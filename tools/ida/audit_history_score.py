"""唯讀匯出 MOO2 歷史記錄與最終得分的原始函式、caller 與 callee。"""

import hashlib
import json
import os

import ida_auto
import ida_bytes
import ida_funcs
import ida_hexrays
import ida_ida
import ida_kernwin
import ida_nalt
import ida_pro
import idautils
import idc


OUT = os.environ.get("MOO2_RE_OUT", "/out/history-score.json")
SOURCE = os.environ.get("MOO2_RE_SOURCE", ida_nalt.get_input_file_path())
ROOTS = {
    "raw_Next_Turn_Calc": 0x136B3,
    "raw_Technology_Value_Consumer": 0x21B6D,
    "raw_Recompute_Technology_Value": 0xE4535,
    "raw_Research_Completion_Caller": 0xE4FA8,
    "raw_Record_History": 0x10208A,
    "raw_History_Building_Summary": 0xE2671,
    "raw_Score_Orchestrator": 0x9D986,
    "raw_Score_Multiplier": 0x58F4A,
    "raw_Get_Antares_Score": 0x9E711,
    "raw_Get_Captured_Colonist_Score": 0x9E735,
    "raw_Get_Elimination_Score": 0x9E84C,
    "raw_Get_Orion_Score": 0x9E8A3,
    "raw_Get_Population_Score": 0x9E8C7,
    "raw_Technology_Count_Helper": 0x9E90B,
    "raw_Get_Technology_Score": 0x9E973,
    "raw_Get_Time_Score": 0x9E993,
    "raw_Get_Won_Council_Score": 0x9EA17,
}
WINDOWS = {
    "raw_Antares_Score_Window": (0x9E6E0, 0x9E735),
    "raw_Orion_Score_Window": (0x9E880, 0x9E8C7),
    "raw_Council_Score_Window": (0x9E9D0, 0x9EA50),
}
NAMED_ROOTS = {
    "raw_Draw_History_Subscreen": ["Draw_History_Subscreen_", "Draw_History_Subscreen"],
    "raw_Draw_Hi_Score_Screen": ["Draw_Hi_Score_Screen_", "Draw_Hi_Score_Screen"],
    "raw_Draw_Hall_Of_Fame_Screen": ["Draw_Hall_Of_Fame_Screen_", "Draw_Hall_Of_Fame_Screen"],
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
    calls = []
    for x in idautils.FuncItems(f.start_ea):
        if idc.print_insn_mnem(x).lower() == "call":
            target = idc.get_operand_value(x, 0)
            calls.append({"instruction": insn(x), "target": f"0x{target:X}",
                          "target_name": idc.get_name(target) or "<unnamed>"})
    pseudo = None
    if ida_hexrays.init_hexrays_plugin():
        try:
            pseudo = str(ida_hexrays.decompile(f.start_ea))
        except Exception as exc:
            pseudo = f"<decompile failed: {exc}>"
    return {"requested": f"0x{ea:X}", "start": f"0x{f.start_ea:X}", "end": f"0x{f.end_ea:X}",
            "original_name": idc.get_name(f.start_ea) or "<unnamed>",
            "pseudocode_navigation_only": pseudo,
            "instructions": [insn(x) for x in idautils.FuncItems(f.start_ea)],
            "callers": [insn(x) for x in idautils.CodeRefsTo(f.start_ea, 0)], "callees": calls}


def resolve_named(candidates):
    for name in candidates:
        ea = idc.get_name_ea_simple(name)
        if ea != idc.BADADDR:
            return {"matched_name": name, "evidence": function(ea)}
    return {"matched_name": None, "candidates": candidates, "error": "name not found"}


def linear_window(start, end):
    out = []
    ea = start
    while ea < end:
        out.append(insn(ea))
        nxt = idc.next_head(ea, end)
        if nxt == idc.BADADDR or nxt <= ea:
            break
        ea = nxt
    return {"start": f"0x{start:X}", "end": f"0x{end:X}", "instructions": out}


def technology_value_operand_functions():
    """列出所有直接含 player-relative +0x224 的函式；語意留待外部證據審查。"""
    out = {}
    for fea in idautils.Functions():
        hits = []
        for ea in idautils.FuncItems(fea):
            line = (idc.generate_disasm_line(ea, 0) or "").lower()
            if "+224h]" in line or "+224h," in line:
                hits.append(insn(ea))
        if hits:
            out[f"0x{fea:X}"] = {
                "original_name": idc.get_name(fea) or "<unnamed>",
                "hits": hits,
                "callers": [insn(x) for x in idautils.CodeRefsTo(fea, 0)],
            }
    return out


def technology_value_topic_records():
    base = 0x17D904
    stride = 23
    rows = []
    for topic in range(83):
        ea = base + stride * topic
        rows.append({
            "topic": topic,
            "ea": f"0x{ea:X}",
            "bytes": (ida_bytes.get_bytes(ea, stride) or b"").hex(),
            "research_choice_ids_navigation_only": [ida_bytes.get_word(ea + off) for off in (6, 8)],
            "technology_value_slots": [ida_bytes.get_word(ea + off) for off in (10, 12, 14, 16)],
            "base_cost": ida_bytes.get_dword(ea + 18),
        })
    return {
        "base": f"0x{base:X}",
        "stride": stride,
        "count": len(rows),
        "records": rows,
    }


def main():
    ida_auto.auto_wait()
    report = {
        "schema": "moo2.ida.re-evidence.v1", "evidence_scope": "static_only", "mutation": "none",
        "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version(),
                 "script": "tools/ida/audit_history_score.py"},
        "input": {"database": ida_nalt.get_input_file_path(), "source": SOURCE,
                  "source_sha256": sha256(SOURCE), "processor": ida_ida.inf_get_procname(),
                  "min_ea": f"0x{ida_ida.inf_get_min_ea():X}", "max_ea": f"0x{ida_ida.inf_get_max_ea():X}"},
        "address_basis": "IDA linear; DOS/4GW LE object #1",
        "roots": {name: function(ea) for name, ea in ROOTS.items()},
        "windows": {name: linear_window(start, end) for name, (start, end) in WINDOWS.items()},
        "named_roots": {name: resolve_named(candidates) for name, candidates in NAMED_ROOTS.items()},
        "technology_value_operand_functions": technology_value_operand_functions(),
        "technology_value_topic_records": technology_value_topic_records(),
    }
    os.makedirs(os.path.dirname(OUT), exist_ok=True)
    with open(OUT, "w", encoding="utf-8") as f:
        json.dump(report, f, ensure_ascii=False, indent=2)
        f.write("\n")
    ida_pro.qexit(0)


if __name__ == "__main__":
    main()
