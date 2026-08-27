# RACES 種族／間諜文案與版面稽核（2026-08-27）

## 證據分級

- **已證實**：`Race_Screen_ @ 0x10ACBA` 是頂層，`Draw_Race_Screen_ @ 0x10C8E0` 負責重繪；
  原版沒有獨立 Spy／Espionage 畫面。函式與負面搜尋證據見
  `docs/re/screen-coords-spy-leader.md` §3.1–3.2。
- **已證實**：`_race_spy_btns @ 0x18406F`、`_race_bar @ 0x18400D` 與 `_race_spies @ 0x184053`
  建立左四右三、最多七個已接觸種族的固定版面。每列三槽 x 偏移為 `0/+76/+149`；座標與原始
  位址見同文件 §3.3。
- **未知**：三槽分別對應 Espionage／Sabotage／Hide 的原版左右順序，以及完整 click callback
  仍未閉合。現行「增派／循環任務／隱匿」是明標 remake adapter，不得升格為原版精確控制流。
- **已證實（remake source）**：`races()` 仍有八組 `tr`／格式字串；英文模式直接輸出中文
  `AIOpponent.StanceName` 與 `GameSession.AIRelationName`。`shell.SpyMissionLabel` 也把玩家顯示字串
  放在規則層。
- **已證實（remake geometry）**：五行帝國資訊已有 `textSafeRect`；三個任務槽與 Agent 控制仍在
  `races()` 內臨時建立矩形，Agent 文字使用只有寬度限制的舊置中 helper。
- **已證實（runtime geometry）**：資料區高約 73px，既有五行各自高 16px，且 y 節奏後四行只有
  12px，因此測試雖確認各行位於大框內，實際相鄰字形仍必然重疊。

## 本輪邊界

將 RACES 動態模板、任務名、Agent 控制與轉場名稱移到 `ui.json`；態勢及 AI 關係沿用 INFO
畫面既有的 typed 分數／相容存檔映射。三槽與 Agent 文字框改由命中矩形推導。不改間諜成本、
任務結算、外交或原版未知 callback。

資訊區改為四條 16px 字形、17px 節奏；移除第五行重複操作提示。整個資訊區仍是同一外交熱區，
hover 外框與點擊行為不變。

## 畫廊勘誤

- **已證實（remake runtime）**：新增正常路徑 `15a_races.png` 後，確認舊 Agent 控制位於
  `(20..228,396..434)`，會直接占用左欄第 4 個帝國槽（其關係區自 y=370 開始、間諜列在
  y=445）。「左下空白區」註解錯誤。
- Agent 控制已移至右下 BONUSES 面板：狀態 `(340,384,184,16)`，訓練／解散分別
  `(340,400,90,18)`、`(434,400,90,18)`；底緣恰為 y=418，不侵入從 y=418 起的外交按鈕列。
  這是 remake adapter 版面修正，不宣稱為原版 Agent widget 座標。
- **已證實（remake runtime）**：修正後繁中與英文正版資料畫廊各 36/36；目視檢查新增的
  `15a_races.png`，四條帝國資訊無上下重疊，三個間諜槽文字在框內，Agent 控制位於 BONUSES
  面板且未遮住第 4 個帝國槽或外交按鈕。
