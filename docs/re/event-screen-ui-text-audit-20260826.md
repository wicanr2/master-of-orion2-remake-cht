# 事件快報畫面與文案 RE 稽核（2026-08-26）

## 問題與舊歧義

`cmd/moo2/eventscreen.go` 把事件畫面描述成「沒有原版 `EVENTS.LBX` 版面資料」，因此以
`TURNSUM.LBX` 加黑底面板重建；固定台標、標記與按鈕仍以中英成對字串內嵌。需要確認原版
畫面鏈、可用資產、事件圖索引與輸入範圍，才能決定外部文案及安全 fallback。

## 輸入與工具

- 輸入：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- IDA 資料庫：`Orion2.exe.i64`，SHA-256
  `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- 工具：IDA Pro 9.4／IDAPython，處理器 `metapc`；本文位址均為 DOS/4GW image 的
  IDA linear address。非破壞性匯出腳本為 `tools/ida/audit_event_screen_ui.py`。
- 資產：正版 `EVENTS.LBX`，只讀解析；Go `internal/lbx` 解碼器另行核對 count、尺寸、
  幀數及內嵌調色盤，不把檔名當成影像形狀證據。

## 函式邊界與符號勘誤

以下結論為**已證實**：IDA 函式邊界、直接 caller 與原始指令一致。

| raw 函式 | 邊界 | 直接關係 | 可證實用途 |
|---|---|---|---|
| `sub_2031D` | `0x2031D..0x203CB` | `sub_21371 @ 0x219C8` 呼叫；`0x2039E` 呼叫 `sub_20538`，`0x203A3` 把 `sub_20460` 當 callback | 啟動事件畫面 |
| `sub_203CB` | `0x203CB..0x20400` | `sub_21371 @ 0x21A65` 呼叫 | 以 `EVENTS.LBX`、`eventIndex+2` 載入事件插圖 |
| `sub_20400` | `0x20400..0x20460` | `sub_21371 @ 0x21A4D／0x21AC1` 呼叫；`0x20445` 呼叫 `sub_20460` | 事件畫面輸入／重繪迴圈 |
| `sub_20460` | `0x20460..0x20538` | `sub_20400 @ 0x20445` 呼叫 | 繪製事件畫面 |
| `sub_20538` | `0x20538..0x20612` | `sub_2031D @ 0x2039E` 呼叫 | 載入 `EVENTS.LBX#0/#1` 並淡入 |

`symbols_fixed.tsv` 對應上述五個起點。`func_names.txt` 則把名稱依序錯放在下一個相鄰函式：
例如把 `Start_Main_Event_` 放在 `0x203CB`、`Draw_Event_Screen_` 放在 `0x20538`，並把
`Event_Fade_In_` 放到大型事件 consumer `sub_206A2 @ 0x206A2..0x212E1`。因此本題保留 raw
函式名與位址，不以衝突的外部名稱覆蓋定位。

## 畫面、插圖、文字與輸入

- **已證實**：`sub_20538` 於 `0x20555..0x2056A` 載入 `EVENTS.LBX#0`，於
  `0x205D0..0x205DA` 載入 `EVENTS.LBX#1`。正版資產解析結果為：`#0` 24×24、1 幀、
  256 色內嵌調色盤；`#1` 640×480、31 幀、無內嵌調色盤。
- **已證實**：`sub_203CB @ 0x203D9..0x203E9` 對事件索引加 2 後載入
  `EVENTS.LBX`。資產 `#2..37` 恰為 36 張 157×125、1 幀、64 色內嵌調色盤影像。
- **已證實**：`sub_20460 @ 0x204A0..0x204B0` 以 `(320,14)` 繪製已載入事件插圖。
  `0x204B5..0x204FF` 量測並以格式化段落 helper 繪製事件字串；可直接回查的常數包含
  x 中心 `0x1BD=445`、段落上限 `0x226=550` 與 y=`0x37=55`。參數的完整文字
  baseline／對齊旗標語意仍是**強推論**，不把現行中文字框宣稱為逐像素原版座標。
- **已證實**：`sub_20400 @ 0x20415..0x2042A` 建立 `(0,0)..(639,479)` 的全畫面
  輸入欄；原版不是只有一顆 100×24 的「繼續」按鈕可點。
- **已證實**：`#1` 是 delta 動畫資料；31 幀可由既有 `AccumulatedUpToRGBA` 正確累積。
  原版 wall-clock／fade 細節尚未逐週期追回，remake 固定每 3 tick 一幀只能標為
  **remake timing approximation**。

## Remake 映射與停止線

1. 有 `EVENTS.LBX` 時，優先顯示 `#1` 累積動畫；合法事件 ID `0..35` 再疊上 `ID+2`
   插圖。索引前先驗證 archive count，任何缺檔、解碼或越界均安全退回現有自繪面板。
2. 原版整張畫面可點，remake 保留可見的「繼續」按鈕作現代操作提示，但點畫面其他位置也應
   繼續；按鈕不是原版熱區證據。
3. 固定台標、好壞標記、勘查標記、按鈕與轉場名稱統一放入 `assets/i18n/ui.json`。
   `EventReport`／`DiscoveryReport` 的動態內容由玩法資料提供；其餘事件資料外部化另依活表逐批處理。
4. 星系發現沒有 `EVENTS.LBX eventIndex+2` 證據，仍使用共用安全面板，不捏造事件插圖。
5. 本輪停止於玩家可見畫面、資產選擇、輸入與文字安全框；不深挖淡入 driver 或硬體計時。

## 驗證

- catalog 雙語完整性與 `eventscreen.go` 無 `.tr(`／固定雙語句／直接字型繪製的靜態測試。
- 事件標題、標記、正文、按鈕的 runtime bitmap-font 雙軸 containment 測試。
- `EVENTS.LBX` count／尺寸／合法與越界事件 ID 測試；負數測試 fixture 不得索引；缺資產 fallback 測試。
- Docker＋Xvfb 以正版資料抽取畫廊 `05_event.png`，人工確認原版 GNN 背景、插圖、文字與按鈕
  沒有互相遮蔽；一般畫廊仍驗證缺 `EVENTS.LBX` 的 fallback。
