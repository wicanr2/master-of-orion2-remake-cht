# SABOTAGE 兩張計分表上游稽核（2026-08-25）

## 證據契約

- 輸入：`Orion2.exe`，SHA-256 `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- 工具：IDA Pro 9.4、IDAPython；位址均為 IDA 線性位址（DOS/4GW LE object #1）。
- 匯出器：`tools/ida/audit_sabotage_score_tables.py`。匯出保留原始函式名、位址、bytes 與 operand，未改名資料庫。
- 等級：以下直接算式與表值為**已證實**；領袖 raw status／技能旗標到 remake `Leader` 的欄位映射沿用既有領袖證據，屬**強推論**。

## 已證實結果

`Resolve_Spies_ @ 0x10192B` 以玩家索引乘 4，將 `word_1ACE78` 與 `word_1ACE7A` 的地址傳給
`sub_100A83 @ 0x100A83`。兩者因此不是靜態常數表，而是每位玩家交錯排列的 16-bit runtime
攻擊／防禦能力值；檔案中的初始 `FF FF` 不能代表回合中的分數。

`sub_100A83` 的共同基底為：

```text
base = technologyBonus + signed(player+0x8A8) + 10*signed(player+0x8B8)
attackTable[player]  = base + bestSpyMaster
defenseTable[player] = base + bestTelepath + governmentDefense[player+0x89F]
```

- `player+0x8A8 = player+0x89F+9`，對應 `TRAIT_SPYING`。
- `player+0x8B8 = player+0x89F+25`，對應 `TRAIT_TELEPATHIC`。
- `sub_100A3E @ 0x100A3E` 只在五個科技狀態 byte 等於 3 時加分：四項各 `+10`，一項 `+5`；與手冊的 Neural Scanner、Cyber Security Link、Stealth Suit、Psionics、Telepathic Training 表一致。
- `byte_100A36 @ 0x100A36` 的八個有號值為 `0,0,10,15,-10,-10,15,15`，與八種政府防禦加成逐格一致。
- 領袖掃描只納入同帝國且 raw status `<3`；Spy Master 與 Telepath 各保留最大有效加成。

`sub_101483 @ 0x101483` 的人數加成不寫入兩表。`sub_1014A4 @ 0x1014A4` 在任務消費端另讀 packed low-six-bit 人數並套用三段式 helper。因此 remake 可以把「帝國能力表」與「對手派駐人數」分開保存，只要最後在同一判定中相加。

## 勘誤與停止線

- 舊文件所稱「兩表以 `0xFF` runtime 初始化／上游未知」已被上述寫入鏈推翻。
- `sub_1018A3 @ 0x1018A3` 不是重設兩表；它在 `Resolve_Spies_` 前處理間諜相關狀態與隨機分支。其完整玩家語意不影響兩表算式，本切片不以猜測命名。
- AI 任務政策、訓練成本與維護仍是另一條玩家機制，不因本次閉合兩表而宣稱完整原版 parity。
