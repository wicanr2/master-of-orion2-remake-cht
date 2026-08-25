# 銀河議會召開排程稽核（2026-08-24）

## 結論

- **已證實**：`Check_For_Council_Meeting_ @ IDA linear 0x168AF` 每回合只由
  `Next_Turn_Calc_ @ 0x136B3` 直接呼叫一次。
- **已證實**：相對開局日曆未達 25 時不召開；第一次符合其他條件後可於第 25 回合召開。
- **已證實**：每次召開後把下一屆門檻寫成「目前相對回合 + 25」，不是 remake 原有的 8 回合。
- **已證實**：至少要 3 個存續帝國；已殖民星球門檻是正整數 `total / 2` 向下取整。24 顆星時
  仍是 12；奇數星數時與 remake 原先的向上取整不同。
- **已證實**：`sub_15239 @ 0x15239` 是實際議會流程，於 `0x152BA` 把 `word_19A0E2` 寫為 1；
  `Check_For_Council_Meeting_` 以此區分第一次召開與後續日期門檻。

候選人、棄權及外交投票仍是不同函式的待查切片，不由本排程結論外推。

## 輸入與工具

| 項目 | 值 |
|---|---|
| 原版 ZIP 成員 | `mastori2/Orion2.exe` |
| 執行檔大小 | 2,644,842 bytes |
| 執行檔 SHA-256 | `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5` |
| IDA 資料庫 | `/tmp/moo2-Orion2-consumer-work.i64`（未加入 Git） |
| 資料庫 SHA-256 | `6562313be340a6bb80d43f25478446ba0bae24285ac86f0419b4f7de02a14fd0` |
| 工具 | IDA Pro／Hex-Rays 9.4.0.260610，IDAPython |
| 位址空間 | IDA linear |
| 可重生匯出 | `tools/ida/audit_council_schedule.py` |

腳本只讀函式、bytes 與交叉參照，不改名、不套型別、不寫回資料庫；輸出固定保留 raw 位址、附加
語意、推論等級與來源。

## 原始指令證據

### 最小時間

`0x16901..0x1690E`：

```text
mov eax, dword_192FD8
sub eax, 88B8h          ; 35000，轉成相對開局日曆
cmp eax, 19h            ; 25
jl  0x16955             ; 未達 25 直接返回
```

### 第一次與後續門檻

`0x16910..0x16923` 比較相對日曆與 `word_19A0E4`；若尚未到期，只有
`word_19A0E2 == 0` 才能繼續。實際議會函式 `sub_15239` 在 `0x152BA` 寫入
`word_19A0E2 = 1`，所以這個旁路只適用於第一次召開。

### 三個帝國與半數殖民

`0x168C1..0x168D2` 以 `0xEA9` stride 統計 record `+0x24 == 0` 的存續帝國；
`0x16925 cmp dx,2 / jle` 證實必須大於 2。`0x168E6..0x168F7` 以 `0x71` stride 統計
record `+0x14 != -1` 的已殖民星球；`0x1692B..0x1693C` 對正的總星數做有號除二並要求
`settled >= trunc(total/2)`。

### 下一屆為目前回合加 25

`0x1693E` 呼叫 `sub_15239` 後，`0x16943..0x1694F` 執行：

```text
mov ax, word ptr dword_192FD8
sub eax, 889Fh          ; 34975
mov word_19A0E4, ax
```

只寫回低 16 位，因此 `sub_15239` 留在 `EAX` 高半部的值不參與結果。若
`currentOffset = currentDate - 35000`，則寫入值為
`currentDate - 34975 = currentOffset + 25`。這同時閉合第一次之後的固定 25 回合間隔。

## Remake 對映與限制

- `GameSession.Turn` 是相對回合；`EndTurn` 在呼叫 `advanceCouncil` 前已遞增，因此 `Turn >= 25`
  可直接對應原版相對日曆檢查。
- `CouncilMeetings == 0` 對應尚未寫入原版議會旗標；後續以 `lastCouncilTurn + 25` 判定。
- `lastCouncilTurn` 已存在 JSON 存檔與熱座狀態，無需新增格式欄位。
- `DisableEvents` 是測試隔離政策，不是本次原版排程證據；本輪保留，不宣稱 parity。
