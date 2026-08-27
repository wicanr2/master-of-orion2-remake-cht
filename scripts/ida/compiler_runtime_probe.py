"""匯出 MOO2 的編譯器與 runtime helper 非破壞性證據。"""

import hashlib
import json
import os

import ida_bytes
import ida_funcs
import ida_ida
import ida_kernwin
import ida_name
import ida_nalt
import ida_ua
import idautils
import idc


TARGETS_BY_INPUT = {
    "orion2.exe": {
        0x13F1E0: "__CHK",
        0x13F1F0: "__GRO",
        0x13F1F3: "__STK",
        0x13F211: "__STKOVERFLOW_",
        0x153E1B: "stackavail_",
        0x153E92: "__fatal_runtime_error_",
        0x1640E4: "__chk8087_",
    },
    "orion95.exe": {
        0x5405C0: "__alloca_probe",
        0x542E80: "___CxxFrameHandler",
        0x5436B0: "__CxxThrowException@8",
        0x5486BC: "__except_handler3",
    },
}


def function_record(ea, expected_name):
    fn = ida_funcs.get_func(ea)
    if fn is None:
        return {"ea": hex(ea), "expected_name": expected_name, "error": "no_function"}

    instructions = []
    cursor = fn.start_ea
    while cursor < fn.end_ea and len(instructions) < 96:
        instructions.append(
            {
                "ea": hex(cursor),
                "bytes": ida_bytes.get_bytes(cursor, ida_ua.decode_insn(ida_ua.insn_t(), cursor) or 1).hex(),
                "disasm": idc.generate_disasm_line(cursor, 0) or "",
            }
        )
        cursor = idc.next_head(cursor, fn.end_ea)

    callers = []
    for xref in idautils.XrefsTo(fn.start_ea, 0):
        caller = ida_funcs.get_func(xref.frm)
        callers.append(
            {
                "call_ea": hex(xref.frm),
                "caller_start": hex(caller.start_ea) if caller else None,
                "caller_name": ida_name.get_name(caller.start_ea) if caller else None,
                "type": int(xref.type),
            }
        )

    return {
        "ea": hex(ea),
        "expected_name": expected_name,
        "ida_name": ida_name.get_name(fn.start_ea),
        "start_ea": hex(fn.start_ea),
        "end_ea": hex(fn.end_ea),
        "size": fn.end_ea - fn.start_ea,
        "flags": int(fn.flags),
        "is_library": bool(fn.flags & ida_funcs.FUNC_LIB),
        "bytes_sha256": hashlib.sha256(ida_bytes.get_bytes(fn.start_ea, fn.end_ea - fn.start_ea)).hexdigest(),
        "instructions": instructions,
        "callers": callers,
    }


def main():
    output = os.environ.get("MOO2_IDA_PROBE_OUT", "/tmp/moo2-compiler-runtime-probe.json")
    root_filename = ida_nalt.get_root_filename()
    targets = TARGETS_BY_INPUT.get(root_filename.lower())
    if targets is None:
        raise RuntimeError(f"unsupported input: {root_filename}")
    result = {
        "tool": "IDA Pro 9.4 IDAPython",
        "root_filename": root_filename,
        "input_path": ida_nalt.get_input_file_path(),
        "input_sha256": ida_nalt.retrieve_input_file_sha256().hex(),
        "processor": ida_ida.inf_get_procname(),
        "compiler_id": int(ida_ida.inf_get_cc_id()),
        "address_space": "IDA linear EA",
        "functions": [function_record(ea, name) for ea, name in targets.items()],
    }
    with open(output, "w", encoding="utf-8") as stream:
        json.dump(result, stream, ensure_ascii=False, indent=2, sort_keys=True)
        stream.write("\n")
    ida_kernwin.qexit(0)


main()
