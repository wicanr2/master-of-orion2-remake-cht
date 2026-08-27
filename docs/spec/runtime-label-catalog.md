# Runtime 顯示標籤外部化規格

## 鍵值範圍

- `info.history.metric.*`：population、technology、fleet、buildings。
- `newgame.value.difficulty.*`、`newgame.value.size.*`、`newgame.value.age.*`、
  `newgame.value.tech.*`：索引順序必須與 typed 設定表一致。
- `galaxy.system.more_bodies`：一個 `%d`。
- `info.history.legend.you`。
- `tactical.fighter.kind.*`：interceptor、heavy、bomber、assault_shuttle。
- `common.unknown_empire`、`common.unknown_ship`、`common.unknown_player`。
- `combat.ship.named`：兩個 `%s`；`hotseat.player.numbered`：一個 `%s`；
  `hotseat.player.numbered_race`：兩個 `%s`。

## 實作契約

1. Go 只保存 typed enum／索引到穩定鍵的映射，不保存固定中英文顯示值。
2. 無效索引使用當前語言的 `common.unknown`；不得一律回英文。
3. 新遊戲選項索引、星數、帝國數及開局規則不變。
4. 敵艦、熱座與舊存檔 fallback 只改顯示層，不回寫原始名稱。
5. 格式模板的雙語參數數量與順序必須一致，格式化不得出現 `%!`。

## 驗收

- 所有有效索引雙語非空，無效索引回傳各語言 Unknown。
- `englishlabels.go` 不再出現被遷移的固定英文句子或固定顯示陣列。
- 新遊戲、INFO 歷史、行星列表、戰術與熱座既有測試通過。
- 中英文畫廊抽樣確認新遊戲、INFO、行星列表與戰術畫面沒有 key 洩漏或文字越框。
