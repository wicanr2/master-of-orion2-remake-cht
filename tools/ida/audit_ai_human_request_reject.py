"""非破壞性匯出 AI 對真人要求拒絕 callback 與 Change_Relations_ 方向。"""

import hashlib
import json
import os
import traceback

import ida_auto
import ida_bytes
import ida_funcs
import ida_hexrays
import ida_ida
import ida_kernwin
import ida_name
import ida_pro
import ida_ua
import idautils
import idc


ROOTS = {
    "raw_diplomacy_dispatch_1afa6": 0x1AFA6,
    "raw_change_relations_4e3b5": 0x4E3B5,
    "raw_declare_war_51078": 0x51078,
    "raw_human_target_producer_53edb": 0x53EDB,
    "raw_target_outputs_544a1": 0x544A1,
}


def digest(path):
    h = hashlib.sha256()
    with open(path, "rb") as f:
        for chunk in iter(lambda: f.read(1 << 20), b""):
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
    f = ida_funcs.get_func(ea)
    if not f:
        return {"requested": f"0x{ea:X}", "error": "missing"}
    raw = ida_bytes.get_bytes(f.start_ea, f.end_ea - f.start_ea) or b""
    pseudo = None
    if ida_hexrays.init_hexrays_plugin():
        try:
            pseudo = str(ida_hexrays.decompile(f.start_ea))
        except Exception as exc:
            pseudo = f"<decompile failed: {exc}>"
    return {
        "requested": f"0x{ea:X}",
        "original_name": ida_name.get_name(f.start_ea) or "<unnamed>",
        "start_ea": f"0x{f.start_ea:X}",
        "end_ea": f"0x{f.end_ea:X}",
        "bytes_sha256": hashlib.sha256(raw).hexdigest(),
        "pseudocode_navigation_only": pseudo,
        "instructions": [instruction(x) for x in idautils.FuncItems(f.start_ea)],
    }


def main():
    ida_auto.auto_wait()
    src = os.environ["MOO2_IDA_INPUT"]
    db = os.environ["MOO2_IDA_SOURCE_DATABASE"]
    report = {
        "schema": "moo2.ida.re-evidence.v1",
        "contract": "raw location/name/bytes/xrefs; semantics reviewed externally",
        "mutation": "none",
        "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version()},
        "input": {
            "file": os.path.basename(src),
            "source_sha256": digest(src),
            "database_sha256": digest(db),
            "processor": ida_ida.inf_get_procname(),
        },
        "address_basis": "IDA linear; DOS/4GW LE image",
        "semantic_status": "unknown_pending_review",
        "roots": {key: function_record(ea) for key, ea in ROOTS.items()},
    }
    with open(os.environ["MOO2_IDA_OUTPUT"], "w", encoding="utf-8") as f:
        json.dump(report, f, ensure_ascii=False, indent=2)
        f.write("\n")


try:
    main()
except Exception:
    path = os.environ.get("MOO2_IDA_OUTPUT", "/tmp/ai-human-request-reject.json")
    with open(path + ".error", "w", encoding="utf-8") as f:
        f.write(traceback.format_exc())
    ida_pro.qexit(1)
else:
    ida_pro.qexit(0)
