# 格子戰術控制列外部文案規格

## 範圍

本切片涵蓋 `tacticalbar.go` 的 AUTO／SCAN／BOARD／RETREAT／WAIT／DONE／OPTIONS 行為訊息、
掃描摘要、登艦結果、模式提示，以及 `interactive.go` 的七顆按鈕標籤。原版入口與證據界線見
`docs/re/tactical-control-bar-text-audit-20260826.md`。

## 文案契約

- Go 只保存 action ID（`auto`、`scan` 等）與 `tactical.*` 鍵；中英文句子及格式模板全部放在
  `assets/i18n/ui.json`。
- `fmt.Sprintf` 模板的兩種語言必須維持相同參數數量與型別順序。掃描摘要依序是名稱、結構
  現值／上限、裝甲、攻擊、防禦、傷害上下限、護盾減傷、陸戰隊；登艦戰報不得交換攻守存活數。
- 正版英文控制列不疊動態標籤，讓烘字露出；正版繁中與兩種 fallback 都以相同鍵取文案。
- OPTIONS 提示只能說明「戰鬥內返回 SETTINGS 尚未接線」，不得再宣稱完整設定畫面不存在。

## 幾何契約

- 七顆按鈕沿用 54×18 熱區與 52×16 擦底板，文字中心和熱區中心共用同一組整數座標。
- 動態戰術訊息只進 `tacticalMessageX/Y/W/H` 的單行安全帶，先以實際字型量測，再依既有
  `truncateToWidth` 截斷；模式提示與訊息組合後仍只繪製一次，不得直接繞過安全帶。
- 掃描與登艦的最長雙語模板以實際代入值測試；按鈕標籤須在 52×16 內置中。

## 驗收

- `tacticalbar.go` 不得再含 `.tr(` 或玩家句子；按鈕表不得保存中文／英文顯示標籤。
- `tactical.*` 所有鍵在中英文 catalog 都存在，printf 佔位符相容。
- 七顆按鈕熱區、共享中心、模式切換、WAIT／DONE 與 Ship Initiative 行為測試維持通過。
- Ebitengine 套件與全套測試通過；正常畫廊的戰術控制列仍由 640×480 安全框約束。

