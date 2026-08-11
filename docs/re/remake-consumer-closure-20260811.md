# Remake 消費端收尾：CMBTSHP／爆炸／SABOTAGE／領袖 ETA

日期：2026-08-11。這份紀錄只描述本輪已接入的 remake runtime 消費端；原版
IDA 證據仍以 [`oracle-static-ida-20260811.md`](oracle-static-ida-20260811.md)
的 raw 位址、輸入雜湊與工具版本為準。

## 本輪完成

| 消費端 | remake 行為 | 證據邊界 |
|---|---|---|
| `CMBTSHP` timer | `CMBTSHPFrameAtTick` 以固定 4 tick/frame 播放 `[0,1,2,1]` 短掃掠；戰術艦移動時啟動，16 tick 後固定回最近朝向幀 | IDA 深度掃描 `sub_30062`／`sub_30631`／`sub_31F25`／`sub_3F628` 只找到資產載入、繪圖、輸入迴圈與 heading；未找到 frame counter 寫入或 timer 常數。20 幀、`45*color+rawPicture`、16 向 heading 已證實；timer 是可重播近似 |
| 艦船爆炸事件 | 事件 8 呼叫 `resolveStrategicShipExplosion`：`Random(201)+74` 取主勢能、移除主艦、每步 `-20`，以 type/resistance consumer 對最多三艘倖存艦寫入 `Ship.Damage` | raw `0x3868F`／`0x39985`／`0x40C2A` 的純公式已證實；raw fleet／colony record 與 resistance 表未完整命名，`/20` 是 remake 尺度 adapter |
| `SABOTAGE` 分數 | `spyMissionScore` 明列攻方 Spies slot／科技／種族＋領袖、守方 Agents slot／科技／政府／種族＋領袖，再計算 `E=T+DB-AB` 與 `p`；SABOTAGE 使用 `T=70` 與原版建造成本加權池。`SpySlotBonus` 已與 raw `sub_101483 @ 0x101483` 對齊 | `sub_1014A4 @ 0x1014A4` 的 packed relationship byte、兩張 raw table、亂數位置與 60／70／80／90／±80 分支已追回；兩張 table 是 runtime `0xFF` 初始化／上游填值，完整欄位語意與 `toggle_flag` 邊界仍是 oracle 差異。remake 以玩家可感知 AB／DB／E 近似完成，不宣稱 raw score parity |
| 防守 Agent | 玩家可訓練／解除 Agent（63 上限）；AI 依既有 remake 週期補充；Spy-vs-Spy 判定擊殺防守方時，runtime 會實際扣一名 Agent，雙向攻擊皆適用 | 手冊的 slot bonus／±80 門檻已接；訓練成本 30 BC、AI 週期是 remake 拍板值 |
| 領袖 ETA callback | `RawStatus=1`、`RawETA:1→0`、`RawLocation=1` 觸發 `applyLeaderETACallback`：撤銷／重套領袖增量、刷新所有殖民地士氣；領袖保留，不把 ETA=0 誤解成解雇 | `sub_E2AB1 @ 0xE2AB1` 的六槽掃描與 `sub_E1D59`／`sub_DF8F0`／`sub_E2710` callback 鏈已由 IDA 證實；raw 設計／艦隊欄位無安全一對一模型，remake 採完整玩家可感知近似 |

## Win95／VESA 呼叫邊界

本輪不再深挖與遊戲規則無關的 Win95 視窗、VESA、GDI／音訊裝置或輸入 API
內部呼叫。這些呼叫只要不改變玩家可見的遊戲狀態，就以 remake 自己的跨平台
視窗／音訊／輸入抽象層近似；只有它們的回傳值直接影響上述 timer、分數或 callback
消費端時，才保留原始呼叫位址作為證據索引。`VESA.COM` 尚未取得可啟動 runtime，
因此不把 API 路徑的靜態猜測升格成原版行為。

## 公式與安全護欄

- 爆炸事件只在全帝國艦數大於 1 時移除主艦，避免一次事件把最後一艘艦直接清空。
  次級傷害只寫入既有 `Ship.Damage`，不在 raw fleet／colony 欄位尚未對回時自行增加全滅規則。
- `Ship.Damage` 的尺度化上限沿用 `shipMaxHP - ShipDamageFloorHP`；既有修復、戰鬥與靠港
  修復路徑因此可以消費這份傷害，不需要另一套不可存檔的事件血量。
- SABOTAGE 的 Agent 與進攻 Spy 是兩個獨立 slot；Agent 被擊殺後才從防守方數量扣除，
  不會誤扣進攻方 Spy，也不會只在新聞文字中顯示一個永遠不變的 Agent 數。
- raw `sub_1026CF @ 0x1026CF` 讀取 `record + 0xE57 + otherEmpire` 的低 6 位，
  `sub_1026F1 @ 0x1026F1` 讀取同一 byte 的高 2 位；`sub_10278D @ 0x10278D` 寫入時保留
  高 2 位。`sub_101483 @ 0x101483` 的 helper 是 `n<=5:2n`、`6..10:n+5`、
  `>10:floor((n-10)/2)+15`，已由 `gamedata.OriginalSpyScoreHelper` 使用。
- `sub_1014A4` 第一段分數形狀為 helper(raw count) + `Random(100)` +
  `word_1ACE78[row*4]` − `Random(100)` − `word_1ACE7A[col*4]` − helper(other count)，
  另有 raw context 修正與 mission mode 分支；第二方向帶 `+0x14`。這是原版 raw
  Spy-vs-Spy／Agent 決策鏈的已證實算術形狀，不把兩張尚未由上游填值的表誤命名成
  remake 的科技／政府欄位。
- CMBTSHP timer 不使用 wall-clock；同一回合與同一移動輸入會得到同一組 frame，符合本專案
  可重播抽樣測試與多人 deterministic turn 的要求。

## 抽樣驗證

均在既有 `pto2-remake-build:latest` Docker image 內執行，沒有在主機啟動 Go／Ebiten：

- `go test ./internal/gamedata -run 'Original(Event|Explosion)' -count=1`：通過。
- `go test ./internal/shell -run 'CMBTSHP|StrategicShipExplosion|Sabotage|SpyMissionResult|AdvanceEspionageConsumesAIDefensiveAgent|TrainAndDismissDefensiveAgent|AdvanceActiveLeaderETA' -count=1`：通過。
- 事件／間諜／領袖相鄰回歸抽樣：通過。
- Docker `mm2-go:latest` + trap 管理的 Xvfb 執行 `go test ./cmd/moo2 -run '^$' -count=1`：退出碼 0；只做 UI package 編譯，不代表完整遊戲測試。

`gamedata` 的 raw spy packing／helper、`shell` 的 AB／DB／E、SABOTAGE 建築消費、Agent
回寫、ETA callback 與 CMBTSHP 固定 tick 抽樣已通過；完整 `go test ./...` 仍不是本輪驗收目標；
既有 `TestPopulationGrowthWriteback` 的人口分配差異另行列管，
不與本輪四個消費端混報。
