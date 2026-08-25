"""唯讀匯出 MOO2 一般隨機事件排程、初始化與事件 ID 抽樣證據。"""

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

OUT = os.environ.get("MOO2_RE_OUT", "/out/event-schedule.json")
SOURCE = os.environ.get("MOO2_RE_SOURCE", ida_nalt.get_input_file_path())
ROOTS = {
    "raw_Event_Record_Create_Check": 0x201A4,
    "raw_Event_Record_Turn_Check": 0x201F9,
    "raw_Event_Turn_Orchestrator": 0x2023F,
    "raw_Event_Record_Maintenance": 0x20316,
    "raw_Determine_Event": 0x2230A,
    "raw_Event_Victim_Chooser": 0x22D57,
    "raw_Event_Victim_Eligibility": 0x230B6,
    "raw_Event_Alternate_Chooser": 0x22F5C,
    "raw_Weighted_Choice": 0x586D4,
    "raw_Player_A6_Writer": 0xE2710,
    "raw_History_Consumer": 0x10208A,
    "raw_Lucky_Selector": 0x24511,
    "raw_Lucky_Counter": 0x245C4,
    "raw_Next_Turn_Calc": 0x136B3,
    "raw_New_Game_Event_Init_Caller": 0xC5F57,
    "raw_Random": 0x1247A0,
}
GLOBALS = {
    "raw_event_last_date": 0x19ACE8,
    "raw_event_attempt_counter": 0x19ACEC,
    "raw_random_events_enabled": 0x199BDE,
    "raw_difficulty": 0x199CB0,
    "raw_stardate": 0x192FD8,
    "raw_event_good_flags": 0x180E84,
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
    return {
        "requested": f"0x{ea:X}",
        "start": f"0x{f.start_ea:X}",
        "end": f"0x{f.end_ea:X}",
        "original_name": idc.get_name(f.start_ea) or "<unnamed>",
        "instructions": [insn(x) for x in idautils.FuncItems(f.start_ea)],
        "callers": [insn(x) for x in idautils.CodeRefsTo(f.start_ea, 0)],
    }


def refs(ea):
    out = []
    for ref in idautils.DataRefsTo(ea):
        f = ida_funcs.get_func(ref)
        out.append({
            "instruction": insn(ref),
            "function_start": f"0x{f.start_ea:X}" if f else None,
            "function_name": idc.get_name(f.start_ea) if f else None,
        })
    return out


def operand_hits(needle):
    out = []
    for f_ea in idautils.Functions():
        hits = []
        for ea in idautils.FuncItems(f_ea):
            line = idc.generate_disasm_line(ea, 0) or ""
            if needle in line:
                hits.append(insn(ea))
        if hits:
            f = ida_funcs.get_func(f_ea)
            out.append({"function_start": f"0x{f.start_ea:X}",
                        "function_name": idc.get_name(f.start_ea) or "<unnamed>",
                        "hits": hits})
    return out


def main():
    ida_auto.auto_wait()
    report = {
        "schema": "moo2.ida.re-evidence.v1",
        "evidence_scope": "static_only",
        "mutation": "none",
        "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version(),
                 "script": "tools/ida/audit_event_schedule.py"},
        "input": {"database": ida_nalt.get_input_file_path(), "source": SOURCE,
                  "source_sha256": sha256(SOURCE), "processor": ida_ida.inf_get_procname(),
                  "min_ea": f"0x{ida_ida.inf_get_min_ea():X}",
                  "max_ea": f"0x{ida_ida.inf_get_max_ea():X}"},
        "address_basis": "IDA linear; DOS/4GW LE object #1",
        "roots": {name: function(ea) for name, ea in ROOTS.items()},
        "global_refs": {name: refs(ea) for name, ea in GLOBALS.items()},
        "relative_operand_hits": {"raw_player_plus_A6": operand_hits("+0A6h")},
        "tables": {"raw_event_good_flags": {
            "ea": "0x180E84", "size": 29,
            "bytes": (ida_bytes.get_bytes(0x180E84, 29) or b"").hex()}},
    }
    os.makedirs(os.path.dirname(OUT), exist_ok=True)
    with open(OUT, "w", encoding="utf-8") as f:
        json.dump(report, f, ensure_ascii=False, indent=2)
        f.write("\n")
    ida_pro.qexit(0)


if __name__ == "__main__":
    main()
