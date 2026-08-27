# 原版 AI 稅率寫入鏈靜態稽核

日期：2026-08-28

## 證據契約

- 輸入：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- IDA 資料庫 SHA-256：
  `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- 工具：IDA Pro 9.4；位址基準為 DOS/4GW LE image 的 IDA linear EA。
- 非破壞性匯出：
  [`evidence/ai-tax-rate-ida-20260828.json`](evidence/ai-tax-rate-ida-20260828.json)。
  匯出保留原函式名、位址、bytes 與上下文，沒有修改正式 `.i64`。
- 動態資料交叉驗證：既有原版新局自動存檔 `SAVE10.GAM`，SHA-256
  `ece2eb06d782078dd0a6f746020a05691355303ceb02bbfbbe2233e987272be1`；以
  `internal/save.Load` 唯讀解析，五名有效玩家的 BC 均為 50、稅率均為 0。

## 已證實

1. `.GAM` 的 player record 大小為 `0xEA9`；依欄位順序，`TaxRate` 是
   `player+0x31` byte。`sub_E08F6 @ 0xE08F6` 以
   `dword_197F98 + playerIndex*0xEA9 + 0x31` 讀取該值，乘殖民地工業後除以 100，證明它是
   百分比稅率 consumer。
2. 全 executable 的 displacement `0x31` 掃描得到 74 個候選。逐一審查基址後，能證明以
   `dword_197F98 + playerIndex*0xEA9` 寫入的只有 `sub_CC198 @ 0xCC198`：
   `0xCC1C4` 與 `0xCC1E8` 把真人稅率視窗的 `word_1A122E` 寫回目前玩家，接著呼叫
   `sub_E2D72` 重算帝國。
3. 沒有 AI 回合函式直接寫入 `player+0x31`；掃描到的 `lea ...+0x31` 也沒有以 player stride／
   player base 建立稅率指標。其餘寫入分屬 `0x139` 等其他 record，不能因同偏移誤認為稅率。
4. `SAVE10.GAM` 的五名有效玩家（含真人與四名 AI）開局稅率均為 0，交叉證實 AI 新局的
   0% 初始化值。原版 AI 不會像 remake 的 `DecideTaxRate` 一樣依國庫門檻每回合切換
   10／30／50%；載入既有存檔時則保持存檔值。

第 1、2、4 項為已證實。第 3 項的「不存在另一條經間接函式指標寫入」受靜態分析能力限制，
證據等級為強推論；但它已足以否定 remake 每回合主動調稅，因該行為沒有任何原版 producer
證據，且新局存檔也顯示所有 AI 保持 0%。

## 推翻的舊實作

正常 AI 回合先前無條件呼叫 `RemakeDecider.TaxRate`，國庫低時拉到 50%、中段 30%、充裕時 10%。
這會改變 AI 的工業建造速度與 BC，且是專案自行設計，不是原版 AI。另有 IDA 腳本曾把
`sub_D6E1D` 標為 `raw_AI_Colony_Tax_Dispatch`；該函式實為殖民地職務／建造 dispatcher，已修正
導覽名稱，原位址保留不變。
