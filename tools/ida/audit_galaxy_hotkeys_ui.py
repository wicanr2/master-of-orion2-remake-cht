"""非破壞性匯出星圖快捷鍵、測距與存檔相關的原始定位證據。"""

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


ROOTS = {
    "galaxy_input_dispatch_candidate": 0x825A8,
    "parsecs_between_points": 0xEBE79,
}
NAME_PATTERNS = ("save", "parsec", "distance", "cycle_ship_icons")


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
        "code_refs": [f"0x{x:X}" for x in idautils.CodeRefsFrom(ea, 0)],
        "data_refs": [f"0x{x:X}" for x in idautils.DataRefsFrom(ea)],
    }


def function_record(ea, include_instructions=True):
    function = ida_funcs.get_func(ea)
    if function is None:
        return {"requested": f"0x{ea:X}", "error": "function missing"}
    record = {
        "requested": f"0x{ea:X}",
        "start": f"0x{function.start_ea:X}",
        "end_exclusive": f"0x{function.end_ea:X}",
        "raw_name": ida_name.get_name(function.start_ea) or "<unnamed>",
        "callers": [],
    }
    for ref in idautils.CodeRefsTo(function.start_ea, 0):
        owner = ida_funcs.get_func(ref)
        record["callers"].append({
            "site": instruction(ref),
            "function_start": f"0x{owner.start_ea:X}" if owner else None,
            "raw_name": ida_name.get_name(owner.start_ea) if owner else None,
        })
    if include_instructions:
        record["instructions"] = [instruction(x) for x in idautils.FuncItems(function.start_ea)]
    return record


def matching_functions():
    matches = []
    for ea in idautils.Functions():
        name = ida_name.get_name(ea) or ""
        if any(pattern in name.lower() for pattern in NAME_PATTERNS):
            matches.append(function_record(ea, include_instructions=False))
    return matches


def matching_strings():
    matches = []
    for value in idautils.Strings():
        text = str(value)
        lowered = text.lower()
        if any(pattern in lowered for pattern in ("save", "parsec", "distance")):
            matches.append({
                "ea": f"0x{int(value.ea):X}",
                "text": text,
                "xrefs": [f"0x{x.frm:X}" for x in idautils.XrefsTo(value.ea)],
            })
    return matches


def main():
    ida_auto.auto_wait()
    source = os.environ["MOO2_IDA_INPUT"]
    database = os.environ["MOO2_IDA_DATABASE"]
    output = {
        "contract": "原始定位＋原始名稱＋指令 bytes；語意由受版控 RE 文件分級",
        "ida_version": ida_kernwin.get_kernel_version(),
        "address_space": "IDA linear address (DOS/4GW image)",
        "processor": ida_ida.inf_get_procname(),
        "input": {"path": source, "sha256": digest(source)},
        "database": {"path": database, "sha256": digest(database)},
        "roots": {name: function_record(ea) for name, ea in ROOTS.items()},
        "matching_functions": matching_functions(),
        "matching_strings": matching_strings(),
    }
    with open(os.environ["MOO2_IDA_OUTPUT"], "w", encoding="utf-8") as target:
        json.dump(output, target, ensure_ascii=False, indent=2)
    ida_pro.qexit(0)


if __name__ == "__main__":
    main()
