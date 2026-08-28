"""唯讀匯出 MOO2 殖民地回合計算與兩遍之間的原始呼叫鏈。"""

import csv
import hashlib
import json
import os

import ida_auto
import ida_bytes
import ida_funcs
import ida_hexrays
import ida_ida
import ida_kernwin
import ida_nalt
import ida_pro
import idautils
import idc


OUT = os.environ.get("MOO2_RE_OUT", "/out/colony-turn-chain.json")
SOURCE = os.environ.get("MOO2_RE_SOURCE", ida_nalt.get_input_file_path())
SYMBOLS = os.environ.get("MOO2_RE_SYMBOLS", "")
ROOTS = {
    "raw_Next_Turn_Calc": 0x136B3,
    "raw_Do_Colony_Calculations": 0xE2B31,
    "raw_Apply_All_Colony_Changes": 0xE3FDC,
    "raw_Compute_Blockades": 0xE5097,
    "raw_Move_Settlers": 0xFF212,
    "raw_Pre_Import_Computing": 0xE1D59,
    "raw_Colony_Industry_Per_Worker": 0xDEC95,
    "raw_Colony_Empire_Base_Industry_Produced": 0xDED47,
    "raw_Pass_Out_Imports": 0xDF8F0,
    "raw_Colony_Pop_Grows": 0xE1839,
    "raw_Colony_Specialty": 0xE1E1F,
    "raw_Update_Player_Stats": 0xE2710,
    "raw_Do_Player_Colony_Post_Production": 0xE2D09,
    "raw_Apply_Colony_Changes": 0xE3F6E,
    "raw_Extra_Caller_A": 0xE587C,
    "raw_Extra_Caller_B": 0xE5ADD,
    "raw_Colony_Environmental_Stuff": 0xE1CED,
    "raw_Colony_Replicators": 0xDF66F,
    "raw_Colony_BC_Maintenance": 0xE094F,
    "raw_Colony_Industry_Maintenance": 0xDF546,
    "raw_Colony_Industry_Production": 0xDEE1B,
    "raw_Colony_Job_Production": 0xDE280,
    "raw_Colony_Pop_Base_Prod_Produced": 0xDE22C,
    "raw_Gravity_Player_Production_Factor": 0xDDF2C,
    "raw_Apply_Production": 0xE36DF,
}


def sha256(path):
    digest = hashlib.sha256()
    with open(path, "rb") as stream:
        for chunk in iter(lambda: stream.read(1024 * 1024), b""):
            digest.update(chunk)
    return digest.hexdigest()


def load_symbols(path):
    if not path:
        return {}, None
    symbols = {}
    with open(path, newline="", encoding="utf-8") as stream:
        for row in csv.DictReader(stream, delimiter="\t"):
            symbols.setdefault(int(row["ea"], 16), []).append(row["name"])
    return symbols, sha256(path)


def insn(ea):
    size = idc.get_item_size(ea)
    return {
        "ea": f"0x{ea:X}",
        "bytes": (ida_bytes.get_bytes(ea, size) or b"").hex(),
        "mnem": idc.print_insn_mnem(ea),
        "op0": idc.print_operand(ea, 0),
        "op1": idc.print_operand(ea, 1),
        "line": idc.generate_disasm_line(ea, 0) or "<unavailable>",
    }


def function(ea, symbols):
    func = ida_funcs.get_func(ea)
    if not func:
        return {"requested": f"0x{ea:X}", "error": "no function"}
    items = list(idautils.FuncItems(func.start_ea))
    calls = []
    for order, call_ea in enumerate(
        (item for item in items if idc.print_insn_mnem(item).lower() == "call"), start=1
    ):
        target = idc.get_operand_value(call_ea, 0)
        target_func = ida_funcs.get_func(target)
        calls.append({
            "order": order,
            "call": insn(call_ea),
            "target": f"0x{target:X}",
            "target_function_start": f"0x{target_func.start_ea:X}" if target_func else None,
            "ida_original_name": idc.get_name(target) or "<unnamed>",
            "external_symbol_names_navigation_only": symbols.get(target, []),
        })
    pseudo = None
    if ida_hexrays.init_hexrays_plugin():
        try:
            pseudo = str(ida_hexrays.decompile(func.start_ea))
        except Exception as exc:
            pseudo = f"<decompile failed: {exc}>"
    return {
        "requested": f"0x{ea:X}",
        "start": f"0x{func.start_ea:X}",
        "end": f"0x{func.end_ea:X}",
        "ida_original_name": idc.get_name(func.start_ea) or "<unnamed>",
        "external_symbol_names_navigation_only": symbols.get(func.start_ea, []),
        "pseudocode_navigation_only": pseudo,
        "callers": [insn(ref) for ref in idautils.CodeRefsTo(func.start_ea, 0)],
        "instructions": [insn(item) for item in items],
        "calls": calls,
    }


def main():
    ida_auto.auto_wait()
    symbols, symbols_hash = load_symbols(SYMBOLS)
    report = {
        "schema": "moo2.ida.colony-turn-chain.v1",
        "evidence_scope": "static_only",
        "mutation": "none",
        "tool": {
            "name": "IDA Pro",
            "version": ida_kernwin.get_kernel_version(),
            "script": "tools/ida/audit_colony_turn_chain.py",
        },
        "input": {
            "database": ida_nalt.get_input_file_path(),
            "source": SOURCE,
            "source_sha256": sha256(SOURCE),
            "symbols": SYMBOLS or None,
            "symbols_sha256": symbols_hash,
            "processor": ida_ida.inf_get_procname(),
            "min_ea": f"0x{ida_ida.inf_get_min_ea():X}",
            "max_ea": f"0x{ida_ida.inf_get_max_ea():X}",
        },
        "address_basis": "IDA linear; DOS/4GW LE object #1",
        "data": {
            "raw_industry_per_worker_base_table": {
                "ea": "0xDD4B5",
                "bytes": (ida_bytes.get_bytes(0xDD4B5, 5) or b"").hex(),
                "values": [ida_bytes.get_byte(0xDD4B5 + index) for index in range(5)],
            },
            "raw_colony_bc_maintenance_modifier_table": {
                "ea": "0xDD4BA",
                "bytes": (ida_bytes.get_bytes(0xDD4BA, 10) or b"").hex(),
                "signed_values": [ida_bytes.get_byte(0xDD4BA + index) - 256
                                  if ida_bytes.get_byte(0xDD4BA + index) >= 128
                                  else ida_bytes.get_byte(0xDD4BA + index)
                                  for index in range(10)],
            },
            "raw_ai_job_difficulty_bonus_table": {
                "ea": "0xDD4D7",
                "bytes": (ida_bytes.get_bytes(0xDD4D7, 5) or b"").hex(),
                "signed_values": [ida_bytes.get_byte(0xDD4D7 + index) - 256
                                  if ida_bytes.get_byte(0xDD4D7 + index) >= 128
                                  else ida_bytes.get_byte(0xDD4D7 + index)
                                  for index in range(5)],
            },
            "raw_robotic_factory_bonus_table": {
                "ea": "0xDD4DC",
                "bytes": (ida_bytes.get_bytes(0xDD4DC, 5) or b"").hex(),
                "values": [ida_bytes.get_byte(0xDD4DC + index) for index in range(5)],
            },
            "raw_pollution_tolerance_by_planet_size_table": {
                "ea": "0xDD4E1",
                "bytes": (ida_bytes.get_bytes(0xDD4E1, 5) or b"").hex(),
                "values": [ida_bytes.get_byte(0xDD4E1 + index) for index in range(5)],
            },
        },
        "roots": {name: function(ea, symbols) for name, ea in ROOTS.items()},
    }
    os.makedirs(os.path.dirname(OUT), exist_ok=True)
    with open(OUT, "w", encoding="utf-8") as stream:
        json.dump(report, stream, ensure_ascii=False, indent=2)
        stream.write("\n")
    ida_pro.qexit(0)


if __name__ == "__main__":
    main()
