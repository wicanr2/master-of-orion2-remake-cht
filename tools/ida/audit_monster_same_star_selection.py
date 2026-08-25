"""唯讀匯出同星多怪獸戰鬥陣營建立與選擇控制流。"""

import json
import os

import ida_auto
import ida_funcs
import ida_hexrays
import ida_ida
import ida_kernwin
import ida_lines
import ida_nalt
import ida_pro
import idautils
import idc


OUT = os.environ.get("MOO2_RE_OUT", "/out/monster-same-star-selection.json")
SOURCE_SHA256 = "7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5"
ROOTS = (0xE9D62, 0xE8029, 0xE8194, 0xE82B4, 0xE84A5, 0xFE9B5, 0x242CF,
         0xE938C, 0xE9CD3,
         0xE9927, 0xE7CDB, 0xE7DCA)


def line(ea):
    return ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or "")


def function(ea):
    f = ida_funcs.get_func(ea)
    if not f:
        return None
    instructions = [{"ea": f"0x{x:X}", "text": line(x)} for x in idautils.FuncItems(f.start_ea)]
    callers = []
    for ref in idautils.CodeRefsTo(f.start_ea, False):
        caller = ida_funcs.get_func(ref)
        callers.append({
            "call_ea": f"0x{ref:X}",
            "function_ea": f"0x{caller.start_ea:X}" if caller else None,
            "original_name": ida_funcs.get_func_name(caller.start_ea) if caller else None,
            "text": line(ref),
        })
    callees = []
    for x in idautils.FuncItems(f.start_ea):
        for target in idautils.CodeRefsFrom(x, False):
            target_func = ida_funcs.get_func(target)
            if target_func and target_func.start_ea == target:
                callees.append({
                    "call_ea": f"0x{x:X}",
                    "target_ea": f"0x{target:X}",
                    "original_name": ida_funcs.get_func_name(target),
                    "text": line(x),
                })
    try:
        pseudo = str(ida_hexrays.decompile(f.start_ea))
    except Exception as exc:
        pseudo = f"<unavailable: {exc}>"
    return {
        "ea": f"0x{f.start_ea:X}",
        "original_name": ida_funcs.get_func_name(f.start_ea),
        "end_ea": f"0x{f.end_ea:X}",
        "callers": callers,
        "callees": callees,
        "instructions": instructions,
        "pseudocode_navigation_only": pseudo,
    }


def main():
    ida_auto.auto_wait()
    report = {
        "schema": "moo2.ida.re-evidence.v1",
        "evidence_scope": "static_only",
        "mutation": "none",
        "tool": {
            "name": "IDA Pro",
            "version": ida_kernwin.get_kernel_version(),
            "script": "tools/ida/audit_monster_same_star_selection.py",
        },
        "input": {
            "database": ida_nalt.get_input_file_path(),
            "source_sha256": SOURCE_SHA256,
            "processor": ida_ida.inf_get_procname(),
            "min_ea": f"0x{ida_ida.inf_get_min_ea():X}",
            "max_ea": f"0x{ida_ida.inf_get_max_ea():X}",
        },
        "address_basis": "IDA linear; DOS/4GW LE object #1",
        "roots": {f"0x{ea:X}": function(ea) for ea in ROOTS},
        "warning": "pseudocode is navigation only; conclusions require instructions/xrefs",
    }
    os.makedirs(os.path.dirname(OUT), exist_ok=True)
    with open(OUT, "w", encoding="utf-8") as handle:
        json.dump(report, handle, ensure_ascii=False, indent=2)
        handle.write("\n")
    ida_pro.qexit(0)


if __name__ == "__main__":
    main()
