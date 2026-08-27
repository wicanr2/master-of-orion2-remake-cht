"""匯出殖民入口與殖民地建立函式的非破壞性證據。"""

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


TARGETS = (
    0xBB082,
    0xC0B87,
    0xE5EB3,
    0xE6071,
    0xE65F8,
    0x8B2DE,
    0xFDB01,
    0x12D75,
    0xE2A70,
    0xE5296,
)


def instruction_record(ea):
    insn = ida_ua.insn_t()
    size = ida_ua.decode_insn(insn, ea) or 1
    return {
        "ea": hex(ea),
        "bytes": (ida_bytes.get_bytes(ea, size) or b"").hex(),
        "disasm": idc.generate_disasm_line(ea, 0) or "",
    }


def function_record(ea, instruction_limit=1600):
    fn = ida_funcs.get_func(ea)
    if fn is None:
        return {"requested_ea": hex(ea), "error": "no_function"}
    raw = ida_bytes.get_bytes(fn.start_ea, fn.end_ea - fn.start_ea) or b""
    instructions = []
    cursor = fn.start_ea
    while cursor < fn.end_ea and len(instructions) < instruction_limit:
        instructions.append(instruction_record(cursor))
        cursor = idc.next_head(cursor, fn.end_ea)
    callers = []
    for xref in idautils.XrefsTo(fn.start_ea, 0):
        caller = ida_funcs.get_func(xref.frm)
        callers.append(
            {
                "call_ea": hex(xref.frm),
                "caller_start": hex(caller.start_ea) if caller else None,
                "caller_name": ida_name.get_name(caller.start_ea) if caller else None,
                "xref_type": int(xref.type),
            }
        )
    callees = []
    for item_ea in idautils.FuncItems(fn.start_ea):
        for xref in idautils.XrefsFrom(item_ea, 0):
            callee = ida_funcs.get_func(xref.to)
            if callee is None or callee.start_ea == fn.start_ea:
                continue
            key = (item_ea, callee.start_ea)
            if any(x["reference_ea"] == hex(key[0]) and x["callee_start"] == hex(key[1]) for x in callees):
                continue
            callees.append(
                {
                    "reference_ea": hex(item_ea),
                    "callee_start": hex(callee.start_ea),
                    "callee_name": ida_name.get_name(callee.start_ea),
                    "xref_type": int(xref.type),
                }
            )
    record = {
        "requested_ea": hex(ea),
        "original_name": ida_name.get_name(fn.start_ea),
        "start_ea": hex(fn.start_ea),
        "end_ea": hex(fn.end_ea),
        "size": fn.end_ea - fn.start_ea,
        "bytes_sha256": hashlib.sha256(raw).hexdigest(),
        "instructions": instructions,
        "callers": callers,
        "callees": callees,
    }
    try:
        record["decompiler_navigation_only"] = str(ida_hexrays.decompile(fn.start_ea))
    except Exception as exc:
        record["decompiler_error"] = str(exc)
    return record


def main():
    output = os.environ.get("MOO2_IDA_PROBE_OUT", "/tmp/colonization-ida.json")
    result = {
        "tool": "IDA Pro 9.4 IDAPython",
        "root_filename": ida_nalt.get_root_filename(),
        "input_path": ida_nalt.get_input_file_path(),
        "input_sha256": ida_nalt.retrieve_input_file_sha256().hex(),
        "processor": ida_ida.inf_get_procname(),
        "compiler_id": int(ida_ida.inf_get_cc_id()),
        "address_space": "IDA linear EA",
        "functions": [function_record(ea) for ea in TARGETS],
        "semantic_claim": {
            "level": "unknown",
            "note": "函式語意須由原始指令、caller 與 record consumer 人工審查；反編譯文字只供導覽。",
        },
    }
    with open(output, "w", encoding="utf-8") as stream:
        json.dump(result, stream, ensure_ascii=False, indent=2, sort_keys=True)
        stream.write("\n")
    idc.qexit(0)


main()
