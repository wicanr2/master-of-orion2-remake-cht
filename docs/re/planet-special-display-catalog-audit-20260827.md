# 行星特殊物產與勘查報告顯示稽核（2026-08-27）

## 已證實證據

- `PlanetSpecial` 的 0..11 編碼、`_planet_special_weighted_chance @ 0x17D832` 權重、
  `Do_System_Discoveries_At_Star_ @ 0xE9927` 分派及 `Make_New_Colony_Or_Outpost_ @ 0xE5EB3`
  殖民效果已有位址與消費端證據；詳見 `internal/gamedata/planet_special.go`。
- 原版一次性結果已閉合：太空殘骸 50 BC、海盜藏寶 100 BC、失散殖民地最多 3 人口、
  受困領袖及遠古文物科技。這些是玩法證據，不等於 remake 中英文句子是原版逐字。
- **已證實（稽核前 remake source）**：十二格中英文名稱硬編於 `internal/gamedata`；星圖、
  殖民地主畫面及行星列表直接取名稱，後兩者另自行拼接 `★`。
- **已證實（稽核前 remake source）**：`internal/shell/discovery.go` 在規則結算時同時組出八種
  中英文完整報告，遠古科技還先以中文頓號連接英文科技鍵。這是 remake 文案，不是原版
  `EVENTMSG.LBX` 字串證據。

## 實作結論

- `gamedata` 只保存 enum→`planet.special.*` 語意鍵；名稱與標記模板放入 `ui.json`。
- `SystemDiscovery` 保存 typed 結果；新結算不再產生 Name／Message 雙語句子。舊欄位保留為
  舊存檔相容 fallback，不作新資料來源。
- 勘查畫面依 typed 結果套用 `event.discovery.*` 外部模板；科技主題逐項經 `tech.json` 翻譯後，
  再使用語言對應的外部 separator。
- `★` 與勘查句型是 remake adapter；原版報告逐字及精確排版仍未知，不升格 parity。
