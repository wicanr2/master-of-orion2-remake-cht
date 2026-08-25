# 真人開局科技 profile 與正式新局順序稽核（2026-08-25）

## 原版證據

- 輸入：`Orion2.exe`，SHA-256 `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- 工具：IDA Pro 9.4／IDAPython；IDA 線性位址，DOS/4GW LE object #1
- `sub_12983 @ 0x12983`：非真人分支設定種族、複製特性並轉成 runtime 真值；
  真人 `[player+0x28] == 100` 走 `0x12C1E`，保留新遊戲畫面已設的種族特性。
- `sub_589D6 @ 0x589D6`：仍對每位玩家建立 6／4／7 三組權重；寫回時只對
  raw6 做真人特判，`0x58E26..0x58E3D` 把真人 raw6 強制保留 100，但 raw4 與
  raw7 仍分別在 `0x58E93..0x58E9A`、`0x58EF5..0x58F01` 寫入。
- `sub_FC845 @ 0xFC845`：真人分支會覆寫 AI 初始 category 權重，但共用後段仍
  消費 raw4 的對手類別分支。因此真人開局不能永遠把 raw4 標成 unknown。

## remake 程式順序稽核

**已證實不符**：正式 UI 路徑原本是：

```text
SetupNewGame → applyStartingTech → ApplyRace / ApplyCustomRaceBonuses → ApplyGovernment
```

先進級十九次抽選因此看不到玩家最後選定的種族與政府。另外，`ApplyRace`
只套數值加成，沒有從 `OrigRaceTraits[TRAIT_GOVERNMENT]` 套內建種族政體，導致
人類、克拉肯等都沿用 demo 獨裁。

## 修正後垂直鏈

1. `SetupNewGame` 保留一次 provisional 開局科技，供直接 API caller 相容，同時設置只存在於記憶體的 pending 旗標。
2. 內建種族從 31 格表套政體；客製種族保存數值 1..9 與布林 10..30 的完整 runtime 陣列。
3. 種族與政府完成後，只在 pending 狀態清掉 provisional research maps，以真實特性重建 1／6／25 次開局科技。
4. 真人也執行 raw27 與 raw6／4／7 抽選；raw6 仍由人類估值分支取代，raw4 送入共用後段。
5. pending 清除後的遊戲中政府變更不得重置已研究科技或進度。

## 證據邊界

- 客製種族目前沒有獨立的原版外觀／基礎種族索引，無法還原真人 `+0x27`
  的來源；remake 對客製種族以 raw27=0 建七項表。raw27 只改七項 raw2 權重，
  而真人 `sub_FC845` 會在後面覆寫這層初始值；raw4 與後續科技抽選的 RNG 消費次數不受影響。
- 開局亂數仍是 remake 可重播流，不宣稱與原版 PRNG 逐位元相同。
