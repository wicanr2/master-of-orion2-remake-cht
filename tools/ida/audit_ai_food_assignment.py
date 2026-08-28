"""非破壞性匯出 MOO2 AI 職務後續食物／殖民地 caller 鏈。"""

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
    "raw_unblockaded_food_score": 0xD682A,
    "raw_unblockaded_food_sort": 0xD68CB,
    "raw_assign_additional_farmer": 0xD6A00,
    "raw_post_job_colony_pass": 0xD6AD4,
    "raw_post_job_empire_pass": 0xD6D80,
    "raw_ai_colony_job_and_build_dispatch": 0xD6E1D,
    "raw_colony_ai": 0xD6ED4,
    "raw_all_colony_ai": 0xD6F67,
    "raw_recompute_food_transport": 0xE2D09,
    "raw_recompute_empire_for_player": 0xE2D72,
    "raw_recompute_player_food": 0xDF8F0,
    "raw_colony_food_transport_a": 0xE1839,
    "raw_colony_food_transport_b": 0xE1E1F,
    "raw_player_maintenance": 0xE2000,
}


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
    pseudo = None
    if ida_hexrays.init_hexrays_plugin():
        try:
            pseudo = str(ida_hexrays.decompile(fn.start_ea))
        except Exception as exc:
            pseudo = f"<decompile failed: {exc}>"
    return {
        "requested": f"0x{requested:X}",
        "original_name": ida_name.get_name(fn.start_ea) or "<unnamed>",
        "start_ea": f"0x{fn.start_ea:X}",
        "end_ea": f"0x{fn.end_ea:X}",
        "bytes_sha256": hashlib.sha256(raw).hexdigest(),
        "pseudocode_navigation_only": pseudo,
        "instructions": [instruction(ea) for ea in idautils.FuncItems(fn.start_ea)],
        "callers": [
            instruction(xref.frm)
            for xref in idautils.XrefsTo(fn.start_ea, 0)
            if ida_funcs.get_func(xref.frm) is not None
        ],
    }


def transport_operand_functions():
    out = {}
    needles = ("+36h]", "+38h]", "+3eh]", "+40h]", "+0f3h]", "+f3h]")
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
        "transport_operand_functions": transport_operand_functions(),
    }
    with open(os.environ["MOO2_IDA_OUTPUT"], "w", encoding="utf-8") as stream:
        json.dump(report, stream, ensure_ascii=False, indent=2)
        stream.write("\n")


try:
    main()
except Exception:
    error = traceback.format_exc()
    out = os.environ.get("MOO2_IDA_OUTPUT", "/tmp/ai-food-assignment.json")
    with open(out + ".error", "w", encoding="utf-8") as stream:
        stream.write(error)
    ida_kernwin.msg(error + "\n")
    ida_pro.qexit(1)
else:
    ida_pro.qexit(0)
