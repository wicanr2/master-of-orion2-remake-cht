"""唯讀匯出事件 34 帝國投降判定、資產轉移與新聞 record 的靜態證據。"""

import json
import os

import ida_auto
import ida_funcs
import ida_ida
import ida_kernwin
import ida_nalt
import ida_pro
import idautils
import idc

import audit_event_warp_beast as common

OUT = os.environ.get("MOO2_RE_OUT", "/out/empire-surrender.json")
SOURCE = os.environ.get("MOO2_RE_SOURCE", ida_nalt.get_input_file_path())
ROOTS = {
    "raw_Surrender_Diplomacy_Caller": 0x2670A,
    "raw_Surrender_Pretransfer_Helper": 0x27A3D,
    "raw_Surrender_Asset_Transfer": 0xE4D06,
    "raw_Event34_Setter": 0x233D2,
    "raw_Event34_News_Callsite": 0x2686B,
    "raw_Surrender_Callsite_Council": 0x16827,
    "raw_Surrender_Callsite_Other": 0x1D17C,
    "raw_Empire_Elimination_Cleanup": 0xE4EB3,
    "raw_Surrender_Deferred_Consumer": 0xE4DC9,
    "raw_Surrender_Transfer_One_Empire": 0xE4B5F,
    "raw_Surrender_Colony_Transfer": 0xE481F,
    "raw_Surrender_Ship_Helper": 0x8A4C4,
    "raw_Surrender_Spy_Getter": 0x1026CF,
    "raw_Surrender_Spy_Setter": 0x10278D,
}
EVENT34_FIELDS = range(0x19ACD6, 0x19ACDF)


def direct_callee_starts(ea):
    f = ida_funcs.get_func(ea)
    if not f:
        return []
    result = set()
    for insn_ea in idautils.FuncItems(f.start_ea):
        if idc.print_insn_mnem(insn_ea).lower() != "call":
            continue
        for target in idautils.CodeRefsFrom(insn_ea, 0):
            target_func = ida_funcs.get_func(target)
            if target_func:
                result.add(target_func.start_ea)
    return sorted(result)


def main():
    ida_auto.auto_wait()
    functions = {}
    for root in ROOTS.values():
        f = ida_funcs.get_func(root)
        if not f:
            continue
        functions[f.start_ea] = common.function(f.start_ea)
        for callee in direct_callee_starts(f.start_ea):
            functions.setdefault(callee, common.function(callee))

    field_refs = {}
    for field in EVENT34_FIELDS:
        refs = sorted(set(idautils.DataRefsTo(field)))
        field_refs[f"0x{field:X}"] = [common.insn(ea) for ea in refs]
        for ref in refs:
            f = ida_funcs.get_func(ref)
            if f:
                functions.setdefault(f.start_ea, common.function(f.start_ea))

    surrender_field_operands = []
    for seg_start in idautils.Segments():
        for ea in idautils.Heads(seg_start, idc.get_segm_end(seg_start)):
            operands = " ".join(idc.print_operand(ea, n) for n in range(2))
            if "0E72h" not in operands and "+E72h" not in operands:
                continue
            surrender_field_operands.append(common.insn(ea))
            f = ida_funcs.get_func(ea)
            if f:
                functions.setdefault(f.start_ea, common.function(f.start_ea))

    report = {
        "schema": "moo2.ida.re-evidence.v1",
        "evidence_scope": "static_only",
        "mutation": "none",
        "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version(),
                 "script": "tools/ida/audit_empire_surrender.py"},
        "input": {"database": ida_nalt.get_input_file_path(), "source": SOURCE,
                  "source_sha256": common.sha256(SOURCE), "processor": ida_ida.inf_get_procname(),
                  "min_ea": f"0x{ida_ida.inf_get_min_ea():X}",
                  "max_ea": f"0x{ida_ida.inf_get_max_ea():X}"},
        "address_basis": "IDA linear; DOS/4GW LE object #1",
        "event34_record": {"range": "0x19ACD6..0x19ACDE", "direct_refs": field_refs},
        "player_surrender_field_operands": surrender_field_operands,
        "roots": {name: common.function(ea) for name, ea in ROOTS.items()},
        "functions": {f"0x{ea:X}": value for ea, value in sorted(functions.items())},
    }
    os.makedirs(os.path.dirname(OUT), exist_ok=True)
    with open(OUT, "w", encoding="utf-8") as f:
        json.dump(report, f, ensure_ascii=False, indent=2)
        f.write("\n")
    ida_pro.qexit(0)


if __name__ == "__main__":
    main()
