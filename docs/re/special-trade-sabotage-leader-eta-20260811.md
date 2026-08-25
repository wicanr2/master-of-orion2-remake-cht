# 特殊貿易、SABOTAGE 與活動領袖 ETA（2026-08-11）

本頁把本輪追回的玩家可感知行為集中記錄。所有位址都是 IDA Pro 線性位址；
輸入為 `Orion2.exe.i64`，SHA-256
`4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`，工具為
IDA Pro 9.4／Hex-Rays 9.4.0.260610。原始名稱、raw operand 與推論等級仍以
[`oracle-static-ida-20260811.md`](oracle-static-ida-20260811.md) 為準。

## 1. 原版貿易目標與活動 Trader

`sub_101BA4 @ 0x101BA4` 的可重算鏈如下：

```text
base = sub_101B3C(min(player +0x5A2, opponent +0x5A2) / 2)
percent = government 4 ? 150 : government 5 ? 175 : 100
if player +0x8B7 != 0: percent += 50
leaderBonus = 0
for leader in 0..66:
    if leader.player == empire && leader.rawStatus < 3:
        if leader.raw +0x28 has 0x08: bonus = (experienceBucket + 1) * 15
        else if leader.raw +0x28 has 0x04: bonus = (experienceBucket + 1) * 10
        leaderBonus = max(leaderBonus, bonus)
percent += leaderBonus
target = base * percent / 100
```

上面 `+0x28` 的兩個 bit 依原版 2-bit common-skill layout 對回 Trader tier 2／1；
raw 位元仍保留在證據匯出中，名稱屬**強推論**。`sub_94951 @ 0x94951` 讀領袖
`+0x24` 經驗，`sub_93D4B @ 0x93D4B` 的 bucket 是：`<60→0`、`60..149→1`、
`150..299→2`、`300..499→3`、`500..999→4`；`>=1000` 且玩家 `+0x8BD`
（Warlord）非零時為 5。領袖 ID `0x42` 強制不讀 Warlord 旗標。

因此活動 Trader 的完整加成表為：

| 經驗 bucket | Trader tier 1（×10） | Trader tier 2（×15） |
|---:|---:|---:|
| 0 | 10 | 15 |
| 1 | 20 | 30 |
| 2 | 30 | 45 |
| 3 | 40 | 60 |
| 4 | 50 | 75 |
| 5（Warlord 高經驗） | 60 | 90 |

`internal/gamedata/original_diplomacy_oracle.go` 保存這條純公式；GAM 匯入現在保留
raw `Experience`，`TreatyState` 只取活動領袖的最大值。舊 JSON／demo 沒有 raw 經驗
時，以顯示等級的 bucket 起點作 fallback，並在程式註解中明示這不是逐值 runtime。
食物／研究交換是 remake 可保存的特殊貿易垂直切片；原版其他 raw 上游／創造力係數
尚未形成完整 UI 對映，不將兩者混稱。

## 2. SABOTAGE 完整成本表與抽樣鏈

`Steal_App @ 0x10130A` 對每個殖民地的 49 個建築槽逐一掃描：record stride
`0x13`，建築旗標在殖民地 `+0x136+slot`，slot `9` 明確跳過，權重讀自
`off_17EB3D + 8 + slot*0x13`。累積總和後呼叫原版 1-based `Random(total)`，
選中項目再呼叫 `Add_Building @ 0x145EA` 清除旗標。

完整 49 筆 `(ID, prereq, productionCost, maintenance, category)` raw 快照已放在
[`internal/gamedata/original_building_table.go`](../../internal/gamedata/original_building_table.go)，
remake 的 `spySabotageCandidates` 已改用該表的 `productionCost`，不是自有建築
顯示名稱的近似成本。原版任務碼 `1` 與 SABOTAGE 成功門檻 `70` 也已保留。

證據等級：任務碼、slot、表位址、`+8` 權重、slot 9 skip、清除效果與
`sub_101483 @ 0x101483` 的三段式 slot helper 為**已證實**；`sub_1014A4 @ 0x1014A4`
的 raw relationship byte（低 6 位數量／高 2 位 mode）、兩張 score table 的讀取位置、
亂數位置與 60／70／80／90／±80 分支也已由 IDA 追回。兩張表在目前資料庫的 raw bytes
是 `0xFF` 初始化值，上游填值與 `toggle_flag` 完整邊界仍是**未知**，不把它們臆測成
科技／政府欄位。remake 的 `spyMissionScore` 已接目前資料模型的完整 AB/DB 來源，
Agent 也已訓練／扣除，玩家可感知的 SABOTAGE 迴圈已用可重播近似完成，不宣稱 raw score parity。

## 3. 活動領袖 ETA

`Deassign_Officer @ 0x934CF` 遍歷 `0x3B×0x43` 全局領袖記錄：

- `RawStatus=4`：每回合 `+0x37` 遞增；達 `30`（`0x1E`）呼叫
  `Check_Officer_Fields @ 0x933F2`。
- `RawStatus=1`：若 `+0x37>0`，每回合遞減；減到 0 且 `+0x23==1` 時呼叫
  `sub_E2AB1`（殖民地計算入口）。
- `Get_Ship_Leader_ETA @ 0x98F42` 的一般 fallback 是 5；暫時領袖池會讀其 raw
  ETA／狀態欄，不能把 fallback 5 稱為所有在職領袖任期。

`sub_E2AB1 @ 0xE2AB1` 在 callback 內掃描六個 `+0x48..+0x52` raw 槽，符合條件時呼叫
`sub_E1D59` 清／重算設計衍生欄位、`sub_DF8F0` 消費艦隊／帝國資料，最後由
`sub_E2710` 重算帝國彙總欄位；這三個下游不是 Win95 UI 呼叫，而是原版遊戲資料消費端，
但 raw 設計表與 remake 結構沒有安全的一對一欄位。remake 因此接上 status 1 的 ETA
decrement、status 4 的 30 回合清理與 GAM raw 欄位保存；`RawETA:1→0` 且 location=1 時，
`applyLeaderETACallback` 以可撤銷增量重整已指派殖民地的領袖效果，刷新所有殖民地士氣，
保留領袖。這是已完成的玩家可感知 callback 近似，不宣稱逐值 raw callback parity。

## 4. 可回查程式與測試

- `internal/shell/treaty.go`：活動 Trader 最大加成與協議目標。
- `internal/shell/spy.go`：raw 49 槽 SABOTAGE 成本加權。
- `internal/shell/leader_tenure.go`：ETA／limbo 門檻。
- `internal/shell/events_ship_explosion.go`：事件 8 的 reservoir 選艦、單艦移除與軍官死亡。
- `internal/shell/spy.go`：`spyMissionScore`、Agent 訓練與 Spy-vs-Spy Agent 扣除。
- `internal/gamedata/original_diplomacy_oracle_test.go`、
  `internal/gamedata/original_building_table_test.go`、
  `internal/shell/treaty_test.go`：邊界與抽樣回歸。
