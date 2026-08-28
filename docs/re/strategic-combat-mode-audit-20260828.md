# 戰略戰鬥模式旗標與戰機駐防 120 分支稽核（2026-08-28）

## 問題與結論

`Fighter_Garrison_Strength_ @ 0x5F64C` 的舊稽核已算出公式分支，卻沒有證明
`byte_199CB4 == 1` 時固定回傳 120 的模式語意。IDA 的全域讀寫端證實：

- **已證實**：`byte_199CB4` 是選擇 `Strategic_Combat_ @ 0x40148` 或格子戰術解算
  `sub_47939` 的全域模式旗標；值 1 走前者，值 0 走後者。
- **已證實**：戰機駐防在戰略解算模式固定貢獻 120；40／0、40／24、32／24 權重公式是
  值 0 時使用的分支，不是「一般戰略模式公式」。
- **已證實**：原版把此旗標納入新遊戲設定、存檔、多人同步及科技／設計合法性判斷。
- **remake 差異**：目前只有權重公式，沒有對應的全域模式選項，也沒有固定 120 分支。

這項結論只關閉戰略強度模式。Fighter Garrison 在格子戰術中如何建立逐架單位，仍是獨立
玩家可見證據鏈。

## 證據契約

- 輸入：`Orion2.exe`
- SHA-256：`7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- IDA 資料庫 SHA-256：
  `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`
- 外部符號索引 SHA-256：
  `f83116049964c14072547dd087a6316117cf62808c328cc7c087984467030b28`
- 工具：IDA Pro 9.4／IDAPython；位址基準為 IDA linear、DOS/4GW LE object #1
- 探針：`tools/ida/audit_strategic_combat_mode.py`
- 非破壞性證據：
  `docs/re/evidence/strategic-combat-mode-ida-20260828.json`

探針保留 52 個直接 operand site、39 個 owner 函式、原始名稱、位址、bytes、指令、周邊視窗、
caller 與外部導覽名稱。正式 `.i64` 唯讀掛載後複製到 `/tmp`；沒有改名、套型別或寫回資料庫。

## 模式選擇（已證實）

`Russ_Combat_ @ 0xE7343` 先把旗標複製到區域值，再依真人／AI 與多人條件調整：

- 區域值非零時，`0xE749C` 呼叫 `Strategic_Combat_ @ 0x40148`。
- 區域值為零時，`0xE75FA／0xE7651` 呼叫 raw `sub_47939 @ 0x47939`；該分支帶有真人／網路同步與
  戰術畫面控制流程。

因此 `byte_199CB4 == 1` 可非破壞性命名為「選擇戰略戰鬥解算的模式旗標」。這個語意來自
同一 selector 的兩個 consumer，不只來自外部符號名稱。

`Set_Default_Game_Settings_ @ 0x127E1` 在 `0x128B3` 寫入 1。`Newgame_Screen_ @ 0xCD435`
把畫面暫存值設為 `byte_199CB4 == 0`，確認 UI toggle 與底層 strategic flag 互為反相；接受設定
時再於 `0xCD6B6／0xCD87D` 寫回反相值。本切片沒有以 JIMTEXT 證實該 toggle 的正式英文標籤，
所以文件只陳述控制流語意，不冒稱精確 UI 字句。

## 持久化與多人同步（已證實）

- `Load_Game_ @ 0x10E2F` 讀取 553-byte game-settings block，最後以 `v51[216]` 寫回
  `byte_199CB4`；依 `0x199BDC + 216 = 0x199CB4`，它是該 block 的直接欄位。
- `Main_Receive_Message_ @ 0xF5A9F` 可由網路訊息寫入旗標。
- `Broadcast_Game_Data_Differential_ @ 0xF74CD` 與
  `Decode_Broadcast_Game_Data_Differential_ @ 0xF79F6` 都把旗標位址納入差異同步。
- `Start_Net_Screen_ @ 0xFB7E5` 讀取並傳送此值。

`fopen`、`fread`、`fclose` 及網路傳輸 helper 的內部實作均屬 C runtime／平台邊界，不納入
remake 或 RE 分母；這裡只保存「設定可存讀、可同步」的玩家可見契約。

## 戰機駐防分支（已證實）

`Fighter_Garrison_Strength_ @ 0x5F64C` 在 `0x5F67E` 比較旗標：

- 值 1：固定回傳 120，不讀玩家戰機科技、武器傷害或裝甲。
- 值 0：依已知科技選攔截機／轟炸機／重戰機權重，取最佳 beam、bomb 與 armor，計算
  `min(64000, (max(beam-armor,0)*beamWeight + max(bomb-armor,0)*bombWeight)/2)`。

直接 caller `Colony_Ground_Strength_Vs_Player_ @ 0x5F747` 只在 colony `+0x165` 非零時加入
此值。相同函式對另一殖民地防禦在旗標值 1／0 時分別使用 200／500，交叉支持這是一整組
戰略模式縮尺度，而不是戰機特例。

## 其他 consumer 邊界

39 個 owner 涵蓋 `Strategic_Combat_`、轟炸、戰略船艦設計、艦艇／衛星／行星防禦強度、
建造畫面、科技合法性、戰鬥回寫與多人流程。這些直接讀端證實旗標是全局玩法模式，不代表
本文件已閉合每一個 consumer 的公式；各子系統仍由 parity matrix 自己的列判定。

編譯器 helper、`memset`、C 檔案函式與網路平台內部只作邊界定位，不逐行判讀，也不計入
RE 知識庫完成分母。

## 證據分級與剩餘未知

- **已證實**：旗標的兩條戰鬥 selector、預設值、UI 反相寫回、存檔欄位、多人同步與
  Fighter Garrison 120／公式兩分支。
- **強推論**：文件中的「戰略戰鬥模式旗標」是依 selector 與 `Strategic_Combat_` 命名的工程
  語意；原版畫面上的正式英文設定名稱尚未由文字資產交叉確認。
- **未知且另案**：Fighter Garrison 的格子戰術逐架建立、數量、位置及回寫。
