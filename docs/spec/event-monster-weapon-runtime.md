# 事件怪物武器執行期規格

## 目標

快速怪物戰鬥須消費 `MonsterBlueprint.Weapons` 的每個非空槽與 `Count`，並套用已證實的 raw
HV／PD 傷害倍率；不得繼續用 `MonsterStats.DamageMin/Max` 代表整艘怪物每回合一發。

## 規則

- 每個 mount 依 `OrigWeaponTable` 取得類別與傷害；炸彈不攻擊艦艇。
- 每門可對艦武器各自擲傷害與一般命中；ID 40 Dragon Breath 與 ID 43 Plasma Breath 依手冊
  跳過命中骰。
- raw `0x0002` 對每發傷害乘 3/2；raw `0x0004` 對每發傷害除 2。最低正傷害為 1。
- raw `0x4000` 已由後續證據閉合為 OVR：快速結算採近距 +50%；格子戰術經 typed mod
  先加成，再套 Dragon Breath 每格 -15。怪物專用特殊效果依各自後續規格消費。
- 怪物群的每個個體都使用一份完整藍圖；快速路徑仍以現有最弱艦選目標，是明示近似。
- Guardian 沒有事件怪物藍圖，維持既有明示近似反擊。

## 驗收

- 純資料測試釘住 HV／PD mask 與逐槽傷害範圍。
- 快速戰鬥測試證明炸彈不對艦、必中武器不受一般命中骰、怪物群會倍增武器數。
- Docker 內執行 focused tests、全套測試及格式／擁有權／容器清理檢查。
