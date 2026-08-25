# Gameplay 系統忠實化現況

> 更新：2026-08-24。這是導覽文件，不是待辦複本。當前工作以
> [`WORKLIST.md`](../../WORKLIST.md) 頂端活表為準；逐系統證據邊界以
> [`docs/re/parity-matrix.tsv`](../re/parity-matrix.tsv) 為準。

## 狀態定義

- **remake 已實作**：有玩家可觸及的消費端，且有相稱測試；不自動代表原版精確。
- **已證實 parity**：輸入、順序、常數或表格可回查原版執行檔、手冊或可重現實驗。
- **remake 近似**：功能可玩，但仍使用調校值、簡化資料模型或未閉合的原版分支。
- **未知／oracle 依賴**：現有靜態證據不足；只有影響玩家路徑時才開啟有界的原版實驗。

## 玩法縱向鏈

| 系統 | remake 現況 | 原版對齊邊界 |
|---|---|---|
| 新遊戲與開局 | 設定、種族／客製種族、命名與母星流程可玩 | 難度倍率、開局建築與先進科技選取尚未完全對齊 |
| 殖民地與帝國經濟 | 食物、工業、研究、污染、士氣、建造、稅收、六項維護與拆船／建築／間諜／領袖的負國庫處分可運作 | `Next_Turn_Calc_` 的完整順序與研究溢出仍未閉合；戰鬥艦破產排序採 remake 戰力近似 |
| 研究 | 玩家／AI 研究前選 application、Creative／Uncreative 分支、原版超額突破率、成功清零、解鎖與存檔消費 | 殖民地其餘人口修正、原版 AI 主題估值與少數特殊科技 callback 仍需對齊 |
| 外交、間諜、領袖 | 主要 UI、條約、貿易、研究協議、任務與管理可玩 | 關係、記憶、特殊交易／SABOTAGE 上游分數與領袖 callback 仍含近似 |
| AI 與艦隊 | 多 AI 擴張、研究、造艦、攻擊與艦隊航行可運作 | 原版 `AI_Screen_` 排程、命令鏈、目標與戰爭／和平邏輯尚未完整重建 |
| 戰鬥 | 快速結算與格子戰術皆有光束、飛彈、魚雷、戰機、登艦、轟炸與地面戰消費端 | 命中／傷害次序、效果表、爆炸半徑、射界、小型化與敵艦設計仍有空白，不得總稱「兩路徑皆為原版真公式」 |
| 議會與勝利 | 人口票數、最早第 25 回合／後續每 25 回合召開、玩家／AI 投票與主要勝利路徑可運作 | 候選、棄權、外交投票、話術及異議／投降後敵對流程尚未全數對齊 |
| 多人 | 現代化的房間／座位／狀態同步核心已有可測路徑 | 重連、心跳、身份驗證、加密與 NAT relay／UPnP 是現代化擴充，不屬於原版 parity |

## 重要技術文件

- 經濟與維護：[`colony-economy-maintenance.md`](colony-economy-maintenance.md)
- 研究：[`research-system-status.md`](research-system-status.md)
- 地面戰：[`ground-combat-algorithm.md`](ground-combat-algorithm.md)
- 艦艇設計：[`ship-design-space.md`](ship-design-space.md)
- 間諜：[`spy-system.md`](spy-system.md)
- 勝利與議會：[`victory-conditions.md`](victory-conditions.md)

## 更新規則

1. 實作完成時，更新 `WORKLIST.md` 的唯一活躍條目。
2. 解出原版規則時，先回填 parity 矩陣與對應規格，再改程式與測試。
3. 過期斷言直接刪除；錯誤形成過程只保留在 [`docs/re/01-gap-report.md`](../re/01-gap-report.md) 與 Git 歷史。
