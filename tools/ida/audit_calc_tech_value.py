"""唯讀匯出 Calc_Tech_Value／Choose_Tech_Application 與相關資料交叉參照。

保留原始函式名、位址、運算元與 bytes；不改名、不套型別、不儲存 IDB。
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


OUT = os.environ.get("MOO2_RE_OUT", "/out/calc-tech-value.json")
SOURCE = os.environ.get("MOO2_RE_SOURCE", ida_nalt.get_input_file_path())

FUNCTIONS = {
    "raw_sub_FC7AC": 0xFC7AC,
    "raw_sub_FC845": 0xFC845,
    "raw_sub_FD335": 0xFD335,
    "raw_sub_FD219": 0xFD219,
    "raw_sub_FD199": 0xFD199,
    "raw_sub_FD2F9": 0xFD2F9,
    "raw_sub_FE96F": 0xFE96F,
    "raw_sub_589D6": 0x589D6,
    "raw_sub_E412B": 0xE412B,
    "raw_sub_12983": 0x12983,
}

DATA = {
    "difficulty": (0x199CB0, 1),
    "tech_research_level_values": (0x18360C, 30 * 4),
    "tech_categories": (0x17D196, 49 * 2),
    "race_trait_level_values": (0x17D1F9, 10 * 3),
    "tech_items": (0x17E07F, 212 * 13),
    "topic_records": (0x17D90C, 83 * 23),
    "race_raw27_table": (0x181090, 13 * 10),
    "best_category_region": (0x1AB10C, 0x18),
}


def sha256(path):
    h = hashlib.sha256()
    with open(path, "rb") as f:
        for chunk in iter(lambda: f.read(1024 * 1024), b""):
            h.update(chunk)
    return h.hexdigest()


def insn(ea):
    size = idc.get_item_size(ea)
    raw = ida_bytes.get_bytes(ea, size) or b""
    return {
        "ea": f"0x{ea:X}",
        "bytes": raw.hex(),
        "mnem": idc.print_insn_mnem(ea),
        "op0": idc.print_operand(ea, 0),
        "op1": idc.print_operand(ea, 1),
        "line": idc.generate_disasm_line(ea, 0) or "<unavailable>",
    }


def func_record(requested):
    f = ida_funcs.get_func(requested)
    if not f:
        return {"requested": f"0x{requested:X}", "error": "no function"}
    callers = []
    for ea in idautils.CodeRefsTo(f.start_ea, 0):
        cf = ida_funcs.get_func(ea)
        callers.append({
            "instruction": insn(ea),
            "function_start": f"0x{cf.start_ea:X}" if cf else None,
            "function_name": idc.get_name(cf.start_ea) if cf else None,
        })
    callees = []
    for ea in idautils.FuncItems(f.start_ea):
        if idc.print_insn_mnem(ea).lower() != "call":
            continue
        targets = list(idautils.CodeRefsFrom(ea, 0))
        callees.append({"instruction": insn(ea), "targets": [f"0x{x:X}" for x in targets]})
    return {
        "requested": f"0x{requested:X}",
        "start": f"0x{f.start_ea:X}",
        "end": f"0x{f.end_ea:X}",
        "original_name": idc.get_name(f.start_ea) or "<unnamed>",
        "instructions": [insn(ea) for ea in idautils.FuncItems(f.start_ea)],
        "callers": callers,
        "callees": callees,
    }


def xrefs_to(ea):
    out = []
    for ref in idautils.XrefsTo(ea, 0):
        f = ida_funcs.get_func(ref.frm)
        out.append({
            "from": f"0x{ref.frm:X}",
            "type": int(ref.type),
            "function_start": f"0x{f.start_ea:X}" if f else None,
            "function_name": idc.get_name(f.start_ea) if f else None,
            "instruction": insn(ref.frm),
        })
    return out


def data_record(ea, size):
    raw = ida_bytes.get_bytes(ea, size) or b""
    return {
        "ea": f"0x{ea:X}",
        "original_name": idc.get_name(ea) or "<unnamed>",
        "bytes": raw.hex(),
        "xrefs": xrefs_to(ea),
    }


def player_offset_mentions(offsets):
    needles = tuple(f"{x:X}h".lower() for x in offsets)
    out = []
    for ea in idautils.Functions():
        hits = []
        for item in idautils.FuncItems(ea):
            line = (idc.generate_disasm_line(item, 0) or "").lower()
            if any(n in line for n in needles):
                hits.append(insn(item))
        if hits:
            out.append({
                "function_start": f"0x{ea:X}",
                "function_name": idc.get_name(ea) or "<unnamed>",
                "hits": hits,
            })
    return out


def main():
    ida_auto.auto_wait()
    report = {
        "schema": "moo2.ida.re-evidence.v1",
        "evidence_scope": "static_only",
        "mutation": "none",
        "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version(),
                 "script": "tools/ida/audit_calc_tech_value.py"},
        "input": {
            "database": ida_nalt.get_input_file_path(),
            "source": SOURCE,
            "source_sha256": sha256(SOURCE),
            "processor": ida_ida.inf_get_procname(),
            "min_ea": f"0x{ida_ida.inf_get_min_ea():X}",
            "max_ea": f"0x{ida_ida.inf_get_max_ea():X}",
        },
        "address_basis": "IDA linear; DOS/4GW LE object #1",
        "functions": {name: func_record(ea) for name, ea in FUNCTIONS.items()},
        "data": {name: data_record(ea, size) for name, (ea, size) in DATA.items()},
        "player_offset_mentions": player_offset_mentions([0x28, 0x205, 0x206, 0x89F]),
    }
    os.makedirs(os.path.dirname(OUT), exist_ok=True)
    with open(OUT, "w", encoding="utf-8") as f:
        json.dump(report, f, ensure_ascii=False, indent=2)
        f.write("\n")
    ida_pro.qexit(0)


if __name__ == "__main__":
    main()
