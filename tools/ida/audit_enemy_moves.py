"""非破壞性匯出 Enemy Moves 字串、交叉參照及其函式上下文。"""

import hashlib
import json
import os

import ida_auto
import ida_bytes
import ida_funcs
import ida_ida
import ida_kernwin
import ida_name
import ida_pro
import idautils
import idc


PATTERNS = ("enemy moves", "enemy move")
ROOTS = {
    "enemy_ship_going_to_colony_candidate": 0x77F5D,
    "enemy_ship_heading_to_colony_candidate": 0xA13B0,
}


def digest(path):
    value = hashlib.sha256()
    with open(path, "rb") as source:
        for chunk in iter(lambda: source.read(1024 * 1024), b""):
            value.update(chunk)
    return value.hexdigest()


def instruction(ea):
    size = max(1, idc.get_item_size(ea))
    return {
        "ea": f"0x{ea:X}",
        "bytes": (ida_bytes.get_bytes(ea, size) or b"").hex(),
        "text": idc.generate_disasm_line(ea, 0) or "<unavailable>",
    }


def function_context(ea):
    function = ida_funcs.get_func(ea)
    if function is None:
        return {"site": instruction(ea), "function": None}
    callers = []
    for ref in idautils.CodeRefsTo(function.start_ea, 0):
        owner = ida_funcs.get_func(ref)
        callers.append({
            "site": instruction(ref),
            "function_start": f"0x{owner.start_ea:X}" if owner else None,
            "raw_name": ida_name.get_name(owner.start_ea) if owner else None,
        })
    return {
        "site": instruction(ea),
        "function": {
            "start": f"0x{function.start_ea:X}",
            "end_exclusive": f"0x{function.end_ea:X}",
            "raw_name": ida_name.get_name(function.start_ea) or "<unnamed>",
            "callers": callers,
            "instructions": [instruction(x) for x in idautils.FuncItems(function.start_ea)],
        },
    }


def function_record(ea):
    function = ida_funcs.get_func(ea)
    if function is None:
        return {"requested": f"0x{ea:X}", "error": "function missing"}
    context = function_context(function.start_ea)["function"]
    context["requested"] = f"0x{ea:X}"
    return context


def main():
    ida_auto.auto_wait()
    matches = []
    for item in idautils.Strings():
        value = str(item)
        if not any(pattern in value.lower() for pattern in PATTERNS):
            continue
        ea = int(item.ea)
        matches.append({
            "address": f"0x{ea:X}",
            "value": value,
            "refs": [function_context(ref.frm) for ref in idautils.XrefsTo(ea, 0)],
        })
    source = os.environ["MOO2_IDA_INPUT"]
    database = os.environ["MOO2_IDA_DATABASE"]
    output = {
        "contract": "原始字串／定位＋原始函式名＋指令 bytes；語意另於受版控 RE 文件分級",
        "ida_version": ida_kernwin.get_kernel_version(),
        "address_space": "IDA linear address",
        "processor": ida_ida.inf_get_procname(),
        "input": {"path": source, "sha256": digest(source)},
        "database": {"path": database, "sha256": digest(database)},
        "matches": matches,
        "roots": {name: function_record(ea) for name, ea in ROOTS.items()},
    }
    with open(os.environ["MOO2_IDA_OUTPUT"], "w", encoding="utf-8") as target:
        json.dump(output, target, ensure_ascii=False, indent=2)
    ida_pro.qexit(0)


if __name__ == "__main__":
    main()
