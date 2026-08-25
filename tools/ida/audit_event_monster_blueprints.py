"""唯讀匯出五種事件怪獸設計、引擎衍生與戰鬥船載入資料流。"""

import json
import os

import ida_auto
import ida_bytes
import ida_ida
import ida_kernwin
import ida_nalt
import ida_pro

import audit_event_warp_beast as common

OUT = os.environ.get("MOO2_RE_OUT", "/out/event-monster-blueprints.json")
SOURCE = os.environ.get("MOO2_RE_SOURCE", ida_nalt.get_input_file_path())
ROOTS = {
    "raw_Load_Dragon_Design": 0x57A02,
    "raw_Load_Hydra_Design": 0x57B1C,
    "raw_Load_Amoeba_Design": 0x57C0B,
    "raw_Load_Eel_Design": 0x57D14,
    "raw_Load_Crystal_Design": 0x57E1B,
    "raw_Drive_From_Raw_Type": 0x56726,
    "raw_Drive_Table_Helper_A": 0x5680D,
    "raw_Drive_Table_Helper_B": 0x5685F,
    "raw_Design_Derived_Value": 0x575D6,
    "raw_Load_Combat_Ship": 0x4954A,
    "raw_Max_Armor": 0x58387,
    "raw_Max_Structure": 0x58425,
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
            "script": "tools/ida/audit_event_monster_blueprints.py",
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
        "raw_tables": {
            "drive_size_records": {
                "base": "0x17FE90", "stride": 46,
                "u8": [ida_bytes.get_byte(0x17FE90 + i) for i in range(46 * 7)],
            },
            "hull_records": {
                "base": "0x180020", "stride": 36,
                "bytes": (ida_bytes.get_bytes(0x180020, 36 * 9) or b"").hex(),
            },
        },
    }
    os.makedirs(os.path.dirname(OUT), exist_ok=True)
    with open(OUT, "w", encoding="utf-8") as f:
        json.dump(report, f, ensure_ascii=False, indent=2)
        f.write("\n")
    ida_pro.qexit(0)


if __name__ == "__main__":
    main()
