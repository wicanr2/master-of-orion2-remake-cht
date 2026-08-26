# 最終得分／名人堂畫面 RE 稽核（2026-08-26）

## 問題與舊歧義

`cmd/moo2/hiscore.go` 宣稱沒有原版 Hi Score 底圖，使用 `TURNSUM.LBX` 加自繪黑色面板；
固定勝負文案、勝利原因、摘要格式、分數列名稱與繼續按鈕仍分散在 UI 及 `internal/shell`。
原版具名 `Load_Hi_Score_Screen_`、`Draw_Hi_Score_Screen_` 與獨立 Hall of Fame 流程，故先
核對函式鏈、資產、列座標與輸入，再決定 remake 映射。

## 輸入與工具

- `Orion2.exe` SHA-256：
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- `Orion2.exe.i64` SHA-256：
  `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- IDA Pro 9.4／IDAPython，處理器 `metapc`；本文位址是 DOS/4GW image 的 IDA linear
  address。非破壞性匯出腳本：`tools/ida/audit_hi_score_screen_ui.py`。
- 正版 `SCORE.LBX` 只讀解析；Go `internal/lbx` 解碼器核對 count、尺寸、幀數及 palette。

## 函式邊界與符號勘誤

以下為**已證實**的 IDA 函式邊界與直接 caller：

| raw 函式 | 邊界 | 關係／用途 |
|---|---|---|
| `sub_9D7EA` | `0x9D7EA..0x9D816` | `sub_9EA3B @ 0x9EA93` 呼叫；建立整張 Hi Score 輸入欄 |
| `sub_9E27A` | `0x9E27A..0x9E3AD` | `sub_9EA3B @ 0x9EAF8／0x9EB0C` 呼叫；繪製 Hi Score |
| `sub_9EA3B` | `0x9EA3B..0x9EB42` | `sub_9F712 @ 0x9F72E` 呼叫；Hi Score screen loop |
| `sub_9EB42` | `0x9EB42..0x9EC32` | `sub_9EA3B @ 0x9EA69` 呼叫；載入 Hi Score 資產 |
| `sub_9EDE1` | `0x9EDE1..0x9EE43` | `sub_9F286 @ 0x9F356` 呼叫；載入 Hall of Fame 資產 |
| `sub_9EE43` | `0x9EE43..0x9F286` | Hall of Fame 繪製函式 |
| `sub_9F286` | `0x9F286..0x9F447` | Hall of Fame screen loop |
| `sub_9F712` | `0x9F712..0x9F981` | 回合流程 `sub_13870` 兩處呼叫；先跑 Hi Score，再更新／顯示 Hall of Fame |

`symbols_fixed.tsv` 與上述邊界一致。`func_names.txt` 把 `Hi_Score_Screen_` 放到 loader
`0x9EB42`、把 `Hall_Of_Fame_Screen_` 放到無 caller 的相鄰 `sub_9F447`、把
`End_Of_Game_Hi_Score_` 放到 `sub_9F981`。本題保留 raw 名稱／位址，不以衝突索引改名。

## 資產、繪製與輸入

- **已證實**：`sub_9EB42 @ 0x9EBC5..0x9EC27` 載入 `SCORE.LBX#0`、`#1`，再迴圈
  載入 `#2..14` 共 13 張圖。`sub_9EDE1 @ 0x9EE03..0x9EE3E` 則載入 `#15/#16`。
- **已證實**：正版 `SCORE.LBX` 有 17 個資產：`#0` 是 640×480、1 幀、256 色內嵌
  palette 的 Hi Score 背景；`#1` 是 42×48；`#2..14` 是 13 張 40×46；`#15` 是
  640×480、1 幀、256 色 Hall of Fame 十列表格背景；`#16` 是 23×23。
- **已證實**：`sub_9E27A @ 0x9E2AB..0x9E2B7` 先把 `#0` 畫在 `(0,0)`。現行
  `TURNSUM.LBX` 自繪面板不是原版主要資產。
- **已證實**：`Draw_Generic_Score_ = sub_9E207 @ 0x9E207..0x9E27A` 以 x=`0x8C=140`
  與 record 內的動態 y 繪製分數列，並以實際字高加 2 更新下一列。`Draw_Time_Score_`
  使用 x=`0x8C`／`0x8B`；其他逐項 helper 共同走 `sub_9E207`。原版不是 remake 的固定
  24px row step，但兩者都是有界的逐列版面。
- **已證實**：`sub_9D7EA @ 0x9D7ED..0x9D802` 建立 `(0,0)..(639,479)` 的輸入欄，
  動作字串為 raw `0x1B`。這證實整張畫面是退出／繼續控制，不只一顆 100×24 按鈕。
- **已證實**：`sub_9E27A` 依固定順序呼叫時間、generic 人口、淘汰、俘虜殖民地、科技、
  獵戶座、議會、安塔蘭與總分 helper。得分公式另由既有 `internal/gamedata/score.go` 證據
  管理；本輪不以 UI 反推或重寫公式。

## Remake 映射與未知

1. 有合法 `SCORE.LBX#0` 時優先使用原版 640×480 背景；缺檔、錯誤 count／尺寸、palette
   或解碼失敗時，才退回現有 `TURNSUM.LBX` 自繪面板。
2. `#1..14` 的逐張人口／種族圖選擇牽涉 captured-colonist array 與 race icon mapping；
   本輪只記錄資產形狀，未閉合 writer→consumer，不把它們任意配給九個分數類別。
3. 原版背景烘有 `HALL OF FAME`。英文模式保留原字；繁中模式只覆蓋中間文字牌，外部文案
   顯示「名人堂」，不可整片蓋掉雕像或框架。
4. 勝負標題、摘要與可見繼續提示是 remake 增補；保留是為了說明對局結果，但必須使用外部
   文案及安全框，不冒稱為原版逐像素文字。
5. 原版整張畫面接受退出／繼續；remake 應同步接受畫面任意點擊。Win95／輸入 helper 內部
   不深挖，停止於玩家可見輸入契約。

## 驗證

- catalog 雙語鍵與 `ScoreLine` typed key 測試；`hiscore.go` 無 `.tr(`、固定雙語玩家句及
  直接字型繪製。
- `SCORE.LBX` count／`#0` 尺寸與缺資產 fallback 測試。
- 標題、摘要、九列分數及提示的 runtime bitmap-font 雙軸 containment 測試。
- Docker＋Xvfb 分別用正版資料與缺 `SCORE.LBX` 的既有包跑 35 張中文畫廊，人工抽查
  `11_hiscore.png`；單元測試只證 remake 自洽，不取代原版畫面證據。

