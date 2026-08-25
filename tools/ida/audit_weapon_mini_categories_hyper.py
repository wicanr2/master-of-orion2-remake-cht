"""非破壞性匯出武器微型化分類與 Hyper-Advanced 等級證據。"""

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


FUNCTIONS = (0x5E1E3, 0x654DD, 0x6B519, 0x6D048, 0x6E60E, 0x6E70A,
             0x6F11C, 0x78D3E, 0xE44E0, 0xFC734)
CALL_SITES = (0x60056, 0x618CE, 0x68345, 0x68DB3, 0x6E4E9,
              0x6EDEF, 0x6EE22, 0x6EE59, 0x6EFAB)
WORDS = (0x17FB71, 0x17FB55, 0x17FB1D, 0x17FB39,
         0x17EF44, 0x17F3DB, 0x17F0EB, 0x17F178, 0x17F582)


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
        size = 1
    return ({
        "ea": f"0x{ea:X}",
        "bytes": (ida_bytes.get_bytes(ea, size) or b"").hex(" "),
        "text": ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or ""),
        "code_refs": [f"0x{x:X}" for x in idautils.CodeRefsFrom(ea, 0)],
        "data_refs": [f"0x{x:X}" for x in idautils.DataRefsFrom(ea)],
    }, size)


def range_rows(start, end):
    rows = []
    ea = start
    while ea < end:
        row, size = instruction(ea)
        rows.append(row)
        ea += size
    return rows


def function_payload(ea):
    func = ida_funcs.get_func(ea)
    if not func:
        return {"requested": f"0x{ea:X}", "error": "function missing"}
    return {
        "requested": f"0x{ea:X}",
        "start": f"0x{func.start_ea:X}",
        "end_exclusive": f"0x{func.end_ea:X}",
        "raw_name": ida_name.get_name(func.start_ea),
        "callers": [f"0x{x.frm:X}" for x in idautils.XrefsTo(func.start_ea, 0) if x.iscode],
        "instructions": range_rows(func.start_ea, func.end_ea),
    }


def call_site_payload(ea):
    func = ida_funcs.get_func(ea)
    return {
        "ea": f"0x{ea:X}",
        "containing_function": f"0x{func.start_ea:X}" if func else None,
        "raw_name": ida_name.get_name(func.start_ea) if func else None,
        "context": range_rows(max(func.start_ea if func else ea, ea - 24), ea + 24),
    }


def xrefs_to_payload(ea):
    rows = []
    for xref in idautils.XrefsTo(ea, 0):
        func = ida_funcs.get_func(xref.frm)
        row, _ = instruction(xref.frm)
        rows.append({
            "from": f"0x{xref.frm:X}",
            "xref_type": xref.type,
            "containing_function": f"0x{func.start_ea:X}" if func else None,
            "raw_name": ida_name.get_name(func.start_ea) if func else None,
            "instruction": row,
        })
    return {"ea": f"0x{ea:X}", "raw_name": ida_name.get_name(ea), "xrefs": rows}


def main():
    input_path = os.environ["MOO2_IDA_INPUT"]
    db_path = os.environ["MOO2_IDA_DATABASE"]
    payload = {
        "contract": "raw-location + navigation-label + reviewed confidence + source",
        "input": os.path.basename(ida_nalt.get_input_file_path()),
        "input_sha256": digest(input_path),
        "database_sha256": digest(db_path),
        "ida_sdk_version": str(getattr(ida_pro, "IDA_SDK_VERSION", "9.4")),
        "address_space": "IDA linear",
        "functions": [function_payload(ea) for ea in FUNCTIONS],
        "call_sites": [call_site_payload(ea) for ea in CALL_SITES],
        "word_constants": [{
            "ea": f"0x{ea:X}",
            "raw_name": ida_name.get_name(ea),
            "value_u16": ida_bytes.get_word(ea),
            "bytes": (ida_bytes.get_bytes(ea, 2) or b"").hex(" "),
        } for ea in WORDS],
        "space_jump_tables": [{
            "ea": f"0x{ea:X}",
            "targets": [f"0x{ida_bytes.get_dword(ea + i * 4):X}" for i in range(5)],
            "bytes": (ida_bytes.get_bytes(ea, 20) or b"").hex(" "),
        } for ea in (0x6E5E6, 0x6E5FA)],
        "field_xrefs": [xrefs_to_payload(ea) for ea in (0x17E085, 0x17F80F)],
        "technology_record_bytes": (ida_bytes.get_bytes(0x17E07F, 212 * 13) or b"").hex(" "),
        "weapon_record_bytes": (ida_bytes.get_bytes(0x17F800, 46 * 28) or b"").hex(" "),
    }
    with open(os.environ["MOO2_IDA_OUTPUT"], "w", encoding="utf-8") as fh:
        json.dump(payload, fh, ensure_ascii=False, indent=2)
        fh.write("\n")


try:
    main()
except Exception:
    error = traceback.format_exc()
    output = os.environ.get("MOO2_IDA_OUTPUT", "/tmp/moo2-weapon-mini-categories-hyper.json")
    with open(output + ".error", "w", encoding="utf-8") as fh:
        fh.write(error)
    ida_kernwin.msg(error + "\n")
    ida_pro.qexit(1)
else:
    ida_pro.qexit(0)
