"""非破壞性匯出 sub_E87D2 的 call graph、原始指令與 Hex-Rays 導覽文字。

輸出保留 raw name／IDA linear address／bytes／operand；不改名、不套型別、不寫回資料庫。
Hex-Rays 文字只供導覽，所有結論仍須回到原始指令與 caller／consumer。
"""

import hashlib
import json
import os
import traceback

import ida_bytes
import ida_funcs
import ida_hexrays
import ida_kernwin
import ida_lines
import ida_name
import ida_nalt
import ida_pro
import ida_ua
import idautils
import idc


ROOT = 0xE87D2
SEED_HELPERS = (0xE85F7, 0xE86A0, 0xE87D2, 0xE938C, 0x4267B, 0xDD2F2)


def raw_name(ea):
    func = ida_funcs.get_func(ea)
    return ida_name.get_name(func.start_ea) if func else ida_name.get_name(ea)


def callers(ea):
    rows = []
    for xref in idautils.XrefsTo(ea, 0):
        if not xref.iscode:
            continue
        owner = ida_funcs.get_func(xref.frm)
        rows.append(
            {
                "from": f"0x{xref.frm:X}",
                "type": int(xref.type),
                "owner_start": f"0x{owner.start_ea:X}" if owner else None,
                "owner_raw_name": raw_name(owner.start_ea) if owner else "outside-function",
            }
        )
    return rows


def instruction_rows(func):
    rows = []
    calls = []
    ea = func.start_ea
    while ea < func.end_ea:
        insn = ida_ua.insn_t()
        size = ida_ua.decode_insn(insn, ea)
        if size <= 0:
            size = 1
        refs = list(idautils.CodeRefsFrom(ea, 0))
        text = ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or "")
        row = {
            "ea": f"0x{ea:X}",
            "bytes": (ida_bytes.get_bytes(ea, size) or b"").hex(" "),
            "text": text,
            "code_refs": [f"0x{x:X}" for x in refs],
            "data_refs": [f"0x{x:X}" for x in idautils.DataRefsFrom(ea)],
        }
        rows.append(row)
        if idc.print_insn_mnem(ea).lower().startswith("call"):
            calls.append(
                {
                    "site": f"0x{ea:X}",
                    "text": text,
                    "targets": [
                        {"ea": f"0x{x:X}", "raw_name": raw_name(x)} for x in refs
                    ],
                }
            )
        ea += size
    return rows, calls


def pseudocode(ea):
    try:
        if not ida_hexrays.init_hexrays_plugin():
            return {"error": "Hex-Rays unavailable"}
        cfunc = ida_hexrays.decompile(ea)
        return {
            "warning": "navigation only; not evidence without raw instruction confirmation",
            "text": "\n".join(ida_lines.tag_remove(line.line) for line in cfunc.get_pseudocode()),
        }
    except Exception as exc:  # noqa: BLE001 - evidence exporter must preserve failure
        return {"error": repr(exc)}


def function_payload(ea):
    func = ida_funcs.get_func(ea)
    if not func:
        return {"requested": f"0x{ea:X}", "error": "function missing"}
    instructions, calls = instruction_rows(func)
    return {
        "requested": f"0x{ea:X}",
        "start": f"0x{func.start_ea:X}",
        "end": f"0x{func.end_ea:X}",
        "raw_name": raw_name(func.start_ea),
        "callers": callers(func.start_ea),
        "calls": calls,
        "instructions": instructions,
        "pseudocode": pseudocode(func.start_ea),
    }


def digest(path):
    if path and os.path.isfile(path):
        with open(path, "rb") as fh:
            return hashlib.sha256(fh.read()).hexdigest()
    return "database-input-not-present; see project RE note"


def main():
    root_func = ida_funcs.get_func(ROOT)
    if not root_func:
        raise RuntimeError("sub_E87D2 function missing")
    _, root_calls = instruction_rows(root_func)
    targets = set(SEED_HELPERS)
    for call in root_calls:
        for target in call["targets"]:
            targets.add(int(target["ea"], 16))

    path = ida_nalt.get_input_file_path()
    payload = {
        "contract": "raw-location + semantic-label + confidence + source",
        "input": os.path.basename(path) if path else ida_nalt.get_root_filename(),
        "input_sha256": digest(path),
        "ida_sdk_version": str(getattr(ida_pro, "IDA_SDK_VERSION", "9.4")),
        "address_space": "IDA linear",
        "root": f"0x{ROOT:X}",
        "functions": [function_payload(ea) for ea in sorted(targets)],
    }
    with open(os.environ["MOO2_IDA_OUTPUT"], "w", encoding="utf-8") as fh:
        json.dump(payload, fh, ensure_ascii=False, indent=2)
        fh.write("\n")


try:
    main()
except Exception:
    error = traceback.format_exc()
    output = os.environ.get("MOO2_IDA_OUTPUT", "/tmp/moo2-bombardment-writeback.json")
    with open(output + ".error", "w", encoding="utf-8") as fh:
        fh.write(error)
    ida_kernwin.msg(error + "\n")
    ida_pro.qexit(1)
else:
    ida_pro.qexit(0)
