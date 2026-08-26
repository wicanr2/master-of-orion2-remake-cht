# 自訂種族 UI 字串稽核（2026-08-26）

## 問題

`customrace.go` 原本用一組自訂的「無／差／佳／優」表示所有數值 picks，並把中英文標籤同時
放在 Go。既有 `docs/tech/custom-race-picks.md` 已承認這些級距名不是原版字串，但該限制沒有回寫
顯示端。本輪直接解碼正版資料，確認原版實際語料後再外部化。

## 輸入、工具與格式

- 輸入：`RACESTUF.LBX`，SHA-256
  `a7942f4f13ca081c6d8d4f53266a3ee79198f92c786402bd41c33abbeecbb5b4`。
- 工具：專案受版控的 `cmd/lbxstrings`，以 C-string 模式解碼 asset 0：
  `lbxstrings --asset 0 --cstr --tsv RACESTUF.LBX`。
- 資料唯讀掛載；輸出只寫至終端，不修改正版資料。
- Picks 成本與效果仍以 `docs/tech/custom-race-picks.md` 所列 patch 1.50 `config.json`
  為來源，本次字串解碼不重新推論數值。

## 已證實字串順序

RACESTUF asset 0 依序給出：

- `Population`，接 `-50% Growth / +50% Growth / +100% Growth`；
- `Farming`，接 `-1/2 Food / +1 Food / +2 Food`；
- `Industry`，接 `-1 Production / +1 Production / +2 Production`；
- `Science`，接 `-1 Research / +1 Research / +2 Research`；
- `Money`，接 `-0.5 BC / +0.5 BC / +1.0 BC`；
- `Ship Defense` 接 `-20/+25/+50`，`Ship Attack` 接 `-20/+20/+50`；
- `Ground Combat` 與 `Spying` 各接 `-10/+10/+20`；
- `Governments` 接 `Feudal/Dictatorship/Democracy/Unification`；
- `Special Abilities` 接 22 項能力，尾端原文是 `Trans Dimensional`，不是
  `Trans-Dimensional`。

以上字串及順序是**已證實**。asset 中沒有 `None/Poor/Good/Great`；未選數值 pick 應顯示空白，
不能再把 remake 的泛用級距名冒充原版。

## Remake 對映

- 規則資料以 `pickCat.id`、`RaceTrait` 與數值欄位識別，不再以中文或英文顯示字串分支。
- Go 只保存 `customrace.*` 語意鍵；RACESTUF 英文與繁中譯文都置於
  `assets/i18n/ui.json`。
- 各列仍沿用 remake 既有合成版面；原版完整畫面幾何與逐像素對照沒有因字串解碼升格。

## 驗收與限制

- 測試需驗證 10 類選項、22 項能力及全部語意鍵均有中英文值。
- 代表性的 `Population/-50% Growth`、`Ship Defense/+25`、`Trans Dimensional` 必須與
  asset dump 相符。
- `customrace.go` 不得再出現 `tr(...)` 或玩家文案常值。
- 本輪只修正文案與規則／顯示識別分離；點數效果的玩法 consumer 仍以各自 RE/spec 為準。

