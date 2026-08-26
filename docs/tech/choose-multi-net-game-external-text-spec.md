# 區網對局選擇畫面外部文案規格

## 文案鍵

- `netgames.title`
- `netgames.empty.title`、`netgames.empty.hint`
- `netgames.button.direct_address`、`netgames.button.cancel`
- `netgames.error.selected_game`、`netgames.error.direct_address`
- `netgames.demo.{orion,sakkra,antares}`

程式不得保存上述中英文句子。`direct_address` 錯誤以 `%s` 插入玩家輸入的位址；兩種語言的
格式佔位符必須一致。對局名稱、位址、玩家數是動態資料，不翻譯；只有固定畫廊名稱外移，避免
正式畫廊仍依賴 Go 內嵌玩家文字。

## 幾何與繪製

1. 標題框為 `(winX+86,winY+26,308,26)`。英文有正版 panel 時保留烘字；繁中或缺 panel 時
   擦底後由 `ui.json` 重繪。
2. 每列沿用原版 362×22 熱區，文字再分成名稱、位址與人數三個不相交安全框；三框不得超出
   row，且 row 文字不得掉入下一列。
3. 空清單標題與提示各用一個 20px 高單行框，不重疊。
4. 直接位址文字框與 `cmngDirectRect` 共用中心；取消文字框與現行量圖按鈕矩形共用中心。
5. 訊息列放在最後一列與取消按鈕之間，單行截斷；不得壓到清單、取消或直接位址入口。
6. 所有固定文字經 `textSafeRect` 繪製，不再直接呼叫 `Draw`／`DrawCentered`。

## 驗證

- 雙語鍵與 `%s` 格式完整，不產生 `%!`。
- 實際 bitmap font 下，最長固定文案不被裁切；動態超長名稱／位址依各自欄位截斷。
- 三個 row 欄互不相交，按鈕文字與點擊熱區中心一致。
- 靜態測試禁止 `choosemultinetgame.go` 使用 `.tr(` 或內嵌本規格固定玩家文案。
- 既有原版座標測試、UDP browser／TCP join 測試保持通過。
- 中英文 35 張畫廊抽查 `33_netgames.png`，確認繁中不露烘字、英文不重繪原版標題。
