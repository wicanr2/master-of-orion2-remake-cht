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

## 強推論與未知

- **強推論**：`-12/-17/-11` 對應殖民船、前哨船、運兵船三項支援產品；目前仍缺
  各自完工 callback 與字串指標的最後一對一閉合，不在程式中固定名稱。
- **未知**：`byte_1A7236` 的 status 1／2 子集對 `player+0x38` 配額公式的完整玩法名稱。
- **未知**：`byte_1A7285`、`sub_1026CF` 與 `-7` 的完整產品語意；不得用現行每 6／8 回合
  免費增加 Spy／Agent 的 remake 政策冒稱相同。
- **未知**：case 4 的 type 2 支援艦／raw 22 fallback、三組 quota 的全部難度與外交輸入，
  以及戰鬥艦 role 0..4 的完整選擇權重。

## 對 remake 的約束

- 保留現有逐殖民地產品資料形狀；不得再把全帝國空閒工業直接倍增後灌入單一造艦池。
- 下一個實作切片先閉合一個產品的「quota producer → 殖民地門檻 → 產品與進度 → 完工
  callback → typed 狀態 → 存檔與測試」。只解出產品碼或成本不算完成。
- 現行 `advanceAIShipProduction` 與免費週期 Spy／Agent 仍是明示 fallback；本輪證據不支持
  把它們升格為原版忠實。
