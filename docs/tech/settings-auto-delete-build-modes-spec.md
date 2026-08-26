# SETTINGS：自動清除阻塞建造模式規格

## 玩家契約

玩家在建造佇列按 `OK` 時，若 Housing 或 Trade Goods 位於其他非空項之前：

- 設定關閉：顯示原版語意的確認框；選「否」留在佇列，選「是」清除阻塞模式並返回殖民地。
- 設定開啟：不顯示確認框，直接清除阻塞模式並返回殖民地。
- `CANCEL` 不觸發這條掃描。
- 位於最後的 Housing／Trade Goods 是有效持續模式，不得刪除。

## 規則層

`BlockingBuildMode(colony)` 以 `Builds[colony] + BuildQueue[colony]` 的七格順序找第一個
Housing／Trade Goods；只有後方還有非空項才回傳。`DeleteBlockingBuildModes` 重複刪除第一個
阻塞模式，直到沒有模式位於其他項之前，並經 `DequeueBuild` 記錄多人鎖步命令。

原版 raw `-10` 是 `^ Repeat ^` 佇列標記；remake 以獨立 `RepeatBuild[]` 保存重複目標，沒有
等價 raw 格可刪。不得為追求表面一致而同時清空 typed Repeat 目標，因那會改變現行資料契約。

## UI 與文案

確認句由 `assets/i18n/ui.json` 的
`buildqueue.confirm.delete_mode_above_product` 提供，兩個 `%s` 均由外部 Housing／Trade Goods
顯示名代入。確認框沿用 `confirmScreen` 與其雙軸文字安全框；Go 不得內嵌句子。

## 驗收

- 單一模式在末尾不刪。
- 單一模式在產品前刪除。
- 多個模式在產品前全部刪除；只有模式時保留最後一項。
- 一般產品在前、模式在末尾不變。
- 外部雙語文案的 placeholder 與確認框高度通過。

RE 位址、bytes、正版字串池與證據等級見
[`game-menu-popup-ui-text-audit-20260826.md`](../re/game-menu-popup-ui-text-audit-20260826.md)。
