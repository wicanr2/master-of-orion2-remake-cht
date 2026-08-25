"""非破壞性匯出武器成本、佔格、微型化與改造表證據。"""

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


FUNCTIONS = (0x6A406, 0x6A636, 0x6B519, 0x6D048, 0x6E60E, 0x6E70A,
             0x6EC74, 0x6EDC6, 0x6EE8E, 0x6F11C, 0x8E94D)
RANGES = (
    (0x6A406, 0x6A636, "regular mod bonus raw flags and table reads"),
    (0x6A636, 0x6A6A0, "special mod bonus raw flags and table reads"),
    (0x6B519, 0x6B5A0, "cost miniaturization percentage ladder"),
    (0x6D048, 0x6D0D5, "completed-topic chain counter"),
    (0x6E60E, 0x6E6B8, "space miniaturization category ladders"),
    (0x6E70A, 0x6E774, "weapon technology level gate"),
    (0x6EC74, 0x6EDC6, "weapon cost consumer"),
    (0x6EDC6, 0x6EE8E, "one weapon space consumer"),
    (0x6EE8E, 0x6EFEB, "weapon stack space consumer"),
    (0x6F11C, 0x6F1F0, "miniaturization category selector"),
    (0x8E94D, 0x8E9B0, "raw flag to table index helper"),
)
DATA_RANGES = (
    (0x17FC80, 0x17FE20, "weapon modification records vicinity"),
    (0x17E060, 0x17E130, "technology record vicinity"),
)


def digest(path):
    if path and os.path.isfile(path):
        h = hashlib.sha256()
        with open(path, "rb") as fh:
            for chunk in iter(lambda: fh.read(1024 * 1024), b""):
                h.update(chunk)
        return h.hexdigest()
    return "database-input-not-present; see project RE note"


def raw_name(ea):
    func = ida_funcs.get_func(ea)
    return ida_name.get_name(func.start_ea) if func else ida_name.get_name(ea)


def instruction(ea):
    insn = ida_ua.insn_t()
    size = ida_ua.decode_insn(insn, ea)
    if size <= 0:
        size = 1
    return ({
        "ea": f"0x{ea:X}",
        "bytes": (ida_bytes.get_bytes(ea, size) or b"").hex(" "),
        "text": ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or ""),
        "code_refs": [f"0x{x:X}" for x in idautils.CodeRefsFrom(ea, 0)],
        "data_refs": [f"0x{x:X}" for x in idautils.DataRefsFrom(ea)],
    }, size)


def range_payload(start, end, label):
    rows = []
    ea = start
    while ea < end:
        row, size = instruction(ea)
        rows.append(row)
        ea += size
    return {"start": f"0x{start:X}", "end_exclusive": f"0x{end:X}",
            "navigation_label": label,
            "warning": "label is navigation only; raw instructions are evidence",
            "instructions": rows}


def function_payload(ea):
    func = ida_funcs.get_func(ea)
    if not func:
        return {"requested": f"0x{ea:X}", "error": "function missing"}
    return {"requested": f"0x{ea:X}", "start": f"0x{func.start_ea:X}",
            "end": f"0x{func.end_ea:X}", "raw_name": raw_name(func.start_ea),
            "callers": [f"0x{x.frm:X}" for x in idautils.XrefsTo(func.start_ea, 0)
                        if x.iscode]}


def data_payload(start, end, label):
    raw = ida_bytes.get_bytes(start, end - start) or b""
    return {"start": f"0x{start:X}", "end_exclusive": f"0x{end:X}",
            "navigation_label": label, "bytes": raw.hex(" ")}


def main():
    path = ida_nalt.get_input_file_path()
    payload = {
        "contract": "raw-location + navigation-label + reviewed confidence + source",
        "input": os.path.basename(path) if path else ida_nalt.get_root_filename(),
        "input_sha256": digest(os.environ.get("MOO2_IDA_INPUT", path)),
        "database_sha256": digest(os.environ.get("MOO2_IDA_DATABASE", "")),
        "ida_sdk_version": str(getattr(ida_pro, "IDA_SDK_VERSION", "9.4")),
        "address_space": "IDA linear",
        "functions": [function_payload(ea) for ea in FUNCTIONS],
        "ranges": [range_payload(*item) for item in RANGES],
        "data_ranges": [data_payload(*item) for item in DATA_RANGES],
    }
    with open(os.environ["MOO2_IDA_OUTPUT"], "w", encoding="utf-8") as fh:
        json.dump(payload, fh, ensure_ascii=False, indent=2)
        fh.write("\n")


try:
    main()
except Exception:
    error = traceback.format_exc()
    output = os.environ.get("MOO2_IDA_OUTPUT", "/tmp/moo2-weapon-cost-space.json")
    with open(output + ".error", "w", encoding="utf-8") as fh:
        fh.write(error)
    ida_kernwin.msg(error + "\n")
    ida_pro.qexit(1)
else:
    ida_pro.qexit(0)
