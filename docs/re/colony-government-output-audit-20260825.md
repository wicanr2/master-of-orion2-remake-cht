# 殖民地政體／士氣／領袖產出稽核（2026-08-25）

## 證據身分

- 輸入：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- 工具：IDA Pro 9.4／IDAPython；唯讀腳本
  `tools/ida/audit_colony_output_modifiers.py`，並於 2026-08-28 以
  `tools/ida/audit_colony_turn_chain.py` 重新匯出完整函式、兩個直接 helper 與難度表。
- 位址基準：IDA linear，DOS/4GW LE object #1
- 主要函式：`sub_DE280 @ 0xDE280`、`sub_DE22C @ 0xDE22C`、
  `sub_DDF2C @ 0xDDF2C`、`sub_DD9F2 @ 0xDD9F2`
- 可重生證據：
  [`evidence/colony-turn-chain-ida-20260828.json`](evidence/colony-turn-chain-ida-20260828.json)；
  位址、bytes、原始名稱與外部導航名稱分欄保存。

## 已證實

1. `sub_DE280` 逐一掃描 packed colonist，只處理 WORKING 且 job 符合者。每人口先由
   `sub_DE22C` 取得基礎值 `P`，再以 `20P` 為共同基準累加修正，最後工業／研究用
   `(sum+10)/20`，食物用 `(sum+20)/40` 取整。
2. 重力項為 `-5P*(4-gravityCode)`；prisoner 項為 `-5P`。兩者與基礎項在同一整數加總中，
   不是先後相乘的兩次取整。
3. 指派殖民地領袖由 `sub_DD9F2` 取得；農業官／勞工官／科學官依技能 bit 對相應 job 加
   `2*(level+1)P` 或 `3*(level+1)P`，即在共同分母 20 下的 10%／15% 階梯。
4. colony `+7` 的士氣值乘 `P` 併入同一加總。政體 raw 值除 2 等於 3（統一兩級）時完全
   跳過一般士氣項；Android race slot 8 在食物／研究亦跳過士氣。
5. 政體逐職務立即數如下：統一／銀河統一對食物與工業 `+50%／+100%`；封建／邦聯對研究
   `-50%／-25%`；民主／聯邦對研究 `+50%／+75%`。其餘組合為 0。
6. `0xDE28F..0xDE2B6` 由 colony 的 planet 找到 star，再讀 `star+0x2A` 的 owner bit。
   該 bit 已由獨立封鎖 producer／consumer 證據閉合為「此殖民地 owner 正被封鎖」；命中時
   食物與工業各加入 `-10P`，即共同分母 20 下的 `-50%`，研究不受此項影響。
7. 非真人玩家（owner player record `+0x28 != 100`）每名符合職務的人口另加
   `byte_DD4D7[difficulty]`。`byte_DD4D7 @ 0xDD4D7` 的原始 bytes 為
   `F6 00 0A 14 28`，signed 值是 `[-10, 0, 10, 20, 40]`。這是直接加入共同 raw accumulator
   的固定值，**不乘 `P`**；食物仍在全體累加後用 `(sum+20)/40`，工業／研究用
   `(sum+10)/20`，不能改寫成每名人口先取整的百分比。
8. 每名人口的 raw 小計只有在大於 0 時才加入總和；因此極端負修正不會讓單一人口提供負產出。
   `sub_DE22C @ 0xDE22C` 依 job dispatch 基礎 producer，並在非食物或 colony `+0xDD != 0`
   時把非正值的最低基礎抬到 1。該 dispatch 的食物／工業／研究 producer 已由混合種族與
   各產出外層文件分別追蹤，不能把 Hex-Rays 的跨函式 `JUMPOUT` 當作獨立公式。

## 實作勘誤

remake 先前在 `ApplyGovernment` 直接改寫 `FoodPerFarmer`、`IndustryPerWorker`、
`ResearchPerScientist`。此作法會在重複切換政體時累乘，也無法在取得進階政體後從
`+50%` 升為 `+75%` 或 `+100%`。現改為每回合依生效政體重建三個獨立百分比欄位，並與
士氣、重力及領袖百分點在單次整數公式中相加。

## 閉合結論與證據限制

- **已證實**：三職務 dispatch、逐 packed colonist 選取、基礎值、重力、prisoner、士氣、
  統一系士氣豁免、Android 士氣豁免、三種行政領袖技能、八種政體、封鎖、五級 AI 難度表、
  單人口正值 gate 與三職務最終取整均有原始指令與 consumer。
- colony `+7` 已由 consumer、統一系豁免及既有士氣 writer 交叉支持為士氣 raw 刻度；精確欄名
  維持**強推論**，不以語意名稱覆蓋原 operand。remake 的 `MoralePercent` 是玩家可見 adapter。
- optional 598-byte breakdown 只影響殖民地報表的分項顯示，不改變上述總產出。其完整每格 UI
  文案索引尚未逐項命名，但不再列為玩法公式缺口。
- `memset_` 僅清除 local breakdown；屬 C runtime，依專案停止線排除於 RE 與 remake 分母。
