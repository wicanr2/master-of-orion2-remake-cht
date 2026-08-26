# Serious Turn Summary 設定稽核（2026-08-27）

## 證據契約

- 輸入：`Orion2.exe`，SHA-256 `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- IDA 資料庫：`Orion2.exe.i64`，SHA-256 `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- 工具：IDA Pro 9.4／IDAPython；位址均為 DOS/4GW 映像的 IDA linear address。
- 可重生匯出：`tools/ida/audit_game_menu_popup_ui.py`，保留 raw 名稱、位址、bytes 與 caller。

## 已證實

- 設定 byte 是 `byte_199BE8 @ 0x199BE8`。`sub_127E1 @ 0x127E1` 在 `0x12848` 寫入預設 0；`sub_7EFEF @ 0x7EFEF` 在 `0x7F10C` 讀入設定頁；`sub_7F14C @ 0x7F14C` 在 `0x7F1DC` 回寫。
- 唯一玩法直接 consumer 是 `sub_FE0EA @ 0xFE0EA..0xFE250`，caller 為 `sub_FE63E @ 0xFE63E` 的 `0xFE751`。函式先在 `0xFE121` 檢查 End Of Turn Summary，再從 `dword_1AA414` 指向的 18-byte 記錄陣列逐筆掃描。
- `0xFE14C`、`0xFE157`、`0xFE162` 將三個預設允許旗標設為 `Serious==0`。Serious 開啟時，只有明列分支能把 `dh` 設為 1：raw type 0 的 subtype 4／6／7／8／9、raw type 4／5／7，以及滿足額外欄位條件的 raw type 9。其餘類型不會觸發摘要。
- 找到合格記錄且 caller 要求顯示時，`0xFE223..0xFE244` 設定畫面狀態並把該玩家 bit 寫入 `byte_1AB50F`；函式回傳狀態 4。這證實選項決定的是「是否收到整張回合摘要」，不是在已開啟的摘要裡逐行刪字。
- 官方 `help.json` 明列 serious 範例為飢荒、叛亂，以及因無 BC 支付維護而被迫報廢；並明寫只有摘要包含 serious reports 時才會收到摘要。

## 強推論與未知

- raw type／subtype 與全部原版報告名稱尚未建立完整表；本輪不把未命名編號硬套到 remake 欄位。
- remake 以官方已列舉且已有 typed 玩家結果的飢荒、叛亂、破產資產處分作最低充分對映；安塔蘭攻擊與敵方突襲同樣是已有結構化戰損／威脅結果，列為強推論的 serious 擴充。原版所有 raw 類型逐值 parity 仍未知。

## Remake 映射

- 設定關閉：沿用 End Of Turn Summary，正常顯示摘要。
- 設定開啟：只有 `HasSeriousTurnSummaryReport` 為真才顯示摘要；摘要一旦開啟仍顯示完整內容。
- GNN 關閉後被強制送入摘要的特殊事件優先於本篩選，避免事件消失；這是兩個已證實 help 契約的組合結果。

