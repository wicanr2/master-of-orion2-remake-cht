"""非破壞性匯出 Get_Colony_Hits_ 的指令、caller、callee 與資料參照。

保留 IDA linear address、原始名稱、bytes 與 operand；不改名、不套型別、不寫資料庫。
輸出路徑由 MOO2_IDA_OUTPUT 指定。
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


TARGETS = (0x42371, 0x4257E, 0x4267B)


def owner(ea):
    func = ida_funcs.get_func(ea)
    if not func:
        return {"start": None, "raw_name": "outside-function"}
    return {"start": f"0x{func.start_ea:X}", "raw_name": ida_name.get_name(func.start_ea)}


def refs_to(ea):
    return [
        {"from": f"0x{x.frm:X}", "type": int(x.type), "is_code": bool(x.iscode), "owner": owner(x.frm)}
        for x in idautils.XrefsTo(ea, 0)
    ]


def instruction_rows(func):
    rows = []
    ea = func.start_ea
    while ea < func.end_ea:
        insn = ida_ua.insn_t()
        size = ida_ua.decode_insn(insn, ea)
        if size <= 0:
            size = 1
        rows.append(
            {
                "ea": f"0x{ea:X}",
                "bytes": (ida_bytes.get_bytes(ea, size) or b"").hex(" "),
                "text": ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or ""),
                "code_refs": [f"0x{x:X}" for x in idautils.CodeRefsFrom(ea, 0)],
                "data_refs": [f"0x{x:X}" for x in idautils.DataRefsFrom(ea)],
            }
        )
        ea += size
    return rows


def digest(path):
    if path and os.path.isfile(path):
        with open(path, "rb") as fh:
            return hashlib.sha256(fh.read()).hexdigest()
    return "database-input-not-present; see project RE note"


def main():
    functions = []
    for ea in TARGETS:
        func = ida_funcs.get_func(ea)
        if not func:
            functions.append({"requested": f"0x{ea:X}", "error": "function missing"})
            continue
        functions.append(
            {
                "requested": f"0x{ea:X}",
                "start": f"0x{func.start_ea:X}",
                "end": f"0x{func.end_ea:X}",
                "raw_name": ida_name.get_name(func.start_ea),
                "confidence": "unknown until instruction and data-flow review",
                "source": "ORION2.EXE IDA database",
                "refs_to": refs_to(func.start_ea),
                "instructions": instruction_rows(func),
            }
        )

    path = ida_nalt.get_input_file_path()
    input_name = os.path.basename(path) if path else ida_nalt.get_root_filename()
    payload = {
        "contract": "raw-location + semantic-label + confidence + source",
        "input": input_name,
        "input_sha256": digest(path),
        "ida_sdk_version": str(getattr(ida_pro, "IDA_SDK_VERSION", "9.4")),
        "address_space": "IDA linear",
        "functions": functions,
    }
    with open(os.environ["MOO2_IDA_OUTPUT"], "w", encoding="utf-8") as fh:
        json.dump(payload, fh, ensure_ascii=False, indent=2)
        fh.write("\n")


try:
    main()
except Exception:
    error = traceback.format_exc()
    output = os.environ.get("MOO2_IDA_OUTPUT", "/tmp/moo2-colony-hits.json")
    with open(output + ".error", "w", encoding="utf-8") as fh:
        fh.write(error)
    ida_kernwin.msg(error + "\n")
    ida_pro.qexit(1)
else:
    ida_pro.qexit(0)
