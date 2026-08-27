"""匯出 Diplomacy_Growth_ @ 0x4DD6B 的非破壞性證據。

保留原始函式名、位址、operand、bytes、caller、callee 與資料交叉參照；
語意分級由外部稽核文件負責，不在 IDA 資料庫改名。
"""

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


TARGET_EA = 0x4DD6B
TARGET_RELATION_INIT_EA = 0x4D78E
BASE_RELATION_TABLE_EA = 0x180ED4


def instruction(ea):
    insn = ida_ua.insn_t()
    size = ida_ua.decode_insn(insn, ea) or 1
    refs = []
    for ref in idautils.DataRefsFrom(ea):
        refs.append({"ea": hex(ref), "original_name": ida_name.get_name(ref)})
    return {
        "ea": hex(ea),
        "bytes": (ida_bytes.get_bytes(ea, size) or b"").hex(),
        "disasm": idc.generate_disasm_line(ea, 0) or "",
        "data_refs": refs,
    }


def function_instructions(fn, limit=1600):
    out = []
    ea = fn.start_ea
    while ea < fn.end_ea and len(out) < limit:
        out.append(instruction(ea))
        ea = idc.next_head(ea, fn.end_ea)
    return out


def function_identity(fn):
    raw = ida_bytes.get_bytes(fn.start_ea, fn.end_ea - fn.start_ea) or b""
    return {
        "original_name": ida_name.get_name(fn.start_ea),
        "start_ea": hex(fn.start_ea),
        "end_ea": hex(fn.end_ea),
        "size": fn.end_ea - fn.start_ea,
        "bytes_sha256": hashlib.sha256(raw).hexdigest(),
    }


def main():
    fn = ida_funcs.get_func(TARGET_EA)
    if fn is None:
        raise RuntimeError("Diplomacy_Growth_ function missing")

    target = function_identity(fn)
    target["instructions"] = function_instructions(fn)
    target["callers"] = []
    for xref in idautils.XrefsTo(fn.start_ea, 0):
        caller = ida_funcs.get_func(xref.frm)
        if caller is None:
            continue
        item_eas = list(idautils.FuncItems(caller.start_ea))
        try:
            index = item_eas.index(xref.frm)
        except ValueError:
            index = 0
        target["callers"].append({
            "call_ea": hex(xref.frm),
            "xref_type": int(xref.type),
            "caller": function_identity(caller),
            "window": [instruction(ea) for ea in item_eas[max(0, index - 16):index + 9]],
        })

    callees = {}
    for call_ea in idautils.FuncItems(fn.start_ea):
        insn = ida_ua.insn_t()
        if not ida_ua.decode_insn(insn, call_ea) or not ida_idp.is_call_insn(insn):
            continue
        for target_ea in idautils.CodeRefsFrom(call_ea, 0):
            callee = ida_funcs.get_func(target_ea)
            if callee is None:
                continue
            key = hex(callee.start_ea)
            entry = callees.setdefault(key, {
                "function": function_identity(callee),
                "call_sites": [],
            })
            entry["call_sites"].append(hex(call_ea))
    target["callees"] = list(callees.values())

    try:
        target["decompiler_navigation_only"] = str(ida_hexrays.decompile(fn.start_ea))
    except Exception as exc:
        target["decompiler_error"] = str(exc)

    # `[base+slot+61Fh]` 這類間接欄位不會出現在 IDA 直接 DataRefs。
    # 另存原始 operand 位點與所屬函式，供外部追 writer／consumer。
    relative_sites = []
    for ea in idautils.Heads():
        line = idc.generate_disasm_line(ea, 0) or ""
        if "61Fh]" not in line and "+61Fh]" not in line:
            continue
        owner = ida_funcs.get_func(ea)
        relative_sites.append({
            "instruction": instruction(ea),
            "function": function_identity(owner) if owner is not None else None,
        })
    target["relative_operand_61f_sites"] = relative_sites

    init_fn = ida_funcs.get_func(TARGET_RELATION_INIT_EA)
    if init_fn is not None:
        init_export = function_identity(init_fn)
        init_export["instructions"] = function_instructions(init_fn)
        init_export["callers"] = []
        for xref in idautils.XrefsTo(init_fn.start_ea, 0):
            owner = ida_funcs.get_func(xref.frm)
            init_export["callers"].append({
                "call_ea": hex(xref.frm),
                "caller": function_identity(owner) if owner is not None else None,
            })
        try:
            init_export["decompiler_navigation_only"] = str(ida_hexrays.decompile(init_fn.start_ea))
        except Exception as exc:
            init_export["decompiler_error"] = str(exc)
        target["target_relation_initializer"] = init_export

    table_raw = ida_bytes.get_bytes(BASE_RELATION_TABLE_EA, 14 * 14 * 2) or b""
    target["base_relation_table"] = {
        "start_ea": hex(BASE_RELATION_TABLE_EA),
        "shape": [14, 14],
        "stride_bytes": 28,
        "element_bytes": 2,
        "raw_hex": table_raw.hex(),
        "bytes_sha256": hashlib.sha256(table_raw).hexdigest(),
        "signed_low_bytes": [
            [
                int.from_bytes(table_raw[(row * 14 + col) * 2:(row * 14 + col) * 2 + 1],
                               "little", signed=True)
                for col in range(14)
            ]
            for row in range(14)
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
            "note": "本匯出只保存原始定位與資料流線索；語意需另經證據審查。",
        },
    }
    output = os.environ.get("MOO2_IDA_PROBE_OUT", "/tmp/diplomacy-growth-ida.json")
    with open(output, "w", encoding="utf-8") as stream:
        json.dump(result, stream, ensure_ascii=False, indent=2, sort_keys=True)
        stream.write("\n")
    idc.qexit(0)


main()
