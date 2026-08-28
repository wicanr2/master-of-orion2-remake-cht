"""唯讀稽核 Stealthy Ships 與匿蹤特殊裝置在兩條戰鬥路徑的邊界。"""

import csv
import hashlib
import json
import os
import re
import traceback

import ida_auto
import ida_bytes
import ida_funcs
import ida_kernwin
import ida_name
import ida_pro
import ida_ua
import idautils
import idc


ROOTS = {
    "ship_has_special": 0x5D393,
    "ship_has_stealth_device": 0x5D3DB,
    "qload_ships": 0x416CF,
    "load_combat_ship": 0x4954A,
    "strategic_combat": 0x40148,
    "tactical_combat": 0x47939,
}
TEST_BIT_FIELD = 0x1276F0
SPECIAL_IDS = {6: "CLOAKING_DEVICE", 23: "PHASING_CLOAK", 31: "STEALTH_FIELD"}
SPECIAL_EFFECT_TYPE_BASE = 0x17EF0C
SPECIAL_RECORD_SIZE = 0x2F
TRAIT_PATTERN = re.compile(r"(?:\+|\b)8BBh\]", re.IGNORECASE)


def digest(path):
    out = hashlib.sha256()
    with open(path, "rb") as source:
        for block in iter(lambda: source.read(1024 * 1024), b""):
            out.update(block)
    return out.hexdigest()


def load_symbols(path):
    out = {}
    with open(path, "r", encoding="utf-8", newline="") as source:
        for row in csv.DictReader(source, delimiter="\t"):
            try:
                ea = int(row["ea"], 16)
            except (KeyError, TypeError, ValueError):
                continue
            out[ea] = row.get("name") or "<unnamed>"
    return out


def instruction(ea):
    decoded = ida_ua.insn_t()
    if ida_ua.decode_insn(decoded, ea) <= 0:
        return {"ea": f"0x{ea:X}", "error": "decode failed"}
    return {
        "ea": f"0x{ea:X}",
        "bytes": (ida_bytes.get_bytes(ea, decoded.size) or b"").hex(),
        "text": idc.generate_disasm_line(ea, 0) or "",
        "mnemonic": ida_ua.print_insn_mnem(ea),
        "operands": [idc.print_operand(ea, i) for i in range(8) if idc.print_operand(ea, i)],
        "code_refs_from": [f"0x{x:X}" for x in idautils.CodeRefsFrom(ea, 0)],
    }


def owner(ea, external):
    fn = ida_funcs.get_func(ea)
    if fn is None:
        return {"start": None, "ida_name": None, "external_symbol": None}
    return {"start": f"0x{fn.start_ea:X}", "ida_name": ida_name.get_name(fn.start_ea) or "<unnamed>",
            "external_symbol": external.get(fn.start_ea)}


def call_window(call_ea, external, radius=14):
    fn = ida_funcs.get_func(call_ea)
    if fn is None:
        return {"call": f"0x{call_ea:X}", "error": "no owner"}
    items = list(idautils.FuncItems(fn.start_ea))
    try:
        index = items.index(call_ea)
    except ValueError:
        index = 0
    return {
        "call": f"0x{call_ea:X}", "owner": owner(call_ea, external),
        "window": [instruction(x) for x in items[max(0, index-radius):min(len(items), index+radius+1)]],
    }


def function_record(ea, external):
    fn = ida_funcs.get_func(ea)
    if fn is None:
        return {"requested": f"0x{ea:X}", "error": "no function"}
    return {
        "requested": f"0x{ea:X}", "start": f"0x{fn.start_ea:X}", "end": f"0x{fn.end_ea:X}",
        "ida_name": ida_name.get_name(fn.start_ea) or "<unnamed>", "external_symbol": external.get(fn.start_ea),
        "chunks": [
            {"start": f"0x{s:X}", "end": f"0x{e:X}",
             "instructions": [instruction(x) for x in idautils.Heads(s, e)
                              if ida_bytes.is_code(ida_bytes.get_flags(x))]}
            for s, e in idautils.Chunks(fn.start_ea)
        ],
        "callers": [call_window(x.frm, external) for x in idautils.XrefsTo(fn.start_ea, 0) if x.iscode],
    }


def direct_trait_sites(external):
    out = []
    for start in idautils.Functions():
        for ea in idautils.FuncItems(start):
            text = idc.generate_disasm_line(ea, 0) or ""
            if TRAIT_PATTERN.search(text):
                out.append({"owner": owner(ea, external), "instruction": instruction(ea)})
    return out


def helper_calls(target, external):
    return [call_window(x.frm, external) for x in idautils.XrefsTo(target, 0) if x.iscode]


def inferred_special_id(call_ea, window):
    call_index = next((index for index, record in enumerate(window) if record.get("ea") == call_ea), len(window))
    for record in reversed(window[:call_index]):
        if record.get("mnemonic") != "mov" or len(record.get("operands", [])) < 2:
            continue
        if record["operands"][0].lower() != "edx":
            continue
        raw = record["operands"][1].lower().rstrip("h")
        try:
            value = int(raw, 16 if record["operands"][1].lower().endswith("h") else 10)
        except ValueError:
            continue
        return value
    return None


def bit_field_special_calls(external):
    out = []
    for record in helper_calls(TEST_BIT_FIELD, external):
        value = inferred_special_id(record["call"], record["window"])
        if value in SPECIAL_IDS:
            record["inferred_edx_special_id"] = value
            record["navigation_special_name"] = SPECIAL_IDS[value]
            record["evidence_level"] = "candidate; verify EDX is not redefined across control-flow join"
            out.append(record)
    return out


def special_device_records():
    out = []
    for special_id, navigation_name in SPECIAL_IDS.items():
        ea = SPECIAL_EFFECT_TYPE_BASE + SPECIAL_RECORD_SIZE * special_id
        out.append({
            "special_id": special_id,
            "navigation_name": navigation_name,
            "effect_type_ea": f"0x{ea:X}",
            "effect_type_u8": ida_bytes.get_byte(ea),
            "effect_value_s16": ida_bytes.get_word(ea + 1) - (
                0x10000 if ida_bytes.get_word(ea + 1) >= 0x8000 else 0
            ),
            "record_tail_bytes_from_effect_type": (
                ida_bytes.get_bytes(ea, SPECIAL_RECORD_SIZE) or b""
            ).hex(),
            "evidence_level": "confirmed raw bytes; semantic labels remain navigation-only",
        })
    return out


def table_xrefs(external):
    out = []
    for ea in range(SPECIAL_EFFECT_TYPE_BASE, SPECIAL_EFFECT_TYPE_BASE + SPECIAL_RECORD_SIZE * 39):
        for xref in idautils.XrefsTo(ea, 0):
            if xref.iscode:
                out.append(call_window(xref.frm, external, radius=8))
    return out


def main():
    ida_auto.auto_wait()
    source = os.environ["MOO2_RE_SOURCE"]
    database = os.environ["MOO2_RE_DATABASE"]
    symbols_path = os.environ["MOO2_RE_SYMBOLS"]
    out_path = os.environ["MOO2_RE_OUT"]
    external = load_symbols(symbols_path)
    payload = {
        "schema": "moo2-stealth-battle-boundary-evidence-v1",
        "inputs": {
            "source": {"name": os.path.basename(source), "sha256": digest(source)},
            "database": {"name": os.path.basename(database), "sha256": digest(database)},
            "symbols": {"name": os.path.basename(symbols_path), "sha256": digest(symbols_path)},
            "ida_version": ida_kernwin.get_kernel_version(),
            "address_space": "IDA linear address in Orion2.exe.i64",
        },
        "mutation": "none; read-only export",
        "warning": "absence claims require reviewing complete direct trait list and helper-mediated device consumers",
        "roots": {name: function_record(ea, external) for name, ea in ROOTS.items()},
        "direct_trait_0x8BB_sites": direct_trait_sites(external),
        "ship_has_special_calls": helper_calls(ROOTS["ship_has_special"], external),
        "ship_has_stealth_device_calls": helper_calls(ROOTS["ship_has_stealth_device"], external),
        "test_bit_field_calls_for_6_23_31": bit_field_special_calls(external),
        "special_device_records_6_23_31": special_device_records(),
        "direct_code_xrefs_into_effect_table_range": table_xrefs(external),
    }
    with open(out_path, "w", encoding="utf-8") as target:
        json.dump(payload, target, ensure_ascii=False, indent=2)
        target.write("\n")


if __name__ == "__main__":
    try:
        main()
    except Exception:
        traceback.print_exc()
        raise
    finally:
        ida_pro.qexit(0)
