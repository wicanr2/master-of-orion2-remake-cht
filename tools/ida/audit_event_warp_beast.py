"""唯讀匯出事件 26（Warp Beast）的 record、選艦、逐回合 consumer 與艦艇移除鏈。"""

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

OUT = os.environ.get("MOO2_RE_OUT", "/out/event-warp-beast.json")
SOURCE = os.environ.get("MOO2_RE_SOURCE", ida_nalt.get_input_file_path())
ROOTS = {
    "raw_Event_Effect_Consumer": 0x206A2,
    "raw_Event_State_Dispatcher": 0x21371,
    "raw_Determine_Event": 0x2230A,
    "raw_Event_Target_Ship": 0x23CED,
    "raw_Event26_Empire_Helper": 0x100618,
    "raw_Event26_Ship_Remove_Helper": 0x941C6,
    "raw_Event26_Ship_Leader_Helper": 0x944A3,
    "raw_Event26_Post_Remove_Helper": 0x8A4C4,
    "raw_Event26_Message_Helper": 0xEF629,
    "raw_Move_All_Ships_Toward_Stars": 0xFFEEA,
}
EVENT26_FIELDS = range(0x19AC8E, 0x19AC97)
EVENT_RECORD_BASE_FIELDS = range(0x19ABA4, 0x19ABAD)


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
    return {"requested": f"0x{ea:X}", "start": f"0x{f.start_ea:X}", "end": f"0x{f.end_ea:X}",
            "original_name": idc.get_name(f.start_ea) or "<unnamed>",
            "instructions": [insn(x) for x in items],
            "callers": [insn(x) for x in idautils.CodeRefsTo(f.start_ea, 0)],
            "callees": [insn(x) for x in items if idc.print_insn_mnem(x).lower() == "call"]}


def main():
    ida_auto.auto_wait()
    direct_refs = {}
    ref_functions = {}
    for field in EVENT26_FIELDS:
        refs = sorted(set(idautils.DataRefsTo(field)))
        direct_refs[f"0x{field:X}"] = [insn(ea) for ea in refs]
        for ea in refs:
            f = ida_funcs.get_func(ea)
            if f:
                ref_functions[f.start_ea] = function(f.start_ea)
    indexed_refs = {}
    for field in EVENT_RECORD_BASE_FIELDS:
        refs = sorted(set(idautils.DataRefsTo(field)))
        indexed_refs[f"0x{field:X}"] = [insn(ea) for ea in refs]
        for ea in refs:
            f = ida_funcs.get_func(ea)
            if f:
                ref_functions[f.start_ea] = function(f.start_ea)
    report = {
        "schema": "moo2.ida.re-evidence.v1", "evidence_scope": "static_only", "mutation": "none",
        "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version(),
                 "script": "tools/ida/audit_event_warp_beast.py"},
        "input": {"database": ida_nalt.get_input_file_path(), "source": SOURCE,
                  "source_sha256": sha256(SOURCE), "processor": ida_ida.inf_get_procname(),
                  "min_ea": f"0x{ida_ida.inf_get_min_ea():X}", "max_ea": f"0x{ida_ida.inf_get_max_ea():X}"},
        "address_basis": "IDA linear; DOS/4GW LE object #1",
        "event26_record": {"slot": 26, "stride": 9, "base": "0x19ABA4",
                           "range": "0x19AC8E..0x19AC96", "direct_refs": direct_refs,
                           "indexed_base_refs": indexed_refs},
        "roots": {name: function(ea) for name, ea in ROOTS.items()},
        "direct_ref_functions": {f"0x{ea:X}": value for ea, value in sorted(ref_functions.items())},
    }
    os.makedirs(os.path.dirname(OUT), exist_ok=True)
    with open(OUT, "w", encoding="utf-8") as f:
        json.dump(report, f, ensure_ascii=False, indent=2)
        f.write("\n")
    ida_pro.qexit(0)


if __name__ == "__main__":
    main()
