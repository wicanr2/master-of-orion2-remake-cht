# 自動艦艇設計規格

## 目標

在目前單槽編輯器中，實作可回查 `Auto_Design_Ship_ @ 0x616A5` 的自動初始裝備；產出會寫入
六艦體持久設計庫，不以殖民地 `AUTO BUILD` 或 AI 抽象軍力冒充 blueprint。

## 輸入與輸出

輸入：

- `GameSession` 的玩家科技與版本規則
- 艦體名稱
- 原版 raw design type 0..7

輸出 `AutoDesignLoadout`：武器、裝甲、護盾、特殊裝置索引，武器改造、火線角、彈架容量與
原始 role。

## 規則

1. 只可選已解鎖元件。
2. raw 2／3 只選飛彈家族；raw 1／4 的特殊槽只選三種機庫；raw 5 優先球形／特殊武器；
   其餘不選炸彈作單一艦對艦武器。
3. 裝甲與護盾選目前最高已解鎖項。
4. 候選由高至低嘗試，最終必須通過 `DesignFitsWithLoadout`；放不下時依序降低特殊裝置與武器。
5. 火線角與彈架容量經現有 `DefaultWeaponArc`／`NormalizeWeaponAmmo` 正規化。
6. 舊存檔缺設計庫時，六種艦體各套用一次 mixed role 模板；之後畫面重建不得覆寫玩家修改。

## 明示近似

- 原版八武器槽／八特殊槽已可保存、編輯、建造並由快速／格子戰術消費；格子戰術另有逐武器槽可用／待命／關閉與右鍵資訊。逐槽點防會在飛彈／戰機接觸前自動開火，且刻意忽略紅色關閉狀態。一般 AI 已保存六筆 blueprint 與實艦，詳見 `ai-ship-blueprints-and-build.md`。
- raw type 的產生 RNG 與 player `+0x205` enum 尚未映射到 remake；UI 初始值固定使用 raw 0。
- 引擎、電腦與燃料尚不是 remake 可選欄位，不能保存原版 `+0x12..+0x16` 全部裝備。
- AI 的 `FleetStrength` 只由實艦艦體強度推導；藍圖、造艦與兩條戰鬥路徑直接消費逐艦資料，不把 loadout 人為換算成軍力。

## 驗收

- mixed role 產生已解鎖且不超出巡洋艦空間的 loadout。
- missile role 不選 beam；fighter role 不選非機庫特殊裝置。
- 六艦體設計可獨立修改、存檔與切換，不被畫面重建覆寫。
- 存量完整 `go test ./...` 通過。
