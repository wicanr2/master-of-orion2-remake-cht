"""唯讀匯出隨機事件 19..23 的建立、維護、新聞與怪獸航行／戰鬥候選鏈。"""

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

OUT = os.environ.get("MOO2_RE_OUT", "/out/random-event-monsters.json")
SOURCE = os.environ.get("MOO2_RE_SOURCE", ida_nalt.get_input_file_path())
ROOTS = {
    "raw_Event_Effect_Consumer": 0x206A2,
    "raw_Event_Dispatcher": 0x21371,
    "raw_Determine_Event": 0x2230A,
    "raw_Event_Record_Maintenance": 0x20316,
    "raw_Monster_Target_Candidate_A": 0x23BEC,
    "raw_Monster_Target_Candidate_B": 0x23DA0,
    "raw_Monster_Target_Candidate_C": 0x23D44,
    "raw_Event_Star_Conflict_Filter": 0x242FC,
    "raw_Spawn_Event_Monster_Fleet": 0xA16BF,
    "raw_Event_Monster_Fleet_Helper": 0xA1A23,
    "raw_Move_All_Ships": 0xFFEEA,
    "raw_Search_For_Battles": 0xE9D62,
}
NEEDLES = (
    "19ABA4", "19ABA5", "19ABA7", "19ABA9", "19ABAB",
    "monster", "Monster", "eel", "Eel", "amoeba", "Amoeba",
    "crystal", "Crystal", "dragon", "Dragon", "hydra", "Hydra",
)


def sha256(path):
    h = hashlib.sha256()
    with open(path, "rb") as f:
        for chunk in iter(lambda: f.read(1024 * 1024), b""):
            h.update(chunk)
    return h.hexdigest()


def insn(ea):
    size = idc.get_item_size(ea)
    return {"ea": f"0x{ea:X}", "bytes": (ida_bytes.get_bytes(ea, size) or b"").hex(),
            "mnem": idc.print_insn_mnem(ea), "op0": idc.print_operand(ea, 0),
            "op1": idc.print_operand(ea, 1), "line": idc.generate_disasm_line(ea, 0) or "<unavailable>"}


def function(ea):
    f = ida_funcs.get_func(ea)
    if not f:
        return {"requested": f"0x{ea:X}", "error": "no function"}
    items = list(idautils.FuncItems(f.start_ea))
    return {"requested": f"0x{ea:X}", "start": f"0x{f.start_ea:X}", "end": f"0x{f.end_ea:X}",
            "original_name": idc.get_name(f.start_ea) or "<unnamed>",
            "instructions": [insn(x) for x in items],
            "callers": [insn(x) for x in idautils.CodeRefsTo(f.start_ea, 0)],
            "calls": [insn(x) for x in items if idc.print_insn_mnem(x).lower() == "call"]}


def interesting_functions():
    out = {}
    for ea in idautils.Functions():
        name = idc.get_name(ea) or ""
        if any(term.lower() in name.lower() for term in NEEDLES[5:]):
            out[f"0x{ea:X}"] = function(ea)
            continue
        hits = []
        for x in idautils.FuncItems(ea):
            line = idc.generate_disasm_line(x, 0) or ""
            if any(term in line for term in NEEDLES):
                hits.append(insn(x))
        if hits:
            out[f"0x{ea:X}"] = {"start": f"0x{ea:X}", "original_name": name or "<unnamed>", "hits": hits}
    return out


def main():
    ida_auto.auto_wait()
    report = {
        "schema": "moo2.ida.re-evidence.v1", "evidence_scope": "static_only", "mutation": "none",
        "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version(),
                 "script": "tools/ida/audit_random_event_monsters.py"},
        "input": {"database": ida_nalt.get_input_file_path(), "source": SOURCE,
                  "source_sha256": sha256(SOURCE), "processor": ida_ida.inf_get_procname(),
                  "min_ea": f"0x{ida_ida.inf_get_min_ea():X}", "max_ea": f"0x{ida_ida.inf_get_max_ea():X}"},
        "address_basis": "IDA linear; DOS/4GW LE object #1",
        "roots": {name: function(ea) for name, ea in ROOTS.items()},
        "interesting_functions": interesting_functions(),
    }
    os.makedirs(os.path.dirname(OUT), exist_ok=True)
    with open(OUT, "w", encoding="utf-8") as f:
        json.dump(report, f, ensure_ascii=False, indent=2)
        f.write("\n")
    ida_pro.qexit(0)


if __name__ == "__main__":
    main()
