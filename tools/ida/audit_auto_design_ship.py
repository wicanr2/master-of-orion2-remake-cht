"""非破壞性匯出 Auto_Design_Ship_ 與直接決策鏈的原始證據。"""

import hashlib
import json
import os
import traceback

import ida_auto
import ida_bytes
import ida_funcs
import ida_ida
import ida_kernwin
import ida_lines
import ida_name
import ida_nalt
import ida_pro
import ida_ua
import idautils
import idc


ROOT = 0x616A5


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


def function_record(ea, include_instructions=True):
    fn = ida_funcs.get_func(ea)
    if fn is None:
        return {"requested": f"0x{ea:X}", "error": "function missing"}
    calls = []
    for item in idautils.FuncItems(fn.start_ea):
        if idc.print_insn_mnem(item).lower().startswith("call"):
            targets = list(idautils.CodeRefsFrom(item, 0))
            calls.append({
                "site": instruction(item),
                "targets": [{"ea": f"0x{x:X}", "raw_name": ida_name.get_name(x)} for x in targets],
            })
    out = {
        "requested": f"0x{ea:X}",
        "start": f"0x{fn.start_ea:X}",
        "end_exclusive": f"0x{fn.end_ea:X}",
        "raw_name": ida_name.get_name(fn.start_ea),
        "callers": [instruction(x.frm) for x in idautils.XrefsTo(fn.start_ea, 0) if x.iscode],
        "direct_calls": calls,
    }
    if include_instructions:
        out["instructions"] = [instruction(x) for x in idautils.FuncItems(fn.start_ea)]
    return out


def caller_record(ref_ea):
    fn = ida_funcs.get_func(ref_ea)
    if fn is None:
        return {"reference": instruction(ref_ea), "error": "containing function missing"}
    items = list(idautils.FuncItems(fn.start_ea))
    pos = items.index(ref_ea)
    return {
        "reference": instruction(ref_ea),
        "function_start": f"0x{fn.start_ea:X}",
        "function_end_exclusive": f"0x{fn.end_ea:X}",
        "raw_name": ida_name.get_name(fn.start_ea),
        "context": [instruction(x) for x in items[max(0, pos - 24):pos + 9]],
    }


def main():
    ida_auto.auto_wait()
    source = os.environ["MOO2_IDA_INPUT"]
    database = os.environ["MOO2_IDA_DATABASE"]
    root = function_record(ROOT)
    direct_targets = []
    seen = set()
    for call in root.get("direct_calls", []):
        for target in call["targets"]:
            ea = int(target["ea"], 16)
            fn = ida_funcs.get_func(ea)
            if fn and fn.start_ea not in seen:
                seen.add(fn.start_ea)
                direct_targets.append(function_record(fn.start_ea))
    payload = {
        "schema": "moo2.ida.re-evidence.v1",
        "contract": "raw-location + navigation-label + reviewed confidence + source",
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
        "root": root,
        "caller_contexts": [caller_record(x.frm) for x in idautils.XrefsTo(ROOT, 0) if x.iscode],
        "caller_functions": [
            function_record(ida_funcs.get_func(x.frm).start_ea)
            for x in idautils.XrefsTo(ROOT, 0)
            if x.iscode and ida_funcs.get_func(x.frm) is not None
        ],
        "direct_targets": direct_targets,
    }
    with open(os.environ["MOO2_IDA_OUTPUT"], "w", encoding="utf-8") as fh:
        json.dump(payload, fh, ensure_ascii=False, indent=2)
        fh.write("\n")


try:
    main()
except Exception:
    error = traceback.format_exc()
    out = os.environ.get("MOO2_IDA_OUTPUT", "/tmp/auto-design-ship.json")
    with open(out + ".error", "w", encoding="utf-8") as fh:
        fh.write(error)
    ida_kernwin.msg(error + "\n")
    ida_pro.qexit(1)
else:
    ida_pro.qexit(0)
