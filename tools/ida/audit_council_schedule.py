"""非破壞性匯出銀河議會排程的原始指令與資料交叉參照。

用法（IDA Pro 9.4 batch）：
    idat64 -A -S/tools/ida/audit_council_schedule.py Orion2.i64

輸出固定使用 raw linear address；語意只寫在輸出標籤，不改名、不改型別、不寫資料庫。
"""

import hashlib
import json
import os
import traceback

import ida_bytes
import ida_funcs
import ida_ida
import ida_idaapi
import ida_kernwin
import ida_lines
import ida_name
import ida_nalt
import ida_pro
import ida_ua
import idautils
import idc


TARGET = 0x168AF
COUNCIL_MEETING = 0x15239
GLOBALS = (0x192FD8, 0x19306C, 0x197F98, 0x199998, 0x19999A, 0x19A0E2, 0x19A0E4)


def instruction_rows(start, end):
    rows = []
    ea = start
    while ea < end:
        size = ida_ua.decode_insn(ida_ua.insn_t(), ea)
        if size <= 0:
            size = 1
        rows.append(
            {
                "ea": f"0x{ea:X}",
                "bytes": ida_bytes.get_bytes(ea, size).hex(" "),
                "text": ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or ""),
            }
        )
        ea += size
    return rows


def xref_rows(ea):
    rows = []
    for x in idautils.XrefsTo(ea, 0):
        owner = ida_funcs.get_func(x.frm)
        rows.append(
            {
                "from": f"0x{x.frm:X}",
                "type": int(x.type),
                "function": f"0x{owner.start_ea:X}" if owner else "outside-function",
            }
        )
    return rows


def main():
    func = ida_funcs.get_func(TARGET)
    if func is None:
        raise RuntimeError(f"找不到函式 0x{TARGET:X}")
    input_path = ida_nalt.get_input_file_path()
    digest = ""
    if input_path and os.path.isfile(input_path):
        with open(input_path, "rb") as fh:
            digest = hashlib.sha256(fh.read()).hexdigest()
    payload = {
        "contract": "raw-location + semantic-label + confidence + source",
        "input": os.path.basename(input_path),
        "input_sha256": digest,
        "ida_version": str(getattr(ida_pro, "IDA_SDK_VERSION", "9.4")),
        "address_space": "IDA linear",
        "function": {
            "raw_name": ida_name.get_name(TARGET),
            "start": f"0x{func.start_ea:X}",
            "end": f"0x{func.end_ea:X}",
            "semantic": "銀河議會召開條件與下一屆日期候選",
            "confidence": "strong inference until instruction consumers are reviewed",
            "source": "ORION2.EXE IDA database",
            "instructions": instruction_rows(func.start_ea, func.end_ea),
            "callers": xref_rows(TARGET),
        },
        "meeting_function": {},
        "globals": [
            {
                "ea": f"0x{ea:X}",
                "raw_name": ida_name.get_name(ea),
                "semantic": "unknown schedule input/output candidate",
                "confidence": "unknown",
                "source": f"xrefs to IDA linear 0x{ea:X}",
                "xrefs": xref_rows(ea),
            }
            for ea in GLOBALS
        ],
    }
    meeting = ida_funcs.get_func(COUNCIL_MEETING)
    if meeting:
        payload["meeting_function"] = {
            "raw_name": ida_name.get_name(COUNCIL_MEETING),
            "start": f"0x{meeting.start_ea:X}",
            "end": f"0x{meeting.end_ea:X}",
            "semantic": "議會召開副作用候選；用於確認 word_19A0E2 寫入",
            "confidence": "strong inference until consumers are reviewed",
            "source": "ORION2.EXE IDA database",
            "instructions": instruction_rows(meeting.start_ea, meeting.end_ea),
            "callers": xref_rows(COUNCIL_MEETING),
        }

    output_path = os.environ.get("MOO2_IDA_OUTPUT")
    if output_path:
        with open(output_path, "w", encoding="utf-8") as fh:
            json.dump(payload, fh, ensure_ascii=False, indent=2)
            fh.write("\n")
    ida_kernwin.msg("MOO2_COUNCIL_SCHEDULE_JSON_BEGIN\n")
    ida_kernwin.msg(json.dumps(payload, ensure_ascii=False, indent=2) + "\n")
    ida_kernwin.msg("MOO2_COUNCIL_SCHEDULE_JSON_END\n")


try:
    main()
except Exception:
    error = traceback.format_exc()
    output_path = os.environ.get("MOO2_IDA_OUTPUT")
    if output_path:
        with open(output_path + ".error", "w", encoding="utf-8") as fh:
            fh.write(error)
    ida_kernwin.msg(error + "\n")
    ida_pro.qexit(1)
else:
    ida_pro.qexit(0)
