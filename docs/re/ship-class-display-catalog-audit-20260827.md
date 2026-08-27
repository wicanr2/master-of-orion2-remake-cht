# 艦級顯示名稱目錄稽核（2026-08-27）

## 證據分級

- **已證實（原版資料／手冊）**：六個戰鬥艦體依序為 Frigate、Destroyer、Cruiser、Battleship、Titan、Doom Star；順序與 `gamedata.CombatShipClass`、原版艦艇設計六槽一致。既有 `tech.json` 亦保存這組原版英文鍵及繁中譯文。
- **已證實（手冊 p.119）**：Colony Ship、Scout、Outpost Ship 是支援艦；remake 另以 Freighter 表示運輸用途。
- **已證實（remake source）**：`shell` 以十個中文 `Ship.Class` 值作成本、支援艦分類與存檔鍵；直接改成英文會破壞規則查表。
- **已證實（稽核前 remake source）**：`shipClassLabel` 曾在繁中模式直接回傳規則鍵，英文則由 Go 內嵌的兩張名稱表提供；艦艇設計的標題、六個艦體名與三顆按鈕也曾把英文文字直接寫在 Go，再借 `tech.json` 翻譯。本切片已依下方結論移除這些顯示路徑。
- **未知**：舊存檔或外部 `.GAM` 若帶有十種清單之外的艦級，其原版玩家可見名稱無法安全推回。

## 結論

規則鍵與玩家文字必須分離。Go 只保存「規則鍵／typed 艦體索引 → `ui.json` 語意鍵」；十種艦級、未知 fallback、艦艇設計標題及按鈕全部由 `ui.json` 提供。艦艇設計點擊 action 改用 `hull:<index>`，不得再以英文顯示名稱當控制識別字。
