"""非破壞性匯出 AI opportunity attack 多艦隊搜尋垂直鏈。"""

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
    "set_opportunity_attacks": 0x1FD80,
    "find_opportunity_attack": 0xDBC5C,
    "find_opportunity_attack_aux": 0xDBB9F,
    "find_best_attack_from_star": 0xD94B3,
    "gather_player_ships_at_star": 0xD93F8,
    "ai_may_attack_player": 0xD7669,
    "compute_attack_worthiness": 0xD8ED2,
    "player_is_hostile_to_player": 0xD8DE1,
    "enemy_colony_worth_to_player": 0xD8D11,
    "enemy_stars_reachable": 0xD8E52,
    "colony_space_strength_vs_player": 0x5F804,
    "colony_ground_strength_vs_player": 0x5F747,
    "launch_attacks": 0xD97EE,
    "ships_try_to_move_to": 0xFF799,
    "make_ships_move_to": 0xFFD08,
    "weighted_choice_short": 0xFE92D,
}


def digest(path):
    h = hashlib.sha256()
    with open(path, "rb") as source:
        for block in iter(lambda: source.read(1024 * 1024), b""):
            h.update(block)
    return h.hexdigest()


def load_symbols(path):
    result = {}
    with open(path, "r", encoding="utf-8", errors="replace") as source:
        for line in source:
            fields = line.rstrip("\n").split("\t")
            if len(fields) >= 2:
                try:
                    result.setdefault(int(fields[0], 16), []).append(fields[1])
                except ValueError:
                    pass
    return result


def instruction(ea):
    decoded = ida_ua.insn_t()
    ida_ua.decode_insn(decoded, ea)
    return {
        "ea": f"0x{ea:X}",
        "bytes": (ida_bytes.get_bytes(ea, decoded.size) or b"").hex(),
        "text": idc.generate_disasm_line(ea, 0) or "",
        "mnemonic": ida_ua.print_insn_mnem(ea),
        "operands": [idc.print_operand(ea, i) for i in range(8) if idc.print_operand(ea, i)],
        "code_refs": [f"0x{x:X}" for x in idautils.CodeRefsFrom(ea, 0)],
        "data_refs": [f"0x{x:X}" for x in idautils.DataRefsFrom(ea)],
    }


def function_record(ea, symbols):
    fn = ida_funcs.get_func(ea)
    if fn is None:
        return {"requested": f"0x{ea:X}", "error": "no function"}
    items = list(idautils.FuncItems(fn.start_ea))
    raw = b"".join(ida_bytes.get_bytes(x, ida_bytes.get_item_size(x)) or b"" for x in items)
    pseudo = None
    if ida_hexrays.init_hexrays_plugin():
        try:
            pseudo = str(ida_hexrays.decompile(fn.start_ea))
        except Exception as error:
            pseudo = f"<decompile failed: {error}>"
    return {
        "requested": f"0x{ea:X}",
        "start_ea": f"0x{fn.start_ea:X}",
        "end_ea": f"0x{fn.end_ea:X}",
        "original_name": ida_name.get_name(fn.start_ea) or "<unnamed>",
        "external_names_navigation_only": symbols.get(fn.start_ea, []),
        "bytes_sha256": hashlib.sha256(raw).hexdigest(),
        "pseudocode_navigation_only": pseudo,
        "instructions": [instruction(x) for x in items],
    }


def graph(ea, symbols):
    fn = ida_funcs.get_func(ea)
    if fn is None:
        return {"calls": [], "callers": []}
    calls = []
    for item in idautils.FuncItems(fn.start_ea):
        if ida_ua.print_insn_mnem(item) != "call":
            continue
        for target in idautils.CodeRefsFrom(item, 0):
            callee = ida_funcs.get_func(target)
            if callee is not None and callee.start_ea == target:
                calls.append({"call_ea": f"0x{item:X}", "target_ea": f"0x{target:X}",
                              "target_original_name": ida_name.get_name(target) or "<unnamed>",
                              "target_external_names_navigation_only": symbols.get(target, [])})
    callers = []
    for ref in idautils.XrefsTo(fn.start_ea, 0):
        owner = ida_funcs.get_func(ref.frm)
        if owner is not None:
            callers.append({"call_ea": f"0x{ref.frm:X}", "caller_start": f"0x{owner.start_ea:X}",
                            "caller_original_name": ida_name.get_name(owner.start_ea) or "<unnamed>",
                            "caller_external_names_navigation_only": symbols.get(owner.start_ea, [])})
    return {"calls": calls, "callers": callers}


def main():
    ida_auto.auto_wait()
    source = os.environ["MOO2_RE_SOURCE"]
    database = os.environ["MOO2_RE_DATABASE"]
    symbols_path = os.environ["MOO2_RE_SYMBOLS"]
    symbols = load_symbols(symbols_path)
    report = {
        "schema": "moo2.ida.opportunity-attack-search.v1",
        "evidence_scope": "static_only",
        "mutation": "none",
        "semantic_status": "reviewed_against_raw_instructions",
        "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version(),
                 "script": "tools/ida/audit_opportunity_attack_search.py"},
        "input": {"file": os.path.basename(source), "source_sha256": digest(source),
                  "database_sha256": digest(database), "symbols_sha256": digest(symbols_path),
                  "processor": ida_ida.inf_get_procname()},
        "address_basis": "IDA linear; DOS/4GW LE object #1",
        "roots": {name: {"function": function_record(ea, symbols), **graph(ea, symbols)}
                  for name, ea in ROOTS.items()},
    }
    with open(os.environ["MOO2_RE_OUT"], "w", encoding="utf-8") as target:
        json.dump(report, target, ensure_ascii=False, indent=2)
        target.write("\n")


try:
    main()
except Exception:
    output = os.environ.get("MOO2_RE_OUT", "/tmp/opportunity-attack-search.json")
    with open(output + ".error", "w", encoding="utf-8") as target:
        target.write(traceback.format_exc())
finally:
    ida_pro.qexit(0)
