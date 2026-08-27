# NPC 外交事件記憶 `+0x64F／+0x65F／+0x71F` 稽核（2026-08-28）

## 證據身分

- 輸入 `Orion2.exe` SHA-256：`7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- 正式 `Orion2.exe.i64` SHA-256：`4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`；
  IDA Pro 9.4 `ida-pro-9.4-idapython:locked-v1` 只分析 `/tmp` 副本。
- 位址均為 IDA linear EA；函式原名、邊界、bytes、運算元與全域相對欄位站點見
  [`evidence/npc-treaty-negotiations-ida-20260827.json`](evidence/npc-treaty-negotiations-ida-20260827.json)。

## 已證實：舊暫名錯誤

舊文件把 `+0x71F` 暫稱 treaty-break writer，證據不足。`Change_Relations_ @ 0x4E3B5`
實際在新變動絕對值大於舊值時，把 reason 寫入方向 byte `+0x64F`、套完政府／性格／難度後的
關係變動幅度寫入方向 word `+0x65F`。`sub_252D5 @ 0x252D5..0x2552D` 每回合消費這兩欄，才更新
byte `+0x71F`。因此三欄應稱「待處理 reason／幅度／重複事件記憶」，不是單一解約計數器。

負面幅度使用政府表 `word_180CDC @ 0x180CDC`：`40／30／20／5／0／-10／50／0`，門檻為
`table[government] - 2*magnitude`。government 4 且 `+0x727==1` 時改用
`word_180CE8 @ 0x180CE8 = 50`。有正式 policy 1..3、貿易、研究或負納貢狀態時，事件記憶固定
加 2；否則骰值低於門檻加 1，失敗清 pending reason。正面幅度在骰值小於 `2*magnitude` 時讓
記憶減一（最低 0），否則清 pending reason。更新後 `+0x71F` 鏡射至反方向。

`sub_5232E` 締約、`sub_52049` 納貢、`sub_51078` 宣戰與 `sub_524FB` 停戰都會把雙向
`+0x71F` 清零。NPC 談判 `sub_2552D` 對 inner 與每個第三方檢查此欄，正值就為 base 加 5。

## Remake 狀態與限制

- `OriginalNPCIncidentMemoryStep` 已承接上述一般政府分支與亂數順序；非法資料失敗即關閉。
- `GameSession` 保存三張方向矩陣，按原版 ordered pair 推進、鏡射記憶、供談判第三方 +5 消費，
  並支援 JSON 往返及熱座索引壓縮。
- AI↔AI 納貢的 reason 14 正面關係變動已寫入 pending；事件 4／5 在現行對稱戰爭路徑會依
  `Change_Relations_` 早退，因此不會偽造 pending。
- `+0x727` 已由 42 個全域 operand 站點確認：`sub_4D78E` 初始化清零，之後沒有 clearer；
  `sub_5138E` 在 actor 破壞既有正式條約時，寫入 target→actor 方向。一般 AI 宣戰現先執行這個
  writer，再改成戰爭政策。事件記憶的 government 4 特殊門檻已改讀該方向旗標。
- `word_18105C @ 0x18105C` 的締約／納貢 cooldown 政府表為
  `5／10／20／5／50／40／5／0`。government 4 有 `+0x727` 時改走索引 6；互不侵犯加 1.5 倍，
  同盟與納貢加 2 倍到 `+0x72F`。remake 已保留原版 signed-byte 寫回及雙向各自政府／旗標。
- **尚未閉合**：其他 Change_Relations_ reason 的 AI↔AI 可達 caller、`sub_4EB06／sub_4F0DC／
  sub_533F4` 其餘 `+0x727` writer 的玩家回應路徑，
  以及原版 `.GAM` 外交矩陣匯入。這些留白不影響已接的一般記憶公式，但目前仍不可宣稱所有
  外交 incident writer 完整 parity。
