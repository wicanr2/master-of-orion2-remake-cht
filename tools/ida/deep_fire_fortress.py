"""Non-destructive IDA Pro 9.4 probe for fighter fire and Antaran fortress data.

This probe intentionally keeps IDA's raw names, linear addresses and operands.  It
does not rename, comment, retype or save the input database.  The output is a
reproducible evidence dump; interpretation belongs in the versioned research note.
"""

import hashlib
import os

import ida_bytes
import ida_funcs
import ida_hexrays
import ida_idaapi
import ida_kernwin
import ida_nalt
import idautils
import idc


OUT = os.environ.get("DEEP_OUT", "/host-tmp/moo2-fire-fortress-ida.txt")
EXE = os.environ.get("ORACLE_EXE", "/private/re/Orion2.exe")
DB = os.environ.get("ORACLE_DB", ida_nalt.get_input_file_path())


def sha256(path):
    try:
        digest = hashlib.sha256()
        with open(path, "rb") as handle:
            for chunk in iter(lambda: handle.read(1024 * 1024), b""):
                digest.update(chunk)
        return digest.hexdigest()
    except OSError as exc:
        return "unavailable:%s" % exc


def disasm(ea):
    return idc.generate_disasm_line(ea, 0) or "<no disassembly>"


def name_at(ea):
    return idc.get_name(ea, idc.GN_VISIBLE) or "<unnamed>"


def fn_for(ea):
    return ida_funcs.get_func(ea)


def ensure_fn(ea):
    fn = fn_for(ea)
    if fn is None:
        ida_funcs.add_func(ea)
        fn = fn_for(ea)
    return fn


def instruction_range(fn, limit=None):
    if fn is None:
        return
    cur = fn.start_ea
    count = 0
    while cur != ida_idaapi.BADADDR and cur < fn.end_ea:
        yield cur
        count += 1
        if limit is not None and count >= limit:
            break
        nxt = idc.next_head(cur, fn.end_ea)
        if nxt == ida_idaapi.BADADDR or nxt <= cur:
            break
        cur = nxt


def prev_heads(ea, count):
    out = []
    cur = ea
    for _ in range(count):
        cur = idc.prev_head(cur)
        if cur == ida_idaapi.BADADDR:
            break
        out.append(cur)
    return list(reversed(out))


def next_heads(ea, count, limit=0x100000):
    out = []
    cur = ea
    for _ in range(count):
        cur = idc.next_head(cur, limit)
        if cur == ida_idaapi.BADADDR:
            break
        out.append(cur)
    return out


def call_target(ea):
    if idc.print_insn_mnem(ea).lower() not in ("call", "callf"):
        return None
    value = idc.get_operand_value(ea, 0)
    if value in (None, ida_idaapi.BADADDR):
        return None
    return value


def function_header(handle, ea, label):
    fn = ensure_fn(ea)
    handle.write("\n[function %s]\n" % label)
    handle.write("requested=0x%X raw_name=%s\n" % (ea, name_at(ea)))
    if fn is None:
        handle.write("status=no_function\n")
        return None
    handle.write("start=0x%X end=0x%X name_at_start=%s\n" %
                 (fn.start_ea, fn.end_ea, name_at(fn.start_ea)))
    return fn


def dump_function(handle, ea, label, limit=2000):
    fn = function_header(handle, ea, label)
    if fn is None:
        return
    calls = []
    for cur in instruction_range(fn, limit):
        line = disasm(cur)
        target = call_target(cur)
        if target is not None:
            calls.append((cur, target, line))
        handle.write("0x%X %s\n" % (cur, line))
    if len(list(instruction_range(fn, limit))) >= limit:
        handle.write("... instruction limit reached\n")
    handle.write("calls=%d\n" % len(calls))
    for call_ea, target, line in calls:
        handle.write("call_site=0x%X target=0x%X target_name=%s %s\n" %
                     (call_ea, target, name_at(target), line))


def dump_callers(handle, ea, label, context=12):
    handle.write("\n[callers %s]\n" % label)
    refs = list(idautils.CodeRefsTo(ea, 0))
    handle.write("target=0x%X target_name=%s count=%d\n" % (ea, name_at(ea), len(refs)))
    for ref in refs:
        fn = fn_for(ref)
        handle.write("caller_site=0x%X caller_fn=%s caller_start=%s\n" %
                     (ref, name_at(fn.start_ea) if fn else "<none>",
                      "0x%X" % fn.start_ea if fn else "<none>"))
        for cur in prev_heads(ref, context) + [ref] + next_heads(ref, context, fn.end_ea if fn else ref + 0x80):
            handle.write("  0x%X %s\n" % (cur, disasm(cur)))


def dump_data(handle, ea, label, size=64, unit=1):
    handle.write("\n[data %s]\n" % label)
    handle.write("ea=0x%X name=%s size=%d unit=%d\n" % (ea, name_at(ea), size, unit))
    blob = ida_bytes.get_bytes(ea, size)
    if blob is None:
        handle.write("bytes=<unavailable>\n")
        return
    handle.write("bytes=%s\n" % blob.hex())
    for offset in range(0, size, unit):
        if offset + unit > len(blob):
            break
        if unit == 1:
            value = blob[offset]
        elif unit == 2:
            value = int.from_bytes(blob[offset:offset + 2], "little")
        elif unit == 4:
            value = int.from_bytes(blob[offset:offset + 4], "little")
        else:
            value = blob[offset:offset + unit].hex()
        handle.write("  +0x%X = 0x%X\n" % (offset, value) if isinstance(value, int)
                     else "  +0x%X = %s\n" % (offset, value))


def dump_drefs(handle, ea, label, limit=100):
    handle.write("\n[data_xrefs %s]\n" % label)
    refs = list(idautils.DataRefsTo(ea))
    handle.write("ea=0x%X name=%s count=%d\n" % (ea, name_at(ea), len(refs)))
    for ref in refs[:limit]:
        handle.write("  ref=0x%X fn=%s %s\n" % (ref, name_at(ref), disasm(ref)))


def dump_decomp(handle, ea, label):
    handle.write("\n[decompiler %s]\n" % label)
    fn = fn_for(ea)
    if fn is None:
        handle.write("status=no_function\n")
        return
    try:
        if not ida_hexrays.init_hexrays_plugin():
            handle.write("status=hexrays_unavailable\n")
            return
        cfunc = ida_hexrays.decompile(fn.start_ea)
        handle.write(str(cfunc) if cfunc else "status=no_cfunc")
        handle.write("\n")
    except Exception as exc:
        handle.write("status=decompile_error:%r\n" % (exc,))


def main():
    ida_kernwin.msg("deep fighter/fortress probe: %s\n" % OUT)
    with open(OUT, "w", encoding="utf-8") as handle:
        handle.write("MOO2 DEEP FIGHTER/FORTRESS IDA PROBE\n")
        handle.write("tool=IDA Pro 9.4; image=ida-pro-9.4-ver2:latest; static_only=true\n")
        handle.write("input_db=%s\n" % DB)
        handle.write("input_db_sha256=%s\n" % sha256(DB))
        handle.write("input_exe=%s\n" % EXE)
        handle.write("input_exe_sha256=%s\n" % sha256(EXE))
        handle.write("address_basis=IDA linear addresses; raw names/operands preserved\n")
        handle.write("min_ea=0x%X max_ea=0x%X\n" %
                     (ida_idaapi.get_inf_structure().min_ea,
                      ida_idaapi.get_inf_structure().max_ea))

        fighter = [
            (0x3AC20, "raw_sub_3AC20_Fire_Fighter_Bomb_candidate"),
            (0x3AD57, "raw_sub_3AD57_Fire_Fighter_Beam_candidate"),
            (0x3DF8D, "raw_sub_3DF8D_fighter_downstream"),
            (0x3DFE0, "raw_sub_3DFE0_fighter_downstream"),
            (0x3CD21, "raw_sub_3CD21_fighter_speed_or_ocv"),
            (0x3C892, "raw_sub_3C892_fighter_runtime_copy"),
            (0x39985, "raw_sub_39985_damage_consumer"),
            (0x3A0B9, "raw_sub_3A0B9_damage_consumer"),
            (0x3E095, "raw_sub_3E095_missile_or_fighter_dcv"),
            (0x2B7CC, "raw_sub_2B7CC_fighter_callsite"),
            (0x3D839, "raw_sub_3D839_fighter_callsite"),
            (0x3D884, "raw_sub_3D884_fighter_callsite"),
            (0x3D2DF, "raw_sub_3D2DF_fighter_callsite"),
            (0x38B5E, "raw_sub_38B5E_fighter_callsite"),
        ]
        for ea, label in fighter:
            dump_function(handle, ea, label, 1600)
            dump_callers(handle, ea, label)
            dump_decomp(handle, ea, label)

        fortress = [
            (0x4D18E, "raw_sub_4D18E_antaran_fortress_loader"),
            (0x6EE8E, "raw_sub_6EE8E_fortress_capacity_candidate"),
            (0x6EFEB, "raw_sub_6EFEB_fortress_helper_candidate"),
            (0x6A636, "raw_sub_6A636_fortress_helper_candidate"),
            (0x6A406, "raw_sub_6A406_fortress_helper_candidate"),
            (0x6F11C, "raw_sub_6F11C_fortress_helper_candidate"),
            (0x6E70A, "raw_sub_6E70A_fortress_helper_candidate"),
            (0x6E60E, "raw_sub_6E60E_fortress_helper_candidate"),
        ]
        for ea, label in fortress:
            dump_function(handle, ea, label, 1800)
            dump_callers(handle, ea, label)
            dump_decomp(handle, ea, label)

        for ea, label, size, unit in [
            (0x192864, "global_design_table_candidate", 0x139 * 13, 1),
            (0x19917A, "antaran_size_counts_candidate", 0x20, 2),
            (0x180140, "fortress_constant_block_candidate", 0x30, 1),
            (0x19988E, "fortress_global_word_candidate", 0x20, 2),
            (0x17F642, "ship_class_power_table_candidate", 0x90, 2),
            (0x17F807, "weapon_table_candidate", 0x1C * 40, 1),
        ]:
            dump_data(handle, ea, label, size, unit)
            dump_drefs(handle, ea, label)

        handle.write("\n[all_calls_to_fighter_and_fortress_ranges]\n")
        ranges = [(0x3A000, 0x3E300), (0x4D000, 0x4D900), (0x6E000, 0x6F400)]
        targets = {ea for ea, _ in fighter + fortress}
        for fn_ea in idautils.Functions():
            fn = fn_for(fn_ea)
            if fn is None:
                continue
            if not any(lo <= fn.start_ea < hi for lo, hi in ranges):
                continue
            for cur in instruction_range(fn, 2200):
                target = call_target(cur)
                if target in targets:
                    handle.write("caller_fn=0x%X site=0x%X target=0x%X target_name=%s %s\n" %
                                 (fn.start_ea, cur, target, name_at(target), disasm(cur)))

    idc.qexit(0)


if __name__ == "__main__":
    main()
