"""唯讀匯出 MOO2 貿易／研究協議的建立、每回合推進與回寫鏈。"""

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


OUT = os.environ.get("MOO2_RE_OUT", "/out/trade-research-agreements.json")
SOURCE = os.environ.get("MOO2_RE_SOURCE", ida_nalt.get_input_file_path())
ROOTS = {
    "raw_Process_Agreements_worklist_anchor": 0x101E77,
    "raw_Start_Trade_Agreement": 0x101EE3,
    "raw_Advance_One_Trade_Agreement": 0x101D53,
    "raw_Advance_One_Research_Agreement": 0x101DE5,
    "raw_Agreement_Postprocess": 0x101A42,
    "raw_Start_Research_Agreement": 0x101F82,
    "raw_Start_Agreement_UI_Caller": 0x5223F,
    "raw_Trade_Agreement_Base": 0x101B3C,
    "raw_Base_Trade_Agreement_Goal": 0x101BA4,
    "raw_Trade_Agreement_Response": 0x101C93,
    "raw_Research_Agreement_Response": 0x101CC5,
    "raw_Start_Trade_Treaty": 0x5232E,
    "raw_Advance_Trade_Value": 0x524C3,
    "raw_Trade_Relation_Delta": 0x524FB,
    "raw_Next_Turn_Calc": 0x136B3,
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


def data_refs(ea, size):
    out = []
    for address in range(ea, ea + size):
        for ref in idautils.XrefsTo(address, 0):
            f = ida_funcs.get_func(ref.frm)
            out.append({
                "target": f"0x{address:X}",
                "from": insn(ref.frm),
                "function_start": f"0x{f.start_ea:X}" if f else None,
                "function_name": idc.get_name(f.start_ea) if f else None,
            })
    return out


def agreement_offset_hits():
    """找外交 record 目前值／目標／旗標位移的直接顯示端。"""
    needles = ("+5A4h", "+5B4h", "+5C6h", "+5D6h", "+62Fh", "+637h")
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
            "script": "tools/ida/audit_trade_research_agreements.py",
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
        "raw_tables": {
            "trade_agreement_values_0x18105C": {
                "ea": "0x18105C",
                "bytes": (ida_bytes.get_bytes(0x18105C, 0x20) or b"").hex(),
                "dwords": [idc.get_wide_dword(0x18105C + i * 4) for i in range(8)],
                "xrefs": data_refs(0x18105C, 0x20),
            },
            "trade_event_values_0x181070": {
                "ea": "0x181070",
                "bytes": (ida_bytes.get_bytes(0x181070, 0x20) or b"").hex(),
                "dwords": [idc.get_wide_dword(0x181070 + i * 4) for i in range(8)],
                "xrefs": data_refs(0x181070, 0x20),
            },
        },
        "agreement_offset_hits": agreement_offset_hits(),
    }
    os.makedirs(os.path.dirname(OUT), exist_ok=True)
    with open(OUT, "w", encoding="utf-8") as f:
        json.dump(report, f, ensure_ascii=False, indent=2)
        f.write("\n")
    ida_pro.qexit(0)


if __name__ == "__main__":
    main()
