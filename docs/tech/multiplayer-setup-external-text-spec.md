# 多人主設定畫面外部文案規格

## 目標

將 `cmd/moo2/multiplayer.go` 的玩家可見文字移至 `assets/i18n/ui.json`。程式只保留按鈕資產、
座標、動作識別字與穩定文案鍵；不得再接受或保存中英文玩家句子。

## 文案鍵

- `multiplayer.setup.title`
- `multiplayer.setup.button.{network,modem,null_modem,hotseat,start,load,join,comm_info,ten,cancel}`
- `multiplayer.setup.button.hotseat_count`
- `multiplayer.setup.note.{hotseat,network}`
- `multiplayer.setup.message.{legacy_transport,choose_supported,no_saves,load_failed,comm_legacy,ten_closed}`
- `multiplayer.setup.error.{host,join}`
- `multiplayer.setup.transition.{new_game,main_menu}`

`hotseat_count`、兩個 note 與兩個 error 鍵使用 `fmt.Sprintf`；JSON 的中英文格式必須保持相同
佔位符形狀。底層網路／檔案錯誤字串可作 `%s` 動態資料插入，但固定前綴必須在 JSON。

## 資料與繪製契約

1. `mpButton` 只保留 `asset`、`dx／dy`、`textKey` 與 `act`。
2. 英文且原版按鈕資產存在時保留烘字；中文或缺資產 fallback 才由同一 `textKey` 重繪。
3. 按鈕文字安全框等於實際熱區內縮 3px；中心必須由同一個 `btnRect` 推導。
4. 標題安全框為面板 `(x+30,y+16,w-60,24)`；英文有原版面板時保留烘字，中文或缺面板時
   從 JSON 重繪。
5. 說明列位於面板下方第一個 20px 區，錯誤列位於第二個 20px 區；兩者寬度限制在
   640×480 邏輯畫布內。每列單行截斷，不允許進入相鄰列或畫布外。
6. `msg` 保存已格式化結果，不保存於規則層或存檔；所有固定來源仍須由 JSON 鍵產生。

## 測試

- 雙語 catalog 鍵完整，格式模板可成功代入且不產生 `%!`。
- 所有按鈕、標題、note 與錯誤模板在實際 bitmap font 下落在安全框；按鈕文字與熱區同中心。
- 靜態測試禁止 `multiplayer.go` 出現 `.tr(`、`mpButton` 的 `zh／en` 欄位，以及本規格列出的
  玩家句子。
- 既有 TCP host／join、停用 legacy transport 與缺資產 fallback 測試不得退化。
- 中英文正常畫廊各抽查多人設定畫面；中文不得露出烘字，英文原版資產不得被多畫一次。
