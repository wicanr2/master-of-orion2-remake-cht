"""匯出 sub_2552D 的非破壞性 IDA 證據。"""

import hashlib
import json
import os

import ida_bytes
import ida_funcs
import ida_hexrays
import ida_ida
import ida_idp
import ida_name
import ida_nalt
import ida_ua
import idautils
import idc


TARGET_EA = 0x2552D
RELATED_EAS = (0x252A7, 0x25AD2, 0x25BC6, 0x5232E, 0x500CF, 0x52049, 0x524C3)
GOVERNMENT_SCORE_TABLE_EA = 0x180CCC
RELATIVE_OPERANDS = ("+617h]", "+627h]", "+62Fh]", "+637h]", "+63Fh]",
                     "+68Fh]", "+69Fh]", "+6D7h]", "+71Fh]")


def instruction(ea):
    insn = ida_ua.insn_t()
    size = ida_ua.decode_insn(insn, ea) or 1
    return {
        "ea": hex(ea),
        "bytes": (ida_bytes.get_bytes(ea, size) or b"").hex(),
        "disasm": idc.generate_disasm_line(ea, 0) or "",
        "code_refs": [hex(x) for x in idautils.CodeRefsFrom(ea, 0)],
        "data_refs": [
            {"ea": hex(x), "original_name": ida_name.get_name(x)}
            for x in idautils.DataRefsFrom(ea)
        ],
    }


def identity(fn):
    raw = ida_bytes.get_bytes(fn.start_ea, fn.end_ea - fn.start_ea) or b""
    return {
        "original_name": ida_name.get_name(fn.start_ea),
        "start_ea": hex(fn.start_ea),
        "end_ea": hex(fn.end_ea),
        "size": fn.end_ea - fn.start_ea,
        "bytes_sha256": hashlib.sha256(raw).hexdigest(),
    }


def instructions(fn):
    return [instruction(ea) for ea in idautils.FuncItems(fn.start_ea)]


def main():
    fn = ida_funcs.get_func(TARGET_EA)
    if fn is None:
        raise RuntimeError("sub_2552D missing")
    target = identity(fn)
    target["instructions"] = instructions(fn)
    target["callers"] = []
    for xref in idautils.XrefsTo(fn.start_ea, 0):
        caller = ida_funcs.get_func(xref.frm)
        if caller is None:
            continue
        items = list(idautils.FuncItems(caller.start_ea))
        pos = items.index(xref.frm) if xref.frm in items else 0
        target["callers"].append({
            "call_ea": hex(xref.frm),
            "caller": identity(caller),
            "window": [instruction(ea) for ea in items[max(0, pos - 20):pos + 13]],
        })

    callees = {}
    for ea in idautils.FuncItems(fn.start_ea):
        insn = ida_ua.insn_t()
        if not ida_ua.decode_insn(insn, ea) or not ida_idp.is_call_insn(insn):
            continue
        for ref in idautils.CodeRefsFrom(ea, 0):
            callee = ida_funcs.get_func(ref)
            if callee is None:
                continue
            item = callees.setdefault(hex(callee.start_ea), {
                "function": identity(callee), "call_sites": [],
            })
            item["call_sites"].append(hex(ea))
    target["callees"] = list(callees.values())
    try:
        target["decompiler_navigation_only"] = str(ida_hexrays.decompile(fn.start_ea))
    except Exception as exc:
        target["decompiler_error"] = str(exc)

    target["related_functions"] = []
    for ea in RELATED_EAS:
        related = ida_funcs.get_func(ea)
        if related is None:
            continue
        item = identity(related)
        item["instructions"] = instructions(related)
        try:
            item["decompiler_navigation_only"] = str(ida_hexrays.decompile(related.start_ea))
        except Exception as exc:
            item["decompiler_error"] = str(exc)
        target["related_functions"].append(item)

    target["relative_operand_sites"] = []
    for ea in idautils.Heads():
        line = idc.generate_disasm_line(ea, 0) or ""
        if not any(token in line for token in RELATIVE_OPERANDS):
            continue
        owner = ida_funcs.get_func(ea)
        target["relative_operand_sites"].append({
            "instruction": instruction(ea),
            "function": identity(owner) if owner is not None else None,
        })

    table = ida_bytes.get_bytes(GOVERNMENT_SCORE_TABLE_EA, 16) or b""
    target["government_score_table"] = {
        "start_ea": hex(GOVERNMENT_SCORE_TABLE_EA),
        "raw_hex": table.hex(),
        "bytes_sha256": hashlib.sha256(table).hexdigest(),
        "signed_words": [
            int.from_bytes(table[i:i + 2], "little", signed=True)
            for i in range(0, len(table), 2)
        ],
    }

    result = {
        "tool": "IDA Pro 9.4 IDAPython",
        "root_filename": ida_nalt.get_root_filename(),
        "input_path": ida_nalt.get_input_file_path(),
        "input_sha256": ida_nalt.retrieve_input_file_sha256().hex(),
        "processor": ida_ida.inf_get_procname(),
        "compiler_id": int(ida_ida.inf_get_cc_id()),
        "address_space": "IDA linear EA",
        "target": target,
        "semantic_claim": {
            "level": "unknown",
            "note": "只保存原始定位與資料流；外部稽核另行分級。",
        },
    }
    output = os.environ.get("MOO2_IDA_PROBE_OUT", "/tmp/npc-treaty-negotiations.json")
    with open(output, "w", encoding="utf-8") as stream:
        json.dump(result, stream, ensure_ascii=False, indent=2, sort_keys=True)
        stream.write("\n")
    idc.qexit(0)


main()
