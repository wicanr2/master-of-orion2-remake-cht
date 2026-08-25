"""唯讀匯出事件怪獸建立、路徑規劃與航程計算的直接資料流。"""

import json
import os

import ida_auto
import ida_ida
import ida_kernwin
import ida_nalt
import ida_pro

import audit_event_warp_beast as common

OUT = os.environ.get("MOO2_RE_OUT", "/out/event-monster-route.json")
SOURCE = os.environ.get("MOO2_RE_SOURCE", ida_nalt.get_input_file_path())
ROOTS = {
    "raw_Spawn_Event_Monster": 0xA16BF,
    "raw_Create_Monster_Ship": 0xA1762,
    "raw_Load_Amoeba_Design": 0x57C0B,
    "raw_Load_Crystal_Design": 0x57E1B,
    "raw_Load_Dragon_Design": 0x57A02,
    "raw_Load_Eel_Design": 0x57D14,
    "raw_Load_Hydra_Design": 0x57B1C,
    "raw_Route_Monster_Ship": 0xA1A23,
    "raw_Check_Route_To_Star": 0xFF799,
    "raw_Commit_Route_To_Star": 0xFFD08,
    "raw_Compute_Travel_State": 0xEBE79,
    "raw_Compute_Target_Distance_Class": 0xEBEB7,
    "raw_Finalize_Route_Group": 0xA1C0D,
    "raw_Create_Ship_At_Coordinates": 0x100010,
    "raw_Move_Ship_Coordinates": 0xEBB0C,
    "raw_Complete_Ship_Arrival": 0xFFDDA,
    "raw_Move_All_Ships": 0xFFEEA,
    "raw_Monster_Turn_Consumer": 0xDB8D8,
    "raw_Monster_Target_Selector": 0xDB6D2,
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
            "script": "tools/ida/audit_event_monster_route.py",
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
