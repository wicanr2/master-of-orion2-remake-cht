# 隨機事件 BC 效果逆向稽核（2026-08-25）

## 證據契約

- 輸入：`Orion2.exe`，SHA-256 `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- 工具：IDA Pro 9.4／IDAPython，DOS/4GW object #1 線性位址空間
- 可重生匯出：`tools/ida/audit_event_effects.py`
- 推論等級：下列公式均為「已證實」；位址保留 raw 定位，不用語意改名取代原名。

## 事件 6：富商捐獻

`sub_2230A @ 0x2230A` 的 case 6 在 `0x22A11..0x22A31` 建立事件紀錄：

```text
record.word+3 = ((relative_stardate / 20) * 100) + 100
```

`sub_206A2 @ 0x206A2` 的 case 6 在 `0x20A1C..0x20A44` 讀取目標玩家與
`record.word+3`，再加到 `player+0x32` 國庫。remake 的 `Turn` 為 1 起算，因此輸入
相對星曆 `Turn-1`。

## 事件 15：海盜劫掠

`sub_2230A` 在 `0x2259F..0x225BB` 要求目標國庫至少 100 BC。case 15 的
`0x22B36..0x22B91` 執行 `Random(0x15)`，原版亂數為 1..21，再加 29 得到
30..50%；`floor(treasury * percent / 100)` 寫入 `record.word+5`。

`sub_206A2` 的 case 15 在 `0x20D99..0x20E00` 從目標 `player+0x32` 扣除
`record.word+5` 並防禦性夾至零。remake 以同一條純公式供目前玩家、熱座非目前席位與
AI 使用。

## 邊界與未延伸聲明

- 富商捐獻在相對星曆 0..19 為 100 BC，20..39 為 200 BC。
- 海盜劫掠以整數除法向下取整；99 BC 不適用，100 BC 的損失為 30..50 BC。
- 本切片沒有宣稱其他事件已忠實；電腦病毒已由後續
  [`random-event-computer-virus-audit-20260825.md`](random-event-computer-virus-audit-20260825.md)
  獨立閉合；古代科技亦已由後續
  [`random-event-ancient-tech-audit-20260825.md`](random-event-ancient-tech-audit-20260825.md)
  閉合；氣候改善亦已由後續
  [`random-event-climate-audit-20260825.md`](random-event-climate-audit-20260825.md)
  閉合。外交、其餘殖民地效果及持續事件仍各自需要完整的建立紀錄與消費端證據鏈。
