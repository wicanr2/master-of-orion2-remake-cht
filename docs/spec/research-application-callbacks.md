# 科技應用授予副作用規格

## 狀態

- `ChosenTech` 保留研究主題的主要 application。
- `GrantedTechs` 保存額外取得、不能安全表示成單一主題選擇的 application；舊存檔缺欄位時視為空集合。
- 科技擁有權判定先查 `GrantedTechs`，再套用既有主題完成／明確選擇規則。

## Battleoids

任何正常授予入口使帝國取得 `TECH_BATTLEOIDS` 時：

1. 若已知 `TECH_ARMOR_BARRACKS`，不改變狀態。
2. 否則把 `TECH_ARMOR_BARRACKS` 加入 `GrantedTechs`。
3. 不得覆蓋 `TOPIC_ASTRO_ENGINEERING` 的既有 `ChosenTech` 或 `ExplicitChoice`。
4. 重複執行結果相同。

正常授予入口包含研究完成、開局額外科技、遺跡免費科技、間諜偷竊與外交科技餽贈。玩家、AI 與熱座席位採相同規則。

## 原版衍生快取

原版在授予後重建殖民地、帝國、星系遮罩與引擎衍生欄位。remake 維持來源狀態與動態計算，不新增對等 raw cache；現有設計更新入口仍須在研究／偷竊／遺跡後執行。

## 驗證

- Astro Engineering 已選 Fighter Garrison，再取得 Battleoids，兩者與 Armor Barracks 均保持已知。
- callback 重跑不改寫主要選擇。
- 額外科技經 JSON 存讀與多人 session snapshot 往返不遺失。
- 玩家、AI、熱座與偷竊／餽贈授予路徑至少各有規則測試或共用入口測試。
