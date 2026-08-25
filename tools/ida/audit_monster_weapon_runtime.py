"""唯讀搜尋 combat weapon record +0x56 raw flags 的讀取端與鄰近射擊函式。"""

import json
import os

import ida_auto
import ida_bytes
import ida_funcs
import ida_ida
import ida_kernwin
import ida_lines
import ida_nalt
import ida_pro
import idautils
import idc

import audit_event_warp_beast as common

OUT = os.environ.get("MOO2_RE_OUT", "/out/monster-weapon-runtime.json")
SOURCE = os.environ.get("MOO2_RE_SOURCE", ida_nalt.get_input_file_path())
ROOTS = {
    "raw_Load_Combat_Ship": 0x4954A,
    "raw_Fire_Fighter_Bomb": 0x3AC20,
    "raw_Fire_Fighter_Beam": 0x3AD57,
    "raw_Do_Combat_Turn": 0x42F7F,
    "raw_Do_One_Combat": 0xE938C,
    "raw_Weapon_Usable_Check": 0x389E8,
    "raw_Weapon_Value_Helper": 0x39434,
    "raw_Weapon_Target_Check": 0x3A7AA,
    "raw_Weapon_Damage_Value": 0x2B9E3,
    "raw_Tactical_Weapon_Dispatch": 0x39F1D,
    "raw_Strategic_Weapon_Attack": 0x41F80,
    "raw_Strategic_Special_Attack": 0x420C0,
    "raw_Strategic_Missile_Attack": 0x4221F,
    "raw_Caustic_Slime_Apply": 0xACF83,
    "raw_Special_Weapon_Visual": 0xADE18,
    "raw_Tactical_Status_Display": 0x37308,
    "raw_Apply_Damage_To_Facing": 0x39985,
    "raw_Caustic_Slime_Turn_Tick": 0x4A5CE,
}


def interesting_function(ea):
    f = ida_funcs.get_func(ea)
    if not f:
        return False
    lines = []
    has_field, has_stride = False, False
    for insn_ea in idautils.FuncItems(f.start_ea):
        line = ida_lines.tag_remove(idc.generate_disasm_line(insn_ea, 0) or "")
        low = line.lower()
        if "+56h" in low or "+ 56h" in low:
            has_field = True
        if "139h" in low or "0bh" in low:
            has_stride = True
        lines.append((insn_ea, low))
    return has_field and has_stride


def main():
    ida_auto.auto_wait()
    candidates = {}
    mask_4000 = {}
    monster_id_candidates = {}
    slime_state_candidates = {}
    high_byte_40_candidates = {}
    for ea in idautils.Functions():
        if interesting_function(ea):
            candidates[f"0x{ea:X}"] = common.function(ea)
        f = ida_funcs.get_func(ea)
        if f and any("4000h" in (ida_lines.tag_remove(idc.generate_disasm_line(x, 0) or "")).lower()
                     for x in idautils.FuncItems(f.start_ea)):
            mask_4000[f"0x{ea:X}"] = common.function(ea)
        if f:
            lines = [(ida_lines.tag_remove(idc.generate_disasm_line(x, 0) or "")).lower()
                     for x in idautils.FuncItems(f.start_ea)]
            has_mask_40 = any((line.startswith("test") or line.startswith("and")) and "40h" in line
                              for line in lines)
            has_weapon_shape = any(token in line for line in lines
                                   for token in ("139h", "1ah", "+56h", "+11h", "+12h"))
            if has_mask_40 and has_weapon_shape:
                high_byte_40_candidates[f"0x{ea:X}"] = common.function(ea)
            if any("+43h" in line or "+ 43h" in line for line in lines):
                slime_state_candidates[f"0x{ea:X}"] = common.function(ea)
            if any("dword_192864" in line for line in lines) and any(
                    token in line for line in lines for token in ("2ch", "2dh", "28h")):
                monster_id_candidates[f"0x{ea:X}"] = common.function(ea)
    report = {
        "schema": "moo2.ida.re-evidence.v1", "evidence_scope": "static_only", "mutation": "none",
        "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version(),
                 "script": "tools/ida/audit_monster_weapon_runtime.py"},
        "input": {"database": ida_nalt.get_input_file_path(), "source": SOURCE,
                  "source_sha256": common.sha256(SOURCE), "processor": ida_ida.inf_get_procname(),
                  "min_ea": f"0x{ida_ida.inf_get_min_ea():X}", "max_ea": f"0x{ida_ida.inf_get_max_ea():X}"},
        "address_basis": "IDA linear; DOS/4GW LE object #1",
        "roots": {name: common.function(ea) for name, ea in ROOTS.items()},
        "combat_weapon_flag_candidates": candidates,
        "constant_4000_functions": mask_4000,
        "monster_weapon_id_candidates": monster_id_candidates,
        "slime_state_candidates": slime_state_candidates,
        "high_byte_40_candidates": high_byte_40_candidates,
        "ui_flag_strings": {
            "0x17FD1E": (ida_bytes.get_bytes(0x17FD1E, 32) or b"").hex(),
            "0x17FD2D": (ida_bytes.get_bytes(0x17FD2D, 32) or b"").hex(),
        },
        "raw_tables": {
            "weapon_mod_records_0x17FD15": (ida_bytes.get_bytes(0x17FD15, 15 * 15) or b"").hex(),
            "owner_ordnance_percent_0x199224": [ida_bytes.get_word(0x199224 + i * 2) for i in range(16)],
        },
    }
    os.makedirs(os.path.dirname(OUT), exist_ok=True)
    with open(OUT, "w", encoding="utf-8") as f:
        json.dump(report, f, ensure_ascii=False, indent=2)
        f.write("\n")
    ida_pro.qexit(0)


if __name__ == "__main__":
    main()
