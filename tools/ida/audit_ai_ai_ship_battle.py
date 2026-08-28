"""非破壞性匯出 AI↔AI 實艦戰鬥、選艦、傷亡與回寫鏈。"""

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
    "search_for_battles": 0xE9D62,
    "pick_star_and_attacker": 0xE84A5,
    "get_ai_target": 0xE7DCA,
    "get_npc_target": 0xE8029,
    "do_one_combat": 0xE938C,
    "get_combat_ships": 0xE6A0C,
    "russ_combat": 0xE7343,
    "strategic_combat": 0x40148,
    "tactical_combat": 0x47939,
    "end_of_tactical_combat": 0x4B184,
    "kill_noncombat_ships": 0xE6B44,
    "determine_retreat_ships": 0xE6CAA,
    "reset_combat_ships_at_star": 0xE9358,
    "process_retreating_ships": 0xE6E52,
    "kill_ship": 0xA163A,
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
        "requested": f"0x{ea:X}", "start_ea": f"0x{fn.start_ea:X}",
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
        "schema": "moo2.ida.ai-ai-ship-battle.v1",
        "evidence_scope": "static_only", "mutation": "none",
        "semantic_status": "reviewed_against_raw_instructions",
        "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version(),
                 "script": "tools/ida/audit_ai_ai_ship_battle.py"},
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
    output = os.environ.get("MOO2_RE_OUT", "/tmp/ai-ai-ship-battle.json")
    with open(output + ".error", "w", encoding="utf-8") as target:
        target.write(traceback.format_exc())
finally:
    ida_pro.qexit(0)
