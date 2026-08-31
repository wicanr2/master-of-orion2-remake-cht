"""非破壞性匯出戰術 raw deployment、camera、sprite center 與縮圖候選 consumer。"""

import hashlib
import json
import os
import re
import traceback

import ida_auto
import ida_bytes
import ida_funcs
import ida_ida
import ida_kernwin
import ida_lines
import ida_name
import ida_pro
import ida_ua
import idautils
import idc


ROOTS = [
    0x2F4EE,
    0x30062,
    0x34454,
    0x465D0,
    0x46CC8,
    0x47939,
    0x49043,
    0x49A41,
    0x49D09,
    0x4446A,
    0x444EE,
    0x4A6C5,
    0x12A478,
    0x12ACA4,
]
DATA_ROOTS = [0x1998F0, 0x1998F2]
NAME_PATTERN = re.compile(r"camera|viewport|mini.?map|ship.*center|center.*ship|draw.*combat|combat.*draw", re.I)


def digest(path):
    h = hashlib.sha256()
    with open(path, "rb") as fh:
        for chunk in iter(lambda: fh.read(1024 * 1024), b""):
            h.update(chunk)
    return h.hexdigest()


def instruction(ea):
    insn = ida_ua.insn_t()
    size = ida_ua.decode_insn(insn, ea)
    if size <= 0:
        size = max(1, idc.get_item_size(ea))
    return {
        "ea": f"0x{ea:X}",
        "bytes": (ida_bytes.get_bytes(ea, size) or b"").hex(" "),
        "text": ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or ""),
        "code_refs": [f"0x{x:X}" for x in idautils.CodeRefsFrom(ea, 0)],
        "data_refs": [f"0x{x:X}" for x in idautils.DataRefsFrom(ea)],
    }


def function_record(ea):
    fn = ida_funcs.get_func(ea)
    if fn is None:
        return {"requested": f"0x{ea:X}", "error": "function missing"}
    return {
        "requested": f"0x{ea:X}",
        "start": f"0x{fn.start_ea:X}",
        "end_exclusive": f"0x{fn.end_ea:X}",
        "raw_name": ida_name.get_name(fn.start_ea),
        "bytes_sha256": hashlib.sha256(b"".join(
            ida_bytes.get_bytes(x, max(1, idc.get_item_size(x))) or b""
            for x in idautils.FuncItems(fn.start_ea)
        )).hexdigest(),
        "callers": [instruction(x.frm) for x in idautils.XrefsTo(fn.start_ea, 0) if x.iscode],
        "instructions": [instruction(x) for x in idautils.FuncItems(fn.start_ea)],
    }


def direct_callee_starts(ea):
    fn = ida_funcs.get_func(ea)
    if fn is None:
        return []
    out = set()
    for item in idautils.FuncItems(fn.start_ea):
        if idc.print_insn_mnem(item).lower().startswith("call"):
            for target in idautils.CodeRefsFrom(item, 0):
                callee = ida_funcs.get_func(target)
                if callee is not None:
                    out.add(callee.start_ea)
    return sorted(out)


def data_xrefs(ea):
    refs = []
    for x in idautils.XrefsTo(ea, 0):
        owner = ida_funcs.get_func(x.frm)
        refs.append({
            "site": instruction(x.frm),
            "owner_start": f"0x{owner.start_ea:X}" if owner else None,
            "owner_raw_name": ida_name.get_name(owner.start_ea) if owner else None,
        })
    return {"address": f"0x{ea:X}", "raw_name": ida_name.get_name(ea), "refs": refs}


def main():
    ida_auto.auto_wait()
    source = os.environ["MOO2_IDA_INPUT"]
    database = os.environ["MOO2_IDA_DATABASE"]
    root_records = [function_record(ea) for ea in ROOTS]
    renderer_callees = [function_record(ea) for ea in direct_callee_starts(ROOTS[0])]
    named = []
    for ea in idautils.Functions():
        name = ida_name.get_name(ea)
        if name and NAME_PATTERN.search(name):
            named.append(function_record(ea))
    payload = {
        "schema": "moo2.ida.re-evidence.v1",
        "contract": "raw-location + additive-semantics + reviewed-confidence + source",
        "mutation": "none",
        "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version(),
                 "script": "tools/ida/audit_tactical_deployment_renderer.py"},
        "input": {
            "file": os.path.basename(source),
            "source_sha256": digest(source),
            "database_sha256": digest(database),
            "processor": ida_ida.inf_get_procname(),
        },
        "address_basis": "IDA linear; DOS/4GW LE image",
        "semantic_status": "unknown_pending_review",
        "roots": root_records,
        "renderer_direct_callees": renderer_callees,
        "named_candidates": named,
        "camera_data_xrefs": [data_xrefs(ea) for ea in DATA_ROOTS],
    }
    with open(os.environ["MOO2_IDA_OUTPUT"], "w", encoding="utf-8") as fh:
        json.dump(payload, fh, ensure_ascii=False, indent=2)
        fh.write("\n")


try:
    main()
except Exception:
    error = traceback.format_exc()
    out = os.environ.get("MOO2_IDA_OUTPUT", "/tmp/tactical-deployment-renderer.json")
    with open(out + ".error", "w", encoding="utf-8") as fh:
        fh.write(error)
    ida_kernwin.msg(error + "\n")
    ida_pro.qexit(1)
else:
    ida_pro.qexit(0)
