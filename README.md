# 銀河霸主 2:安塔瑞斯之戰 — go/ebiten 重製 + 繁體中文化

以 [OpenOrion2](https://github.com/next-ghost/openorion2) 為參考基底,用 Go + [Ebitengine](https://ebitengine.org/) 重新打造《Master of Orion II: Battle at Antares》(1996),並提供完整**繁體中文**在地化與英文原文切換,支援 **1.3 / 1.5** 兩個版本的規則與資料。

> 目前狀態:**Phase 0 — 可行性研究與 kick-off 完成**。詳見 `docs/kickoff/` 與 `PLAN.md`。

## 這是什麼

《銀河霸主 2》是 4X 太空策略遊戲的經典。本專案的目標:

- 基於 OpenOrion2 的資產解碼與存檔資料模型,移植到跨平台的 go/ebiten。
- 完整中文化**所有訊息,包含按鈕**;主選單可切換 中文 / 英文。
- 主選單可選 **1.3** 或 **1.5** 版本(對應各自的資料與規則)。
- 重新檢視並依原版手冊重建遊戲規則(OpenOrion2 本身尚未實作 gameplay)。
- 收錄華人圈的中文討論考據與文化現象整理。

誠實揭露工作量:OpenOrion2 自述為「partial savegame viewer, no gameplay」,經查證屬實。因此本專案分兩軌 —— **移植現成的檢視器/資產層**,以及**從手冊從零重建整個回合制遊戲引擎**。後者是絕大部分工作量。詳見 `docs/kickoff/00-feasibility.md`。

## 遊戲資料(玩家自備正版)

本 repo **不含**任何原版遊戲檔、手冊或官方 patch(版權所有)。你需要自備正版《Master of Orion II》,並將遊戲的 `*.lbx` 資料指給程式讀取。取得正版途徑如 GOG。

## 建置與執行

編譯與測試一律在 Docker 進行(細節見 `docs/kickoff/06-ebiten-porting.md`)。Phase 1/2 完成後補上具體指令。

## 文件

- `PLAN.md` — 分階段計畫與里程碑
- `WORKLIST.md` — 可勾選工作清單
- `docs/kickoff/` — 可行性研究與 kick-off 知識庫(openorion2 盤點、中文化/按鈕/字型策略、LBX/patch、ebiten 移植)

## 致謝

這個專案站在許多人的肩膀上:

- **[OpenOrion2](https://github.com/next-ghost/openorion2)**(next_ghost,GPL v2)—— 提供 LBX 資產解碼器與完整的 MOO2 存檔資料模型逆向,是本專案的參考基底。
- **[1oom](https://gitlab.com/1oom-fork/1oom) 社群** —— 前作《銀河霸主 1》繁中化的引擎與經驗來源;其 CJK 渲染與按鈕烘字踩雷經驗直接指引了本專案的中文化策略。
- **MOO2 1.5 社群 patch 團隊** —— 持續維護的非官方 patch(至 2026 仍在更新),為 1.5 版規則與資料的權威來源。
- **像素中文字型作者** —— Cubic 11(方舟像素字體)、Fusion Pixel 等開源字型,讓繁中能以契合原版美術的像素風呈現(最終選用見 `docs/kickoff/04-font-choice.md`,授權於定案時標明)。
- **華人玩家社群** —— 數十年來的攻略、考據與討論,是文化考究章節的基礎。
- 原作 **Simtex / MicroProse** —— 創造了這款不朽的經典。

## 授權

本專案的原始碼衍生自 GPL v2 的 OpenOrion2,故以 **GPL v2** 釋出。原版遊戲資產與字型各依其授權,不包含於本 repo。
