"""唯讀匯出原版議會投票公式的 IDA 證據。

輸出保留原始位址、名稱、運算元與位元組，不修改或儲存 IDA 資料庫。
"""

import hashlib
import json
import os

import ida_auto
import ida_bytes
import ida_funcs
import ida_ida
import ida_idaapi
import ida_kernwin
import ida_nalt
import ida_pro
import idautils
import idc


TARGET_EA = 0x15B90
OUT = os.environ.get("MOO2_RE_OUT", "/out/council-vote.json")
SOURCE_EXE = os.environ.get("MOO2_RE_EXE", ida_nalt.get_input_file_path())


def sha256(path):
    digest = hashlib.sha256()
    with open(path, "rb") as handle:
        for chunk in iter(lambda: handle.read(1024 * 1024), b""):
            digest.update(chunk)
    return digest.hexdigest()


def original_name(ea):
    return idc.get_name(ea, idc.GN_VISIBLE) or "<unnamed>"


def instruction(ea):
    size = idc.get_item_size(ea)
    raw = ida_bytes.get_bytes(ea, size) or b""
    return {
        "ea": f"0x{ea:X}",
        "bytes": raw.hex(),
        "mnemonic": idc.print_insn_mnem(ea),
        "operand_0": idc.print_operand(ea, 0),
        "operand_1": idc.print_operand(ea, 1),
        "disassembly": idc.generate_disasm_line(ea, 0) or "<unavailable>",
    }


def instruction_window(center_ea, function, before=16, after=16):
    items = list(idautils.FuncItems(function.start_ea))
    try:
        center_index = items.index(center_ea)
    except ValueError:
        return []
    start = max(0, center_index - before)
    end = min(len(items), center_index + after + 1)
    return [instruction(item_ea) for item_ea in items[start:end]]


def function_record(ea):
    function = ida_funcs.get_func(ea)
    if function is None:
        raise RuntimeError(f"0x{ea:X} 沒有函式邊界")
    instructions = [instruction(item_ea) for item_ea in idautils.FuncItems(function.start_ea)]
    callers = []
    for ref in idautils.CodeRefsTo(function.start_ea, 0):
        caller = ida_funcs.get_func(ref)
        callers.append({
            "call_site": f"0x{ref:X}",
            "call_instruction": instruction(ref),
            "caller_start": f"0x{caller.start_ea:X}" if caller else None,
            "caller_original_name": original_name(caller.start_ea) if caller else None,
            "context": instruction_window(ref, caller) if caller else [],
        })
    return {
        "requested_ea": f"0x{ea:X}",
        "start_ea": f"0x{function.start_ea:X}",
        "end_ea": f"0x{function.end_ea:X}",
        "original_name": original_name(function.start_ea),
        "instructions": instructions,
        "direct_callers": callers,
    }


def main():
    ida_auto.auto_wait()
    report = {
        "schema": "moo2.ida.re-evidence.v1",
        "evidence_scope": "static_only",
        "mutation": "none",
        "tool": {
            "name": "IDA Pro",
            "version": ida_kernwin.get_kernel_version(),
            "script": "tools/ida/audit_council_vote.py",
        },
        "input": {
            "ida_database_path": ida_nalt.get_input_file_path(),
            "source_executable_path": SOURCE_EXE,
            "source_executable_sha256": sha256(SOURCE_EXE),
            "processor": ida_ida.inf_get_procname(),
            "file_type_id": ida_ida.inf_get_filetype(),
        },
        "address_basis": "IDA linear address; DOS/4GW LE object #1 code base 0x10000",
        "semantic_claim": {
            "level": "unknown_pending_review",
            "note": "本檔只匯出原始證據；語意與證據等級由受版控 RE 文件審查。",
        },
        "target": function_record(TARGET_EA),
    }
    os.makedirs(os.path.dirname(OUT), exist_ok=True)
    with open(OUT, "w", encoding="utf-8") as handle:
        json.dump(report, handle, ensure_ascii=False, indent=2)
        handle.write("\n")
    ida_pro.qexit(0)


if __name__ == "__main__":
    main()
