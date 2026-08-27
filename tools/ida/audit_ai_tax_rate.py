"""非破壞性盤點 MOO2 raw record +0x31 byte 的讀寫端，供 AI 稅率鏈審查。"""

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


RAW_TAX_RATE_OFFSET = 0x31


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
    }


def context(ea, radius=6):
    rows = []
    cursor = ea
    for _ in range(radius):
        cursor = idc.prev_head(cursor)
    for _ in range(radius * 2 + 1):
        if cursor == idc.BADADDR:
            break
        rows.append(instruction(cursor))
        cursor = idc.next_head(cursor)
    return rows


def matching_operands(insn):
    matches = []
    for index, op in enumerate(insn.ops):
        if op.type == ida_ua.o_void:
            break
        if op.type == ida_ua.o_displ and op.addr == RAW_TAX_RATE_OFFSET:
            matches.append(index)
    return matches


def main():
    ida_auto.auto_wait()
    source = os.environ["MOO2_IDA_INPUT"]
    database = os.environ["MOO2_IDA_DATABASE"]
    hits = []
    for segment in idautils.Segments():
        for ea in idautils.Heads(segment, idc.get_segm_end(segment)):
            if not ida_bytes.is_code(ida_bytes.get_full_flags(ea)):
                continue
            insn = ida_ua.insn_t()
            if not ida_ua.decode_insn(insn, ea):
                continue
            operands = matching_operands(insn)
            if not operands:
                continue
            fn = ida_funcs.get_func(ea)
            hits.append({
                "instruction": instruction(ea),
                "matching_operand_indices": operands,
                "original_function_name": ida_name.get_name(fn.start_ea) if fn else "<no-function>",
                "function_start_ea": f"0x{fn.start_ea:X}" if fn else None,
                "context": context(ea),
            })
    report = {
        "schema": "moo2.ida.raw-offset-audit.v1",
        "contract": "raw-location + original-name + bytes; semantics reviewed externally",
        "mutation": "none",
        "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version()},
        "input": {
            "file": os.path.basename(source),
            "source_sha256": digest(source),
            "database_sha256": digest(database),
            "processor": ida_ida.inf_get_procname(),
        },
        "address_basis": "IDA linear; DOS/4GW LE image",
        "raw_offset": f"0x{RAW_TAX_RATE_OFFSET:X}",
        "semantic_status": "unknown_pending_player_base_review",
        "hits": hits,
    }
    with open(os.environ["MOO2_IDA_OUTPUT"], "w", encoding="utf-8") as stream:
        json.dump(report, stream, ensure_ascii=False, indent=2)
        stream.write("\n")


try:
    main()
except Exception:
    error = traceback.format_exc()
    out = os.environ.get("MOO2_IDA_OUTPUT", "/tmp/ai-tax-rate.json")
    with open(out + ".error", "w", encoding="utf-8") as stream:
        stream.write(error)
    ida_kernwin.msg(error + "\n")
    ida_pro.qexit(1)
else:
    ida_pro.qexit(0)
