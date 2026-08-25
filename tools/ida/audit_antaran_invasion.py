"""唯讀匯出安塔蘭週期入侵的排程、資源、建艦、目標與部署鏈。"""

import hashlib
import json
import os

import ida_auto
import ida_bytes
import ida_funcs
import ida_ida
import ida_kernwin
import ida_nalt
import ida_pro
import idautils
import idc

OUT = os.environ.get("MOO2_RE_OUT", "/out/antaran-invasion.json")
SOURCE = os.environ.get("MOO2_RE_SOURCE", ida_nalt.get_input_file_path())
ROOT = 0x63D92
ROOTS = {
    "raw_Next_Turn_Calc": 0x136B3,
    "raw_New_Game_Setting_Gates": 0x13B0F,
    "raw_Antaran_Invasion_Check": 0x63D92,
    "raw_Antaran_Base_Resource": 0x63DDA,
    "raw_Antaran_Last_Invasion_Turn": 0x63E4C,
    "raw_Antaran_Resource_Component": 0x63E73,
    "raw_Build_Antaran_Defensive_Ships": 0x63F9C,
    "raw_Build_Antaran_Offensive_Ships": 0x63FCB,
    "raw_Build_Antaran_Ships_Common": 0x63FF0,
    "raw_Select_Antaran_Target_Or_Deploy": 0x643A0,
    "raw_Antaran_Ship_Size_Selector": 0x645A4,
    "raw_Accumulate_Antaran_Resources": 0x645EC,
    "raw_Antaran_Resource_Split": 0x646BD,
    "raw_Antaran_Target_Selector": 0x6478D,
    "raw_Antaran_Target_Candidates": 0x64311,
    "raw_Antaran_Defensive_Pressure": 0x647D7,
    "raw_Antaran_Offensive_Pressure": 0x6481B,
    "raw_Event_Empire_Target_Helper": 0x22F5C,
    "raw_Find_Player_Stack": 0xA1762,
    "raw_Find_Target_Stack": 0x100010,
    "raw_Antaran_Target_Stack_Eligible": 0x7826A,
}
RAW_DATA = {
    "raw_Antaran_Ship_Strength_Table": (0x181734, 9),
    "raw_Antaran_Offensive_Ship_Maxima": (0x18173D, 9),
    "raw_Antaran_Defensive_Ship_Maxima": (0x181746, 9),
}
RAW_GLOBALS = {
    "raw_Current_Stardate": 0x192FD8,
    "raw_Difficulty": 0x199CB0,
    "raw_Antaran_Setting": 0x199CB5,
    "raw_Antaran_Trigger_Accumulator": 0x199174,
    "raw_Antaran_Offensive_Resource": 0x199176,
    "raw_Antaran_Defensive_Resource": 0x199178,
    "raw_Antaran_Last_Invasion_Stardate": 0x1991B1,
}


def sha256(path):
    h = hashlib.sha256()
    with open(path, "rb") as f:
        for chunk in iter(lambda: f.read(1024 * 1024), b""):
            h.update(chunk)
    return h.hexdigest()


def insn(ea):
    size = idc.get_item_size(ea)
    return {"ea": f"0x{ea:X}", "bytes": (ida_bytes.get_bytes(ea, size) or b"").hex(),
            "mnem": idc.print_insn_mnem(ea), "op0": idc.print_operand(ea, 0),
            "op1": idc.print_operand(ea, 1), "line": idc.generate_disasm_line(ea, 0) or "<unavailable>"}


def function(ea):
    f = ida_funcs.get_func(ea)
    if not f:
        return {"requested": f"0x{ea:X}", "error": "no function"}
    items = list(idautils.FuncItems(f.start_ea))
    calls = []
    for x in items:
        if idc.print_insn_mnem(x).lower() != "call":
            continue
        targets = list(idautils.CodeRefsFrom(x, 0))
        calls.append({"instruction": insn(x), "targets": [f"0x{target:X}" for target in targets]})
    return {"requested": f"0x{ea:X}", "start": f"0x{f.start_ea:X}", "end": f"0x{f.end_ea:X}",
            "original_name": idc.get_name(f.start_ea) or "<unnamed>",
            "instructions": [insn(x) for x in items],
            "callers": [insn(x) for x in idautils.CodeRefsTo(f.start_ea, 0)], "calls": calls}


def antaran_names():
    matches = []
    for ea, name in idautils.Names():
        low = name.lower()
        if "antaran" not in low and "antares" not in low:
            continue
        size = idc.get_item_size(ea)
        matches.append({"ea": f"0x{ea:X}", "name": name, "item_size": size,
                        "bytes_64": (ida_bytes.get_bytes(ea, min(max(size, 1), 64)) or b"").hex(),
                        "data_refs_to": [insn(x) for x in idautils.DataRefsTo(ea)],
                        "code_refs_to": [insn(x) for x in idautils.CodeRefsTo(ea, 0)]})
    return matches


def main():
    ida_auto.auto_wait()
    root = function(ROOT)
    callees = {}
    for call in root.get("calls", []):
        for target_text in call["targets"]:
            target = int(target_text, 16)
            f = ida_funcs.get_func(target)
            if f:
                callees[f.start_ea] = function(f.start_ea)
    named_functions = {}
    for ea, name in idautils.Names():
        if ("antaran" in name.lower() or "antares" in name.lower()) and ida_funcs.get_func(ea):
            f = ida_funcs.get_func(ea)
            named_functions[f.start_ea] = function(f.start_ea)
    report = {
        "schema": "moo2.ida.re-evidence.v1", "evidence_scope": "static_only", "mutation": "none",
        "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version(),
                 "script": "tools/ida/audit_antaran_invasion.py"},
        "input": {"database": ida_nalt.get_input_file_path(), "source": SOURCE,
                  "source_sha256": sha256(SOURCE), "processor": ida_ida.inf_get_procname(),
                  "min_ea": f"0x{ida_ida.inf_get_min_ea():X}", "max_ea": f"0x{ida_ida.inf_get_max_ea():X}"},
        "address_basis": "IDA linear; DOS/4GW LE object #1",
        "root": root,
        "roots": {name: function(ea) for name, ea in ROOTS.items()},
        "raw_data": {name: {"ea": f"0x{ea:X}", "size": size,
                             "bytes": (ida_bytes.get_bytes(ea, size) or b"").hex(),
                             "refs": [insn(x) for x in idautils.DataRefsTo(ea)]}
                     for name, (ea, size) in RAW_DATA.items()},
        "raw_globals": {name: {"ea": f"0x{ea:X}",
                               "refs": [insn(x) for x in idautils.DataRefsTo(ea)]}
                        for name, ea in RAW_GLOBALS.items()},
        "direct_callees": {f"0x{ea:X}": value for ea, value in sorted(callees.items())},
        "named_functions": {f"0x{ea:X}": value for ea, value in sorted(named_functions.items())},
        "named_symbols": antaran_names(),
    }
    os.makedirs(os.path.dirname(OUT), exist_ok=True)
    with open(OUT, "w", encoding="utf-8") as f:
        json.dump(report, f, ensure_ascii=False, indent=2)
        f.write("\n")
    ida_pro.qexit(0)


if __name__ == "__main__":
    main()
