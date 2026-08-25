"""唯讀匯出殖民地逐人口產出的士氣、領袖與忠誠修正鏈。"""

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


OUT = os.environ.get("MOO2_RE_OUT", "/out/colony-output-modifiers.json")
SOURCE_SHA256 = "7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5"
ROOTS = (0xDE280, 0xDE22C, 0xDDF2C, 0xDDFD3, 0xDD9F2, 0xE1D59)


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
                    "call_ea": f"0x{x:X}", "target_ea": f"0x{target:X}",
                    "original_name": ida_funcs.get_func_name(target), "text": line(x),
                })
    try:
        pseudo = str(ida_hexrays.decompile(f.start_ea))
    except Exception as exc:
        pseudo = f"<unavailable: {exc}>"
    return {
        "ea": f"0x{f.start_ea:X}", "end_ea": f"0x{f.end_ea:X}",
        "original_name": ida_funcs.get_func_name(f.start_ea), "callers": callers,
        "callees": callees, "instructions": instructions, "pseudocode_navigation_only": pseudo,
    }


def main():
    ida_auto.auto_wait()
    report = {
        "schema": "moo2.ida.re-evidence.v1", "evidence_scope": "static_only", "mutation": "none",
        "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version(),
                 "script": "tools/ida/audit_colony_output_modifiers.py"},
        "input": {"database": ida_nalt.get_input_file_path(), "source_sha256": SOURCE_SHA256,
                  "processor": ida_ida.inf_get_procname(),
                  "min_ea": f"0x{ida_ida.inf_get_min_ea():X}",
                  "max_ea": f"0x{ida_ida.inf_get_max_ea():X}"},
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
