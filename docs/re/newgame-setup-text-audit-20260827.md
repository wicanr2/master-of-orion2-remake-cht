# 新遊戲設定動態文案稽核（2026-08-27）

## 已證實證據

- 原版 `Newgame_Screen_ @ 0xCD435` 的五個選擇器、三個開關、兩顆按鈕與立即數座標已記錄於 `interactive.go`；值圖格、數值列與熱區由 `sub_CCE2E`、`sub_CCC3D` 及 `NEWGAME.LBX#1..22` 交叉確認。
- `NEWGAME.LBX` 1.31／1.50 的滿版背景索引差異已由實檔資產數與雙版本畫廊驗證；本輪不改資產選擇。
- 固定 DIFFICULTY、GALAXY SIZE、PLAYERS、開關與 ACCEPT／CANCEL 標籤已由 `menu.json` overlay catalog 提供。
- 本輪修改前仍有兩個 `tr` 格式模板（星系大小＋星數、帝國數）及兩個中文轉場名稱留在 Go。

完整中英文格式句是 remake 顯示 adapter，列為**強推論**；設定 enum、星數與帝國數為 typed 動態資料。`ngStripTextRect` 與 12→11→10px 實際字高降級是既有防越框契約，本輪保留並補測試。
