# 研究應用選擇畫面外部文案規格

證據來源：[`docs/re/research-choice-ui-text-audit-20260826.md`](../re/research-choice-ui-text-audit-20260826.md)。

## 範圍

本切片包含 `researchChoiceScreen` 的固定玩家文案：

- 畫面標題；
- 目前 field 與「突破後取得」說明模板；
- 選定後返回回合摘要或星系主畫面的轉場名稱。

科技名稱與研究主題名稱已由 `tech.json` 提供，不複製到 `ui.json`。程式內只允許保留
`research.choice.*` 鍵、typed `ResearchTopic`／`Technology`、格式參數與座標。

## 文案鍵與格式契約

| 鍵 | 格式參數 | 用途 |
|---|---:|---|
| `research.choice.title` | 0 | 畫面標題 |
| `research.choice.topic_summary` | 1 個 `%s` | 已翻譯的 topic 名稱 |
| `research.choice.transition.summary` | 0 | 選定後進回合摘要 |
| `research.choice.transition.galaxy` | 0 | 選定後回星系 |

英文與繁中的 `%s` 數量及型別必須一致。不得把完整 `fmt.Sprintf` 結果或科技名稱寫回 catalog。

## 邏輯座標與安全框

畫布固定 640×480：

| 欄位 | 文字安全框 | 字級 | 溢出策略 |
|---|---:|---:|---|
| 標題 | `(20,52,600,36)` | 18 | 單行省略 |
| field 摘要 | `(20,92,600,24)` | 12 | 單行省略 |
| application row | `rowRect(i)` 內縮 6 px | 16 | 單行省略 |

row 的繪圖、hover 邊框、點擊熱區與文字框必須由同一個 `rowRect(i)` 推導。不得直接呼叫字型
繪製繞過安全框；直向字墨高度不得超出各框。

## 驗收

- `researchchoice.go` 不含 `.tr(`，也不保存上述中英文玩家句子。
- 四個 `ui.json` 鍵在英文與繁中均存在，格式參數契約一致。
- 最長測試 topic／technology 名稱經實際 bitmap font 量測後仍在安全框內；超長值必須穩定省略。
- 點擊 application row 仍只寫目前 application，不提前授予科技。
- 全套 Go 測試通過；本切片不宣稱 remake 的獨立 adapter 與原版整合式 `TECHSEL` 畫面像素相同。
