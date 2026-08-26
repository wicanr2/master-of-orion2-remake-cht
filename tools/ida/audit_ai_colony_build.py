"""非破壞性匯出 MOO2 AI 殖民地建造選擇與生產呼叫鏈。"""

import hashlib
import json
import os
import re
import traceback

import ida_auto
import ida_bytes
import ida_funcs
import ida_ida
import ida_kernwin
import ida_name
import ida_nalt
import ida_pro
import idautils
import idc


ROOTS = {
    "raw_Next_Turn_Calc": 0x136B3,
    "raw_Colony_Building_Score": 0xD0036,
    "raw_Assign_Colony_New_Building": 0xD0B08,
    "raw_AI_Build_Dispatch": 0xD0D2F,
    "raw_Assign_Empire_Building": 0xD10EE,
    "raw_Assign_Colony_Building": 0xD2754,
    "raw_Player_Colony_Autobuild": 0xD2783,
    "raw_AI_Colony_Primary_Player": 0xD2A08,
    "raw_Compute_AI_Data": 0xD3D34,
    "raw_Collect_AI_Colonies": 0xD5795,
    "raw_Assign_Buildings": 0xD589B,
    "raw_Clear_Build_Queues_If_New_Tech": 0xD58D4,
    "raw_Keep_Ship_Designs_Up_To_Date": 0xD5CEE,
    "raw_Do_Ship_Upgrade_Cheats": 0xD5E19,
    "raw_Assign_Required_Colonists": 0xD5FE1,
    "raw_Do_Blockaded_Colony": 0xD61E7,
    "raw_Do_Unblockaded_Colony": 0xD652C,
    "raw_AI_Colony_Tax_Dispatch": 0xD6E1D,
    "raw_Colony_AI": 0xD6ED4,
    "raw_All_Colony_AI": 0xD6F67,
    "raw_Colony_Product_Cost": 0xE0DD6,
    "raw_AI_Colony_Secondary_Value": 0xE0C1D,
    "raw_Colony_Can_Build_Product": 0xE11BC,
    "raw_Apply_Production": 0xE36DF,
    "raw_AI_Choose_Research": 0xDC288,
    "raw_AI_Empire_Output_Cache": 0xDF8F0,
    "raw_Colony_Population_Capacity_Core": 0xE0A18,
    "raw_Colony_Secondary_Base_Value": 0xE0A93,
    "raw_Colony_Food_Per_Farmer": 0xDE03E,
    "raw_Colony_Industry_Production": 0xDEE1B,
    "raw_Recompute_Colony_Output": 0xE1D59,
    "raw_Recompute_Player_Economy": 0xE2710,
    "raw_Do_Colony_Calculations": 0xE2B31,
    "raw_Apply_Player_Economy": 0xE4F49,
    "raw_Integer_Sqrt": 0x134C92,
}


TRACKED_RECORD_OFFSETS = {
    "colony_pollution_word": 0x08,
    "colony_raw_e9_word": 0xE9,
    "player_food_balance_word": 0xB0,
    "colony_food_output_byte": 0xDD,
    "player_cybernetic_trait": 0x8B0,
    "player_lithovore_trait": 0x8B1,
    "player_aquatic_trait": 0x8AB,
    "player_late_tech": 0x59D,
    "shared_gate_125": 0x125,
    "shared_gate_12d": 0x12D,
    "shared_gate_17e": 0x17E,
    "shared_gate_138": 0x138,
    "shared_gate_13d": 0x13D,
    "shared_gate_14c": 0x14C,
}

REVIEWED_OFFSET_SEMANTICS = {
    "colony_pollution_word": {
        "semantic": "base-dependent raw +0x08; colony base stores current pollution cleanup burden",
        "confidence": "confirmed_from_sub_DEE1B_write_and_industry_consumer",
    },
    "colony_raw_e9_word": {
        "semantic": "colony-context signed word storing current net industry after pollution and downstream production diversions",
        "confidence": "confirmed_from_sub_DEE1B_write_and_sub_E2710_empire_sum",
    },
    "player_food_balance_word": {
        "semantic": "base-dependent raw +0xB0; player base stores empire food production minus consumption",
        "confidence": "confirmed_only_for_reviewed_player_base_context",
    },
    "colony_food_output_byte": {
        "semantic": "base-dependent raw +0xDD; colony-context meaning pending write-origin review",
        "confidence": "unknown_pending_write_origin_review",
    },
    "player_cybernetic_trait": {
        "semantic": "base-dependent raw +0x8B0; candidate player Cybernetic trait byte pending write-origin review",
        "confidence": "hypothesis_pending_write_origin_review",
    },
    "player_lithovore_trait": {
        "semantic": "base-dependent raw +0x8B1; candidate player Lithovore trait byte pending write-origin review",
        "confidence": "strong_inference_from_trait_table_index_and_consumers",
    },
    "player_aquatic_trait": {
        "semantic": "base-dependent raw +0x8AB; player Aquatic trait byte from runtime trait index 12",
        "confidence": "confirmed_from_trait_table_index_and_independent_food_consumers",
    },
    "player_late_tech": {
        "semantic": "player late-tech flag when research field >= 75",
        "confidence": "confirmed_direct_write_and_consumers",
    },
    "shared_gate_125": {
        "semantic": "base-dependent raw +0x125; player base maps technology 14 in D0036/D58D4",
        "confidence": "confirmed_only_for_reviewed_base_context",
    },
    "shared_gate_12d": {
        "semantic": "base-dependent raw +0x12D; player base maps technology 22 in D0036/D58D4",
        "confidence": "confirmed_only_for_reviewed_base_context",
    },
    "shared_gate_17e": {
        "semantic": "base-dependent raw +0x17E; player base maps technology 103 in D0036/D58D4",
        "confidence": "confirmed_only_for_reviewed_base_context",
    },
    "shared_gate_138": {
        "semantic": "base-dependent raw +0x138; colony base maps building 2 in D0036/D58D4",
        "confidence": "confirmed_only_for_reviewed_base_context",
    },
    "shared_gate_13d": {
        "semantic": "base-dependent raw +0x13D; colony base maps building 7 in D0036/D58D4",
        "confidence": "confirmed_only_for_reviewed_base_context",
    },
    "shared_gate_14c": {
        "semantic": "base-dependent raw +0x14C; colony base maps building 22 in D0036/D58D4",
        "confidence": "confirmed_only_for_reviewed_base_context",
    },
}


def operand_mentions_offset(ea, offset):
    """只比對帶 `+offset` 的原始運算元，避免把立即數或絕對位址混入。"""
    operands = [idc.print_operand(ea, i) for i in range(2)]
    normalized = " ".join(operands).lower().replace("0x", "")
    needle = f"+{offset:x}h"
    spaced = f"+ {offset:x}h"
    if needle in normalized or spaced in normalized:
        return True
    # IDA 對首碼為 A..F 的 displacement 會補前導 0（例如 +0B0h）。只正規化
    # `+...h` 內的十六進位數，不碰立即數或絕對位址。
    for match in re.finditer(r"\+\s*([0-9a-f]+)h", normalized):
        if int(match.group(1), 16) == offset:
            return True
    # 小位移有時以無尾碼十進位顯示（例如 `[ecx+8]`）。只接受緊鄰 `]`／`,` 的
    # displacement 位置，避免把比例或立即數 8 混入。
    for match in re.finditer(r"\+\s*([0-9]+)(?=\s*[\],])", normalized):
        if int(match.group(1), 10) == offset:
            return True
    return False


def instruction_context(fn_ea, ea, radius=6):
    items = list(idautils.FuncItems(fn_ea))
    try:
        index = items.index(ea)
    except ValueError:
        return []
    lo = max(0, index - radius)
    hi = min(len(items), index + radius + 1)
    return [instruction(item) for item in items[lo:hi]]


def direct_record_offset_refs(offset):
    refs = []
    for fn_ea in idautils.Functions():
        for ea in idautils.FuncItems(fn_ea):
            if operand_mentions_offset(ea, offset):
                refs.append({
                    "function_start": f"0x{fn_ea:X}",
                    "raw_name": ida_name.get_name(fn_ea) or "<unnamed>",
                    "instruction": instruction(ea),
                    "context": instruction_context(fn_ea, ea),
                })
    return refs


def global_data_refs(ea):
    """保留全域本體的直接 xref 與函式內前後文，供人工續追指標式間接讀寫。"""
    refs = []
    for ref in idautils.DataRefsTo(ea):
        owner = ida_funcs.get_func(ref)
        refs.append({
            "global_ea": f"0x{ea:X}",
            "site": instruction(ref),
            "function_start": f"0x{owner.start_ea:X}" if owner else None,
            "raw_name": ida_name.get_name(owner.start_ea) if owner else None,
            "context": instruction_context(owner.start_ea, ref) if owner else [],
        })
    return refs


def digest(path):
    h = hashlib.sha256()
    with open(path, "rb") as fh:
        for chunk in iter(lambda: fh.read(1024 * 1024), b""):
            h.update(chunk)
    return h.hexdigest()


def instruction(ea):
    size = max(1, idc.get_item_size(ea))
    return {
        "ea": f"0x{ea:X}",
        "bytes": (ida_bytes.get_bytes(ea, size) or b"").hex(),
        "mnem": idc.print_insn_mnem(ea),
        "op0": idc.print_operand(ea, 0),
        "op1": idc.print_operand(ea, 1),
        "text": idc.generate_disasm_line(ea, 0) or "<unavailable>",
        "code_refs": [f"0x{x:X}" for x in idautils.CodeRefsFrom(ea, 0)],
        "data_refs": [f"0x{x:X}" for x in idautils.DataRefsFrom(ea)],
    }


def function_record(ea):
    fn = ida_funcs.get_func(ea)
    if fn is None:
        return {"requested": f"0x{ea:X}", "error": "function missing"}
    calls = []
    for item in idautils.FuncItems(fn.start_ea):
        if idc.print_insn_mnem(item).lower().startswith("call"):
            calls.append({
                "site": instruction(item),
                "targets": [
                    {"ea": f"0x{x:X}", "raw_name": ida_name.get_name(x) or "<unnamed>"}
                    for x in idautils.CodeRefsFrom(item, 0)
                ],
            })
    callers = []
    for ref in idautils.CodeRefsTo(fn.start_ea, 0):
        owner = ida_funcs.get_func(ref)
        callers.append({
            "site": instruction(ref),
            "function_start": f"0x{owner.start_ea:X}" if owner else None,
            "raw_name": ida_name.get_name(owner.start_ea) if owner else None,
        })
    return {
        "requested": f"0x{ea:X}",
        "start": f"0x{fn.start_ea:X}",
        "end_exclusive": f"0x{fn.end_ea:X}",
        "raw_name": ida_name.get_name(fn.start_ea) or "<unnamed>",
        "callers": callers,
        "direct_calls": calls,
        "instructions": [instruction(x) for x in idautils.FuncItems(fn.start_ea)],
    }


def main():
    ida_auto.auto_wait()
    source = os.environ["MOO2_IDA_INPUT"]
    database = os.environ["MOO2_IDA_DATABASE"]
    score_switch_table = 0xCFF62
    score_switch = []
    for case_index in range(47):
        entry_ea = score_switch_table + case_index * 4
        target = ida_bytes.get_dword(entry_ea)
        score_switch.append({
            "raw_building_id": case_index + 1,
            "entry_ea": f"0x{entry_ea:X}",
            "entry_bytes": (ida_bytes.get_bytes(entry_ea, 4) or b"").hex(),
            "target_ea": f"0x{target:X}",
            "target_name": ida_name.get_name(target) or "<unnamed>",
        })
    terraform_switch_table = 0xD001E
    terraform_switch = []
    for case_index in range(6):
        entry_ea = terraform_switch_table + case_index * 4
        target = ida_bytes.get_dword(entry_ea)
        terraform_switch.append({
            "raw_climate": case_index + 2,
            "entry_ea": f"0x{entry_ea:X}",
            "entry_bytes": (ida_bytes.get_bytes(entry_ea, 4) or b"").hex(),
            "target_ea": f"0x{target:X}",
            "target_name": ida_name.get_name(target) or "<unnamed>",
        })
    report = {
        "schema": "moo2.ida.re-evidence.v1",
        "contract": "raw-location + navigation-label + reviewed confidence + source",
        "mutation": "none",
        "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version()},
        "input": {
            "file": os.path.basename(source),
            "source_sha256": digest(source),
            "database_sha256": digest(database),
            "processor": ida_ida.inf_get_procname(),
        },
        "address_basis": "IDA linear; DOS/4GW LE image",
        "semantic_status": "unknown_pending_review",
        "score_switch": {
            "jump_ea": "0xD01BF",
            "table_ea": f"0x{score_switch_table:X}",
            "case_count": 47,
            "entries": score_switch,
        },
        "terraform_score_switch": {
            "jump_ea": "0xD0A67",
            "table_ea": f"0x{terraform_switch_table:X}",
            "case_count": 6,
            "entries": terraform_switch,
        },
        "direct_record_offset_refs": {
            name: {
                "offset": f"0x{offset:X}",
                "reviewed_semantic": REVIEWED_OFFSET_SEMANTICS[name]["semantic"],
                "confidence": REVIEWED_OFFSET_SEMANTICS[name]["confidence"],
                "evidence_source": "docs/re/ai-colony-build-selection-audit-20260826.md",
                "refs": direct_record_offset_refs(offset),
            }
            for name, offset in TRACKED_RECORD_OFFSETS.items()
        },
        "global_data_refs": {
            "raw_ai_colony_cache_pointer": global_data_refs(0x1AA1EC),
        },
        "roots": {name: function_record(ea) for name, ea in ROOTS.items()},
    }
    with open(os.environ["MOO2_IDA_OUTPUT"], "w", encoding="utf-8") as fh:
        json.dump(report, fh, ensure_ascii=False, indent=2)
        fh.write("\n")


try:
    main()
except Exception:
    error = traceback.format_exc()
    out = os.environ.get("MOO2_IDA_OUTPUT", "/tmp/ai-colony-build.json")
    with open(out + ".error", "w", encoding="utf-8") as fh:
        fh.write(error)
    ida_kernwin.msg(error + "\n")
    ida_pro.qexit(1)
else:
    ida_pro.qexit(0)
