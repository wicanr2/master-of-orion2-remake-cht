"""唯讀匯出 Advanced Civilization 行星選取與回寫完整函式鏈。"""

import csv, hashlib, json, os, traceback
import ida_auto, ida_bytes, ida_funcs, ida_kernwin, ida_name, ida_pro, ida_ua, idautils, idc

ROOTS = {
    "num_planets": 0x62BB7,
    "planet_worth_sort": 0x62BE1,
    "orchestrator": 0x62C70,
    "build_star_list": 0x62E98,
    "build_planet_list": 0x63035,
    "worth_all_planets": 0x63156,
    "planet_worthiness": 0x63259,
    "proximity_bonus": 0x63312,
    "player_has_colony_at_star": 0x633FE,
    "next_planet": 0x6341C,
    "choose_planets": 0x63577,
    "assign_starting_ships": 0x63848,
    "saved_planet_worth": 0x63861,
    "twiddle_selected": 0x638A9,
    "init_homeworlds": 0x63B5A,
    "twiddle_planet": 0x63B8F,
    "player_worth_average": 0x63CA0,
    "write_planet_info": 0x63D0A,
    "modify_homeworlds": 0x7C4AF,
}

def digest(path):
    h = hashlib.sha256()
    with open(path, "rb") as f:
        for block in iter(lambda: f.read(1024 * 1024), b""): h.update(block)
    return h.hexdigest()

def load_symbols(path):
    out = {}
    with open(path, encoding="utf-8", newline="") as f:
        for row in csv.DictReader(f, delimiter="\t"):
            try: out[int(row["ea"], 16)] = row.get("name") or "<unnamed>"
            except (KeyError, TypeError, ValueError): pass
    return out

def instruction(ea):
    ins = ida_ua.insn_t(); ida_ua.decode_insn(ins, ea)
    return {"ea": f"0x{ea:X}", "bytes": (ida_bytes.get_bytes(ea, ins.size) or b"").hex(),
            "text": idc.generate_disasm_line(ea, 0) or ""}

def owner(ea, names):
    fn = ida_funcs.get_func(ea)
    if not fn: return None
    return {"start": f"0x{fn.start_ea:X}", "end": f"0x{fn.end_ea:X}",
            "ida_name": ida_name.get_name(fn.start_ea) or "<unnamed>",
            "external_symbol": names.get(fn.start_ea)}

def window(ea, names, radius=20):
    fn = ida_funcs.get_func(ea)
    if not fn: return {"site": f"0x{ea:X}", "owner": None}
    items = list(idautils.FuncItems(fn.start_ea)); n = items.index(ea)
    return {"site": f"0x{ea:X}", "owner": owner(ea, names),
            "instructions": [instruction(x) for x in items[max(0,n-radius):n+radius+1]]}

def record(ea, names):
    fn = ida_funcs.get_func(ea)
    return {"function": owner(ea, names),
            "chunks": [{"start": f"0x{s:X}", "end": f"0x{e:X}",
                        "instructions": [instruction(x) for x in idautils.Heads(s,e)
                                         if ida_bytes.is_code(ida_bytes.get_flags(x))]}
                       for s,e in idautils.Chunks(fn.start_ea)],
            "callers": [window(x.frm, names) for x in idautils.XrefsTo(fn.start_ea,0) if x.iscode]}

def main():
    ida_auto.auto_wait(); src=os.environ["MOO2_RE_SOURCE"]; db=os.environ["MOO2_RE_DATABASE"]
    syms=os.environ["MOO2_RE_SYMBOLS"]; out=os.environ["MOO2_RE_OUT"]; names=load_symbols(syms)
    payload={"schema":"moo2-advanced-civilization-evidence-v1","inputs":{
        "source":{"name":os.path.basename(src),"sha256":digest(src)},
        "database":{"name":os.path.basename(db),"sha256":digest(db)},
        "symbols":{"name":os.path.basename(syms),"sha256":digest(syms)},
        "ida_version":ida_kernwin.get_kernel_version(),"address_space":"IDA linear address in Orion2.exe.i64"},
        "mutation":"none; read-only export","warning":"外部符號只供導覽；語意須由 raw 資料流分級。",
        "roots":{k:record(v,names) for k,v in ROOTS.items()}}
    with open(out,"w",encoding="utf-8") as f: json.dump(payload,f,ensure_ascii=False,indent=2); f.write("\n")

if __name__ == "__main__":
    try: main()
    except Exception: traceback.print_exc(); raise
    finally: ida_pro.qexit(0)
