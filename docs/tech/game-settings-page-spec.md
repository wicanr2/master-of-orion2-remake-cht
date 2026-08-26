# 對局內 SETTINGS 分頁規格

## 證據與範圍

本規格對應 `sub_7E00F`、`sub_7FA28`、`sub_7EFEF`、`sub_7F14C` 與
`sub_127E1`。詳細位址、雜湊與推論分級見
[`game-menu-popup-ui-text-audit-20260826.md`](../re/game-menu-popup-ui-text-audit-20260826.md)。

## 畫面契約

- SETTINGS 從對局內 GAME 選單進入，使用 `GAME.LBX#29` 背景、`#9` 兩幀開關與
  `#10` ACCEPT；缺資產時必須保持同一熱區與可讀 fallback。
- 固定 13 列、列距 17 邏輯像素。每列整行可點，文字另有 207×16 的雙軸安全框，
  不得越入下一列或開關圖示。
- ACCEPT 才把畫面暫存值套入工作階段並返回星圖；畫面文案只可從
  `assets/i18n/ui.json` 的 `gamesettings.*` 鍵取得。

## 狀態與相容契約

- `shell.GameSettings` 保存 13 個 typed bool 與版本。版本 0 代表舊 JSON，讀取時補
  `sub_127E1` 的原版預設；不得用全 false 猜測舊檔意圖。
- `.GAM` 匯入採其 12 個現有設定欄位；原版 `.GAM` 沒有 Auto Save 欄位，該項採
  `sub_127E1` 預設開啟，證據邊界必須保留。
- `ShowRelocationLines` 在過渡期同步既有相容欄位；新 UI 不直接改動該舊欄位。

## 本切片消費端

- 已接：Show Relocation Lines、Auto Save Game、Animations（結局過場 gate）、
  End Of Turn Summary（一般回合與事件快報後）。
- 其餘開關目前可往返保存，但在相應玩家路徑閉合前不得宣稱玩法效果已完成；它們留在
  WORKLIST 的玩家機制稽核，而不是以「畫面能切換」冒充功能完成。

## 驗收

- 原版預設值測試、13 列雙語 catalog、列框 containment、切換與 JSON 往返。
- Ebitengine 套件測試、正版資產繁中／英文畫廊及缺 `GAME.LBX` fallback 畫廊。
