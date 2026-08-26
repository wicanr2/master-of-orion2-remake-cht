# 安塔蘭王座廳外部文案與版面規格

## 分層

`internal/shell` 以 `AntaranAssaultBlockReason` enum 表示可發動、對局結束、停用安塔蘭事件、
缺次元傳送門及無艦隊。`cmd/moo2` 才把 enum、艦數與戰力映射到 `assets/i18n/ui.json`；規則層
不得組裝中英文句子。

## 外部鍵群

- `antaran.room.title`、`antaran.room.subtitle`：adapter 標題與情境說明。
- `antaran.room.defense`、`antaran.room.player_power`：typed 整數格式模板。
- `antaran.room.odds.low`、`antaran.room.odds.ready`：依雙方戰力切換的判讀。
- `antaran.room.button.*`：發動與撤退。
- `antaran.room.block.prefix`、`antaran.room.block.*`：typed 阻擋原因。
- `antaran.room.transition.*`：返回艦隊與進入戰報的轉場名稱。

所有格式參數須在雙語保持相同數量與型別。Go 只保存鍵值、enum 與數值。

## 文字安全框

- 標題帶 `0..640 × 24..98` 分成 36px 標題與 30px 副標題；22px runtime bitmap font 的
  實際字墨高為 32px，不依名義字級或單一 `maxWidth` 猜測。
- 戰力帶 `0..640 × 300..376` 分成三個互不重疊的固定高度列。
- 兩顆 190×44 按鈕的文字安全框由按鈕矩形內縮 5px 推導，中心必須與熱區中心相同。
- 阻擋列限定於 `0..640 × 448..474`；前綴與原因組裝後仍只能占一行並安全省略。
- 所有欄位同時限制寬度與高度；原版背景、按鈕熱區與玩家狀態不可由文案長度改變。

## 驗證

- enum 與 `CanAssaultAntares`／`AssaultAntares` 使用同一前置條件契約。
- catalog 雙語鍵、格式模板、最長阻擋原因及按鈕文字通過實際 bitmap glyph 量測。
- `antaranroom.go` 與 `internal/shell/antaran_victory.go` 不含本畫面的玩家句子；前者不呼叫
  `.tr(`。
- 中英文畫廊各 35 張，目視抽查 `08_antaranroom.png` 的標題、三列戰力、兩顆按鈕及阻擋列。
