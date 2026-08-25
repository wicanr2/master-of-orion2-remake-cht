# 事件怪物格子戰術規格

## 正常入口

- 星系畫面的「攻擊怪物」先驗證怪物、玩家艦隊位置及戰鬥艦，再建立格子戰術雙方。
- 玩家側沿用 `StartCombat` 的玩家艦 adapter；敵方改由 `MonsterBlueprint` 建立，不生成 AI
  帝國代理艦隊。
- `AttackMonster` 保留為快速／命令重播 API，但不再是 GUI 唯一路徑。

## 怪物戰術艦

- 每個怪物個體建立一艘 `CombatShip`；聚合血池按個體順序分配，總裝甲與總結構不得改變。
- 使用 blueprint 的 size、combat speed、computer、picture 與逐槽 weapon count。
- raw HV／PD 轉成既有 typed mod；未知 bit 留在 `RawMods`。
- 炸彈槽保留供狀態顯示，但艦對艦公式會拒絕開火。

## 戰後

- 玩家倖存艦依現有戰術契約回寫。
- 將所有存活怪物戰術艦的 `HP` 與 `ArmorHP` 分別加總回 `MonsterGuard`；全滅才移除怪物。
- 戰敗或撤退時怪物保留實際剩餘雙血池；不得回滿。
- `LastBattle` 必須顯示怪物名稱與雙方損失。

## 驗收

- 測試建構後個體數、總血池、逐槽武器、戰速與 sprite adapter。
- 測試勝利移除、撤退／敗北保留剩餘雙血池與玩家艦損失。
- UI 測試證明星系 action 進入 `tacticalScreen`，不是立即快速結算。
