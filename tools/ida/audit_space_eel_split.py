"""唯讀追查太空鰻 type 13 loader 與可能的 30 回合／owner 8 分裂 consumer。"""

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

OUT = os.environ.get("MOO2_RE_OUT", "/out/space-eel-split.json")
SOURCE = os.environ.get("MOO2_RE_SOURCE", ida_nalt.get_input_file_path())
ROOTS = {
    "raw_Load_Eel_Design": 0x57D14,
    "raw_Spawn_Event_Monster": 0xA16BF,
    "raw_Create_Monster_Ship": 0xA1762,
    "raw_Route_Monster_Ship": 0xA1A23,
    "raw_Move_All_Ships": 0xFFEEA,
    "raw_Search_For_Battles": 0xE9D62,
}


def immediate_value(ea, op):
    if idc.get_operand_type(ea, op) != idc.o_imm:
        return None
    return idc.get_operand_value(ea, op)


def main():
    ida_auto.auto_wait()
    candidates = {}
    for start in idautils.Functions():
        f = ida_funcs.get_func(start)
        if not f:
            continue
        has_30 = False
        has_8 = False
        hits = []
        for ea in idautils.FuncItems(start):
            values = [immediate_value(ea, op) for op in range(2)]
            if 30 in values:
                has_30 = True
                hits.append(common.insn(ea))
            if 8 in values:
                has_8 = True
                hits.append(common.insn(ea))
        if has_30 and has_8:
            candidates[start] = {"function": common.function(start), "matching_instructions": hits}

    report = {
        "schema": "moo2.ida.re-evidence.v1", "evidence_scope": "static_only", "mutation": "none",
        "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version(),
                 "script": "tools/ida/audit_space_eel_split.py"},
        "input": {"database": ida_nalt.get_input_file_path(), "source": SOURCE,
                  "source_sha256": common.sha256(SOURCE), "processor": ida_ida.inf_get_procname(),
                  "min_ea": f"0x{ida_ida.inf_get_min_ea():X}", "max_ea": f"0x{ida_ida.inf_get_max_ea():X}"},
        "address_basis": "IDA linear; DOS/4GW LE object #1",
        "roots": {name: common.function(ea) for name, ea in ROOTS.items()},
        "functions_with_immediate_30_and_8": {f"0x{ea:X}": value for ea, value in sorted(candidates.items())},
    }
    os.makedirs(os.path.dirname(OUT), exist_ok=True)
    with open(OUT, "w", encoding="utf-8") as f:
        json.dump(report, f, ensure_ascii=False, indent=2)
        f.write("\n")
    ida_pro.qexit(0)


if __name__ == "__main__":
    main()
