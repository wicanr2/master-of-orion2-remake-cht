"""唯讀索引 MOO2 player+0x89F 的 31-byte 種族特性陣列及全部直接 consumer。"""

import hashlib
import json
import os
import re
import csv
import bisect
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


TRAIT_BASE = 0x89F
TRAIT_NAMES = [
    "GOVERNMENT", "POPULATION", "FARMING", "INDUSTRY", "SCIENCE", "MONEY",
    "SHIP_DEFENSE", "SHIP_ATTACK", "GROUND_COMBAT", "SPYING", "LOW_G", "HIGH_G",
    "AQUATIC", "SUBTERRANEAN", "LARGE_HOMEWORLD", "RICH_OR_POOR_HOMEWORLD",
    "ARTIFACTS_HOMEWORLD", "CYBERNETIC", "LITHOVORE", "REPULSIVE", "CHARISMATIC",
    "UNCREATIVE", "CREATIVE", "TOLERANT", "FANTASTIC_TRADERS", "TELEPATHIC",
    "LUCKY", "OMNISCIENCE", "STEALTHY_SHIPS", "TRANS_DIMENSIONAL", "WARLORD",
]
TRAIT_END = TRAIT_BASE + len(TRAIT_NAMES) - 1
DISPLACEMENT = re.compile(r"(?P<sign>[+-])(?P<hex>[0-9A-Fa-f]+)h")

ROOTS = {
    "convert_custom_race_flags": 0x5BC24,
    "initialize_player_tech": 0x5E55F,
    "calculate_tech_value": 0xFC845,
    "choose_tech_application": 0xFD335,
    "assimilation": 0xE3456,
    "spy_attack_table": 0x100A3E,
    "spy_defense_table": 0x100A83,
    "lucky_event_counter": 0x245C4,
    "lucky_event_target": 0x22D57,
    "ai_data": 0xD3D34,
    "colony_population_growth": 0xE1839,
    "colony_food": 0xDE0C6,
    "colony_industry_base": 0xDED47,
    "colony_industry_total": 0xDEE1B,
    "colony_industry_maintenance": 0xDF546,
    "colony_research_base": 0xDFE77,
    "ground_combat_bonus": 0xEC15C,
    "council_vote_check": 0x16021,
    "diplomacy_response": 0x5272D,
    "ai_human_target_score": 0x544A1,
    "fleet_movement": 0xFF799,
    "ship_crew_level": 0x56726,
    "telepathic_ai_self_destruct": 0x28168,
    "telepathic_boarding_action_type": 0x2C129,
    "telepathic_capture_ship": 0x38312,
    "telepathic_diplomacy_test": 0x53146,
    "telepathic_mind_control_with_help": 0xC613B,
    "telepathic_mind_control": 0xC622A,
    "telepathic_colony_combat_fields": 0xCADC8,
    "telepathic_enemy_colony_worth": 0xD8D11,
    "telepathic_best_colony_target": 0xE78A7,
    "telepathic_attacker_vs_colony": 0xE87D2,
    "telepathic_change_population_ownership": 0xECBF7,
    "telepathic_calculate_tech_value": 0xFC845,
    "telepathic_compute_spy_bonuses": 0x100A83,
    "omniscience_find_ship_stacks": 0x5D41E,
    "omniscience_print_scanned_star_name": 0x761A5,
    "omniscience_print_scanned_ship_data": 0x7670E,
    "omniscience_player_is_omniscient": 0x79E06,
    "omniscience_star_owner": 0x7A1A8,
    "omniscience_check_hot_keys": 0x82809,
    "omniscience_explored_new_star": 0xFD95A,
    "stealth_add_specials_to_design": 0x5FE14,
    "stealth_compute_ai_data": 0xD3D34,
    "stealth_initialize_npc_profiles": 0x589D6,
    "ship_defense_defensive_combat_bonus": 0x35D0D,
    "ship_defense_missile_dcv": 0x3DFE0,
    "ship_defense_qload_ships": 0x416CF,
    "ship_defense_get_ship_combat_bonuses": 0x54E5B,
    "ship_defense_base_generic_dcv": 0x5EAE9,
    "ship_defense_determine_beam_modifications": 0x60B59,
    "ship_attack_auto_ship_turn": 0x29837,
    "ship_attack_offensive_combat_bonus": 0x366DD,
    "ship_attack_fighter_ocv": 0x3DF8D,
    "ship_attack_base_generic_ocv": 0x5EB39,
    "ship_attack_satellite_strength": 0x5EB72,
    "ship_attack_strength_vs_shield": 0x5EF4B,
    "transdimensional_npc_war_declarations": 0x25DF1,
    "transdimensional_missile_speed": 0x3CD21,
    "transdimensional_current_speed": 0x4528F,
    "transdimensional_display_combat_ship": 0x4CB1E,
    "transdimensional_get_ship_combat_bonuses": 0x54E5B,
    "transdimensional_calc_player_ftl": 0x57597,
    "transdimensional_get_ftl": 0x575D6,
    "transdimensional_base_generic_dcv": 0x5EAE9,
    "transdimensional_ai_move_all": 0xDBB29,
    "transdimensional_opportunity_attack": 0xDBC5C,
    "transdimensional_player_threat": 0xDBCC8,
    "transdimensional_determine_retreat": 0xE6CAA,
    "transdimensional_ship_try_move": 0xFF799,
    "repchar_vote_check": 0x16021,
    "repulsive_diplomacy_screen": 0x16C4E,
    "repulsive_net_diplomacy_choices": 0x1DEF8,
    "repulsive_npc_diplomacy": 0x252A7,
    "repulsive_npc_treaty_negotiation": 0x2552D,
    "charismatic_tech_exchange_reaction": 0x26BBD,
    "charismatic_tech_exchange_expectation": 0x26FBA,
    "charismatic_change_relations": 0x4E3B5,
    "repulsive_determine_diplomacy_messages": 0x4EB06,
    "repulsive_determine_bad_message": 0x4F0DC,
    "repchar_diplomacy_test": 0x53146,
    "repchar_sneak_attack_evaluations": 0x544A1,
    "repchar_chance_hire_hero": 0x9781D,
    "repchar_select_leader_for_hire": 0x97B2D,
    "repulsive_allocate_advanced_officers": 0x98489,
    "repulsive_ai_leaders": 0xD7439,
    "repchar_apply_assimilation": 0xE3456,
    "repchar_next_ai_talker": 0xFA1A3,
    "government_init_players": 0x12983,
    "government_init_homeworld": 0x13A3D,
    "government_race_customization": 0x150FB,
    "government_differential_tech_list": 0x27094,
    "government_evolutionary_setup": 0x5B0AA,
    "government_tech_legal": 0x5E481,
    "government_init_player_tech": 0x5E55F,
    "government_cost_reduction": 0x6E1A0,
    "government_advanced_type": 0x93D0A,
    "government_draw_info_morale": 0xB53CC,
    "government_leader_technology_string": 0xC760E,
    "government_occupation_policy_popup": 0xCD969,
    "government_colony_building_score": 0xD0036,
    "government_colony_worth": 0xD2CAE,
    "government_clear_build_queues": 0xD58D4,
    "government_assign_star_officer": 0xD7171,
    "government_apply_evolutionary_upgrades": 0xDCBB0,
    "government_colony_morale": 0xDDB25,
    "government_colony_job_production": 0xDE280,
    "government_colony_bc_production": 0xE03F1,
    "government_colony_can_build": 0xE11BC,
    "government_player_maintenance": 0xE2000,
    "government_apply_assimilation": 0xE3456,
    "government_player_gets_tech": 0xE4204,
    "government_change_pop_ownership": 0xECBF7,
    "government_building_worth": 0xED9EC,
    "government_calc_tech_value": 0xFC845,
    "government_compute_spy_bonuses": 0x100A83,
    "government_trade_agreement_goal": 0x101BA4,
    "government_research_agreement_goal": 0x101CC5,
    "cybernetic_strategic_combat": 0x40148,
    "cybernetic_repair_combat_ship": 0x35821,
    "cybernetic_fleet_engineer_bonus": 0x35E0C,
    "cybernetic_end_of_combat": 0x4B184,
    "cybernetic_repair_all_combat_ships": 0x4CFE7,
    "cybernetic_uncolonized_planet_worth": 0xD27A7,
    "cybernetic_colony_worth": 0xD2CAE,
    "cybernetic_food_maintenance": 0xDEB4B,
    "cybernetic_industry_maintenance": 0xDF546,
    "cybernetic_apply_population_growth": 0xE2DCA,
    "lithovore_init_homeworld": 0x13A3D,
    "lithovore_tech_legal": 0x5E481,
    "lithovore_ensure_food_planet": 0x7B631,
    "lithovore_uncolonized_planet_worth": 0xD27A7,
    "lithovore_colony_worth": 0xD2CAE,
    "lithovore_compute_ai_data": 0xD3D34,
    "lithovore_food_maintenance": 0xDEB4B,
    "lithovore_apply_population_growth": 0xE2DCA,
    "homeworld_twiddle_advanced": 0x638A9,
    "homeworld_modify": 0x7C4AF,
    "homeworld_twiddle_initial": 0xE5832,
    "gravity_enforce": 0x7B45C,
    "gravity_resolve_bomb_hit": 0xDCEBD,
    "warlord_calc_ship_level": 0x147E7,
    "warlord_owned_officer_level": 0x94951,
    "warlord_ground_combat_bonus": 0xEC15C,
    "warlord_ground_military": 0xE3616,
    "creative_research_breakthrough": 0xE4410,
    "uncreative_limit_field_applications": 0xE408F,
    "creative_player_gets_tech": 0xE4204,
    "spying_compute_needs": 0xCF40D,
    "spying_compute_bonuses": 0x100A83,
    "spying_allocate_ai_spies": 0x100D19,
    "spying_resolve_player_spies": 0x1014A4,
    "spying_resolve_spies": 0x10192B,
}


def load_external_symbols(path):
    symbols = {}
    with open(path, "r", encoding="utf-8", newline="") as source:
        for row in csv.DictReader(source, delimiter="\t"):
            try:
                ea = int(row["ea"], 16)
            except (KeyError, TypeError, ValueError):
                continue
            symbols[ea] = {
                "name": row.get("name") or "<unnamed>",
                "type": row.get("type") or "",
                "segment": row.get("seg") or "",
                "module": row.get("module") or "",
            }
    return symbols


def digest(path):
    h = hashlib.sha256()
    with open(path, "rb") as source:
        for block in iter(lambda: source.read(1024 * 1024), b""):
            h.update(block)
    return h.hexdigest()


def insn(ea):
    decoded = ida_ua.insn_t()
    ida_ua.decode_insn(decoded, ea)
    return {
        "ea": f"0x{ea:X}",
        "bytes": (ida_bytes.get_bytes(ea, decoded.size) or b"").hex(),
        "text": idc.generate_disasm_line(ea, 0) or "",
        "mnemonic": ida_ua.print_insn_mnem(ea),
        "operands": [idc.print_operand(ea, i) for i in range(8) if idc.print_operand(ea, i)],
    }


def annotation(offset):
    index = offset - TRAIT_BASE
    return {
        "raw_offset": f"0x{offset:X}",
        "trait_index": index,
        "navigation_semantic": TRAIT_NAMES[index],
        "evidence_level": "candidate_pending_consumer_review",
        "evidence_source": "31-byte enum mapping plus direct operand candidate",
        "warning": "列舉名稱只供導覽；必須審查原始 producer／consumer 才能升格",
    }


def function_record(ea, external_symbols):
    fn = ida_funcs.get_func(ea)
    if fn is None:
        return {"requested": f"0x{ea:X}", "error": "no function"}
    items = list(idautils.FuncItems(fn.start_ea))
    raw = b"".join(ida_bytes.get_bytes(x, ida_bytes.get_item_size(x)) or b"" for x in items)
    pseudo = None
    if ida_hexrays.init_hexrays_plugin():
        try:
            pseudo = str(ida_hexrays.decompile(fn.start_ea))
        except Exception as error:
            pseudo = f"<decompile failed: {error}>"
    return {
        "requested": f"0x{ea:X}",
        "start_ea": f"0x{fn.start_ea:X}",
        "end_ea": f"0x{fn.end_ea:X}",
        "original_name": ida_name.get_name(fn.start_ea) or "<unnamed>",
        "external_symbol_exact_address": external_symbols.get(fn.start_ea),
        "code_callers": [
            {
                "call_ea": f"0x{ref:X}",
                "owner_start": f"0x{ida_funcs.get_func(ref).start_ea:X}" if ida_funcs.get_func(ref) else None,
                "owner_original_name": ida_name.get_name(ida_funcs.get_func(ref).start_ea)
                if ida_funcs.get_func(ref) else None,
                "owner_external_symbol_exact_address": external_symbols.get(ida_funcs.get_func(ref).start_ea)
                if ida_funcs.get_func(ref) else None,
                "instruction": insn(ref),
            }
            for ref in idautils.CodeRefsTo(fn.start_ea, False)
        ],
        "bytes_sha256": hashlib.sha256(raw).hexdigest(),
        "instruction_count": len(items),
        "navigation_label_not_evidence": True,
        "pseudocode_navigation_only": pseudo,
        "instructions": [insn(x) for x in items],
    }


def direct_sites(external_symbols):
    by_offset = {}
    external_addresses = sorted(external_symbols)
    for fn_ea in idautils.Functions():
        fn = ida_funcs.get_func(fn_ea)
        items = list(idautils.FuncItems(fn_ea))
        player_base_positions = [
            position for position, x in enumerate(items) if
            "dword_197F98" in (idc.generate_disasm_line(x, 0) or "")
            or "dword_197FB8" in (idc.generate_disasm_line(x, 0) or "")
        ]
        if not player_base_positions:
            continue
        for position, ea in enumerate(items):
            operands = " ".join(idc.print_operand(ea, i) for i in range(4))
            offsets = set()
            for match in DISPLACEMENT.finditer(operands):
                value = int(match.group("hex"), 16)
                if match.group("sign") == "-":
                    value = -value
                if TRAIT_BASE <= value <= TRAIT_END:
                    offsets.add(value)
            for offset in offsets:
                prior_distances = [position - p for p in player_base_positions if p <= position]
                following_distances = [p - position for p in player_base_positions if p > position]
                nearest_prior = min(prior_distances) if prior_distances else None
                nearest_any = min(prior_distances + following_distances) if prior_distances or following_distances else None
                symbol_position = bisect.bisect_right(external_addresses, ea) - 1
                nearest_symbol_ea = external_addresses[symbol_position] if symbol_position >= 0 else None
                next_symbol_ea = external_addresses[symbol_position + 1] if 0 <= symbol_position + 1 < len(external_addresses) else ida_ida.inf_get_max_ea()
                external_extent_has_player_base = False
                if nearest_symbol_ea is not None:
                    external_extent_has_player_base = any(
                        "dword_197F98" in (idc.generate_disasm_line(head, 0) or "")
                        or "dword_197FB8" in (idc.generate_disasm_line(head, 0) or "")
                        for head in idautils.Heads(nearest_symbol_ea, next_symbol_ea)
                    )
                by_offset.setdefault(f"0x{offset:X}", []).append({
                    "semantic_annotation": annotation(offset),
                    "owner_start": f"0x{fn.start_ea:X}",
                    "owner_end": f"0x{fn.end_ea:X}",
                    "owner_original_name": ida_name.get_name(fn.start_ea) or "<unnamed>",
                    "owner_external_symbol_exact_address": external_symbols.get(fn.start_ea),
                    "nearest_preceding_external_symbol": {
                        "ea": f"0x{nearest_symbol_ea:X}" if nearest_symbol_ea is not None else None,
                        "record": external_symbols.get(nearest_symbol_ea),
                        "agrees_with_ida_owner": nearest_symbol_ea == fn.start_ea,
                        "next_symbol_ea": f"0x{next_symbol_ea:X}",
                        "extent_has_player_base_reference": external_extent_has_player_base,
                    },
                    "player_base_context": {
                        "nearest_prior_instruction_distance": nearest_prior,
                        "nearest_any_instruction_distance": nearest_any,
                        "locally_supported": nearest_prior is not None and nearest_prior <= 16,
                        "warning": "false positives remain possible; manually review register provenance",
                    },
                    "instruction": insn(ea),
                    "window": [insn(x) for x in items[max(0, position - 6):position + 7]],
                })
    return dict(sorted(by_offset.items(), key=lambda item: int(item[0], 16)))


def main():
    ida_auto.auto_wait()
    source = os.environ["MOO2_RE_SOURCE"]
    database = os.environ["MOO2_RE_DATABASE"]
    symbols_path = os.environ["MOO2_RE_SYMBOLS"]
    external_symbols = load_external_symbols(symbols_path)
    report = {
        "schema": "moo2.ida.custom-race-trait-consumers.v1",
        "evidence_scope": "static_only",
        "mutation": "none",
        "semantic_status": "candidate_operand_sites_pending_register_provenance_review",
        "tool": {
            "name": "IDA Pro",
            "version": ida_kernwin.get_kernel_version(),
            "script": "tools/ida/audit_custom_race_trait_consumers.py",
        },
        "input": {
            "file": os.path.basename(source),
            "source_sha256": digest(source),
            "database_sha256": digest(database),
            "external_symbols_file": os.path.basename(symbols_path),
            "external_symbols_sha256": digest(symbols_path),
            "processor": ida_ida.inf_get_procname(),
        },
        "address_basis": "IDA linear; DOS/4GW LE object #1",
        "trait_array": {
            "base": f"0x{TRAIT_BASE:X}",
            "end_inclusive": f"0x{TRAIT_END:X}",
            "count": len(TRAIT_NAMES),
            "poor_homeworld_encoding": "signed -1 in index 15; no runtime byte index 31",
        },
        "direct_operand_sites": direct_sites(external_symbols),
        "roots": {name: function_record(ea, external_symbols) for name, ea in ROOTS.items()},
    }
    with open(os.environ["MOO2_RE_OUT"], "w", encoding="utf-8") as target:
        json.dump(report, target, ensure_ascii=False, indent=2)
        target.write("\n")


try:
    main()
except Exception:
    output = os.environ.get("MOO2_RE_OUT", "/tmp/custom-race-trait-consumers.json")
    with open(output + ".error", "w", encoding="utf-8") as target:
        target.write(traceback.format_exc())
finally:
    ida_pro.qexit(0)
