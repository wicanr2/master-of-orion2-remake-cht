# NPC 宣戰與停戰規格

**狀態：CONFORMED（一般 AI↔AI reason 23 垂直切片）**

證據：[`../re/npc-war-ceasefire-audit-20260827.md`](../re/npc-war-ceasefire-audit-20260827.md)。

## 狀態

每個 AI ordered pair 除正式 policy／war state 外，保存：

- `AIWarDurationRaw`：原版 `+0x717`，0..250；
- `AIDiplomacyCooldownRaw`：原版 `+0x72F`，停戰時 30；
- 既有 relation、treaty/agreement bias、trade、research 與 tribute mode。

新增矩陣必須隨 JSON 存檔及熱座 AI 索引壓縮；舊存檔缺欄以 0 補入。

## 回合規則

1. 冷卻大於 0 時每方向減一；雙向歸零且 policy 3 時回到 policy 0。
2. 正式戰爭 policy 至少 4 時，雙向 duration 各加一並夾在 250。
3. NPC 條約談判完成後，依來源低到高掃一般 reason 23 候選；來源已有戰爭不再建立一般戰爭。
4. 候選必須通過輪轉目標、冷卻、難度第三方戰爭 veto、政府／條約／協議／納貢門檻與國力比例。
5. 有多個候選時均勻選一個。AI↔AI 宣戰固定 policy 4，清協議與納貢，relation 寫為 -75..-99，
   談判記憶寫 -200，duration 與 cooldown 歸零。
6. duration 大於 `90-15*difficulty` 時直接停戰：policy 3、war false、relation +50 且最高 0、
   cooldown 30，並清協議與納貢。

## 驗收

- 純規則測試涵蓋國力 cap／折半、政府與政策門檻、冷卻及高難度 veto。
- shell 測試涵蓋宣戰 writer、停戰計時、30 回合解除、存檔往返與熱座壓縮。
- 不得再以顯示關係常數建立或解除戰爭，也不得把 AI↔AI policy 4 正規化為人類戰爭 policy 5。
