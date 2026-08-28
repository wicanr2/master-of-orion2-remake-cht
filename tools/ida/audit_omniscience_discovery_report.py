"""唯讀匯出 MOO2 Omniscience discovery report、star+0x28／+0x34 producer 與 consumer。"""

import csv
import hashlib
import json
import os
import re
import traceback

import ida_auto
import ida_bytes
import ida_funcs
import ida_kernwin
import ida_name
import ida_pro
import ida_ua
import idautils
import idc


ROOTS = {
    "explored_new_star": 0xFD95A,
    "draw_special": 0xFD875,
    "new_system_discovery_popup": 0xC8556,
    "display_report_aux": 0xFE63E,
    "reset_one_shot_system_special": 0xE98F2,
    "do_system_discoveries": 0xE9927,
    "make_ship_arrive": 0xFFDDA,
}
STAR_BASE_GLOBAL = 0x19306C
FIELD_PATTERN = re.compile(r"(?:\+|-)\s*(?:28|34|49|3D)h\]", re.IGNORECASE)
STRIDE_PATTERN = re.compile(r"(?:\b71h\b|\b113\b)", re.IGNORECASE)


def digest(path):
    out = hashlib.sha256()
    with open(path, "rb") as source:
        for block in iter(lambda: source.read(1024 * 1024), b""):
            out.update(block)
    return out.hexdigest()


def load_symbols(path):
    out = {}
    with open(path, "r", encoding="utf-8", newline="") as source:
        for row in csv.DictReader(source, delimiter="\t"):
            try:
                ea = int(row["ea"], 16)
            except (KeyError, TypeError, ValueError):
                continue
            out[ea] = row.get("name") or "<unnamed>"
    return out


def insn(ea):
    decoded = ida_ua.insn_t()
    if ida_ua.decode_insn(decoded, ea) <= 0:
        return {"ea": f"0x{ea:X}", "error": "decode failed"}
    return {
        "ea": f"0x{ea:X}",
        "bytes": (ida_bytes.get_bytes(ea, decoded.size) or b"").hex(),
        "text": idc.generate_disasm_line(ea, 0) or "",
        "mnemonic": ida_ua.print_insn_mnem(ea),
        "operands": [idc.print_operand(ea, i) for i in range(8) if idc.print_operand(ea, i)],
        "data_refs_from": [f"0x{x:X}" for x in idautils.DataRefsFrom(ea)],
        "code_refs_from": [f"0x{x:X}" for x in idautils.CodeRefsFrom(ea, 0)],
    }


def context(items, index, radius=8):
    return [insn(ea) for ea in items[max(0, index - radius):min(len(items), index + radius + 1)]]


def function_record(ea, symbols):
    fn = ida_funcs.get_func(ea)
    if fn is None:
        return {"requested": f"0x{ea:X}", "error": "no function"}
    chunks = []
    for start, end in idautils.Chunks(fn.start_ea):
        chunks.append({
            "start": f"0x{start:X}",
            "end": f"0x{end:X}",
            "instructions": [insn(x) for x in idautils.Heads(start, end)
                             if ida_bytes.is_code(ida_bytes.get_flags(x))],
        })
    return {
        "requested": f"0x{ea:X}",
        "start": f"0x{fn.start_ea:X}",
        "end": f"0x{fn.end_ea:X}",
        "ida_name": ida_name.get_name(fn.start_ea) or "<unnamed>",
        "external_symbol": symbols.get(fn.start_ea),
        "callers": [
            {"call": f"0x{x.frm:X}", "instruction": insn(x.frm),
             "owner": None if ida_funcs.get_func(x.frm) is None else f"0x{ida_funcs.get_func(x.frm).start_ea:X}"}
            for x in idautils.XrefsTo(fn.start_ea, 0) if x.iscode
        ],
        "chunks": chunks,
    }


def star_field_candidates(symbols):
    records = []
    for start in idautils.Functions():
        items = list(idautils.FuncItems(start))
        if not items:
            continue
        texts = [idc.generate_disasm_line(ea, 0) or "" for ea in items]
        has_star_base = any(STAR_BASE_GLOBAL in list(idautils.DataRefsFrom(ea)) for ea in items)
        has_stride = any(STRIDE_PATTERN.search(text) for text in texts)
        if not (has_star_base and has_stride):
            continue
        for index, text in enumerate(texts):
            if not FIELD_PATTERN.search(text):
                continue
            records.append({
                "function_start": f"0x{start:X}",
                "ida_name": ida_name.get_name(start) or "<unnamed>",
                "external_symbol": symbols.get(start),
                "has_star_base_data_ref": has_star_base,
                "has_0x71_stride": has_stride,
                "candidate": insn(items[index]),
                "context": context(items, index),
                "evidence_level": "candidate_pending_manual_base_provenance",
            })
    return records


def main():
    ida_auto.auto_wait()
    source = os.environ["MOO2_RE_SOURCE"]
    database = os.environ["MOO2_RE_DATABASE"]
    symbols_path = os.environ["MOO2_RE_SYMBOLS"]
    out_path = os.environ["MOO2_RE_OUT"]
    symbols = load_symbols(symbols_path)
    payload = {
        "schema": "moo2-omniscience-discovery-report-evidence-v1",
        "inputs": {
            "source": {"name": os.path.basename(source), "sha256": digest(source)},
            "database": {"name": os.path.basename(database), "sha256": digest(database)},
            "symbols": {"name": os.path.basename(symbols_path), "sha256": digest(symbols_path)},
            "ida_version": ida_kernwin.get_kernel_version(),
            "address_space": "IDA linear address in Orion2.exe.i64",
        },
        "mutation": "none; read-only export",
        "warning": "field candidates require manual star-base provenance; names are navigation only",
        "roots": {name: function_record(ea, symbols) for name, ea in ROOTS.items()},
        "star_field_candidates": star_field_candidates(symbols),
    }
    with open(out_path, "w", encoding="utf-8") as target:
        json.dump(payload, target, ensure_ascii=False, indent=2)
        target.write("\n")


if __name__ == "__main__":
    try:
        main()
    except Exception:
        traceback.print_exc()
        raise
    finally:
        ida_pro.qexit(0)
