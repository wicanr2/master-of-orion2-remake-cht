# 事件怪物同星多群組規格

## 資料與查詢

- `GameSession.Monsters` 可同時保存相同 `StarIndex` 的多筆 `MonsterGuard`。
- `MonsterGroupsAtStar(star)` 只回傳 `TransitETA == 0` 的停泊群組，保持切片順序；航行中
  record 不得被當成守衛。
- `MonsterAtStar` 保留相容 API，回傳第一個停泊群組；選擇是可重播 adapter，不宣稱原版 RNG。
- `MonsterNameAtStar` 必須顯示該星所有停泊種類，避免 UI 隱藏第二群；同名只顯示一次。

## 戰鬥與刪除

- 一次玩家戰鬥只處理一個群組，符合原版 owner/type side 分離。
- 原版 `Search_For_Battles_` 會洗牌星系、以 reservoir sampling 選尚未處理的 owner/type side，
  並在同一輪持續消費全部可交戰 side。這是全銀河自動排程契約，不套用到 remake 的單次玩家
  「攻擊怪獸」指令；該指令固定選第一個停泊群組，使指令重播與戰後回寫不額外消耗事件 RNG。
- 同種類多個個體應由單一 `MonsterGuard.Count` 與聚合雙血池表示；不得只因同星便把不同
  `Kind`、航行狀態或太空鰻 `EelAges` 的 record 合併。
- 戰勝或群組結構歸零時，只刪除本次選中的第一個停泊群組；不得刪除同星另一種類，也不得
  誤刪仍在航行的同目的星 record。
- 第一群被清除後，下一群立即成為 `MonsterAtStar` 的結果；星系在所有停泊群組清空前仍阻擋
  殖民與前哨站。

## 驗收

- 同星 Amoeba／Crystal 均可列舉並同時顯示。
- 清除第一群後第二群仍存在且繼續阻擋星系。
- 同目的星的航行中 record 排在前面時，清除停泊群不得刪除航行中 record。
- 存檔往返保留群組順序與全部 record。
- 同一份鎖步指令在相同快照上重播，必須選中相同的第一個群組且只改寫該群組。
