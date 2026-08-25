# 隨機事件 2：彗星來襲規格

依據：`docs/re/random-event-comet-audit-20260825.md`。

## 持續 record

新增 `PersistentComet`，保存穩定 `PlanetIndex`、`StarIndex`、倒數、目前耐久與初始耐久。
該 record 必須可經 JSON、熱座及多人快照往返；不得只保存 UI 訊息。

## 建立與互斥

1. elapsed turn 未滿 200 不得建立。
2. 從事件分派已選中的帝國抽殖民地；若 planet 已有彗星、瘟疫、人口暴增，或所在
   star 已有超新星／時空異象／海盜活動 record，建立失敗。事件 14 的同星互斥已由
   `random-event-pirate-activity.md` 接回。
3. 耐久為 `10×(Random(5)+10+difficulty)`，倒數為
   `Random(5)+10-difficulty`；使用可存檔事件亂數流。

## 每回合

1. 計算目標星系所有已抵達艦艇的 `艦體級+1`，不按 owner 過濾；玩家、熱座席位與 AI
   都納入。航行中的艦艇不計。
2. 從目前耐久扣攔截力。耐久 `<=0` 時成功攔截並移除 record。
3. 尚未攔截時倒數減一；仍大於零則保存進度並產生雙語進度訊息。
4. 倒數歸零時，依原版撞擊公式送入共用戰略殖民地 resolver，同步人口群組、駐軍、
   建築效果、建造進度、殖民地／星系 owner 與所有平行陣列，再移除 record。
5. 目標殖民地先消失時，record 安全結束，不得改傷同索引的另一座殖民地。

## 驗收

- 公式邊界與難度方向有純規則測試。
- 只有停泊於目標星的艦艇貢獻，且敵我 owner 不影響總和。
- 成功攔截與撞擊各有測試；撞擊不是固定全滅。
- 玩家、非目前熱座席位與 AI 的撞擊回寫均有抽樣；存讀後倒數、耐久與目標不變。
- 事件 2 進入 `ImplementedRandomEvents`，正常排程可選到，不靠 direct-entry 才能運作。
