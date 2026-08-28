"""非破壞性匯出 MOO2 1.31 戰略轟炸的完整快速戰鬥與殖民防禦鏈。"""

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
    "raw_qload_ships": 0x416CF,
    "raw_strategic_beam_attack": 0x41F80,
    "raw_strategic_special_attack": 0x420C0,
    "raw_strategic_missile_attack": 0x4221F,
    "raw_strategic_beam_attack_count": 0x57651,
    "raw_apply_strategic_damage": 0x40C2A,
    "raw_pick_strategic_target": 0x41E88,
    "raw_get_colony_hits": 0x42371,
    "raw_strategic_bombardment": 0x4257E,
    "raw_build_bombardment_result": 0x4267B,
    "raw_search_battles": 0xE9D62,
    "raw_do_one_combat": 0xE938C,
    "raw_colony_battle_resolution": 0xE87D2,
    "raw_apply_battle_ship_losses": 0xF975C,
    "raw_colony_damage_dispatch": 0xDD2F2,
    "raw_special_damage_pool": 0xDD13E,
    "raw_general_colony_damage": 0xDCEBD,
    "raw_remove_building": 0x145EA,
    "raw_destroy_colony_defense": 0x3A19E,
    "raw_apply_damage_to_planet": 0x3A3C3,
    "raw_load_colony_defense": 0x4A9E9,
    "raw_load_tactical_colony": 0x4AA36,
}

DATA_RANGES = {
    "strategic_record_type_table": (0x18002E, 9 * 36),
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
    return {"ea": f"0x{ea:X}", "bytes": (ida_bytes.get_bytes(ea, size) or b"").hex(),
            "text": idc.generate_disasm_line(ea, 0) or "", "mnem": idc.print_insn_mnem(ea),
            "op0": idc.print_operand(ea, 0), "op1": idc.print_operand(ea, 1),
            "code_refs": [f"0x{x:X}" for x in idautils.CodeRefsFrom(ea, 0)],
            "data_refs": [f"0x{x:X}" for x in idautils.DataRefsFrom(ea)]}


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
    return {"requested": f"0x{ea:X}", "start_ea": f"0x{function.start_ea:X}",
            "end_ea": f"0x{function.end_ea:X}",
            "original_name": ida_name.get_name(function.start_ea) or "<unnamed>",
            "external_names_navigation_only": symbols.get(function.start_ea, []),
            "bytes_sha256": hashlib.sha256(raw).hexdigest(),
            "pseudocode_navigation_only": pseudocode,
            "instructions": [instruction(item) for item in items]}


def graph(ea, symbols):
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
                          "window": [instruction(x) for x in items[max(0,index-16):index+17]]})
    return {"calls": calls, "callsites": callsites}


def main():
    ida_auto.auto_wait()
    source = os.environ["MOO2_RE_SOURCE"]
    database = os.environ["MOO2_RE_DATABASE"]
    symbols, symbols_hash = load_symbols(os.environ["MOO2_RE_SYMBOLS"])
    report = {
        "schema": "moo2.ida.strategic-bombardment-full.v1",
        "evidence_scope": "static_only", "mutation": "none",
        "semantic_status": "reviewed_against_raw_instructions",
        "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version(),
                 "script": "tools/ida/audit_strategic_bombardment_full.py"},
        "input": {"file": os.path.basename(source), "source_sha256": digest(source),
                  "database_sha256": digest(database), "symbols_sha256": symbols_hash,
                  "processor": ida_ida.inf_get_procname()},
        "address_basis": "IDA linear; DOS/4GW LE object #1",
        "roots": {},
        "data_ranges": {
            name: {"start_ea": f"0x{ea:X}", "length": length,
                   "bytes": (ida_bytes.get_bytes(ea, length) or b"").hex()}
            for name, (ea, length) in DATA_RANGES.items()
        },
    }
    for name, ea in ROOTS.items():
        report["roots"][name] = {"function": function_record(ea, symbols), **graph(ea, symbols)}
    with open(os.environ["MOO2_RE_OUT"], "w", encoding="utf-8") as target:
        json.dump(report, target, ensure_ascii=False, indent=2)
        target.write("\n")


try:
    main()
except Exception:
    output = os.environ.get("MOO2_RE_OUT", "/tmp/strategic-bombardment-full.json")
    with open(output + ".error", "w", encoding="utf-8") as target:
        target.write(traceback.format_exc())
finally:
    ida_pro.qexit(0)
