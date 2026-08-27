"""非破壞性匯出 MOO2 AI 殖民地職務分配鏈。"""

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
import ida_name
import ida_nalt
import ida_pro
import ida_ua
import idautils
import idc


ROOTS = {
    "raw_initial_sort": 0xD5FA9,
    "raw_collect_required_colonists": 0xD5FE1,
    "raw_blockaded_sort": 0xD614D,
    "raw_blockaded_assignment": 0xD61E7,
    "raw_unblockaded_production_sort": 0xD6315,
    "raw_unblockaded_colonist_sort": 0xD63A6,
    "raw_unblockaded_marginal_output": 0xD648A,
    "raw_unblockaded_recompute_candidate": 0xD664B,
    "raw_unblockaded_assignment": 0xD652C,
    "raw_empire_job_balance": 0xD66B3,
    "raw_ai_colony_turn_dispatch": 0xD6E1D,
    "raw_colonist_output_for_job": 0xDDFD3,
    "raw_recompute_colony_output": 0xE1D59,
    "raw_recompute_colony_for_player": 0xE2A70,
    "raw_recompute_player_economy": 0xE2710,
    "raw_recompute_empire_for_player": 0xE2D72,
    "raw_integer_sqrt": 0x134C92,
}

RAW_WINDOWS = (0xD578F, 0xD61D2, 0xD61D6)
TRACKED_OFFSETS = (0xAA, 0xAC)


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
        "callers": [
            instruction(xref.frm)
            for xref in idautils.XrefsTo(fn.start_ea, 0)
            if ida_funcs.get_func(xref.frm) is not None
        ],
    }


def instruction_window(ea, radius=5):
    fn = ida_funcs.get_func(ea)
    if fn is None:
        return {"requested": f"0x{ea:X}", "error": "function missing"}
    items = list(idautils.FuncItems(fn.start_ea))
    index = min(range(len(items)), key=lambda i: abs(items[i] - ea))
    return {
        "requested": f"0x{ea:X}",
        "owner_start": f"0x{fn.start_ea:X}",
        "owner_original_name": ida_name.get_name(fn.start_ea) or "<unnamed>",
        "instructions": [instruction(item) for item in items[max(0, index-radius):index+radius+1]],
    }


def operand_mentions_offset(ea, offset):
    operands = " ".join(idc.print_operand(ea, i) for i in range(2)).lower().replace("0x", "")
    for match in re.finditer(r"\+\s*([0-9a-f]+)h", operands):
        if int(match.group(1), 16) == offset:
            return True
    for match in re.finditer(r"\+\s*([0-9]+)(?=\s*[\],])", operands):
        if int(match.group(1), 10) == offset:
            return True
    return False


def direct_offset_refs(offset):
    refs = []
    for fn_ea in idautils.Functions():
        items = list(idautils.FuncItems(fn_ea))
        for index, ea in enumerate(items):
            if not operand_mentions_offset(ea, offset):
                continue
            refs.append({
                "function_start": f"0x{fn_ea:X}",
                "original_name": ida_name.get_name(fn_ea) or "<unnamed>",
                "instruction": instruction(ea),
                "context": [instruction(item) for item in items[max(0, index-5):index+6]],
            })
    return refs


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
        "raw_windows": [instruction_window(ea) for ea in RAW_WINDOWS],
        "direct_offset_refs": {
            f"0x{offset:X}": direct_offset_refs(offset) for offset in TRACKED_OFFSETS
        },
    }
    with open(os.environ["MOO2_IDA_OUTPUT"], "w", encoding="utf-8") as stream:
        json.dump(report, stream, ensure_ascii=False, indent=2)
        stream.write("\n")


try:
    main()
except Exception:
    error = traceback.format_exc()
    out = os.environ.get("MOO2_IDA_OUTPUT", "/tmp/ai-colony-jobs.json")
    with open(out + ".error", "w", encoding="utf-8") as stream:
        stream.write(error)
    ida_kernwin.msg(error + "\n")
    ida_pro.qexit(1)
else:
    ida_pro.qexit(0)
