"""非破壞性匯出飛彈彈架容量、成本／佔格與 caller 證據。"""

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


FUNCTIONS = (0x6EC74, 0x6EDC6, 0x6EE8E, 0x6E69E, 0x6F1CC)
RANGES = (
    (0x6E69E, 0x6E70A, "ammo or miniaturization gate helper"),
    (0x6EC74, 0x6EDAC, "Weapon_Cost raw function"),
    (0x6EDC6, 0x6EE8E, "One_Weapon_Space raw function"),
    (0x6EE8E, 0x6EFE9, "Weapon_Space raw function"),
    (0x6F1CC, 0x6F230, "ammo value mapping helper"),
)


def digest(path):
    if not path or not os.path.isfile(path):
        return "database-input-not-present; see project RE note"
    h = hashlib.sha256()
    with open(path, "rb") as fh:
        for chunk in iter(lambda: fh.read(1024 * 1024), b""):
            h.update(chunk)
    return h.hexdigest()


def insn(ea):
    decoded = ida_ua.insn_t()
    size = ida_ua.decode_insn(decoded, ea)
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
    rows, ea = [], start
    while ea < end:
        row, size = insn(ea)
        rows.append(row)
        ea += size
    return {"start": f"0x{start:X}", "end_exclusive": f"0x{end:X}",
            "navigation_label": label,
            "warning": "label is navigation only; raw instructions are evidence",
            "instructions": rows}


def caller_window(xref_ea):
    owner = ida_funcs.get_func(xref_ea)
    start = max(owner.start_ea if owner else xref_ea - 48, xref_ea - 64)
    end = min(owner.end_ea if owner else xref_ea + 48, xref_ea + 64)
    return range_payload(start, end, "caller window around raw xref")


def function_payload(ea):
    f = ida_funcs.get_func(ea)
    if not f:
        return {"requested": f"0x{ea:X}", "error": "function missing"}
    callers = []
    for x in idautils.XrefsTo(f.start_ea, 0):
        if not x.iscode:
            continue
        owner = ida_funcs.get_func(x.frm)
        callers.append({
            "xref": f"0x{x.frm:X}",
            "owner_start": f"0x{owner.start_ea:X}" if owner else None,
            "owner_raw_name": ida_name.get_name(owner.start_ea) if owner else "outside-function",
            "window": caller_window(x.frm),
        })
    return {"requested": f"0x{ea:X}", "start": f"0x{f.start_ea:X}",
            "end": f"0x{f.end_ea:X}", "raw_name": ida_name.get_name(f.start_ea),
            "callers": callers}


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
    }
    with open(os.environ["MOO2_IDA_OUTPUT"], "w", encoding="utf-8") as fh:
        json.dump(payload, fh, ensure_ascii=False, indent=2)
        fh.write("\n")


try:
    main()
except Exception:
    error = traceback.format_exc()
    output = os.environ.get("MOO2_IDA_OUTPUT", "/tmp/moo2-missile-ammo-racks.json")
    with open(output + ".error", "w", encoding="utf-8") as fh:
        fh.write(error)
    ida_kernwin.msg(error + "\n")
    ida_pro.qexit(1)
else:
    ida_pro.qexit(0)
