# 多人資訊面板外部文案規格

## 範圍

本規格只處理 `netInfoScreen` 的標題、狀態說明、STATUS 欄位與 START NET GAME 按鈕文案。
狀態編號、MULTIGM 資產、動畫、位置、熱區與多人流程不因文案外部化改變。

## 資料契約

文案置於 `assets/i18n/ui.json`。每筆使用不會顯示給玩家的穩定語意鍵：

```json
{
  "key": "netinfo.button.start",
  "english": "START NET GAME",
  "value": "開始連線對局",
  "note": "等待其他玩家加入畫面的主機按鈕"
}
```

- `english` 與 `value` 都是外部玩家文案；Go 只可持有 `netinfo.*` 鍵。
- 繁中模式優先回 `value`；英文模式回 `english`。
- 指定語言缺值時退回另一語言，兩者皆缺才顯示語意鍵，使缺檔可見而不 panic。
- 原版字串表仍以英文原文作 `Translate` key；語意鍵只供 remake 自繪 UI 的 `TextFor` 使用，
  不把兩種來源模型混為同一 catalog 契約。

## 狀態鍵值

七個 raw 狀態各有 `netinfo.caption.*` 與 `netinfo.title.*`；共用欄位使用
`netinfo.label.status`，主機按鈕使用 `netinfo.button.start`。

## 驗收

1. `internal/i18n` 驗證相同語意鍵可依語言取得外部中英文。
2. `cmd/moo2` 驗證全部 16 個鍵的實際回傳值。
3. `netinfo.go` 不得呼叫 `tr`，也不得再出現代表性中英文玩家句子常值。
4. `assets/i18n/*.json` 全檔解碼測試必須通過；三平台打包仍須包含整個 i18n 目錄。
5. 畫面驗收仍使用既有文字安全框與 headless 畫廊；靜態測試不取代 glyph ink 檢查。

