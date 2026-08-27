"""匯出 Change_Relations_ 與所有直接 caller 的非破壞性證據。"""

import hashlib
import json
import os

import ida_bytes
import ida_funcs
import ida_hexrays
import ida_ida
import ida_name
import ida_nalt
import ida_ua
import idautils
import idc


TARGET_EA = 0x4E3B5


def instruction(ea):
    insn = ida_ua.insn_t()
    size = ida_ua.decode_insn(insn, ea) or 1
    return {"ea": hex(ea), "bytes": (ida_bytes.get_bytes(ea, size) or b"").hex(),
            "disasm": idc.generate_disasm_line(ea, 0) or ""}


def function_instructions(fn, limit=1200):
    out = []
    ea = fn.start_ea
    while ea < fn.end_ea and len(out) < limit:
        out.append(instruction(ea))
        ea = idc.next_head(ea, fn.end_ea)
    return out


def caller_window(call_ea, fn, before=18, after=8):
    eas = list(idautils.FuncItems(fn.start_ea))
    try:
        index = eas.index(call_ea)
    except ValueError:
        return []
    return [instruction(ea) for ea in eas[max(0, index-before):index+after+1]]


def main():
    fn = ida_funcs.get_func(TARGET_EA)
    if fn is None:
        raise RuntimeError("Change_Relations_ function missing")
    raw = ida_bytes.get_bytes(fn.start_ea, fn.end_ea - fn.start_ea) or b""
    callers = []
    for xref in idautils.XrefsTo(fn.start_ea, 0):
        caller = ida_funcs.get_func(xref.frm)
        if caller is None:
            continue
        callers.append({
            "call_ea": hex(xref.frm),
            "caller_start": hex(caller.start_ea),
            "caller_end": hex(caller.end_ea),
            "caller_original_name": ida_name.get_name(caller.start_ea),
            "xref_type": int(xref.type),
            "window": caller_window(xref.frm, caller),
        })
    target = {
        "original_name": ida_name.get_name(fn.start_ea),
        "start_ea": hex(fn.start_ea),
        "end_ea": hex(fn.end_ea),
        "size": fn.end_ea - fn.start_ea,
        "bytes_sha256": hashlib.sha256(raw).hexdigest(),
        "instructions": function_instructions(fn),
        "callers": callers,
    }
    try:
        target["decompiler_navigation_only"] = str(ida_hexrays.decompile(fn.start_ea))
    except Exception as exc:
        target["decompiler_error"] = str(exc)
    result = {
        "tool": "IDA Pro 9.4 IDAPython",
        "root_filename": ida_nalt.get_root_filename(),
        "input_path": ida_nalt.get_input_file_path(),
        "input_sha256": ida_nalt.retrieve_input_file_sha256().hex(),
        "processor": ida_ida.inf_get_procname(),
        "compiler_id": int(ida_ida.inf_get_cc_id()),
        "address_space": "IDA linear EA",
        "target": target,
        "semantic_claim": {"level": "unknown", "note": "caller window 只保存參數線索；語意須人工分級。"},
    }
    output = os.environ.get("MOO2_IDA_PROBE_OUT", "/tmp/change-relations-ida.json")
    with open(output, "w", encoding="utf-8") as stream:
        json.dump(result, stream, ensure_ascii=False, indent=2, sort_keys=True)
        stream.write("\n")
    idc.qexit(0)


main()
