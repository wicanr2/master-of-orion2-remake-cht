"""非破壞性匯出 MOO2 1.31 Fighter Garrison 格子戰術建立與回寫鏈。"""

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
    "raw_check_launched_fighters": 0x29480,
    "raw_fire_fighter_bomb": 0x3AC20,
    "raw_fire_fighter_beam": 0x3AD57,
    "raw_draw_fighter": 0x3DBD1,
    "raw_qload_ships": 0x416CF,
    "raw_get_colony_hits": 0x42371,
    "raw_tactical_combat": 0x47939,
    "raw_load_combat_ship": 0x4954A,
    "raw_load_colony_defense": 0x4A9E9,
    "raw_load_tactical_colony": 0x4AA36,
    "raw_load_combat_satellite": 0x4BBD5,
    "raw_planet_has_defenses": 0x3A142,
    "raw_destroy_colony_defense": 0x3A19E,
    "raw_apply_damage_to_planet": 0x3A3C3,
    "raw_deploy_ships": 0x49043,
    "raw_create_missile_runtime": 0x3C892,
    "raw_fighters_available": 0x5F968,
    "raw_add_fighters_to_design": 0x601AC,
    "raw_russ_combat": 0xE7343,
}

COLONY_PATTERNS = ("+165h", "+160h", "+151h", "+150h")


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


def function_record(ea, symbols):
    function = ida_funcs.get_func(ea)
    if function is None:
        return {"requested": f"0x{ea:X}", "error": "function missing"}
    items = list(idautils.FuncItems(function.start_ea))
    raw = ida_bytes.get_bytes(function.start_ea, function.end_ea - function.start_ea) or b""
    pseudocode = None
    if ida_hexrays.init_hexrays_plugin():
        try:
            pseudocode = str(ida_hexrays.decompile(function.start_ea))
        except Exception as error:
            pseudocode = f"<decompile failed: {error}>"
    return {
        "requested": f"0x{ea:X}",
        "start_ea": f"0x{function.start_ea:X}",
        "end_ea": f"0x{function.end_ea:X}",
        "original_name": ida_name.get_name(function.start_ea) or "<unnamed>",
        "external_names_navigation_only": symbols.get(function.start_ea, []),
        "bytes_sha256": hashlib.sha256(raw).hexdigest(),
        "pseudocode_navigation_only": pseudocode,
        "instructions": [instruction(item) for item in items],
    }


def direct_calls(ea, symbols):
    function = ida_funcs.get_func(ea)
    if function is None:
        return []
    rows = []
    for item in idautils.FuncItems(function.start_ea):
        for target in idautils.CodeRefsFrom(item, 0):
            callee = ida_funcs.get_func(target)
            if callee is None or callee.start_ea != target:
                continue
            rows.append({
                "call_ea": f"0x{item:X}",
                "target_ea": f"0x{target:X}",
                "target_original_name": ida_name.get_name(target) or "<unnamed>",
                "target_external_names_navigation_only": symbols.get(target, []),
                "instruction": instruction(item),
            })
    return rows


def direct_callsites(ea, symbols):
    function = ida_funcs.get_func(ea)
    if function is None:
        return []
    rows = []
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
            "window": [instruction(item) for item in items[max(0, index - 12):min(len(items), index + 13)]],
        })
    return rows


def colony_operand_sites(symbols):
    rows = []
    for function_ea in idautils.Functions():
        function = ida_funcs.get_func(function_ea)
        if function is None:
            continue
        items = list(idautils.FuncItems(function.start_ea))
        for index, ea in enumerate(items):
            operands = f"{idc.print_operand(ea, 0)} {idc.print_operand(ea, 1)}".lower()
            if not any(pattern.lower() in operands for pattern in COLONY_PATTERNS):
                continue
            rows.append({
                "reference_ea": f"0x{ea:X}",
                "owner_start": f"0x{function.start_ea:X}",
                "owner_original_name": ida_name.get_name(function.start_ea) or "<unnamed>",
                "owner_external_names_navigation_only": symbols.get(function.start_ea, []),
                "window": [instruction(item) for item in items[max(0, index - 10):min(len(items), index + 11)]],
            })
    return rows


def main():
    ida_auto.auto_wait()
    source = os.environ["MOO2_RE_SOURCE"]
    database = os.environ["MOO2_RE_DATABASE"]
    symbols, symbols_hash = load_symbols(os.environ["MOO2_RE_SYMBOLS"])
    report = {
        "schema": "moo2.ida.fighter-garrison-tactical.v1",
        "evidence_scope": "static_only",
        "mutation": "none",
        "semantic_status": "reviewed_against_raw_instructions",
        "tool": {
            "name": "IDA Pro",
            "version": ida_kernwin.get_kernel_version(),
            "script": "tools/ida/audit_fighter_garrison_tactical.py",
        },
        "input": {
            "file": os.path.basename(source),
            "source_sha256": digest(source),
            "database_sha256": digest(database),
            "symbols_sha256": symbols_hash,
            "processor": ida_ida.inf_get_procname(),
        },
        "address_basis": "IDA linear; DOS/4GW LE object #1",
        "roots": {
            name: {
                "function": function_record(ea, symbols),
                "direct_calls": direct_calls(ea, symbols),
                "direct_callsites": direct_callsites(ea, symbols),
            }
            for name, ea in ROOTS.items()
        },
        "colony_operand_sites": colony_operand_sites(symbols),
    }
    with open(os.environ["MOO2_RE_OUT"], "w", encoding="utf-8") as target:
        json.dump(report, target, ensure_ascii=False, indent=2)
        target.write("\n")


try:
    main()
except Exception:
    output = os.environ.get("MOO2_RE_OUT", "/tmp/fighter-garrison-tactical.json")
    with open(output + ".error", "w", encoding="utf-8") as target:
        target.write(traceback.format_exc())
finally:
    ida_pro.qexit(0)
