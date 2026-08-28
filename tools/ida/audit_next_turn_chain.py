"""唯讀匯出 MOO2 回合主鏈的固定呼叫順序與分支邊界。"""

import csv
import hashlib
import json
import os

import ida_auto
import ida_bytes
import ida_funcs
import ida_ida
import ida_kernwin
import ida_nalt
import ida_pro
import idautils
import idc


ROOT = 0x136B3
OUT = os.environ.get("MOO2_RE_OUT", "/out/next-turn-chain.json")
SOURCE = os.environ.get("MOO2_RE_SOURCE", ida_nalt.get_input_file_path())
SYMBOLS = os.environ.get("MOO2_RE_SYMBOLS", "")


def sha256(path):
    digest = hashlib.sha256()
    with open(path, "rb") as stream:
        for chunk in iter(lambda: stream.read(1024 * 1024), b""):
            digest.update(chunk)
    return digest.hexdigest()


def insn(ea):
    size = idc.get_item_size(ea)
    return {
        "ea": f"0x{ea:X}",
        "bytes": (ida_bytes.get_bytes(ea, size) or b"").hex(),
        "mnem": idc.print_insn_mnem(ea),
        "op0": idc.print_operand(ea, 0),
        "op1": idc.print_operand(ea, 1),
        "line": idc.generate_disasm_line(ea, 0) or "<unavailable>",
    }


def load_symbols(path):
    if not path:
        return {}, None
    symbols = {}
    with open(path, newline="", encoding="utf-8") as stream:
        for row in csv.DictReader(stream, delimiter="\t"):
            symbols.setdefault(int(row["ea"], 16), []).append(row["name"])
    return symbols, sha256(path)


def main():
    ida_auto.auto_wait()
    symbols, symbols_hash = load_symbols(SYMBOLS)
    function = ida_funcs.get_func(ROOT)
    if not function:
        raise RuntimeError(f"no function at 0x{ROOT:X}")

    instructions = [insn(ea) for ea in idautils.FuncItems(function.start_ea)]
    calls = []
    for order, ea in enumerate(
        (ea for ea in idautils.FuncItems(function.start_ea)
         if idc.print_insn_mnem(ea).lower() == "call"),
        start=1,
    ):
        target = idc.get_operand_value(ea, 0)
        target_function = ida_funcs.get_func(target)
        calls.append({
            "order": order,
            "call": insn(ea),
            "target": f"0x{target:X}",
            "target_function_start": f"0x{target_function.start_ea:X}" if target_function else None,
            "ida_original_name": idc.get_name(target) or "<unnamed>",
            "external_symbol_names_navigation_only": symbols.get(target, []),
        })

    report = {
        "schema": "moo2.ida.next-turn-chain.v1",
        "evidence_scope": "static_only",
        "mutation": "none",
        "tool": {
            "name": "IDA Pro",
            "version": ida_kernwin.get_kernel_version(),
            "script": "tools/ida/audit_next_turn_chain.py",
        },
        "input": {
            "database": ida_nalt.get_input_file_path(),
            "source": SOURCE,
            "source_sha256": sha256(SOURCE),
            "symbols": SYMBOLS or None,
            "symbols_sha256": symbols_hash,
            "processor": ida_ida.inf_get_procname(),
            "min_ea": f"0x{ida_ida.inf_get_min_ea():X}",
            "max_ea": f"0x{ida_ida.inf_get_max_ea():X}",
        },
        "address_basis": "IDA linear; DOS/4GW LE object #1",
        "root": {
            "requested": f"0x{ROOT:X}",
            "start": f"0x{function.start_ea:X}",
            "end": f"0x{function.end_ea:X}",
            "ida_original_name": idc.get_name(function.start_ea) or "<unnamed>",
            "external_symbol_names_navigation_only": symbols.get(function.start_ea, []),
            "instructions": instructions,
            "calls": calls,
        },
    }
    os.makedirs(os.path.dirname(OUT), exist_ok=True)
    with open(OUT, "w", encoding="utf-8") as stream:
        json.dump(report, stream, ensure_ascii=False, indent=2)
        stream.write("\n")
    ida_pro.qexit(0)


if __name__ == "__main__":
    main()
