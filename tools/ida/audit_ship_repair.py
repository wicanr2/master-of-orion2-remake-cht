"""非破壞性匯出 Repair_Ships_At_Colonies_ 與 Repair_Ship_Full_ 證據。

保留 raw name／IDA linear address／bytes／operand、caller 與 direct callee；不改名、
不套型別、不寫回資料庫。Hex-Rays 只供導覽，結論必須回查原始指令。
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


ROOTS = tuple(
    int(value, 0)
    for value in os.environ.get("MOO2_IDA_ROOTS", "0x580F5,0x581F3").split(",")
    if value.strip()
)


def raw_name(ea):
    func = ida_funcs.get_func(ea)
    return ida_name.get_name(func.start_ea) if func else ida_name.get_name(ea)


def callers(ea):
    rows = []
    for xref in idautils.XrefsTo(ea, 0):
        if not xref.iscode:
            continue
        owner = ida_funcs.get_func(xref.frm)
        rows.append({
            "from": f"0x{xref.frm:X}",
            "type": int(xref.type),
            "owner_start": f"0x{owner.start_ea:X}" if owner else None,
            "owner_raw_name": raw_name(owner.start_ea) if owner else "outside-function",
        })
    return rows


def instruction_rows(func):
    rows, calls = [], []
    ea = func.start_ea
    while ea < func.end_ea:
        insn = ida_ua.insn_t()
        size = ida_ua.decode_insn(insn, ea)
        if size <= 0:
            size = 1
        refs = list(idautils.CodeRefsFrom(ea, 0))
        text = ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or "")
        rows.append({
            "ea": f"0x{ea:X}",
            "bytes": (ida_bytes.get_bytes(ea, size) or b"").hex(" "),
            "text": text,
            "code_refs": [f"0x{x:X}" for x in refs],
            "data_refs": [f"0x{x:X}" for x in idautils.DataRefsFrom(ea)],
        })
        if idc.print_insn_mnem(ea).lower().startswith("call"):
            calls.append({
                "site": f"0x{ea:X}",
                "text": text,
                "targets": [{"ea": f"0x{x:X}", "raw_name": raw_name(x)} for x in refs],
            })
        ea += size
    return rows, calls


def pseudocode(ea):
    try:
        if not ida_hexrays.init_hexrays_plugin():
            return {"error": "Hex-Rays unavailable"}
        cfunc = ida_hexrays.decompile(ea)
        return {
            "warning": "navigation only; confirm against instructions",
            "text": "\n".join(ida_lines.tag_remove(line.line) for line in cfunc.get_pseudocode()),
        }
    except Exception as exc:  # noqa: BLE001
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
    targets = set(ROOTS)
    for root in ROOTS:
        func = ida_funcs.get_func(root)
        if not func:
            raise RuntimeError(f"function missing at 0x{root:X}")
        _, calls = instruction_rows(func)
        for call in calls:
            for target in call["targets"]:
                targets.add(int(target["ea"], 16))

    path = ida_nalt.get_input_file_path()
    payload = {
        "contract": "raw-location + semantic-label + confidence + source",
        "input": os.path.basename(path) if path else ida_nalt.get_root_filename(),
        "input_sha256": digest(path),
        "ida_sdk_version": str(getattr(ida_pro, "IDA_SDK_VERSION", "9.4")),
        "address_space": "IDA linear",
        "roots": [f"0x{x:X}" for x in ROOTS],
        "data_ranges": [
            {
                "start": "0x17D160",
                "end": "0x17D1A0",
                "bytes": (ida_bytes.get_bytes(0x17D160, 0x40) or b"").hex(" "),
                "note": "raw crew XP thresholds/cap neighborhood; interpretation requires consumers",
            }
        ],
        "functions": [function_payload(ea) for ea in sorted(targets)],
    }
    with open(os.environ["MOO2_IDA_OUTPUT"], "w", encoding="utf-8") as fh:
        json.dump(payload, fh, ensure_ascii=False, indent=2)
        fh.write("\n")


try:
    main()
except Exception:
    error = traceback.format_exc()
    output = os.environ.get("MOO2_IDA_OUTPUT", "/tmp/moo2-ship-repair.json")
    with open(output + ".error", "w", encoding="utf-8") as fh:
        fh.write(error)
    ida_kernwin.msg(error + "\n")
    ida_pro.qexit(1)
else:
    ida_pro.qexit(0)
