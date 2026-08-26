# 命名／旗色外部文案規格

## 範圍

`cmd/moo2/nameflag.go` 不得內嵌玩家可見的繁中或英文句子。畫面標題、輸入提示、預設帝國名、
按鈕、目前旗色格式與八個旗色名稱，統一由 `assets/i18n/ui.json` 的 `nameflag.*` 語意鍵提供。

## 資料邊界

- `internal/shell.FlagColors` 保留 `Key`、RGB 與陣列索引。`Key` 固定為小寫 ASCII 資料識別字，
  不可直接顯示給玩家。
- 顯示名稱由 `nameflag.color.<key>` 取得；目前旗色使用 `nameflag.label.banner` 的單一 `%s` 插槽。
- 空名稱接受時使用 `nameflag.default_empire_name`，不得在 Go 建立第二份安全 fallback 文案。
- 外部文案缺鍵時沿用 catalog 的鍵值 fallback，測試必須把這種輸出視為失敗。

## 驗收

1. 八個旗色索引順序與 RGB 護欄維持通過。
2. 繁中與英文 catalog 均涵蓋全部 `nameflag.*` 鍵，格式化後不得殘留 `%s`。
3. 原始碼護欄拒絕 `.tr(` 與代表性內嵌玩家句子。
4. 本規格不宣稱現行合併版面、純色色塊或 `RACEOPT.LBX#0` 背景與原版一致；這三項須由後續
   單機函式／LBX 資產證據另行解鎖。
