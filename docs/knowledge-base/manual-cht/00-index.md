# 銀河霸主 II 遊戲機制大全(繁體中文)

《Master of Orion II: Battle at Antares》(銀河霸主 II:安塔瑞斯之戰,1996)的遊戲機制中文整理與考據,供玩家速查、也作為本 remake 專案對齊原版的規格真相來源。

> **版權聲明**:本系列為遊戲**機制、數值、公式**的中文整理與考據,**非手冊原文的翻譯**。所有內容以編纂者自己的話重述為表格與條列,並標註手冊頁碼供查證。遊戲規則與數字屬事實範疇;原遊戲、手冊與其具體文字表達之版權屬原作者(MicroProse / Simtex)。本 repo 不含原手冊檔案或其逐字譯本。

## 章節

| 章 | 主題 | 內容 |
|---|---|---|
| [第一章 種族](01-races.md) | 種族系統 | 13 個經典種族招牌特性與數值、自訂種族點數(Picks)成本表 |
| [第二章 建築](02-buildings.md) | 殖民地建築 | 全 40+ 棟建築/衛星:效果/維護費/前置科技/成本(附誠實缺口標註) |
| [第三章 戰鬥](03-combat.md) | 太空戰/地面戰/軌道轟炸 | 命中/傷害/護盾/裝甲/飛彈/武器 mod/艦艇設計/地面戰公式 |
| [第四章 經濟與帝國](04-economy-empire.md) | 經濟/政府/勝利 | BC 收入模型(人頭收入為核心)、殖民地產出、政府型態、指揮評等、外交、勝利條件、間諜 |

## 使用說明

- 每項數值後標的「(手冊 p.NN)」為 `GAME_MANUAL.pdf`(patch 1.5 隨附)印刷頁碼;`M150` 指 `MANUAL_150.html`(1.5 patch 的機制註記)。
- 誠實標註:凡「手冊未給數字」「社群逆向推定」「待實機驗證」「remake 近似」之處均就地標明,不假裝確定。
- 與 remake 實作的對照散見各章「交叉參考」段(對應 `internal/gamedata/`、`internal/engine/`、`docs/tech/`)。

## 相關文件

- [`../original-gameplay-reference.md`](../original-gameplay-reference.md) — 原版玩法考據 + SAVE10.GAM oracle 反推的開局真值
- [`../../tech/moo2-formulas-reference.md`](../../tech/moo2-formulas-reference.md) — 公式技術參考
