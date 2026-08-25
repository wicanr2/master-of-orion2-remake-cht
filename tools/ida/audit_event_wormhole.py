"""唯讀匯出事件 28（Wormhole）的候選 helper、record 與航行寫回。"""

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

OUT = os.environ.get("MOO2_RE_OUT", "/out/random-event-wormhole.json")
SOURCE = os.environ.get("MOO2_RE_SOURCE", ida_nalt.get_input_file_path())
ROOTS = {
    "raw_Determine_Event": 0x2230A,
    "raw_Event28_Apply": 0x100519,
    "raw_Event28_Move_Ship_Helper": 0xFFDDA,
    "raw_Event_State_Dispatcher": 0x21371,
    "raw_Move_All_Ships_Toward_Stars": 0xFFEEA,
}
EVENT28_FIELDS = range(0x19ACA0, 0x19ACA9)
EVENT_RECORD_BASE_FIELDS = range(0x19ABA4, 0x19ABAD)


def main():
    ida_auto.auto_wait()
    direct_refs, indexed_refs, ref_functions = {}, {}, {}
    for field in EVENT28_FIELDS:
        refs = sorted(set(idautils.DataRefsTo(field)))
        direct_refs[f"0x{field:X}"] = [common.insn(ea) for ea in refs]
        for ea in refs:
            f = ida_funcs.get_func(ea)
            if f:
                ref_functions[f.start_ea] = common.function(f.start_ea)
    for field in EVENT_RECORD_BASE_FIELDS:
        refs = sorted(set(idautils.DataRefsTo(field)))
        indexed_refs[f"0x{field:X}"] = [common.insn(ea) for ea in refs]
        for ea in refs:
            f = ida_funcs.get_func(ea)
            if f:
                ref_functions[f.start_ea] = common.function(f.start_ea)
    report = {
        "schema": "moo2.ida.re-evidence.v1", "evidence_scope": "static_only", "mutation": "none",
        "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version(),
                 "script": "tools/ida/audit_event_wormhole.py"},
        "input": {"database": ida_nalt.get_input_file_path(), "source": SOURCE,
                  "source_sha256": common.sha256(SOURCE), "processor": ida_ida.inf_get_procname(),
                  "min_ea": f"0x{ida_ida.inf_get_min_ea():X}", "max_ea": f"0x{ida_ida.inf_get_max_ea():X}"},
        "address_basis": "IDA linear; DOS/4GW LE object #1",
        "event28_record": {"slot": 28, "stride": 9, "base": "0x19ABA4",
                           "range": "0x19ACA0..0x19ACA8", "direct_refs": direct_refs,
                           "indexed_base_refs": indexed_refs},
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
