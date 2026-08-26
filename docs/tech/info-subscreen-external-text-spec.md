# INFO 五子畫面外部文案與版面規格

## 目的

INFO 的歷史、科技、種族、回合摘要與參考資料五頁，不得在 Go 原始碼內嵌玩家
可見句子。繁中、英文、格式模板與 remake 補充項目統一放在
`assets/i18n/ui.json` 的 `info.*` 命名空間。

## 顯示契約

- 程式只保存穩定語意鍵、數值、格式參數與純裝飾符號。
- 原版已知標題沿用 `BILLTEXT.LBX` 用語；remake 新增內容的 `note` 必須標明用途，
  不得偽稱原版字串。
- 動態模板由 JSON 提供完整句型，Go 端只用 `fmt.Sprintf` 帶入數值或名稱。
- INFO 所有文字只能經 `textSafeRect`；不得直接新增 `extraText` 或繞過裁切。
- 英文模式不得直接顯示 `shell` 保存的中文 `StanceName`／關係名稱；顯示層須先
  轉成外部語意鍵。
- Reference 底部不得暴露開發機路徑或儲存庫結構。

## 回歸閘門

`infosubscreens_text_test.go` 驗證雙語鍵存在、代表性最長動態值經裁切後位於安全框，
並禁止 `infosubscreens.go` 重新出現 `.tr(` 或代表性內嵌玩家句子；既有 AST 測試
持續禁止 INFO 直接建立 `extraText`。
