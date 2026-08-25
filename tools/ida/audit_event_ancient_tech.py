"""唯讀匯出事件 0 的科技選擇 helper、caller 與 application 授予鏈。"""

import hashlib
import json
import os

import ida_auto
import ida_bytes
import ida_funcs
import ida_ida
import ida_kernwin
import ida_nalt
import ida_pro
import idautils
import idc

OUT = os.environ.get("MOO2_RE_OUT", "/out/event-ancient-tech.json")
SOURCE = os.environ.get("MOO2_RE_SOURCE", ida_nalt.get_input_file_path())
ROOTS = {
    "raw_Event_Effect_Consumer": 0x206A2,
    "raw_Determine_Event": 0x2230A,
    "raw_Event_Tech_Chooser": 0x58853,
    "raw_Player_Tech_Benchmark": 0x5679E,
    "raw_Player_Weapon_Benchmark": 0x568EB,
    "raw_Tech_Eligibility": 0xE412B,
    "raw_Grant_Application": 0xE4204,
}


def sha256(path):
    h = hashlib.sha256()
    with open(path, "rb") as f:
        for chunk in iter(lambda: f.read(1024 * 1024), b""):
            h.update(chunk)
    return h.hexdigest()


def instruction(ea):
    size = idc.get_item_size(ea)
    return {
        "ea": f"0x{ea:X}",
        "bytes": (ida_bytes.get_bytes(ea, size) or b"").hex(),
        "mnem": idc.print_insn_mnem(ea),
        "op0": idc.print_operand(ea, 0),
        "op1": idc.print_operand(ea, 1),
        "line": idc.generate_disasm_line(ea, 0) or "<unavailable>",
    }


def function(ea):
    f = ida_funcs.get_func(ea)
    if not f:
        return {"requested": f"0x{ea:X}", "error": "no function"}
    return {
        "requested": f"0x{ea:X}",
        "start": f"0x{f.start_ea:X}",
        "end": f"0x{f.end_ea:X}",
        "original_name": idc.get_name(f.start_ea) or "<unnamed>",
        "instructions": [instruction(x) for x in idautils.FuncItems(f.start_ea)],
        "callers": [instruction(x) for x in idautils.CodeRefsTo(f.start_ea, 0)],
        "callees": [instruction(x) for x in idautils.FuncItems(f.start_ea)
                    if idc.print_insn_mnem(x).lower() == "call"],
    }


def main():
    ida_auto.auto_wait()
    report = {
        "schema": "moo2.ida.re-evidence.v1",
        "evidence_scope": "static_only",
        "mutation": "none",
        "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version(),
                 "script": "tools/ida/audit_event_ancient_tech.py"},
        "input": {"database": ida_nalt.get_input_file_path(), "source": SOURCE,
                  "source_sha256": sha256(SOURCE), "processor": ida_ida.inf_get_procname(),
                  "min_ea": f"0x{ida_ida.inf_get_min_ea():X}",
                  "max_ea": f"0x{ida_ida.inf_get_max_ea():X}"},
        "address_basis": "IDA linear; DOS/4GW LE object #1",
        "roots": {name: function(ea) for name, ea in ROOTS.items()},
        "raw_tables": {
            "record_3B_application_word_17F6A7": [
                {"index": i, "ea": f"0x{0x17F6A7 + i * 0x3B:X}",
                 "word": ida_bytes.get_word(0x17F6A7 + i * 0x3B)}
                for i in range(0, 8)
            ],
            "weapon_application_word_17F80D": [
                {"index": i, "ea": f"0x{0x17F80D + i * 0x1C:X}",
                 "word": ida_bytes.get_word(0x17F80D + i * 0x1C),
                 "category": ida_bytes.get_byte(0x17F80F + i * 0x1C),
                 "max_damage": ida_bytes.get_word(0x17F819 + i * 0x1C)}
                for i in range(0, 40)
            ],
        },
    }
    os.makedirs(os.path.dirname(OUT), exist_ok=True)
    with open(OUT, "w", encoding="utf-8") as f:
        json.dump(report, f, ensure_ascii=False, indent=2)
        f.write("\n")
    ida_pro.qexit(0)


if __name__ == "__main__":
    main()
