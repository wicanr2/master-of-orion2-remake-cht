"""非破壞性匯出 MOO2 帝國命名與旗色選擇的函式、caller 與原始指令。"""

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
    "raw_Player_Name": 0xE5E09,
    "raw_Add_Banner_Field": 0xEFABA,
    "raw_Draw_Multi_Player_Banner": 0xF27B7,
    "raw_Request_Banner_Color": 0xFBEE1,
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
        })
    return {
        "start": f"0x{fn.start_ea:X}",
        "end_exclusive": f"0x{fn.end_ea:X}",
        "raw_name": ida_name.get_name(fn.start_ea) or "<unnamed>",
        "callers": callers,
        "instructions": [instruction(x) for x in idautils.FuncItems(fn.start_ea)],
    }


def string_records():
    needles = ("NAME YOUR", "BANNER", "EMPIRE NAME", "LEADER NAME", "HOMEWORLD")
    records = []
    for item in idautils.Strings():
        value = str(item)
        if not any(needle in value.upper() for needle in needles):
            continue
        refs = []
        for ref in idautils.DataRefsTo(item.ea):
            owner = ida_funcs.get_func(ref)
            refs.append({
                "site": instruction(ref),
                "function_start": f"0x{owner.start_ea:X}" if owner else None,
                "raw_name": ida_name.get_name(owner.start_ea) if owner else None,
            })
        records.append({"ea": f"0x{item.ea:X}", "value": value, "refs": refs})
    return records


def main():
    ida_auto.auto_wait()
    source = os.environ["MOO2_IDA_INPUT"]
    database = os.environ["MOO2_IDA_DATABASE"]
    output = {
        "contract": "原始定位＋原始名稱＋指令 bytes；語意須由受版控 RE 文件另行分級",
        "ida_version": ida_kernwin.get_kernel_version(),
        "address_space": "IDA linear address (DOS/4GW image)",
        "processor": ida_ida.inf_get_procname(),
        "input": {"path": source, "sha256": digest(source)},
        "database": {"path": database, "sha256": digest(database)},
        "roots": {name: function_record(ea) for name, ea in ROOTS.items()},
        "matching_strings": string_records(),
    }
    with open(os.environ["MOO2_IDA_OUTPUT"], "w", encoding="utf-8") as target:
        json.dump(output, target, ensure_ascii=False, indent=2)
    ida_pro.qexit(0)


if __name__ == "__main__":
    main()
