"""非破壞性匯出 MOO2 1.31 艦隊派遣、星際移動、抵達與 ETA 鏈。"""

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
    "raw_fleet_screen_move_ships": 0x75D48,
    "raw_xy_to_star_id": 0x7A3E3,
    "raw_update_ship_icons_between_turns": 0x8A4C4,
    "raw_clear_all_auto_colonization": 0xA1C0D,
    "raw_move_ships_with_possible_intermediate": 0xD7923,
    "raw_ai_interceptions": 0xD9A7E,
    "raw_collect_black_holes": 0xEB7C5,
    "raw_black_hole_blocks_points": 0xEB7FD,
    "raw_initialize_black_hole_blocks": 0xEB87D,
    "raw_point_is_in_nebula_n": 0xEB9C8,
    "raw_point_is_in_some_nebula": 0xEBA3A,
    "raw_point_is_in_warp_interdictor": 0xEBAC3,
    "raw_move_player_one_turn_to_star": 0xEBB0C,
    "raw_ship_range": 0xFF496,
    "raw_star_in_range_aux1": 0xFF4E9,
    "raw_star_in_range_aux2": 0xFF593,
    "raw_star_in_range": 0xFF5F8,
    "raw_star_in_normal_range": 0xFF666,
    "raw_star_in_extended_range": 0xFF68A,
    "raw_can_order_ship": 0xFF6BE,
    "raw_ships_try_to_move_to": 0xFF799,
    "raw_player_turns_star_to_star": 0xFFBD6,
    "raw_make_ships_move_to": 0xFFD08,
    "raw_make_ship_arrive_at_star": 0xFFDDA,
    "raw_move_all_ships_toward_stars": 0xFFEEA,
    "raw_create_nonplayer_ship": 0x100010,
    "raw_update_player_ship_range": 0x10034D,
    "raw_update_all_ship_ranges": 0x10038C,
    "raw_ship_destination_star": 0x1003F2,
    "raw_update_ship_etas": 0x10041C,
    "raw_test_bit_field": 0x1276F0,
    "raw_unknown_movement_notifier": 0x16915C,
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
                calls.append({
                    "call_ea": f"0x{item:X}",
                    "target_ea": f"0x{target:X}",
                    "target_original_name": ida_name.get_name(target) or "<unnamed>",
                    "target_external_names_navigation_only": symbols.get(target, []),
                })
    callsites = []
    for ref in idautils.XrefsTo(function.start_ea, 0):
        owner = ida_funcs.get_func(ref.frm)
        if owner is None:
            continue
        items = list(idautils.FuncItems(owner.start_ea))
        if ref.frm not in items:
            continue
        index = items.index(ref.frm)
        callsites.append({
            "call_ea": f"0x{ref.frm:X}",
            "caller_start": f"0x{owner.start_ea:X}",
            "caller_original_name": ida_name.get_name(owner.start_ea) or "<unnamed>",
            "caller_external_names_navigation_only": symbols.get(owner.start_ea, []),
            "window": [instruction(item) for item in items[max(0, index - 12):index + 13]],
        })
    return {"calls": calls, "callsites": callsites}


def main():
    ida_auto.auto_wait()
    source = os.environ["MOO2_RE_SOURCE"]
    database = os.environ["MOO2_RE_DATABASE"]
    symbols, symbols_hash = load_symbols(os.environ["MOO2_RE_SYMBOLS"])
    report = {
        "schema": "moo2.ida.fleet-interstellar-movement.v1",
        "evidence_scope": "static_only",
        "mutation": "none",
        "semantic_status": "reviewed_against_raw_instructions",
        "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version(),
                 "script": "tools/ida/audit_fleet_interstellar_movement.py"},
        "input": {
            "file": os.path.basename(source),
            "source_sha256": digest(source),
            "database_sha256": digest(database),
            "symbols_sha256": symbols_hash,
            "processor": ida_ida.inf_get_procname(),
        },
        "address_basis": "IDA linear; DOS/4GW LE object #1",
        "roots": {},
    }
    for name, ea in ROOTS.items():
        report["roots"][name] = {"function": function_record(ea, symbols), **xrefs(ea, symbols)}
    with open(os.environ["MOO2_RE_OUT"], "w", encoding="utf-8") as target:
        json.dump(report, target, ensure_ascii=False, indent=2)
        target.write("\n")


try:
    main()
except Exception:
    output = os.environ.get("MOO2_RE_OUT", "/tmp/fleet-interstellar-movement.json")
    with open(output + ".error", "w", encoding="utf-8") as target:
        target.write(traceback.format_exc())
finally:
    ida_pro.qexit(0)
