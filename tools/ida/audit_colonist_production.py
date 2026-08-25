"""唯讀匯出 MOO2 逐 colonist 產出、忠誠與職務修正鏈。"""
import hashlib
import json
import os

import ida_auto
import ida_bytes
import ida_funcs
import ida_ida
import ida_kernwin
import ida_nalt
import ida_pro
import idautils
import idc

OUT = os.environ.get("MOO2_RE_OUT", "/out/colonist-production.json")
SOURCE = os.environ.get("MOO2_RE_SOURCE", ida_nalt.get_input_file_path())
ROOTS = {
    "raw_Colony_Base_Production_By_Job": 0xDE280,
    "raw_Colonist_Production_Adjustment": 0xDDF2C,
    "raw_Colonist_Base_Production_Unit": 0xDE22C,
    "raw_Research_Colonist_Base": 0xDFE77,
    "raw_Food_Colonist_Base": 0xDE0C6,
    "raw_Industry_Colonist_Base": 0xDED47,
    "raw_Research_Colonist_Recompute": 0xDFDC6,
    "raw_Gravity_Adjusted_Colonist_Value": 0xDDFD3,
    "raw_Colony_Assigned_Officer": 0xDD9F2,
    "raw_Apply_Colony_Pop_Growth": 0xE2DCA,
    "raw_Apply_Colonist_Assimilation": 0xE3456,
    "raw_Race_Population_Limit": 0xE0C1D,
    "raw_Food_Environment_Base": 0xDE03E,
    "raw_Industry_Environment_Base": 0xDEC95,
    "raw_Recompute_Food_Consumption": 0xDEB4B,
    "raw_Recompute_Industry_Consumption": 0xDF546,
    "raw_Recompute_Race_Population_Growth": 0xE1839,
}

MASKS = {0x7, 0x70, 0x200, 0x400}


def sha256(path):
    h = hashlib.sha256()
    with open(path, "rb") as f:
        for chunk in iter(lambda: f.read(1024 * 1024), b""):
            h.update(chunk)
    return h.hexdigest()


def insn(ea):
    size = idc.get_item_size(ea)
    return {"ea": f"0x{ea:X}", "bytes": (ida_bytes.get_bytes(ea, size) or b"").hex(),
            "line": idc.generate_disasm_line(ea, 0) or "<unavailable>"}


def function(ea):
    f = ida_funcs.get_func(ea)
    if not f:
        return {"requested": f"0x{ea:X}", "error": "no function"}
    callers = []
    for ref in idautils.CodeRefsTo(f.start_ea, 0):
        cf = ida_funcs.get_func(ref)
        callers.append({"instruction": insn(ref),
                        "function_start": f"0x{cf.start_ea:X}" if cf else None,
                        "function_name": idc.get_name(cf.start_ea) if cf else None})
    callees = []
    for item in idautils.FuncItems(f.start_ea):
        if idc.print_insn_mnem(item).lower() == "call":
            callees.append({"instruction": insn(item),
                            "targets": [{"ea": f"0x{x:X}", "name": idc.get_name(x) or "<unnamed>"}
                                        for x in idautils.CodeRefsFrom(item, 0)]})
    return {"start": f"0x{f.start_ea:X}", "end": f"0x{f.end_ea:X}",
            "original_name": idc.get_name(f.start_ea) or "<unnamed>",
            "instructions": [insn(x) for x in idautils.FuncItems(f.start_ea)],
            "callers": callers, "callees": callees}


def raw(ea, size):
    return {"ea": f"0x{ea:X}", "size": size, "bytes": (ida_bytes.get_bytes(ea, size) or b"").hex()}


def immediate_mask_uses():
    """列出 packed colonist 候選遮罩的立即數使用點，保留函式脈絡。"""
    hits = []
    for f_ea in idautils.Functions():
        f = ida_funcs.get_func(f_ea)
        if not f:
            continue
        for ea in idautils.FuncItems(f.start_ea):
            values = []
            for op in range(3):
                if idc.get_operand_type(ea, op) == idc.o_imm:
                    values.append(idc.get_operand_value(ea, op))
            matched = sorted(set(values) & MASKS)
            if matched:
                hits.append({
                    "function_start": f"0x{f.start_ea:X}",
                    "function_name": idc.get_name(f.start_ea) or "<unnamed>",
                    "matched_immediates": [f"0x{x:X}" for x in matched],
                    "instruction": insn(ea),
                })
    return hits


def packed_colonist_functions():
    """找同函式內同時出現 colony +0x0C 陣列、4-byte stride 與 packed 位元操作的候選。"""
    out = []
    for f_ea in idautils.Functions():
        f = ida_funcs.get_func(f_ea)
        if not f:
            continue
        items = list(idautils.FuncItems(f.start_ea))
        lines = [idc.generate_disasm_line(ea, 0) or "" for ea in items]
        has_colonist_base = any("+0Ch" in line or "+0C" in line for line in lines)
        has_packed_bits = any(
            (idc.print_insn_mnem(ea).lower() in {"and", "test", "shr", "shl"})
            and any(idc.get_operand_type(ea, op) == idc.o_imm and
                    idc.get_operand_value(ea, op) in {2, 4, 7, 0xF, 0x70, 0x180}
                    for op in range(3))
            for ea in items
        )
        has_stride = any("*4" in line or "shl" in line and ", 2" in line for line in lines)
        if has_colonist_base and has_packed_bits and has_stride:
            out.append(function(f.start_ea))
    return out


def colony_offset_functions():
    """列出直接顯示 race-pop/growth/assimilation 欄位位移的函式，供追寫入端。"""
    needles = ("+0B4h", "+0C8h", "+12Eh", "+12Fh")
    out = []
    for f_ea in idautils.Functions():
        f = ida_funcs.get_func(f_ea)
        if not f:
            continue
        hits = []
        for ea in idautils.FuncItems(f.start_ea):
            line = idc.generate_disasm_line(ea, 0) or ""
            if any(n in line for n in needles):
                hits.append(insn(ea))
        if hits:
            out.append({"start": f"0x{f.start_ea:X}", "end": f"0x{f.end_ea:X}",
                        "original_name": idc.get_name(f.start_ea) or "<unnamed>", "hits": hits})
    return out


def consumption_offset_functions():
    """列出殖民地逐類食物／工業消耗欄位 +0xFC..+0x103 的直接讀寫端。"""
    needles = tuple(f"+{x:X}h" for x in range(0xFC, 0x104))
    out = []
    for f_ea in idautils.Functions():
        f = ida_funcs.get_func(f_ea)
        if not f:
            continue
        hits = []
        for ea in idautils.FuncItems(f.start_ea):
            line = idc.generate_disasm_line(ea, 0) or ""
            if any(n in line for n in needles):
                hits.append(insn(ea))
        if hits:
            out.append({"start": f"0x{f.start_ea:X}", "end": f"0x{f.end_ea:X}",
                        "original_name": idc.get_name(f.start_ea) or "<unnamed>", "hits": hits})
    return out


def main():
    ida_auto.auto_wait()
    roots = {name: function(ea) for name, ea in ROOTS.items()}
    direct = {}
    for root in roots.values():
        for call in root.get("callees", []):
            for target in call["targets"]:
                ea = int(target["ea"], 16)
                direct[f"{target['name']}@0x{ea:X}"] = function(ea)
    report = {
        "schema": "moo2.ida.re-evidence.v1", "evidence_scope": "static_only", "mutation": "none",
        "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version(),
                 "script": "tools/ida/audit_colonist_production.py"},
        "input": {"database": ida_nalt.get_input_file_path(), "source": SOURCE,
                  "source_sha256": sha256(SOURCE), "processor": ida_ida.inf_get_procname(),
                  "min_ea": f"0x{ida_ida.inf_get_min_ea():X}", "max_ea": f"0x{ida_ida.inf_get_max_ea():X}"},
        "address_basis": "IDA linear; DOS/4GW LE object #1",
        "roots": roots, "direct_callees": direct,
        "tables": {
            "raw_byte_DD4D7": raw(0xDD4D7, 16),
            "raw_byte_DD4F5": raw(0xDD4F5, 16),
        },
        "packed_colonist_mask_candidates": immediate_mask_uses(),
        "packed_colonist_function_candidates": packed_colonist_functions(),
        "colony_offset_function_candidates": colony_offset_functions(),
        "consumption_offset_function_candidates": consumption_offset_functions(),
    }
    os.makedirs(os.path.dirname(OUT), exist_ok=True)
    with open(OUT, "w", encoding="utf-8") as f:
        json.dump(report, f, ensure_ascii=False, indent=2)
        f.write("\n")
    ida_pro.qexit(0)


if __name__ == "__main__":
    main()
