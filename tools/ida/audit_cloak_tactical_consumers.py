"""唯讀稽核 raw 6／23／31 匿蹤裝置在格子戰術的間接 consumer。"""

import csv
import hashlib
import json
import os
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
    "ship_specials_defensive_bonus": 0x36A63,
    "ship_specials_offensive_bonus": 0x36B46,
    "get_combat_ship_special_id": 0x37254,
    "special_is_on": 0x3AFE5,
    "special_is_turned_on": 0x3B004,
    "pre_weapon_specials": 0x3B332,
    "does_combat_ship_have_special": 0x4B0D3,
    "init_special_devices": 0x4C9F6,
    "draw_cloak": 0xACB4A,
}
SPECIAL_IDS = {6: "CLOAKING_DEVICE", 23: "PHASING_CLOAK", 31: "STEALTH_FIELD"}
HELPERS = {
    0x3AFE5: "Special_Is_On_",
    0x3B004: "Special_Is_Turned_On_",
    0x4B0D3: "Does_Combat_Ship_Have_Special_",
}
DATA_RANGES = {
    "cloaking_spcl": (0x181150, 0x78),
    "special_weapons_spcl": (0x1811C8, 0x14),
    "special_device_spcl": (0x1811DC, 0x30),
    "defense_specials_f_spcl": (0x18120C, 0x4C),
    "defense_specials_t_spcl": (0x181258, 0x4C),
}


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
                out[int(row["ea"], 16)] = row.get("name") or "<unnamed>"
            except (KeyError, TypeError, ValueError):
                continue
    return out


def insn(ea):
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


def owner(ea, symbols):
    fn = ida_funcs.get_func(ea)
    if fn is None:
        return {"start": None, "ida_name": None, "external_symbol": None}
    return {
        "start": f"0x{fn.start_ea:X}",
        "ida_name": ida_name.get_name(fn.start_ea) or "<unnamed>",
        "external_symbol": symbols.get(fn.start_ea),
    }


def window(ea, symbols, radius=18):
    fn = ida_funcs.get_func(ea)
    if fn is None:
        return {"site": f"0x{ea:X}", "error": "no owner"}
    items = list(idautils.FuncItems(fn.start_ea))
    try:
        index = items.index(ea)
    except ValueError:
        index = 0
    return {
        "site": f"0x{ea:X}",
        "owner": owner(ea, symbols),
        "instructions": [insn(x) for x in items[max(0, index-radius):index+radius+1]],
    }


def function_record(ea, symbols):
    fn = ida_funcs.get_func(ea)
    if fn is None:
        return {"requested": f"0x{ea:X}", "error": "no function"}
    return {
        "requested": f"0x{ea:X}",
        "start": f"0x{fn.start_ea:X}",
        "end": f"0x{fn.end_ea:X}",
        "ida_name": ida_name.get_name(fn.start_ea) or "<unnamed>",
        "external_symbol": symbols.get(fn.start_ea),
        "chunks": [
            {
                "start": f"0x{start:X}",
                "end": f"0x{end:X}",
                "instructions": [insn(x) for x in idautils.Heads(start, end)
                                 if ida_bytes.is_code(ida_bytes.get_flags(x))],
            }
            for start, end in idautils.Chunks(fn.start_ea)
        ],
        "callers": [window(x.frm, symbols) for x in idautils.XrefsTo(fn.start_ea, 0) if x.iscode],
    }


def parse_immediate(text):
    raw = text.strip().lower()
    if raw.endswith("h"):
        raw = raw[:-1]
        base = 16
    else:
        base = 10
    try:
        return int(raw, base)
    except ValueError:
        return None


def nearby_special_constants(call_ea):
    fn = ida_funcs.get_func(call_ea)
    if fn is None:
        return []
    items = list(idautils.FuncItems(fn.start_ea))
    try:
        index = items.index(call_ea)
    except ValueError:
        return []
    found = []
    for ea in items[max(0, index-24):index]:
        operands = [idc.print_operand(ea, i) for i in range(8) if idc.print_operand(ea, i)]
        for operand in operands:
            value = parse_immediate(operand)
            if value in SPECIAL_IDS:
                found.append({
                    "instruction": insn(ea),
                    "special_id": value,
                    "navigation_name": SPECIAL_IDS[value],
                    "evidence_level": "candidate proximity only; verify register/control-flow provenance",
                })
    return found


def helper_calls(symbols):
    out = []
    for target, navigation_name in HELPERS.items():
        for xref in idautils.XrefsTo(target, 0):
            if not xref.iscode:
                continue
            out.append({
                "target": f"0x{target:X}",
                "navigation_name": navigation_name,
                "call": window(xref.frm, symbols),
                "nearby_special_constants": nearby_special_constants(xref.frm),
            })
    return out


def data_records():
    out = {}
    for name, (ea, size) in DATA_RANGES.items():
        raw = ida_bytes.get_bytes(ea, size) or b""
        out[name] = {
            "ea": f"0x{ea:X}",
            "size": size,
            "bytes": raw.hex(),
            "u16_le": [raw[i] | (raw[i+1] << 8) for i in range(0, len(raw)-1, 2)],
            "code_xrefs": [window(x.frm, symbols_global) for x in idautils.XrefsTo(ea, 0) if x.iscode],
            "evidence_level": "confirmed raw bytes; external symbol is navigation-only",
        }
    return out


symbols_global = {}


def main():
    global symbols_global
    ida_auto.auto_wait()
    source = os.environ["MOO2_RE_SOURCE"]
    database = os.environ["MOO2_RE_DATABASE"]
    symbols_path = os.environ["MOO2_RE_SYMBOLS"]
    output = os.environ["MOO2_RE_OUT"]
    symbols_global = load_symbols(symbols_path)
    payload = {
        "schema": "moo2-cloak-tactical-consumers-evidence-v1",
        "inputs": {
            "source": {"name": os.path.basename(source), "sha256": digest(source)},
            "database": {"name": os.path.basename(database), "sha256": digest(database)},
            "symbols": {"name": os.path.basename(symbols_path), "sha256": digest(symbols_path)},
            "ida_version": ida_kernwin.get_kernel_version(),
            "address_space": "IDA linear address in Orion2.exe.i64",
        },
        "mutation": "none; read-only export",
        "warning": "nearby constants are candidates until argument provenance is reviewed",
        "roots": {name: function_record(ea, symbols_global) for name, ea in ROOTS.items()},
        "helper_calls": helper_calls(symbols_global),
        "named_data_ranges": data_records(),
    }
    with open(output, "w", encoding="utf-8") as target:
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
