"""非破壞性盤點戰術畫面的武器清單／狀態控制候選函式。"""

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


PATTERN = re.compile(r"weapon|combat", re.I)
ROOTS = [0x2E2CD, 0x2E98D, 0x67FFC, 0x68177, 0x47939,
         0x2C57C, 0x2C63E, 0x2F7E3, 0x2F8AE, 0x2F91C, 0x2FA38,
         0x338F4, 0x33C61, 0x3EFF8]


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
    }


def function_summary(ea):
    fn = ida_funcs.get_func(ea)
    if fn is None:
        return None
    callers = []
    for x in idautils.XrefsTo(fn.start_ea, 0):
        if x.iscode:
            callers.append(instruction(x.frm))
    calls = []
    for item in idautils.FuncItems(fn.start_ea):
        if idc.print_insn_mnem(item).lower().startswith("call"):
            calls.append({
                "site": instruction(item),
                "targets": [{"ea": f"0x{x:X}", "raw_name": ida_name.get_name(x)}
                            for x in idautils.CodeRefsFrom(item, 0)],
            })
    return {
        "start": f"0x{fn.start_ea:X}",
        "end_exclusive": f"0x{fn.end_ea:X}",
        "raw_name": ida_name.get_name(fn.start_ea),
        "callers": callers,
        "direct_calls": calls,
        "instructions": [instruction(item) for item in idautils.FuncItems(fn.start_ea)],
    }


def main():
    ida_auto.auto_wait()
    source = os.environ["MOO2_IDA_INPUT"]
    database = os.environ["MOO2_IDA_DATABASE"]
    candidates = []
    field_hits = []
    for ea in idautils.Functions():
        name = ida_name.get_name(ea)
        if name and PATTERN.search(name):
            rec = function_summary(ea)
            if rec:
                candidates.append(rec)
        for item in idautils.FuncItems(ea):
            line = ida_lines.tag_remove(idc.generate_disasm_line(item, 0) or "")
            if "5Bh]" in line or "5Ch]" in line or "+5Bh" in line or "+5Ch" in line:
                field_hits.append({"function_start": f"0x{ea:X}",
                                   "raw_name": ida_name.get_name(ea),
                                   "instruction": instruction(item)})
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
        "candidate_count": len(candidates),
        "candidates": candidates,
        "field_hits": field_hits,
        "roots": [{"requested": f"0x{ea:X}", "record": function_summary(ea)} for ea in ROOTS],
    }
    with open(os.environ["MOO2_IDA_OUTPUT"], "w", encoding="utf-8") as fh:
        json.dump(payload, fh, ensure_ascii=False, indent=2)
        fh.write("\n")


try:
    main()
except Exception:
    error = traceback.format_exc()
    out = os.environ.get("MOO2_IDA_OUTPUT", "/tmp/tactical-weapon-controls.json")
    with open(out + ".error", "w", encoding="utf-8") as fh:
        fh.write(error)
    ida_kernwin.msg(error + "\n")
    ida_pro.qexit(1)
else:
    ida_pro.qexit(0)
