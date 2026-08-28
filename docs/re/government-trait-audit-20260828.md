# 政體欄位與玩家可見消費鏈稽核（2026-08-28）

## 證據契約

- 輸入：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- 資料庫：`Orion2.exe.i64`，SHA-256
  `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- 工具：IDA Pro 9.4、`tools/ida/audit_custom_race_trait_consumers.py`；位址均為
  IDA linear、DOS/4GW LE object #1。
- 可重生證據：
  [`evidence/custom-race-trait-consumers-ida-20260828.json`](evidence/custom-race-trait-consumers-ida-20260828.json)。
  外部符號只供導覽；原始 `sub_xxx`、位址、bytes 與 operand 均保留。

## 欄位、編號與升級寫入

`player+0x89F` 是 31-byte signed race trait 陣列的 index 0，也是目前生效政體。原始值
`0..7` 依序為封建、邦聯、獨裁、帝國、民主、聯邦、統一、銀河統一；偶數是基本政體，
同族進階政體是下一個奇數。

`Player_Gets_Tech_App_`（raw `sub_E4204 @ 0xE4204`）在取得四項政府科技時分別寫入
`1/3/5/7`。這四筆寫入有玩家記錄基址的同函式資料流支持，不再只是文字搜尋命中。
`Get_Advanced_Government_Type_ @ 0x93D0A` 與多個 consumer 的 `raw/2` 分族判定交叉支持
同一編號契約。

## 已證實的主要玩家效果

| 消費鏈 | 原版行為 |
| --- | --- |
| 生產成本 | `Cost_Reduction_For_Govt_Type_ @ 0x6E1A0`：封建為 `floor((2C+2)/3)`，邦聯為 `floor((C+2)/3)`；其餘不改。呼叫端涵蓋戰略設計、自動設計、總設計成本、改裝、艦型與殖民地產品成本。 |
| 食物／工業／研究 | `Colony_Job_Production_ @ 0xDE280`：統一／銀河統一對食物與工業 `+50%/+100%`；封建／邦聯研究 `-50%/-25%`；民主／聯邦研究 `+50%/+75%`。 |
| 士氣 | `Colony_Morale_ @ 0xDDB25`：前三種基本政體為 `-20%`，軍營抵銷 `+20%`；統一族跳過一般士氣鏈；失都、建築、科技與領袖仍依既有士氣稽核表計算。 |
| 殖民地 BC | `Colony_BC_Production_ @ 0xE03F1`：民主加入基礎 BC 的 `1/2`，聯邦加入 `3/4`；統一族不套一般士氣 BC 項。 |
| 指揮點 | `Compute_Player_Maintenance_ @ 0xE2000`：帝國在既有指揮點合計上再加 `1/2`；這是 command-point capacity，不是直接把艦艇維護費減半。 |
| 間諜防禦 | `Compute_Spy_Bonuses_ @ 0x100A83` 的八格表為 `0,0,10,15,-10,-10,15,15`，只加入 defense table；攻擊表不吃政體。 |
| 同化 | `Apply_Assimilation_ @ 0xE3456` 的每回合 raw 進度為 `30,60,30,60,60,120,12,16`，門檻 240。 |
| 貿易協議 | `Trade_Agreement_Goal_ @ 0x101BA4` 以共同基值為 100；民主為 150、聯邦為 175，Fantastic Traders 另加 50，商業領袖取有效最高加成。 |
| 研究協議 | `Research_Agreement_Goal_ @ 0x101CC5` 先取雙方研究基值較小者的一半，再依發起方政體縮放：封建 `1/2`、邦聯 `3/4`、獨裁／帝國／統一族 `1`、民主 `3/2`、聯邦 `7/4`。 |

詳細整數尺度分別見
[`colony-government-output-audit-20260825.md`](colony-government-output-audit-20260825.md)、
[`colony-morale-audit-20260828.md`](colony-morale-audit-20260828.md)、
[`colony-bc-production-tax-audit-20260828.md`](colony-bc-production-tax-audit-20260828.md)、
[`sabotage-score-upstream-audit-20260825.md`](sabotage-score-upstream-audit-20260825.md) 與
[`assimilation-race-traits-audit-20260825.md`](assimilation-race-traits-audit-20260825.md)。

## 證據限制與停止線

- 上表為直接 consumer 與 caller 已閉合的玩家可見規則；AI 建築評分、科技估值、人口所有權
  轉換及 occupation policy 的完整分支仍須各自保留在後續窄切片，不能由函式名推定公式。
- `Colony_Can_Build_Product_` 的大量 caller 證明政體參與建造合法性，但產品 raw ID 表尚未逐項
  命名；因此不在本文件冒稱完整建造表已閉合。
- `memset_`、`sprintf_`、整數除法輔助碼、Watcom stack probe、C runtime 與平台 API 不屬
  玩家玩法，依專案停止線不納入 RE 或 remake 分母。遠端尾區塊仍須追控制流與暫存器來源，
  不以 IDA owner 或最近外部符號單獨定案。
