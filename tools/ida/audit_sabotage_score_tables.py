"""唯讀匯出 MOO2 SABOTAGE 兩個 score global 的全部直接讀寫與上游函式。"""

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


OUT = os.environ.get("MOO2_RE_OUT", "/out/sabotage-score-tables.json")
SOURCE = os.environ.get("MOO2_RE_SOURCE", ida_nalt.get_input_file_path())
GLOBALS = {"raw_score_a": 0x1ACE78, "raw_score_b": 0x1ACE7A}
DATA = {
    "raw_government_defense_table": (0x100A36, 8),
}
ROOTS = {
    "raw_Spy_Technology_Bonus": 0x100A3E,
    "raw_Compute_Spy_Bonuses": 0x100A83,
    "raw_PreResolve_Spy_Status": 0x1018A3,
    "raw_AI_Spy_Decision": 0x100D19,
    "raw_N_Spies_Bonus": 0x1014A4,
    "raw_Resolve_Spies": 0x10192B,
    "raw_Spy_Slot_Helper": 0x101483,
    "raw_Relationship_Count": 0x1026CF,
    "raw_Relationship_Mode": 0x1026F1,
    "raw_Relationship_Pack": 0x10278D,
}


def data_evidence(ea, size):
    raw = ida_bytes.get_bytes(ea, size) or b""
    return {"ea": f"0x{ea:X}", "original_name": idc.get_name(ea) or "<unnamed>",
            "size": size, "bytes": raw.hex(),
            "signed_bytes": [x if x < 0x80 else x - 0x100 for x in raw]}


def sha256(path):
    h = hashlib.sha256()
    with open(path, "rb") as f:
        for chunk in iter(lambda: f.read(1024 * 1024), b""):
            h.update(chunk)
    return h.hexdigest()


def insn(ea):
    size = idc.get_item_size(ea)
    return {"ea": f"0x{ea:X}", "bytes": (ida_bytes.get_bytes(ea, size) or b"").hex(),
            "mnem": idc.print_insn_mnem(ea), "op0": idc.print_operand(ea, 0),
            "op1": idc.print_operand(ea, 1), "line": idc.generate_disasm_line(ea, 0) or "<unavailable>"}


def function(ea):
    f = ida_funcs.get_func(ea)
    if not f:
        return {"requested": f"0x{ea:X}", "error": "no function"}
    return {"requested": f"0x{ea:X}", "start": f"0x{f.start_ea:X}", "end": f"0x{f.end_ea:X}",
            "original_name": idc.get_name(f.start_ea) or "<unnamed>",
            "instructions": [insn(x) for x in idautils.FuncItems(f.start_ea)],
            "callers": [{"instruction": insn(ref),
                         "function_start": f"0x{ida_funcs.get_func(ref).start_ea:X}" if ida_funcs.get_func(ref) else None,
                         "function_name": idc.get_name(ida_funcs.get_func(ref).start_ea) if ida_funcs.get_func(ref) else None}
                        for ref in idautils.CodeRefsTo(f.start_ea, 0)]}


def global_evidence(ea):
    refs = sorted(set(idautils.DataRefsTo(ea)))
    functions = {}
    for ref in refs:
        f = ida_funcs.get_func(ref)
        key = f"0x{f.start_ea:X}" if f else "outside_function"
        functions.setdefault(key, {"original_name": idc.get_name(f.start_ea) if f else None, "refs": []})
        functions[key]["refs"].append(insn(ref))
    return {"ea": f"0x{ea:X}", "original_name": idc.get_name(ea) or "<unnamed>",
            "bytes": (ida_bytes.get_bytes(ea, 2) or b"").hex(), "word_le": ida_bytes.get_word(ea),
            "direct_refs": [insn(x) for x in refs], "functions": functions}


def main():
    ida_auto.auto_wait()
    report = {"schema": "moo2.ida.re-evidence.v1", "evidence_scope": "static_only", "mutation": "none",
              "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version(),
                       "script": "tools/ida/audit_sabotage_score_tables.py"},
              "input": {"database": ida_nalt.get_input_file_path(), "source": SOURCE,
                        "source_sha256": sha256(SOURCE), "processor": ida_ida.inf_get_procname(),
                        "min_ea": f"0x{ida_ida.inf_get_min_ea():X}", "max_ea": f"0x{ida_ida.inf_get_max_ea():X}"},
              "address_basis": "IDA linear; DOS/4GW LE object #1",
              "globals": {name: global_evidence(ea) for name, ea in GLOBALS.items()},
              "data": {name: data_evidence(ea, size) for name, (ea, size) in DATA.items()},
              "roots": {name: function(ea) for name, ea in ROOTS.items()}}
    os.makedirs(os.path.dirname(OUT), exist_ok=True)
    with open(OUT, "w", encoding="utf-8") as f:
        json.dump(report, f, ensure_ascii=False, indent=2)
        f.write("\n")
    ida_pro.qexit(0)


if __name__ == "__main__":
    main()
