"""非破壞性匯出地面戰畫面、兵力面板、資產 loader 與輸入鏈。"""

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
    "load_colony_combat_assets": 0xB6D51,
    "colony_combat_loop": 0xB7289,
    "colony_combat_setup": 0xB7491,
    "colony_combat_screen_candidate": 0xB771D,
    "ground_unit_animation_step": 0xB7825,
    "ground_unit_record_reset": 0xB86B4,
    "ground_unit_record_allocate": 0xB86F6,
    "place_ground_unit": 0xB88B2,
    "print_troop_totals": 0xB896D,
    "draw_attacker_panel": 0xB8BC7,
    "draw_defender_panel": 0xB8C8B,
    "replace_colgcbt_player_colors": 0xB8EFB,
    "compute_player_ground_combat_bonuses": 0xEC15C,
    "compute_ground_combat_info": 0xEC3CE,
    "ground_combat_round": 0xEC4FE,
    "resolve_ground_combat": 0xEC601,
}

DATA_RANGES = {
    "ground_color_map_c0_8x8": (0x1828D3, 8 * 8),
    "ground_color_map_e8_8x4": (0x182913, 8 * 8),
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
        "data_ranges": {
            name: {"start": f"0x{ea:X}", "size": size,
                   "bytes": (ida_bytes.get_bytes(ea, size) or b"").hex()}
            for name, (ea, size) in DATA_RANGES.items()
        },
    }
    with open(os.environ["MOO2_IDA_OUTPUT"], "w", encoding="utf-8") as target:
        json.dump(output, target, ensure_ascii=False, indent=2)
    ida_pro.qexit(0)


if __name__ == "__main__":
    main()
