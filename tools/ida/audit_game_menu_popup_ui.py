"""非破壞性匯出對局內 Game Popup 的入口、繪製、資產與音量鏈。"""

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
    "do_main_game_popup_candidate": 0x7DD41,
    "do_options_game_popup_candidate": 0x7E00F,
    "load_game_popup_pictures_candidate": 0x7EA5C,
    "print_main_game_to_bitmap_candidate": 0x7ED66,
    "print_options_to_bitmap_candidate": 0x7EDB1,
    "draw_main_game_popup_candidate": 0x7F701,
    "draw_options_game_popup_candidate": 0x7FA28,
    "game_popup_candidate": 0x8012F,
    "set_music_for_game_popup_candidate": 0x80892,
    "set_sound_for_game_popup_candidate": 0x80918,
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
        "data_refs_to_start": [f"0x{x:X}" for x in idautils.DataRefsTo(function.start_ea)],
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
