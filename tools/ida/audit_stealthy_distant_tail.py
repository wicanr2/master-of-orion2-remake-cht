"""唯讀匯出 MOO2 Stealthy Ships 遠端尾區塊的控制流與原始定位證據。"""

import bisect
import csv
import hashlib
import json
import os
import traceback

import ida_auto
import ida_bytes
import ida_funcs
import ida_gdl
import ida_ida
import ida_kernwin
import ida_name
import ida_pro
import ida_ua
import idautils
import idc


ROOT = 0x5D953
TAIL_HIT = 0xF4D5E
RETURN_TARGET = 0x5DC77


def digest(path):
    result = hashlib.sha256()
    with open(path, "rb") as source:
        for block in iter(lambda: source.read(1024 * 1024), b""):
            result.update(block)
    return result.hexdigest()


def load_symbols(path):
    records = {}
    with open(path, "r", encoding="utf-8", newline="") as source:
        for row in csv.DictReader(source, delimiter="\t"):
            try:
                ea = int(row["ea"], 16)
            except (KeyError, TypeError, ValueError):
                continue
            records[ea] = {
                "name": row.get("name") or "<unnamed>",
                "type": row.get("type") or "",
                "segment": row.get("seg") or "",
                "module": row.get("module") or "",
            }
    return records


def instruction(ea):
    decoded = ida_ua.insn_t()
    size = ida_ua.decode_insn(decoded, ea)
    if size <= 0:
        return {"ea": f"0x{ea:X}", "error": "decode failed"}
    return {
        "ea": f"0x{ea:X}",
        "bytes": (ida_bytes.get_bytes(ea, decoded.size) or b"").hex(),
        "text": idc.generate_disasm_line(ea, 0) or "",
        "mnemonic": ida_ua.print_insn_mnem(ea),
        "operands": [idc.print_operand(ea, i) for i in range(8) if idc.print_operand(ea, i)],
    }


def instructions(start, end):
    return [instruction(ea) for ea in idautils.Heads(start, end) if ida_bytes.is_code(ida_bytes.get_flags(ea))]


def xrefs_to(ea):
    return [
        {
            "from": f"0x{x.frm:X}",
            "to": f"0x{x.to:X}",
            "type": x.type,
            "is_code": bool(x.iscode),
            "instruction": instruction(x.frm),
        }
        for x in idautils.XrefsTo(ea, 0)
    ]


def xrefs_from(ea):
    return [
        {
            "from": f"0x{x.frm:X}",
            "to": f"0x{x.to:X}",
            "type": x.type,
            "is_code": bool(x.iscode),
        }
        for x in idautils.XrefsFrom(ea, 0)
    ]


def symbol_context(ea, symbols):
    keys = sorted(symbols)
    at = bisect.bisect_right(keys, ea) - 1
    before = keys[at] if at >= 0 else None
    after = keys[at + 1] if at + 1 < len(keys) else None
    return {
        "exact": symbols.get(ea),
        "nearest_preceding": None if before is None else {"ea": f"0x{before:X}", "record": symbols[before]},
        "next": None if after is None else {"ea": f"0x{after:X}", "record": symbols[after]},
    }


def block_record(block):
    return {
        "start": f"0x{block.start_ea:X}",
        "end": f"0x{block.end_ea:X}",
        "preds": [f"0x{x.start_ea:X}" for x in block.preds()],
        "succs": [f"0x{x.start_ea:X}" for x in block.succs()],
        "instructions": instructions(block.start_ea, block.end_ea),
    }


def containing_blocks(fn, addresses):
    result = {}
    for block in ida_gdl.FlowChart(fn, flags=ida_gdl.FC_PREDS):
        for ea in addresses:
            if block.start_ea <= ea < block.end_ea:
                result[f"0x{ea:X}"] = block_record(block)
    return result


def main():
    ida_auto.auto_wait()
    source = os.environ["MOO2_RE_SOURCE"]
    database = os.environ["MOO2_RE_DATABASE"]
    symbols_path = os.environ["MOO2_RE_SYMBOLS"]
    out = os.environ["MOO2_RE_OUT"]
    symbols = load_symbols(symbols_path)
    fn = ida_funcs.get_func(ROOT)
    if fn is None:
        raise RuntimeError(f"no function at 0x{ROOT:X}")

    chunks = []
    for start, end in idautils.Chunks(fn.start_ea):
        incoming = []
        for ea in idautils.Heads(start, end):
            for xref in idautils.XrefsTo(ea, 0):
                if xref.iscode and not (start <= xref.frm < end):
                    incoming.append({
                        "target": f"0x{ea:X}",
                        "source": f"0x{xref.frm:X}",
                        "source_instruction": instruction(xref.frm),
                    })
        chunks.append({
            "start": f"0x{start:X}",
            "end": f"0x{end:X}",
            "bytes_sha256": hashlib.sha256(ida_bytes.get_bytes(start, end - start) or b"").hexdigest(),
            "incoming_code_edges_from_outside_chunk": incoming,
            "instructions": instructions(start, end),
        })

    tail_fn = ida_funcs.get_func(TAIL_HIT)
    payload = {
        "schema": "moo2-watcom-distant-tail-evidence-v1",
        "evidence_contract": {
            "ownership_and_field_semantics_are_separate": True,
            "automated_result_level": "candidate_pending_manual_register_provenance",
            "warning": "IDA owner／最近外部符號／線性鄰接均不可單獨證明區塊歸屬或欄位語意",
        },
        "inputs": {
            "source": {"name": os.path.basename(source), "sha256": digest(source)},
            "database": {"name": os.path.basename(database), "sha256": digest(database)},
            "symbols": {"name": os.path.basename(symbols_path), "sha256": digest(symbols_path)},
            "ida_version": ida_kernwin.get_kernel_version(),
            "address_space": "IDA linear address in Orion2.exe.i64",
        },
        "root": {
            "requested": f"0x{ROOT:X}",
            "ida_start": f"0x{fn.start_ea:X}",
            "ida_end": f"0x{fn.end_ea:X}",
            "ida_original_name": ida_name.get_name(fn.start_ea) or "<unnamed>",
            "external_symbol": symbol_context(fn.start_ea, symbols),
        },
        "tail": {
            "hit": f"0x{TAIL_HIT:X}",
            "ida_owner_start": None if tail_fn is None else f"0x{tail_fn.start_ea:X}",
            "ida_owner_name": None if tail_fn is None else (ida_name.get_name(tail_fn.start_ea) or "<unnamed>"),
            "external_symbol_context": symbol_context(TAIL_HIT, symbols),
            "instruction": instruction(TAIL_HIT),
            "xrefs_to_hit": xrefs_to(TAIL_HIT),
            "xrefs_from_hit": xrefs_from(TAIL_HIT),
        },
        "return_target": {
            "ea": f"0x{RETURN_TARGET:X}",
            "instruction": instruction(RETURN_TARGET),
            "xrefs_to": xrefs_to(RETURN_TARGET),
            "xrefs_from": xrefs_from(RETURN_TARGET),
        },
        "containing_blocks": containing_blocks(fn, [TAIL_HIT, RETURN_TARGET]),
        "chunks": chunks,
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
