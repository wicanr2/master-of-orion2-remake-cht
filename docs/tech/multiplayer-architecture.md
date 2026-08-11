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

## 4. 落地順序與完成界線

1. **熱座(hotseat)**——多位真人同機,在回合迴圈裡同時/輪流下令。零網路、零決定性風險,原版列為
   多人一種,最忠實也最省。先做這個驗證「多人回合流程」本身。
2. **引擎決定性化**——收斂 RNG(統一種子)+ map 迭代順序;加「兩機各跑同指令序列 → 狀態雜湊比對」
   的決定性回歸測試(desync 偵測器)。這是網路對戰的地基,且獨立於網路碼可先驗。
3. **區網 lockstep over TCP**——host/client、config／快照廣播、逐回合指令收齊→依席位重播→
   第二階段狀態指紋→結算。最低鏈已完成；可選的重連／心跳／身份驗證／TLS 1.3 也已接入，
   NAT 穿透仍不是本專案內建能力。

> 歷史排序保留作決策紀錄；目前 1–3 的區網最低鏈均已接入，下一輪只追公網可靠性或原版
> oracle，不重新挖已完成的傳輸骨架。

## 5. 現況

> 本節依 2026-08-10 生產呼叫圖重新盤點；本文其餘各節保留原版考據與「保留 lockstep、
> 傳輸改 TCP」的架構決策。

- **熱座已實作**(上表第 1 步):`internal/shell/hotseat.go`(席位交換模型)、
  `cmd/moo2/hotseat.go`(交接畫面)、`cmd/moo2/multiplayer.go`(原版 MULTI-PLAYER 設定畫面)。
  主選單的「多人對戰」接上了,版面座標全部取自反組譯(`Multi_Player_Screen_` @ 0xF4D99)。
  「結束回合」改成全員下完令才推進世界;席位進存檔。
- **網路對戰最低可玩鏈已完成**：`cmd/moo2` 由大廳名冊進入主機共同新局；主機廣播設定／種子／
  席位快照，客戶端套用同一快照。玩家狀態入口會記錄 `PlayerCommand`，每回合依序走
  `turn_done` → 依玩家編號從共同基準重播 → `turn_ready`；`NetworkStateHash` 分歧、未知指令、
  錯席位與重複封包都失敗即關閉。`networkWaitScreen` 是正式回合等待畫面，`netNextTurn` 仍只供畫廊示範。
- **可選公網可靠性已接入**：`internal/netplay` 在每次連線先送一次性 challenge，非空
  `MOO2_NET_AUTH` 時以 HMAC proof 驗證共享身份；`MOO2_NET_TLS=1` 時再套 TLS 1.3（預設為
  記憶體內短期憑證，正式憑證可由 `LobbyOptions.TLSConfig` 注入）。第一次加入取得 resume token，
  斷線後可在寬限期內恢復原玩家編號；Session 以 Ping/Pong、逾時與重連 callback 維持連線。
  逾時後才會把對局標為失敗即關閉。
- **仍未由本專案處理的部分**：NAT 穿透。跨網段公網部署要使用外部 relay 或 UPnP；這不是原版
  IPX／數據機／序列／TEN 的恢復，也不能把區網 TCP 直連寫成 NAT 解法。
- **數據機／序列埠／TEN 直連:明確不做**——硬體與服務已不存在，remake 走 TCP／熱座。

> 先前把「netplay 測試與畫面骨架完成」誤報成「網路對戰已可玩」；本輪已補上生產呼叫圖，並以
> 兩個實際 process 的 loopback TCP 加上席位快照／指令重播測試驗證最低鏈，再以 TLS／HMAC／
> resume／心跳測試驗證可選可靠性。這不等於跨網段 NAT 穿透或原版逐值 oracle 完成，也不恢復
> 已失效的 IPX／數據機／序列／TEN。
- 熱座本身的已知不對稱(非當前席位的結算時點、勝負判定只對當前席位跑)見
  `docs/re/01-gap-report.md` 第 3 項(Colony+Event 畫面)。
- 對應 WORKLIST「2026-08-10 盤點結論」的多人網路對局調整項。
