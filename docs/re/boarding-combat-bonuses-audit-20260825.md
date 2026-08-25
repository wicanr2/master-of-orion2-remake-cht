# 登艦戰加成鏈稽核（2026-08-25）

## 輸入與工具

- 輸入：`Orion2.exe`
- SHA-256：`7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- IDA 資料庫 SHA-256：`4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`
- 工具：IDA Pro 9.4，處理器 `metapc`
- 位址空間：IDA linear；DOS/4GW LE 載入後位址
- 非破壞性探針：`tools/ida/audit_boarding_combat.py`

本文件保留原始 `sub_*` 名稱與 linear 位址；外部符號名稱只作導覽，不取代原始定位。

## 已證實

### 艦員登艦加成

`sub_35CAD @ 0x35CAD`（外部索引：`Crew_Boarding_Combat_Bonus_`）以 combat ship index 為輸入，乘上記錄大小 `0x139`，讀取戰鬥艦記錄 `+0xAC` 的艦員等級，再索引 `0x17D174 + level*8` 的第一個 word。有效等級 0..4 得到：

| 艦員等級 | Bo |
|---:|---:|
| 0 | 0 |
| 1 | 5 |
| 2 | 10 |
| 3 | 15 |
| 4 | 20 |

表後資料屬相鄰結構，不得延伸成額外艦員等級。

### AI 登艦評估的完整可見加成鏈

`sub_2C129 @ 0x2C129`（`Boarding_Action_Type_`）在 `0x2C37F..0x2C3C6` 組合兩方數值：

- 攻方：`Get_Fleet_Commando_Bonus_(攻方 owner)` + `Crew_Boarding_Combat_Bonus_(攻方 ship)` + 攻方 boarding info。
- 守方：`Get_Fleet_Commando_Bonus_(守方 owner)` + `Get_Fleet_Security_Bonus_(守方 owner)` + `count(special 0x1D)*20` + `Crew_Boarding_Combat_Bonus_(守方 ship)` + 守方 boarding info。

因此攻守雙方都消費艦員 Bo 與艦隊 Commando；Security 與保安站只加在守方。保安站特殊元件 ID `0x1D` 每個給 `20`。

### 艦隊技能取最大值，不疊加

- `sub_35FDA @ 0x35FDA`（`Get_Fleet_Commando_Bonus_`）掃描同 owner 的有效戰鬥艦，排除 `+0xB0 != 0` 或 `+0x24 != 0`，只讀有軍官索引（`+0xAD >= 0`）的艦艇，保留最大 Commando 值。
- `sub_360AA @ 0x360AA`（`Get_Fleet_Security_Bonus_`）使用相同掃描形狀並保留最大 Security 值。
- `sub_36036 @ 0x36036`（`Commando_Bonus_`）：一般階 `2*(level+1)`，進階階 `3*(level+1)`。
- `sub_36106 @ 0x36106`（`Security_Bonus_`）：一般階 `2*(level+1)`，進階階沿共用分支得到 `3*(level+1)`。

這與 `gamedata.LeaderSkillBonus` 的 common Commando 與 captain Security 表一致。

## 強推論與未知

- **強推論**：`sub_EC767 @ 0xEC767`（`Get_Boarding_Info_`）建立雙方各 `0x37` bytes 的地面戰資料並呼叫 `sub_EC15C`、`sub_EC3CE`，是種族／科技陸戰隊數值進入登艦鏈的來源；其 owner/player 參數證明兩方資料必須分開取得。
- **未知**：兩個下游 helper 每個 raw 欄位的完整名稱尚未逐欄證實。本 remake 已有經獨立 RE 錨定的地面戰科技、種族與 hits-to-kill 規則；為避免重開無限研究，本切片只要求登艦兩方各自使用其帝國的同一套 typed 規則。

## 被推翻的舊斷言

- 「守方可暫用玩家相同基礎戰力」錯誤：原版 boarding info 明確帶雙方 owner/player。
- 「Security 只看被登艦那艘船的軍官」錯誤：原版掃描同 owner 的有效參戰艦，取最高值。
- 「Commando 可由帝國全域領袖清單代理」不適用於艦戰：艦戰只看指派在有效參戰艦上的軍官。

