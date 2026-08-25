"""唯讀匯出 Plasma Flux（weapon ID 44）目標枚舉與傷害資料流。"""

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


OUT = os.environ.get("MOO2_RE_OUT", "/out/plasma-flux-spread.json")
SOURCE_SHA256 = "7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5"
ROOTS = (0x289C4, 0x2A545, 0x38B5E, 0x39F1D, 0x394F7, 0x39E15, 0x3A82F, 0x3BB3D,
         0x3C539, 0x3E095, 0xADE18)


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
        callers.append({"call_ea": f"0x{ref:X}", "function_ea": f"0x{caller.start_ea:X}" if caller else None,
                        "function_name": ida_funcs.get_func_name(caller.start_ea) if caller else None,
                        "text": line(ref)})
    pseudo = None
    try:
        pseudo = str(ida_hexrays.decompile(f.start_ea))
    except Exception as exc:
        pseudo = f"<unavailable: {exc}>"
    return {"ea": f"0x{f.start_ea:X}", "original_name": ida_funcs.get_func_name(f.start_ea),
            "end_ea": f"0x{f.end_ea:X}", "callers": callers, "instructions": instructions,
            "pseudocode_navigation_only": pseudo}


def main():
    ida_auto.auto_wait()
    id44 = {}
    for ea in idautils.Functions():
        f = ida_funcs.get_func(ea)
        if not f:
            continue
        lines = [line(x).lower() for x in idautils.FuncItems(f.start_ea)]
        if any(token in text for text in lines for token in ("2ch", "44")):
            # 限縮到戰鬥位址區及有武器表／combat-record 形狀者，避免把十進位 44 全庫灌入。
            if ea < 0x60000 and any(token in text for text in lines
                                   for token in ("dword_192864", "139h", "0bh", "+25h", "+ 25h")):
                id44[f"0x{ea:X}"] = function(ea)
    report = {
        "schema": "moo2.ida.re-evidence.v1",
        "evidence_scope": "static_only",
        "mutation": "none",
        "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version(),
                 "script": "tools/ida/audit_plasma_flux_spread.py"},
        "input": {"database": ida_nalt.get_input_file_path(), "source_sha256": SOURCE_SHA256,
                  "processor": ida_ida.inf_get_procname(), "min_ea": f"0x{ida_ida.inf_get_min_ea():X}",
                  "max_ea": f"0x{ida_ida.inf_get_max_ea():X}"},
        "address_basis": "IDA linear; DOS/4GW LE object #1",
        "roots": {f"0x{ea:X}": function(ea) for ea in ROOTS},
        "weapon_id_44_candidates": id44,
        "warning": "pseudocode is navigation only; conclusions require instructions/xrefs",
    }
    os.makedirs(os.path.dirname(OUT), exist_ok=True)
    with open(OUT, "w", encoding="utf-8") as handle:
        json.dump(report, handle, ensure_ascii=False, indent=2)
        handle.write("\n")
    ida_pro.qexit(0)


if __name__ == "__main__":
    main()
