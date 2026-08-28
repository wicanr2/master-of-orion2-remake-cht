"""非破壞性匯出 Change_Relations 本體、全部直接 caller 與 reason consumer。"""

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
    "raw_change_relations": 0x4E3B5,
    "raw_reason_payload_helper": 0x4EA03,
    "raw_determine_diplomacy_messages": 0x4EB06,
    "raw_determine_bad_message": 0x4F0DC,
}


def digest(path):
    value = hashlib.sha256()
    with open(path, "rb") as source:
        for chunk in iter(lambda: source.read(1024 * 1024), b""):
            value.update(chunk)
    return value.hexdigest()


def load_symbols(path):
    result = {}
    if not path:
        return result, None
    with open(path, newline="", encoding="utf-8") as source:
        for row in csv.DictReader(source, delimiter="\t"):
            result.setdefault(int(row["ea"], 16), []).append(row["name"])
    return result, digest(path)


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
    function = ida_funcs.get_func(ea)
    if function is None:
        return {"requested": f"0x{ea:X}", "error": "function missing"}
    items = list(idautils.FuncItems(function.start_ea))
    raw = ida_bytes.get_bytes(function.start_ea, function.end_ea - function.start_ea) or b""
    pseudo = None
    if ida_hexrays.init_hexrays_plugin():
        try:
            pseudo = str(ida_hexrays.decompile(function.start_ea))
        except Exception as error:
            pseudo = f"<decompile failed: {error}>"
    return {
        "requested": f"0x{ea:X}",
        "start_ea": f"0x{function.start_ea:X}",
        "end_ea": f"0x{function.end_ea:X}",
        "original_name": ida_name.get_name(function.start_ea) or "<unnamed>",
        "external_names_navigation_only": symbols.get(function.start_ea, []),
        "bytes_sha256": hashlib.sha256(raw).hexdigest(),
        "pseudocode_navigation_only": pseudo,
        "instructions": [instruction(item) for item in items],
    }


def callsite_record(ref_ea, symbols):
    owner = ida_funcs.get_func(ref_ea)
    if owner is None:
        return {"call_ea": f"0x{ref_ea:X}", "error": "caller missing"}
    items = list(idautils.FuncItems(owner.start_ea))
    index = items.index(ref_ea)
    return {
        "call_ea": f"0x{ref_ea:X}",
        "caller_start": f"0x{owner.start_ea:X}",
        "caller_original_name": ida_name.get_name(owner.start_ea) or "<unnamed>",
        "caller_external_names_navigation_only": symbols.get(owner.start_ea, []),
        "window": [instruction(x) for x in items[max(0, index - 24):min(len(items), index + 10)]],
    }


def main():
    ida_auto.auto_wait()
    source = os.environ["MOO2_RE_SOURCE"]
    database = os.environ["MOO2_RE_DATABASE"]
    symbols, symbols_hash = load_symbols(os.environ.get("MOO2_RE_SYMBOLS", ""))
    roots = {name: function_record(ea, symbols) for name, ea in ROOTS.items()}
    root_direct_callsites = {}
    for root_name, root_ea in ROOTS.items():
        root_function = ida_funcs.get_func(root_ea)
        root_direct_callsites[root_name] = [
            callsite_record(ref.frm, symbols)
            for ref in idautils.XrefsTo(root_function.start_ea, 0)
            if ida_funcs.get_func(ref.frm) is not None
        ]
    target = ida_funcs.get_func(ROOTS["raw_change_relations"])
    callsites = []
    callers = {}
    for ref in idautils.XrefsTo(target.start_ea, 0):
        owner = ida_funcs.get_func(ref.frm)
        if owner is None:
            continue
        callsites.append(callsite_record(ref.frm, symbols))
        key = f"0x{owner.start_ea:X}"
        callers[key] = function_record(owner.start_ea, symbols)
    report = {
        "schema": "moo2.ida.change-relations-callers.v1",
        "evidence_scope": "static_only",
        "mutation": "none",
        "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version(),
                 "script": "tools/ida/audit_change_relations_callers.py"},
        "input": {"file": os.path.basename(source), "source_sha256": digest(source),
                  "database_sha256": digest(database), "symbols_sha256": symbols_hash,
                  "processor": ida_ida.inf_get_procname()},
        "address_basis": "IDA linear; DOS/4GW LE object #1",
        "semantic_status": "reviewed_with_evidence_levels",
        "roots": roots,
        "root_direct_callsites": root_direct_callsites,
        "change_relations_callsites": callsites,
        "direct_caller_functions": callers,
    }
    with open(os.environ["MOO2_RE_OUT"], "w", encoding="utf-8") as target_file:
        json.dump(report, target_file, ensure_ascii=False, indent=2)
        target_file.write("\n")


try:
    main()
except Exception:
    error = traceback.format_exc()
    output = os.environ.get("MOO2_RE_OUT", "/tmp/change-relations-callers.json")
    with open(output + ".error", "w", encoding="utf-8") as target:
        target.write(error)
    ida_kernwin.msg(error + "\n")
    ida_pro.qexit(1)
else:
    ida_pro.qexit(0)
