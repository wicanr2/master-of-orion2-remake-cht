# NPC 方向國力矩陣規格

**狀態：CONFORMED（一般 AI 艦隊 producer 與外交消費端已接；特殊 owner 分支除外）**

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

`OriginalNPCShipPower`、`OriginalNPCShipDurability` 與 `OriginalNPCShipBeamAttack` 已實作並測試。
shell producer 逐艦接入 `sub_54E5B／sub_5679E／sub_5EAE9／sub_5F2F6` 的 typed 輸入：電腦、
可用戰鬥掃描器、Weaponry 軍官、艦員、種族艦攻、observer 引擎／種族艦防／跨維度、裝甲科技、
強化船體、重型裝甲與兩條獨立損傷。最佳戰機光束與炸彈亦依已知科技及原版嚴格最大值掃描。

## 接線與驗收

- NPC 條約納貢與一般宣戰 reason 23 都必須改讀同一矩陣。
- 測試至少涵蓋：同 hull 不同武器得到不同 power、觀察者防禦造成方向差、八槽累加、受損／船員
  修正、零艦隊 `+1` ratio、防止超過 800，以及存檔後重建一致。
- `FleetStrength` 投影只保留給非零純量但完全沒有實艦 raw 資料的舊存檔／精簡測試；矩陣另回傳
  `exact=false`，不可把 fallback 升格為原版精確值。
