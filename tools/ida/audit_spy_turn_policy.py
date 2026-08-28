"""非破壞性匯出 MOO2 1.31 間諜回合、AI 配置與 RACES 任務控制鏈。"""

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
    "raw_compute_needs": 0xCF40D,
    "raw_compute_empire_building_needs": 0xCFCB6,
    "raw_apply_production": 0xE36DF,
    "raw_compute_player_maintenance": 0xE2000,
    "raw_kill_spy": 0xEDCBF,
    "raw_compute_spy_bonuses": 0x100A83,
    "raw_allocate_ai_spies": 0x100D19,
    "raw_steal_app": 0x10119C,
    "raw_destroy_random_building": 0x10130A,
    "raw_n_spies_bonus": 0x101483,
    "raw_resolve_player_spies": 0x1014A4,
    "raw_leader_assassinates": 0x101768,
    "raw_resolve_assassins": 0x1018A3,
    "raw_resolve_spies": 0x10192B,
    "raw_bring_home_spies_vs": 0x1019F0,
    "raw_get_their_spy_number": 0x1026CF,
    "raw_get_their_spy_mission": 0x1026F1,
    "raw_get_my_spy_number": 0x102711,
    "raw_get_my_spy_mission": 0x102739,
    "raw_get_my_agent_number": 0x10275F,
    "raw_get_my_agent_mission": 0x102776,
    "raw_set_their_spy_number": 0x10278D,
    "raw_set_their_spy_mission": 0x1027B5,
    "raw_set_my_spy_number": 0x1027D9,
    "raw_set_my_spy_mission": 0x10280D,
    "raw_set_my_agent_number": 0x10283D,
    "raw_set_my_agent_mission": 0x102872,
    "raw_take_their_spy": 0x1028A2,
    "raw_take_my_spy": 0x1028D5,
    "raw_take_my_agent": 0x102912,
    "raw_add_their_spy": 0x10294B,
    "raw_add_my_spy": 0x102982,
    "raw_add_my_agent": 0x102994,
    "raw_init_race_display_data": 0x10BA3D,
    "raw_update_spy_stuff": 0x10C88D,
    "raw_get_spy_group_for_field": 0x10CBDF,
    "raw_redraw_spy_mouse": 0x10CC4D,
    "raw_adjust_spy_mission_data": 0x10CC23,
    "raw_move_spy_group_from_mouse": 0x10CCC5,
    "raw_move_spy_group_to_mouse": 0x10CD65,
}


RAW_RANGES = {
    "apply_production_spy_branch": (0xE3B20, 0xE3BE0),
    "ai_mission_assignment": (0x101130, 0x10119C),
    "spy_add_shared_tail": (0x102940, 0x1029D1),
    "race_spy_mouse_helpers": (0x10CBC0, 0x10CE20),
}


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
            "window": [instruction(item) for item in items[max(0, index - 14):min(len(items), index + 15)]],
        })
    return rows


def raw_range(start, end):
    rows = []
    ea = ida_bytes.next_head(start - 1, end)
    while ea != idc.BADADDR and ea < end:
        rows.append(instruction(ea))
        ea = ida_bytes.next_head(ea, end)
    return {
        "start_ea": f"0x{start:X}",
        "end_ea_exclusive": f"0x{end:X}",
        "instructions": rows,
    }


def main():
    ida_auto.auto_wait()
    source = os.environ["MOO2_RE_SOURCE"]
    database = os.environ["MOO2_RE_DATABASE"]
    symbols, symbols_hash = load_symbols(os.environ["MOO2_RE_SYMBOLS"])
    report = {
        "schema": "moo2.ida.spy-turn-policy.v1",
        "evidence_scope": "static_only",
        "mutation": "none",
        "semantic_status": "reviewed_against_raw_instructions",
        "tool": {
            "name": "IDA Pro",
            "version": ida_kernwin.get_kernel_version(),
            "script": "tools/ida/audit_spy_turn_policy.py",
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
        "raw_ranges": {
            name: raw_range(start, end)
            for name, (start, end) in RAW_RANGES.items()
        },
    }
    with open(os.environ["MOO2_RE_OUT"], "w", encoding="utf-8") as target:
        json.dump(report, target, ensure_ascii=False, indent=2)
        target.write("\n")


try:
    main()
except Exception:
    output = os.environ.get("MOO2_RE_OUT", "/tmp/spy-turn-policy.json")
    with open(output + ".error", "w", encoding="utf-8") as target:
        target.write(traceback.format_exc())
finally:
    ida_pro.qexit(0)
