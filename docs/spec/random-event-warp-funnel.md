# 隨機事件 27「曲速漏斗」規格

## 範圍

依 `docs/re/random-event-warp-funnel-audit-20260825.md` 實作 1.31 的報告型持續事件。
本規格涵蓋一般玩家、熱座席位、AI 目標、存檔往返與雙語新聞；不增加 ETA 凍結。

## 建立

- 事件只有在目標帝國至少有一艘有效艦艇時成功。
- 玩家／熱座由所有 `Fleets[].Ships` 均勻選一艘作 record 目標；AI 目前只有聚合艦隊
  模型，以 `FleetStrength > 0` 作有效艦艇閘門。
- 建立 `PersistentWarpFunnel`，初始 `Turns=-1`，使公告回合不算 active turn。
- 新聞說明一支艦隊被曲速漏斗困住，但不改寫 ETA、目的地、船數或船體狀態。

## 推進與解除

- 每次 `advancePersistentEvents` 先把 `Turns` 加一。
- `Turns < 5` 不擲骰。
- `Turns >= 5` 時，每回合 `eventRoll(20)==1` 即解除。
- 若骰失敗，`Turns >= 21` 時仍強制解除（對應原版在遞增前檢查 `age > 20`）。
- 解除只移除 persistent record，播報全艦無損脫困；不需要補償航行回合。

## 目標路徑

- 一般玩家與目前熱座席位走 `applyRandomEventLocalized`。
- 非目前熱座席位由既有 save/load seat 包裝落地，record 留在全局事件列。
- AI 在 `applyRandomEventLocalizedToAI` 建立同種全局 record；AI 聚合艦隊資料不變。

## 驗收

1. 無艦艇／無 AI 艦力時事件失敗。
2. 建立事件不改玩家與 AI 的 ETA、目的地、船數／艦力。
3. 前五個 active turn 不解除；第六個 active turn 起採 1/20；第 22 個 active turn
   最遲解除。
4. persistent record 可經 JSON 存檔往返。
5. 玩家、非目前熱座席位與 AI 各有抽樣測試，並跑 `go test ./... -count=1`。
