"""非破壞性匯出 AI 對真人 directional incident producer 證據。"""

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
    "raw_world_diplomacy_dispatch_4eb06": 0x4EB06,
    "raw_human_incident_consumer_4f0dc": 0x4F0DC,
    "raw_change_relations_4e3b5": 0x4E3B5,
    "raw_human_target_score_544a1": 0x544A1,
    "raw_spy_score_caller_10119c": 0x10119C,
    "raw_steal_app_10130a": 0x10130A,
    "raw_spy_resolution_1014a4": 0x1014A4,
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
        "callers": [f"0x{x:X}" for x in idautils.CodeRefsTo(f.start_ea, 0)],
        "pseudocode_navigation_only": pseudo,
        "instructions": [instruction(x) for x in idautils.FuncItems(f.start_ea)],
    }


def callsite_windows(target, radius=16):
    rows = []
    for call_ea in idautils.CodeRefsTo(target, 0):
        f = ida_funcs.get_func(call_ea)
        if not f:
            continue
        items = list(idautils.FuncItems(f.start_ea))
        try:
            pos = items.index(call_ea)
        except ValueError:
            continue
        rows.append({
            "caller_original_name": ida_name.get_name(f.start_ea) or "<unnamed>",
            "caller_start_ea": f"0x{f.start_ea:X}",
            "call_ea": f"0x{call_ea:X}",
            "window": [instruction(x) for x in items[max(0, pos - radius):pos + 4]],
        })
    return rows


def incident_operand_functions():
    out = {}
    needles = ("+64fh]", "+65fh]", "+6cfh]", "+71fh]")
    for fea in idautils.Functions():
        hits = []
        for ea in idautils.FuncItems(fea):
            text = (idc.generate_disasm_line(ea, 0) or "").lower()
            if any(needle in text for needle in needles):
                hits.append(instruction(ea))
        if hits:
            out[f"0x{fea:X}"] = {
                "original_name": ida_name.get_name(fea) or "<unnamed>",
                "hits": hits,
                "callers": [f"0x{x:X}" for x in idautils.CodeRefsTo(fea, 0)],
            }
    return out


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
        "change_relations_callsite_windows": callsite_windows(0x4E3B5),
        "incident_consumer_callsite_windows": callsite_windows(0x4F0DC),
        "incident_operand_functions": incident_operand_functions(),
    }
    with open(os.environ["MOO2_IDA_OUTPUT"], "w", encoding="utf-8") as f:
        json.dump(report, f, ensure_ascii=False, indent=2)
        f.write("\n")


try:
    main()
except Exception:
    path = os.environ.get("MOO2_IDA_OUTPUT", "/tmp/ai-human-incident-producer.json")
    with open(path + ".error", "w", encoding="utf-8") as f:
        f.write(traceback.format_exc())
    ida_pro.qexit(1)
else:
    ida_pro.qexit(0)
