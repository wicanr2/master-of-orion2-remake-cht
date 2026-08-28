"""唯讀匯出 Stealthy Ships 的 AI profile accumulator 與科技估值 category 資料流。"""

import csv
import hashlib
import json
import os
import traceback

import ida_auto
import ida_bytes
import ida_funcs
import ida_gdl
import ida_kernwin
import ida_name
import ida_pro
import ida_ua
import idautils
import idc


ROOTS = {"init_npc_profiles": 0x589D6, "calc_tech_value": 0xFC845}
HITS = {"profile_trait_hit": 0x58D53, "tech_trait_hit": 0xFCDDB}
TECH_ITEMS = 0x17E07F
TECH_ITEM_COUNT = 212
TECH_ITEM_STRIDE = 13
TECH_CATEGORIES = 0x17D196
TECH_CATEGORY_COUNT = 49


def digest(path):
    value = hashlib.sha256()
    with open(path, "rb") as source:
        for block in iter(lambda: source.read(1024 * 1024), b""):
            value.update(block)
    return value.hexdigest()


def symbols(path):
    out = {}
    with open(path, "r", encoding="utf-8", newline="") as source:
        for row in csv.DictReader(source, delimiter="\t"):
            try:
                ea = int(row["ea"], 16)
            except (KeyError, TypeError, ValueError):
                continue
            out[ea] = row.get("name") or "<unnamed>"
    return out


def instruction(ea):
    decoded = ida_ua.insn_t()
    if ida_ua.decode_insn(decoded, ea) <= 0:
        return {"ea": f"0x{ea:X}", "error": "decode failed"}
    return {
        "ea": f"0x{ea:X}",
        "bytes": (ida_bytes.get_bytes(ea, decoded.size) or b"").hex(),
        "text": idc.generate_disasm_line(ea, 0) or "",
        "mnemonic": ida_ua.print_insn_mnem(ea),
        "operands": [idc.print_operand(ea, i) for i in range(8) if idc.print_operand(ea, i)],
        "code_refs_from": [f"0x{x:X}" for x in idautils.CodeRefsFrom(ea, 0)],
        "data_refs_from": [f"0x{x:X}" for x in idautils.DataRefsFrom(ea)],
    }


def block(block):
    return {
        "start": f"0x{block.start_ea:X}",
        "end": f"0x{block.end_ea:X}",
        "preds": [f"0x{x.start_ea:X}" for x in block.preds()],
        "succs": [f"0x{x.start_ea:X}" for x in block.succs()],
        "instructions": [instruction(ea) for ea in idautils.Heads(block.start_ea, block.end_ea)
                         if ida_bytes.is_code(ida_bytes.get_flags(ea))],
    }


def function_record(ea, external):
    fn = ida_funcs.get_func(ea)
    if fn is None:
        return {"requested": f"0x{ea:X}", "error": "no function"}
    chunks = []
    for start, end in idautils.Chunks(fn.start_ea):
        chunks.append({
            "start": f"0x{start:X}", "end": f"0x{end:X}",
            "instructions": [instruction(x) for x in idautils.Heads(start, end)
                             if ida_bytes.is_code(ida_bytes.get_flags(x))],
        })
    return {
        "requested": f"0x{ea:X}", "start": f"0x{fn.start_ea:X}", "end": f"0x{fn.end_ea:X}",
        "ida_name": ida_name.get_name(fn.start_ea) or "<unnamed>",
        "external_symbol": external.get(fn.start_ea),
        "callers": [{"from": f"0x{x.frm:X}", "instruction": instruction(x.frm)}
                    for x in idautils.XrefsTo(fn.start_ea, 0) if x.iscode],
        "chunks": chunks,
    }


def hit_record(ea):
    fn = ida_funcs.get_func(ea)
    if fn is None:
        return {"ea": f"0x{ea:X}", "error": "no function"}
    selected = None
    all_blocks = list(ida_gdl.FlowChart(fn, flags=ida_gdl.FC_PREDS))
    for item in all_blocks:
        if item.start_ea <= ea < item.end_ea:
            selected = item
            break
    if selected is None:
        return {"ea": f"0x{ea:X}", "error": "no block"}
    neighboring = {selected.start_ea: selected}
    for pred in selected.preds():
        neighboring[pred.start_ea] = pred
    for succ in selected.succs():
        neighboring[succ.start_ea] = succ
    return {
        "ea": f"0x{ea:X}",
        "instruction": instruction(ea),
        "block": block(selected),
        "neighbor_blocks": [block(x) for _, x in sorted(neighboring.items())],
        "xrefs_to_block_start": [
            {"from": f"0x{x.frm:X}", "instruction": instruction(x.frm)}
            for x in idautils.XrefsTo(selected.start_ea, 0) if x.iscode
        ],
    }


def tech_items():
    rows = []
    for tech_id in range(TECH_ITEM_COUNT):
        ea = TECH_ITEMS + tech_id * TECH_ITEM_STRIDE
        raw = ida_bytes.get_bytes(ea, TECH_ITEM_STRIDE) or b""
        rows.append({
            "tech_id": tech_id,
            "ea": f"0x{ea:X}",
            "bytes": raw.hex(),
            "topic_word": ida_bytes.get_word(ea),
            "category_byte": ida_bytes.get_byte(ea + 3),
        })
    return rows


def tech_categories():
    return [
        {
            "category": category,
            "ea": f"0x{TECH_CATEGORIES + category * 2:X}",
            "value_byte": ida_bytes.get_byte(TECH_CATEGORIES + category * 2),
            "flag_byte": ida_bytes.get_byte(TECH_CATEGORIES + category * 2 + 1),
        }
        for category in range(TECH_CATEGORY_COUNT)
    ]


def main():
    ida_auto.auto_wait()
    source = os.environ["MOO2_RE_SOURCE"]
    database = os.environ["MOO2_RE_DATABASE"]
    symbols_path = os.environ["MOO2_RE_SYMBOLS"]
    out = os.environ["MOO2_RE_OUT"]
    external = symbols(symbols_path)
    payload = {
        "schema": "moo2-stealthy-profile-tech-evidence-v1",
        "inputs": {
            "source": {"name": os.path.basename(source), "sha256": digest(source)},
            "database": {"name": os.path.basename(database), "sha256": digest(database)},
            "symbols": {"name": os.path.basename(symbols_path), "sha256": digest(symbols_path)},
            "ida_version": ida_kernwin.get_kernel_version(),
            "address_space": "IDA linear address in Orion2.exe.i64",
        },
        "mutation": "none; read-only export",
        "warning": "external names and automated category rows are navigation; semantic promotion requires control/data-flow review",
        "roots": {name: function_record(ea, external) for name, ea in ROOTS.items()},
        "hits": {name: hit_record(ea) for name, ea in HITS.items()},
        "tech_item_table": {
            "base": f"0x{TECH_ITEMS:X}", "stride": TECH_ITEM_STRIDE,
            "count": TECH_ITEM_COUNT, "rows": tech_items(),
        },
        "tech_category_table": {
            "base": f"0x{TECH_CATEGORIES:X}", "stride": 2,
            "count": TECH_CATEGORY_COUNT, "rows": tech_categories(),
        },
    }
    with open(out, "w", encoding="utf-8") as target:
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
