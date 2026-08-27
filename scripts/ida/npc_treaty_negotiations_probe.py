"""匯出 sub_2552D 的非破壞性 IDA 證據。"""

import hashlib
import json
import os

import ida_bytes
import ida_funcs
import ida_hexrays
import ida_ida
import ida_idp
import ida_name
import ida_nalt
import ida_ua
import idautils
import idc


TARGET_EA = 0x2552D
RELATED_EAS = (
    0x252A7, 0x252D5, 0x25AD2, 0x25BC6, 0x25DF1, 0x2670A,
    0x4E3B5, 0x4EB06, 0x4F0DC, 0x4F694, 0x4FE25, 0x500CF, 0x5090C, 0x51078,
    0x5138E, 0x52049, 0x5232E, 0x524C3,
    0x524FB, 0x52602,
    0x53EDB,
    0xD3D34,
    0x54D80, 0x54E5B, 0x56726, 0x5679E, 0x5680D, 0x5685F, 0x5699C, 0x56CA2,
    0x582BF, 0x58387, 0x58425, 0x5EAE9, 0x5EE27, 0x5EED4,
    0x5EF09, 0x5EF17, 0x5EF4B, 0x5F2C3, 0x5F2F6, 0x5F379, 0x5F871,
)
GOVERNMENT_SCORE_TABLE_EA = 0x180CCC
INCIDENT_GOVERNMENT_THRESHOLD_TABLE_EA = 0x180CDC
INCIDENT_DEMOCRACY_MEMORY_THRESHOLD_EA = 0x180CE8
TREATY_COOLDOWN_GOVERNMENT_TABLE_EA = 0x18105C
NPC_POWER_MOD_WORD_EAS = (
    0x17FD26, 0x17FD35, 0x17FD62, 0x17FD80, 0x17FD8F, 0x17FD9E,
    0x17FDAD, 0x17FDBC, 0x17FDCB, 0x17FDDA, 0x17FDE9,
)
NPC_POWER_DRIVE_DEFENSE_BASE_EA = 0x17FE92
NPC_POWER_DRIVE_RECORD_SIZE = 46
NPC_POWER_COMPUTER_TECH_WORD_BASE_EA = 0x17F6A7
NPC_POWER_COMPUTER_TECH_RECORD_SIZE = 59
NPC_POWER_COMPUTER_REDUCTION_BASE_EA = 0x17F6C1
NPC_POWER_STRUCTURE_BASE_EA = 0x180026
NPC_POWER_ARMOR_BASE_EA = 0x180024
NPC_POWER_HULL_RECORD_SIZE = 36
NPC_POWER_ARMOR_PERCENT_BASE_EA = 0x17F642
NPC_POWER_ARMOR_RECORD_SIZE = 15
NPC_POWER_COMPUTER_ATTACK_BASE_EA = 0x17FE00
NPC_POWER_COMPUTER_RECORD_SIZE = 22
NPC_POWER_WEAPON_BASE_EA = 0x17F807
NPC_POWER_WEAPON_RECORD_SIZE = 28
NPC_POWER_FIGHTER_MINI_THRESHOLD_EA = 0x17FD32
RELATIVE_OPERANDS = ("+5ECh]", "+617h]", "+627h]", "+62Fh]", "+637h]", "+63Fh]",
                     "+64Fh]", "+65Fh]", "+68Fh]", "+69Fh]", "+6D7h]", "+717h]",
                     "+71Fh]", "+727h]", "+72Fh]", "+737h]")


def instruction(ea):
    insn = ida_ua.insn_t()
    size = ida_ua.decode_insn(insn, ea) or 1
    return {
        "ea": hex(ea),
        "bytes": (ida_bytes.get_bytes(ea, size) or b"").hex(),
        "disasm": idc.generate_disasm_line(ea, 0) or "",
        "code_refs": [hex(x) for x in idautils.CodeRefsFrom(ea, 0)],
        "data_refs": [
            {"ea": hex(x), "original_name": ida_name.get_name(x)}
            for x in idautils.DataRefsFrom(ea)
        ],
    }


def identity(fn):
    raw = ida_bytes.get_bytes(fn.start_ea, fn.end_ea - fn.start_ea) or b""
    return {
        "original_name": ida_name.get_name(fn.start_ea),
        "start_ea": hex(fn.start_ea),
        "end_ea": hex(fn.end_ea),
        "size": fn.end_ea - fn.start_ea,
        "bytes_sha256": hashlib.sha256(raw).hexdigest(),
    }


def instructions(fn):
    return [instruction(ea) for ea in idautils.FuncItems(fn.start_ea)]


def main():
    fn = ida_funcs.get_func(TARGET_EA)
    if fn is None:
        raise RuntimeError("sub_2552D missing")
    target = identity(fn)
    target["instructions"] = instructions(fn)
    target["callers"] = []
    for xref in idautils.XrefsTo(fn.start_ea, 0):
        caller = ida_funcs.get_func(xref.frm)
        if caller is None:
            continue
        items = list(idautils.FuncItems(caller.start_ea))
        pos = items.index(xref.frm) if xref.frm in items else 0
        target["callers"].append({
            "call_ea": hex(xref.frm),
            "caller": identity(caller),
            "window": [instruction(ea) for ea in items[max(0, pos - 20):pos + 13]],
        })

    callees = {}
    for ea in idautils.FuncItems(fn.start_ea):
        insn = ida_ua.insn_t()
        if not ida_ua.decode_insn(insn, ea) or not ida_idp.is_call_insn(insn):
            continue
        for ref in idautils.CodeRefsFrom(ea, 0):
            callee = ida_funcs.get_func(ref)
            if callee is None:
                continue
            item = callees.setdefault(hex(callee.start_ea), {
                "function": identity(callee), "call_sites": [],
            })
            item["call_sites"].append(hex(ea))
    target["callees"] = list(callees.values())
    try:
        target["decompiler_navigation_only"] = str(ida_hexrays.decompile(fn.start_ea))
    except Exception as exc:
        target["decompiler_error"] = str(exc)

    target["related_functions"] = []
    for ea in RELATED_EAS:
        related = ida_funcs.get_func(ea)
        if related is None:
            continue
        item = identity(related)
        item["instructions"] = instructions(related)
        try:
            item["decompiler_navigation_only"] = str(ida_hexrays.decompile(related.start_ea))
        except Exception as exc:
            item["decompiler_error"] = str(exc)
        target["related_functions"].append(item)

    target["relative_operand_sites"] = []
    for ea in idautils.Heads():
        line = idc.generate_disasm_line(ea, 0) or ""
        if not any(token in line for token in RELATIVE_OPERANDS):
            continue
        owner = ida_funcs.get_func(ea)
        target["relative_operand_sites"].append({
            "instruction": instruction(ea),
            "function": identity(owner) if owner is not None else None,
        })

    table = ida_bytes.get_bytes(GOVERNMENT_SCORE_TABLE_EA, 16) or b""
    target["government_score_table"] = {
        "start_ea": hex(GOVERNMENT_SCORE_TABLE_EA),
        "raw_hex": table.hex(),
        "bytes_sha256": hashlib.sha256(table).hexdigest(),
        "signed_words": [
            int.from_bytes(table[i:i + 2], "little", signed=True)
            for i in range(0, len(table), 2)
        ],
    }

    incident_table = ida_bytes.get_bytes(INCIDENT_GOVERNMENT_THRESHOLD_TABLE_EA, 16) or b""
    target["incident_memory_thresholds"] = {
        "government_table_ea": hex(INCIDENT_GOVERNMENT_THRESHOLD_TABLE_EA),
        "government_signed_words": [
            int.from_bytes(incident_table[i:i + 2], "little", signed=True)
            for i in range(0, len(incident_table), 2)
        ],
        "democracy_memory_ea": hex(INCIDENT_DEMOCRACY_MEMORY_THRESHOLD_EA),
        "democracy_memory_signed_word": int.from_bytes(
            ida_bytes.get_bytes(INCIDENT_DEMOCRACY_MEMORY_THRESHOLD_EA, 2) or b"\x00\x00",
            "little", signed=True),
    }
    cooldown_table = ida_bytes.get_bytes(TREATY_COOLDOWN_GOVERNMENT_TABLE_EA, 16) or b""
    target["treaty_cooldown_government_table"] = {
        "ea": hex(TREATY_COOLDOWN_GOVERNMENT_TABLE_EA),
        "signed_words": [
            int.from_bytes(cooldown_table[i:i + 2], "little", signed=True)
            for i in range(0, len(cooldown_table), 2)
        ],
    }

    target["npc_power_modifier_words"] = [
        {
            "ea": hex(ea),
            "raw_hex": (ida_bytes.get_bytes(ea, 2) or b"").hex(),
            "signed_word": int.from_bytes(
                ida_bytes.get_bytes(ea, 2) or b"\x00\x00", "little", signed=True),
        }
        for ea in NPC_POWER_MOD_WORD_EAS
    ]
    target["npc_power_drive_defense_bytes"] = [
        {
            "drive_index": index,
            "ea": hex(NPC_POWER_DRIVE_DEFENSE_BASE_EA + NPC_POWER_DRIVE_RECORD_SIZE * index),
            "signed_byte": int.from_bytes(
                ida_bytes.get_bytes(NPC_POWER_DRIVE_DEFENSE_BASE_EA + NPC_POWER_DRIVE_RECORD_SIZE * index, 1)
                or b"\x00", "little", signed=True),
        }
        for index in range(8)
    ]
    target["npc_power_computer_tech_offsets"] = [
        {
            "computer_index": index,
            "ea": hex(NPC_POWER_COMPUTER_TECH_WORD_BASE_EA + NPC_POWER_COMPUTER_TECH_RECORD_SIZE * index),
            "player_tech_status_offset": int.from_bytes(
                ida_bytes.get_bytes(
                    NPC_POWER_COMPUTER_TECH_WORD_BASE_EA + NPC_POWER_COMPUTER_TECH_RECORD_SIZE * index, 2)
                or b"\x00\x00", "little", signed=True),
        }
        for index in range(1, 6)
    ]
    target["npc_power_computer_reduction_words"] = [
        {
            "computer_index": index,
            "ea": hex(NPC_POWER_COMPUTER_REDUCTION_BASE_EA + NPC_POWER_COMPUTER_TECH_RECORD_SIZE * index),
            "signed_word": int.from_bytes(
                ida_bytes.get_bytes(
                    NPC_POWER_COMPUTER_REDUCTION_BASE_EA + NPC_POWER_COMPUTER_TECH_RECORD_SIZE * index, 2)
                or b"\x00\x00", "little", signed=True),
        }
        for index in range(6)
    ]
    target["npc_power_hull_capacity_words"] = [
        {
            "size": size,
            "structure": int.from_bytes(
                ida_bytes.get_bytes(NPC_POWER_STRUCTURE_BASE_EA + NPC_POWER_HULL_RECORD_SIZE * size, 2)
                or b"\x00\x00", "little", signed=True),
            "armor": int.from_bytes(
                ida_bytes.get_bytes(NPC_POWER_ARMOR_BASE_EA + NPC_POWER_HULL_RECORD_SIZE * size, 2)
                or b"\x00\x00", "little", signed=True),
        }
        for size in range(6)
    ]
    target["npc_power_armor_percent_words"] = [
        {
            "armor_index": index,
            "ea": hex(NPC_POWER_ARMOR_PERCENT_BASE_EA + NPC_POWER_ARMOR_RECORD_SIZE * index),
            "percent": int.from_bytes(
                ida_bytes.get_bytes(NPC_POWER_ARMOR_PERCENT_BASE_EA + NPC_POWER_ARMOR_RECORD_SIZE * index, 2)
                or b"\x00\x00", "little", signed=True),
        }
        for index in range(8)
    ]
    target["npc_power_computer_attack_words"] = [
        {
            "computer_index": index,
            "ea": hex(NPC_POWER_COMPUTER_ATTACK_BASE_EA + NPC_POWER_COMPUTER_RECORD_SIZE * index),
            "attack": int.from_bytes(
                ida_bytes.get_bytes(NPC_POWER_COMPUTER_ATTACK_BASE_EA + NPC_POWER_COMPUTER_RECORD_SIZE * index, 2)
                or b"\x00\x00", "little", signed=True),
        }
        for index in range(6)
    ]
    target["npc_power_fighter_beam_gate"] = {
        "mini_threshold_ea": hex(NPC_POWER_FIGHTER_MINI_THRESHOLD_EA),
        "mini_threshold": ida_bytes.get_byte(NPC_POWER_FIGHTER_MINI_THRESHOLD_EA),
        "weapon_static_flags": [
            {
                "weapon_id": weapon_id,
                "ea": hex(NPC_POWER_WEAPON_BASE_EA + NPC_POWER_WEAPON_RECORD_SIZE * weapon_id + 22),
                "flags": ida_bytes.get_word(
                    NPC_POWER_WEAPON_BASE_EA + NPC_POWER_WEAPON_RECORD_SIZE * weapon_id + 22),
            }
            for weapon_id in range(40)
        ],
    }

    result = {
        "tool": "IDA Pro 9.4 IDAPython",
        "root_filename": ida_nalt.get_root_filename(),
        "input_path": ida_nalt.get_input_file_path(),
        "input_sha256": ida_nalt.retrieve_input_file_sha256().hex(),
        "processor": ida_ida.inf_get_procname(),
        "compiler_id": int(ida_ida.inf_get_cc_id()),
        "address_space": "IDA linear EA",
        "target": target,
        "semantic_claim": {
            "level": "unknown",
            "note": "只保存原始定位與資料流；外部稽核另行分級。",
        },
    }
    output = os.environ.get("MOO2_IDA_PROBE_OUT", "/tmp/npc-treaty-negotiations.json")
    with open(output, "w", encoding="utf-8") as stream:
        json.dump(result, stream, ensure_ascii=False, indent=2, sort_keys=True)
        stream.write("\n")
    idc.qexit(0)


main()
