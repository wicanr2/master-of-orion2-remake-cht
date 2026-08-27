"""匯出 Fighter_Garrison_Strength_ 與其直接消費端的非破壞性證據。"""

import hashlib
import json
import os

import ida_bytes
import ida_funcs
import ida_hexrays
import ida_ida
import ida_name
import ida_nalt
import ida_ua
import idautils
import idc


TARGET_EA = 0x5F64C


def instruction_record(ea):
    insn = ida_ua.insn_t()
    size = ida_ua.decode_insn(insn, ea) or 1
    return {
        "ea": hex(ea),
        "bytes": (ida_bytes.get_bytes(ea, size) or b"").hex(),
        "disasm": idc.generate_disasm_line(ea, 0) or "",
    }


def function_instructions(fn, limit=512):
    records = []
    ea = fn.start_ea
    while ea < fn.end_ea and len(records) < limit:
        records.append(instruction_record(ea))
        ea = idc.next_head(ea, fn.end_ea)
    return records


def function_summary(fn, include_instructions=True):
    raw = ida_bytes.get_bytes(fn.start_ea, fn.end_ea - fn.start_ea) or b""
    record = {
        "original_name": ida_name.get_name(fn.start_ea),
        "start_ea": hex(fn.start_ea),
        "end_ea": hex(fn.end_ea),
        "size": fn.end_ea - fn.start_ea,
        "bytes_sha256": hashlib.sha256(raw).hexdigest(),
    }
    if include_instructions:
        record["instructions"] = function_instructions(fn)
    try:
        record["decompiler_navigation_only"] = str(ida_hexrays.decompile(fn.start_ea))
    except Exception as exc:
        record["decompiler_error"] = str(exc)
    return record


def main():
    output = os.environ.get("MOO2_IDA_PROBE_OUT", "/tmp/fighter-garrison-probe.json")
    fn = ida_funcs.get_func(TARGET_EA)
    if fn is None:
        raise RuntimeError(f"no function at {TARGET_EA:#x}")

    callers = []
    for xref in idautils.XrefsTo(fn.start_ea, 0):
        caller = ida_funcs.get_func(xref.frm)
        if caller is None:
            continue
        item = function_summary(caller)
        item["call_ea"] = hex(xref.frm)
        item["xref_type"] = int(xref.type)
        callers.append(item)

    callees = []
    seen = set()
    for ea in idautils.FuncItems(fn.start_ea):
        for xref in idautils.XrefsFrom(ea, 0):
            callee = ida_funcs.get_func(xref.to)
            if callee is None or callee.start_ea == fn.start_ea or callee.start_ea in seen:
                continue
            seen.add(callee.start_ea)
            item = function_summary(callee, include_instructions=False)
            item["reference_ea"] = hex(ea)
            item["xref_type"] = int(xref.type)
            callees.append(item)

    weapon_table = []
    weapon_base = 0x17F807
    for weapon_id in range(46):
        ea = weapon_base + 0x1C * weapon_id
        raw = ida_bytes.get_bytes(ea, 0x1C) or b""
        weapon_table.append(
            {
                "weapon_id": weapon_id,
                "ea": hex(ea),
                "raw_hex": raw.hex(),
                "tech_id": ida_bytes.get_word(ea + 6),
                "category": ida_bytes.get_byte(ea + 8),
                "damage_max": ida_bytes.get_word(ea + 18),
                "raw_flags_at_22": ida_bytes.get_word(ea + 22),
                "fighter_eligible_bit_0x4": bool(ida_bytes.get_word(ea + 22) & 4),
            }
        )

    result = {
        "tool": "IDA Pro 9.4 IDAPython",
        "root_filename": ida_nalt.get_root_filename(),
        "input_path": ida_nalt.get_input_file_path(),
        "input_sha256": ida_nalt.retrieve_input_file_sha256().hex(),
        "processor": ida_ida.inf_get_procname(),
        "compiler_id": int(ida_ida.inf_get_cc_id()),
        "address_space": "IDA linear EA",
        "target": function_summary(fn),
        "callers": callers,
        "callees": callees,
        "weapon_table": weapon_table,
        "best_armor_lookup": [
            {
                "armor_index": index,
                "ea": hex(0x17F6C1 + 59 * index),
                "raw_word": ida_bytes.get_word(0x17F6C1 + 59 * index),
            }
            for index in range(6)
        ],
        "semantic_claim": {
            "level": "unknown",
            "note": "此匯出只保留原始定位與資料流；玩法語意須經人工審查後另行分級。",
        },
    }
    with open(output, "w", encoding="utf-8") as stream:
        json.dump(result, stream, ensure_ascii=False, indent=2, sort_keys=True)
        stream.write("\n")
    idc.qexit(0)


main()
