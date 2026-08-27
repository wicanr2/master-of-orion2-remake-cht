"""非破壞性匯出 MOO2 Compute_Blockades_ 與直接資料流。"""

import hashlib
import json
import os
import traceback

import ida_auto
import ida_bytes
import ida_funcs
import ida_ida
import ida_kernwin
import ida_name
import ida_pro
import ida_ua
import idautils
import idc


ROOT = 0xE5097


def digest(path):
    h = hashlib.sha256()
    with open(path, "rb") as stream:
        for chunk in iter(lambda: stream.read(1024 * 1024), b""):
            h.update(chunk)
    return h.hexdigest()


def instruction(ea):
    insn = ida_ua.insn_t()
    size = ida_ua.decode_insn(insn, ea) or 1
    return {
        "ea": f"0x{ea:X}",
        "bytes": (ida_bytes.get_bytes(ea, size) or b"").hex(),
        "text": idc.generate_disasm_line(ea, 0) or "",
        "code_refs": [f"0x{x:X}" for x in idautils.CodeRefsFrom(ea, 0)],
        "data_refs": [f"0x{x:X}" for x in idautils.DataRefsFrom(ea)],
    }


def function_record(ea):
    fn = ida_funcs.get_func(ea)
    if fn is None:
        return {"requested": f"0x{ea:X}", "error": "function missing"}
    raw = ida_bytes.get_bytes(fn.start_ea, fn.end_ea - fn.start_ea) or b""
    return {
        "requested": f"0x{ea:X}",
        "original_name": ida_name.get_name(fn.start_ea) or "<unnamed>",
        "start_ea": f"0x{fn.start_ea:X}",
        "end_ea": f"0x{fn.end_ea:X}",
        "bytes_sha256": hashlib.sha256(raw).hexdigest(),
        "instructions": [instruction(item) for item in idautils.FuncItems(fn.start_ea)],
        "callers": [
            instruction(xref.frm)
            for xref in idautils.XrefsTo(fn.start_ea, 0)
            if ida_funcs.get_func(xref.frm) is not None
        ],
    }


def direct_callees(ea):
    fn = ida_funcs.get_func(ea)
    if fn is None:
        return []
    found = set()
    for item in idautils.FuncItems(fn.start_ea):
        for target in idautils.CodeRefsFrom(item, 0):
            callee = ida_funcs.get_func(target)
            if callee is not None and callee.start_ea != fn.start_ea:
                found.add(callee.start_ea)
    return sorted(found)


def caller_functions(ea):
    found = set()
    for xref in idautils.XrefsTo(ea, 0):
        fn = ida_funcs.get_func(xref.frm)
        if fn is not None:
            found.add(fn.start_ea)
    return sorted(found)


def main():
    ida_auto.auto_wait()
    source = os.environ["MOO2_IDA_INPUT"]
    database = os.environ["MOO2_IDA_DATABASE"]
    callees = direct_callees(ROOT)
    report = {
        "schema": "moo2.ida.re-evidence.v1",
        "contract": "raw-location + original-name + bytes + xrefs; semantics reviewed externally",
        "mutation": "none",
        "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version()},
        "input": {
            "file": os.path.basename(source),
            "source_sha256": digest(source),
            "database_sha256": digest(database),
            "processor": ida_ida.inf_get_procname(),
        },
        "address_basis": "IDA linear; DOS/4GW LE image",
        "semantic_status": "unknown_pending_review",
        "root": function_record(ROOT),
        "caller_functions": [function_record(ea) for ea in caller_functions(ROOT)],
        "direct_callees": [function_record(ea) for ea in callees],
    }
    with open(os.environ["MOO2_IDA_OUTPUT"], "w", encoding="utf-8") as stream:
        json.dump(report, stream, ensure_ascii=False, indent=2)
        stream.write("\n")


try:
    main()
except Exception:
    error = traceback.format_exc()
    out = os.environ.get("MOO2_IDA_OUTPUT", "/tmp/blockades.json")
    with open(out + ".error", "w", encoding="utf-8") as stream:
        stream.write(error)
    ida_kernwin.msg(error + "\n")
    ida_pro.qexit(1)
else:
    ida_pro.qexit(0)
