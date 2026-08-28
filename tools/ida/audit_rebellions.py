"""非破壞性匯出 MOO2 叛亂檢定、地面戰與殖民地回寫鏈。"""

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
    "raw_Compute_Player_Ground_Combat_Bonuses": 0xEC15C,
    "raw_Compute_Ground_Combat_Info": 0xEC3CE,
    "raw_Resolve_Ground_Combat": 0xEC601,
    "raw_Get_Rebellion_Info": 0xEC65A,
    "raw_Change_Pop_Ownership": 0xECBF7,
    "raw_Resolve_Rebellion_Troops": 0xECECA,
    "raw_Change_Colony_Ownership": 0xECF41,
    "raw_Check_Rebellion": 0xED260,
    "raw_Check_All_Rebellions": 0xED44A,
    "raw_Do_Colony_Owner_Change_Star_Colony_Stuff": 0xE53CD,
    "raw_Check_For_Eliminated_Players": 0xE4EB3,
    "raw_Eliminate_Player": 0xE45FF,
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
        "schema": "moo2.ida.rebellions.v1",
        "evidence_scope": "static_only",
        "mutation": "none",
        "tool": {
            "name": "IDA Pro",
            "version": ida_kernwin.get_kernel_version(),
            "script": "tools/ida/audit_rebellions.py",
        },
        "input": {
            "file": os.path.basename(source),
            "source_sha256": digest(source),
            "database_sha256": digest(database),
            "symbols_sha256": symbols_hash,
            "processor": ida_ida.inf_get_procname(),
        },
        "address_basis": "IDA linear; DOS/4GW LE object #1",
        "semantic_status": "reviewed_in_docs/re/rebellions-audit-20260828.md",
        "roots": {name: function_record(ea, symbols) for name, ea in ROOTS.items()},
    }
    out = os.environ["MOO2_RE_OUT"]
    with open(out, "w", encoding="utf-8") as stream:
        json.dump(report, stream, ensure_ascii=False, indent=2)
        stream.write("\n")


try:
    main()
except Exception:
    error = traceback.format_exc()
    out = os.environ.get("MOO2_RE_OUT", "/tmp/rebellions.json")
    with open(out + ".error", "w", encoding="utf-8") as stream:
        stream.write(error)
    ida_kernwin.msg(error + "\n")
    ida_pro.qexit(1)
else:
    ida_pro.qexit(0)
