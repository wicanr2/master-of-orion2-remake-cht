"""唯讀匯出 MOO2 玩家維護費函式與 caller/callee。

保留原始名稱、IDA 線性位址、bytes 與運算元；不修改或儲存 IDB。
"""
import hashlib, json, os
import ida_auto, ida_bytes, ida_funcs, ida_ida, ida_kernwin, ida_nalt, ida_pro, idautils, idc

OUT = os.environ.get("MOO2_RE_OUT", "/out/player-maintenance.json")
SOURCE = os.environ.get("MOO2_RE_SOURCE", ida_nalt.get_input_file_path())
FUNCTIONS = {
    "raw_Compute_Player_Maintenance": 0xE2000,
    "raw_Player_Maintenance": 0xEE0B0,
    "raw_Next_Turn_Calc": 0x136B3,
    "raw_Officer_Maintenance_Helper": 0x94A9D,
    "raw_Spy_Maintenance_Helper": 0x1026CF,
    "raw_Tribute_Maintenance_Helper": 0xE1FC7,
    "raw_Maintenance_Crisis_Candidate": 0xEDDF7,
    "raw_Maintenance_Crisis_Caller": 0xE4F49,
    "raw_Maintenance_Crisis_Prepare": 0xED908,
    "raw_Maintenance_Crisis_ColonyScore": 0xED9EC,
    "raw_Maintenance_Crisis_Group": 0xEDB35,
    "raw_Maintenance_Crisis_Special": 0xEDCBF,
    "raw_Maintenance_Crisis_Final": 0xEDFE0,
    "raw_Maintenance_Crisis_Filter_A": 0xEDAE2,
    "raw_Maintenance_Crisis_Filter_B": 0xEDB1D,
    "raw_Maintenance_Crisis_Value": 0xEDD73,
    "raw_Maintenance_Crisis_ShipScore": 0x5EF17,
    "raw_Maintenance_Crisis_ShipScore_A": 0x5F871,
    "raw_Maintenance_Crisis_ShipScore_B": 0x5EF4B,
    "raw_Maintenance_Crisis_ColonyWrite": 0xE2D09,
    "raw_Maintenance_Crisis_Notify": 0xEF629,
}

def sha256(path):
    h = hashlib.sha256()
    with open(path, "rb") as f:
        for chunk in iter(lambda: f.read(1024*1024), b""): h.update(chunk)
    return h.hexdigest()

def insn(ea):
    size=idc.get_item_size(ea); raw=ida_bytes.get_bytes(ea,size) or b""
    return {"ea":f"0x{ea:X}","bytes":raw.hex(),"mnem":idc.print_insn_mnem(ea),
            "op0":idc.print_operand(ea,0),"op1":idc.print_operand(ea,1),
            "line":idc.generate_disasm_line(ea,0) or "<unavailable>"}

def function(ea):
    f=ida_funcs.get_func(ea)
    if not f: return {"requested":f"0x{ea:X}","error":"no function"}
    callers=[]
    for ref in idautils.CodeRefsTo(f.start_ea,0):
        cf=ida_funcs.get_func(ref)
        callers.append({"instruction":insn(ref),"function_start":f"0x{cf.start_ea:X}" if cf else None,
                        "function_name":idc.get_name(cf.start_ea) if cf else None})
    callees=[]
    for item in idautils.FuncItems(f.start_ea):
        if idc.print_insn_mnem(item).lower()=="call":
            callees.append({"instruction":insn(item),"targets":[f"0x{x:X}" for x in idautils.CodeRefsFrom(item,0)]})
    return {"start":f"0x{f.start_ea:X}","end":f"0x{f.end_ea:X}",
            "original_name":idc.get_name(f.start_ea) or "<unnamed>",
            "instructions":[insn(x) for x in idautils.FuncItems(f.start_ea)],
            "callers":callers,"callees":callees}

def main():
    ida_auto.auto_wait()
    crisis_filter_ptr = idc.get_wide_dword(0xEDB2D)
    building_table = []
    for building_id in range(49):
        ea = 0x17EB3D + building_id * 0x13
        building_table.append({"id":building_id,"ea":f"0x{ea:X}",
                               "cost":idc.get_wide_dword(ea+8),
                               "maintenance":idc.get_wide_word(ea+12),
                               "raw_category":idc.get_wide_byte(ea+14)})
    report={"schema":"moo2.ida.re-evidence.v1","evidence_scope":"static_only","mutation":"none",
            "tool":{"name":"IDA Pro","version":ida_kernwin.get_kernel_version(),
                    "script":"tools/ida/audit_player_maintenance.py"},
            "input":{"database":ida_nalt.get_input_file_path(),"source":SOURCE,
                     "source_sha256":sha256(SOURCE),"processor":ida_ida.inf_get_procname(),
                     "min_ea":f"0x{ida_ida.inf_get_min_ea():X}","max_ea":f"0x{ida_ida.inf_get_max_ea():X}"},
            "address_basis":"IDA linear; DOS/4GW LE object #1",
            "raw_data":{"off_EDB2D":{"ea":"0xEDB2D","bytes":(ida_bytes.get_bytes(0xEDB2D,16) or b"").hex(),
                                         "first_dword":f"0x{crisis_filter_ptr:X}",
                                         "first_name":idc.get_name(crisis_filter_ptr) or "<unnamed>"},
                        "building_table_17EB3D":building_table,
                        "locret_ED903":{"ea":"0xED903",
                                         "bytes":(ida_bytes.get_bytes(0xED903,5) or b"").hex(),
                                         "u8":[idc.get_wide_byte(0xED903+i) for i in range(5)]}},
            "functions":{name:function(ea) for name,ea in FUNCTIONS.items()}}
    os.makedirs(os.path.dirname(OUT),exist_ok=True)
    with open(OUT,"w",encoding="utf-8") as f: json.dump(report,f,ensure_ascii=False,indent=2); f.write("\n")
    ida_pro.qexit(0)

if __name__=="__main__": main()
