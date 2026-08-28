"""非破壞性匯出領袖回合、ETA callback 與 AI 任命資料流。"""

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
    "raw_check_officer_level": 0x92FDA,
    "raw_deassign_officer": 0x933F2,
    "raw_decrement_officer_eta": 0x934CF,
    "raw_get_officer_base_level": 0x93D4B,
    "raw_set_officer_to_player": 0x9718F,
    "raw_chance_to_hire": 0x9781D,
    "raw_random_officer_check": 0x97A66,
    "raw_generate_random_officer": 0x97AD4,
    "raw_select_leader_for_hire": 0x97B2D,
    "raw_get_star_leader_eta": 0x98F42,
    "raw_assign_fleet_officer": 0xD6FDA,
    "raw_reassign_fleet_officer": 0xD7078,
    "raw_assign_star_officer": 0xD7171,
    "raw_assign_unassigned_leaders": 0xD73D4,
    "raw_do_ai_leaders": 0xD7439,
    "raw_colony_dominant_player": 0xD2A08,
    "raw_find_players_in_range": 0xD3A68,
    "raw_compute_star_player_range_info": 0xD3BA0,
    "raw_compute_ai_data": 0xD3D34,
    "raw_colony_race_pop_limit": 0xE0C1D,
    "raw_pass_out_imports": 0xDF8F0,
    "raw_pre_import_computing": 0xE1D59,
    "raw_update_player_stats": 0xE2710,
    "raw_star_colony_calculation": 0xE2AB1,
}

OFFICER_FIELDS = {
    "raw_type_plus_23": "+23h",
    "raw_experience_plus_24": "+24h",
    "raw_assignment_plus_34": "+34h",
    "raw_assignment_plus_35": "+35h",
    "raw_assignment_plus_36": "+36h",
    "raw_eta_plus_37": "+37h",
    "raw_status_plus_39": "+39h",
    "raw_owner_plus_3A": "+3Ah",
}

AI_CACHE_RANGES = {
    # 兩者是指向 heap buffer 的 dword globals，不是 inline array。只收精確
    # pointer slot，避免把後續全域資料誤列成同一快取的 reference。
    "raw_colony_ai_cache_pointer_1AA1EC": (0x1AA1EC, 0x1AA1F0),
    "raw_star_player_ai_cache_pointer_1AA1F8": (0x1AA1F8, 0x1AA1FC),
}


def digest(path):
    h = hashlib.sha256()
    with open(path, "rb") as source:
        for chunk in iter(lambda: source.read(1024 * 1024), b""):
            h.update(chunk)
    return h.hexdigest()


def load_symbols(path):
    result = {}
    if not path:
        return result, None
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


def callsite_record(ref_ea, symbols):
    owner = ida_funcs.get_func(ref_ea)
    if owner is None:
        return {"call_ea": f"0x{ref_ea:X}", "error": "caller missing"}
    items = list(idautils.FuncItems(owner.start_ea))
    index = items.index(ref_ea)
    return {
        "call_ea": f"0x{ref_ea:X}",
        "caller_start": f"0x{owner.start_ea:X}",
        "caller_original_name": ida_name.get_name(owner.start_ea) or "<unnamed>",
        "caller_external_names_navigation_only": symbols.get(owner.start_ea, []),
        "window": [instruction(x) for x in items[max(0, index - 24):min(len(items), index + 10)]],
    }


def root_callsites(ea, symbols):
    function = ida_funcs.get_func(ea)
    if function is None:
        return []
    return [
        callsite_record(ref.frm, symbols)
        for ref in idautils.XrefsTo(function.start_ea, 0)
        if ida_funcs.get_func(ref.frm) is not None
    ]


def officer_field_sites():
    found = {name: [] for name in OFFICER_FIELDS}
    for function_ea in idautils.Functions():
        function = ida_funcs.get_func(function_ea)
        if function is None:
            continue
        owner = ida_name.get_name(function.start_ea) or "<unnamed>"
        for ea in idautils.FuncItems(function.start_ea):
            operands = f"{idc.print_operand(ea, 0)} {idc.print_operand(ea, 1)}"
            for name, pattern in OFFICER_FIELDS.items():
                if pattern.lower() in operands.lower():
                    row = instruction(ea)
                    row["owner_start"] = f"0x{function.start_ea:X}"
                    row["owner_original_name"] = owner
                    found[name].append(row)
    return found


def ai_cache_reference_sites(symbols):
    """保留 raw data reference，並附帶所在函式與足夠的 producer/consumer 視窗。"""
    found = {name: [] for name in AI_CACHE_RANGES}
    for function_ea in idautils.Functions():
        function = ida_funcs.get_func(function_ea)
        if function is None:
            continue
        items = list(idautils.FuncItems(function.start_ea))
        owner = ida_name.get_name(function.start_ea) or "<unnamed>"
        for index, ea in enumerate(items):
            refs = list(idautils.DataRefsFrom(ea))
            for name, (start, end) in AI_CACHE_RANGES.items():
                matched = [ref for ref in refs if start <= ref < end]
                if not matched:
                    continue
                found[name].append({
                    "reference_ea": f"0x{ea:X}",
                    "referenced_addresses": [f"0x{x:X}" for x in matched],
                    "owner_start": f"0x{function.start_ea:X}",
                    "owner_original_name": owner,
                    "owner_external_names_navigation_only": symbols.get(function.start_ea, []),
                    "window": [instruction(x) for x in items[max(0, index - 20):min(len(items), index + 21)]],
                })
    return found


def main():
    ida_auto.auto_wait()
    source = os.environ["MOO2_RE_SOURCE"]
    database = os.environ["MOO2_RE_DATABASE"]
    symbols, symbols_hash = load_symbols(os.environ.get("MOO2_RE_SYMBOLS", ""))
    report = {
        "schema": "moo2.ida.leader-turn-chain.v1",
        "evidence_scope": "static_only",
        "mutation": "none",
        "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version(),
                 "script": "tools/ida/audit_leader_turn_chain.py"},
        "input": {"file": os.path.basename(source), "source_sha256": digest(source),
                  "database_sha256": digest(database), "symbols_sha256": symbols_hash,
                  "processor": ida_ida.inf_get_procname()},
        "address_basis": "IDA linear; DOS/4GW LE object #1",
        "semantic_status": "reviewed_against_raw_instructions",
        "roots": {name: function_record(ea, symbols) for name, ea in ROOTS.items()},
        "root_direct_callsites": {name: root_callsites(ea, symbols) for name, ea in ROOTS.items()},
        "officer_field_operand_sites": officer_field_sites(),
        "ai_cache_reference_sites": ai_cache_reference_sites(symbols),
    }
    with open(os.environ["MOO2_RE_OUT"], "w", encoding="utf-8") as target:
        json.dump(report, target, ensure_ascii=False, indent=2)
        target.write("\n")


try:
    main()
except Exception:
    error = traceback.format_exc()
    output = os.environ.get("MOO2_RE_OUT", "/tmp/leader-turn-chain.json")
    with open(output + ".error", "w", encoding="utf-8") as target:
        target.write(error)
    ida_kernwin.msg(error + "\n")
    ida_pro.qexit(1)
else:
    ida_pro.qexit(0)
