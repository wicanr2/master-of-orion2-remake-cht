"""非破壞性索引 player record +0x584..+0x88F 外交／AI 區段的全部直接 operand。"""

import hashlib
import json
import os
import re
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


BLOCK_START = 0x584
BLOCK_END = 0x88F
DISPLACEMENT = re.compile(r"(?P<sign>[+-])(?P<hex>[0-9A-Fa-f]+)h")

ROOTS = {
    "initialize_diplomacy": 0x4D78E,
    "turn_diplomacy_producers": 0x4DAB2,
    "diplomacy_growth": 0x4DD6B,
    "change_relations": 0x4E3B5,
    "determine_messages": 0x4EB06,
    "npc_negotiations": 0x2552D,
    "declare_war": 0x51078,
    "ceasefire": 0x524FB,
    "trade_turn": 0x101E77,
    "create_trade": 0x101EE3,
    "create_research": 0x101F82,
    "compute_contacts": 0xEB192,
    "broken_contacts": 0x52602,
    "set_opportunity_attacks": 0x1FD80,
    "incident_memory": 0x252D5,
    "clear_diplomacy_messages_and_war_timers": 0x5090C,
    "declare_war_tail": 0x5138E,
    "proposal_score_a": 0x5175B,
    "proposal_score_b": 0x519AC,
    "proposal_score_c": 0x51C02,
    "proposal_dispatch": 0x5272D,
    "player_response": 0x533F4,
    "demand_score": 0x539D9,
    "ai_human_dispatch": 0x53EDB,
    "ai_human_target_score": 0x544A1,
    "ai_human_payload_tail": 0x54CC0,
    "npc_give_gift": 0x1AEB5,
    "npc_diplomacy_screen": 0x1AFA6,
    "non_player_proposals": 0x1AC12,
    "diplomacy_payload_codec": 0x1D565,
    "message_payload_prepare": 0x4FE25,
    "message_payload_consume": 0x501CA,
    "war_candidate_scan": 0x25DF1,
    "npc_war_candidate": 0x2670A,
    "npc_response_a": 0x26BBD,
    "npc_response_b": 0x26FBA,
    "npc_response_memory_a": 0x2755F,
    "npc_response_memory_b": 0x276E6,
    "npc_good_proposal_check": 0x2736E,
    "clear_last_proposal_turn": 0x27507,
    "npc_treaty_hatred": 0x277CF,
    "bad_message": 0x4F0DC,
    "sneak_attack_or_declare_war": 0x4F59B,
    "peace_proposal": 0x4F694,
    "sneak_attack_war_check": 0x50827,
    "total_war_loss_comparison": 0x51F0F,
    "war_loss_comparison": 0x51FE5,
    "start_tribute_treaty": 0x52049,
    "start_research_treaty": 0x52150,
    "start_treaty": 0x5232E,
    "gift_response": 0x53723,
    "allocate_ai_spies": 0x100D19,
}

# 附加語意不取代 raw offset；未列出的候選一律保持 unknown，避免直接 xref 掃描把
# 同函式內其他 record 的相同 displacement 誤升格成 player 欄位。
SEMANTIC_INDEX = {
    0x584: ("方向接觸成立旗標", "proven", "Compute_Contacts_／Contact_With_Player_ producer-consumer"),
    0x60F: ("曾建立接觸旗標", "proven", "Init_Diplomatic_Relations_／Compute_Contacts_"),
    0x617: ("目前方向關係", "proven", "Diplomacy_Growth_／Change_Relations_"),
    0x61F: ("關係自然漂移目標", "proven", "Init_Diplomatic_Relations_／Diplomacy_Growth_"),
    0x627: ("正式外交 policy", "proven", "Start_Treaty_／Declare_War_／Ceasefire_"),
    0x62F: ("貿易協議狀態", "proven", "Create_Trade_／Trade_Turn_／Break_Trade_"),
    0x637: ("研究協議狀態", "proven", "Create_Research_／Trade_Turn_／Break_Research_"),
    0x63F: ("方向納貢 raw mode／值", "proven", "Start_Tribute_Treaty_／Break_Tribute_"),
    0x64F: ("pending 外交 reason", "proven", "Change_Relations_／Determine_Diplomacy_Messages_"),
    0x657: ("本回合選定外交訊息 raw ID", "proven", "Determine_Diplomacy_Messages_ family"),
    0x65F: ("pending 最強事件幅度", "proven", "Change_Relations_／Clear_Diplomacy_Messages_"),
    0x66F: ("reason-specific payload B", "proven", "Change_Relations_／message codec"),
    0x67F: ("reason-specific payload A", "proven", "Change_Relations_／message codec"),
    0x68F: ("條約提案持久 modifier", "proven", "NPC treaty score／modifier recovery"),
    0x69F: ("協議提案持久 modifier", "proven", "NPC agreement score／modifier recovery"),
    0x6AF: ("外交 modifier 第三槽", "proven", "technology exchange／proposal／war consumers"),
    0x6BF: ("外交 modifier 第四槽", "proven", "treaty／blockade／proposal／war consumers"),
    0x6CF: ("保存的 bad-message reason 1..9", "proven", "Determine_Bad_Message_／AI-human evaluator"),
    0x6D7: ("方向 reputation", "proven", "war／break／response writers and diplomacy consumers"),
    0x6DF: ("解約通知 raw 類型", "proven", "Break_Treaties_／Break_Trade_／Break_Research_／codec"),
    0x6E7: ("餽贈回應類別 1..4", "proven", "Get_Gift_Response_／codec"),
    0x6EF: ("方向戰爭損失累積", "proven", "battle writer／War_Loss_Comparison_"),
    0x6FF: ("殖民破壞／易手怨值", "proven", "colony damage writers／reason 22"),
    0x70F: ("只有初始化／宣戰／停戰清零，非零語意未知", "unknown", "direct operand inventory"),
    0x717: ("戰爭持續回合", "proven", "Clear_Diplomacy_Messages_／Declare_War_／Ceasefire_"),
    0x71F: ("重複外交事件記憶", "proven", "incident-memory producer and diplomacy consumers"),
    0x727: ("單向永久違約旗標", "proven", "Break_Treaties_ and downstream consumers"),
    0x72F: ("方向外交／停戰冷卻", "proven", "contact／treaty／ceasefire producers and turn decrement"),
    0x737: ("剛清除的 pending magnitude 影子", "proven", "Clear_Diplomacy_Messages_／Diplomacy_Growth_"),
    0x747: ("最近中止協議 raw 類型", "proven", "Break_Treaties_／Break_Trade_／Break_Research_"),
    0x74F: ("上一回合正式 policy 影子", "proven", "End_Of_Turn_Diplomacy_Adjustments_"),
    0x757: ("只有初始化與回合末清除，語意未知", "unknown", "direct operand inventory"),
    0x75F: ("餽贈／需求 payload 類別或 tier", "proven", "Set_Demands_／Npc_Give_Gift_／codec"),
    0x767: ("科技 application payload", "proven", "Set_Demands_／Npc_Give_Gift_／codec"),
    0x76F: ("只有初始化與回合末清除，語意未知", "unknown", "direct operand inventory"),
    0x777: ("BC payload", "proven", "Set_Demands_／Npc_Give_Gift_／codec"),
    0x787: ("第三帝國／關聯帝國 payload", "proven", "alliance proposal／codec"),
    0x78F: ("NPC 科技交換 raw 32-bit payload", "proven", "NPC_Tech_Exchange_Check_／non-player proposal UI"),
    0x7AF: ("科技交換 application ID", "proven", "NPC_Tech_Exchange_Check_／codec"),
    0x7B7: ("殖民地／行星 payload；-1 無", "proven", "Set_Demands_／Npc_Give_Gift_／codec"),
    0x7C7: ("每帝國待執行目標 planet ID", "proven", "AI-human diplomacy and fleet-target consumers"),
    0x7C9: ("每帝國待執行任務／reason raw code", "proven", "AI-human diplomacy and fleet-target consumers"),
    0x7CA: ("每帝國待執行任務目標 owner slot", "proven", "Get_AI_Target_／Launch_Sneak_Attack_"),
    0x7CB: ("每帝國 pending reason 9／請求關聯 player slot", "proven", "AI-human dispatch／bad-message consumer"),
    0x7CC: ("訊息 BC payload shadow", "proven", "message payload staging producer-consumer"),
    0x7DC: ("訊息科技 payload shadow", "proven", "message payload staging producer-consumer"),
    0x7E4: ("訊息關聯第三帝國 slot；-1 無", "proven", "reward／demand／break consumers"),
    0x7EC: ("每帝國食物赤字連續回合；非方向 array", "proven", "turn producer／special-war consumers"),
    0x7EE: ("方向違約積怨", "proven", "Break_Treaties_／council／proposal／AI-human consumers"),
    0x7F6: ("違約積怨對應受害帝國 slot", "proven", "Break_Treaties_／AI-human reason selection"),
    0x7FE: ("Break_Treaties_ 保存的 policy context；下游未知", "strong_inference", "writer proven; independent consumer absent"),
    0x806: ("需求 mode 5 方向禁止計時", "proven", "Get_Demand_Response_／Allocate_AI_Spies_／turn decrement"),
    0x80E: ("需求 mode 8 方向禁止計時", "proven", "demand resolution／economy gates／turn decrement"),
    0x816: ("每帝國 AI-human 新目標決策冷卻", "proven", "turn decrement／AI-human dispatcher"),
    0x817: ("距上次提案累積權重", "proven", "turn increment／Clear_Last_Proposal_Turn_／gift divisor"),
    0x827: ("NPC proposal response memory", "proven", "proposal rejection-accept writers／council consumer"),
    0x837: ("機會攻擊候選 planet ID", "proven", "Set_Opportunity_Attacks_／AI-human consumers"),
    0x847: ("候選敵方殖民地 worth", "proven", "Set_Opportunity_Attacks_／AI-human consumers"),
    0x857: ("候選攻防 pressure", "proven", "Set_Opportunity_Attacks_／AI-human consumers"),
    0x887: ("AI-human worst reason", "proven", "AI-human evaluator／diplomacy-screen consumer"),
    0x88F: ("AI-human 接觸回合", "proven", "turn increment／AI-human dispatcher reset"),
}


def semantic_annotation(offset):
    semantic, level, source = SEMANTIC_INDEX.get(
        offset,
        ("此 direct displacement 尚未證明屬於 player 外交欄", "unknown",
         "candidate from same-function direct operand scan"),
    )
    return {
        "raw_offset": f"0x{offset:X}",
        "semantic": semantic,
        "evidence_level": level,
        "evidence_source": source,
        "warning": None if level == "proven" else "不得作為已證實玩法語意",
    }


def digest(path):
    h = hashlib.sha256()
    with open(path, "rb") as source:
        for block in iter(lambda: source.read(1024 * 1024), b""):
            h.update(block)
    return h.hexdigest()


def insn(ea):
    decoded = ida_ua.insn_t()
    ida_ua.decode_insn(decoded, ea)
    return {
        "ea": f"0x{ea:X}",
        "bytes": (ida_bytes.get_bytes(ea, decoded.size) or b"").hex(),
        "text": idc.generate_disasm_line(ea, 0) or "",
        "mnemonic": ida_ua.print_insn_mnem(ea),
        "operands": [idc.print_operand(ea, i) for i in range(8) if idc.print_operand(ea, i)],
    }


def function_record(ea):
    fn = ida_funcs.get_func(ea)
    if fn is None:
        return {"requested": f"0x{ea:X}", "error": "no function"}
    items = list(idautils.FuncItems(fn.start_ea))
    raw = b"".join(ida_bytes.get_bytes(x, ida_bytes.get_item_size(x)) or b"" for x in items)
    pseudo = None
    if ida_hexrays.init_hexrays_plugin():
        try:
            pseudo = str(ida_hexrays.decompile(fn.start_ea))
        except Exception as error:
            pseudo = f"<decompile failed: {error}>"
    return {
        "requested": f"0x{ea:X}", "start_ea": f"0x{fn.start_ea:X}",
        "end_ea": f"0x{fn.end_ea:X}",
        "original_name": ida_name.get_name(fn.start_ea) or "<unnamed>",
        "bytes_sha256": hashlib.sha256(raw).hexdigest(), "instruction_count": len(items),
        "pseudocode_navigation_only": pseudo,
        "instructions": [insn(x) for x in items],
    }


def direct_sites():
    by_offset = {}
    for fn_ea in idautils.Functions():
        fn = ida_funcs.get_func(fn_ea)
        items = list(idautils.FuncItems(fn_ea))
        # Player-record consumers almost invariably materialize this global base in the same function.
        has_player_base = any("dword_197F98" in (idc.generate_disasm_line(x, 0) or "")
                              or "dword_197FB8" in (idc.generate_disasm_line(x, 0) or "")
                              for x in items)
        if not has_player_base:
            continue
        for index, ea in enumerate(items):
            operands = " ".join(idc.print_operand(ea, i) for i in range(4))
            offsets = set()
            for match in DISPLACEMENT.finditer(operands):
                value = int(match.group("hex"), 16)
                if match.group("sign") == "-":
                    value = -value
                if BLOCK_START <= value <= BLOCK_END:
                    offsets.add(value)
            for offset in offsets:
                by_offset.setdefault(f"0x{offset:X}", []).append({
                    "semantic_annotation": semantic_annotation(offset),
                    "owner_start": f"0x{fn.start_ea:X}",
                    "owner_end": f"0x{fn.end_ea:X}",
                    "owner_original_name": ida_name.get_name(fn.start_ea) or "<unnamed>",
                    "instruction": insn(ea),
                    "window": [insn(x) for x in items[max(0, index - 5):index + 6]],
                })
    return dict(sorted(by_offset.items(), key=lambda item: int(item[0], 16)))


def main():
    ida_auto.auto_wait()
    source = os.environ["MOO2_RE_SOURCE"]
    database = os.environ["MOO2_RE_DATABASE"]
    report = {
        "schema": "moo2.ida.directional-diplomacy-block.v2",
        "evidence_scope": "static_only", "mutation": "none",
        "semantic_status": "reviewed_against_raw_instructions",
        "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version(),
                 "script": "tools/ida/audit_directional_diplomacy_block.py"},
        "input": {"file": os.path.basename(source), "source_sha256": digest(source),
                  "database_sha256": digest(database), "processor": ida_ida.inf_get_procname()},
        "address_basis": "IDA linear; DOS/4GW LE object #1",
        "block": {"start": f"0x{BLOCK_START:X}", "end_inclusive": f"0x{BLOCK_END:X}"},
        "direct_operand_sites": direct_sites(),
        "roots": {name: function_record(ea) for name, ea in ROOTS.items()},
    }
    with open(os.environ["MOO2_RE_OUT"], "w", encoding="utf-8") as target:
        json.dump(report, target, ensure_ascii=False, indent=2)
        target.write("\n")


try:
    main()
except Exception:
    output = os.environ.get("MOO2_RE_OUT", "/tmp/directional-diplomacy-block.json")
    with open(output + ".error", "w", encoding="utf-8") as target:
        target.write(traceback.format_exc())
finally:
    ida_pro.qexit(0)
