# 多人對戰:原版通訊考據 + 重製架構建議(2026-07-11)

> 起因:使用者問「主選單的『多人對戰』,go/ebiten 能做到嗎?」→ 先考據原版走什麼通訊,再定重製方向。
> 方法:第一性原理 + 手冊為據(rulebook 62/65)。原版 CD 手冊是掃描本(`pdftotext` 僅 9 字元),
> 用 tesseract(docker `ocr-tesseract:local`,200dpi)OCR 出多人頁原文;架構面用 patch 1.5 手冊
> (`moo2_patch1.5/MANUAL_150.html`,可文字擷取)佐證。

## 1. 原版多人通訊方式(手冊原文,權威)

原版 CD 手冊「The multiplayer option requires the following」逐字列出(OCR 出處:
`original_game/Master of Orion 2 - CD Manual.pdf` 第 3 頁系統需求段):

| 傳輸 | 人數 | 手冊原文 |
|---|---|---|
| **序列直連線** | 2 人 | 「Null-modem serial cable (2 players)」——兩台 PC 序列纜線對接 |
| **數據機** | 2 人 | 「Windows-compatible 9600 baud modem or faster (2 players)」——撥接直連 |
| **IPX 區域網路** | **2–8 人** | 「Local area network on IPX protocol (2-8 players)」 |
| **網際網路(TEN)** | — | 「You can play Master of Orion II over the Internet on the **TEN service**」——需另裝 TEN 軟體 |

**底層**:安裝要裝 **DirectX 6.1**(手冊安裝步驟)→ 網路走 **DirectPlay**(DirectX 的網路抽象層,
統一封裝序列/數據機/IPX)。網際網路對戰**不是遊戲原生 TCP**,是靠第三方 **TEN
(Total Entertainment Network,1990 年代線上對戰服務)**轉接。

> 註:此 CD 為 Hasbro 再版(手冊提 DirectX 6.1 / Acrobat 4.0 / TEN,約 1998–99),但四種傳輸方式
> 與 1996 MicroProse 原版一致(IPX/數據機/序列 + 熱座)。

## 2. 架構:決定性 lockstep + host/client(patch 1.5 手冊佐證)

patch 1.5 手冊「Config → Network Synchronization」節原文:

> "In network multiplayer the host's config is broadcast and applied by clients so that all sides have
> the same game settings. … Only non-interface options are broadcast …"

加上 1.5 patch 大量修復「desync and stall」類 bug(如 Techfield Desync、Zero Marines Raid Desync、
Mutation-on-battle-turn desync),可反推原版網路架構為:

- **決定性 lockstep**:每台機器各跑**同一份決定性模擬**,網路上只交換玩家指令,不傳整份遊戲狀態;
  各機狀態一旦分歧就 **desync**(卡住、要 crash 重載)。這正是 1.5 一直在修的東西 → 證明模型如此。
- **host/client**:有一台 host 廣播遊戲設定(config),clients 套用,確保各方**規則完全一致**
  (lockstep 的前提)。
- **同時回合(simultaneous turns)**:所有玩家同一回合同時下令,回合末一起結算(非嚴格輪流),
  最多 8 位(人/AI 混合)。

## 3. 對 go/ebiten 重製的意義

**語言/引擎不是障礙**:ebiten 只管畫面/輸入/音效;網路是 Go 標準庫 `net` 的強項(Go 本為高併發
網路服務而生)。ebiten 上做連線對戰的遊戲很多。真正的成本在**多人架構**,與語言無關。

**忠實 ≠ 重現舊傳輸**:原版四種傳輸(序列/數據機/IPX/TEN)在現代作業系統**全數死亡**——IPX 協定、
撥接、序列對接、TEN 服務都不存在了。所以「忠實」是**保留架構、換掉傳輸**:

| 面向 | 原版 | 重製建議 |
|---|---|---|
| 同步模型 | 決定性 lockstep | **保留 lockstep** |
| 回合制 | 同時回合、host 廣播設定、指令同步 | **保留** |
| 傳輸 | 序列 / 數據機 / IPX / TEN | **換成 TCP**(必要時 UDP;Go `net`) |
| 玩家數 | 2–8(人/AI 混合) | 同 |
| 熱座 | 同機多人(無網路) | **保留,且最省——起步先做這個** |

**關鍵工程難點 = 引擎決定性化**:lockstep 要求同一份模擬在不同機器跑出**逐位元相同**的結果。
現行 remake 引擎有兩個已知不確定性來源要收斂:
1. **RNG**:戰鬥已用「回合數+星索引」種子化(可重現),但要全域統一種子廣播 + 所有隨機路徑走同一
   PRNG,不可用 `math/rand` 全域源或 wall-clock(對齊 memory `moo2-headless-gui-loop-must-bound`
   同源的決定性紀律,及 Workflow「Date.now/rand 不可用」的同一道理)。
2. **map 迭代順序**:Go `map` 迭代隨機——任何影響模擬的 `range map` 都要改成排序後迭代或用 slice,
   否則各機結算順序不同 → desync。這正是 1.5 patch 修 desync 的同一類問題。

## 4. 建議落地順序(投報比)

1. **熱座(hotseat)**——多位真人同機,在回合迴圈裡同時/輪流下令。零網路、零決定性風險,原版列為
   多人一種,最忠實也最省。先做這個驗證「多人回合流程」本身。
2. **引擎決定性化**——收斂 RNG(統一種子)+ map 迭代順序;加「兩機各跑同指令序列 → 狀態雜湊比對」
   的決定性回歸測試(desync 偵測器)。這是網路對戰的地基,且獨立於網路碼可先驗。
3. **區網/線上 lockstep over TCP**——host/client、config 廣播、逐回合指令收齊→同步→結算、
   斷線/重連處理、狀態雜湊校驗(偵測 desync)。中大型獨立子專案。

> 排序原則:1、2 可先做且低風險;3 是大工程,排在音樂/忠實新遊戲流程/像素對齊之後(見
> `remaining-work-roadmap.md` 的阻塞分類——多人屬「需你授權方向的大工程」)。

## 5. 現況

- remake 目前**單人**(玩家 + 3 AI);多人**未接**;主選單「多人對戰」按鈕是**沿用原版美術**,
  尚無功能。
- 本文只做考據 + 方向定案,無程式改動。對應 WORKLIST「Phase 9 — 多人對戰」。
