"""非破壞性匯出銀河議會投票狀態機的 raw 指令與 references。

保留 IDA linear address、raw name、bytes 與 operand；不改名、不套型別、不寫資料庫。
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


FUNCTIONS = {
    0x156D4: "議會票表／候選初始化候選",
    0x15BE9: "玩家投票互動候選",
    0x15DF8: "逐票後勝負門檻候選",
    0x15EBC: "全體帝國投票候選",
    0x16021: "Vote_Check_ 候選",
    0x161E4: "單一帝國票數回寫候選",
    0x1633C: "選舉結果／2/3 判定候選",
    0x1660B: "議會結果收尾候選",
}


def owner_name(ea):
    owner = ida_funcs.get_func(ea)
    if not owner:
        return "outside-function"
    return f"0x{owner.start_ea:X}:{ida_name.get_name(owner.start_ea)}"


def instruction_rows(start, end):
    rows = []
    ea = start
    while ea < end:
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


def callers(ea):
    return [
        {"from": f"0x{x.frm:X}", "type": int(x.type), "owner": owner_name(x.frm)}
        for x in idautils.XrefsTo(ea, 0)
        if x.iscode
    ]


def input_digest():
    path = ida_nalt.get_input_file_path()
    if path and os.path.isfile(path):
        with open(path, "rb") as fh:
            return hashlib.sha256(fh.read()).hexdigest()
    return "database-input-not-present; see project RE note"


def main():
    functions = []
    for ea, semantic in FUNCTIONS.items():
        func = ida_funcs.get_func(ea)
        if not func:
            functions.append({"start": f"0x{ea:X}", "error": "function missing"})
            continue
        functions.append(
            {
                "raw_name": ida_name.get_name(ea),
                "start": f"0x{func.start_ea:X}",
                "end": f"0x{func.end_ea:X}",
                "semantic": semantic,
                "confidence": "hypothesis until callers and consumers are interpreted",
                "source": "ORION2.EXE IDA database",
                "callers": callers(ea),
                "instructions": instruction_rows(func.start_ea, func.end_ea),
            }
        )
    payload = {
        "contract": "raw-location + semantic-label + confidence + source",
        "input": os.path.basename(ida_nalt.get_input_file_path()),
        "input_sha256": input_digest(),
        "ida_sdk_version": str(getattr(ida_pro, "IDA_SDK_VERSION", "9.4")),
        "address_space": "IDA linear",
        "functions": functions,
    }
    output = os.environ["MOO2_IDA_OUTPUT"]
    with open(output, "w", encoding="utf-8") as fh:
        json.dump(payload, fh, ensure_ascii=False, indent=2)
        fh.write("\n")


try:
    main()
except Exception:
    error = traceback.format_exc()
    output = os.environ.get("MOO2_IDA_OUTPUT", "/tmp/moo2-council-voting.json")
    with open(output + ".error", "w", encoding="utf-8") as fh:
        fh.write(error)
    ida_kernwin.msg(error + "\n")
    ida_pro.qexit(1)
else:
    ida_pro.qexit(0)
