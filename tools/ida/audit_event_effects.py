"""唯讀匯出 MOO2 事件模組函式、事件 record 參照與國庫欄位候選消費端。"""

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

OUT = os.environ.get("MOO2_RE_OUT", "/out/event-effects.json")
SOURCE = os.environ.get("MOO2_RE_SOURCE", ida_nalt.get_input_file_path())
EVENT_START = 0x20000
EVENT_END = 0x24A00
ROOTS = {
	"raw_Event_Effect_Consumer": 0x206A2,
    "raw_Event_Record_Maintenance": 0x2027E,
    "raw_Event_Dispatcher": 0x21371,
    "raw_Determine_Event": 0x2230A,
    "raw_Event_Colony_Chooser": 0x23BEC,
    "raw_Event_Colony_Value": 0x23B64,
    "raw_Event_Target_Relation": 0x23B7D,
    "raw_Event_Ship_Chooser": 0x23CED,
    "raw_Event_Colony_Chooser_A": 0x23DA0,
    "raw_Event_Colony_Chooser_B": 0x23D44,
    "raw_Event_Colony_Filter": 0x23E60,
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


def event_module_inventory():
    out = []
    for ea in idautils.Functions(EVENT_START, EVENT_END):
        f = ida_funcs.get_func(ea)
        out.append({"start": f"0x{f.start_ea:X}", "end": f"0x{f.end_ea:X}",
                    "original_name": idc.get_name(f.start_ea) or "<unnamed>",
                    "calls": [insn(x) for x in idautils.FuncItems(f.start_ea)
                              if idc.print_insn_mnem(x).lower() == "call"]})
    return out


def operand_hits(needles, start=0x10000, end=0x1D5CD0):
    out = []
    for f_ea in idautils.Functions(start, end):
        hits = []
        for ea in idautils.FuncItems(f_ea):
            line = idc.generate_disasm_line(ea, 0) or ""
            if any(needle in line for needle in needles):
                hits.append(insn(ea))
        if hits:
            f = ida_funcs.get_func(f_ea)
            out.append({"function_start": f"0x{f.start_ea:X}",
                        "function_name": idc.get_name(f.start_ea) or "<unnamed>", "hits": hits})
    return out


def main():
    ida_auto.auto_wait()
    report = {
        "schema": "moo2.ida.re-evidence.v1", "evidence_scope": "static_only", "mutation": "none",
        "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version(),
                 "script": "tools/ida/audit_event_effects.py"},
        "input": {"database": ida_nalt.get_input_file_path(), "source": SOURCE,
                  "source_sha256": sha256(SOURCE), "processor": ida_ida.inf_get_procname(),
                  "min_ea": f"0x{ida_ida.inf_get_min_ea():X}", "max_ea": f"0x{ida_ida.inf_get_max_ea():X}"},
        "address_basis": "IDA linear; DOS/4GW LE object #1",
        "roots": {name: function(ea) for name, ea in ROOTS.items()},
        "event_module_inventory": event_module_inventory(),
        "event_record_operand_hits": operand_hits(["19ABA4", "19ABA5", "19ABA7", "19ABA9", "19ABAB"]),
        "player_treasury_operand_hits": operand_hits(["+32h"]),
    }
    os.makedirs(os.path.dirname(OUT), exist_ok=True)
    with open(OUT, "w", encoding="utf-8") as f:
        json.dump(report, f, ensure_ascii=False, indent=2)
        f.write("\n")
    ida_pro.qexit(0)


if __name__ == "__main__":
    main()
