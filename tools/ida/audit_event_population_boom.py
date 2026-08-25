"""唯讀匯出事件 17 人口暴增的 record、持續期與殖民地成長消費端。"""

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

OUT = os.environ.get("MOO2_RE_OUT", "/out/event-population-boom.json")
SOURCE = os.environ.get("MOO2_RE_SOURCE", ida_nalt.get_input_file_path())
ROOTS = {
    "raw_Event_Effect_Consumer": 0x206A2,
    "raw_Event_Dispatcher": 0x21371,
    "raw_Determine_Event": 0x2230A,
    "raw_Find_Active_Event_For_Colony": 0x230B6,
    "raw_Colony_Event_Flag_A": 0x234B8,
    "raw_Colony_Event_Flag_B": 0x23509,
    "raw_Event_Colony_Growth_Modifier": 0x23DFE,
    "raw_Event_On_Colony_Query": 0x23E60,
    "raw_Colony_Max_Population": 0xE0B4F,
    "raw_Persistent_Colony_Event_Turn": 0x24069,
    "raw_Persistent_Colony_Event_Consumer": 0x2434D,
    "raw_Recalculate_Colony": 0xE2A70,
    "raw_Colony_Precalculation": 0xE1D59,
    "raw_Recalculate_Colony_Growth_Rates": 0xDF8F0,
    "raw_Colony_Population_Growth_Calculation": 0xDFDC6,
    "raw_Recompute_Race_Growth": 0xE1839,
    "raw_Apply_Colony_Pop_Growth": 0xE2DCA,
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
    return {"requested": f"0x{ea:X}", "start": f"0x{f.start_ea:X}",
            "end": f"0x{f.end_ea:X}", "original_name": idc.get_name(f.start_ea) or "<unnamed>",
            "instructions": [insn(x) for x in items],
            "callers": [insn(x) for x in idautils.CodeRefsTo(f.start_ea, 0)],
            "callees": [insn(x) for x in items if idc.print_insn_mnem(x).lower() == "call"]}


def data_refs(ea):
    return [insn(x) for x in idautils.DataRefsTo(ea)]


def main():
    ida_auto.auto_wait()
    report = {"schema": "moo2.ida.re-evidence.v1", "evidence_scope": "static_only", "mutation": "none",
              "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version(),
                       "script": "tools/ida/audit_event_population_boom.py"},
              "input": {"database": ida_nalt.get_input_file_path(), "source": SOURCE,
                        "source_sha256": sha256(SOURCE), "processor": ida_ida.inf_get_procname(),
                        "min_ea": f"0x{ida_ida.inf_get_min_ea():X}",
                        "max_ea": f"0x{ida_ida.inf_get_max_ea():X}"},
              "address_basis": "IDA linear; DOS/4GW LE object #1",
              "roots": {name: function(ea) for name, ea in ROOTS.items()},
              "raw_global_refs": {
                  "raw_19AC35": data_refs(0x19AC35),
                  "raw_19AC37": data_refs(0x19AC37),
                  "raw_19AC3E": data_refs(0x19AC3E),
                  "raw_19AC40": data_refs(0x19AC40),
              }}
    os.makedirs(os.path.dirname(OUT), exist_ok=True)
    with open(OUT, "w", encoding="utf-8") as f:
        json.dump(report, f, ensure_ascii=False, indent=2)
        f.write("\n")
    ida_pro.qexit(0)


if __name__ == "__main__":
    main()
