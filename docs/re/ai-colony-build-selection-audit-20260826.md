# AI 殖民地建造選擇靜態稽核

日期：2026-08-26

## 證據契約

- 輸入：`Orion2.exe`
- SHA-256：`7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- IDA 資料庫 SHA-256：`4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`
- 工具：IDA Pro 9.4，處理器 `metapc`
- 位址基準：IDA linear，DOS/4GW LE image
- 非破壞性匯出：`tools/ida/audit_ai_colony_build.py`

正式 `.i64` 唯讀掛載後複製到一次性容器的 `/tmp`；探針只匯出原始函式邊界、名稱、
指令、bytes、caller／callee 與資料參照，沒有保存資料庫或覆蓋原始名稱。

## 已證實

### 回合呼叫鏈是逐殖民地，而非全帝國單一造艦池

- `Next_Turn_Calc_ @ 0x136B3` 於 `0x13715` 呼叫 raw `sub_D6F67 @ 0xD6F67`
  （外部索引：`All_Colony_AI_`）。
- `sub_D6F67 @ 0xD6F67..0xD6FDA` 逐 AI 呼叫 raw `sub_D6ED4 @ 0xD6ED4`
  （`Colony_AI_`）。
- `sub_D6ED4 @ 0xD6ED4..0xD6F67` 先收集殖民地、更新科技／艦型，再於 `0xD6F4B`
  呼叫 raw `sub_D0D2F`，於 `0xD6F50` 呼叫 raw `sub_D6E1D`。
- `sub_D6E1D @ 0xD6E1D..0xD6ED4` 分別處理封鎖／未封鎖殖民地，於 `0xD6E75`
  呼叫 raw `sub_D10EE @ 0xD10EE`，接著逐殖民地呼叫 raw `sub_D2754 @ 0xD2754`。

因此原版產品與進度的權威單位是殖民地。remake 原先直接把 `EmpireOutput.TotalNetIndustry`
全數投入 `ShipBuildProgress`，沒有殖民地產品選擇，資料形狀已被直接反證。

### 空產品的建築候選與加權門

raw `sub_D2754 @ 0xD2754..0xD2783` 只在殖民地有效、有人口，且目前產品屬特定空／特殊
範圍時呼叫 raw `sub_D0B08 @ 0xD0B08`。後者：

1. `0xD0B32..0xD0B53` 掃 raw building ID `1..48`（上界 `0x31` 不含）。
2. 每項先呼叫 `Colony_Can_Build_Product_ @ 0xE11BC`；不可建者權重為 0。
3. 可建者呼叫 raw `sub_D0036 @ 0xD0036` 計分；最大分數至少夾到 1。
4. `0xD0BC5..0xD0BD9` 將 `(6-difficulty)×score < maxScore` 的候選歸零。
5. `0xD0C1B..0xD0C27` 以 51 槽權重陣列呼叫 raw `sub_FE92D` 抽選；結果 1..48
   透過 raw `sub_B206F` 轉成產品，49／50 分別寫入 `0xFFFD／0xFFFE`。

這證實了候選範圍、可建 gate、難度濾門與加權抽選控制流。

## 強推論、近似與未知

- **強推論**：`colony+0x115` 是目前產品欄；它被 `Colony_Product_Cost_`、
  `Apply_Production_` 與上述 AI 指派鏈共同讀寫。
- **已標註近似**：remake 已有 typed 建築分類、科技 gate 與殖民地輸出，但尚未一對一表示
  `sub_D0036` 使用的全部 raw 玩家／殖民地欄位；本輪只以 typed 類別與糧食／人口狀態重建分數。
- **已標註近似**：原版共用 PRNG 的精確全局位置未閉合；remake 以回合與殖民地 key
  決定性取樣，保存／讀取不會換產品。
- **未知**：`sub_D10EE` 中完整的帝國級建築、支援艦、戰鬥艦、改裝與購買配額；不可把
  本輪建築子鏈誤寫成完整原版 AI 生產 parity。

