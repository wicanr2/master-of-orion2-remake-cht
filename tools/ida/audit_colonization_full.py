"""非破壞性匯出 MOO2 1.31 殖民／前哨站垂直鏈與 colony raw 欄位 consumer。"""

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
    "raw_init_colony": 0x12D75,
    "raw_planet_is_colonizable": 0x714A1,
    "raw_set_colonize_button_flag": 0x71F35,
    "raw_planet_has_outpost": 0x79C1D,
    "raw_player_has_outpost": 0x79D86,
    "raw_star_is_outpost_star": 0x7A133,
    "raw_planet_colonization_main_screen": 0x8B2DE,
    "raw_colony_ship_auto_colonization": 0x9952D,
    "raw_check_colony_outpost_ships": 0x99E07,
    "raw_planet_is_outpost_planet": 0x9AC0D,
    "raw_colonize_planet_ui": 0xBB082,
    "raw_info_decision_popups": 0xC035E,
    "raw_get_colonizable_planets": 0xD76B8,
    "raw_colony_calculation": 0xE2A70,
    "raw_compute_star_colony_stuff": 0xE5296,
    "raw_do_system_discoveries": 0xE9927,
    "raw_make_new_colony_or_outpost": 0xE5EB3,
    "raw_make_new_colony": 0xE6071,
    "raw_make_new_outpost": 0xE607F,
    "raw_star_colonizable_count": 0xE60C8,
    "raw_star_outpostable_count": 0xE6132,
    "raw_find_colonize_players": 0xE6170,
    "raw_ai_colonize": 0xE65F8,
    "raw_all_ai_colonize": 0xE67F6,
    "raw_remove_building": 0x145EA,
    "raw_player_colony_base_at_star": 0xFDA3F,
    "raw_player_ship_type_at_star": 0xFDAA7,
    "raw_colonization_dispatch": 0xFDB01,
}


RAW_RANGES = {
    "new_colony_outpost_shared_wrappers": (0xE6060, 0xE60C8),
}


def digest(path):
    value = hashlib.sha256()
    with open(path, "rb") as source:
        for chunk in iter(lambda: source.read(1024 * 1024), b""):
            value.update(chunk)
    return value.hexdigest()


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


def xrefs(ea, symbols):
    function = ida_funcs.get_func(ea)
    if function is None:
        return {"calls": [], "callsites": []}
    calls = []
    for item in idautils.FuncItems(function.start_ea):
        for target in idautils.CodeRefsFrom(item, 0):
            callee = ida_funcs.get_func(target)
            if callee is not None and callee.start_ea == target:
                calls.append({"call_ea": f"0x{item:X}", "target_ea": f"0x{target:X}",
                              "target_original_name": ida_name.get_name(target) or "<unnamed>",
                              "target_external_names_navigation_only": symbols.get(target, [])})
    callsites = []
    for ref in idautils.XrefsTo(function.start_ea, 0):
        owner = ida_funcs.get_func(ref.frm)
        if owner is None:
            continue
        items = list(idautils.FuncItems(owner.start_ea))
        if ref.frm not in items:
            continue
        index = items.index(ref.frm)
        callsites.append({"call_ea": f"0x{ref.frm:X}", "caller_start": f"0x{owner.start_ea:X}",
                          "caller_original_name": ida_name.get_name(owner.start_ea) or "<unnamed>",
                          "caller_external_names_navigation_only": symbols.get(owner.start_ea, []),
                          "window": [instruction(x) for x in items[max(0, index - 12):index + 13]]})
    return {"calls": calls, "callsites": callsites}


def operand_sites(patterns, symbols):
    rows = []
    seen = set()
    for segment_start in idautils.Segments():
        for ea in idautils.Heads(segment_start, idc.get_segm_end(segment_start)):
            if not ida_bytes.is_code(ida_bytes.get_flags(ea)):
                continue
            line = idc.generate_disasm_line(ea, 0) or ""
            if not any(pattern.lower() in line.lower() for pattern in patterns):
                continue
            owner = ida_funcs.get_func(ea)
            key = (ea, line)
            if key in seen:
                continue
            seen.add(key)
            rows.append({
                "instruction": instruction(ea),
                "owner_start": f"0x{owner.start_ea:X}" if owner else None,
                "owner_original_name": ida_name.get_name(owner.start_ea) if owner else None,
                "owner_external_names_navigation_only": symbols.get(owner.start_ea, []) if owner else [],
            })
    return rows


def raw_range(start, end):
    rows = []
    ea = ida_bytes.next_head(start - 1, end)
    while ea != idc.BADADDR and ea < end:
        rows.append(instruction(ea))
        ea = ida_bytes.next_head(ea, end)
    return {"start_ea": f"0x{start:X}", "end_ea_exclusive": f"0x{end:X}", "instructions": rows}


def main():
    ida_auto.auto_wait()
    source = os.environ["MOO2_RE_SOURCE"]
    database = os.environ["MOO2_RE_DATABASE"]
    symbols, symbols_hash = load_symbols(os.environ["MOO2_RE_SYMBOLS"])
    report = {
        "schema": "moo2.ida.colonization-full.v1",
        "evidence_scope": "static_only",
        "mutation": "none",
        "semantic_status": "reviewed_against_raw_instructions",
        "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version(),
                 "script": "tools/ida/audit_colonization_full.py"},
        "input": {"file": os.path.basename(source), "source_sha256": digest(source),
                  "database_sha256": digest(database), "symbols_sha256": symbols_hash,
                  "processor": ida_ida.inf_get_procname()},
        "address_basis": "IDA linear; DOS/4GW LE object #1",
        "roots": {},
        "relative_operand_sites": {
            "raw_colony_plus_0x123": operand_sites(("+123h]", "+123h,"), symbols),
            "raw_colony_plus_0x14c": operand_sites(("+14Ch]", "+14Ch,"), symbols),
        },
        "raw_ranges": {name: raw_range(start, end) for name, (start, end) in RAW_RANGES.items()},
    }
    for name, ea in ROOTS.items():
        report["roots"][name] = {"function": function_record(ea, symbols), **xrefs(ea, symbols)}
    with open(os.environ["MOO2_RE_OUT"], "w", encoding="utf-8") as target:
        json.dump(report, target, ensure_ascii=False, indent=2)
        target.write("\n")


try:
    main()
except Exception:
    output = os.environ.get("MOO2_RE_OUT", "/tmp/colonization-full.json")
    with open(output + ".error", "w", encoding="utf-8") as target:
        target.write(traceback.format_exc())
finally:
    ida_pro.qexit(0)
