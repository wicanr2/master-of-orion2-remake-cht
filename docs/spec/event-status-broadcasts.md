# GNN 狀態播報 29–35 規格

**狀態：CONFORMED（可表示 record／caller 契約；事件 33 觸發仍為明示近似）**

RE-TRACE: dos-orion2-1.31:0x233AB

CORRECTION-20260830-UNUSED-SETTERS-33-35

證據來源：[`../re/event-status-broadcasts-audit-20260825.md`](../re/event-status-broadcasts-audit-20260825.md)。

## 共用狀態

- 狀態播報不進 0..28 隨機候選池，也不改 `EventLastTurn`／`EventAttemptCounter`。
- `LastEventReport` 使用原版 ID 29..35、正確 `TargetKind/TargetIndex/TargetName`，熱座時廣播給所有席位。
  事件 29 與 34 不在規則層保存成品訊息；外部顯示契約分別見
  [`empire-elimination-notice-external-text.md`](empire-elimination-notice-external-text.md) 與
  [`empire-surrender-notice-external-text.md`](empire-surrender-notice-external-text.md)。
- 為避免每回合重播，保存已消費的帝國滅亡、Orion 發現、安塔蘭勝利、投降／叛亂事件 key，
  以及事件 30 的 stage。這些欄位必須進 JSON 與多人快照。
- 同一回合有多則狀態新聞時使用可存檔佇列；不得讓最後一則覆蓋前一則而永久遺失。

## 觸發

1. **29 帝國滅亡**：存續帝國由有殖民地轉為無殖民地時建立一次；玩家、熱座、AI 一致；report
   只保存 typed target，固定雙語通知由外部 catalog 提供。
2. **30 帝國壯大**：依 RE 文件的殖民星系最大值與三階段門檻；平手按原版掃描順序保留首個最大值。
3. **31 排行榜播報**：只在一般事件確實進入 `Determine_Event_` 候選處理的時點檢查；elapsed>50、
   議會未成立，`Random(40)==1` 後以 `Random(4)-1` 選艦隊／科技／人口／建築類別。
4. **32 發現 Orion**：一般玩家／熱座／AI 首次讓艦隊抵達 Orion 星時建立一次；原版 raw star
   type 無直接 remake 欄位時，以 `MonsterGuardian` 守衛星作強推論對映。
5. **33 擊敗安塔蘭**：remake 的安塔蘭母星勝利正式成立時建立；標記 trigger approximation。
6. **34 帝國投降**：成功建立合法 pending surrender 時立即播報並保存投降者與接收者；下一個
   surrender consumer 才依 [`empire-surrender.md`](empire-surrender.md) 完成資產轉移。只有外交
   態勢變化或文字宣告、卻沒有可消費 pending record 時不得播報。
7. **35 殖民地叛亂**：殖民地真正易手時建立並保存新主人與殖民地索引；鎮壓成功不播。

## 驗收

1. 每種播報各有正向與不適用測試，ID、目標、雙語訊息與去重 key 正確。
2. 事件 30 測 stage 0..2 一般門檻、第三階段議會前窄分支、平手與最大三次。
3. 事件 31 測 elapsed、議會、1/40 與四類別邊界，不改一般事件排程狀態。
4. 熱座非目前席位、AI、JSON／多人快照與同回合多則佇列測試。
5. 全專案測試、格式、擁有權與 Docker 清理通過。
