# 星圖快捷鍵外部文案與安全框規格

## 文案契約

固定文案一律由 `assets/i18n/ui.json` 的下列鍵提供：

- `hotkey.quicksave.no_slot`
- `hotkey.quicksave.failed`：一個 `%s` runtime 錯誤參數
- `hotkey.quicksave.success`
- `hotkey.measure.select_origin`
- `hotkey.measure.hover_target`
- `hotkey.measure.distance`：一個 `%d` 秒差距參數

Go 不得保存上述中英文句子；檔案路徑與作業系統錯誤維持 runtime 資料，不寫進 catalog。

## 幾何契約

- 星圖可用 viewport 為 `starVX0..starVX1`、`starVY0..starVY1`。
- 測距提示框為 210×22，預設置於游標右上；右側不足時翻到左邊，上方不足時翻到下方，
  最後夾限於 viewport。文字置中、水平內縮 5px、垂直內縮 2px。
- 距離框為 96×20，以兩顆星的線段中點為中心，向上偏移 8px後夾限於 viewport。
- 快速存檔框固定在星圖右下可用區 `x=240..523`、`y=396..418`，不得侵入左側
  `x<=238` 的選星資訊面板。
- 三個框都先填固定底板，再以相同 `textSafeRect` 截斷／置中；不得依未截斷字寬擴張底板。

## 行為與驗收

- F9 仍使用 `GameSession.ParsecsBetweenStars`，不另建顯示公式。
- F10 仍直接覆寫 `savePath`，不新增確認框；成功、無槽位及帶 runtime 錯誤三路皆有回饋。
- 雙語 catalog 格式參數分別為 0、1、0、0、0、1；格式化後不得出現 `%!`。
- 四角游標、極端線段中點與長錯誤字串均須留在各自安全框及 viewport。
- `hotkeys.go` 的玩家文案來源切片不得再含 `.tr(` 或被遷移的固定句子。
