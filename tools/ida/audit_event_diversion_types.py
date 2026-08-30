"""匯出 sub_23DFE 的 raw 事件型別、事件表欄位與所有 caller，供非破壞性語意審查。"""

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


OUT = os.environ.get("MOO2_RE_OUT", "/out/event-diversion-types.json")
SOURCE = os.environ.get("MOO2_RE_SOURCE", ida_nalt.get_input_file_path())
ROOTS = [0x2027E, 0x206A2, 0x2230A, 0x23DFE, 0xE2710, 0xE2DCA]


def sha256(path):
    digest = hashlib.sha256()
    with open(path, "rb") as stream:
        for chunk in iter(lambda: stream.read(1024 * 1024), b""):
            digest.update(chunk)
    return digest.hexdigest()


def instruction(ea):
    size = idc.get_item_size(ea)
    return {
        "ea": f"0x{ea:X}",
        "bytes": (ida_bytes.get_bytes(ea, size) or b"").hex(),
        "mnem": idc.print_insn_mnem(ea),
        "op0": idc.print_operand(ea, 0),
        "op1": idc.print_operand(ea, 1),
        "line": idc.generate_disasm_line(ea, 0) or "<unavailable>",
    }


def function(ea):
    func = ida_funcs.get_func(ea)
    if not func:
        return {"requested": f"0x{ea:X}", "error": "no function"}
    return {
        "requested": f"0x{ea:X}",
        "start": f"0x{func.start_ea:X}",
        "end": f"0x{func.end_ea:X}",
        "original_name": idc.get_name(func.start_ea) or "<unnamed>",
        "instructions": [instruction(item) for item in idautils.FuncItems(func.start_ea)],
        "callers": [instruction(item) for item in idautils.CodeRefsTo(func.start_ea, 0)],
    }


def operand_hits(needles):
    hits = []
    for func_ea in idautils.Functions(ida_ida.inf_get_min_ea(), ida_ida.inf_get_max_ea()):
        found = []
        for ea in idautils.FuncItems(func_ea):
            line = idc.generate_disasm_line(ea, 0) or ""
            if any(needle.lower() in line.lower() for needle in needles):
                found.append(instruction(ea))
        if found:
            hits.append({
                "function_start": f"0x{func_ea:X}",
                "original_name": idc.get_name(func_ea) or "<unnamed>",
                "hits": found,
            })
    return hits


def main():
    ida_auto.auto_wait()
    report = {
        "schema": "moo2.ida.re-evidence.v1",
        "evidence_scope": "static_only",
        "mutation": "none",
        "tool": {
            "name": "IDA Pro",
            "version": ida_kernwin.get_kernel_version(),
            "script": "tools/ida/audit_event_diversion_types.py",
        },
        "input": {
            "database": ida_nalt.get_input_file_path(),
            "source": SOURCE,
            "source_sha256": sha256(SOURCE),
            "processor": ida_ida.inf_get_procname(),
            "min_ea": f"0x{ida_ida.inf_get_min_ea():X}",
            "max_ea": f"0x{ida_ida.inf_get_max_ea():X}",
        },
        "address_basis": "IDA linear; DOS/4GW LE object #1",
        "roots": {f"0x{ea:X}": function(ea) for ea in ROOTS},
        "event_global_hits": operand_hits(["19ABA4", "19ABA5", "19ABA7", "19ABA9", "19ABAB"]),
    }
    os.makedirs(os.path.dirname(OUT), exist_ok=True)
    with open(OUT, "w", encoding="utf-8") as stream:
        json.dump(report, stream, ensure_ascii=False, indent=2)
        stream.write("\n")
    ida_pro.qexit(0)


if __name__ == "__main__":
    main()
