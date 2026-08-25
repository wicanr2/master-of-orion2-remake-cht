# 武器成本、佔格與微型化 IDA 稽核（2026-08-24）

## 證據契約

- 輸入：`Orion2.exe`，SHA-256 `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- IDA database：`Orion2.exe.i64`，SHA-256 `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`
- 工具：IDA Pro 9.4（SDK 940）；位址皆為 IDA linear。
- 探針：`tools/ida/audit_weapon_cost_space.py`。探針不改名、不套型別、不寫回 database。

## 已證實

- `sub_6B519 @ 0x6B519` 的成本比例依微型化等級為 `100/75/55/40/30/25%`，整數截斷，正成本保底 1。
- `sub_6E60E @ 0x6E60E` 的一般武器佔格比例為 `100/80/65/50/35/25%`，以千分比計算並加 500 四捨五入，正佔格保底 1。
- `sub_6D048 @ 0x6D048` 先讀解鎖科技所屬主題，再從 `OrigTopicNext` 的下一個主題開始計算連續完成數；故剛解鎖是 0 級。
- `sub_6A636 @ 0x6A636` 先套 HV／PD；`sub_6A406 @ 0x6A406` 再加總其餘改造百分比。兩階段是相乘關係，不是所有百分比一次相加。
- `word_17FD15/word_17FD17` 所在 15-byte 記錄表給 AF 的成本與佔格均為 50；`Weapon_Cost_ @ 0x6EC74` 與 `Weapon_Space_ @ 0x6EE8E` 均除以 100 套用。AF 是 `+50%`，不是固定 `+50`。
- 同表門檻：HV／PD／EMG 為 0；AP／CO／NR／SP／ECCM／ARM／FST／OVR 為 1；AF／ENV／MIRV 為 2。

## 邊界

- 後續勘誤（2026-08-24）：飛彈 2/5/10/15/20 彈架容量已由獨立 IDA 切片閉合並接入完整 UI；
  證據移至 `missile-ammo-rack-audit-20260824.md`，本句不再代表現況。
- 後續勘誤（2026-08-24）：`sub_6F11C` 三種 raw category 與 `sub_6D048` 八個 Hyper byte
  已由獨立切片閉合並接線；舊的「只接 category 0／沒有重複等級欄位」不再代表現況。
  見 `weapon-mini-categories-hyper-audit-20260824.md`。
