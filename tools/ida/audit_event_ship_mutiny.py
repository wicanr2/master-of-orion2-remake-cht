"""唯讀匯出事件 13 艦船叛變的候選、record 與艦艇所有權回寫鏈。"""

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

OUT = os.environ.get("MOO2_RE_OUT", "/out/event-ship-mutiny.json")
SOURCE = os.environ.get("MOO2_RE_SOURCE", ida_nalt.get_input_file_path())
ROOTS = {
    "raw_Event_Effect_Consumer": 0x206A2,
    "raw_Event_Record_Maintenance": 0x2027E,
    "raw_Event_State_Dispatcher": 0x21371,
    "raw_Determine_Event": 0x2230A,
    "raw_Event_Target_Relation": 0x23B7D,
    "raw_Event_Ship_Chooser": 0x23CED,
    "raw_Post_Candidate_Helper": 0x23563,
    "raw_Ship_Owner_Transfer_A": 0xAF7B4,
    "raw_Ship_Owner_Write_Helper": 0xE4B5F,
    "raw_Ship_Owner_Transfer_B": 0x100010,
    "raw_Ship_Owner_Transfer_C": 0x10011B,
}


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
    items = list(idautils.FuncItems(f.start_ea))
    return {"requested": f"0x{ea:X}", "start": f"0x{f.start_ea:X}", "end": f"0x{f.end_ea:X}",
            "original_name": idc.get_name(f.start_ea) or "<unnamed>",
            "instructions": [insn(x) for x in items],
            "callers": [insn(x) for x in idautils.CodeRefsTo(f.start_ea, 0)],
            "callees": [insn(x) for x in items if idc.print_insn_mnem(x).lower() == "call"]}


def operand_hits(needles, start=0x10000, end=0x1D5CD0):
    out = []
    for f_ea in idautils.Functions(start, end):
        hits = []
        for ea in idautils.FuncItems(f_ea):
            line = idc.generate_disasm_line(ea, 0) or ""
            if any(needle in line for needle in needles):
                hits.append(insn(ea))
        if hits:
            f = ida_funcs.get_func(f_ea)
            out.append({"function_start": f"0x{f.start_ea:X}",
                        "function_name": idc.get_name(f.start_ea) or "<unnamed>", "hits": hits})
    return out


def data_refs(addresses):
    out = {}
    for address in addresses:
        rows = []
        for ea in idautils.DataRefsTo(address):
            f = ida_funcs.get_func(ea)
            rows.append({"function_start": f"0x{f.start_ea:X}" if f else None,
                         "function_name": idc.get_name(f.start_ea) if f else None,
                         "instruction": insn(ea)})
        out[f"0x{address:X}"] = rows
    return out


def jump_table(name, count):
    ea = idc.get_name_ea_simple(name)
    if ea == idc.BADADDR:
        return {"name": name, "error": "name not found"}
    return {"name": name, "ea": f"0x{ea:X}", "entries": [
        {"case": index, "entry_ea": f"0x{ea + index * 4:X}",
         "raw_dword": f"0x{ida_bytes.get_dword(ea + index * 4):X}"}
        for index in range(count)
    ]}


def local_writes(function_ea, names):
    f = ida_funcs.get_func(function_ea)
    if not f:
        return []
    return [insn(ea) for ea in idautils.FuncItems(f.start_ea)
            if idc.print_insn_mnem(ea).lower() == "mov"
            and any(name in idc.print_operand(ea, 0) for name in names)]


def main():
    ida_auto.auto_wait()
    report = {"schema": "moo2.ida.re-evidence.v1", "evidence_scope": "static_only", "mutation": "none",
              "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version(),
                       "script": "tools/ida/audit_event_ship_mutiny.py"},
              "input": {"database": ida_nalt.get_input_file_path(), "source": SOURCE,
                        "source_sha256": sha256(SOURCE), "processor": ida_ida.inf_get_procname(),
                        "min_ea": f"0x{ida_ida.inf_get_min_ea():X}", "max_ea": f"0x{ida_ida.inf_get_max_ea():X}"},
              "address_basis": "IDA linear; DOS/4GW LE object #1",
              "roots": {name: function(ea) for name, ea in ROOTS.items()},
              "determine_event_jump_table": jump_table("jpt_22836", 29),
              "determine_event_local_writes": local_writes(0x2230A, ["var_4", "var_18"]),
              "ship_owner_status_hits": operand_hits(["+63h]", "+64h]", "+74h]"]),
              "event13_fixed_record_refs": data_refs(range(0x19AC19, 0x19AC22))}
    os.makedirs(os.path.dirname(OUT), exist_ok=True)
    with open(OUT, "w", encoding="utf-8") as f:
        json.dump(report, f, ensure_ascii=False, indent=2)
        f.write("\n")
    ida_pro.qexit(0)


if __name__ == "__main__":
    main()
