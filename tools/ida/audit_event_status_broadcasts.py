"""唯讀匯出事件 29..35 狀態播報 record 的全部直接讀寫函式與新聞 dispatcher。"""

import json
import os

import ida_auto
import ida_funcs
import ida_ida
import ida_kernwin
import ida_nalt
import ida_pro
import idautils

import audit_event_warp_beast as common

OUT = os.environ.get("MOO2_RE_OUT", "/out/event-status-broadcasts.json")
SOURCE = os.environ.get("MOO2_RE_SOURCE", ida_nalt.get_input_file_path())
ROOTS = {
    "raw_Event_Effect_Consumer": 0x206A2,
    "raw_Event_State_Dispatcher": 0x21371,
    "raw_Determine_Event": 0x2230A,
    "raw_Setup_Next_Event": 0x201A4,
    "raw_Event29_Caller": 0xE4F34,
    "raw_Event32_Caller": 0xFFE31,
    "raw_Event34_Caller": 0x2686B,
}
EVENT_RECORD_BASE = 0x19ABA4
EVENT_STRIDE = 9


def main():
    ida_auto.auto_wait()
    records, ref_functions = {}, {}
    for event_id in range(29, 36):
        start = EVENT_RECORD_BASE + event_id * EVENT_STRIDE
        refs_by_field = {}
        for field in range(start, start + EVENT_STRIDE):
            refs = sorted(set(idautils.DataRefsTo(field)))
            refs_by_field[f"0x{field:X}"] = [common.insn(ea) for ea in refs]
            for ea in refs:
                f = ida_funcs.get_func(ea)
                if f:
                    ref_functions[f.start_ea] = common.function(f.start_ea)
        records[str(event_id)] = {"range": f"0x{start:X}..0x{start + EVENT_STRIDE - 1:X}",
                                  "direct_refs": refs_by_field}
    report = {
        "schema": "moo2.ida.re-evidence.v1", "evidence_scope": "static_only", "mutation": "none",
        "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version(),
                 "script": "tools/ida/audit_event_status_broadcasts.py"},
        "input": {"database": ida_nalt.get_input_file_path(), "source": SOURCE,
                  "source_sha256": common.sha256(SOURCE), "processor": ida_ida.inf_get_procname(),
                  "min_ea": f"0x{ida_ida.inf_get_min_ea():X}", "max_ea": f"0x{ida_ida.inf_get_max_ea():X}"},
        "address_basis": "IDA linear; DOS/4GW LE object #1",
        "event_records": records,
        "roots": {name: common.function(ea) for name, ea in ROOTS.items()},
        "direct_ref_functions": {f"0x{ea:X}": value for ea, value in sorted(ref_functions.items())},
    }
    os.makedirs(os.path.dirname(OUT), exist_ok=True)
    with open(OUT, "w", encoding="utf-8") as f:
        json.dump(report, f, ensure_ascii=False, indent=2)
        f.write("\n")
    ida_pro.qexit(0)


if __name__ == "__main__":
    main()
