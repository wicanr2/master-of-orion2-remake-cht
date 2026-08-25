"""非破壞性匯出登艦戰、艦員 Bo 加成與上下游原始證據。"""

import hashlib
import json
import os
import traceback

import ida_auto
import ida_bytes
import ida_funcs
import ida_ida
import ida_kernwin
import ida_lines
import ida_name
import ida_pro
import ida_ua
import idautils
import idc


TARGETS = {
    "resolve_ai_boarding": 0x2BF73,
    "boarding_action_type": 0x2C129,
    "crew_boarding_combat_bonus": 0x35CAD,
    "boarding_empire_bonus": 0x35FDA,
    "boarding_empire_leader_skill": 0x36036,
    "boarding_ship_bonus": 0x360AA,
    "boarding_ship_leader_skill": 0x36106,
    "init_boarding_action": 0x372BD,
    "capture_ship": 0x38312,
    "initiate_boarding_action": 0x459E7,
    "get_boarding_info": 0xEC767,
}


def digest(path):
    h = hashlib.sha256()
    with open(path, "rb") as fh:
        for chunk in iter(lambda: fh.read(1024 * 1024), b""):
            h.update(chunk)
    return h.hexdigest()


def instruction(ea):
    insn = ida_ua.insn_t()
    size = ida_ua.decode_insn(insn, ea)
    if size <= 0:
        size = max(1, idc.get_item_size(ea))
    return {
        "ea": f"0x{ea:X}",
        "bytes": (ida_bytes.get_bytes(ea, size) or b"").hex(" "),
        "text": ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or ""),
        "code_refs": [f"0x{x:X}" for x in idautils.CodeRefsFrom(ea, 0)],
        "data_refs": [f"0x{x:X}" for x in idautils.DataRefsFrom(ea)],
    }


def function_record(ea):
    fn = ida_funcs.get_func(ea)
    if fn is None:
        return {"requested": f"0x{ea:X}", "error": "function missing"}
    return {
        "requested": f"0x{ea:X}",
        "start": f"0x{fn.start_ea:X}",
        "end_exclusive": f"0x{fn.end_ea:X}",
        "raw_name": ida_name.get_name(fn.start_ea),
        "callers": [instruction(x.frm) for x in idautils.XrefsTo(fn.start_ea, 0) if x.iscode],
        "instructions": [instruction(x) for x in idautils.FuncItems(fn.start_ea)],
    }


def main():
    ida_auto.auto_wait()
    source = os.environ["MOO2_IDA_INPUT"]
    database = os.environ["MOO2_IDA_DATABASE"]
    payload = {
        "schema": "moo2.ida.re-evidence.v1",
        "contract": "raw-location + navigation-label + reviewed confidence + source",
        "mutation": "none",
        "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version()},
        "input": {
            "file": os.path.basename(source),
            "source_sha256": digest(source),
            "database_sha256": digest(database),
            "processor": ida_ida.inf_get_procname(),
        },
        "address_basis": "IDA linear; DOS/4GW LE image",
        "semantic_status": "unknown_pending_review",
        "targets": {name: function_record(ea) for name, ea in TARGETS.items()},
        "crew_bonus_table": [
            {
                "index": i,
                "ea": f"0x{0x17D174 + i * 8:X}",
                "bytes": (ida_bytes.get_bytes(0x17D174 + i * 8, 8) or b"").hex(" "),
                "word0": idc.get_wide_word(0x17D174 + i * 8),
            }
            for i in range(8)
        ],
    }
    with open(os.environ["MOO2_IDA_OUTPUT"], "w", encoding="utf-8") as fh:
        json.dump(payload, fh, ensure_ascii=False, indent=2)
        fh.write("\n")


try:
    main()
except Exception:
    error = traceback.format_exc()
    out = os.environ.get("MOO2_IDA_OUTPUT", "/tmp/boarding-combat.json")
    with open(out + ".error", "w", encoding="utf-8") as fh:
        fh.write(error)
    ida_kernwin.msg(error + "\n")
    ida_pro.qexit(1)
else:
    ida_pro.qexit(0)
