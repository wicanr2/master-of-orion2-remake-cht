# 電腦病毒隨機事件規格

## 範圍

事件 3 對目前玩家、熱座席位與 AI 使用相同研究進度規則。訊息顯示的損失必須等於實際扣除值。

## 適用與公式

若目前研究進度低於 10 RP，事件對該目標不適用，且不得消耗效果亂數。否則從事件專用、
可存檔亂數流取得 `roll ∈ [1,50]`：

```text
loss = min(research_progress, roll + 50)
research_progress_after = research_progress - loss
```

事件不改變目前研究 field、已選 application 或已取得科技。

## 驗收

- 純規則覆蓋 9／10 RP 適用邊界、51／100 RP 亂數邊界及低進度夾制。
- shell 抽樣確認 AI 使用 51..100 RP 損失，並只回寫目標帝國。
- 既有熱座通用目標回寫測試證明非目前席位透過同一玩家消費端保存結果。
- 完整 `go test ./...` 在 Docker／Xvfb 內通過。
