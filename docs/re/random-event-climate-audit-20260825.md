# 氣候改善隨機事件逆向稽核（2026-08-25）

## 證據身分

- 輸入：`Orion2.exe`
- SHA-256：`7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- 工具：IDA Pro 9.4／IDAPython
- 位址基準：DOS/4GW object #1 的 IDA 線性位址
- 非破壞性匯出：`tools/ida/audit_event_climate.py`
- 推論等級：下列目標選擇、氣候值與消費端回寫均為「已證實」。

## 事件適用殖民地

`sub_2230A @ 0x2230A` 在事件 ID 1 的前置檢查 `0x225C0..0x225D7`，以目標 player slot
呼叫 `sub_2310C @ 0x2310C`；回傳 −1 時事件不適用。

`sub_23D44 @ 0x23D44` 掃描全部殖民地 record，接受 owner 等於目標 player 且 raw `+6=0`
的正常殖民地。每遇一個候選便執行 `Random(candidateCount)==1`，因此是 reservoir sampling，
結果均勻分布於該玩家全部正常殖民地。

`sub_2310C`：

1. 最多 200 次呼叫 `sub_23D44`。
2. 取得殖民地後，以 `colony+2` 找 planet record，只有 `planet+8 < 8` 才接受。
3. 200 次未命中時，自殖民地 record 低索引起找第一個 owner 相符且 climate `<8` 的殖民地。
4. 沒有候選才回傳 −1。

原版氣候 enum 已由多處資料表證實為 Toxic=0 至 Gaia=9；所以事件接受 Toxic、Radiated、
Barren、Desert、Tundra、Ocean、Swamp、Arid，排除 Terran 與 Gaia。這不是一般 Terraforming
的一級轉換規則。

## 事件紀錄

`sub_2230A` case 1 在 `0x228E1..0x22903` 由選中殖民地 `+2` 取 planet index，寫入
`record.word+3`。事件保存的是 planet index，不是 remake 的顯示名稱或氣候階級。

## 效果消費端

`sub_206A2 @ 0x206A2` case 1：

1. `0x207AC..0x207C9` 以 `record.word+3` 找 0x11-byte planet record，保存舊 `planet+8`，
   再把該 byte 直接寫為 8（Terran）。
2. `0x207CD..0x207E0` 由 `byte_17D81C[new]-byte_17D81C[old]` 調整 planet `+0x0B`
   食物基值；該表是十氣候每農夫食物表。
3. `0x207E3..0x2081A` 由 planet record 的 colony index 呼叫 `sub_E2A70` 重算殖民地。
4. `0x2081F..0x2083E` 呼叫顯示更新 helper；它不是另一套玩法公式。

## Remake 對映與限制

- 依原版 reservoir／200 次重試／線性 fallback 消耗事件亂數，玩家、熱座與 AI 相同。
- 同時更新 `Planet.ClimateID/Climate` 與 `ColonyState.Climate`，並透過既有氣候變更 helper
  重算食物與人口容量，不能只改其中一層。
- 原版重算人口容量使用完整 planet／colony raw 狀態；remake 既有
  `TerraformPopMaxAfterClimateChange` 對已烘入建築加成的 `PopMax` 採比例重算，仍屬已記錄近似。
  本切片修正事件目標與 Terran 結果，不把該既有跨系統近似冒稱逐欄完全相同。
