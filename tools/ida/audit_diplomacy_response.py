"""非破壞性匯出 MOO2 玩家外交提案評分、回應模式與直接 consumer。"""

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
    "raw_diplomacy_score_buckets": 0x53146,
    "raw_diplomacy_random_test": 0x533F4,
    "raw_diplomacy_response_modes": 0x539D9,
    "raw_response_gate_helper": 0x53E96,
    "raw_declare_war": 0x51078,
    "raw_set_formal_policy": 0x5232E,
    "raw_human_offer_gift_message": 0x186D3,
    "raw_diplomacy_action_consumer": 0x1DEF8,
    "raw_apply_accepted_demand": 0x1B487,
    "raw_score_clamp_or_adjust": 0x524C3,
    "raw_pair_base_score": 0x51DCE,
    "raw_pair_score_modifier_a": 0x51E98,
    "raw_pair_score_modifier_b": 0x51E3B,
    "raw_pair_relation_score": 0x50FDF,
    "raw_diplomat_leader_bonus": 0xE5E09,
}


def digest(path):
    value = hashlib.sha256()
    with open(path, "rb") as source:
        for chunk in iter(lambda: source.read(1024 * 1024), b""):
            value.update(chunk)
    return value.hexdigest()


def load_symbols(path):
    symbols = {}
    if not path:
        return symbols, None
    with open(path, newline="", encoding="utf-8") as source:
        for row in csv.DictReader(source, delimiter="\t"):
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
    function = ida_funcs.get_func(ea)
    if function is None:
        return {"requested": f"0x{ea:X}", "error": "function missing"}
    raw = ida_bytes.get_bytes(function.start_ea, function.end_ea - function.start_ea) or b""
    pseudo = None
    if ida_hexrays.init_hexrays_plugin():
        try:
            pseudo = str(ida_hexrays.decompile(function.start_ea))
        except Exception as error:
            pseudo = f"<decompile failed: {error}>"
    callers = []
    for ref in idautils.XrefsTo(function.start_ea, 0):
        owner = ida_funcs.get_func(ref.frm)
        if owner is None:
            continue
        item = instruction(ref.frm)
        item.update({
            "caller_start": f"0x{owner.start_ea:X}",
            "caller_original_name": ida_name.get_name(owner.start_ea) or "<unnamed>",
            "caller_external_names_navigation_only": symbols.get(owner.start_ea, []),
        })
        callers.append(item)
    calls = []
    for item_ea in idautils.FuncItems(function.start_ea):
        if idc.print_insn_mnem(item_ea).lower() != "call":
            continue
        target = idc.get_operand_value(item_ea, 0)
        target_function = ida_funcs.get_func(target)
        calls.append({
            "callsite": instruction(item_ea),
            "target": f"0x{target:X}",
            "target_start": f"0x{target_function.start_ea:X}" if target_function else None,
            "target_original_name": ida_name.get_name(target) or "<unnamed>",
            "target_external_names_navigation_only": symbols.get(target, []),
        })
    return {
        "requested": f"0x{ea:X}",
        "start_ea": f"0x{function.start_ea:X}",
        "end_ea": f"0x{function.end_ea:X}",
        "original_name": ida_name.get_name(function.start_ea) or "<unnamed>",
        "external_names_navigation_only": symbols.get(function.start_ea, []),
        "bytes_sha256": hashlib.sha256(raw).hexdigest(),
        "pseudocode_navigation_only": pseudo,
        "instructions": [instruction(x) for x in idautils.FuncItems(function.start_ea)],
        "callers": callers,
        "calls": calls,
    }


def main():
    ida_auto.auto_wait()
    source = os.environ["MOO2_RE_SOURCE"]
    database = os.environ["MOO2_RE_DATABASE"]
    symbols, symbols_hash = load_symbols(os.environ.get("MOO2_RE_SYMBOLS", ""))
    roots = {name: function_record(ea, symbols) for name, ea in ROOTS.items()}
    caller_functions = {}
    for root_name in ("raw_diplomacy_score_buckets", "raw_diplomacy_random_test",
                      "raw_diplomacy_response_modes"):
        for caller in roots[root_name].get("callers", []):
            ea = int(caller["caller_start"], 16)
            caller_functions[caller["caller_start"]] = function_record(ea, symbols)
    report = {
        "schema": "moo2.ida.diplomacy-response.v1",
        "evidence_scope": "static_only",
        "mutation": "none",
        "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version(),
                 "script": "tools/ida/audit_diplomacy_response.py"},
        "input": {"file": os.path.basename(source), "source_sha256": digest(source),
                  "database_sha256": digest(database), "symbols_sha256": symbols_hash,
                  "processor": ida_ida.inf_get_procname()},
        "address_basis": "IDA linear; DOS/4GW LE object #1",
        "semantic_status": "reviewed_with_per_item_confidence_in_diplomacy_response_audit_20260828",
        "roots": roots,
        "direct_caller_functions": caller_functions,
    }
    with open(os.environ["MOO2_RE_OUT"], "w", encoding="utf-8") as target:
        json.dump(report, target, ensure_ascii=False, indent=2)
        target.write("\n")


try:
    main()
except Exception:
    error = traceback.format_exc()
    output = os.environ.get("MOO2_RE_OUT", "/tmp/diplomacy-response.json")
    with open(output + ".error", "w", encoding="utf-8") as target:
        target.write(error)
    ida_kernwin.msg(error + "\n")
    ida_pro.qexit(1)
else:
    ida_pro.qexit(0)
