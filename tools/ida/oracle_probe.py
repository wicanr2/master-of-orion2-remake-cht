"""IDA Pro 9.4 static oracle probe for the private DOS MOO2 executable.

This script is deliberately non-destructive: it only reads the opened database,
keeps original IDA names/addresses, and writes a bounded text report. It does
not rename functions, change comments, or save the private .i64 database.
"""

import hashlib
import os

import ida_bytes
import ida_funcs
import ida_idaapi
import ida_nalt
import idautils
import idc


OUT = os.environ.get("ORACLE_OUT", "/tmp/oracle-ida-report.txt")
INPUT_SHA_PATH = os.environ.get("ORACLE_INPUT", "/tmp/Orion2.exe")


def sha256(path):
    h = hashlib.sha256()
    with open(path, "rb") as f:
        while True:
            chunk = f.read(1024 * 1024)
            if not chunk:
                break
            h.update(chunk)
    return h.hexdigest()


def name_at(ea):
    return idc.get_name(ea, idc.GN_VISIBLE) or "<unnamed>"


def disasm(ea):
    return idc.generate_disasm_line(ea, 0) or "<no disassembly>"


def function_for(ea):
    fn = ida_funcs.get_func(ea)
    if fn is None:
        return None
    return fn


def xref_lines(ea, limit=20):
    out = []
    for i, ref in enumerate(idautils.XrefsTo(ea, 0)):
        if i >= limit:
            out.append("    ... xref limit reached")
            break
        out.append("    0x%X %s" % (ref.frm, disasm(ref.frm)))
    return out


def function_dump(ea, max_insns=96):
    fn = function_for(ea)
    if fn is None:
        return ["  no function boundary at 0x%X" % ea]
    lines = [
        "  function original_name=%s start=0x%X end=0x%X" %
        (name_at(fn.start_ea), fn.start_ea, fn.end_ea),
    ]
    cur = fn.start_ea
    count = 0
    while cur != ida_idaapi.BADADDR and cur < fn.end_ea and count < max_insns:
        lines.append("    0x%X %s" % (cur, disasm(cur)))
        nxt = idc.next_head(cur, fn.end_ea)
        if nxt == ida_idaapi.BADADDR or nxt <= cur:
            break
        cur = nxt
        count += 1
    if count >= max_insns:
        lines.append("    ... instruction limit reached")
    return lines


def matching_functions(patterns, limit=80):
    found = []
    for ea in idautils.Functions():
        n = name_at(ea)
        if any(p in n.upper() for p in patterns):
            found.append((ea, n))
    return found[:limit]


def matching_strings(patterns, limit=120):
    found = []
    for s in idautils.Strings():
        text = str(s)
        upper = text.upper()
        if any(p in upper for p in patterns):
            found.append((s.ea, text))
            if len(found) >= limit:
                break
    return found


def exact_symbol_candidates(names):
    out = []
    for name in names:
        ea = idc.get_name_ea_simple(name)
        if ea != ida_idaapi.BADADDR:
            out.append((name, ea))
    return out


def main():
    inf = ida_idaapi.get_inf_structure()
    lines = []
    lines.append("MOO2 ORACLE STATIC PROBE")
    lines.append("evidence_level=strong_inference_or_proven_only; runtime_not_executed")
    lines.append("ida_version=%s" % ida_idaapi.get_kernel_version())
    lines.append("input_path=%s" % ida_nalt.get_input_file_path())
    lines.append("input_sha256_path=%s" % INPUT_SHA_PATH)
    lines.append("input_sha256=%s" % sha256(INPUT_SHA_PATH))
    lines.append("processor=%s filetype=%s min_ea=0x%X max_ea=0x%X" %
                 (inf.procName, ida_idaapi.get_file_type_name(inf.filetype), inf.min_ea, inf.max_ea))
    lines.append("address_basis=IDA linear addresses; DOS LE object #1 code base 0x10000")
    lines.append("")

    exact = [
        "VESA_Init_", "VESA.COM", "Fire_Fighter_Bomb", "Fire_Fighter_Bomb_",
        "Start_Diplomacy_Music_", "Play_Diplomacy_Music_", "Load_Game_",
        "Load_Leaders_", "Save_Game_", "Read_Game_", "Write_Game_",
    ]
    lines.append("[exact symbol candidates]")
    for requested, ea in exact_symbol_candidates(exact):
        lines.append("symbol requested=%s ea=0x%X original_name=%s" % (requested, ea, name_at(ea)))
        lines.extend(xref_lines(ea))
        lines.extend(function_dump(ea, 72))

    patterns = [
        "VESA", "FIRE_FIGHTER", "FIGHTER", "BOMB", "DIPLOM", "LEADER",
        "GAM", "SAVE", "LOAD", "BLUEPRINT", "CMBTSHP", "ARM", "FST",
    ]
    lines.append("")
    lines.append("[matching original function names]")
    for ea, n in matching_functions(patterns):
        lines.append("function ea=0x%X original_name=%s" % (ea, n))
        lines.extend(xref_lines(ea, 8))

    lines.append("")
    lines.append("[matching string literals and xrefs]")
    for ea, text in matching_strings(patterns):
        safe = text.replace("\n", "\\n")[:240]
        lines.append("string ea=0x%X text=%r" % (ea, safe))
        lines.extend(xref_lines(ea, 8))

    lines.append("")
    lines.append("[bounded function dumps around selected names]")
    dumped = set()
    for ea, n in matching_functions(
            ["VES", "DIPLOM", "FIGHTER", "FIRE", "BOMB", "LEADER", "GAME", "GAM"], 40):
        fn = function_for(ea)
        if fn is None or fn.start_ea in dumped:
            continue
        dumped.add(fn.start_ea)
        lines.append("selected original_name=%s ea=0x%X" % (n, fn.start_ea))
        lines.extend(function_dump(fn.start_ea, 48))

    with open(OUT, "w", encoding="utf-8") as f:
        f.write("\n".join(lines) + "\n")
    print("oracle probe wrote %s (%d lines)" % (OUT, len(lines)))
    idc.qexit(0)


if __name__ == "__main__":
    main()
