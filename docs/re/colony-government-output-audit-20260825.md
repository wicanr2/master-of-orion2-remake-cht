# 殖民地政體／士氣／領袖產出稽核（2026-08-25）

## 證據身分

- 輸入：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- 工具：IDA Pro 9.4／IDAPython；唯讀腳本
  `tools/ida/audit_colony_output_modifiers.py`
- 位址基準：IDA linear，DOS/4GW LE object #1
- 主要函式：`sub_DE280 @ 0xDE280`、`sub_DE22C @ 0xDE22C`、
  `sub_DDF2C @ 0xDDF2C`、`sub_DD9F2 @ 0xDD9F2`

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
6. 非真人玩家另讀難度表 `byte_DD4D7`；本切片只閉合真人與共用公式，不把 AI 難度值猜入
   玩家公式。

## 實作勘誤

remake 先前在 `ApplyGovernment` 直接改寫 `FoodPerFarmer`、`IndustryPerWorker`、
`ResearchPerScientist`。此作法會在重複切換政體時累乘，也無法在取得進階政體後從
`+50%` 升為 `+75%` 或 `+100%`。現改為每回合依生效政體重建三個獨立百分比欄位，並與
士氣、重力及領袖百分點在單次整數公式中相加。

## 證據限制

- 本輪未重新命名 colony `+7` 的 raw 儲存尺度；玩家可見 10% 士氣百分點仍由既有手冊型
  `MoralePercent` adapter 提供。
- AI 的額外 `byte_DD4D7[difficulty]` 仍待 AI 經濟切片閉合，不阻擋玩家政體公式。

