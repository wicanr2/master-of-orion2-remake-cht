# CMBTSHP.LBX 戰鬥艦 sprite 結構與艦級對照(2026-07-11)

> 對應 WORKLIST task #12「艦型 sprite 完整對照」。方法:網路查證(社群文件)定結構 +
> 視覺比對(解碼出的 sprite 挑大小)補尺寸層。**不需反組譯、不需 DOSBox oracle**——
> 尺寸層用我方 lbxdump 解出的現成 sprite 目視挑選(rulebook 64 截圖式對照,以自產解碼輸出當 oracle)。

## 1. 結構(社群多方來源確證)

`CMBTSHP.LBX` 共 **360 個資產**(59×60、各 20 幀動畫、無內嵌調色盤)。

- **按「玩家顏色」分組,不是按「種族」**——同一設計圖對應玩家在遊戲裡選的 banner 顏色,
  與所選種族無關。來源:
  - karoltomala(Wolverine)逆向結論,經 Apolyton「The MOO2 LBX format explained」引用:
    "ship design DOESN'T belong to RACE, but to COLOR you choose"。
  - The Spriters Resource 把素材切成 **8 個色系 sheet**:Blue/Brown/Green/Grey/Orange/Purple/Red/Yellow。
  - smuchadissertation「Master of Orion visual style」部落格:"each ship type is limited to the
    faction colour and not the actual faction being played"(佐證)。
- **8 色 × 45 = 360**(算術推導):每個色塊 45 個資產 = **44 張艦艇 sprite(索引 0..43)+ 1 張
  palette-holder 小圖(索引 44)**。palette-holder 是 2×1 內嵌 32 色調色盤的佔位圖,提供該色塊上色用
  (先前不帶 `--pal` 的 lbxdump 只解出這 8 張 = idx 44/89/134/179/224/269/314/359,正是各色塊界)。

色塊 k(k=0..7)的資產配置:
```
sprite 索引  = 45*k + 0 .. 45*k + 43   (44 張艦艇)
palette 索引 = 45*k + 44               (該色塊的調色盤 holder)
```

## 2. 艦級 → sprite 對照(視覺比對)

社群**無人記錄**「同色 45 張裡各艦體尺寸對應哪張」(moo2mod「The Book」/ModdingWiki/LBX 工具皆無)。
本專案用 `lbxdump --pal CMBTSHP.LBX:44 CMBTSHP.LBX <out>`(旗標須在位置參數前,否則 Go flag 不解析)
解出色塊 0 的 44 張,拼成對照網格目視——**索引 0→43 大小單調遞增**(0-10 最小如巡防、33-43 最大填滿
59×60 格)。挑 6 個跨全範圍、逐一放大確認單調變大者當各艦級代表:

| 艦級(remake) | 色塊內索引 | 觀感 |
|---|---|---|
| 巡防艦 Frigate | 3 | 最小 |
| 驅逐艦 Destroyer | 12 | 小 |
| 巡洋艦 Cruiser | 20 | 中 |
| 戰艦 Battleship | 28 | 中大(雙節結構) |
| 泰坦 Titan | 36 | 大 |
| 末日之星 Doom Star | 43 | 最大,填滿格 |

> 這是**近似對照**(把 44 張視為約略尺寸序,取代表值),非「原版設計 picture 欄位的精確映射」——
> 後者仍需實機。但足以讓戰術戰鬥畫面**依艦級顯示不同大小 sprite**,取代先前所有艦共用單一 placeholder
> (CMBTSHP#30)。

## 3. 接線方案

- **玩家艦**用色塊 0(索引 0-43,palette holder 44);**敵艦**用另一色塊(如色塊 1,索引 45-88,
  palette holder 89)以視覺區分敵我。各用自己色塊的 palette-holder 上色。
- 艦級→索引用上表;`CombatShip` 需帶艦級或算好的 sprite 索引(玩家由 `sh.Class` 推,敵艦由
  `genEnemyFleet` 戰力值反推近似艦級)。
- sprite 依 `(色塊, 艦級)` 快取,避免每幀重解。

## 4. 未決 / 誠實標註

- 8 色的實際顏色順序(哪個索引段是 Blue/Red…)未逐一驗證,接線只需「兩個相異色塊」故不影響。
- 艦級→索引為視覺近似;若日後有實機 oracle 可校正為原版精確 picture 映射。
- 44 張的完整語意(是否 6 尺寸×N 設計、各尺寸幾張)未定論(社群軼事「6 設計/色/尺寸」數字對不上
  360,未採信);本文只取「約略尺寸序」這個目視可靠的結論。
