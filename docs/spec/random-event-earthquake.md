# 隨機事件 7「地震」規格

## 範圍

本規格把原版地震的目標選擇、強度、殖民地傷亡與回寫接到玩家、熱座及 AI 正常回合。
證據來源為
[`docs/re/random-event-earthquake-audit-20260825.md`](../re/random-event-earthquake-audit-20260825.md)。

## 規則

1. 在目標帝國有效、非前哨且沒有 Capitol（raw `+0x13F==0`）的殖民地中，依原始索引做
   reservoir sampling；remake 以各帝國 `ColonyBuildings` 的 typed raw 9 狀態限制候選。
2. 令 `P = 人口 + 49 槽中目前已建且可對回原版 raw ID 的建築數`。
3. 依序從事件亂數流取得 `r3 ∈ [1,3]`、`r2 ∈ [1,2]`，計算
   `damage = max(1, P × (r3+r2) / 10)`，整數除法向零截斷。
4. 將 damage 送入 `gamedata.ResolveStrategicColonyDamage`；建築成本 1，陸戰隊與戰車
   依該帝國種族、Powered Armor 與 Battleoids 的既有 hit cost。
5. 回寫人口、最後人口點數、建造進度、陸戰隊、戰車及被摧毀建築；同步人口職務與
   `PopulationGroups`，玩家側並重算建築衍生士氣。
6. 人口歸零時移除殖民地與所有平行陣列；只有該帝國不再擁有同星系殖民地時才清除
   星系所有權。
7. GNN 訊息使用同一次結算結果，至少包含目標、人口損失與建築損失數；中英文不得
   重新擲骰。

## 決定性與失敗條件

- 無殖民地時事件失敗，呼叫端照原版候選流程繼續下一次嘗試。
- 所有 reservoir、強度與傷亡候選都只消費 `eventRand`，存讀檔後序列可續接。
- 不從 Go map 的迭代順序產生亂數順序；raw 建築 ID 必須排序後交給共用 resolver。
- 玩家、非目前熱座席位與 AI 對同一狀態／亂數序列應得到相同 damage 與傷亡類型。

## 驗收

- 純公式測試覆蓋最低值、整數截斷與最大擲骰。
- 目標測試證明 reservoir sampling 的消費順序，而非直接 `Intn(len(colonies))`。
- 垂直測試至少涵蓋人口、建築、駐軍／建造進度其中兩類回寫，以及殖民地摧毀的平行陣列。
- AI 與熱座抽樣證明不是只修目前玩家。
- 完整 `go test ./...` 在既有 Docker／Xvfb 工具鏈通過；單元測試綠只證 remake 接線，
  原版 parity 仍以 RE 證據範圍為限。
