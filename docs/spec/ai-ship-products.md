# AI 艦艇與特殊產品規格

狀態：PARTIAL；raw -15 貨運艦隊已接，其他產品仍為 DRAFT，不宣稱完整 AI 造艦器完成。

## 已固定契約

1. AI 生產單位是殖民地；每座殖民地一次只能有一個產品與一份進度。
2. 帝國回合先重建艦艇／特殊單位摘要，再建立產品配額，最後由 `sub_D10EE` 把配額分派給
   合法殖民地。
3. `netCapacity = population/8 + industry - pollution`；原版比較使用整數與嚴格門檻。
4. raw `-15` 是貨運艦隊，成本 50 PP，完工增加 5 艘貨運艦；它不建立 `Ship` 記錄。
   配額為 `ceil(max(0,movingColonyShips-surplusFreighters)/5)`；同輪最多分派兩座殖民地，
   且要求 Freighters 科技、未封鎖及 `netCapacity>=12`。
5. raw `-7` 成本 100 PP，但產品身分與 callback 尚未閉合，不得實作成前哨船。
6. raw `-12/-17/-11` 暫只保存 raw code；名稱、修正後成本與 callback 閉合前不得由成本猜測。

## 實作順序

1. （已完成）`sub_CFCB6 → sub_CF3BD → sub_D10EE(case 1) → sub_E36DF` 的 `-15`
   貨運艦隊垂直鏈。
2. 追回 `sub_1026CF → sub_CF40D → sub_D10EE(case 3) → sub_E36DF` 的 `-7` 完整身分，
   再取代固定週期 Spy／Agent fallback。
3. 逐一閉合 `-12/-17/-11` 的產品字串、成本修正與完工 callback，再接殖民船／前哨船／
   運兵船配額。
4. 最後處理戰鬥艦 role、改裝與多艦隊分派；不讓後者阻塞前三條玩家可見產品鏈。

## 驗收

- 每個產品都有 raw code、quota producer、殖民地 gate、成本、完工 callback 與 typed consumer。
- 生產中途存檔／讀檔不換產品、不重置進度、不重複完工。
- 貨運艦、支援艦與戰鬥艦維持不同資料模型；快速與戰術戰鬥不載入貨運艦隊。
- 缺 raw profile 的舊存檔走明示 fallback，不把未知欄位補成零後宣稱 exact。
