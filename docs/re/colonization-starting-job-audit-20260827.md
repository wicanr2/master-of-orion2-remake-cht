# 殖民地起始職務稽核（2026-08-27）

## 問題與證據身分

remake 目前把所有新殖民地人口都設為農夫；需確認原版是否依行星或種族改變第一位殖民者職務。

- 輸入：`Orion2.exe`
- SHA-256：`7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- 工具：IDA Pro 9.4／IDAPython；位址均為 IDA 線性位址
- 原始函式：`sub_E5EB3 @ 0xE5EB3..0xE6071`
- 匯出：`docs/re/evidence/colonization-ida-20260827.json`

## 定位勘誤

外部符號曾把 `Colonize_Planet_` 放在 `sub_BB082 @ 0xBB082`。該函式只有 65 bytes，僅在
planet record 尚無 colony index 時配置暫存 colony record，直接 caller 是畫面函式
`sub_C0B87`；它不是完整殖民規則。真正建立殖民地／前哨站的共用資料流是
`sub_E5EB3`，一般殖民薄包裝為 `sub_E6071 @ 0xE6071`，玩家正常 caller 為
`sub_8B2DE @ 0x8B2DE`。

## 已證實

1. `0xE5F6D` 將一般殖民地人口寫為 1，`0xE5F71` 取第一筆 packed colonist record。
2. `0xE5F74..0xE5FA7` 依序檢查：

   - `planet+0x0B == 0`；
   - `player+0x8B1 != 0`（Lithovore）；
   - `player+0x8B0 != 0`（Cybernetic）。

   任一成立即把 packed colonist 的 bit 7 設為 1；三者皆不成立則清除 bits 7–8。
3. packed colonist bits 7–8 是職務；值 0 為農夫、值 1 為工人。此格式已由
   `Make_New_Colony_Or_Outpost_` 的原住民分支及 `Colonist::load` 交叉驗證。
4. 因此一般殖民者在「行星沒有自然食物」或種族為 Lithovore／Cybernetic 時從工人開始，
   其餘從農夫開始。原住民特殊物產額外建立的三位人口仍固定為農夫。
5. `sub_8B2DE` 呼叫一般殖民 wrapper 後，把實際殖民船 ship record `+0x64` 寫為 5、
   遞減艦隊數與玩家船數；這證實殖民船由正常玩家路徑消耗。

## 強推論與停止線

- `planet+0x0B` 的完整欄位名稱仍以 raw offset 保留；它與 remake 的
  `ClimateFoodPerFarmer(climate) == 0` 在十種氣候表逐值對應，故以「自然食物為零」作 typed 映射
  是強推論，不把欄位直接改名成氣候。
- 本頁原先未追完的前哨站重用、`colony+0x14C`、`+0x123=-39`、完整資格與重算 callback，
  已於 2026-08-28 閉合：`+0x14C` 是 raw building 22 Marine Barracks，`+0x123` 是一次性
  玩家通知欄而 `-39` 是新 colony 類通知碼。完整證據見
  [`colonization-full-audit-20260828.md`](colonization-full-audit-20260828.md)。
