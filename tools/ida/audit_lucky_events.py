"""唯讀匯出 MOO2 Lucky 種族旗標與隨機事件排程／目標鏈。"""

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

OUT = os.environ.get("MOO2_RE_OUT", "/out/lucky-events.json")
SOURCE = os.environ.get("MOO2_RE_SOURCE", ida_nalt.get_input_file_path())
ROOTS = {
    "raw_Event_Record_Turn_Check": 0x201F9,
    "raw_Event_Record_Create_Check": 0x201A4,
    "raw_Event_Dispatcher": 0x21371,
    "raw_Event_Fleet_Eligibility": 0x2230A,
    "raw_Determine_Event_Target": 0x22D57,
    "raw_Event_Target_Eligibility": 0x230B6,
    "raw_Event_Target_Alternate": 0x22F5C,
    "raw_Lucky_Event_Gate_A": 0x24511,
    "raw_Lucky_Event_Gate_B": 0x245C4,
    "raw_Weighted_Choice": 0x586D4,
    "raw_Convert_Custom_Race_Flags": 0x5BC24,
    "raw_Game_Config_Init": 0x10E2F,
    "raw_Game_Config_Override": 0x127E1,
    "raw_Next_Turn_Calc": 0x136B3,
    "raw_Tech_Gate_A": 0xE408F,
    "raw_Tech_Gate_B": 0xE412B,
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
    return {"requested": f"0x{ea:X}", "start": f"0x{f.start_ea:X}", "end": f"0x{f.end_ea:X}",
            "original_name": idc.get_name(f.start_ea) or "<unnamed>",
            "instructions": [insn(x) for x in idautils.FuncItems(f.start_ea)],
            "callers": [insn(x) for x in idautils.CodeRefsTo(f.start_ea, 0)]}


def lucky_operand_hits():
    out = []
    for f_ea in idautils.Functions():
        f = ida_funcs.get_func(f_ea)
        hits = []
        for ea in idautils.FuncItems(f_ea):
            line = idc.generate_disasm_line(ea, 0) or ""
            if "+8B9h" in line:
                hits.append(insn(ea))
        if hits:
            out.append({"function_start": f"0x{f.start_ea:X}",
                        "function_name": idc.get_name(f.start_ea) or "<unnamed>", "hits": hits})
    return out


def raw(ea, size):
    return {"ea": f"0x{ea:X}", "size": size, "bytes": (ida_bytes.get_bytes(ea, size) or b"").hex()}


def data_refs(ea):
    out = []
    for ref in idautils.DataRefsTo(ea):
        f = ida_funcs.get_func(ref)
        out.append({"instruction": insn(ref),
                    "function_start": f"0x{f.start_ea:X}" if f else None,
                    "function_name": idc.get_name(f.start_ea) if f else None})
    return out


def main():
    ida_auto.auto_wait()
    report = {
        "schema": "moo2.ida.re-evidence.v1", "evidence_scope": "static_only", "mutation": "none",
        "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version(),
                 "script": "tools/ida/audit_lucky_events.py"},
        "input": {"database": ida_nalt.get_input_file_path(), "source": SOURCE,
                  "source_sha256": sha256(SOURCE), "processor": ida_ida.inf_get_procname(),
                  "min_ea": f"0x{ida_ida.inf_get_min_ea():X}", "max_ea": f"0x{ida_ida.inf_get_max_ea():X}"},
        "address_basis": "IDA linear; DOS/4GW LE object #1",
        "roots": {name: function(ea) for name, ea in ROOTS.items()},
        "lucky_operand_hits": lucky_operand_hits(),
        "tables": {"raw_event_record_slots": raw(0x19ABA5, 36 * 9),
                   "raw_event_good_flags_and_tail": raw(0x180E84, 36)},
        "global_refs": {"raw_flag_199BDE": data_refs(0x199BDE),
                        "raw_flag_199CAF": data_refs(0x199CAF)},
    }
    os.makedirs(os.path.dirname(OUT), exist_ok=True)
    with open(OUT, "w", encoding="utf-8") as f:
        json.dump(report, f, ensure_ascii=False, indent=2)
        f.write("\n")
    ida_pro.qexit(0)


if __name__ == "__main__":
    main()
