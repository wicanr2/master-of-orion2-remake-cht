"""非破壞性匯出殖民地轟炸畫面、資產 loader、炸彈佇列與輸入生命週期。"""

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
    "bombing_cleanup": 0xB432D,
    "load_bombing_anims": 0xB435A,
    "bombing_asset_setup_neighbor": 0xB4379,
    "bombing_queue_reset": 0xB4383,
    "add_bomb_to_queue": 0xB43A6,
    "bomb_queue_seed": 0xB44EB,
    "bomb_sprite_init": 0xB4523,
    "bomb_sprite_draw": 0xB45D4,
    "do_bomb": 0xB4606,
    "bomb_tick_driver": 0xB46CF,
    "bombing_animation_setup": 0xB4733,
    "draw_colony_bombing_screen": 0xB4800,
    "bombing_screen_driver": 0xB48EB,
    "bombing_screen_setup": 0xB494B,
    "bombing_screen_finish": 0xB4B51,
    "colony_bombing_screen_symbol_candidate": 0xB4D02,
    "replace_colgcbt_player_colors": 0xB8EFB,
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
        "nearby_functions": [
            {"start": f"0x{x:X}", "raw_name": ida_name.get_name(x) or "<unnamed>"}
            for x in idautils.Functions(0xB4300, 0xB4D50)
        ],
    }
    with open(os.environ["MOO2_IDA_OUTPUT"], "w", encoding="utf-8") as target:
        json.dump(output, target, ensure_ascii=False, indent=2)
    ida_pro.qexit(0)


if __name__ == "__main__":
    main()
