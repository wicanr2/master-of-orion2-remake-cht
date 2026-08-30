"""唯讀匯出 Cybernetic／Lithovore 的行星與殖民地 AI worth 下游。"""

import csv, hashlib, json, os, traceback
import ida_auto, ida_bytes, ida_funcs, ida_kernwin, ida_name, ida_pro, ida_ua, idautils, idc

ROOTS={"uncolonized_worth":0xD27A7,"colony_worth":0xD2CAE,"compute_ai_data":0xD3D34}

def digest(path):
    h=hashlib.sha256()
    with open(path,"rb") as f:
        for b in iter(lambda:f.read(1048576),b""): h.update(b)
    return h.hexdigest()

def symbols(path):
    out={}
    with open(path,encoding="utf-8",newline="") as f:
        for r in csv.DictReader(f,delimiter="\t"):
            try: out[int(r["ea"],16)]=r.get("name") or "<unnamed>"
            except (KeyError,TypeError,ValueError): pass
    return out

def insn(ea):
    x=ida_ua.insn_t(); ida_ua.decode_insn(x,ea)
    return {"ea":f"0x{ea:X}","bytes":(ida_bytes.get_bytes(ea,x.size) or b"").hex(),"text":idc.generate_disasm_line(ea,0) or ""}

def owner(ea,names):
    f=ida_funcs.get_func(ea)
    if not f:return None
    return {"start":f"0x{f.start_ea:X}","end":f"0x{f.end_ea:X}","ida_name":ida_name.get_name(f.start_ea) or "<unnamed>","external_symbol":names.get(f.start_ea)}

def window(ea,names,radius=20):
    f=ida_funcs.get_func(ea)
    if not f:return {"site":f"0x{ea:X}","owner":None}
    items=list(idautils.FuncItems(f.start_ea)); n=items.index(ea)
    return {"site":f"0x{ea:X}","owner":owner(ea,names),"instructions":[insn(x) for x in items[max(0,n-radius):n+radius+1]]}

def record(ea,names):
    f=ida_funcs.get_func(ea)
    return {"function":owner(ea,names),"chunks":[{"start":f"0x{s:X}","end":f"0x{e:X}","instructions":[insn(x) for x in idautils.Heads(s,e) if ida_bytes.is_code(ida_bytes.get_flags(x))]} for s,e in idautils.Chunks(f.start_ea)],"callers":[window(x.frm,names) for x in idautils.XrefsTo(f.start_ea,0) if x.iscode]}

def trait_sites(names):
    out=[]
    for seg in idautils.Segments():
        for ea in idautils.Heads(seg,idc.get_segm_end(seg)):
            if not ida_bytes.is_code(ida_bytes.get_flags(ea)):continue
            text=idc.generate_disasm_line(ea,0) or ""
            if "+8B0h]" in text or "+8B1h]" in text: out.append(window(ea,names))
    return out

def main():
    ida_auto.auto_wait(); src=os.environ["MOO2_RE_SOURCE"]; db=os.environ["MOO2_RE_DATABASE"]; syms=os.environ["MOO2_RE_SYMBOLS"]; out=os.environ["MOO2_RE_OUT"]; names=symbols(syms)
    payload={"schema":"moo2-cybernetic-lithovore-worth-evidence-v1","inputs":{"source":{"name":os.path.basename(src),"sha256":digest(src)},"database":{"name":os.path.basename(db),"sha256":digest(db)},"symbols":{"name":os.path.basename(syms),"sha256":digest(syms)},"ida_version":ida_kernwin.get_kernel_version(),"address_space":"IDA linear address in Orion2.exe.i64"},"mutation":"none; read-only export","warning":"trait direct sites are candidates until base stride and branch consumer are reviewed.","roots":{k:record(v,names) for k,v in ROOTS.items()},"trait_direct_sites":trait_sites(names)}
    with open(out,"w",encoding="utf-8") as f:json.dump(payload,f,ensure_ascii=False,indent=2);f.write("\n")

if __name__=="__main__":
    try:main()
    except Exception:traceback.print_exc();raise
    finally:ida_pro.qexit(0)
