"""非破壞性匯出戰略轟炸快速戰鬥鏈的 raw 指令、caller 與 references。

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


FUNCTIONS = {
    0x416CF: "戰略戰鬥狀態初始化候選",
    0x41F80: "戰略武器攻擊步驟候選",
    0x420C0: "戰略特殊攻擊步驟候選",
    0x4221F: "戰略飛彈攻擊步驟候選",
    0x42371: "殖民地耐久／目標值候選",
    0x4257E: "Strategic_Bombardment_／Resolve_Strat_Colony_Damage 候選",
    0x4267B: "Strategic_Bombardment_ 直接 caller／回傳值消費候選",
    0xE7678: "戰略轟炸結果 caller A 候選",
    0xE87D2: "戰略轟炸結果 caller B 候選",
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
    output = os.environ.get("MOO2_IDA_OUTPUT", "/tmp/moo2-strategic-bombardment.json")
    with open(output + ".error", "w", encoding="utf-8") as fh:
        fh.write(error)
    ida_kernwin.msg(error + "\n")
    ida_pro.qexit(1)
else:
    ida_pro.qexit(0)
