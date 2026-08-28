"""非破壞性匯出 MOO2 戰機炸彈、目標傷害與中隊生命週期鏈。"""

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
    "raw_Check_For_Launched_Fighters": 0x29480,
    "raw_Target_Evaluator": 0x2A239,
    "raw_Target_Ship_Value": 0x2A46A,
    "raw_Fighter_Pilot_Bonus": 0x35E6A,
    "raw_Best_Fighter_Pilot_Bonus": 0x35EAE,
    "raw_Apply_Damage_To_Facing": 0x39985,
    "raw_Weapon_In_Range": 0x3A0B9,
    "raw_Weapon_Target_Check": 0x3A7AA,
    "raw_sub_3AC20": 0x3AC20,
    "raw_sub_3AD57": 0x3AD57,
    "raw_Missile_Owner": 0x3BB3D,
    "raw_Missile_Type": 0x3BB50,
    "raw_Kill_Missile": 0x3BC80,
    "raw_Create_Missile_Runtime": 0x3C892,
    "raw_Runtime_Target_Alive": 0x3D299,
    "raw_Resolve_Missile_Runtime": 0x3D2DF,
    "raw_Draw_Fighter": 0x3DBD1,
    "raw_Retarget_Missile": 0x3DDD8,
    "raw_Fighter_Ocv": 0x3DF8D,
    "raw_Fighter_Target_Defense": 0x3DFE0,
    "raw_Fighter_Combat_SFX": 0x3E1A4,
    "raw_Tactical_Weapon_Dispatch": 0x39F1D,
    "raw_Do_Combat_Turn": 0x42F7F,
    "raw_Load_Combat_Ship": 0x4954A,
}

RAW_DATA = {
    "fighter_bomb_type": (0x199242, 32),
    "fighter_beam_type": (0x199254, 32),
    "weapon_records_head": (0x17F815, 0x1C * 8),
}


def digest(path):
    h = hashlib.sha256()
    with open(path, "rb") as stream:
        for chunk in iter(lambda: stream.read(1024 * 1024), b""):
            h.update(chunk)
    return h.hexdigest()


def load_symbols(path):
    symbols = {}
    if not path:
        return symbols, None
    with open(path, newline="", encoding="utf-8") as stream:
        for row in csv.DictReader(stream, delimiter="\t"):
            symbols.setdefault(int(row["ea"], 16), []).append(row["name"])
    return symbols, digest(path)


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
    fn = ida_funcs.get_func(ea)
    if fn is None:
        return {"requested": f"0x{ea:X}", "error": "function missing"}
    raw = ida_bytes.get_bytes(fn.start_ea, fn.end_ea - fn.start_ea) or b""
    pseudo = None
    if ida_hexrays.init_hexrays_plugin():
        try:
            pseudo = str(ida_hexrays.decompile(fn.start_ea))
        except Exception as exc:
            pseudo = f"<decompile failed: {exc}>"
    callers = []
    for xref in idautils.XrefsTo(fn.start_ea, 0):
        caller_fn = ida_funcs.get_func(xref.frm)
        if caller_fn is None:
            continue
        item = instruction(xref.frm)
        item.update({
            "caller_function_start": f"0x{caller_fn.start_ea:X}",
            "caller_original_name": ida_name.get_name(caller_fn.start_ea) or "<unnamed>",
            "caller_external_symbol_names_navigation_only": symbols.get(caller_fn.start_ea, []),
        })
        callers.append(item)
    calls = []
    for item_ea in idautils.FuncItems(fn.start_ea):
        if idc.print_insn_mnem(item_ea).lower() != "call":
            continue
        target = idc.get_operand_value(item_ea, 0)
        target_fn = ida_funcs.get_func(target)
        calls.append({
            "callsite": instruction(item_ea),
            "target": f"0x{target:X}",
            "target_function_start": f"0x{target_fn.start_ea:X}" if target_fn else None,
            "ida_original_name": ida_name.get_name(target) or "<unnamed>",
            "external_symbol_names_navigation_only": symbols.get(target, []),
        })
    return {
        "requested": f"0x{ea:X}",
        "original_name": ida_name.get_name(fn.start_ea) or "<unnamed>",
        "external_symbol_names_navigation_only": symbols.get(fn.start_ea, []),
        "start_ea": f"0x{fn.start_ea:X}",
        "end_ea": f"0x{fn.end_ea:X}",
        "bytes_sha256": hashlib.sha256(raw).hexdigest(),
        "pseudocode_navigation_only": pseudo,
        "instructions": [instruction(x) for x in idautils.FuncItems(fn.start_ea)],
        "callers": callers,
        "calls": calls,
    }


def main():
    ida_auto.auto_wait()
    source = os.environ["MOO2_RE_SOURCE"]
    database = os.environ["MOO2_RE_DATABASE"]
    symbols, symbols_hash = load_symbols(os.environ.get("MOO2_RE_SYMBOLS", ""))
    report = {
        "schema": "moo2.ida.fighter-bomb.v1",
        "evidence_scope": "static_only",
        "mutation": "none",
        "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version(),
                 "script": "tools/ida/audit_fighter_bomb.py"},
        "input": {"file": os.path.basename(source), "source_sha256": digest(source),
                  "database_sha256": digest(database), "symbols_sha256": symbols_hash,
                  "processor": ida_ida.inf_get_procname()},
        "address_basis": "IDA linear; DOS/4GW LE object #1",
        "semantic_status": "reviewed_with_per_item_confidence_in_fighter_runtime_audit_20260828",
        "roots": {name: function_record(ea, symbols) for name, ea in ROOTS.items()},
        "raw_data": {name: {"ea": f"0x{ea:X}", "bytes": (ida_bytes.get_bytes(ea, size) or b"").hex()}
                     for name, (ea, size) in RAW_DATA.items()},
    }
    with open(os.environ["MOO2_RE_OUT"], "w", encoding="utf-8") as stream:
        json.dump(report, stream, ensure_ascii=False, indent=2)
        stream.write("\n")


try:
    main()
except Exception:
    error = traceback.format_exc()
    out = os.environ.get("MOO2_RE_OUT", "/tmp/fighter-bomb.json")
    with open(out + ".error", "w", encoding="utf-8") as stream:
        stream.write(error)
    ida_kernwin.msg(error + "\n")
    ida_pro.qexit(1)
else:
    ida_pro.qexit(0)
