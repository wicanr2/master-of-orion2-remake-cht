"""唯讀匯出 owner 8 怪物抵達星系後的戰鬥建立與殖民地戰後回寫鏈。"""

import json
import os

import ida_auto
import ida_ida
import ida_kernwin
import ida_nalt
import ida_pro

import audit_event_warp_beast as common

OUT = os.environ.get("MOO2_RE_OUT", "/out/monster-colony-battle.json")
SOURCE = os.environ.get("MOO2_RE_SOURCE", ida_nalt.get_input_file_path())
ROOTS = {
    "raw_Search_For_Battles": 0xE9D62,
    "raw_Select_Owner8_Opponent_After_Side_Chosen": 0xE8029,
    "raw_Do_One_Combat": 0xE938C,
    "raw_Prepare_Combat_Record": 0xE6A0C,
    "raw_Run_Combat_Core": 0xE7343,
    "raw_Post_Combat_Result_A": 0xE6B44,
    "raw_Post_Combat_Result_B": 0xE6CAA,
    "raw_Owner8_Post_Combat": 0xE9358,
    "raw_Final_Combat_Cleanup": 0xE87D2,
    "raw_Build_Colony_Bombard_Result": 0x4267B,
    "raw_Strategic_Bombardment": 0x4257E,
    "raw_Consume_Colony_Damage": 0xDD2F2,
    "raw_Apply_Colony_Damage_Point": 0xDCEBD,
    "raw_Destroy_Colony_A": 0xE1CED,
    "raw_Destroy_Colony_B": 0xEC97C,
    "raw_Remove_Colony": 0xE2A70,
    "raw_Consume_Battle_Result": 0xE6AA2,
    "raw_Resolve_Monster_Special_Case": 0xE9CD3,
    "raw_Postprocess_Star_Battles": 0xE9927,
    "raw_Colony_Defense_Selector": 0xE7CDB,
    "raw_Normal_Battle_Sides": 0xE7DCA,
}


def main():
    ida_auto.auto_wait()
    report = {
        "schema": "moo2.ida.re-evidence.v1",
        "evidence_scope": "static_only",
        "mutation": "none",
        "tool": {
            "name": "IDA Pro",
            "version": ida_kernwin.get_kernel_version(),
            "script": "tools/ida/audit_monster_colony_battle.py",
        },
        "input": {
            "database": ida_nalt.get_input_file_path(),
            "source": SOURCE,
            "source_sha256": common.sha256(SOURCE),
            "processor": ida_ida.inf_get_procname(),
            "min_ea": f"0x{ida_ida.inf_get_min_ea():X}",
            "max_ea": f"0x{ida_ida.inf_get_max_ea():X}",
        },
        "address_basis": "IDA linear; DOS/4GW LE object #1",
        "roots": {name: common.function(ea) for name, ea in ROOTS.items()},
    }
    os.makedirs(os.path.dirname(OUT), exist_ok=True)
    with open(OUT, "w", encoding="utf-8") as f:
        json.dump(report, f, ensure_ascii=False, indent=2)
        f.write("\n")
    ida_pro.qexit(0)


if __name__ == "__main__":
    main()
