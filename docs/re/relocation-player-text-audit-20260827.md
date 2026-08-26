# 集結點／遷移玩家流程與文案邊界稽核（2026-08-27）

## 問題

集結點玩法與畫線已可使用，但 `internal/shell/relocation.go` 仍把中文拒絕原因與怪獸確認句
當成規則回傳值，`cmd/moo2/relocation.go`／`interactive.go` 再以 `tr(中文,英文)` 組合提示、
按鈕與結果。這使規則資料、玩家文案與語言選擇互相纏繞，也無法證明文句與原版文字表的關係。

## 證據契約

- 輸入：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- IDA 資料庫：`Orion2.exe.i64`，SHA-256
  `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- 工具：IDA Pro 9.4／IDAPython；位址為 DOS/4GW 映像的 IDA linear address。
- 可重生匯出：`tools/ida/audit_relocation_ui.py`，保留 raw 名稱、函式邊界、caller、
  指令 bytes 與交叉參照。

## 已證實

- 起點 wrapper `sub_74F8A @ 0x74F8A..0x74FAA`、終點 wrapper
  `sub_74FAA @ 0x74FAA..0x75035` 共用合法性函式
  `sub_75035 @ 0x75035..0x75180`。完整兩段式互動是
  `sub_75180 @ 0x75180..0x7522B`，由 `sub_73980` 在 `0x73DF2` 呼叫。
- `sub_75035` 在 `0x7505D` 檢查光譜碼 6。依起點／終點方向分別選文字 ID
  `0x84／0x83`；在 `0x75098` 呼叫已探索 predicate，失敗時分別選 `0x86／0x85`。
- `0x750BA` 呼叫 `sub_7A47A` 的怪獸 predicate。終點分支在
  `0x750EE..0x75117` 取文字 ID `0x87`、填入星系與怪獸資料後呼叫 kind 1 使用者確認框；
  起點分支在 `0x7511B` 靜默回傳失敗。
- 起點額外在 `0x7512C` 檢查我方殖民地；失敗時取文字 ID `0x88` 並以星系資料組裝。
- 取消函式是 `sub_7522B @ 0x7522B..0x7527A`。全設函式
  `sub_785EC @ 0x785EC..0x7862B` 在 `0x7860B..0x78619` 只把原值不等於 `-1`
  的 `star+0x54+player×2` 欄位改成新終點；未設定的殖民地不會被順便設定。
- 清除全部是 `sub_77BB1 @ 0x77BB1..0x77BF1`。畫線函式是
  `sub_85320 @ 0x85320..0x853D3`，由三個星圖流程呼叫。

## 證據分級與文字停止線

- 上述合法性、方向差異、怪獸確認、取消及「只改既有設定」均為**已證實**。
- ID `0x83..0x88` 的精確英文內容未在本輪從正版文字資產逐字匯出；remake 的等義文句只標為
  玩家介面轉接，不冒稱原版逐字一致。
- 原版起點怪獸分支是靜默失敗；remake 保留明示提示，避免滑鼠介面看似失效。這是刻意且已揭露
  的可用性差異。
- 規則層應只回傳 typed 原因；星名、怪獸名與數量由 UI 以 `ui.json` 模板組裝。

## Remake 映射

- `CanRelocateFrom`、`CanRelocateTo`、`SetStarRelocation` 回傳 enum，不回傳任何語言句子。
- `RelocateToNeedsConfirm` 只回傳布林契約；星系與怪獸顯示名仍由既有 typed session 查詢。
- 星圖按鈕、艦隊操作列、flash 結果與確認框全部使用外部 catalog；規則／存檔／網路命令形狀
  不變。
