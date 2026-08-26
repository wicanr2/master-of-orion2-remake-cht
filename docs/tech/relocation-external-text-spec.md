# 集結點／遷移外部文案規格

## 分層

`internal/shell` 只保存 `RelocateRefusal` enum、星系索引與布林確認條件；不得保存中英文玩家
句子。`cmd/moo2` 將 enum、星名、怪獸名與數量映射到 `assets/i18n/ui.json`。

## typed 原因

零值 `RelocateAllowed` 代表通過；其餘值分別為：非法星系、起點／終點黑洞、起點／終點未探索、
起點怪獸、起點無我方殖民地及最終寫入失敗。起點與終點必須保留不同 enum，因原版使用不同
文字 ID。

## 外部鍵群

- `relocation.refusal.*`：八種拒絕／失敗原因。
- `relocation.confirm.monster`：星系、怪獸兩個 `%s`；確認框正文。
- `relocation.prompt.*`：選起點、選終點、星圖捷徑、全部改送。
- `relocation.result.*`：清除、設定、無既有設定、改送／清除數量。
- `relocation.button.*`：星圖設定／目前終點、艦隊列表全部改送／清除全部。
- `relocation.fallback.*`：只供無怪獸的畫廊版面示範，仍須由 JSON 提供。
- `fleet.antares.entry`：艦隊列表的王座廳入口提示；它與集結點入口共用左下資訊區，必須保持短句。

所有格式模板必須在中英文保持相同動詞參數數量與型別。Go 只保存鍵值；不得把已翻譯句子塞回
規則層或 `relocatePickState`。

## 顯示與版面

- 星圖 190×20 動作列沿用 `starPanelButtonTextRect`，單行裁切；目前終點名稱可能很長，必須在
  同一安全框省略。
- 艦隊列表新增的兩個 140×18 adapter 按鈕須各有 `textSafeRect`，文字中心與現有熱區中心一致。
- 王座廳入口使用獨立 288×18 安全框，其下緣不得越過兩個集結點安全框的上緣。
- flash 訊息沿用星圖底部既有安全區；確認框沿用 `confirmbox.go` 的定寬、定行數折行契約。
- 任何 catalog 文案即使過長，也不可改變熱區或寫入存檔狀態。

## 驗證

- enum 分支、兩段設定、怪獸確認、取消、全設／清除及存檔行為測試。
- `internal/shell/relocation.go` 不含玩家句子；`cmd/moo2/relocation.go` 不含 `.tr(`。
- 所有鍵雙語存在，格式化不出現 `%!`；最長星名與雙語按鈕通過實際字型安全框量測。
- 完整測試、中英文 35 張畫廊及 `07_fleet.png`／`29_confirm.png` 抽查；英文確認句不得把單字
  硬切成兩行。
