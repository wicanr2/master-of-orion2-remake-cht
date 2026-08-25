"""唯讀匯出 MOO2 研究產出、突破與完成回寫鏈。

保留原始函式名、IDA 線性位址、bytes、運算元、caller/callee；不修改或儲存 IDB。
"""
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

OUT = os.environ.get("MOO2_RE_OUT", "/out/research-breakthrough.json")
SOURCE = os.environ.get("MOO2_RE_SOURCE", ida_nalt.get_input_file_path())
ROOTS = {
    "raw_Next_Turn_Calc": 0x136B3,
    "raw_Colony_Research_Production": 0xDFF74,
    "raw_Colony_Base_Production_By_Job": 0xDE280,
    "raw_Colony_Base_Production_Unit": 0xDE22C,
    "raw_Colony_Production_Adjustment": 0xDDF2C,
    "raw_Colony_Planet_Special_Adjustment": 0xDD9F2,
    "raw_Check_For_Research_Breakthrough": 0xE44E0,
    "raw_Player_Research_Cost": 0xE1E96,
    "raw_Research_Breakthrough_Chance": 0xE1EC6,
    "raw_Research_Complete_Applications": 0xE4410,
    "raw_Research_Choose_Field": 0xE401D,
    "raw_Research_Grant_Application": 0xE4204,
    "raw_Recalc_Colony_After_Tech": 0xE1D59,
    "raw_Recalc_Player_After_Tech": 0xE2D09,
    "raw_Recalc_Player_Drive_Derived": 0x10034D,
    "raw_Get_FTL_Derived": 0x57597,
    "raw_AI_Research_Application_Select": 0xDC288,
    "raw_Player_Research_Application_Select_UI": 0x10DC12,
    "raw_Player_Research_Select_Entry_A": 0x10DB69,
    "raw_Player_Research_Select_Entry_B": 0x10DBCE,
    "raw_Player_Research_Application_Row": 0x10ED00,
    "raw_Player_Research_Application_Rows": 0x10E563,
    "raw_Next_Turn_Player_Loop": 0xE4F49,
}


def sha256(path):
    h = hashlib.sha256()
    with open(path, "rb") as f:
        for chunk in iter(lambda: f.read(1024 * 1024), b""):
            h.update(chunk)
    return h.hexdigest()


def insn(ea):
    size = idc.get_item_size(ea)
    raw = ida_bytes.get_bytes(ea, size) or b""
    return {
        "ea": f"0x{ea:X}",
        "bytes": raw.hex(),
        "mnem": idc.print_insn_mnem(ea),
        "op0": idc.print_operand(ea, 0),
        "op1": idc.print_operand(ea, 1),
        "line": idc.generate_disasm_line(ea, 0) or "<unavailable>",
    }


def function(ea):
    f = ida_funcs.get_func(ea)
    if not f:
        return {"requested": f"0x{ea:X}", "error": "no function"}
    callers = []
    for ref in idautils.CodeRefsTo(f.start_ea, 0):
        cf = ida_funcs.get_func(ref)
        callers.append({
            "instruction": insn(ref),
            "function_start": f"0x{cf.start_ea:X}" if cf else None,
            "function_name": idc.get_name(cf.start_ea) if cf else None,
        })
    callees = []
    for item in idautils.FuncItems(f.start_ea):
        if idc.print_insn_mnem(item).lower() == "call":
            targets = list(idautils.CodeRefsFrom(item, 0))
            callees.append({
                "instruction": insn(item),
                "targets": [{"ea": f"0x{x:X}", "name": idc.get_name(x) or "<unnamed>"} for x in targets],
            })
    return {
        "start": f"0x{f.start_ea:X}",
        "end": f"0x{f.end_ea:X}",
        "original_name": idc.get_name(f.start_ea) or "<unnamed>",
        "instructions": [insn(x) for x in idautils.FuncItems(f.start_ea)],
        "callers": callers,
        "callees": callees,
    }


def main():
    ida_auto.auto_wait()
    roots = {name: function(ea) for name, ea in ROOTS.items()}
    job_dispatch = []
    for mode in range(3):
        slot = 0x18355C + mode * 4
        target = ida_bytes.get_dword(slot)
        job_dispatch.append({
            "mode": mode,
            "slot": f"0x{slot:X}",
            "target": f"0x{target:X}",
            "function": function(target),
        })
    direct = {}
    for root in roots.values():
        for call in root.get("callees", []):
            for target in call["targets"]:
                ea = int(target["ea"], 16)
                direct[f"{target['name']}@0x{ea:X}"] = function(ea)
    field_hits = []
    for f_ea in idautils.Functions():
        for item in idautils.FuncItems(f_ea):
            line = idc.generate_disasm_line(item, 0) or ""
            if any(token in line for token in ("+322h", "+323h", "+8B4h", "+8B5h")):
                field_hits.append({
                    "function_start": f"0x{f_ea:X}",
                    "function_name": idc.get_name(f_ea) or "<unnamed>",
                    "instruction": insn(item),
                })
    report = {
        "schema": "moo2.ida.re-evidence.v1",
        "evidence_scope": "static_only",
        "mutation": "none",
        "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version(),
                 "script": "tools/ida/audit_research_breakthrough.py"},
        "input": {"database": ida_nalt.get_input_file_path(), "source": SOURCE,
                  "source_sha256": sha256(SOURCE), "processor": ida_ida.inf_get_procname(),
                  "min_ea": f"0x{ida_ida.inf_get_min_ea():X}",
                  "max_ea": f"0x{ida_ida.inf_get_max_ea():X}"},
        "address_basis": "IDA linear; DOS/4GW LE object #1",
        "roots": roots,
        "job_dispatch": job_dispatch,
        "research_runtime_field_hits": field_hits,
        "direct_callees": direct,
    }
    os.makedirs(os.path.dirname(OUT), exist_ok=True)
    with open(OUT, "w", encoding="utf-8") as f:
        json.dump(report, f, ensure_ascii=False, indent=2)
        f.write("\n")
    ida_pro.qexit(0)


if __name__ == "__main__":
    main()
