"""非破壞性匯出間諜嫁禍第三方的完整靜態鏈。"""

import hashlib
import json
import os
import traceback

import ida_auto
import ida_bytes
import ida_funcs
import ida_ida
import ida_kernwin
import ida_name
import ida_pro
import ida_ua
import idautils
import idc


ROOTS = {
    "frame_another": 0x100BC5,
    "steal_app": 0x10119C,
    "destroy_random_building": 0x10130A,
    "resolve_player_spies": 0x1014A4,
    "change_relations": 0x4E3B5,
    "spy_result_message": 0xFD846,
}


def digest(path):
    h = hashlib.sha256()
    with open(path, "rb") as source:
        for block in iter(lambda: source.read(1024 * 1024), b""):
            h.update(block)
    return h.hexdigest()


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


def function_record(ea):
    fn = ida_funcs.get_func(ea)
    if fn is None:
        return {"requested": f"0x{ea:X}", "error": "no function"}
    items = list(idautils.FuncItems(fn.start_ea))
    raw = b"".join(ida_bytes.get_bytes(x, ida_bytes.get_item_size(x)) or b"" for x in items)
    return {
        "requested": f"0x{ea:X}",
        "start_ea": f"0x{fn.start_ea:X}",
        "end_ea": f"0x{fn.end_ea:X}",
        "original_name": ida_name.get_name(fn.start_ea) or "<unnamed>",
        "bytes_sha256": hashlib.sha256(raw).hexdigest(),
        "instructions": [instruction(x) for x in items],
    }


def callers(ea):
    fn = ida_funcs.get_func(ea)
    if fn is None:
        return []
    rows = []
    for ref in idautils.XrefsTo(fn.start_ea, 0):
        owner = ida_funcs.get_func(ref.frm)
        if owner is None:
            continue
        items = list(idautils.FuncItems(owner.start_ea))
        if ref.frm not in items:
            continue
        index = items.index(ref.frm)
        rows.append({
            "call_ea": f"0x{ref.frm:X}",
            "caller_start": f"0x{owner.start_ea:X}",
            "caller_original_name": ida_name.get_name(owner.start_ea) or "<unnamed>",
            "window": [instruction(x) for x in items[max(0, index - 18):index + 19]],
        })
    return rows


def main():
    ida_auto.auto_wait()
    source = os.environ["MOO2_RE_SOURCE"]
    database = os.environ["MOO2_RE_DATABASE"]
    report = {
        "schema": "moo2.ida.spy-framing.v1",
        "evidence_scope": "static_only",
        "mutation": "none",
        "semantic_status": "reviewed_against_raw_instructions",
        "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version(),
                 "script": "tools/ida/audit_spy_framing.py"},
        "input": {"file": os.path.basename(source), "source_sha256": digest(source),
                  "database_sha256": digest(database), "processor": ida_ida.inf_get_procname()},
        "address_basis": "IDA linear; DOS/4GW LE object #1",
        "roots": {name: {"function": function_record(ea), "callers": callers(ea)}
                  for name, ea in ROOTS.items()},
    }
    with open(os.environ["MOO2_RE_OUT"], "w", encoding="utf-8") as target:
        json.dump(report, target, ensure_ascii=False, indent=2)
        target.write("\n")


try:
    main()
except Exception:
    output = os.environ.get("MOO2_RE_OUT", "/tmp/spy-framing.json")
    with open(output + ".error", "w", encoding="utf-8") as target:
        target.write(traceback.format_exc())
finally:
    ida_pro.qexit(0)
