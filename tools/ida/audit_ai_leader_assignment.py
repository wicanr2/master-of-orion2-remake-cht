"""唯讀匯出 MOO2 AI 領袖招募、任命與回合 caller 證據。"""

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


OUT = os.environ.get("MOO2_RE_OUT", "/out/ai-leader-assignment.json")
SOURCE = os.environ.get("MOO2_RE_SOURCE", ida_nalt.get_input_file_path())
ROOTS = {
    "raw_Do_AI_Leaders_documented_anchor": 0xD7439,
    "raw_Do_AI_Leaders_old_script_anchor": 0xD7662,
    "raw_Next_Turn_Calc": 0x136B3,
    "raw_Set_Officer_To_Player": 0x97287,
    "raw_Officer_Type_Counts": 0x977F9,
    "raw_Officer_Candidate_Eligible": 0x97C2D,
    "raw_Officer_Hire_Cost_Tail": 0x97A2D,
    "raw_AI_Assign_Ship_Leader": 0xD6FDA,
    "raw_AI_Assign_Admin_Leader_Status0": 0xD7171,
    "raw_AI_Assign_Admin_Leader_Status1": 0xD7078,
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
        if idc.print_insn_mnem(item).lower() != "call":
            continue
        targets = list(idautils.CodeRefsFrom(item, 0))
        callees.append({
            "instruction": insn(item),
            "targets": [{"ea": f"0x{x:X}", "name": idc.get_name(x) or "<unnamed>"} for x in targets],
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


def main():
    ida_auto.auto_wait()
    primary = ida_funcs.get_func(0xD7439)
    direct = set()
    if primary:
        for ea in idautils.FuncItems(primary.start_ea):
            if idc.print_insn_mnem(ea).lower() == "call":
                direct.update(idautils.CodeRefsFrom(ea, 0))
    report = {
        "schema": "moo2.ida.re-evidence.v1",
        "evidence_scope": "static_only",
        "mutation": "none",
        "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version(),
                 "script": "tools/ida/audit_ai_leader_assignment.py"},
        "input": {"database": ida_nalt.get_input_file_path(), "source": SOURCE,
                  "source_sha256": sha256(SOURCE), "processor": ida_ida.inf_get_procname(),
                  "min_ea": f"0x{ida_ida.inf_get_min_ea():X}",
                  "max_ea": f"0x{ida_ida.inf_get_max_ea():X}"},
        "address_basis": "IDA linear; DOS/4GW LE object #1",
        "roots": {name: function(ea) for name, ea in ROOTS.items()},
        "primary_direct_callees": {f"0x{ea:X}": function(ea) for ea in sorted(direct)},
    }
    os.makedirs(os.path.dirname(OUT), exist_ok=True)
    with open(OUT, "w", encoding="utf-8") as f:
        json.dump(report, f, ensure_ascii=False, indent=2)
        f.write("\n")
    ida_pro.qexit(0)


if __name__ == "__main__":
    main()
