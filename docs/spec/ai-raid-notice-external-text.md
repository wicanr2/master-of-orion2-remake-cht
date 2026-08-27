# AI 殖民地突襲通知外部文字規格

## 資料契約

- `AIRaidReport` 只保存 AI／星系雙語名稱、殖民地索引、`Repelled`、`PopLost`、`BCLost`、
  `Building` 與 `FleetLost`；不得保存玩家句子。
- `advanceAIRaids` 每回合先清空 `LastRaidReport`，一回合最多留下第一筆突襲結果。
- `HasSeriousTurnSummaryReport`、熱座席位與所有 UI 消費端只讀 `LastRaidReport`，不得另設
  是否發生的字串旗標。

## 顯示契約

- `aiRaidNoticeText` 依 `Repelled` 選擇擊退或突破模板；突破的建築摧毀、攻方折損是可選片段。
- AI 與星系英文名缺值時安全回退原欄位；星名全空時使用外部 `common.unknown`。
- 被摧毀建築以 `colonyBuildingLabel` 依語言顯示；英文通知不得直接輸出中文規則名稱。
- 主句、可選片段及片段分隔符全部來自 `ui.json`，Go 不保存標點句型。
- 回合摘要沿用既有動態多行安全框，INFO 沿用事件列安全框；不新增直接字型繪製。

## 驗證契約

- 擊退、突破、突破＋建築、突破＋攻方折損及兩片段同時存在均有雙語測試。
- 英文測試必須證明 AI 名、星名與建築名不洩漏中文。
- 規則來源不得出現兩種通知句子或 `Message` 欄位；重要摘要及熱座往返讀 typed report。
- 聚焦規則／UI 測試、雙語正常玩家畫廊、`git diff --check` 與 Docker 清理通過。

