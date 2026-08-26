# 安塔蘭王座廳玩家流程與文案邊界稽核（2026-08-27）

## 問題

王座廳已可由艦隊列表進入，但畫面標題、戰力摘要、按鈕、阻擋原因與轉場仍內嵌於 Go；
`internal/shell` 甚至直接回傳中文阻擋句。現行戰力比較與「發動／撤退」兩顆按鈕也不可因可玩
就冒稱原版畫面。

## 證據契約

- 輸入：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- IDA 資料庫原始輸入：`Orion2.exe.i64`，SHA-256
  `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- 工具：IDA Pro 9.4／IDAPython；位址為 DOS/4GW 映像的 IDA linear address。
- 可重生匯出：`tools/ida/audit_antaran_room_ui.py`；保留 raw 名稱、函式邊界、caller、指令
  bytes、字串與資料交叉參照。匯出在可寫副本上執行，副本關閉後雜湊改變，不把它當原始資料庫
  雜湊。

## 已證實

- 原始函式邊界依次為 `sub_14AAC @ 0x14AAC..0x14BFD`、
  `sub_14BFD @ 0x14BFD..0x14C83`、`sub_14C83 @ 0x14C83..0x14D7C`、
  `sub_14D7C @ 0x14D7C..0x14DE1`。`sub_14AAC` 由 `sub_FE552` 在 `0xFE61A`
  呼叫，且依序呼叫其餘三支。
- `sub_14C83` 在 `0x14CA2` 與 `0x14D3A` 兩度取得 `antaroom.LBX`；後者以 raw asset
  參數 1 呼叫 `sub_12C607`。這與正版 `ANTAROOM.LBX#1` 的 640×480、55 幀結構一致。
- `sub_14D7C` 在 `0x14DBA` 取得 `antarmsg.LBX`，以
  `4 × input + 8 × byte_199CAE - 1` 形成 raw message 索引，再把所得字串複製到
  `dword_19A014` 指向的緩衝區。
- `sub_14BFD` 讀 `dword_19A014`，並在 `0x14C4A`／`0x14C66` 經兩個文字 helper
  消費該訊息。`sub_14AAC` 以整張 `0..639 × 0..479` 區域建立輸入欄，迴圈等待輸入後退出。

## 符號勘誤與證據分級

- 外部 `func_names.txt` 把名稱放在 `0x14BFD／0x14C83／0x14D7C／0x14DE1`；修正後的
  `symbols_fixed.tsv` 則把 `Main_Antaran_Room_Screen_` 放在 raw `0x14AAC`，其餘依序移到
  `0x14BFD／0x14C83／0x14D7C`。本文件以 raw 位址與 IDA 邊界為證據，不以衝突名稱取代定位。
- 原版滿版動畫、`ANTARMSG.LBX` 訊息與整張畫面輸入為**已證實**。
- 現行 remake 的戰力比較、阻擋原因、兩顆按鈕與按鈕座標是可用性介面轉接；現有靜態證據
  未證明原版有相同 widget，故只標 **remake adapter**。
- 原版 55 幀的精確停留時間仍未知；每 3 tick 一幀維持已揭露 timing approximation。

## Remake 映射

- `internal/shell` 只回傳 typed 阻擋原因，不保存玩家句子。
- adapter 的標題、說明、戰力、勝算、按鈕、阻擋原因與轉場全部由 `assets/i18n/ui.json`
  提供；原版 `ANTARMSG.LBX` 訊息尚未 typed 解碼，不假裝現行文案是逐字還原。
- 文字使用明確安全框；動畫與戰鬥規則、存檔及網路命令形狀不因文案遷移而改變。
