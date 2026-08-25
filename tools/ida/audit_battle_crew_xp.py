"""非破壞性匯出原版戰後艦員 XP 的最小充分證據。

不改名、不套型別、不寫回 IDB。每筆保留 raw name、IDA linear address、bytes、
原始運算元與交叉參照；附加語意僅作導覽，信賴等級由研究文件審查後決定。
"""

import hashlib
import json
import os
import traceback

import ida_bytes
import ida_funcs
import ida_kernwin
import ida_lines
import ida_name
import ida_nalt
import ida_pro
import ida_ua
import idautils
import idc


FUNCTIONS = (0x47939, 0x4B184)
RANGES = (
    (0x48D43, 0x48D57, "sub_47939 傳入戰後結果與兩側資料"),
    (0x4B19B, 0x4B1CA, "依戰後結果選出 recipient side raw id"),
    (0x4B2D8, 0x4B32C, "摧毀記錄按 owner 累加 raw hull class + 1"),
    (0x4B810, 0x4B953, "逐戰鬥記錄篩 recipient、除二、最少一點並寫 Ship+0x72"),
)


def raw_name(ea):
    func = ida_funcs.get_func(ea)
    return ida_name.get_name(func.start_ea) if func else ida_name.get_name(ea)


def digest(path):
    if path and os.path.isfile(path):
        with open(path, "rb") as fh:
            return hashlib.sha256(fh.read()).hexdigest()
    return "database-input-not-present; see project RE note"


def instruction(ea):
    insn = ida_ua.insn_t()
    size = ida_ua.decode_insn(insn, ea)
    if size <= 0:
        size = 1
    return {
        "ea": f"0x{ea:X}",
        "bytes": (ida_bytes.get_bytes(ea, size) or b"").hex(" "),
        "text": ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or ""),
        "code_refs": [f"0x{x:X}" for x in idautils.CodeRefsFrom(ea, 0)],
        "data_refs": [f"0x{x:X}" for x in idautils.DataRefsFrom(ea)],
    }, size


def range_payload(start, end, navigation_label):
    rows = []
    ea = start
    while ea < end:
        row, size = instruction(ea)
        rows.append(row)
        ea += size
    return {
        "start": f"0x{start:X}",
        "end_exclusive": f"0x{end:X}",
        "navigation_label": navigation_label,
        "warning": "label is navigation only; raw instructions are evidence",
        "instructions": rows,
    }


def function_payload(ea):
    func = ida_funcs.get_func(ea)
    if not func:
        return {"requested": f"0x{ea:X}", "error": "function missing"}
    callers = []
    for xref in idautils.XrefsTo(func.start_ea, 0):
        if not xref.iscode:
            continue
        owner = ida_funcs.get_func(xref.frm)
        callers.append({
            "from": f"0x{xref.frm:X}",
            "owner_start": f"0x{owner.start_ea:X}" if owner else None,
            "owner_raw_name": raw_name(owner.start_ea) if owner else "outside-function",
            "xref_type": int(xref.type),
        })
    return {
        "requested": f"0x{ea:X}",
        "start": f"0x{func.start_ea:X}",
        "end": f"0x{func.end_ea:X}",
        "raw_name": raw_name(func.start_ea),
        "callers": callers,
    }


def main():
    path = ida_nalt.get_input_file_path()
    hash_path = os.environ.get("MOO2_IDA_INPUT", path)
    payload = {
        "contract": "raw-location + navigation-label + confidence-in-reviewed-note + source",
        "input": os.path.basename(path) if path else ida_nalt.get_root_filename(),
        "input_sha256": digest(hash_path),
        "ida_sdk_version": str(getattr(ida_pro, "IDA_SDK_VERSION", "9.4")),
        "address_space": "IDA linear",
        "functions": [function_payload(ea) for ea in FUNCTIONS],
        "ranges": [range_payload(*item) for item in RANGES],
    }
    with open(os.environ["MOO2_IDA_OUTPUT"], "w", encoding="utf-8") as fh:
        json.dump(payload, fh, ensure_ascii=False, indent=2)
        fh.write("\n")


try:
    main()
except Exception:
    error = traceback.format_exc()
    output = os.environ.get("MOO2_IDA_OUTPUT", "/tmp/moo2-battle-crew-xp.json")
    with open(output + ".error", "w", encoding="utf-8") as fh:
        fh.write(error)
    ida_kernwin.msg(error + "\n")
    ida_pro.qexit(1)
else:
    ida_pro.qexit(0)
