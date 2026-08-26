# End Of Turn Wait 設定稽核（2026-08-27）

## 證據契約

- 輸入：`Orion2.exe`，SHA-256 `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- IDA 資料庫：`Orion2.exe.i64`，SHA-256 `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- 工具：IDA Pro 9.4／IDAPython；位址均為 DOS/4GW 映像的 IDA linear address。
- 可重生匯出：`tools/ida/audit_game_menu_popup_ui.py`，保留 raw 名稱、位址、指令 bytes 與 caller。

## 已證實

- 設定 byte 是 `byte_199BDD @ 0x199BDD`。`sub_127E1 @ 0x127E1` 在 `0x127FB` 寫入預設 1；`sub_7EFEF @ 0x7EFEF` 在 `0x7F071` 讀入設定頁；`sub_7F14C @ 0x7F14C` 在 `0x7F164` 回寫。
- `sub_8AD82 @ 0x8AD82..0x8AF4D` 在 `0x8ADAF` 檢查設定。只有設定為 0，且本地輸入、玩家狀態、待處理畫面與其他中斷旗標均為零／允許時，才會在 `0x8AE05` 把 `byte_19C194` 設為 1，進入後續連續回合鏈。
- `sub_84E9D @ 0x84E9D..0x84F8E` 在 `0x84F2D` 檢查設定；當設定為 0、`byte_19C194` 已啟用且 `sub_124075` 回報輸入時，`0x84F49` 把 `byte_199F2D` 設為 1，形成玩家輸入中斷。
- `sub_83411 @ 0x83411..0x8354D` 在設定為 0 時，於 `0x834DE..0x834F8` 建立覆蓋 0..639、0..479 的全畫面輸入區。這與「點畫面任意處停止」契約一致。
- `sub_84555 @ 0x84555..0x849CB` 在設定為 1 時於 `0x845FC..0x8464B` 顯示星曆／回合資訊；設定為 0 時直接跳過該等待呈現。
- 官方 `help.json` 明寫：開啟時每按一次 TURN 只執行一回合；關閉時一次點擊會連續推進，直到發生有趣事情；回合摘要、GNN 報告或玩家點擊會停止。

## 強推論與停止線

- `byte_19C194`、`byte_199F2D` 是依讀寫角色附加的導覽語意，raw 名稱與位址保留；其完整舊 UI state machine 不逐欄命名。
- 原版連續回合的精確 wall-clock 間隔未由上述控制流給出。remake 採固定 15 Ebitengine tick 的可中斷節奏，標為介面 timing approximation，不追 Win95 輸入／timer 內部。

## Remake 映射

- 單人局且設定開啟：TURN 只推進一回合。
- 單人局且設定關閉：第一次 TURN 正常推進一回合；若沒有研究選擇、勝負、GNN／勘查或應顯示的摘要，回到星圖並每 15 tick 再推進一回合。
- 任意滑鼠按鍵會停止連續回合且不把同一次點擊傳給下層星圖熱區。
- 熱座與網路鎖步不自動連續推進，避免繞過其他玩家交令；這是現代多人安全邊界。

