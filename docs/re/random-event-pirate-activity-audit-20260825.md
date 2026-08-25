# 隨機事件 14「海盜活動」靜態稽核（2026-08-25）

## 證據身分

- 輸入：`Orion2.exe`
- SHA-256：`7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- 工具：IDA Pro 9.4、`tools/ida/audit_event_pirate.py`
- 位址基準：IDA linear；DOS/4GW LE object #1
- 方法：唯讀 `.i64` 交叉參照與指令匯出；未改名、未改資料庫。

## 已證實

1. `sub_2230A` 的事件 14 分支位於 `0x22AEA..0x22B35`。它以目標帝國呼叫
   `sub_23BEC @ 0x23BEC` 選星，把結果寫入事件記錄 `+0x03`；若
   `sub_242FC @ 0x242FC` 回報該星已有互斥事件，建立失敗。
2. `sub_23BEC` 掃描目標帝國具有殖民地的星系，採 reservoir sampling：第 N 個
   合格候選以 `1/N` 機率取代舊候選。因此不是固定挑第一座殖民地。
3. `sub_2448F @ 0x2448F` 依目標星系的帝國 presence bitset，累加各帝國
   `player+0x3C` 的運輸船數 T，再套難度分段：Tutor=`T/5`、Easy=`2T/5`、
   Average=`3T/5`、Hard=`T`、Impossible=`4T/5`。結果小於 5 時強制歸零，
   呼叫端因此取消事件。Hard 比 Impossible 高是原版指令的實際結果，不修正成猜想值。
4. 建立成功時，`event+0x05`（目前強度）與 `event+0x07`（初始強度）都寫同一值。
5. `sub_206A2` 的事件 14 消費端位於 `0x20CD8..0x20D98`。事件未在結束展示狀態時，
   先計算 `floor(current*100/initial)`；`Random(100)` 小於等於該百分比時，逐一檢查
   目標星系 presence bitset，讓每個仍有運輸船的帝國各損失一艘。
6. 同一回合再以 `sub_23B28 @ 0x23B28` 掃描全域艦艇；位置等於目標星、raw status=0
   的每艘船貢獻 `raw size+1`，不檢查 owner。總值自目前海盜強度扣除；降至 0 以下後
   進入共用結束狀態。
7. `sub_242FC` 對同星系排除事件 2、16、17、25、14、24；前三者先由殖民地轉回星系，
   後三者直接比較星索引。這補足先前彗星規格明列但未實作的事件 14 互斥鏈。

## 強推論與 remake 投影

- `star+0x38` 是星系的帝國 presence bitset：同一位元同時被初始強度與每回合損失鏈消費，
  並用來索引帝國 `player+0x3C`。remake 沒有保留這個 raw bitset，以「該帝國在星系內
  至少有一座殖民地」重建 presence。這是玩家可見語意相同的 typed projection，非 raw
  儲存逐位元對齊。
- remake 的 AI 目前沒有獨立運輸艦建造政策，因此 AI 的 `ActiveFreighters` 可能為 0；
  欄位與事件消費端仍一視同仁接線，不能以目前常見零值刪掉 AI 分支。
- raw ship status=0 投影為 `ETA==0`；AI 單一艦隊及熱座快照亦納入同星系清剿火力。

## 未知與停止線

- `star+0x38` 的所有非事件寫入端未在本切片逐一命名；現有讀取端足以閉合事件玩家路徑。
- GNN 畫面狀態 5／6 的逐幀展示與 EVENTMSG 佔位符不影響規則結算，本切片不深挖。

