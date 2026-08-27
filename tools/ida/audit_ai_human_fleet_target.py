"""非破壞性匯出 AI 艦隊目標／抵達鏈與兩層上游 caller。"""

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
    "raw_fleet_target_arrival": 0xDB257,
    "raw_arrival_call_site": 0xDB564,
    "raw_declare_war": 0x51078,
    "raw_prepare_fleet_targets": 0xDB3E5,
    "raw_military_stage_d76b8": 0xD76B8,
    "raw_military_stage_d7764": 0xD7764,
    "raw_military_stage_d97ee": 0xD97EE,
    "raw_military_stage_d9f85": 0xD9F85,
    "raw_military_stage_d896f": 0xD896F,
    "raw_military_stage_da99c": 0xDA99C,
    "raw_military_stage_da485": 0xDA485,
    "raw_military_stage_dafd4": 0xDAFD4,
    "raw_ai_military_outer": 0xDBB29,
    "raw_ai_military_outer_prelude": 0xDB8D8,
    "raw_ai_military_turn_call_site": 0x1371A,
    "raw_target_writer_1afa6": 0x1AFA6,
    "raw_target_writer_4f59b": 0x4F59B,
    "raw_target_writer_50827": 0x50827,
    "raw_target_writer_53edb": 0x53EDB,
    "raw_target_clear_4d78e": 0x4D78E,
    "raw_target_consume_clear_e7dca": 0xE7DCA,
    "raw_human_target_decision": 0x544A1,
    "raw_human_target_commit": 0x54CC0,
    "raw_ai_personality_class": 0x53E96,
    "raw_turn_gate_update": 0x4DAB2,
    "raw_human_target_memory_reset": 0x54D4D,
    "raw_human_target_score_modifier": 0xE5E09,
    "raw_diplomatic_action_availability": 0x4F93B,
}
PERSONALITY_SCORE_TABLE_EA = 0x181080


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
    pseudocode = None
    if ida_hexrays.init_hexrays_plugin():
        try:
            pseudocode = str(ida_hexrays.decompile(fn.start_ea))
        except Exception as exc:
            pseudocode = f"<decompile failed: {exc}>"
    return {
        "requested": f"0x{requested:X}",
        "original_name": ida_name.get_name(fn.start_ea) or "<unnamed>",
        "start_ea": f"0x{fn.start_ea:X}",
        "end_ea": f"0x{fn.end_ea:X}",
        "bytes_sha256": hashlib.sha256(raw).hexdigest(),
        "pseudocode_navigation_only": pseudocode,
        "instructions": [instruction(ea) for ea in idautils.FuncItems(fn.start_ea)],
        "callers": [instruction(x.frm) for x in idautils.XrefsTo(fn.start_ea, 0)
                    if ida_funcs.get_func(x.frm) is not None],
    }


def caller_layers(root, depth):
    layers = []
    frontier = {ida_funcs.get_func(root).start_ea}
    seen = set(frontier)
    for _ in range(depth):
        callers = {}
        for target in frontier:
            for xref in idautils.XrefsTo(target, 0):
                fn = ida_funcs.get_func(xref.frm)
                if fn is None or fn.start_ea in seen:
                    continue
                callers[f"0x{fn.start_ea:X}"] = function_record(fn.start_ea)
                seen.add(fn.start_ea)
        layers.append(callers)
        frontier = {int(ea, 16) for ea in callers}
        if not frontier:
            break
    return layers


def operand_matches(patterns):
    matches = []
    for fn_ea in idautils.Functions():
        fn = ida_funcs.get_func(fn_ea)
        if fn is None:
            continue
        for ea in idautils.FuncItems(fn.start_ea):
            text = idc.generate_disasm_line(ea, 0) or ""
            lowered = text.lower()
            if any(pattern.lower() in lowered for pattern in patterns):
                row = instruction(ea)
                row["function_start_ea"] = f"0x{fn.start_ea:X}"
                row["function_original_name"] = ida_name.get_name(fn.start_ea) or "<unnamed>"
                matches.append(row)
    return matches


def signed_word_table(ea, count):
    values = []
    for index in range(count):
        raw = ida_bytes.get_word(ea + 2 * index)
        values.append(raw - 0x10000 if raw >= 0x8000 else raw)
    return {
        "ea": f"0x{ea:X}",
        "bytes": (ida_bytes.get_bytes(ea, 2 * count) or b"").hex(),
        "values_signed_word": values,
    }


def main():
    ida_auto.auto_wait()
    source = os.environ["MOO2_IDA_INPUT"]
    database = os.environ["MOO2_IDA_DATABASE"]
    source_database = os.environ.get("MOO2_IDA_SOURCE_DATABASE", database)
    report = {
        "schema": "moo2.ida.re-evidence.v1",
        "contract": "raw-location + original-name + bytes + xrefs; semantics reviewed externally",
        "mutation": "none",
        "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version()},
        "input": {
            "file": os.path.basename(source),
            "source_sha256": digest(source),
            "database_sha256": digest(source_database),
            "processor": ida_ida.inf_get_procname(),
        },
        "address_basis": "IDA linear; DOS/4GW LE image",
        "semantic_status": "unknown_pending_review",
        "roots": {name: function_record(ea) for name, ea in ROOTS.items()},
        "raw_fleet_target_arrival_caller_layers": caller_layers(ROOTS["raw_fleet_target_arrival"], 4),
        "raw_ai_record_target_operand_matches": operand_matches(["7C7h", "7CAh"]),
        "raw_human_target_gate_operand_matches": operand_matches(["74Fh", "816h", "88Fh"]),
        "raw_personality_score_table": signed_word_table(PERSONALITY_SCORE_TABLE_EA, 7),
    }
    with open(os.environ["MOO2_IDA_OUTPUT"], "w", encoding="utf-8") as stream:
        json.dump(report, stream, ensure_ascii=False, indent=2)
        stream.write("\n")


try:
    main()
except Exception:
    error = traceback.format_exc()
    out = os.environ.get("MOO2_IDA_OUTPUT", "/tmp/ai-human-fleet-target.json")
    with open(out + ".error", "w", encoding="utf-8") as stream:
        stream.write(error)
    ida_kernwin.msg(error + "\n")
    ida_pro.qexit(1)
else:
    ida_pro.qexit(0)
