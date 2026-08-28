# Spying 種族特性垂直鏈稽核（2026-08-28）

## 證據契約

- 輸入：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- IDA 資料庫：`Orion2.exe.i64`，SHA-256
  `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- 工具：IDA Pro 9.4、`tools/ida/audit_custom_race_trait_consumers.py`；位址均為
  IDA linear、DOS/4GW LE object #1。
- 可重生證據：
  [`evidence/custom-race-trait-consumers-ida-20260828.json`](evidence/custom-race-trait-consumers-ida-20260828.json)
  與 [`evidence/spy-turn-policy-ida-20260828.json`](evidence/spy-turn-policy-ida-20260828.json)。

## signed raw 尺度與攻守表

`player+0x8A8` 是 signed Spying 值；客製種族三檔為 `-10/+10/+20`。`Compute_Spy_Bonuses_
@ 0x100A83` 不做倍率或布林化，直接建立：

```text
base = technologyBonus + signed(Spying) + 10*signed(Telepathic)
attack  = base + bestSpyMaster
defense = base + bestTelepath + governmentDefense[government]
```

因此 Spying 同時改變攻擊與防禦，負值也完整保留；Telepathic 的 `10×raw` 是另一個 trait，
不能套到 Spying。AI 難度不寫進這兩張表，只在 AI 攻方對真人守方的最終任務差值加入
`difficulty-2`。

## Agent pool、訓練與維護

- `player+0xE57+target` 的低六 bit 保存 0..63 人數；self pair 是防守 Agent，其他 pair 是
  外派 Spy，高兩 bit 保存任務。
- 新 spy 產品 raw `-7` 成本 100 industry；完成後先加入 self Agent pool，滿 63 時退回 100
  industry 並重新排入產品。
- `Compute_Player_Maintenance_` 對 self 與所有 target 的人數逐一累加，每名每回合 1 BC。
- `Compute_Needs_ @ 0xCF40D` 在 AI 尚有 Agent 容量且既有戰略需求允許時，直接以
  `Random(100) <= signed(Spying)` 增加一級 spy 生產需求；所以 `+10/+20` 分別提供 10%／20%
  的額外 gate，負值不會在此機率分支產生負需求。函式後段另用 `Spying>0` 選擇積極需求路徑，
  非正值只以 `Random(4)==1` 偶爾保留該需求。

這些 AI 生產 gate 與攻守表是不同 consumer：不能把 `+20` 誤作「直接免費增加兩名間諜」。

## AI 配置與任務

`Resolve_Spies_ @ 0x10192B` 每回合先建立所有活動帝國的 attack／defense 表，再呼叫
`Allocate_AI_Spies_ @ 0x100D19`。配置器會先收回所有外派 Spy：

1. 若尚無任何 spy，只有在已有正常經濟、星曆超過 35030 與難度 gate 通過時才 bootstrap
   一名；正 Spying 額外提供依難度為 `Random(5/6/13)==1` 的啟動機會。
2. attack 分數為負時全部留守；難度 0..2 參考真人與 AI 對手最高防守，難度 3..4 只參考
   真人，保留 Agent 後才分配外派人數。
3. 只攻擊已接觸、仍存續且有可偷科技的帝國；目標權重使用科技價值、目標 Agent、其他帝國
   已派人數及已建立的攻防表。
4. AI 一般選 Espionage；目標 personality raw 3 時有 `Random(8)==1` 改選 Sabotage，從不選 Hide。

## 任務消費端

- Espionage raw 0：淨分至少 80 才偷科技，至少 90 可嫁禍。
- Sabotage raw 1：淨分至少 70 才破壞建築，至少 90 可嫁禍。
- Hide raw 2：不偷竊／破壞，第二次 Spy-vs-Spy 對抗給攻方 +20。
- 第二次對抗至少 80 令守方損失一名 Agent；至多 -80 令攻方損失一名 Spy。

## 閉合結論與 remake 差異

- **已證實**：Spying signed raw 尺度、攻防共同加項、AI 生產需求機率、bootstrap gate、
  packed pool、100 industry 訓練、每名 1 BC、63 上限、AI 留守／外派與三任務主要結果。
- 現行 remake 曾使用「30 BC 直接訓練」「AI 每六回合免費增加一名」及把難度混入共同攻防表；
  均與原版直接資料流不符，應在 RE gate 關閉後進 READY spec 修正。
- AI personality raw 3 的正式名稱與原版 PRNG 逐 bit 不在本切片；數值 gate 與玩家結果已保留。
  C runtime、Watcom helper 與平台 API 不納入玩法分母。
