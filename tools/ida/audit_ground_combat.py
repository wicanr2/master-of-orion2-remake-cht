"""非破壞性匯出 MOO2 一般地面入侵、解算與戰後回寫鏈。"""

import csv
import hashlib
import json
import os
import traceback

import ida_auto
import ida_bytes
import ida_funcs
import ida_hexrays
import ida_ida
import ida_kernwin
import ida_name
import ida_pro
import ida_ua
import idautils
import idc


ROOTS = {
    "raw_AI_Invasion_Will_Succeed": 0xCF26E,
    "raw_AI_Player_Invades": 0xCF34C,
    "raw_Produce_Ground_Military": 0xE3616,
    "raw_Player_Owns_Transports": 0xE8776,
    "raw_Do_Attacker_Beat_Colony_Stuff": 0xE87D2,
    "raw_Do_1_Combat": 0xE938C,
    "raw_Compute_Player_Ground_Combat_Bonuses": 0xEC15C,
    "raw_Compute_Colony_Officer_Bonus": 0xEC2AF,
    "raw_Compute_Attacker_Officer_Bonus": 0xEC2FD,
    "raw_Compute_Ground_Combat_Info": 0xEC3CE,
    "raw_Ground_Combat_Round": 0xEC4FE,
    "raw_Resolve_Ground_Combat": 0xEC601,
    "raw_Colony_N_Militia": 0xEC61E,
    "raw_Get_Invasion_Info": 0xEC831,
    "raw_Resolve_Invasion_Troops": 0xECE05,
    "raw_Change_Pop_Ownership": 0xECBF7,
    "raw_Change_Colony_Ownership": 0xECF41,
    "raw_Invade": 0xED48B,
    "raw_Colony_Infantry_Limit": 0xED59D,
    "raw_Colony_Infantry_Barracks_Limit": 0xED5E7,
    "raw_Colony_Tank_Limit": 0xED61D,
    "raw_Colony_Tank_Barracks_Limit": 0xED674,
    "raw_Unload_Transports": 0xED6B7,
    "raw_Compute_Colony_Ground_Combat_Info": 0xED713,
}


def digest(path):
    h = hashlib.sha256()
    with open(path, "rb") as stream:
        for chunk in iter(lambda: stream.read(1024 * 1024), b""):
            h.update(chunk)
    return h.hexdigest()


def load_symbols(path):
    symbols = {}
    if not path:
        return symbols, None
    with open(path, newline="", encoding="utf-8") as stream:
        for row in csv.DictReader(stream, delimiter="\t"):
            symbols.setdefault(int(row["ea"], 16), []).append(row["name"])
    return symbols, digest(path)


def instruction(ea):
    insn = ida_ua.insn_t()
    size = ida_ua.decode_insn(insn, ea) or 1
    return {
        "ea": f"0x{ea:X}",
        "bytes": (ida_bytes.get_bytes(ea, size) or b"").hex(),
        "text": idc.generate_disasm_line(ea, 0) or "",
        "mnem": idc.print_insn_mnem(ea),
        "op0": idc.print_operand(ea, 0),
        "op1": idc.print_operand(ea, 1),
        "code_refs": [f"0x{x:X}" for x in idautils.CodeRefsFrom(ea, 0)],
        "data_refs": [f"0x{x:X}" for x in idautils.DataRefsFrom(ea)],
    }


def function_record(ea, symbols):
    fn = ida_funcs.get_func(ea)
    if fn is None:
        return {"requested": f"0x{ea:X}", "error": "function missing"}
    raw = ida_bytes.get_bytes(fn.start_ea, fn.end_ea - fn.start_ea) or b""
    pseudo = None
    if ida_hexrays.init_hexrays_plugin():
        try:
            pseudo = str(ida_hexrays.decompile(fn.start_ea))
        except Exception as exc:
            pseudo = f"<decompile failed: {exc}>"
    callers = []
    for xref in idautils.XrefsTo(fn.start_ea, 0):
        caller_fn = ida_funcs.get_func(xref.frm)
        if caller_fn is None:
            continue
        record = instruction(xref.frm)
        record.update({
            "caller_function_start": f"0x{caller_fn.start_ea:X}",
            "caller_original_name": ida_name.get_name(caller_fn.start_ea) or "<unnamed>",
            "caller_external_symbol_names_navigation_only": symbols.get(caller_fn.start_ea, []),
        })
        callers.append(record)
    calls = []
    for item in idautils.FuncItems(fn.start_ea):
        if idc.print_insn_mnem(item).lower() != "call":
            continue
        target = idc.get_operand_value(item, 0)
        target_fn = ida_funcs.get_func(target)
        calls.append({
            "callsite": instruction(item),
            "target": f"0x{target:X}",
            "target_function_start": f"0x{target_fn.start_ea:X}" if target_fn else None,
            "ida_original_name": ida_name.get_name(target) or "<unnamed>",
            "external_symbol_names_navigation_only": symbols.get(target, []),
        })
    return {
        "requested": f"0x{ea:X}",
        "original_name": ida_name.get_name(fn.start_ea) or "<unnamed>",
        "external_symbol_names_navigation_only": symbols.get(fn.start_ea, []),
        "start_ea": f"0x{fn.start_ea:X}",
        "end_ea": f"0x{fn.end_ea:X}",
        "bytes_sha256": hashlib.sha256(raw).hexdigest(),
        "pseudocode_navigation_only": pseudo,
        "instructions": [instruction(item) for item in idautils.FuncItems(fn.start_ea)],
        "callers": callers,
        "calls": calls,
    }


def main():
    ida_auto.auto_wait()
    source = os.environ["MOO2_RE_SOURCE"]
    database = os.environ["MOO2_RE_DATABASE"]
    symbols, symbols_hash = load_symbols(os.environ.get("MOO2_RE_SYMBOLS", ""))
    report = {
        "schema": "moo2.ida.ground-combat.v1",
        "evidence_scope": "static_only",
        "mutation": "none",
        "tool": {
            "name": "IDA Pro",
            "version": ida_kernwin.get_kernel_version(),
            "script": "tools/ida/audit_ground_combat.py",
        },
        "input": {
            "file": os.path.basename(source),
            "source_sha256": digest(source),
            "database_sha256": digest(database),
            "symbols_sha256": symbols_hash,
            "processor": ida_ida.inf_get_procname(),
        },
        "address_basis": "IDA linear; DOS/4GW LE object #1",
        "semantic_status": "reviewed_in_docs/re/ground-combat-audit-20260828.md",
        "roots": {name: function_record(ea, symbols) for name, ea in ROOTS.items()},
    }
    with open(os.environ["MOO2_RE_OUT"], "w", encoding="utf-8") as stream:
        json.dump(report, stream, ensure_ascii=False, indent=2)
        stream.write("\n")


try:
    main()
except Exception:
    error = traceback.format_exc()
    out = os.environ.get("MOO2_RE_OUT", "/tmp/ground-combat.json")
    with open(out + ".error", "w", encoding="utf-8") as stream:
        stream.write(error)
    ida_kernwin.msg(error + "\n")
    ida_pro.qexit(1)
else:
    ida_pro.qexit(0)
