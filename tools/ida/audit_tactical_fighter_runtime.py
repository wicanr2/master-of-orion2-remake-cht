"""非破壞性匯出戰機攻擊分流；不對原始函式改名。"""

import hashlib, json, os, traceback
import ida_auto, ida_bytes, ida_funcs, ida_ida, ida_kernwin, ida_lines, ida_name, ida_pro, ida_ua, idautils, idc

ROOTS = [0x3AC20, 0x3AD57, 0x3D2DF]

def digest(path):
    h = hashlib.sha256()
    with open(path, "rb") as fh:
        for chunk in iter(lambda: fh.read(1024 * 1024), b""):
            h.update(chunk)
    return h.hexdigest()

def insn(ea):
    i = ida_ua.insn_t(); n = ida_ua.decode_insn(i, ea)
    if n <= 0: n = max(1, idc.get_item_size(ea))
    return {"ea": f"0x{ea:X}", "bytes": (ida_bytes.get_bytes(ea, n) or b"").hex(" "),
            "text": ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or "")}

def function(ea):
    fn = ida_funcs.get_func(ea)
    if not fn: return None
    calls = []
    for p in idautils.FuncItems(fn.start_ea):
        if idc.print_insn_mnem(p).lower().startswith("call"):
            calls.append({"site": insn(p), "targets": [{"ea": f"0x{x:X}", "raw_name": ida_name.get_name(x)} for x in idautils.CodeRefsFrom(p, 0)]})
    return {"start": f"0x{fn.start_ea:X}", "end_exclusive": f"0x{fn.end_ea:X}",
            "raw_name": ida_name.get_name(fn.start_ea), "direct_calls": calls,
            "instructions": [insn(p) for p in idautils.FuncItems(fn.start_ea)]}

def main():
    ida_auto.auto_wait(); src=os.environ["MOO2_IDA_INPUT"]; db=os.environ["MOO2_IDA_DATABASE"]
    out={"schema":"moo2.ida.re-evidence.v1","mutation":"none","tool":{"name":"IDA Pro","version":ida_kernwin.get_kernel_version()},
         "input":{"file":os.path.basename(src),"source_sha256":digest(src),"database_sha256":digest(db),"processor":ida_ida.inf_get_procname()},
         "address_basis":"IDA linear; DOS/4GW LE image","semantic_status":"unknown_pending_review",
         "roots":[{"requested":f"0x{x:X}","record":function(x)} for x in ROOTS]}
    with open(os.environ["MOO2_IDA_OUTPUT"],"w",encoding="utf-8") as fh: json.dump(out,fh,ensure_ascii=False,indent=2); fh.write("\n")

try: main()
except Exception:
    with open(os.environ.get("MOO2_IDA_OUTPUT","/tmp/fighter.json")+".error","w") as fh: fh.write(traceback.format_exc())
    ida_pro.qexit(1)
else: ida_pro.qexit(0)
