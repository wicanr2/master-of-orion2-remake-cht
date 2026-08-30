# AI 艦艇與特殊產品反組譯稽核（2026-08-28）

## 證據契約

- 輸入：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- 工具：IDA Pro 9.4；IDA 線性位址，DOS/4GW LE image。
- 非破壞性匯出：
  [`evidence/ai-ship-products-ida-20260828.json`](evidence/ai-ship-products-ida-20260828.json)。
- 匯出器：[`tools/ida/audit_ai_ship_products.py`](../../tools/ida/audit_ai_ship_products.py)。
- 下列語意均保留 raw 函式名、位址與產品碼；反編譯文字只供導覽，不單獨作證。

## 已證實

1. `sub_CFCB6 @ 0xCFCB6` 每輪重建帝國艦艇摘要：raw ship type 0 進戰鬥艦統計，
   type 1 另進 `byte_1A7234／byte_1A7236`，type 2 進 `byte_1A7274`。因此三類不可再
   壓成同一個抽象軍力池。
2. `sub_CF3BD @ 0xCF3BD` 以 `player+0x38` 與 `byte_1A7236[player*3]` 的差建立
   `byte_1A724C[player]` 配額；負差採 `(4-diff)/5`。兩個來源欄位的完整玩家語意仍須
   由 producer／consumer 閉合，配額公式本身已證實。
3. `sub_CF40D @ 0xCF40D` 另建立 `byte_1A7275[player*2]` 與
   `byte_1A7285[player*2]` 兩組配額；它們分別消費 type 2 艦數及
   `sub_1026CF` 的逐帝國加總，不能用性格固定週期取代。
4. `sub_D10EE @ 0xD10EE` 的候選 case 1 在殖民地未受特殊 gate 且
   `sub_CFEDC(colony)>=12` 時寫產品 `-15`，並扣 `byte_1A724C`；case 3 在門檻 15
   時寫 `-7`，並扣 `byte_1A7285`。`sub_CFEDC @ 0xCFEDC` 精確為
   `population/8 + industry - pollution`。
5. 真正的 `Colony_Product_Cost_` 是 `sub_E0DD6 @ 0xE0DD6`，不是舊外部符號曾指向的
   `sub_B206F @ 0xB206F`。後者只反覆移除同類建築／產品。已直接證實的固定成本包括：
   `-15 → 50 PP`、`-7 → 100 PP`；`-12` 的基礎成本為 500 PP，另經
   `sub_6E1A0` 修正。
6. `Apply_Production_` 所在函式實際邊界為 `sub_E36DF @ 0xE36DF..0xE3E9A`。
   產品 `-15` 完工會令 `player+0x36 += 5`；既有證據已閉合 `+0x36` 為貨運艦總數，
   所以 `-15` 是貨運艦隊，而不是戰鬥用運兵船或殖民船。
7. `sub_AFF9E @ 0xAFF9E` 的玩家產品清單把 `-12`、`-17`、`-11` 放在帝國支援科技
   gate 下；`-15` 與 `-7` 另走各自的 `sub_E11BC` 可建 gate。這直接否定「看到 100 PP
   就把 `-7` 當前哨船」的舊猜法。

## 2026-08-30 閉合勘誤

- `-12／-17／-11` 已由 `sub_AFF9E @ 0xAFF9E` 的唯一清單建立端閉合。三項分別受
  `player+0x140`、`+0x184`、`+0x1D4` 的 status 3 gate；以 application base `+0x117`
  回推 raw tech ID 為 41／109／189，受版控 enum 分別是 Colony Ship／Outpost Ship／
  Transport。`-11` 另要求 colony `+0x14C` raw building 22 Marine Barracks。
- 這三個負碼是產品選單的 pseudo-product，不是 `Apply_Production_` 的三條獨立完工 callback。
  選定後建立／選取 129-byte ship slot，殖民地產品改存 `-(shipSlot+100)`；完工由
  `sub_E36DF` 的共用 ship-record consumer 處理。舊待辦要求三個固定 callback 是資料模型誤判。
- `-7` 已由間諜回合專題閉合為 100 PP 的 Agent 訓練：完工先加入 self Agent pool；
  `sub_E36DF @ 0xE3B61..0xE3BC7` 的 `sub_10294B(owner,owner)`、packed pair 更新與
  `sub_B206F(...,-7)` 是 callback，不是免費 Spy／Agent 計時器。
- `sub_CFCB6` 的 type 2 已由地面入侵／卸載 consumer 證實為 Transport；
  `sub_D10EE` case 4 在 `netCapacity>=14` 時，Marine Barracks 已建就以 `sub_5663E` 建
  transport ship slot，否則先指定 raw building 22。這閉合了舊稱「type 2 支援艦／raw 22
  fallback」的語意。
- `sub_D10EE` 的六權重抽選已閉合為：case 0 戰鬥艦／軌道基地、case 1 貨運艦隊、case 2
  艦艇改裝、case 3 Agent、case 4 Transport／Marine Barracks、case 5 無動作。case 0 使用
  `dword_1A745C` 戰力缺額及既有 role 0..4 藍圖；艦體／role 建立器另由
  `ai-ship-blueprint-build-audit-20260825.md` 閉合。
- 三張 quota 的外交、難度與經濟輸入全部留在 `sub_CF3BD／sub_CF40D` 原始指令及
  `evidence/ai-ship-products-ida-20260828.json`；remake 是否已逐式接線屬 source 差異，
  不再列為 RE 未知。

補充可重生證據：`tools/ida/audit_support_product_conversion.py` 與
`evidence/support-product-conversion-ida-20260830.json`。工具、輸入雜湊及位址基準同本文件
證據契約；推論等級為上述 raw code、tech gate、ship slot 與 callback **已證實**。

## 對 remake 的約束

- 保留現有逐殖民地產品資料形狀；不得再把全帝國空閒工業直接倍增後灌入單一造艦池。
- raw -15 已接完整垂直鏈：以移動中殖民船與貨運艦餘額建立無條件進位配額；同輪最多
  分派兩座具 Freighters 科技、未封鎖且 `population/8+netIndustry>=12` 的殖民地；產品保存
  50 PP 進度，完工增加 5 艘貨運艦與版本相依現金回饋，不建立 `Ship` 或常駐建築。
- raw -7 的 RE 垂直鏈已閉合；remake 仍須以殖民地 100 PP Agent 產品取代免費週期 fallback。
- 現行 `advanceAIShipProduction` 與免費週期 Spy／Agent 仍是明示 fallback；本輪證據不支持
  把它們升格為原版忠實。
