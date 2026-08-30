"""唯讀匯出 AI trait profile、科技估值與初始母星調整三條窄鏈。"""

import csv
import hashlib
import json
import os
import traceback

import ida_auto
import ida_bytes
import ida_funcs
import ida_kernwin
import ida_name
import ida_pro
import ida_ua
import idautils
import idc


ROOTS = {
    "npc_profiles": 0x589D6,
    "tech_value": 0xFC845,
    "twiddle_initial_homeworlds": 0xE5832,
}

TRAIT_OFFSETS = tuple(range(0x89F, 0x8BE))


def digest(path):
    value = hashlib.sha256()
    with open(path, "rb") as source:
        for block in iter(lambda: source.read(1048576), b""):
            value.update(block)
    return value.hexdigest()


def symbols(path):
    result = {}
    with open(path, encoding="utf-8", newline="") as source:
        for row in csv.DictReader(source, delimiter="\t"):
            try:
                result[int(row["ea"], 16)] = row.get("name") or "<unnamed>"
            except (KeyError, TypeError, ValueError):
                pass
    return result


def instruction(ea):
    decoded = ida_ua.insn_t()
    ida_ua.decode_insn(decoded, ea)
    return {
        "ea": f"0x{ea:X}",
        "bytes": (ida_bytes.get_bytes(ea, decoded.size) or b"").hex(),
        "text": idc.generate_disasm_line(ea, 0) or "",
    }


def owner(ea, names):
    function = ida_funcs.get_func(ea)
    if not function:
        return None
    return {
        "start": f"0x{function.start_ea:X}",
        "end": f"0x{function.end_ea:X}",
        "ida_name": ida_name.get_name(function.start_ea) or "<unnamed>",
        "external_symbol": names.get(function.start_ea),
    }


def window(ea, names, radius=24):
    function = ida_funcs.get_func(ea)
    if not function:
        return {"site": f"0x{ea:X}", "owner": None, "instructions": []}
    items = list(idautils.FuncItems(function.start_ea))
    index = items.index(ea)
    return {
        "site": f"0x{ea:X}",
        "owner": owner(ea, names),
        "instructions": [instruction(item) for item in items[max(0, index-radius):index+radius+1]],
    }


def record(ea, names):
    function = ida_funcs.get_func(ea)
    return {
        "function": owner(ea, names),
        "chunks": [
            {
                "start": f"0x{start:X}",
                "end": f"0x{end:X}",
                "instructions": [
                    instruction(item)
                    for item in idautils.Heads(start, end)
                    if ida_bytes.is_code(ida_bytes.get_flags(item))
                ],
            }
            for start, end in idautils.Chunks(function.start_ea)
        ],
        "callers": [
            window(reference.frm, names)
            for reference in idautils.XrefsTo(function.start_ea, 0)
            if reference.iscode
        ],
    }


def root_trait_sites(root_ea, names):
    function = ida_funcs.get_func(root_ea)
    result = []
    for ea in idautils.FuncItems(function.start_ea):
        text = idc.generate_disasm_line(ea, 0) or ""
        if any(f"+{offset:X}h]" in text for offset in TRAIT_OFFSETS):
            result.append(window(ea, names))
    return result


def main():
    ida_auto.auto_wait()
    source = os.environ["MOO2_RE_SOURCE"]
    database = os.environ["MOO2_RE_DATABASE"]
    symbol_path = os.environ["MOO2_RE_SYMBOLS"]
    output = os.environ["MOO2_RE_OUT"]
    names = symbols(symbol_path)
    payload = {
        "schema": "moo2-ai-trait-profile-homeworld-evidence-v1",
        "inputs": {
            "source": {"name": os.path.basename(source), "sha256": digest(source)},
            "database": {"name": os.path.basename(database), "sha256": digest(database)},
            "symbols": {"name": os.path.basename(symbol_path), "sha256": digest(symbol_path)},
            "ida_version": ida_kernwin.get_kernel_version(),
            "address_space": "IDA linear address in Orion2.exe.i64",
        },
        "mutation": "none; formal database mounted read-only and analyzed from a /tmp copy",
        "warning": "external symbols aid navigation only; semantics require branch and consumer review",
        "roots": {name: record(ea, names) for name, ea in ROOTS.items()},
        "trait_sites": {name: root_trait_sites(ea, names) for name, ea in ROOTS.items()},
    }
    with open(output, "w", encoding="utf-8") as destination:
        json.dump(payload, destination, ensure_ascii=False, indent=2)
        destination.write("\n")


if __name__ == "__main__":
    try:
        main()
    except Exception:
        traceback.print_exc()
        raise
    finally:
        ida_pro.qexit(0)
