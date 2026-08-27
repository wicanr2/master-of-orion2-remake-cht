# 戰機駐防戰略強度規格

狀態：CONFORMED（2026-08-27）

## 範圍

以 `docs/re/fighter-garrison-strength-audit-20260827.md` 的已證實資料流，取代 remake
殖民地反擊路徑中「中隊數 × 泛用戰機近似攻擊」的舊模型。

## 契約

- 純規則函式接收最高戰機檔、最佳合格 beam 最大傷害、最佳 bomb 最大傷害、最佳裝甲減傷，
  回傳原版戰略強度。
- 三檔權重固定為 `(40,0)`、`(40,24)`、`(32,24)`。
- 每種武器先各自扣裝甲且不得低於零，再乘權重、相加、整數除二，最後上限 64000。
- beam 只可從原版 raw fighter-eligible 的玩家武器 ID `1,3,4,5,9` 選最大已知傷害；
  無已知項時以雷射 ID 3 為基礎值。
- bomb 從原版 category 3 選最大已知傷害；無已知項時以核彈 ID 21 為基礎值。
- 裝甲減傷依鈦／三鈦／佐特／中子素／精金／氙素階梯使用 `0/1/3/5/7/10`。
- 殖民地反擊將此強度作為單一抽象防禦 combatant 的攻擊值；HP 仍屬 remake 戰鬥容器，
  不冒稱原版逐架模型。

## 驗收

- 純規則表格測試涵蓋三檔、裝甲扣減、負值夾零、整數除二與 64000 上限。
- shell 測試涵蓋起始 fallback、解鎖合格 beam／bomb、最佳裝甲，以及科技升級後的實際反擊值。
- 原有殖民地轟炸抽樣測試保持通過。

## 實作與驗證

- `internal/gamedata/planet_defense.go`：原版權重、扣甲、上限、eligible ID 與裝甲 raw 表。
- `internal/shell/orbital_bombardment.go`：玩家科技選擇器與殖民地反擊消費端。
- `go test ./internal/gamedata ./internal/shell` 通過；另以 `FighterGarrison` 聚焦案例重跑通過。
