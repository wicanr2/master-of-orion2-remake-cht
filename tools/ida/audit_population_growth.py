"""非破壞性匯出 MOO2 1.31 人口成長 producer、逐族池與回寫 consumer。"""

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


ROOTS = {
    "raw_event_check_space_anomaly": 0x2341E,
    "raw_event_check_plague": 0x234B8,
    "raw_event_check_population_boom": 0x23509,
    "raw_event_check_colony_researching": 0x23DFE,
    "raw_colony_officer": 0xDD9F2,
    "raw_colony_race_pop_limit": 0xE0C1D,
    "raw_recompute_race_population_growth": 0xE1839,
    "raw_apply_colony_pop_growth": 0xE2DCA,
    "raw_growth_rule_helper_16946E": 0x16946E,
    "raw_growth_rule_helper_F4B81": 0xF4B81,
    "raw_star_zero_flag_initializer": 0x169245,
    "raw_star_zero_flag_source_A": 0x1691A0,
    "raw_star_zero_flag_source_B": 0x16945B,
    "raw_do_change_star_name": 0x922C2,
    "raw_owned_officer_level": 0x94951,
    "raw_integer_square_root": 0x134C92,
    "raw_weighted_choice_int": 0xFE96F,
    "raw_shuffle_sint": 0xFE9F5,
}

GROWTH_OFFSETS = ("+0B4h", "+0C8h", "+0Ah", "+0Ch", "+1D8h")
RAW_GLOBAL_OPTIONS = {
    "raw_base_plus_2": 0x1784DF,
    "raw_base_plus_278": 0x178755,
    "raw_base_plus_2DB": 0x1787B8,
    "raw_base_plus_2E4": 0x1787C1,
}
RAW_GLOBAL_OPTION_OPERANDS = ("+278h", "+2DBh", "+2E4h")


def digest(path):
    h = hashlib.sha256()
    with open(path, "rb") as source:
        for chunk in iter(lambda: source.read(1024 * 1024), b""):
            h.update(chunk)
    return h.hexdigest()


def load_symbols(path):
    result = {}
    with open(path, newline="", encoding="utf-8") as source:
        for row in csv.DictReader(source, delimiter="\t"):
            result.setdefault(int(row["ea"], 16), []).append(row["name"])
    return result, digest(path)


def instruction(ea):
    insn = ida_ua.insn_t()
    size = ida_ua.decode_insn(insn, ea) or 1
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


def function_record(ea, symbols):
    function = ida_funcs.get_func(ea)
    if function is None:
        return {"requested": f"0x{ea:X}", "error": "function missing"}
    items = list(idautils.FuncItems(function.start_ea))
    raw = ida_bytes.get_bytes(function.start_ea, function.end_ea - function.start_ea) or b""
    pseudo = None
    if ida_hexrays.init_hexrays_plugin():
        try:
            pseudo = str(ida_hexrays.decompile(function.start_ea))
        except Exception as error:
            pseudo = f"<decompile failed: {error}>"
    return {
        "requested": f"0x{ea:X}",
        "start_ea": f"0x{function.start_ea:X}",
        "end_ea": f"0x{function.end_ea:X}",
        "original_name": ida_name.get_name(function.start_ea) or "<unnamed>",
        "external_names_navigation_only": symbols.get(function.start_ea, []),
        "bytes_sha256": hashlib.sha256(raw).hexdigest(),
        "pseudocode_navigation_only": pseudo,
        "instructions": [instruction(item) for item in items],
    }


def root_callsites(ea, symbols):
    function = ida_funcs.get_func(ea)
    if function is None:
        return []
    rows = []
    for ref in idautils.XrefsTo(function.start_ea, 0):
        owner = ida_funcs.get_func(ref.frm)
        if owner is None:
            continue
        items = list(idautils.FuncItems(owner.start_ea))
        index = items.index(ref.frm)
        rows.append({
            "call_ea": f"0x{ref.frm:X}",
            "caller_start": f"0x{owner.start_ea:X}",
            "caller_original_name": ida_name.get_name(owner.start_ea) or "<unnamed>",
            "caller_external_names_navigation_only": symbols.get(owner.start_ea, []),
            "window": [instruction(x) for x in items[max(0, index - 16):min(len(items), index + 9)]],
        })
    return rows


def growth_operand_sites():
    rows = []
    for function_ea in idautils.Functions():
        function = ida_funcs.get_func(function_ea)
        if function is None:
            continue
        hits = []
        for ea in idautils.FuncItems(function.start_ea):
            operands = f"{idc.print_operand(ea, 0)} {idc.print_operand(ea, 1)}"
            if any(pattern.lower() in operands.lower() for pattern in GROWTH_OFFSETS):
                hits.append(instruction(ea))
        if hits:
            rows.append({
                "owner_start": f"0x{function.start_ea:X}",
                "owner_original_name": ida_name.get_name(function.start_ea) or "<unnamed>",
                "hits": hits,
            })
    return rows


def star_zero_reserved_flag_helpers(symbols):
    """匯出 0x169xxx out-of-line fragments，追 star #0 name-reserved byte +0x0E。"""
    rows = []
    seen = set()
    for function_ea in idautils.Functions(0x169000, 0x169500):
        function = ida_funcs.get_func(function_ea)
        if function is None or function.start_ea in seen:
            continue
        items = list(idautils.FuncItems(function.start_ea))
        lines = [idc.generate_disasm_line(ea, 0) or "" for ea in items]
        if not any("+0Eh" in line for line in lines):
            continue
        seen.add(function.start_ea)
        rows.append({
            "function": function_record(function.start_ea, symbols),
            "direct_callsites": root_callsites(function.start_ea, symbols),
        })
    return rows


def global_option_refs(symbols):
    found = {}
    for name, target in RAW_GLOBAL_OPTIONS.items():
        rows = []
        for ref in idautils.XrefsTo(target, 0):
            owner = ida_funcs.get_func(ref.frm)
            if owner is None:
                continue
            items = list(idautils.FuncItems(owner.start_ea))
            index = items.index(ref.frm)
            rows.append({
                "reference_ea": f"0x{ref.frm:X}",
                "owner_start": f"0x{owner.start_ea:X}",
                "owner_original_name": ida_name.get_name(owner.start_ea) or "<unnamed>",
                "owner_external_names_navigation_only": symbols.get(owner.start_ea, []),
                "window": [instruction(x) for x in items[max(0, index - 14):min(len(items), index + 15)]],
            })
        found[name] = {"target_ea": f"0x{target:X}", "references": rows}
    return found


def global_option_operand_sites(symbols):
    rows = []
    for function_ea in idautils.Functions():
        function = ida_funcs.get_func(function_ea)
        if function is None:
            continue
        items = list(idautils.FuncItems(function.start_ea))
        for index, ea in enumerate(items):
            operands = f"{idc.print_operand(ea, 0)} {idc.print_operand(ea, 1)}"
            if not any(pattern.lower() in operands.lower() for pattern in RAW_GLOBAL_OPTION_OPERANDS):
                continue
            rows.append({
                "reference_ea": f"0x{ea:X}",
                "owner_start": f"0x{function.start_ea:X}",
                "owner_original_name": ida_name.get_name(function.start_ea) or "<unnamed>",
                "owner_external_names_navigation_only": symbols.get(function.start_ea, []),
                "window": [instruction(x) for x in items[max(0, index - 12):min(len(items), index + 13)]],
            })
    return rows


def main():
    ida_auto.auto_wait()
    source = os.environ["MOO2_RE_SOURCE"]
    database = os.environ["MOO2_RE_DATABASE"]
    symbols, symbols_hash = load_symbols(os.environ["MOO2_RE_SYMBOLS"])
    report = {
        "schema": "moo2.ida.population-growth.v1",
        "evidence_scope": "static_only",
        "mutation": "none",
        "semantic_status": "reviewed_against_raw_instructions",
        "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version(),
                 "script": "tools/ida/audit_population_growth.py"},
        "input": {"file": os.path.basename(source), "source_sha256": digest(source),
                  "database_sha256": digest(database), "symbols_sha256": symbols_hash,
                  "processor": ida_ida.inf_get_procname()},
        "address_basis": "IDA linear; DOS/4GW LE object #1",
        "roots": {name: function_record(ea, symbols) for name, ea in ROOTS.items()},
        "root_direct_callsites": {name: root_callsites(ea, symbols) for name, ea in ROOTS.items()},
        "growth_operand_sites": growth_operand_sites(),
        "star_zero_reserved_flag_helpers": star_zero_reserved_flag_helpers(symbols),
        "global_option_refs": global_option_refs(symbols),
        "global_option_operand_sites": global_option_operand_sites(symbols),
        "raw_global_option_bytes": {
            name: {"ea": f"0x{ea:X}", "bytes": (ida_bytes.get_bytes(ea, 1) or b"").hex(),
                   "ida_name": ida_name.get_name(ea) or "<unnamed>"}
            for name, ea in RAW_GLOBAL_OPTIONS.items()
        },
    }
    with open(os.environ["MOO2_RE_OUT"], "w", encoding="utf-8") as target:
        json.dump(report, target, ensure_ascii=False, indent=2)
        target.write("\n")


try:
    main()
except Exception:
    error = traceback.format_exc()
    output = os.environ.get("MOO2_RE_OUT", "/tmp/population-growth.json")
    with open(output + ".error", "w", encoding="utf-8") as target:
        target.write(error)
    ida_kernwin.msg(error + "\n")
    ida_pro.qexit(1)
else:
    ida_pro.qexit(0)
