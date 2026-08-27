# NPC 對 NPC 外交談判靜態稽核（2026-08-27）

## 證據身分

- 輸入：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- 工具：IDA Pro 9.4／IDAPython，映像 `ida-pro-9.4-idapython:locked-v1`；
  正式 `Orion2.exe.i64` 唯讀掛載後複製到 `/tmp`，未改名或寫回。
- 位址空間：IDA linear EA，DOS/4GW LE object #1。
- 主函式：原名 `sub_2552D @ 0x2552D..0x25AD2`，1445 bytes、371 條指令，
  bytes SHA-256
  `d216cdb84d723af0f28a7048325f742716ce718c5c9eb8dc20bcc138eed6484a`。
- 唯一直接 caller：`sub_252A7` 的 `0x252B6`；該回合外交驅動器依序呼叫
  `sub_53EDB`、`sub_25C7C`、`sub_252D5`、本函式、`sub_25DF1` 與 `sub_2670A`。
- 完整原始指令、caller window、callee、相關 helper、欄位讀寫端與 Hex-Rays
  導覽輸出：[`evidence/npc-treaty-negotiations-ida-20260827.json`](evidence/npc-treaty-negotiations-ida-20260827.json)。

## 已證實：配對與頻率

`0x25536..0x255CE` 掃描所有 ordered pair `(outer, inner)`，排除 self、human marker
`+0x28==100`、淘汰 marker `+0x8B2!=0`，並要求 outer 對 inner 的接觸
`+0x584==1`。每對以 `Random(250-40*difficulty)==1` 才實際談判；difficulty
直接讀 `byte_199CB0`。正式狀態 `+0x627>=4` 或條約記憶 `+0x68F<-30`
時跳過建約分數，但仍執行談判後記憶衰減。

## 已證實：共同基礎分數

`0x255F2..0x256DC` 建立：

```text
base = signed(+0x6D7) + signed(+0x617)
     + 20 if policy == 1
     + 40 if policy == 2
     + 20 if trade +0x62F != 0
     + Σ third party (
           +10 if outer→third policy >=4
           +20 if inner→third policy >=4
            +5 if inner→third +0x71F >0)
```

原始指令在 `0x25642` 與 `0x25666` 都讀同一方向 `+0x62F` 並各加 10；
不可擅自把第二次改解成研究協議。政府分數表 `word_180CCC @ 0x180CCC`
的八個 signed word 為 `[-50,-20,-20,0,20,30,-70,0]`，16 bytes SHA-256
`d045c3a754e49617cce57a4cdbfcd3a1cdf54b955ea21ea36be5b418e6f33f3c`。

## 已證實：條約與協議門檻

1. `0x2577A..0x25866`：
   `treatyScore = base + +0x68F + Random(100) + govOuter + govInner + 50*humanWarCount`。
   分數 `>=200`、目前 policy 1、且兩方沒有 non-human 第三方戰爭時，以
   `sub_5232E(outer,inner,2)` 升同盟；否則分數 `>=100` 且目前非 1／2 時，
   以 policy 1 建互不侵犯。
2. `0x2586B..0x25931`：
   `agreementScore = base + +0x69F + Random(100) + govInner`。
   分數 `>110` 且研究 `+0x637==0` 時優先建立雙向研究協議；否則分數
   `>80` 且貿易 `+0x62F==0` 時建立雙向貿易協議。
3. `sub_5232E @ 0x5232E` 直接雙向寫 `+0x627` 並清 `+0x71F`；policy 1／2
   還依政府表增加另一個 raw 外交欄位。remake 本切片只保存玩家可見的正式
   policy／協議，不把未消費欄位猜成新效果。

## 已證實：納貢要求

`0x25931..0x25A7D` 再算
`base + +0x68F + Random(100) + govOuter + govInner`。分數 `>150`、方向
`+0x63F` 非負、`sub_500CF(outer,inner)<100`、outer `+0xAE` 小於 inner，且
`Random(20)<=difficulty+1` 時，`sub_52049` 把 outer→inner `+0x63F` 設為 2；
接著以 actor=inner、target=outer、reason 14、delta `Random(3)+3` 呼叫
`Change_Relations_`。`sub_500CF` 使用方向性 `+0x5EC` 國力矩陣的
`100*(outer+1)/(inner+1)`，最高 800，outer 每有一場第三方戰爭再除 2。

remake 尚無 `+0x5EC` 原版國力矩陣，現以可觀察 `FleetStrength` 作有明確標註的
資料模型投影；公式、門檻及亂數順序保持原版。這是「強推論投影」，不是
原版國力輸入已證實等價。

## 已證實：談判記憶

- 初始化器 `sub_4D78E` 把 `+0x68F／+0x69F` 設為 0。
- `sub_4DAB2 @ 0x4DAB2` 每回合各加 10，若轉正則夾回 0。
- 每個通過頻率 gate 的 ordered pair 最後三次呼叫 `sub_524C3`；該 helper
  每次把 `+0x68F／+0x69F／+0x6AF／+0x6BF` 各減 10，合計 -30。
- `+0x6D7` 初始化為 0，會由宣戰、解約及外交回應 helper 減少；本切片保存
  方向性 raw 值，但其餘 writer 需隨各玩家可見行為逐條接回。

## Remake 邊界

- **已接**：ordered AI pair、難度頻率、政府表、條約／協議分數、raw 記憶
  回復與 -30、互不侵犯／同盟／貿易／研究、納貢 mode 2 與關係 delta。
- **資料模型投影**：AI 都視為已接觸且存活；國力以 `FleetStrength` 代理；
  `+0x71F` 尚無 writer，維持初始化 0。
- **仍未知／未接**：原版宣戰與停戰建立函式、完整方向接觸、`+0x5EC`
  國力 producer、policy helper 寫入的未消費 raw 欄位，以及玩家／熱座方向矩陣。
- 本頁只證實 1.31 executable；沒有把 1.50 行為外推成相同。
