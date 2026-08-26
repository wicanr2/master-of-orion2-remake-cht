# 隨機事件 16 瘟疫稽核（2026-08-25）

## 證據契約

- 輸入：`Orion2.exe`，SHA-256 `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- 資料庫：`Orion2.exe.i64` 一次性可寫副本
- 工具：IDA Pro 9.4、`tools/ida/audit_event_plague.py`
- 位址：IDA linear，DOS/4GW LE object #1；靜態證據，未做動態 oracle

## 已證實

1. `sub_2230A @ 0x2230A` 的 `0x226CF..0x22712` 以 `sub_23DA0` 均勻抽一次殖民地，
   並拒絕已有互斥殖民地事件的目標；不會因候選不適用而在事件 16 內重抽。
2. 建立端 `0x22B96..0x22BE0` 保存目標殖民地，並令
   `remaining = colonyResearch × (Random(8) + 2×difficultyRaw)`。`sub_23B64 @ 0x23B64`
   直接讀 Colony `+0xEB`，既有殖民地研究證據已確認該欄是當回合 `TotalResearch`。結果小於 1
   時事件建立失敗。
3. `sub_206A2` case 16 `0x20E0D..0x20E67` 在公告期後，每回合從 remaining 扣除當下
   `TotalResearch`、重算殖民地，remaining 小於等於 0 時進入結束狀態。因此研究為 0 時治療
   不會自行前進。
4. `0x20E6E..0x20E9B` 在 age > 4 後每回合以 `Random(20)==1` 進入 status 4；下一輪共同入口
   又把 status 4 轉回 active status 2。這是中途播報狀態，不是解除瘟疫。
5. `sub_E1839 @ 0xE1839` 以 `sub_234B8 @ 0x234B8` 查詢事件 16；命中時
   `0xE1C0F..0xE1C15` 從逐族成長百分點扣除 `200`，再交給原有 signed growth 與人口回寫鏈。
   原版不是固定扣兩人口。

## Remake 對映與限制

- 事件以 `PersistentPlague` 保存目標行星、研究需求、已完成研究與 age；玩家、熱座及 AI
  經濟結算都在暫態殖民地副本套 `GrowthBonusSum -= 200`，並以該回合實際 Research 推進治療。
- `sub_23DA0` 額外要求 `colony+0x13F==0`；該 raw 欄位已證實為 Capitol，remake 尚未把排除
  條件接入事件抽選器，保留為明示近似。
- 原版完整互斥表含事件 2、14、16、17、24、25；目前先對已具行星 record 的 16／17 做互斥，
  其餘待各事件垂直鏈重建後接入。
- status 4 的中途 GNN 文案變體尚未接；不影響負成長與研究解除的玩法效果。
