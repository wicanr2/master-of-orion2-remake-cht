"""唯讀盤點 MOO2 首都／士氣相關符號、字串與交叉參照。

不修改或儲存 IDB；輸出保留原始名稱、IDA 線性位址與引用指令。
"""
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

OUT = os.environ.get("MOO2_RE_OUT", "/out/capital-morale.json")
SOURCE = os.environ.get("MOO2_RE_SOURCE", ida_nalt.get_input_file_path())
TOKENS = ("capitol", "capital", "morale", "anarch")
ROOTS = {
    "raw_sub_23DFE_event_colony_filter": 0x23DFE,
    "raw_Colonize_Planet": 0xBB082,
    "raw_Ground_Combat_Round": 0xEC4FE,
    "raw_sub_E2710_colony_player_adjustment": 0xE2710,
    "raw_Apply_Colony_Pop_Growth": 0xE2DCA,
    "raw_Colony_Base_Production_By_Job": 0xDE280,
}


def sha256(path):
    h = hashlib.sha256()
    with open(path, "rb") as f:
        for chunk in iter(lambda: f.read(1024 * 1024), b""):
            h.update(chunk)
    return h.hexdigest()


def insn(ea):
    size = idc.get_item_size(ea)
    return {
        "ea": f"0x{ea:X}",
        "bytes": (ida_bytes.get_bytes(ea, size) or b"").hex(),
        "line": idc.generate_disasm_line(ea, 0) or "<unavailable>",
    }


def xrefs_to(ea):
    out = []
    for ref in idautils.XrefsTo(ea, 0):
        f = ida_funcs.get_func(ref.frm)
        out.append({
            "from": insn(ref.frm),
            "function_start": f"0x{f.start_ea:X}" if f else None,
            "function_name": idc.get_name(f.start_ea) if f else None,
            "type": ref.type,
        })
    return out


def function(ea):
    f = ida_funcs.get_func(ea)
    if not f:
        return None
    return {
        "start": f"0x{f.start_ea:X}",
        "end": f"0x{f.end_ea:X}",
        "original_name": idc.get_name(f.start_ea) or "<unnamed>",
        "instructions": [insn(x) for x in idautils.FuncItems(f.start_ea)],
        "callers": xrefs_to(f.start_ea),
    }


def main():
    ida_auto.auto_wait()
    names = []
    funcs = {}
    for ea, name in idautils.Names():
        if any(t in name.lower() for t in TOKENS):
            names.append({"ea": f"0x{ea:X}", "original_name": name, "xrefs": xrefs_to(ea)})
            f = ida_funcs.get_func(ea)
            if f:
                funcs[f"0x{f.start_ea:X}"] = function(f.start_ea)
    strings = []
    for item in idautils.Strings():
        text = str(item)
        if any(t in text.lower() for t in TOKENS):
            strings.append({"ea": f"0x{item.ea:X}", "text": text, "xrefs": xrefs_to(item.ea)})
            for ref in idautils.XrefsTo(item.ea, 0):
                f = ida_funcs.get_func(ref.frm)
                if f:
                    funcs[f"0x{f.start_ea:X}"] = function(f.start_ea)
    capitol_status_hits = []
    for f_ea in idautils.Functions():
        for item in idautils.FuncItems(f_ea):
            line = idc.generate_disasm_line(item, 0) or ""
            if "+137h" in line or "+0137h" in line:
                capitol_status_hits.append({
                    "function_start": f"0x{f_ea:X}",
                    "function_name": idc.get_name(f_ea) or "<unnamed>",
                    "instruction": insn(item),
                })
                funcs[f"0x{f_ea:X}"] = function(f_ea)
    report = {
        "schema": "moo2.ida.re-evidence.v1",
        "evidence_scope": "static_only",
        "mutation": "none",
        "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version(),
                 "script": "tools/ida/audit_capital_morale.py"},
        "input": {"database": ida_nalt.get_input_file_path(), "source": SOURCE,
                  "source_sha256": sha256(SOURCE), "processor": ida_ida.inf_get_procname(),
                  "min_ea": f"0x{ida_ida.inf_get_min_ea():X}",
                  "max_ea": f"0x{ida_ida.inf_get_max_ea():X}"},
        "address_basis": "IDA linear; DOS/4GW LE object #1",
        "matching_names": names,
        "matching_strings": strings,
        "capitol_application_status_hits": capitol_status_hits,
        "roots": {name: function(ea) for name, ea in ROOTS.items()},
        "functions": funcs,
    }
    os.makedirs(os.path.dirname(OUT), exist_ok=True)
    with open(OUT, "w", encoding="utf-8") as f:
        json.dump(report, f, ensure_ascii=False, indent=2)
        f.write("\n")
    ida_pro.qexit(0)


if __name__ == "__main__":
    main()
