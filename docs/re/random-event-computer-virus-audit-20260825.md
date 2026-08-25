# 電腦病毒隨機事件逆向稽核（2026-08-25）

## 證據身分

- 輸入：`Orion2.exe`
- SHA-256：`7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- 工具：IDA Pro 9.4／IDAPython
- 位址基準：DOS/4GW object #1 的 IDA 線性位址
- 非破壞性匯出：`tools/ida/audit_event_effects.py`
- 推論等級：下列門檻、亂數範圍、夾制及回寫均為「已證實」。

## 建立事件的適用門檻

`sub_2230A @ 0x2230A` 的 case 3 在 `0x229CC..0x229F7` 讀取目標玩家
`player+0x1EB`：進度至少 10 時保留候選；低於 10 時清除事件狀態並把目標設為 `0xFF`。

`player+0x1EB` 的語意另由 `Check_For_Research_Breakthrough_ @ 0xE44E0` 的累積寫端、
成本比較與成功清零消費端證實為目前研究累積進度，不靠現行 Go 欄位名反證原版。

## 結算公式與回寫

`sub_206A2 @ 0x206A2` 的 case 3：

1. `0x2091D..0x20927` 呼叫 `sub_1247A0(50)`；既有亂數證據確認回傳 1..50。
2. `0x20941` 將結果加 50，形成 51..100 RP。
3. `0x20949..0x2096A` 與目標 `player+0x1EB` 比較；亂數損失較大時改用全部現有進度。
4. `0x2098B` 把實際損失寫入事件 `record.word+3`，供訊息顯示。
5. `0x20992..0x2099B` 同時保存目前研究 field 到 `record.word+5`。
6. `0x209A2..0x209B2` 從同一目標的 `player+0x1EB` 扣除實際損失。

因此公式為：

```text
applicable = progress >= 10
loss = min(progress, Random(50) + 50)
progress_after = progress - loss
```

## Remake 邊界

- 玩家、熱座非目前席位與 AI 必須走同一純規則與可存檔事件亂數流。
- 研究進度低於 10 時先拒絕，不消耗效果亂數。
- remake 保存實際損失於雙語播報，但沒有複製原版九位元組事件 record；此資料模型差異不影響
  玩家可見公式與回寫。
- 本文不宣稱事件 0、18 或其他科技授予事件已閉合。
