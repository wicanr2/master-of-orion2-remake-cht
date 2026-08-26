# 研究應用選擇畫面文案稽核（2026-08-26）

## 證據身分

- 輸入：`Orion2.exe`
- SHA-256：`7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- 工具：IDA Pro 9.4、IDAPython
- 位址基準：IDA linear，DOS/4GW LE object #1
- 唯讀匯出：`tools/ida/audit_research_breakthrough.py`
- 本輪重跑輸出：`/tmp/research-breakthrough-20260826.json`（一次性證據，不提交大型匯出）

本筆記只閉合研究應用選擇畫面的玩家文案與版面邊界；玩法時序的完整證據仍以
[`research-application-selection-audit-20260825.md`](research-application-selection-audit-20260825.md)
為準。

## 已證實

- `sub_10DC12 @ 0x10DC12` 是玩家研究選擇的完整函式；本輪 IDA 重跑仍得到相同函式起點。
- `sub_10DB69 @ 0x10DB69` 與 `sub_10DBCE @ 0x10DBCE` 是進入同一選擇函式的兩個入口。
- `0x10E389..0x10E3AA` 依序回寫目前 field 與 application。application 是研究開始前的
  選擇，不是突破後才產生的獎勵選單。
- `sub_E4410 @ 0xE4410` 在突破時讀取並授予既選 application；本輪重跑亦保持相同函式起點。
- 原版研究選擇使用 `TECHSEL.LBX#0`，調色盤鏈由 `SCIENCE.LBX#0` 提供；按鈕文字烘在
  `TECHSEL` 圖片中，動態項目名稱另由科技字串來源繪製。資產鏈的既有交叉證據見
  [`screen-spec-info-research.md`](../tech/screen-spec-info-research.md) 與
  [`palette-chain.md`](../tech/palette-chain.md)。

## 強推論

- 原版把 field 與 application 選擇整合在同一張 `TECHSEL` 畫面。remake 為了保持既有
  `SetResearchTopic → PendingResearchChoice → ChooseResearchTech` 狀態機，使用獨立的
  application 面板；這是玩家可見但不改規則結果的介面轉接近似。
- remake 面板中的「選擇研究應用」、field 說明、轉場名稱是 adapter 文案，不可標成從原版
  字串表逐字取回。科技與研究主題名稱仍由 `tech.json` 提供，不應複製進 `ui.json`。

## 未知與停止線

- 尚未追回原版每個 application row 的精確文字安全框、字級與點擊 widget ID。
- `researchchoice.go` 現用 `RACEOPT.LBX#0` 作裝飾背景，並非原版研究資產；本輪只外部化文案並
  釘住安全框，不以另一張錯位的 `TECHSEL` 全畫面直接替換。若要把獨立 adapter 合併回原版
  八領域畫面，應另開帶正常玩家路徑截圖的視覺切片。
- 上述未知不影響「選擇後寫 application、突破後授予」的玩法鏈，也不阻塞文案外部化。

## Remake 映射

- `researchchoice.go` 只保存 `research.choice.*` 語意鍵、格式參數與幾何。
- 固定介面文案與轉場名稱放入 `assets/i18n/ui.json`；科技與 topic 名稱繼續讀 `tech.json`。
- 標題、field 摘要與每列科技名都使用明確的雙軸文字安全框；溢出策略為單行省略。
- 原版整合式畫面與 remake 獨立 adapter 的差異保留為已揭露近似，不用測試綠升格為像素對齊。
