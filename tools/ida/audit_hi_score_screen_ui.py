"""非破壞性匯出 Hi Score／Hall of Fame 畫面鏈、符號衝突與原始指令。"""

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
    "symbols_Add_Hi_Score_Screen_Fields": 0x9D7EA,
    "symbols_Draw_Captured_Colonies_Score": 0x9DBDB,
    "symbols_Draw_Total_Score": 0x9DC76,
    "symbols_Draw_Elimination_Score": 0x9E17C,
    "symbols_Draw_Generic_Score": 0x9E207,
    "symbols_Draw_Hi_Score_Screen": 0x9E27A,
    "symbols_Draw_Time_Score": 0x9E68B,
    "symbols_Hi_Score_Screen": 0x9EA3B,
    "symbols_Load_Hi_Score_Screen": 0x9EB42,
    "symbols_Load_Hall_Of_Fame_Screen": 0x9EDE1,
    "symbols_Draw_Hall_Of_Fame_Screen": 0x9EE43,
    "symbols_Hall_Of_Fame_Screen": 0x9F286,
    "symbols_End_Of_Game_Hi_Score": 0x9F712,
    "func_names_Do_Calculate_Hi_Score": 0x9DB96,
    "func_names_Hi_Score_Screen": 0x9EB42,
    "func_names_Hall_Of_Fame_Screen": 0x9F447,
    "func_names_End_Of_Game_Hi_Score": 0x9F981,
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
        "code_refs": [f"0x{x:X}" for x in idautils.CodeRefsFrom(ea, 0)],
        "data_refs": [f"0x{x:X}" for x in idautils.DataRefsFrom(ea)],
    }


def function_record(ea):
    function = ida_funcs.get_func(ea)
    if function is None:
        return {"requested": f"0x{ea:X}", "error": "function missing"}
    callers = []
    for ref in idautils.CodeRefsTo(function.start_ea, 0):
        owner = ida_funcs.get_func(ref)
        callers.append({
            "site": instruction(ref),
            "function_start": f"0x{owner.start_ea:X}" if owner else None,
            "raw_name": ida_name.get_name(owner.start_ea) if owner else None,
        })
    return {
        "requested": f"0x{ea:X}",
        "start": f"0x{function.start_ea:X}",
        "end_exclusive": f"0x{function.end_ea:X}",
        "raw_name": ida_name.get_name(function.start_ea) or "<unnamed>",
        "callers": callers,
        "instructions": [instruction(x) for x in idautils.FuncItems(function.start_ea)],
    }


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
    }
    with open(os.environ["MOO2_IDA_OUTPUT"], "w", encoding="utf-8") as target:
        json.dump(output, target, ensure_ascii=False, indent=2)
    ida_pro.qexit(0)


if __name__ == "__main__":
    main()
