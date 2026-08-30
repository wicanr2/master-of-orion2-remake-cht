"""匯出 Compute_AI_Data_ 與三張暫存表指標的直接讀寫端。"""

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


OUT = os.environ.get("MOO2_RE_OUT", "/out/compute-ai-data-consumers.json")
SOURCE = os.environ.get("MOO2_RE_SOURCE", ida_nalt.get_input_file_path())
ROOTS = [0xD3A68, 0xD3BA0, 0xD3D34, 0xD574D, 0xD6ED4, 0xDCA69]
GLOBALS = [0x1AA1E4, 0x1AA1EC, 0x1AA1F8]


def sha256(path):
    digest = hashlib.sha256()
    with open(path, "rb") as stream:
        for chunk in iter(lambda: stream.read(1024 * 1024), b""):
            digest.update(chunk)
    return digest.hexdigest()


def insn(ea):
    size = idc.get_item_size(ea)
    return {
        "ea": f"0x{ea:X}", "bytes": (ida_bytes.get_bytes(ea, size) or b"").hex(),
        "mnem": idc.print_insn_mnem(ea), "op0": idc.print_operand(ea, 0),
        "op1": idc.print_operand(ea, 1), "line": idc.generate_disasm_line(ea, 0) or "<unavailable>",
    }


def function(ea, include_instructions=True):
    func = ida_funcs.get_func(ea)
    if not func:
        return {"requested": f"0x{ea:X}", "error": "no function"}
    result = {
        "requested": f"0x{ea:X}", "start": f"0x{func.start_ea:X}", "end": f"0x{func.end_ea:X}",
        "original_name": idc.get_name(func.start_ea) or "<unnamed>",
        "callers": [insn(x) for x in idautils.CodeRefsTo(func.start_ea, 0)],
    }
    if include_instructions:
        result["instructions"] = [insn(x) for x in idautils.FuncItems(func.start_ea)]
    return result


def global_refs(ea):
    refs = []
    for ref in idautils.XrefsTo(ea, 0):
        func = ida_funcs.get_func(ref.frm)
        refs.append({
            "reference": insn(ref.frm),
            "function": function(func.start_ea, False) if func else None,
            "xref_type": int(ref.type),
        })
    return refs


def main():
    ida_auto.auto_wait()
    report = {
        "schema": "moo2.ida.re-evidence.v1", "evidence_scope": "static_only", "mutation": "none",
        "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version(),
                 "script": "tools/ida/audit_compute_ai_data_consumers.py"},
        "input": {"database": ida_nalt.get_input_file_path(), "source": SOURCE,
                  "source_sha256": sha256(SOURCE), "processor": ida_ida.inf_get_procname(),
                  "min_ea": f"0x{ida_ida.inf_get_min_ea():X}", "max_ea": f"0x{ida_ida.inf_get_max_ea():X}"},
        "address_basis": "IDA linear; DOS/4GW LE object #1",
        "roots": {f"0x{ea:X}": function(ea) for ea in ROOTS},
        "cache_pointer_xrefs": {f"0x{ea:X}": global_refs(ea) for ea in GLOBALS},
    }
    os.makedirs(os.path.dirname(OUT), exist_ok=True)
    with open(OUT, "w", encoding="utf-8") as stream:
        json.dump(report, stream, ensure_ascii=False, indent=2)
        stream.write("\n")
    ida_pro.qexit(0)


if __name__ == "__main__":
    main()
