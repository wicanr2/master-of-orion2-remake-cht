# 殖民地總覽外部文案規格

## 資料契約

- 表格列、建造／已建、帝國摘要、行星資訊、生產資訊、饑荒、估算標記與轉場皆由 `ui.json` 提供。
- 氣候、重力、礦產與行星大小以 enum 索引產生穩定 key；未知值使用 JSON fallback。
- 動態殖民地序號、建築名、產出及 BC 只作格式參數。

## 幾何契約

- 九列五欄各有固定安全框。runtime 繁中字墨在最小可讀級仍高 16px，30px 列放不下兩條；
  因此建造項與已建摘要合成單列，優先保留目前建造並以省略號收束。
- Planetary Info、Production Info、Empire Summary 各最多五列×17px，均限制於各自黑色面板。
  Planetary Info 不重複上表已顯示且正被懸停的殖民地序號。
- 所有 `postDraw` 動態文字也必須使用 `textSafeRect`，不得直接呼叫字型。

## 驗證

- 雙語最長環境名、建築清單、六位數產出與長建造名通過 runtime 字墨 containment。
- 九列文字不跨列，下方面板文字不跨面板或底部排序列。
- `colonySummary()` source slice不得出現 `.tr(` 或 `font.Draw`。
