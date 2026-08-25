# 最終分數倍率規格

## 輸入

- 八項原版分數的未縮放總和。
- 客製種族確認時未使用的初始 Picks；內建種族與舊存檔為 0。
- 是否已知 Evolutionary Mutation；在 remake 尚無 mutation 消費 UI 時代表另有 4 點未使用。

## 規則

1. `multiplier = 100 + 10×unusedPicks`。
2. Evolutionary Mutation 已知時，現階段 `unusedPicks += 4`。
3. `final = (rawTotal×multiplier+50)/100`，採 Go 整數除法對應原版正分路徑。
4. 不自行把負總分夾成零。
5. 客製種族確認時必須保存基礎倍率；JSON、熱座與多人快照沿用同一 session 欄位。
6. UI 的總分列顯示縮放後結果；逐項列仍顯示原始八項，避免把倍率重複套到每列。

## 驗收

- 5 點未使用得到 150%。
- raw 101、150% 經 `+50/100` 得 152。
- Evolutionary Mutation 在 150% 基礎上得到 190%。
- 存讀檔與熱座切換不遺失基礎倍率。
