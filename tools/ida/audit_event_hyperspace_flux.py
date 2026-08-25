"""唯讀匯出事件 9 超空間亂流的建立、維護、展示及航行消費資料流。"""

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

OUT = os.environ.get("MOO2_RE_OUT", "/out/event-hyperspace-flux.json")
SOURCE = os.environ.get("MOO2_RE_SOURCE", ida_nalt.get_input_file_path())
ROOTS = {
    "raw_Next_Turn_Calc": 0x136B3,
    "raw_Event_Record_Maintenance": 0x2027E,
    "raw_Event_Effect_Consumer": 0x206A2,
    "raw_Event_State_Dispatcher": 0x21371,
    "raw_Determine_Event": 0x2230A,
    "raw_Move_All_Ships_Toward_Stars": 0xFFEEA,
    "raw_Event9_Active_Query": 0x233FA,
    "raw_Event9_Caller_1FDCA": 0x1FDCA,
    "raw_Event9_Caller_25E2D": 0x25E2D,
    "raw_Event9_Caller_87CEC": 0x87CEC,
    "raw_Event9_Caller_9A09E": 0x9A09E,
    "raw_Event9_Caller_9D66F": 0x9D66F,
    "raw_Event9_Caller_9D6CB": 0x9D6CB,
    "raw_Event9_Caller_DBB7F": 0xDBB7F,
    "raw_Event9_Caller_DBC8A": 0xDBC8A,
    "raw_Event9_Caller_DBCF8": 0xDBCF8,
    "raw_Event9_Caller_E6D64": 0xE6D64,
    "raw_Event9_Caller_EF836": 0xEF836,
    "raw_Event9_Caller_FF861": 0xFF861,
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
    return {"requested": f"0x{ea:X}", "start": f"0x{f.start_ea:X}", "end": f"0x{f.end_ea:X}",
            "original_name": idc.get_name(f.start_ea) or "<unnamed>",
            "instructions": [insn(x) for x in items],
            "callers": [insn(x) for x in idautils.CodeRefsTo(f.start_ea, 0)],
            "callees": [insn(x) for x in items if idc.print_insn_mnem(x).lower() == "call"]}


def named_flux_functions():
    out = []
    for ea in idautils.Functions():
        name = idc.get_name(ea) or ""
        if "flux" in name.lower() or "hyperspace" in name.lower():
            out.append(function(ea))
    return out


def event9_operand_refs():
    out = []
    for ea in idautils.Heads(ida_ida.inf_get_min_ea(), ida_ida.inf_get_max_ea()):
        if not ida_bytes.is_code(ida_bytes.get_flags(ea)):
            continue
        line = idc.generate_disasm_line(ea, 0) or ""
        if any(token in line.upper() for token in ("19ABF1", "19ABF2", "19ABF3", "19ABF4",
                                                    "19ABF5", "19ABF6", "19ABF7", "19ABF8", "19ABF9")):
            out.append(insn(ea))
    return out


def main():
    ida_auto.auto_wait()
    report = {"schema": "moo2.ida.re-evidence.v1", "evidence_scope": "static_only", "mutation": "none",
              "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version(),
                       "script": "tools/ida/audit_event_hyperspace_flux.py"},
              "input": {"database": ida_nalt.get_input_file_path(), "source": SOURCE,
                        "source_sha256": sha256(SOURCE), "processor": ida_ida.inf_get_procname(),
                        "min_ea": f"0x{ida_ida.inf_get_min_ea():X}", "max_ea": f"0x{ida_ida.inf_get_max_ea():X}"},
              "address_basis": "IDA linear; DOS/4GW LE object #1",
              "roots": {name: function(ea) for name, ea in ROOTS.items()},
              "named_flux_functions": named_flux_functions(),
              "event9_operand_refs": event9_operand_refs()}
    os.makedirs(os.path.dirname(OUT), exist_ok=True)
    with open(OUT, "w", encoding="utf-8") as f:
        json.dump(report, f, ensure_ascii=False, indent=2)
        f.write("\n")
    ida_pro.qexit(0)


if __name__ == "__main__":
    main()
