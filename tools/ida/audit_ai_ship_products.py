"""非破壞性匯出 MOO2 AI 艦艇／支援產品選擇所需證據。"""

import hashlib
import json
import os
import traceback

import ida_auto
import ida_bytes
import ida_funcs
import ida_hexrays
import ida_ida
import ida_kernwin
import ida_name
import ida_pro
import ida_ua
import idautils
import idc


ROOTS = {
    "raw_colony_product_list_builder": 0xAFF9E,
    "raw_colony_can_build_product": 0xE11BC,
    "raw_colony_product_text_or_status": 0xEEC98,
    "raw_ai_support_quota_a_producer": 0xCF3BD,
    "raw_ai_support_quota_bc_producer": 0xCF40D,
    "raw_ai_colony_product_pass": 0xD10EE,
    "raw_remove_all_buildings_of_type": 0xB206F,
    "raw_colony_product_cost": 0xE0DD6,
    "raw_colony_production_capacity_gate": 0xCFEDC,
    "raw_colony_blockade_or_special_gate": 0xDF8C1,
    "raw_ai_ship_slot_allocator": 0xAF7B4,
    "raw_ai_design_builder": 0x5663E,
    "raw_ai_design_cost": 0xCFAE5,
    "raw_ai_empire_inventory_summary": 0xCFCB6,
    "raw_apply_colony_production": 0xE380E,
}

GLOBALS = {
    "raw_quota_a_baseline": 0x1A7236,
    "raw_quota_group_a": 0x1A724C,
    "raw_quota_b_baseline": 0x1A7274,
    "raw_quota_group_b": 0x1A7275,
    "raw_quota_c_baseline": 0x1A7284,
    "raw_quota_group_c": 0x1A7285,
}


def digest(path):
    h = hashlib.sha256()
    with open(path, "rb") as stream:
        for chunk in iter(lambda: stream.read(1024 * 1024), b""):
            h.update(chunk)
    return h.hexdigest()


def instruction(ea):
    insn = ida_ua.insn_t()
    size = ida_ua.decode_insn(insn, ea) or 1
    return {
        "ea": f"0x{ea:X}",
        "bytes": (ida_bytes.get_bytes(ea, size) or b"").hex(),
        "text": idc.generate_disasm_line(ea, 0) or "",
        "code_refs": [f"0x{x:X}" for x in idautils.CodeRefsFrom(ea, 0)],
        "data_refs": [f"0x{x:X}" for x in idautils.DataRefsFrom(ea)],
    }


def function_record(requested):
    fn = ida_funcs.get_func(requested)
    if fn is None:
        return {"requested": f"0x{requested:X}", "error": "function missing"}
    raw = ida_bytes.get_bytes(fn.start_ea, fn.end_ea - fn.start_ea) or b""
    pseudo = None
    if ida_hexrays.init_hexrays_plugin():
        try:
            pseudo = str(ida_hexrays.decompile(fn.start_ea))
        except Exception as exc:
            pseudo = f"<decompile failed: {exc}>"
    return {
        "requested": f"0x{requested:X}",
        "original_name": ida_name.get_name(fn.start_ea) or "<unnamed>",
        "start_ea": f"0x{fn.start_ea:X}",
        "end_ea": f"0x{fn.end_ea:X}",
        "bytes_sha256": hashlib.sha256(raw).hexdigest(),
        "pseudocode_navigation_only": pseudo,
        "instructions": [instruction(ea) for ea in idautils.FuncItems(fn.start_ea)],
        "callers": [instruction(xref.frm) for xref in idautils.XrefsTo(fn.start_ea, 0)
                    if ida_funcs.get_func(xref.frm) is not None],
    }


def global_xrefs(ea):
    rows = []
    for xref in idautils.XrefsTo(ea, 0):
        fn = ida_funcs.get_func(xref.frm)
        rows.append({
            "reference": instruction(xref.frm),
            "function_start": f"0x{fn.start_ea:X}" if fn else None,
            "function_original_name": ida_name.get_name(fn.start_ea) if fn else None,
        })
    return rows


def product_constant_functions():
    needles = ("0fff1h", "0fff4h", "0fff9h")
    out = {}
    for fea in idautils.Functions():
        hits = []
        for ea in idautils.FuncItems(fea):
            text = (idc.generate_disasm_line(ea, 0) or "").lower()
            if any(needle in text for needle in needles):
                hits.append(instruction(ea))
        if hits:
            out[f"0x{fea:X}"] = {
                "original_name": ida_name.get_name(fea) or "<unnamed>",
                "hits": hits,
                "callers": [f"0x{x:X}" for x in idautils.CodeRefsTo(fea, 0)],
            }
    return out


def main():
    ida_auto.auto_wait()
    source = os.environ["MOO2_IDA_INPUT"]
    database = os.environ["MOO2_IDA_DATABASE"]
    report = {
        "schema": "moo2.ida.re-evidence.v1",
        "contract": "raw-location + original-name + bytes + xrefs; semantics reviewed externally",
        "mutation": "none",
        "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version()},
        "input": {
            "file": os.path.basename(source),
            "source_sha256": digest(source),
            "database_sha256": digest(database),
            "processor": ida_ida.inf_get_procname(),
        },
        "address_basis": "IDA linear; DOS/4GW LE image",
        "semantic_status": "unknown_pending_review",
        "roots": {name: function_record(ea) for name, ea in ROOTS.items()},
        "global_xrefs": {name: {"ea": f"0x{ea:X}", "xrefs": global_xrefs(ea)}
                         for name, ea in GLOBALS.items()},
        "product_constant_functions": product_constant_functions(),
    }
    with open(os.environ["MOO2_IDA_OUTPUT"], "w", encoding="utf-8") as stream:
        json.dump(report, stream, ensure_ascii=False, indent=2)
        stream.write("\n")


try:
    main()
except Exception:
    error = traceback.format_exc()
    out = os.environ.get("MOO2_IDA_OUTPUT", "/tmp/ai-ship-products.json")
    with open(out + ".error", "w", encoding="utf-8") as stream:
        stream.write(error)
    ida_kernwin.msg(error + "\n")
    ida_pro.qexit(1)
else:
    ida_pro.qexit(0)
