"""非破壞性匯出銀河議會投票與未知分數欄位證據。"""

import csv
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
    "raw_reload_council": 0x156D4,
    "raw_calc_vote_count": 0x15B90,
    "raw_council_finished": 0x15DF8,
    "raw_council_votes": 0x15EBC,
    "raw_vote_check": 0x16021,
    "raw_vote_result": 0x161E4,
    "raw_player_vote": 0x1633C,
    "raw_accept_ruling": 0x1660B,
    "raw_unknown_78398": 0x78398,
    "raw_tech_exchange_reaction": 0x26BBD,
    "raw_good_proposal_check": 0x2736E,
    "raw_message_rejection_check": 0x2755F,
    "raw_proposal_rejection_accept": 0x276E6,
    "raw_treaty_hatred": 0x277CF,
    "raw_init_diplomatic_relations": 0x4D78E,
    "raw_break_treaties": 0x5138E,
    "raw_break_research": 0x5175B,
    "raw_break_trade": 0x519AC,
    "raw_break_tribute": 0x51C02,
}

FIELD_PATTERNS = {
    "raw_current_relation_plus_617": "+617h",
    "raw_treaty_bias_plus_6D7": "+6D7h",
    "raw_grievance_plus_7EE": "+7EEh",
    "raw_vote_modifier_plus_827": "+827h",
    "raw_government_plus_1DF": "+1DFh",
    "raw_previous_vote_plus_E71": "+0E71h",
}


def digest(path):
    value = hashlib.sha256()
    with open(path, "rb") as source:
        for chunk in iter(lambda: source.read(1024 * 1024), b""):
            value.update(chunk)
    return value.hexdigest()


def load_symbols(path):
    result = {}
    if not path:
        return result, None
    with open(path, newline="", encoding="utf-8") as source:
        for row in csv.DictReader(source, delimiter="\t"):
            result.setdefault(int(row["ea"], 16), []).append(row["name"])
    return result, digest(path)


def instruction(ea):
    insn = ida_ua.insn_t()
    size = ida_ua.decode_insn(insn, ea) or 1
    return {
        "ea": f"0x{ea:X}",
        "bytes": (ida_bytes.get_bytes(ea, size) or b"").hex(),
        "text": idc.generate_disasm_line(ea, 0) or "",
        "mnem": idc.print_insn_mnem(ea),
        "op0": idc.print_operand(ea, 0),
        "op1": idc.print_operand(ea, 1),
        "code_refs": [f"0x{x:X}" for x in idautils.CodeRefsFrom(ea, 0)],
        "data_refs": [f"0x{x:X}" for x in idautils.DataRefsFrom(ea)],
    }


def function_record(ea, symbols):
    function = ida_funcs.get_func(ea)
    if function is None:
        return {"requested": f"0x{ea:X}", "error": "function missing"}
    items = list(idautils.FuncItems(function.start_ea))
    raw = ida_bytes.get_bytes(function.start_ea, function.end_ea - function.start_ea) or b""
    pseudo = None
    if ida_hexrays.init_hexrays_plugin():
        try:
            pseudo = str(ida_hexrays.decompile(function.start_ea))
        except Exception as error:
            pseudo = f"<decompile failed: {error}>"
    return {
        "requested": f"0x{ea:X}",
        "start_ea": f"0x{function.start_ea:X}",
        "end_ea": f"0x{function.end_ea:X}",
        "original_name": ida_name.get_name(function.start_ea) or "<unnamed>",
        "external_names_navigation_only": symbols.get(function.start_ea, []),
        "bytes_sha256": hashlib.sha256(raw).hexdigest(),
        "pseudocode_navigation_only": pseudo,
        "instructions": [instruction(item) for item in items],
    }


def callsite_record(ref_ea, symbols):
    owner = ida_funcs.get_func(ref_ea)
    if owner is None:
        return {"call_ea": f"0x{ref_ea:X}", "error": "caller missing"}
    items = list(idautils.FuncItems(owner.start_ea))
    index = items.index(ref_ea)
    return {
        "call_ea": f"0x{ref_ea:X}",
        "caller_start": f"0x{owner.start_ea:X}",
        "caller_original_name": ida_name.get_name(owner.start_ea) or "<unnamed>",
        "caller_external_names_navigation_only": symbols.get(owner.start_ea, []),
        "window": [instruction(x) for x in items[max(0, index - 24):min(len(items), index + 10)]],
    }


def root_callsites(ea, symbols):
    function = ida_funcs.get_func(ea)
    if function is None:
        return []
    return [
        callsite_record(ref.frm, symbols)
        for ref in idautils.XrefsTo(function.start_ea, 0)
        if ida_funcs.get_func(ref.frm) is not None
    ]


def field_sites():
    found = {name: [] for name in FIELD_PATTERNS}
    for function_ea in idautils.Functions():
        function = ida_funcs.get_func(function_ea)
        if function is None:
            continue
        owner = ida_name.get_name(function.start_ea) or "<unnamed>"
        for ea in idautils.FuncItems(function.start_ea):
            operands = f"{idc.print_operand(ea, 0)} {idc.print_operand(ea, 1)}"
            for name, pattern in FIELD_PATTERNS.items():
                if pattern.lower() in operands.lower():
                    row = instruction(ea)
                    row["owner_start"] = f"0x{function.start_ea:X}"
                    row["owner_original_name"] = owner
                    found[name].append(row)
    return found


def main():
    ida_auto.auto_wait()
    source = os.environ["MOO2_RE_SOURCE"]
    database = os.environ["MOO2_RE_DATABASE"]
    symbols, symbols_hash = load_symbols(os.environ.get("MOO2_RE_SYMBOLS", ""))
    report = {
        "schema": "moo2.ida.council-voting.v2",
        "evidence_scope": "static_only",
        "mutation": "none",
        "tool": {
            "name": "IDA Pro",
            "version": ida_kernwin.get_kernel_version(),
            "script": "tools/ida/audit_council_voting.py",
        },
        "input": {
            "file": os.path.basename(source),
            "source_sha256": digest(source),
            "database_sha256": digest(database),
            "symbols_sha256": symbols_hash,
            "processor": ida_ida.inf_get_procname(),
        },
        "address_basis": "IDA linear; DOS/4GW LE object #1",
        "semantic_status": "reviewed_with_evidence_levels",
        "roots": {name: function_record(ea, symbols) for name, ea in ROOTS.items()},
        "root_direct_callsites": {
            name: root_callsites(ea, symbols) for name, ea in ROOTS.items()
        },
        "field_operand_sites": field_sites(),
    }
    with open(os.environ["MOO2_RE_OUT"], "w", encoding="utf-8") as target:
        json.dump(report, target, ensure_ascii=False, indent=2)
        target.write("\n")


try:
    main()
except Exception:
    error = traceback.format_exc()
    output = os.environ.get("MOO2_RE_OUT", "/tmp/council-voting.json")
    with open(output + ".error", "w", encoding="utf-8") as target:
        target.write(error)
    ida_kernwin.msg(error + "\n")
    ida_pro.qexit(1)
else:
    ida_pro.qexit(0)
