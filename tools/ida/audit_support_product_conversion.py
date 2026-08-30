"""匯出 -12/-17/-11 支援產品由 UI pseudo-product 轉為 ship slot 的鏈。"""

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

OUT = os.environ.get("MOO2_RE_OUT", "/out/support-product-conversion.json")
SOURCE = os.environ.get("MOO2_RE_SOURCE", ida_nalt.get_input_file_path())
ROOTS = [0xAFF9E, 0xB2D21, 0xB41A2, 0xBDD2F, 0xE0DD6, 0xE11BC, 0xE36DF]


def sha256(path):
    h = hashlib.sha256()
    with open(path, "rb") as stream:
        for chunk in iter(lambda: stream.read(1024 * 1024), b""):
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
            "callers": [insn(x) for x in idautils.CodeRefsTo(f.start_ea, 0)]}


def main():
    ida_auto.auto_wait()
    report = {"schema": "moo2.ida.re-evidence.v1", "evidence_scope": "static_only", "mutation": "none",
              "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version(),
                       "script": "tools/ida/audit_support_product_conversion.py"},
              "input": {"database": ida_nalt.get_input_file_path(), "source": SOURCE,
                        "source_sha256": sha256(SOURCE), "processor": ida_ida.inf_get_procname(),
                        "min_ea": f"0x{ida_ida.inf_get_min_ea():X}", "max_ea": f"0x{ida_ida.inf_get_max_ea():X}"},
              "address_basis": "IDA linear; DOS/4GW LE object #1",
              "roots": {f"0x{ea:X}": function(ea) for ea in ROOTS}}
    os.makedirs(os.path.dirname(OUT), exist_ok=True)
    with open(OUT, "w", encoding="utf-8") as stream:
        json.dump(report, stream, ensure_ascii=False, indent=2); stream.write("\n")
    ida_pro.qexit(0)


if __name__ == "__main__":
    main()
