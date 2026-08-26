# Show GNN Report 設定稽核（2026-08-27）

## 證據契約

- 輸入：`Orion2.exe`，SHA-256 `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- IDA 資料庫：`Orion2.exe.i64`，SHA-256 `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- 工具：IDA Pro 9.4／IDAPython；位址均為 DOS/4GW 映像的 IDA linear address。
- 可重生匯出：`tools/ida/audit_game_menu_popup_ui.py`。raw 函式名、位址、指令 bytes 與 caller 保留於匯出，本文只附加分級語意。

## 已證實

- 設定 byte 是 `byte_199BE5 @ 0x199BE5`。`sub_127E1 @ 0x127E1` 在 `0x12833` 寫入預設 1；`sub_7EFEF @ 0x7EFEF` 在 `0x7F0DF` 讀入設定頁；`sub_7F14C @ 0x7F14C` 在 `0x7F1B8` 回寫。
- `sub_21371 @ 0x21371..0x21AF8` 是事件分派的大型 switch consumer。它在 `0x219B2`、`0x21A58` 與 `0x21AB8` 比較 `byte_199BE5 == 1`，只有開啟時才呼叫 `sub_203CB`／`sub_20400` 等 GNN 畫面鏈；關閉分支仍於 `0x21ACB..0x21AEF` 建立並交付 `JIMTEXT.LBX` 文字資料。
- `sub_21B6D @ 0x21B6D..0x2223C` 是同一事件文字組裝鏈。`0x2217C` 依設定選擇欄位值 `0x78` 或 `0x0A`，證明關閉設定不是丟棄事件，而是改變其報告呈現資料。
- 官方 `help.json` 對應條目明寫：開啟時中斷一般報告並顯示 GNN 畫面；關閉後，特殊事件仍會在一般回合摘要通知。這與上述「略過 GNN 畫面但仍建立文字」控制流一致。

## 強推論與未知

- `0x78`／`0x0A` 欄位的原版完整結構名稱尚未閉合；它們只用來支持「同一事件文字仍流向非 GNN 報告」，不在 remake 中照抄 raw 結構。
- `sub_8B17B`／`sub_8B956` 也讀取設定以控制畫面／回合流程，但其所有狀態 byte 尚未逐項命名。remake 依已證實玩家契約實作，不為舊 UI state machine 逐指令翻譯。

## Remake 映射

- `ShowGNNReport=true`：結算產生 `LastEventReport` 時先開 GNN 快報，再依設定進回合摘要。
- `ShowGNNReport=false`：不開 GNN 快報；事件文字保留在既有回合摘要。即使 `EndOfTurnSummary=false`，本回合有特殊事件仍必須開摘要，否則會違反原版 help 契約。
- 星系勘查 `LastDiscovery` 是玩家自家回報，不受 GNN 設定抑制；事件與勘查同回合且 GNN 關閉時，先顯示勘查回報，事件仍留在摘要。

