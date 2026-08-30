# AI 艦艇與特殊產品規格

狀態：READY

RE-TRACE: dos-orion2-1.31:0xD10EE

訂正標記：`CORRECTION-20260830-AI-PRODUCT-CODES`

## 已固定契約

1. AI 生產單位是殖民地；每座殖民地一次只能有一個產品與一份進度。
2. 帝國回合先重建艦艇／特殊單位摘要，再建立產品配額，最後由 `sub_D10EE` 把配額分派給
   合法殖民地。
3. `netCapacity = population/8 + industry - pollution`；原版比較使用整數與嚴格門檻。
4. raw `-15` 是貨運艦隊，成本 50 PP，完工增加 5 艘貨運艦；它不建立 `Ship` 記錄。
   配額為 `ceil(max(0,movingColonyShips-surplusFreighters)/5)`；同輪最多分派兩座殖民地，
   且要求 Freighters 科技、未封鎖及 `netCapacity>=12`。
5. raw `-7` 是成本 100 PP 的 Agent 訓練，完工加入 self Agent pool。
6. raw `-12/-17/-11` 是 Colony Ship／Outpost Ship／Transport 選單 pseudo-product；選定後
   轉成 `-(shipSlot+100)`，由共用 ship-record 完工鏈處理，不建立三個虛構固定 callback。
7. `sub_D10EE` case 4 在 Marine Barracks 已建時建 Transport slot，否則先建 raw 22；
   case 0／1／2／3／5 分別是戰鬥艦或基地／貨運艦／改裝／Agent／無動作。

## 實作順序

1. （已完成）`sub_CFCB6 → sub_CF3BD → sub_D10EE(case 1) → sub_E36DF` 的 `-15`
   貨運艦隊垂直鏈。
2. 依已閉合 `sub_1026CF → sub_CF40D → sub_D10EE(case 3) → sub_E36DF` 取代免費週期
   Spy／Agent fallback。
3. 依 pseudo-product → ship slot 契約接 Colony Ship／Outpost Ship／Transport 配額。
4. 依既有 role 0..4 藍圖、case 0 戰力缺額與 case 2 改裝鏈接線；多艦隊資料不足時明示近似。

## 驗收

- 每個產品都有 raw code、quota producer、殖民地 gate、成本、完工 callback 與 typed consumer。
- 生產中途存檔／讀檔不換產品、不重置進度、不重複完工。
- 貨運艦、支援艦與戰鬥艦維持不同資料模型；快速與戰術戰鬥不載入貨運艦隊。
- 缺 raw profile 的舊存檔走明示 fallback，不把未知欄位補成零後宣稱 exact。
