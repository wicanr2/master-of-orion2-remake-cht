"""非破壞性匯出 sub_FF593 與 star+0x33／+0x38 mask 的 producer／consumer。"""

import hashlib, json, os, traceback
import ida_auto, ida_bytes, ida_funcs, ida_hexrays, ida_ida, ida_kernwin, ida_name, ida_pro, ida_ua
import idautils, idc

ROOTS = {
    "raw_star_reachability_ff593": 0xFF593,
    "raw_star_distance_ff4e9": 0xFF4E9,
    "raw_compute_blockades_e5097": 0xE5097,
    "raw_mask_writer_12d75": 0x12D75,
    "raw_mask_writer_13a3d": 0x13A3D,
    "raw_mask_writer_79fdb": 0x79FDB,
    "raw_mask_writer_e481f": 0xE481F,
    "raw_mask_writer_e493a": 0xE493A,
    "raw_mask_writer_e5832": 0xE5832,
    "raw_mask_writer_e7cdb": 0xE7CDB,
    "raw_colony_mask_writer_f83d8": 0xF83D8,
    "raw_event_ship_arrival_ffdda": 0xFFDDA,
    "raw_mask_fill_8d65d": 0x8D65D,
    "raw_mask_clear_8d8a1": 0x8D8A1,
}

def digest(path):
    h=hashlib.sha256()
    with open(path,"rb") as f:
        for c in iter(lambda:f.read(1<<20),b""): h.update(c)
    return h.hexdigest()

def ins(ea):
    i=ida_ua.insn_t(); n=ida_ua.decode_insn(i,ea) or 1
    return {"ea":f"0x{ea:X}","bytes":(ida_bytes.get_bytes(ea,n) or b"").hex(),
            "text":idc.generate_disasm_line(ea,0) or "",
            "code_refs":[f"0x{x:X}" for x in idautils.CodeRefsFrom(ea,0)],
            "data_refs":[f"0x{x:X}" for x in idautils.DataRefsFrom(ea)]}

def fnrec(ea):
    f=ida_funcs.get_func(ea)
    if not f: return {"requested":f"0x{ea:X}","error":"missing"}
    raw=ida_bytes.get_bytes(f.start_ea,f.end_ea-f.start_ea) or b""
    pseudo=None
    if ida_hexrays.init_hexrays_plugin():
        try: pseudo=str(ida_hexrays.decompile(f.start_ea))
        except Exception as e: pseudo=f"<decompile failed: {e}>"
    return {"requested":f"0x{ea:X}","original_name":ida_name.get_name(f.start_ea) or "<unnamed>",
            "start_ea":f"0x{f.start_ea:X}","end_ea":f"0x{f.end_ea:X}",
            "bytes_sha256":hashlib.sha256(raw).hexdigest(),"pseudocode_navigation_only":pseudo,
            "instructions":[ins(x) for x in idautils.FuncItems(f.start_ea)]}

def candidate_functions():
    out={}
    for fea in idautils.Functions():
        texts=[idc.generate_disasm_line(x,0) or "" for x in idautils.FuncItems(fea)]
        low="\n".join(texts).lower()
        if ("+33h]" not in low and "+38h]" not in low) or ("dword_19306c" not in low and "71h" not in low):
            continue
        hits=[]
        for x,t in zip(idautils.FuncItems(fea),texts):
            if "+33h]" in t.lower() or "+38h]" in t.lower() or "71h" in t.lower(): hits.append(ins(x))
        out[f"0x{fea:X}"]={"original_name":ida_name.get_name(fea) or "<unnamed>","hits":hits}
    return out

def main():
    ida_auto.auto_wait(); src=os.environ["MOO2_IDA_INPUT"]; db=os.environ["MOO2_IDA_SOURCE_DATABASE"]
    report={"schema":"moo2.ida.re-evidence.v1","contract":"raw location/name/bytes/xrefs; semantics reviewed externally",
            "mutation":"none","tool":{"name":"IDA Pro","version":ida_kernwin.get_kernel_version()},
            "input":{"file":os.path.basename(src),"source_sha256":digest(src),"database_sha256":digest(db),
                     "processor":ida_ida.inf_get_procname()},"address_basis":"IDA linear; DOS/4GW LE image",
            "semantic_status":"unknown_pending_review","roots":{k:fnrec(v) for k,v in ROOTS.items()},
            "star_mask_candidate_functions":candidate_functions()}
    with open(os.environ["MOO2_IDA_OUTPUT"],"w",encoding="utf-8") as f: json.dump(report,f,ensure_ascii=False,indent=2); f.write("\n")

try: main()
except Exception:
    p=os.environ.get("MOO2_IDA_OUTPUT","/tmp/star-mask.json"); open(p+".error","w").write(traceback.format_exc()); ida_pro.qexit(1)
else: ida_pro.qexit(0)
