"""非破壞性匯出 AI 外交剩餘 raw 欄位的 operand、owner 與呼叫關係。"""

import hashlib
import json
import os
import traceback

import ida_auto
import ida_bytes
import ida_funcs
import ida_ida
import ida_kernwin
import ida_name
import ida_pro
import ida_ua
import idautils
import idc


FIELDS = {
    "forced_war_flag_60e": (0x60E,),
    "incident_memory_6af": (0x6AF,),
    "incident_memory_6bf": (0x6BF,),
    "colony_damage_grievance_6ff": (0x6FF,),
    "direction_lock_737": (0x737,),
}

ROOTS = {
    "special_war_candidates": 0x25DF1,
    "initialize_diplomacy": 0x4D78E,
    "turn_diplomacy_producers": 0x4DAB2,
    "declare_war": 0x51078,
    "declare_war_tail": 0x5138E,
    "ceasefire": 0x524FB,
    "war_duration": 0x5090C,
    "player_response": 0x533F4,
    "colony_damage_general": 0xDCEBD,
    "colony_damage_special": 0xDD13E,
    "change_population_ownership": 0xECBF7,
    "resolve_colony_capture": 0xECF41,
}


def digest(path):
    h = hashlib.sha256()
    with open(path, "rb") as source:
        for block in iter(lambda: source.read(1024 * 1024), b""):
            h.update(block)
    return h.hexdigest()


def insn(ea):
    decoded = ida_ua.insn_t()
    ida_ua.decode_insn(decoded, ea)
    return {
        "ea": f"0x{ea:X}",
        "bytes": (ida_bytes.get_bytes(ea, decoded.size) or b"").hex(),
        "mnemonic": ida_ua.print_insn_mnem(ea),
        "operands": [idc_print_operand(ea, i) for i in range(8)
                     if idc_print_operand(ea, i)],
    }


def idc_print_operand(ea, index):
    return idc.print_operand(ea, index)


def function_record(ea):
    fn = ida_funcs.get_func(ea)
    if fn is None:
        return {"requested": f"0x{ea:X}", "error": "no function"}
    items = list(idautils.FuncItems(fn.start_ea))
    raw = b"".join(ida_bytes.get_bytes(x, ida_bytes.get_item_size(x)) or b"" for x in items)
    return {
        "requested": f"0x{ea:X}",
        "start_ea": f"0x{fn.start_ea:X}",
        "end_ea": f"0x{fn.end_ea:X}",
        "original_name": ida_name.get_name(fn.start_ea) or "<unnamed>",
        "bytes_sha256": hashlib.sha256(raw).hexdigest(),
        "instruction_count": len(items),
    }


def all_field_sites():
    found = {name: [] for name in FIELDS}
    needles = {
        name: {f"+{offsets[0]:X}h", f"-{offsets[0]:X}h",
               f"+{offsets[0]:x}h", f"-{offsets[0]:x}h"}
        for name, offsets in FIELDS.items()
    }
    for fn_ea in idautils.Functions():
        fn = ida_funcs.get_func(fn_ea)
        items = list(idautils.FuncItems(fn_ea))
        for index, ea in enumerate(items):
            joined = " ".join(idc.print_operand(ea, i) for i in range(3))
            names = [name for name, values in needles.items() if any(n in joined for n in values)]
            if not names:
                continue
            row = insn(ea)
            site = {
                "instruction": row,
                "owner_start": f"0x{fn.start_ea:X}",
                "owner_end": f"0x{fn.end_ea:X}",
                "owner_original_name": ida_name.get_name(fn.start_ea) or "<unnamed>",
                "window": [insn(x) for x in items[max(0, index - 12):index + 13]],
            }
            for name in names:
                found[name].append(site)
    return found


def callers(ea):
    fn = ida_funcs.get_func(ea)
    if fn is None:
        return []
    result = []
    for ref in idautils.XrefsTo(fn.start_ea, 0):
        owner = ida_funcs.get_func(ref.frm)
        if owner is not None:
            result.append({"call_ea": f"0x{ref.frm:X}",
                           "caller_start": f"0x{owner.start_ea:X}",
                           "caller_original_name": ida_name.get_name(owner.start_ea) or "<unnamed>"})
    return result


def main():
    ida_auto.auto_wait()
    source = os.environ["MOO2_RE_SOURCE"]
    database = os.environ["MOO2_RE_DATABASE"]
    sites = all_field_sites()
    report = {
        "schema": "moo2.ida.diplomacy-residual-fields.v1",
        "evidence_scope": "static_only",
        "mutation": "none",
        "semantic_status": "reviewed_against_raw_instructions",
        "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version(),
                 "script": "tools/ida/audit_diplomacy_residual_fields.py"},
        "input": {"file": os.path.basename(source), "source_sha256": digest(source),
                  "database_sha256": digest(database), "processor": ida_ida.inf_get_procname()},
        "address_basis": "IDA linear; DOS/4GW LE object #1",
        "fields": {name: {"offset": f"0x{offsets[0]:X}", "sites": sites[name]}
                   for name, offsets in FIELDS.items()},
        "roots": {name: {"function": function_record(ea), "callers": callers(ea)}
                  for name, ea in ROOTS.items()},
    }
    with open(os.environ["MOO2_RE_OUT"], "w", encoding="utf-8") as target:
        json.dump(report, target, ensure_ascii=False, indent=2)
        target.write("\n")


try:
    main()
except Exception:
    output = os.environ.get("MOO2_RE_OUT", "/tmp/diplomacy-residual-fields.json")
    with open(output + ".error", "w", encoding="utf-8") as target:
        target.write(traceback.format_exc())
finally:
    ida_pro.qexit(0)
