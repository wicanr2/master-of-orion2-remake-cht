"""非破壞性匯出原版片頭／Smack 玩家路徑候選函式。"""

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
import ida_pro
import ida_ua
import idautils
import idc

ROOTS = [0x14DF7, 0x15085, 0x150E5, 0x24ED3, 0x2518F]
DATA_ROOTS = [0x180178, 0x19A004, 0x19A018, 0x19A01C, 0x19A01E]


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
    callers = [instruction(x.frm) for x in idautils.XrefsTo(fn.start_ea, 0) if x.iscode]
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


def data_summary(ea):
    refs = []
    for x in idautils.XrefsTo(ea, 0):
        owner = ida_funcs.get_func(x.frm)
        refs.append({
            "site": instruction(x.frm),
            "function_start": f"0x{owner.start_ea:X}" if owner else None,
            "raw_name": ida_name.get_name(owner.start_ea) if owner else None,
        })
    return {"address": f"0x{ea:X}", "raw_name": ida_name.get_name(ea), "refs": refs}


def main():
    ida_auto.auto_wait()
    source = os.environ["MOO2_IDA_INPUT"]
    database = os.environ["MOO2_IDA_DATABASE"]
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
        "roots": [{"requested": f"0x{ea:X}", "record": function_summary(ea)} for ea in ROOTS],
        "data_roots": [data_summary(ea) for ea in DATA_ROOTS],
    }
    with open(os.environ["MOO2_IDA_OUTPUT"], "w", encoding="utf-8") as fh:
        json.dump(payload, fh, ensure_ascii=False, indent=2)
        fh.write("\n")


try:
    main()
except Exception:
    output = os.environ.get("MOO2_IDA_OUTPUT", "/tmp/cutscene-player-path.json")
    with open(output + ".error", "w", encoding="utf-8") as fh:
        fh.write(traceback.format_exc())
    ida_pro.qexit(1)
else:
    ida_pro.qexit(0)
