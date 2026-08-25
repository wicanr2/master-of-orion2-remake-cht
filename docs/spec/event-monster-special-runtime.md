# 事件怪物特殊武器規格

## 範圍

本規格只涵蓋事件怪物 weapon ID 44（Plasma Flux）與 ID 45（Caustic Slime）在
remake 兩條戰鬥路徑的玩家可見效果。證據來源見
`docs/re/event-monster-special-runtime-audit-20260825.md`。

## 格子戰術

- ID 44 依 `event-monster-plasma-flux-spread.md` 一次聚合該 mount 全部武器，以 96px
  半徑傷害雙方鄰艦；距離平方衰減後按 `target.SizeClass+1` 逐段擲值，不再把單一傷害
  直接乘尺寸。圈內戰機另依 `event-monster-plasma-flux-fighters.md` 先整隊 50% 迴避，
  未避開再逐架判定傷亡。
- ID 45 每門命中時擲出該 mount 的傷害範圍，累加到
  `CombatShip.CausticSlimeStrength`，命中當下不另外製造第二份直接傷害。
- 每個回合交界，存活艦若黏液強度大於零，依四個護盾朝向各承受一次目前強度；每一面
  護盾不足的溢出量繼續扣裝甲與結構。四面結算後強度減 5，最低為零。
- 黏液狀態是單場戰鬥暫態，不寫回戰略存檔；重複命中採加總。

## 快速結算

- ID 44 的快速結算沒有格位，以中心距離作用於全部玩家 combatant，保留逐尺寸分段；
  缺少原版像素位置的部分明標為可重播近似。
- ID 45 因快速結算沒有逐面護盾與跨回合 `CombatShip`，使用可重播近似：每次命中的
  擲值直接乘四，代表同回合四面包覆傷害；文件必須標成 remake approximation，不能
  宣稱逐回合狀態與原版完全相同。

## 驗收

- 測試 Plasma Flux 的 96px 半徑、距離衰減、雙方鄰艦與尺寸分段亂數。
- 測試 Caustic Slime 堆疊、四面護盾、溢出、每回合減 5 與歸零。
- 測試事件怪物戰術建構仍保留 ID 44／45 mount，且一般武器行為不變。
