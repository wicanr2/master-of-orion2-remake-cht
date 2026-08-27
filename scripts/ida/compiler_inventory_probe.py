"""匯出 IDA 已辨識的 compiler/runtime 函式清冊，不修改資料庫。"""

import hashlib
import json
import os
import re

import ida_bytes
import ida_funcs
import ida_ida
import ida_kernwin
import ida_name
import ida_nalt
import idautils


PATTERN = re.compile(
    r"(?:^|_)(?:STK|CHK|GRO|alloca|chkstk|stack|fatal|runtime|Cxx|except|EH|RTC|security|8087)",
    re.IGNORECASE,
)


def main():
    output = os.environ.get("MOO2_IDA_PROBE_OUT", "/tmp/moo2-compiler-inventory.json")
    matches = []
    for ea in idautils.Functions():
        fn = ida_funcs.get_func(ea)
        name = ida_name.get_name(ea)
        if not name or not (PATTERN.search(name) or (fn.flags & ida_funcs.FUNC_LIB)):
            continue
        body = ida_bytes.get_bytes(fn.start_ea, fn.end_ea - fn.start_ea)
        matches.append(
            {
                "start_ea": hex(fn.start_ea),
                "end_ea": hex(fn.end_ea),
                "size": fn.end_ea - fn.start_ea,
                "name": name,
                "is_library": bool(fn.flags & ida_funcs.FUNC_LIB),
                "bytes_sha256": hashlib.sha256(body).hexdigest(),
                "caller_count": sum(1 for _ in idautils.XrefsTo(fn.start_ea, 0)),
            }
        )

    result = {
        "tool": "IDA Pro 9.4 IDAPython",
        "root_filename": ida_nalt.get_root_filename(),
        "input_path": ida_nalt.get_input_file_path(),
        "input_sha256": ida_nalt.retrieve_input_file_sha256().hex(),
        "processor": ida_ida.inf_get_procname(),
        "compiler_id": int(ida_ida.inf_get_cc_id()),
        "address_space": "IDA linear EA",
        "matches": matches,
    }
    with open(output, "w", encoding="utf-8") as stream:
        json.dump(result, stream, ensure_ascii=False, indent=2, sort_keys=True)
        stream.write("\n")
    ida_kernwin.qexit(0)


main()
