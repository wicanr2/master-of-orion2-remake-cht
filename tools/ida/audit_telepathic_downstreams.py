"""匯出 Telepathic 尚未閉合的五組 raw 下游。"""

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
    "boarding_action_type": 0x2C129,
    "ai_boarding_dispatch": 0x2BF73,
    "resolve_capture": 0x37DA8,
    "capture_ship": 0x38312,
    "mind_control_colony": 0xC622A,
    "general_skill_at_star": 0xC6052,
    "leader_at_star_check": 0x9467D,
    "enemy_colony_worth": 0xD8D11,
    "calc_tech_value": 0xFC845,
}


def digest(path):
    value = hashlib.sha256()
    with open(path, "rb") as source:
        for block in iter(lambda: source.read(1024 * 1024), b""):
            value.update(block)
    return value.hexdigest()


def load_symbols(path):
    result = {}
    with open(path, "r", encoding="utf-8", newline="") as source:
        for row in csv.DictReader(source, delimiter="\t"):
            try:
                result[int(row["ea"], 16)] = row.get("name") or "<unnamed>"
            except (KeyError, TypeError, ValueError):
                pass
    return result


def instruction(ea):
    decoded = ida_ua.insn_t()
    ida_ua.decode_insn(decoded, ea)
    return {
        "ea": f"0x{ea:X}",
        "bytes": (ida_bytes.get_bytes(ea, decoded.size) or b"").hex(),
        "text": idc.generate_disasm_line(ea, 0) or "",
    }


def owner(ea, names):
    fn = ida_funcs.get_func(ea)
    if fn is None:
        return None
    return {
        "start": f"0x{fn.start_ea:X}", "end": f"0x{fn.end_ea:X}",
        "ida_name": ida_name.get_name(fn.start_ea) or "<unnamed>",
        "external_symbol": names.get(fn.start_ea),
    }


def window(ea, names, radius=24):
    fn = ida_funcs.get_func(ea)
    if fn is None:
        return {"site": f"0x{ea:X}", "owner": None}
    items = list(idautils.FuncItems(fn.start_ea))
    index = items.index(ea)
    return {
        "site": f"0x{ea:X}", "owner": owner(ea, names),
        "instructions": [instruction(item) for item in items[max(0, index-radius):index+radius+1]],
    }


def function_record(ea, names):
    fn = ida_funcs.get_func(ea)
    return {
        "function": owner(ea, names),
        "chunks": [
            {"start": f"0x{start:X}", "end": f"0x{end:X}",
             "instructions": [instruction(item) for item in idautils.Heads(start, end)
                              if ida_bytes.is_code(ida_bytes.get_flags(item))]}
            for start, end in idautils.Chunks(fn.start_ea)
        ],
        "callers": [window(xref.frm, names) for xref in idautils.XrefsTo(fn.start_ea, 0) if xref.iscode],
    }


def direct_b0_sites(names):
    result = []
    for segment_start in idautils.Segments():
        for ea in idautils.Heads(segment_start, idc.get_segm_end(segment_start)):
            if not ida_bytes.is_code(ida_bytes.get_flags(ea)):
                continue
            text = idc.generate_disasm_line(ea, 0) or ""
            if "+0B0h]" in text or "+0B0h+" in text:
                result.append(window(ea, names))
    return result


def main():
    ida_auto.auto_wait()
    source = os.environ["MOO2_RE_SOURCE"]
    database = os.environ["MOO2_RE_DATABASE"]
    symbols_path = os.environ["MOO2_RE_SYMBOLS"]
    output = os.environ["MOO2_RE_OUT"]
    names = load_symbols(symbols_path)
    payload = {
        "schema": "moo2-telepathic-downstreams-evidence-v1",
        "inputs": {
            "source": {"name": os.path.basename(source), "sha256": digest(source)},
            "database": {"name": os.path.basename(database), "sha256": digest(database)},
            "symbols": {"name": os.path.basename(symbols_path), "sha256": digest(symbols_path)},
            "ida_version": ida_kernwin.get_kernel_version(),
            "address_space": "IDA linear address in Orion2.exe.i64",
        },
        "mutation": "none; read-only export",
        "warning": "外部符號只供導覽；+0xB0 文字命中須以 0x139 stride 與基址資料流排除假陽性。",
        "roots": {name: function_record(ea, names) for name, ea in ROOTS.items()},
        "direct_combat_ship_b0_candidates": direct_b0_sites(names),
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
