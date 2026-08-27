# NPC 方向國力矩陣規格

**狀態：READY（producer 控制流已閉合；remake 資料鏈施工中）**

證據：[`../re/npc-power-matrix-audit-20260828.md`](../re/npc-power-matrix-audit-20260828.md)。

## 必要資料形狀

- `AIPowerRaw[owner][observer]` 為非負整數方向矩陣；不可壓成每帝國單一純量。
- 每回合由目前實艦重算，不把衍生值當權威存檔；存檔後重建必須得到同值。
- 舊存檔只有 `FleetStrength` 時，先走既有逐艦相容轉換，再計算矩陣。
- 原版 `.GAM` 匯入必須無損保留 `Computer／Size／Armor／Shield／BaseCombatSpeed`、五個
  `DamagedSpecials` bytes，以及 shield／drive／computer／armor／structure damage 與 crew level；
  合法零值一律由 `Known` 旗標和缺欄區分。這條資料鏈已於 2026-08-28 完成並有 JSON 往返測試。

## Producer

1. owner／observer 以穩定 AI 索引掃描。
2. 每艘非支援艦從最多八個武器 mount 計算方向效能；武器 ID、數量、改造、命中、船員、損傷及
   observer 防禦輸入必須來自 typed 狀態，不得解析顯示字串。
3. 同 owner 的逐艦值累加至該 observer 欄；非法／無法映射的 mount 失敗即關閉，不以艦體值猜補。
4. `OriginalNPCPowerRatio` 只消費矩陣值，維持 800 cap 及來源第三方戰爭逐次折半。

`OriginalNPCShipPower` 的純算術層已於 2026-08-28 實作並測試；尚未完成的是 shell 對
`sub_54E5B／sub_5679E／sub_5EAE9／sub_5F2F6` 輸入的完整 typed producer 與回合接線。

## 接線與驗收

- NPC 條約納貢與一般宣戰 reason 23 都必須改讀同一矩陣。
- 測試至少涵蓋：同 hull 不同武器得到不同 power、觀察者防禦造成方向差、八槽累加、受損／船員
  修正、零艦隊 `+1` ratio、防止超過 800，以及存檔後重建一致。
- 完成前，`FleetStrength` 投影只能保留為舊存檔或無 typed mount 的明示 fallback；目前 producer
  尚未接線，因此本規格仍不可標為 `CONFORMED`。
