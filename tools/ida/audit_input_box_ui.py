"""唯讀匯出 MOO2 共用文字輸入彈窗與多人對局名稱 caller。

保留原始名稱、IDA 線性位址、bytes、運算元與 caller／callee；不修改或儲存 IDB。
"""
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

OUT = os.environ.get("MOO2_INPUTBOX_RE_OUT", "/out/input-box-ui.json")
SOURCE = os.environ.get("MOO2_INPUTBOX_RE_SOURCE", ida_nalt.get_input_file_path())
ROOTS = {
    "raw_Input_Callback": 0x91B89,
    "raw_Remapped_Input_Box_Popup": 0x91BB4,
    "raw_Draw_Input_Box": 0x91BD4,
    "raw_Layout_Input_Box": 0x91F14,
    "raw_Change_MP_Game_Name": 0xF5777,
}


def sha256(path):
    h = hashlib.sha256()
    with open(path, "rb") as f:
        for chunk in iter(lambda: f.read(1024 * 1024), b""):
            h.update(chunk)
    return h.hexdigest()


def insn(ea):
    size = idc.get_item_size(ea)
    raw = ida_bytes.get_bytes(ea, size) or b""
    return {
        "ea": f"0x{ea:X}", "bytes": raw.hex(),
        "mnem": idc.print_insn_mnem(ea),
        "op0": idc.print_operand(ea, 0), "op1": idc.print_operand(ea, 1),
        "line": idc.generate_disasm_line(ea, 0) or "<unavailable>",
    }


def function(ea):
    f = ida_funcs.get_func(ea)
    if not f:
        return {"requested": f"0x{ea:X}", "error": "no function"}
    callers = []
    for ref in idautils.CodeRefsTo(f.start_ea, 0):
        caller = ida_funcs.get_func(ref)
        callers.append({
            "instruction": insn(ref),
            "function_start": f"0x{caller.start_ea:X}" if caller else None,
            "function_name": idc.get_name(caller.start_ea) if caller else None,
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
        "requested": f"0x{ea:X}", "start": f"0x{f.start_ea:X}", "end": f"0x{f.end_ea:X}",
        "original_name": idc.get_name(f.start_ea) or "<unnamed>",
        "instructions": [insn(x) for x in idautils.FuncItems(f.start_ea)],
        "callers": callers, "callees": callees,
    }


def main():
    ida_auto.auto_wait()
    report = {
        "schema": "moo2.ida.re-evidence.v1", "evidence_scope": "static_only", "mutation": "none",
        "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version(),
                 "script": "tools/ida/audit_input_box_ui.py"},
        "input": {"database": ida_nalt.get_input_file_path(), "source": SOURCE,
                  "source_sha256": sha256(SOURCE), "processor": ida_ida.inf_get_procname(),
                  "min_ea": f"0x{ida_ida.inf_get_min_ea():X}",
                  "max_ea": f"0x{ida_ida.inf_get_max_ea():X}"},
        "address_basis": "IDA linear; DOS/4GW LE object #1",
        "roots": {name: function(ea) for name, ea in ROOTS.items()},
    }
    os.makedirs(os.path.dirname(OUT), exist_ok=True)
    with open(OUT, "w", encoding="utf-8") as f:
        json.dump(report, f, ensure_ascii=False, indent=2)
        f.write("\n")
    ida_pro.qexit(0)


if __name__ == "__main__":
    main()
