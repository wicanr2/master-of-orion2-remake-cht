# patch 1.5 CFG 檔:潛在資料來源(需版本消歧後才可用)

## 發現

patch 1.5 安裝包 `moo2_patch1.5/MOO2-1.50.26.zip` 解壓後含大量 `.CFG` 檔,是 **1.50「improved」mod** 的
modding 配置(手冊「Modding with Config」章節,ORION2.CFG/PARAMETERS.CFG/USER.CFG 等)。主要檔:

| 檔 | 行數 | 內容 |
|---|---|---|
| `PARAMETERS.CFG` | 3574 | 遊戲參數與公式開關(研究突破、refit 成本、建築排序…),多附 `# default, classic` 註解 |
| `AIRACES.CFG` | 122 | **AI 種族性格分布**(race_personality 每種族 10 檔位)、種族 ground/lowg 修正 |
| `AISHIPS.CFG` | 165 | AI 艦艇設計偏好 |
| `TWEAKS.CFG` / `MELEE.CFG` | 318 / 526 | 戰鬥/規則微調開關 |
| `TECHTREE.CFG` | 138 | 科技樹配置 |
| `BUILD*.CFG` | 小 | **建造佇列**(玩家 build order 偏好,非建造成本) |

## ⚠ 為何尚未移植(版本消歧陷阱)

1. **這是「improved」mod,非 classic 1.5**。專案目標是 classic 1.3/1.5,直接抄 CFG 值會混入非原版行為。
   好消息:CFG 常以註解保留 classic 原值,例如 `AIRACES.CFG`:
   ```
   race_personality Humans = 3 3 3 4 4 4 4 4 5 5; ## 3 4 4 4 4 4 4 4 5 5   ← ## 後為 classic 原值
   ```
   `##` 標記處 mod 值與 classic 不同,要取 classic 必須讀註解、逐行消歧。
2. **核心武器/裝甲/護盾數值表不在 CFG**,仍硬編在 `Orion2.exe`(這些 CFG 只調公式開關與 AI,非基礎數值表)。
3. **AI 性格分布 ≠ AI 決策邏輯**。AIRACES.CFG 給的是「哪個種族傾向哪種性格」(0 排外…6 失格),
   不是「性格如何影響每回合決策」的演算法——後者社群公認未解(見 `community-mechanics-findings.md`)。
4. **建造成本仍無來源**:BUILD*.CFG 是 build queue 不是成本表;成本在 EXE。

## 可謹慎採用的部分(未來工作)

- AIRACES.CFG 的 **classic race_personality 分布**(取 `##` 原值)是官方權威 datum,可作為未來 AI 的種族性格資料表。
- PARAMETERS.CFG 標 `# default/classic` 的公式開關,可交叉驗證已移植公式(非新數值)。

## 結論

CFG 是「有價值但需人工版本消歧」的來源,不適合自動化盲抄。武器/裝甲基礎數值仍須從 `Orion2.exe`
逆向(社群已知偏移)或實機取得。列此供未來謹慎採用,維持「不臆造、不混版」的紀律。
