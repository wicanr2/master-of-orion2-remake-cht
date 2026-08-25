"""唯讀匯出原版新遊戲開局與難度的 IDA 證據。

輸出保留原始位址、名稱、運算元與 bytes；不改名、不施加型別、不儲存資料庫。
"""

import hashlib
import json
import os

import ida_auto
import ida_bytes
import ida_funcs
import ida_ida
import ida_kernwin
import ida_nalt
import ida_pro
import idautils
import idc


OUT = os.environ.get("MOO2_RE_OUT", "/out/newgame-rules.json")
SOURCE = os.environ.get("MOO2_RE_SOURCE", ida_nalt.get_input_file_path())

TARGET_FUNCTIONS = {
    "Init_Player_Tech_candidate": 0x5E55F,
    "Init_Homeworld_Colony2_candidate": 0x13A3D,
}

TARGET_DATA = {
    "difficulty_global_candidate": (0x199CB0, 1),
    "starting_tech_level_global_candidate": (0x199CB5, 1),
    "initial_building_cap_table": (0x13A3A, 3),
    "initial_building_priority_table": (0x17D8AC, 66),
    "fixed_starting_topic_table": (0x18111C, 12),
}


def sha256(path):
    digest = hashlib.sha256()
    with open(path, "rb") as handle:
        for chunk in iter(lambda: handle.read(1024 * 1024), b""):
            digest.update(chunk)
    return digest.hexdigest()


def instruction(ea):
    size = idc.get_item_size(ea)
    raw = ida_bytes.get_bytes(ea, size) or b""
    return {
        "ea": f"0x{ea:X}",
        "bytes": raw.hex(),
        "mnemonic": idc.print_insn_mnem(ea),
        "operand_0": idc.print_operand(ea, 0),
        "operand_1": idc.print_operand(ea, 1),
        "disassembly": idc.generate_disasm_line(ea, 0) or "<unavailable>",
    }


def original_name(ea):
    return idc.get_name(ea, idc.GN_VISIBLE) or "<unnamed>"


def function_record(ea):
    function = ida_funcs.get_func(ea)
    if function is None:
        return {"requested_ea": f"0x{ea:X}", "error": "no function boundary"}
    return {
        "requested_ea": f"0x{ea:X}",
        "start_ea": f"0x{function.start_ea:X}",
        "end_ea": f"0x{function.end_ea:X}",
        "original_name": original_name(function.start_ea),
        "instructions": [instruction(item) for item in idautils.FuncItems(function.start_ea)],
        "direct_callers": [instruction(ref) for ref in idautils.CodeRefsTo(function.start_ea, 0)],
    }


def data_record(ea, size):
    raw = ida_bytes.get_bytes(ea, size) or b""
    refs = []
    for ref in idautils.XrefsTo(ea, 0):
        function = ida_funcs.get_func(ref.frm)
        refs.append({
            "from": f"0x{ref.frm:X}",
            "type": int(ref.type),
            "function_start": f"0x{function.start_ea:X}" if function else None,
            "function_original_name": original_name(function.start_ea) if function else None,
            "instruction": instruction(ref.frm),
        })
    return {
        "ea": f"0x{ea:X}",
        "original_name": original_name(ea),
        "bytes": raw.hex(),
        "xrefs": refs,
    }


def matching_functions():
    needles = ("Difficult", "Init_Player_Tech", "Init_Homeworld", "Twiddle_Initial", "Adv_Civ")
    out = []
    for ea in idautils.Functions():
        name = original_name(ea)
        if any(needle.lower() in name.lower() for needle in needles):
            out.append({"ea": f"0x{ea:X}", "original_name": name})
    return out


def main():
    ida_auto.auto_wait()
    report = {
        "schema": "moo2.ida.re-evidence.v1",
        "evidence_scope": "static_only",
        "mutation": "none",
        "tool": {
            "name": "IDA Pro",
            "version": ida_kernwin.get_kernel_version(),
            "script": "tools/ida/audit_newgame_rules.py",
        },
        "input": {
            "ida_database_path": ida_nalt.get_input_file_path(),
            "source_path": SOURCE,
            "source_sha256": sha256(SOURCE),
            "processor": ida_ida.inf_get_procname(),
            "min_ea": f"0x{ida_ida.inf_get_min_ea():X}",
            "max_ea": f"0x{ida_ida.inf_get_max_ea():X}",
        },
        "address_basis": "IDA linear address; DOS/4GW LE object #1 code base 0x10000",
        "semantic_claim": {
            "level": "unknown_pending_review",
            "note": "本檔只匯出原始證據；語意與證據等級由受版控 RE 文件審查。",
        },
        "matching_original_names": matching_functions(),
        "functions": {name: function_record(ea) for name, ea in TARGET_FUNCTIONS.items()},
        "data": {name: data_record(ea, size) for name, (ea, size) in TARGET_DATA.items()},
    }
    os.makedirs(os.path.dirname(OUT), exist_ok=True)
    with open(OUT, "w", encoding="utf-8") as handle:
        json.dump(report, handle, ensure_ascii=False, indent=2)
        handle.write("\n")
    ida_pro.qexit(0)


if __name__ == "__main__":
    main()
