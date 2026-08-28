# Repulsive／Charismatic 原版消費端稽核（2026-08-28）

## 範圍

本切片審查 `player+0x8B2`（Repulsive）與 `player+0x8B3`（Charismatic）。完整 IDA Pro 9.4
指令、bytes、caller 與外部符號在
`evidence/custom-race-trait-consumers-ida-20260828.json`。兩者不是一個可互換的「外交 bonus」：
原版分別在外交選項、關係變化、議會、科技交換、領袖、AI talker 與同化採用不同尺度。

## 可精確量化的規則

| 玩家可見路徑 | Charismatic | Repulsive | 原始位置 |
|---|---:|---:|---|
| 議會候選評分 | `+40` | `-100` | `Vote_Check_ @ 0x160D0／0x160EB` |
| 條約 proposal 檢定 | `+50` | `-50` | `Diplomacy_Test_ @ 0x53365／0x53373` |
| 科技交換 reaction | `+50` | — | `Get_Tech_Exchange_Reaction_ @ 0x26C3C` |
| 科技交換 expectation | `+50` | — | `Get_Tech_Exchange_Expectation_ @ 0x27001` |
| 隨機領袖出現 accumulator | `+5` | `-10` | `Chance_To_Hire_Hero_ @ 0x97899／0x978B4` |
| 領袖候選選取上限前 score | `+10` | `truncTowardZero(score/2)` | `Select_Leader_For_Hire_ @ 0x97B80／0x97B90` |
| AI talker priority | `3` | `1` | `Get_Next_AI_Talker_ @ 0xFA205／0xFA215` |
| 一般 AI talker priority | — | — | raw `2` |
| 同化進度 | `rate×2` | `truncTowardZero(rate/2)` | `Apply_Assimilation_ @ 0xE3516／0xE3523` |

議會完整 score 還含關係、reputation、違約積怨、proposal memory、Orion owner、政策與難度；
本表只列兩個種族項，不能用 `+40/-100` 取代完整 `Vote_Check_`。

## Charismatic 的關係變化非對稱倍率

`Change_Relations_ @ 0x4E3B5` 對事件 actor（`di`）的 Charismatic trait 檢查後：

```text
if delta > 0:
    delta = delta * 2
else:
    delta = truncTowardZero(delta / 2)
```

也就是正向關係變化加倍、零／負向變化減半，不是把目前關係值乘二。此函式有 30 個直接
caller，覆蓋議會結果、外交畫面、NPC 談判／宣戰、事件、首次接觸、封鎖、間諜偷竊與破壞，
所以這是全局關係 delta producer。

## Repulsive 的外交選項與談判 gate

### 玩家／多人外交畫面

- `Diplomacy_Screen_ @ 0x16C4E`：真人或對方任一方 raw `+0x8B2 == 1` 時，改呼叫
  `sub_17227` 的 Repulsive 分支，不走一般選項路徑。
- `Get_Net_Diplomacy_Choices_ @ 0x1DEF8`：多人同步選項使用相同「任一方 Repulsive」gate，
  再呼叫同一 `sub_17227`。因此不是單機 UI 自行隱藏，而是正式網路 choice producer 也受限。

`sub_17227` 的完整 choice ID 表仍由外交選項樹切片判定；本輪只閉合切換條件。

### AI↔AI 與 AI↔真人

- `NPC_To_NPC_Treaty_Negotiations_ @ 0x2552D`：任一帝國 Repulsive 即直接離開整個 treaty
  negotiation 路徑。
- `NPC_To_Human_Diplomacy_ @ 0x26990`：AI 或真人 Repulsive 時跳過兩組一般主動外交候選；
  這兩個 block 位於 IDA 誤併進 `NPC_Diplomacy_ @ 0x252A7` 的後段，外部原始符號可修正 owner，
  但 raw 位址與 IDA owner 均保留。
- `Determine_Diplomacy_Messages_ @ 0x4EB06` 與 `Determine_Bad_Message_ @ 0x4F0DC`：任一方
  Repulsive 會切換到專用訊息／bad-message 分支；精確 message ID 表未在本切片重複展開。
- `Sneak_Attack_Evaluations_ @ 0x544A1`：任一方 Repulsive 會跳過一段一般隨機評分；目標
  Charismatic 另使前段評分 `+10`。最終 sneak-attack score 尚包含軍力、政策與其他欄位。

## 領袖系統

### 招募機率與候選分數

`Random_Officer_Check_ → Chance_To_Hire_Hero_` 的基礎 accumulator 先含距上次領袖的時間，再套
Charismatic `+5` 或 Repulsive `-10`，之後才掃候選領袖。`Generate_Random_Officer_ →
Select_Leader_For_Hire_` 則在亂數／帝國數基礎 score 上套 Charismatic `+10` 或 Repulsive
有號除二，最後 clamp 到 67 再抽選。這兩條是不同階段，不能合成單一「領袖機率倍率」。

### Advanced Civilisation 與 AI 任命

- `Allocate_Adv_Civ_Game_Officers_ @ 0x98489` 有兩個 Repulsive 分支：第一段只對 Repulsive
  帝國進行一組候選處理，第二段明確讓非 Repulsive 進入另一組配置。候選陣列與 officer flag
  的正式語意仍需完整表格切片，故本輪不把它稱為「多／少幾名領袖」。
- `Do_AI_Leaders_ @ 0xD7439` 對 leader raw flag `+0x28` 的 `0x08／0x04` 分支檢查 AI 帝國
  Repulsive 後才繼續；這證明某些領袖相容性／任命規則直接消費 trait，但 flag 名稱尚未閉合。

## AI talker 排序

`Get_Next_AI_Talker_ @ 0xFA1A3` 對每個候選帝國建立 raw priority：Charismatic 3、一般 2、
Repulsive 1，取最高者；另一個 raw `byte_1AAF8B == 40` 的 gate 可讓候選直接設定特殊狀態，
其正式名稱未閉合。三個 caller 都在 `Main_Receive_Message_`，所以這是回合訊息／會談排序，
不是關係分數。

## 開局 AI profile

`Init_NPC_Personalities_Objectives_Themes_ @ 0x589D6` 對 Charismatic raw 1 令一個 profile
accumulator `+100`。Repulsive 沒有在這一段使用相反的 `-100`；它透過前述外交 gate 直接改變
可達路徑。profile accumulator 的正式欄位由 AI profile 母項繼續追查。

## 閉合判定

### 已證實

- 議會 `+40／-100`、proposal `+50／-50`。
- Charismatic 關係正向 delta ×2、負向 delta ÷2（向零截斷）。
- 科技交換兩個 Charismatic `+50`。
- 領袖出現 `+5／-10` 與候選 score `+10／÷2` 是兩個不同階段。
- 同化 `×2／÷2`。
- AI talker priority 3／2／1。
- Repulsive 切換單機／網路外交選項，阻止 AI↔AI treaty negotiation，並改走專用訊息路徑。

### 仍未知

- `sub_17227` 的完整 Repulsive choice ID 表與每項下游。
- Repulsive 專用 diplomacy message／bad-message 完整表。
- Advanced Civilisation officer 候選陣列與 AI leader raw flag 的正式名稱。
- Sneak attack 中被 Repulsive 跳過之 block 的完整最終 score 權重。
- AI profile accumulator 與 raw 40 talker gate 的正式欄位。
- 原版全域 RNG 序列。

因此兩項特性的主要玩家可見公式與 gate 已閉合，但外交選項表、訊息表與領袖 flag 未完成前，
客製種族 parity 仍維持部分閉合。
