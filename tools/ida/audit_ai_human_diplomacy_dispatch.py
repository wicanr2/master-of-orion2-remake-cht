"""非破壞性匯出 AI／真人外交訊息 dispatcher 與請求燈遮罩讀寫端。"""

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


ROOTS = {
    "raw_main_receive_message": 0xF5A9F,
    "raw_humans_requesting_diplomacy": 0xFA795,
    "raw_draw_diplomacy_request_lights": 0x83D06,
    "raw_change_relations": 0x4E3B5,
    "raw_set_formal_policy": 0x5232E,
}
REQUEST_MASK_EA = 0x1AB054


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


def function_record(requested):
    fn = ida_funcs.get_func(requested)
    if fn is None:
        return {"requested": f"0x{requested:X}", "error": "function missing"}
    raw = ida_bytes.get_bytes(fn.start_ea, fn.end_ea - fn.start_ea) or b""
    return {
        "requested": f"0x{requested:X}",
        "original_name": ida_name.get_name(fn.start_ea) or "<unnamed>",
        "start_ea": f"0x{fn.start_ea:X}",
        "end_ea": f"0x{fn.end_ea:X}",
        "bytes_sha256": hashlib.sha256(raw).hexdigest(),
        "instructions": [instruction(ea) for ea in idautils.FuncItems(fn.start_ea)],
        "callers": [instruction(x.frm) for x in idautils.XrefsTo(fn.start_ea, 0)
                    if ida_funcs.get_func(x.frm) is not None],
    }


def main():
    ida_auto.auto_wait()
    source = os.environ["MOO2_IDA_INPUT"]
    database = os.environ["MOO2_IDA_DATABASE"]
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
        "roots": {name: function_record(ea) for name, ea in ROOTS.items()},
        "request_mask": {
            "ea": f"0x{REQUEST_MASK_EA:X}",
            "original_name": ida_name.get_name(REQUEST_MASK_EA) or "<unnamed>",
            "xrefs": [instruction(x.frm) for x in idautils.XrefsTo(REQUEST_MASK_EA, 0)],
        },
    }
    with open(os.environ["MOO2_IDA_OUTPUT"], "w", encoding="utf-8") as stream:
        json.dump(report, stream, ensure_ascii=False, indent=2)
        stream.write("\n")


try:
    main()
except Exception:
    error = traceback.format_exc()
    out = os.environ.get("MOO2_IDA_OUTPUT", "/tmp/ai-human-diplomacy.json")
    with open(out + ".error", "w", encoding="utf-8") as stream:
        stream.write(error)
    ida_kernwin.msg(error + "\n")
    ida_pro.qexit(1)
else:
    ida_pro.qexit(0)
