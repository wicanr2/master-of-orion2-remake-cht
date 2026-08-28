"""非破壞性匯出 MOO2 1.31 戰略戰鬥模式旗標的全部直接讀寫端。"""

import csv
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


MODE_FLAG = 0x199CB4


def digest(path):
    result = hashlib.sha256()
    with open(path, "rb") as source:
        for chunk in iter(lambda: source.read(1024 * 1024), b""):
            result.update(chunk)
    return result.hexdigest()


def load_symbols(path):
    result = {}
    with open(path, newline="", encoding="utf-8") as source:
        for row in csv.DictReader(source, delimiter="\t"):
            result.setdefault(int(row["ea"], 16), []).append(row["name"])
    return result, digest(path)


def instruction(ea):
    decoded = ida_ua.insn_t()
    size = ida_ua.decode_insn(decoded, ea) or 1
    return {
        "ea": f"0x{ea:X}",
        "bytes": (ida_bytes.get_bytes(ea, size) or b"").hex(),
        "text": idc.generate_disasm_line(ea, 0) or "",
        "mnem": idc.print_insn_mnem(ea),
        "op0": idc.print_operand(ea, 0),
        "op1": idc.print_operand(ea, 1),
        "code_refs": [f"0x{x:X}" for x in idautils.CodeRefsFrom(ea, 0)],
        "data_refs": [f"0x{x:X}" for x in idautils.DataRefsFrom(ea)],
    }


def function_record(start_ea, symbols):
    function = ida_funcs.get_func(start_ea)
    if function is None:
        return {"requested": f"0x{start_ea:X}", "error": "function missing"}
    items = list(idautils.FuncItems(function.start_ea))
    raw = ida_bytes.get_bytes(function.start_ea, function.end_ea - function.start_ea) or b""
    pseudocode = None
    if ida_hexrays.init_hexrays_plugin():
        try:
            pseudocode = str(ida_hexrays.decompile(function.start_ea))
        except Exception as error:
            pseudocode = f"<decompile failed: {error}>"
    return {
        "start_ea": f"0x{function.start_ea:X}",
        "end_ea": f"0x{function.end_ea:X}",
        "original_name": ida_name.get_name(function.start_ea) or "<unnamed>",
        "external_names_navigation_only": symbols.get(function.start_ea, []),
        "bytes_sha256": hashlib.sha256(raw).hexdigest(),
        "pseudocode_navigation_only": pseudocode,
        "instructions": [instruction(ea) for ea in items],
    }


def mode_flag_sites(symbols):
    rows = []
    seen = set()
    refs = list(idautils.XrefsTo(MODE_FLAG, 0))
    for ref in refs:
        owner = ida_funcs.get_func(ref.frm)
        if owner is None:
            continue
        key = (owner.start_ea, ref.frm)
        if key in seen:
            continue
        seen.add(key)
        items = list(idautils.FuncItems(owner.start_ea))
        try:
            index = items.index(ref.frm)
        except ValueError:
            continue
        rows.append({
            "reference_ea": f"0x{ref.frm:X}",
            "xref_type": ref.type,
            "owner_start": f"0x{owner.start_ea:X}",
            "owner_original_name": ida_name.get_name(owner.start_ea) or "<unnamed>",
            "owner_external_names_navigation_only": symbols.get(owner.start_ea, []),
            "instruction": instruction(ref.frm),
            "window": [instruction(ea) for ea in items[max(0, index - 16):min(len(items), index + 17)]],
        })
    return rows


def callers(start_ea, symbols):
    rows = []
    function = ida_funcs.get_func(start_ea)
    if function is None:
        return rows
    for ref in idautils.XrefsTo(function.start_ea, 0):
        owner = ida_funcs.get_func(ref.frm)
        if owner is None:
            continue
        items = list(idautils.FuncItems(owner.start_ea))
        if ref.frm not in items:
            continue
        index = items.index(ref.frm)
        rows.append({
            "call_ea": f"0x{ref.frm:X}",
            "caller_start": f"0x{owner.start_ea:X}",
            "caller_original_name": ida_name.get_name(owner.start_ea) or "<unnamed>",
            "caller_external_names_navigation_only": symbols.get(owner.start_ea, []),
            "window": [instruction(ea) for ea in items[max(0, index - 12):min(len(items), index + 13)]],
        })
    return rows


def main():
    ida_auto.auto_wait()
    source = os.environ["MOO2_RE_SOURCE"]
    database = os.environ["MOO2_RE_DATABASE"]
    symbols, symbols_hash = load_symbols(os.environ["MOO2_RE_SYMBOLS"])
    sites = mode_flag_sites(symbols)
    owner_starts = sorted({int(site["owner_start"], 16) for site in sites})
    report = {
        "schema": "moo2.ida.strategic-combat-mode.v1",
        "evidence_scope": "static_only",
        "mutation": "none",
        "semantic_status": "reviewed_against_raw_instructions",
        "tool": {
            "name": "IDA Pro",
            "version": ida_kernwin.get_kernel_version(),
            "script": "tools/ida/audit_strategic_combat_mode.py",
        },
        "input": {
            "file": os.path.basename(source),
            "source_sha256": digest(source),
            "database_sha256": digest(database),
            "symbols_sha256": symbols_hash,
            "processor": ida_ida.inf_get_procname(),
        },
        "address_basis": "IDA linear; DOS/4GW LE object #1",
        "mode_flag": {
            "ea": f"0x{MODE_FLAG:X}",
            "original_name": ida_name.get_name(MODE_FLAG) or "<unnamed>",
            "raw_byte": (ida_bytes.get_bytes(MODE_FLAG, 1) or b"").hex(),
            "direct_sites": sites,
        },
        "owner_functions": {
            f"0x{start:X}": {
                "function": function_record(start, symbols),
                "direct_callsites": callers(start, symbols),
            }
            for start in owner_starts
        },
    }
    with open(os.environ["MOO2_RE_OUT"], "w", encoding="utf-8") as target:
        json.dump(report, target, ensure_ascii=False, indent=2)
        target.write("\n")


try:
    main()
except Exception:
    output = os.environ.get("MOO2_RE_OUT", "/tmp/strategic-combat-mode.json")
    with open(output + ".error", "w", encoding="utf-8") as target:
        target.write(traceback.format_exc())
finally:
    ida_pro.qexit(0)
