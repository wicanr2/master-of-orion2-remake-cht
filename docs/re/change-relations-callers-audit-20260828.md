# `Change_Relations_` caller／reason 垂直鏈稽核

日期：2026-08-28

## 證據契約

- 輸入 `Orion2.exe` SHA-256：
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- 正式 `Orion2.exe.i64` SHA-256：
  `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- 外部符號表 SHA-256：
  `f83116049964c14072547dd087a6316117cf62808c328cc7c087984467030b28`。
- 工具：IDA Pro 9.4／IDAPython；DOS/4GW LE object #1，位址均為 IDA linear EA。
  正式資料庫唯讀掛載，容器只分析 `/tmp` 副本。
- 非破壞性匯出：
  [`evidence/change-relations-callers-ida-20260828.json`](evidence/change-relations-callers-ida-20260828.json)。
  證據保存四個 raw 函式、完整指令與 bytes hash、30 個直接 callsite、17 個 caller 本體，
  以及四個根函式各自的直接 xref；外部名稱只供導覽。
- Hex-Rays 偽碼的參數與型別不是證據。`Reward_Check_` 的偽碼只保留可達的固定 `200`；
  原始指令另含一段不可達的隨機分支。下文以 `0x4EA4E..0x4EA66` 與
  `0x4EA6B..0x4EAED` 的控制流為準。

## 呼叫契約與共用寫回（已證實）

`Change_Relations_ @ 0x4E3B5..0x4EA03` 的 caller 契約是：`EAX=raw delta`、
`EDX=actor/source`、`EBX=target`、`ECX=reason`，stack 由先後兩次 push 傳入
`payload A`、`payload B`。兩個 payload 的玩家語意依 reason 而異，不應建立一個全域欄位名。

函式先排除越界／死亡／自己／未接觸／policy 6 等不可達配對。`delta=-10000` 是通知 sentinel：
它可寫入 pending reason 與 payload，但不改關係分數。一般 delta 依目前關係、actor 政體、
target Charismatic、AI／AI 百回合後的難度分支縮放，寫入 raw signed relation `+0x617`，
夾在 `-100..100`，再鏡射關係分數。反向正式 policy 4／5 對 reason 1..9 的負面事件另有
戰爭早退／真人 quarter 分支；因此 caller 的基礎 delta 不能直接當成最終變動。

只有新事件的絕對幅度大於既有 `+0x65F` 時，才更新方向 pending record：

| 原始偏移 | 已證實用途 |
|---:|---|
| `+0x64F` | pending reason |
| `+0x657` | 本回合選出的外交訊息 ID |
| `+0x65F` | pending／最強事件幅度 |
| `+0x66F` | payload B |
| `+0x67F` | payload A |
| `+0x6CF` | `Determine_Bad_Message_` 保存的 reason 1..9 |
| `+0x71F` | `sub_252D5` 由 pending 事件演化出的重複事件記憶 |

## 30 個直接 callsite 分類（已證實）

| caller／callsite | 基礎 delta | reason／payload | 玩家可見來源或消費 |
|---|---:|---|---|
| `Council_Vote_Result_ @ 0x161E4`，`0x16220／0x16239` | `-6` | `0／0／0` | 議會無勝者時，投票者對兩名候選的關係變化 |
| 同上，`0x1628D／0x162F5` | `+24` | `16／0／0` | 投票者支持勝者 |
| 同上，`0x162A9／0x16317` | `-12` | `9／0／0` | 投票者反對另一候選；reason 9 進負面訊息鏈 |
| `Npc_Diplomacy_Screen_ @ 0x1AFA6`，`0x1AA9E` | `-50` | `0／0／0` | NPC 外交畫面的 raw message `0x25` 分支 |
| `Accepted_Npc_Demand_Resolution_ @ 0x1ABBE`，`0x1ABE2` | `Random(13)+11 = 12..24` | `0／0／0` | 玩家接受 NPC demand；之後雙向清 `+0x71F` |
| `Diplomacy_Exchange_Technology_ @ 0x1CB4D`，`0x1CF2D` | `+12` | `0／0／0` | 雙方取得交換科技前的關係獎勵 |
| `Event_Twiddle_ @ 0x206A2`，共用 `0x209E7` | case 4 `-100`；case 5 `+100` | `0／0／0` | 隨機外交暗殺／聯姻；case 5 由 `0x20A1A` 回跳同一 callsite，對稱戰爭狀態會在共用函式早退 |
| `NPC_To_NPC_Treaty_Negotiations_ @ 0x2552D`，`0x25A7D` | `Random(3)+3 = 4..6` | `14／0／0` | AI 納貢成立後的正面關係事件 |
| `Honor_Alliances_ @ 0x25AD2`，`0x25B7E` | `-10000` | `19／0／0` | 真人捲入同盟戰爭前的通知 sentinel |
| `NPC_Diplomacy_Buildup_Penalties_ @ 0x25C7C`，`0x25DCB` | 上游軍力差計算值 | `8／0／0` | AI 軍備壓力負面事件 |
| `NPC_Declarations_Of_War_ @ 0x25DF1`，`0x265E4／0x266CE` | `-10000` | 動態 reason／`0／0` | 真人收到 AI 特殊宣戰理由；隨後呼叫宣戰 writer |
| `Diplomacy_Growth_ @ 0x4DD6B`，`0x4DDF8` | `+1` | `0／0／0` | 停戰／冷卻方向的緩和分支 |
| 同上，`0x4DE43` | `Random(3)=1..3` | `0／0／0` | 互不侵犯協議 |
| 同上，`0x4DE93` | `Random(3)=1..3` | `12／0／0` | 貿易協議 |
| 同上，`0x4DEE3` | `Random(3)=1..3` | `13／0／0` | 研究協議 |
| 同上，`0x4DF11` | `Random(5)=1..5` | `0／0／0` | 同盟 |
| 同上，`0x4DF64` | `Random(3)=1..3` | `14／0／0` | 納貢 mode 1 |
| 同上，`0x4DFB7` | `Random(8)=1..8` | `14／0／0` | 納貢 mode 2 |
| 同上，`0x4E0F1` | `-(leaderSkill-4 + strength/50)` | `5／leader index／0` | 真人對 AI 的高技能領袖／國力負面事件 |
| `Reward_Check_ @ 0x4EA03`，共用尾端 `0x4EAED` | `+200` | `15／caller payload／已連結帝國` | 第三方的 `+0x7E4` 已指向被獎勵帝國；這是唯一可達入口 |
| 同上，`0x4EA6B..0x4EAED` dead block | `Random(15)+15 = 16..30` | `15／caller payload／另一帝國` | `0x4EAAC cmp ax,ax` 後 `0x4EAAF jle 0x4EAF2` 恆成立，故在本 binary 不可達，不算玩法公式 |
| `Determine_First_Contacts_ @ 0x50B57`，`0x50D86` | `-10000` | `18／0／0` | 重新接觸／首次接觸通知 sentinel 的其中一條分支 |
| `Compute_Blockades_ @ 0xE5097`，`0xE5253` | `-Random(5)=-1..-5` | `7／star index／0` | 封鎖積怨與星系 payload |
| `Main_Receive_Message_ @ 0xF5A9F`，`0xF675B` | 封包值 | 六個參數皆由封包傳入 | 多人同步 transport consumer；不是獨立玩法 producer |
| `Russ_Change_Relations_ @ 0xF975C`，`0xF97C3` | wrapper 傳入值 | 六個參數透傳 | 多人同步 sender／本機 wrapper；不是獨立玩法 producer |
| `Steal_App_ @ 0x10119C`，`0x1012B0` | `-(Random(15)+Random(5)) = -2..-20` | `1` 或 `2`／`0`／tech ID | 科技竊取；reason 2 是嫁禍第三方 |
| `Destroy_Random_Building_ @ 0x10130A`，`0x101406` | `-(Random(15)+Random(5)) = -2..-20` | `3` 或 `4`／colony ID／building slot | 破壞建築；reason 4 是嫁禍第三方 |

表中的亂數範圍使用本專案已獨立證實的 `Random(n)=1..n`。這訂正舊筆記把
`Random(13)+11` 寫成 11..23、`Random(3)+3` 寫成 3..5 的下界錯誤；
`Random(15)+15` 即使算式範圍為 16..30，在 1.31 中仍是不可達 dead code。

## reason producer → message consumer（已證實）

`Determine_Diplomacy_Messages_ @ 0x4EB06..0x4F0DC` 每回合從 `+0x64F／+0x65F`
選擇 pending 事件：reason 1..9 交給
`Determine_Bad_Message_ @ 0x4F0DC..0x4F59B`；reason 10／11 走獎勵、政府與宣戰分支；
reason 12 以上走正面／通知 message dispatcher。已直接看見的固定分支包括
17（政府相關訊息）、18（首次／重新接觸）、19（同盟戰爭通知）、20（message 108）、
21（82）、22（80）、23（111）。reason 1..9 會保存到 `+0x6CF`，並依政府、正式狀態、
關係、抱怨次數與亂數選擇警告、要求、解約或宣戰；這是 pending record 的玩家可見下游，
不是單純的關係分數欄位。

`Change_Relations_` 只有在 reason 10 時直接呼叫 `Reward_Check_ @ 0x4EA03`；該 helper 的
唯一 caller xref 是 `0x4E483`。`Determine_Diplomacy_Messages_` 的唯一直接 caller 是
`Next_Turn_Calc_ @ 0x137B0`，而 `Determine_Bad_Message_` 的唯一 caller 是前者 `0x4ED02`，
因此 producer、每回合 consumer 與玩家訊息鏈已閉合。

## 證據等級與剩餘邊界

- **已證實**：完整函式本體、30 個直接 callsite、17 個 caller、呼叫暫存器／stack 契約、
  caller 的 delta／reason／payload、sentinel、pending record、每回合訊息 consumer，以及
  `Reward_Check_` 隨機區塊不可達。
- **已證實但不是玩法 producer**：兩個 multiplayer wrapper 只傳輸同一六參數契約；
  它們不應在玩法 RE 分母重複計數。
- **強推論**：外部符號 `Reward_Check_`、`Steal_App_` 等名稱只用於導覽；語意仍由 raw caller、
  欄位與 consumer 支撐。
- **未知但不阻塞本列 raw gameplay**：JIMTEXT 中每個 reason／message ID 的正式英文句子、
  reason 6 是否只可能由動態／網路輸入，以及 1.50 二進位是否改動表值。這些不能反向否定
  已閉合的 1.31 producer／consumer，但正式文案與版本 profile 仍須分列追查。
- **remake 明確不對齊**：現行關係資料模型、一般外交線性變動、嫁禍第三方、完整 pending
  message payload 與部分方向 policy 尚未承接上述原版鏈。依 RE-first gate，本輪只登記差異，
  不撰寫 READY spec，也不修改玩法程式。
