# 隨機事件 BC 效果規格

## 範圍

本規格只涵蓋事件 6「富商捐獻」與事件 15「海盜劫掠」。兩者必須對目前玩家、熱座席位
與 AI 使用相同規則，事件訊息中的金額必須等於實際回寫金額。

## 富商捐獻

令 `elapsed = Turn - 1`：

```text
amount = floor(elapsed / 20) * 100 + 100
treasury_after = treasury_before + amount
```

`elapsed < 0` 時安全拒絕事件，不建立猜測值。

## 海盜劫掠

只有 `treasury_before >= 100` 才適用。目標確定後從事件專用可存檔亂數流取得
`roll ∈ [1,21]`：

```text
percent = roll + 29
loss = floor(treasury_before * percent / 100)
treasury_after = max(0, treasury_before - loss)
```

不適用的候選回傳失敗，交由既有事件候選鏈處理；不得改成固定金額或百分比。

## 驗收

- 純規則測試覆蓋富商 1／20／21 回合邊界，以及海盜 99／100 BC、30%／50% 邊界。
- shell 抽樣確認 AI 與非目前熱座席位只修改目標國庫。
- 完整 `go test ./...` 必須在 Docker／Xvfb 內通過。
