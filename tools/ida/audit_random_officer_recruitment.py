"""唯讀匯出 MOO2 隨機領袖招募、候選篩選與回合 caller 證據。"""

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


OUT = os.environ.get("MOO2_RE_OUT", "/out/random-officer-recruitment.json")
SOURCE = os.environ.get("MOO2_RE_SOURCE", ida_nalt.get_input_file_path())
ROOTS = {
    "raw_Chance_To_Hire_Hero": 0x979A0,
    "raw_Random_Officer_Check_worklist_anchor": 0x97A66,
    "raw_Random_Officer_Check_late_oracle_anchor": 0x97AD4,
    "raw_Generate_Random_Officer": 0x97B2D,
    "raw_Leader_Available_For_Hire": 0x97C64,
    "raw_Set_Officer_To_Player": 0x97287,
    "raw_Officer_Status_Ok": 0x9776C,
    "raw_Officer_Count_Other_Type": 0x977AF,
    "raw_Officer_Recruit_Chance": 0x9781D,
    "raw_Officer_Candidate_Eligible": 0x97C2D,
    "raw_Stardate_Or_Turn_Helper": 0x64395,
    "raw_Officer_Experience_Level_Helper": 0x93D4B,
    "raw_Unowned_Officer_Level_Helper": 0x94951,
    "raw_Init_Officers": 0x1307F,
    "raw_Next_Turn_Calc": 0x136B3,
    "raw_Check_For_Officer_Level": 0x92FDA,
    "raw_Decrement_Officer_ETA_worklist_anchor": 0x934CF,
    "raw_Decrement_Officer_ETA_late_oracle_anchor": 0x93528,
    "raw_Do_AI_Leaders": 0xD7439,
}

DATA_POINTS = {
    "raw_recruit_skill_mask_a": (0x17D2CD, 4),
    "raw_recruit_skill_mask_b": (0x17D2DF, 4),
    "raw_candidate_prefix_modifier": (0x199998, 2),
    "raw_recruit_mode_flag": (0x199CB0, 1),
}


def sha256(path):
    h = hashlib.sha256()
    with open(path, "rb") as f:
        for chunk in iter(lambda: f.read(1024 * 1024), b""):
            h.update(chunk)
    return h.hexdigest()


def insn(ea):
    size = idc.get_item_size(ea)
    return {
        "ea": f"0x{ea:X}",
        "bytes": (ida_bytes.get_bytes(ea, size) or b"").hex(),
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
            callees.append({
                "instruction": insn(item),
                "targets": [
                    {"ea": f"0x{x:X}", "name": idc.get_name(x) or "<unnamed>"}
                    for x in idautils.CodeRefsFrom(item, 0)
                ],
            })
    return {
        "requested": f"0x{ea:X}",
        "start": f"0x{f.start_ea:X}",
        "end": f"0x{f.end_ea:X}",
        "original_name": idc.get_name(f.start_ea) or "<unnamed>",
        "instructions": [insn(x) for x in idautils.FuncItems(f.start_ea)],
        "callers": callers,
        "callees": callees,
    }


def officer_offset_hits():
    """列出 0x43-byte officer record 關鍵欄位的直接顯示端。"""
    needles = tuple(f"+{offset:X}h" for offset in (0x22, 0x23, 0x24, 0x34, 0x35, 0x36, 0x37))
    out = []
    for f_ea in idautils.Functions():
        f = ida_funcs.get_func(f_ea)
        if not f:
            continue
        hits = []
        for ea in idautils.FuncItems(f.start_ea):
            line = idc.generate_disasm_line(ea, 0) or ""
            if any(needle in line for needle in needles):
                hits.append(insn(ea))
        if hits:
            out.append({
                "start": f"0x{f.start_ea:X}",
                "end": f"0x{f.end_ea:X}",
                "original_name": idc.get_name(f.start_ea) or "<unnamed>",
                "hits": hits,
            })
    return out


def main():
    ida_auto.auto_wait()
    report = {
        "schema": "moo2.ida.re-evidence.v1",
        "evidence_scope": "static_only",
        "mutation": "none",
        "tool": {
            "name": "IDA Pro",
            "version": ida_kernwin.get_kernel_version(),
            "script": "tools/ida/audit_random_officer_recruitment.py",
        },
        "input": {
            "database": ida_nalt.get_input_file_path(),
            "source": SOURCE,
            "source_sha256": sha256(SOURCE),
            "processor": ida_ida.inf_get_procname(),
            "min_ea": f"0x{ida_ida.inf_get_min_ea():X}",
            "max_ea": f"0x{ida_ida.inf_get_max_ea():X}",
        },
        "address_basis": "IDA linear; DOS/4GW LE object #1",
        "roots": {name: function(ea) for name, ea in ROOTS.items()},
        "data_points": {
            name: {
                "ea": f"0x{ea:X}",
                "size": size,
                "bytes": (ida_bytes.get_bytes(ea, size) or b"").hex(),
                "value_le": int.from_bytes(ida_bytes.get_bytes(ea, size) or b"\0" * size, "little"),
                "original_name": idc.get_name(ea) or "<unnamed>",
                "data_refs_to": [f"0x{x:X}" for x in idautils.DataRefsTo(ea)],
            }
            for name, (ea, size) in DATA_POINTS.items()
        },
        "officer_offset_hits": officer_offset_hits(),
    }
    os.makedirs(os.path.dirname(OUT), exist_ok=True)
    with open(OUT, "w", encoding="utf-8") as f:
        json.dump(report, f, ensure_ascii=False, indent=2)
        f.write("\n")
    ida_pro.qexit(0)


if __name__ == "__main__":
    main()
