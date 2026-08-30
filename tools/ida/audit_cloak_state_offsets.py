"""匯出戰鬥艦紀錄 cloak state（+0x40/+0x41）的直接讀寫端。"""

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


TARGETS = {
    0x35D0D: "Defensive_Combat_Bonus_",
    0x36A63: "Ship_Specials_Defensive_Bonus_",
    0xACB4A: "Draw_Cloak_",
}


def digest(path):
    value = hashlib.sha256()
    with open(path, "rb") as source:
        for block in iter(lambda: source.read(1024 * 1024), b""):
            value.update(block)
    return value.hexdigest()


def symbols(path):
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


def function(ea, names):
    fn = ida_funcs.get_func(ea)
    if fn is None:
        return None
    return {
        "start": f"0x{fn.start_ea:X}",
        "end": f"0x{fn.end_ea:X}",
        "ida_name": ida_name.get_name(fn.start_ea) or "<unnamed>",
        "external_symbol": names.get(fn.start_ea),
    }


def window(ea, names, radius=16):
    fn = ida_funcs.get_func(ea)
    if fn is None:
        return {"site": f"0x{ea:X}", "owner": None}
    items = list(idautils.FuncItems(fn.start_ea))
    index = items.index(ea)
    return {
        "site": f"0x{ea:X}",
        "owner": function(ea, names),
        "instructions": [instruction(item) for item in items[max(0, index-radius):index+radius+1]],
    }


def direct_offset_sites(names):
    result = []
    for segment_start in idautils.Segments():
        segment_end = idc.get_segm_end(segment_start)
        for ea in idautils.Heads(segment_start, segment_end):
            if not ida_bytes.is_code(ida_bytes.get_flags(ea)):
                continue
            text = idc.generate_disasm_line(ea, 0) or ""
            if "+40h]" in text or "+41h]" in text:
                result.append(window(ea, names))
    return result


def target_records(names):
    result = {}
    for ea, navigation_name in TARGETS.items():
        fn = ida_funcs.get_func(ea)
        result[navigation_name] = {
            "function": function(ea, names),
            "instructions": [instruction(item) for item in idautils.FuncItems(fn.start_ea)],
            "callers": [window(xref.frm, names) for xref in idautils.XrefsTo(fn.start_ea, 0) if xref.iscode],
        }
    return result


def main():
    ida_auto.auto_wait()
    source = os.environ["MOO2_RE_SOURCE"]
    database = os.environ["MOO2_RE_DATABASE"]
    symbols_path = os.environ["MOO2_RE_SYMBOLS"]
    output = os.environ["MOO2_RE_OUT"]
    names = symbols(symbols_path)
    payload = {
        "schema": "moo2-cloak-state-offset-evidence-v1",
        "inputs": {
            "source": {"name": os.path.basename(source), "sha256": digest(source)},
            "database": {"name": os.path.basename(database), "sha256": digest(database)},
            "symbols": {"name": os.path.basename(symbols_path), "sha256": digest(symbols_path)},
            "ida_version": ida_kernwin.get_kernel_version(),
            "address_space": "IDA linear address in Orion2.exe.i64",
        },
        "mutation": "none; read-only export",
        "warning": "位移文字命中仍須由 0x139 戰鬥艦 stride 與基址資料流確認；外部符號只供導覽。",
        "direct_offset_sites": direct_offset_sites(names),
        "targets": target_records(names),
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
