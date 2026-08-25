# 事件怪物藍圖規格

## 目標

讓 raw owner/type 10..14 的事件怪物由同一份 typed 藍圖驅動建立、存檔、快速交戰與殖民地
防禦，不再以威脅敘述猜測 60–120 的單一結構池。

## 資料與相容性

- `gamedata.MonsterBlueprint` 保存 raw type、艦體、引擎、電腦、五個 special bytes、武器槽、
  picture、基礎戰速、結構及裝甲。
- `MonsterStats` 的事件怪物結構／裝甲由藍圖鏡射；Guardian 沒有納入本表，維持明示近似。
- `MonsterGuard` 分開保存剩餘 `Structure` 與 `Armor`。新生成怪物填滿兩池；舊 JSON 的
  `Armor==0` 視為既有狀態，不在載入時自動補滿，以免修復已受傷怪物。
- Eel 分裂時，每個新個體同時增加一份結構與裝甲。

## 戰鬥規則

- 玩家齊射先扣裝甲，再扣結構；只有結構歸零才移除怪物。
- 殖民地固定防禦使用同一雙血池，並把剩餘裝甲／結構寫回事件 record。
- 武器 mount 資料本輪可供後續快速／戰術消費端使用；未證實的 mods runtime 不以猜測效果
  接線。既有怪物反擊近似在對應消費端閉合前仍明確標記為近似。

## 驗收

- 表格測試逐種釘住 raw type、drive、combat speed、結構、裝甲與所有非空武器槽。
- 測試新生成、Eel 分裂、裝甲先吸收、殖民地防禦寫回與舊零裝甲相容。
- Docker 內執行 focused test、`go test ./...`、`git diff --check`，並確認無專案容器殘留。
