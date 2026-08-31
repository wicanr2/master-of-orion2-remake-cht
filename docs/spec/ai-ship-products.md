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
2. （部分完成）已依 `sub_D10EE(case 3) → sub_E36DF` 接上 100 PP、未封鎖、
   `netCapacity>=15` 與 self Agent callback，並移除免費週期 Agent fallback。原版
   `sub_1026CF → sub_CF40D` 是逐對手 packed pair 配額；remake 目前只有帝國總池，配額暫以
   真人外派 Spy 壓力投影，證據等級為強推論。生產中 `ProductKind／Progress／Cost` 已通過
   snapshot 往返，且缺 player slot、封鎖、`netCapacity=14/15` 均有失敗即關閉邊界測試。
   免費週期 Spy 是另一個進攻配置缺口，未混報完成。
3. （部分完成：Colony Ship）raw `-12` 已使用 typed `ai_colony_ship` 產品保存成本與進度；
   只有已知 Colony Ship application、沒有既有／生產中殖民船且存在合法擴張候選時建立一份
   配額。完工建立真正 `RawType=COLONY_SHIP` 的 Ship，通過 snapshot 往返後可由既有
   `aiLaunchColonizationFleet → aiExpand` 航線抵達、建立殖民地並消耗。原版 ship slot 造價
   尚未在 remake 同構，現沿用玩家支援艦的 120 PP 明示估值；單主力艦隊也會把新船加入目前
   艦隊，兩者均不冒稱原版精確。Outpost Ship／Transport 尚未接，因目前缺 AI 前哨站與逐艦
   地面運輸 consumer，不建立無下游的假完成 Ship。
4. （部分完成：case 0 戰鬥艦）逐殖民地 `ai_combat_ship` 產品已保存 role 0..4 持久藍圖的
   深層快照、typed 造價與進度；科技更新不會改寫生產中的武器、特殊裝備或成本。原版多藍圖
   生產評分仍未知，因此目前採可重播近似：補實艦數量最少的 hull role，平手取較低 hull，
   不冒稱原版精確 selector。完工建立真正 `RawType=COMBAT_SHIP` 的 Ship，並同步艦隊戰力與
   指揮點；快速／格子戰術沿用既有 typed Ship consumer。舊存檔的全帝國
   `ShipBuildProgress` 只遷移一次到第一個 ship slot；正常 AI 回合已停止使用全帝國造艦池。
   case 2 改裝、軌道基地與多艦隊分派仍待後續切片。

## 驗收

- 每個產品都有 raw code、quota producer、殖民地 gate、成本、完工 callback 與 typed consumer。
- 生產中途存檔／讀檔不換產品、不重置進度、不重複完工。
- 貨運艦、支援艦與戰鬥艦維持不同資料模型；快速與戰術戰鬥不載入貨運艦隊。
- 缺 raw profile 的舊存檔走明示 fallback，不把未知欄位補成零後宣稱 exact。
