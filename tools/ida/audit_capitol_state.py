"""非破壞性匯出 MOO2 Capitol 行星、攻陷與重建狀態鏈。"""

import hashlib
import json
import os
import re

import ida_auto
import ida_bytes
import ida_funcs
import ida_ida
import ida_kernwin
import ida_name
import ida_pro
import idautils
import idc


ROOTS = {
    # 外部符號名稱只作導覽；RE 文件仍保留 raw 位址與證據分級。
    "raw_Player_Capitol_Fallback_external_adjacent": 0xC5F5C,
    "raw_Player_Capitol_World_external_symbol": 0xC5FB0,
    "raw_Player_Capitol_World_caller": 0xC7ADA,
    "raw_Colony_Morale_external_symbol": 0xDDB25,
    "raw_Show_Morale_external_symbol": 0xDDEFB,
    "raw_Mixed_Race_Morale_external_symbol": 0xDDAD4,
    "raw_colony_137_writer_3B627": 0x3B627,
    "raw_colony_137_writer_4C9F6": 0x4C9F6,
    "raw_colony_137_consumer_E3456": 0xE3456,
    "raw_colony_137_consumer_ECF41": 0xECF41,
    "raw_colony_capture_transfer_ECBF7": 0xECBF7,
    "raw_reassign_lost_capitol_ECB65": 0xECB65,
    "raw_colony_137_consumer_ED260": 0xED260,
    "raw_Resolve_Ground_Combat_external_symbol": 0xEC601,
    "raw_building_completion_consumer": 0x13FD9,
    "raw_colony_building_score": 0xD0036,
}


def digest(path):
    h = hashlib.sha256()
    with open(path, "rb") as source:
        for chunk in iter(lambda: source.read(1024 * 1024), b""):
            h.update(chunk)
    return h.hexdigest()


def instruction(ea):
    size = max(1, idc.get_item_size(ea))
    return {
        "ea": f"0x{ea:X}",
        "bytes": (ida_bytes.get_bytes(ea, size) or b"").hex(),
        "text": idc.generate_disasm_line(ea, 0) or "<unavailable>",
        "code_refs": [f"0x{x:X}" for x in idautils.CodeRefsFrom(ea, 0)],
        "data_refs": [f"0x{x:X}" for x in idautils.DataRefsFrom(ea)],
    }


def context(fn_start, ea, radius=10):
    items = list(idautils.FuncItems(fn_start))
    try:
        index = items.index(ea)
    except ValueError:
        return []
    return [instruction(x) for x in items[max(0, index-radius):index+radius+1]]


def operand_mentions_offset(ea, offset):
    operands = " ".join(idc.print_operand(ea, i) for i in range(2)).lower()
    for match in re.finditer(r"\+\s*([0-9a-f]+)h", operands):
        if int(match.group(1), 16) == offset:
            return True
    for match in re.finditer(r"\+\s*([0-9]+)(?=\s*[\],])", operands):
        if int(match.group(1), 10) == offset:
            return True
    return False


def function_record(ea):
    fn = ida_funcs.get_func(ea)
    if fn is None:
        return {"requested": f"0x{ea:X}", "error": "function missing"}
    callers = []
    for ref in idautils.CodeRefsTo(fn.start_ea, 0):
        owner = ida_funcs.get_func(ref)
        callers.append({
            "site": instruction(ref),
            "function_start": f"0x{owner.start_ea:X}" if owner else None,
            "raw_name": ida_name.get_name(owner.start_ea) if owner else None,
            "context": context(owner.start_ea, ref) if owner else [],
        })
    return {
        "start": f"0x{fn.start_ea:X}",
        "end_exclusive": f"0x{fn.end_ea:X}",
        "raw_name": ida_name.get_name(fn.start_ea) or "<unnamed>",
        "callers": callers,
        "instructions": [instruction(x) for x in idautils.FuncItems(fn.start_ea)],
    }


def main():
    ida_auto.auto_wait()
    source = os.environ["MOO2_IDA_INPUT"]
    database = os.environ["MOO2_IDA_DATABASE"]

    offset_refs = []
    named_candidates = []
    keywords = ("capitol", "capital", "capture", "conquer", "colony", "morale", "building")
    for fn_start in idautils.Functions():
        raw_name = ida_name.get_name(fn_start) or "<unnamed>"
        if any(word in raw_name.lower() for word in keywords):
            named_candidates.append({
                "start": f"0x{fn_start:X}",
                "end_exclusive": f"0x{ida_funcs.get_func(fn_start).end_ea:X}",
                "raw_name": raw_name,
            })
        for ea in idautils.FuncItems(fn_start):
            if operand_mentions_offset(ea, 0x29):
                offset_refs.append({
                    "function_start": f"0x{fn_start:X}",
                    "raw_name": raw_name,
                    "site": instruction(ea),
                    "context": context(fn_start, ea),
                })

    output = {
        "contract": "原始定位＋原始名稱＋指令 bytes；語意須由受版控 RE 文件另行分級",
        "ida_version": ida_kernwin.get_kernel_version(),
        "address_space": "IDA linear address (DOS/4GW image)",
        "processor": ida_ida.inf_get_procname(),
        "input": {"path": source, "sha256": digest(source)},
        "database": {"path": database, "sha256": digest(database)},
        "roots": {name: function_record(ea) for name, ea in ROOTS.items()},
        "direct_plus_29_refs": offset_refs,
        "named_function_candidates": named_candidates,
    }
    with open(os.environ["MOO2_IDA_OUTPUT"], "w", encoding="utf-8") as target:
        json.dump(output, target, ensure_ascii=False, indent=2)
    ida_pro.qexit(0)


if __name__ == "__main__":
    main()
