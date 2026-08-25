# 新遊戲開局規則 RE 審查（2026-08-24）

## 問題

`WORKLIST.md` 將五級難度、開局建築與先進級十九次隨機科技合併為一個
忠實化條目。本輪重新檢查受版控 RE、Go 消費端與測試，以確認真正缺口。

## 輸入、工具與位址基準

- 受版控原版輸入紀錄：`Orion2.exe`，2,644,842 bytes，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- 原有證據工具：IDA Pro 9.4；位址為 DOS/4GW LE object #1 的 IDA 線性位址。
- 本輪探針：[`tools/ida/audit_newgame_rules.py`](../../tools/ida/audit_newgame_rules.py)，固定輸出
  原始位址、bytes、運算元、caller 與 data xref，不改名、不加型別、不儲存資料庫。
- 本輪實際啟動 `ida-pro-9.4-ver3:py312-x11-v4`時，容器回報授權無效，沒有得到
  IDAPython JSON。因此本文沒有把該次失敗執行升格為證據。

## 已有 RE 的重新審查

### 難度

- **已證實**：新遊戲有五個難度，原版全域 `byte_199CB0` 的索引為 `0..4`。
- **已證實**：已追回的 consumer 採各自的整數公式或分支；例如 AI 性格抽樣為
  `clamp(Random(10)+1-difficulty, 1, 10)`，地面戰以普通為中心套整數偏移，
  安塔蘭則分開處理資源率、觸發門與裝甲。
- **已證實不符**：Go `Difficulties.Mult` 的 `0.3/0.6/1.0/1.5/2.2` 自註為 remake
  調校值，卻同時縮放敵艦代理屬性與外交關係；這不是原版的分系統契約。
- **未知**：原版 AI 經濟、建造、科技、艦隊與外交的全部難度分支尚未追完。

### 開局科技

- **已證實**：`Init_Player_Tech_ @ 0x5E55F` 執行 `1/6/25` 次；前六次讀固定表，
  先進級其餘十九次由可研究集合重新選擇。
- **已證實**：原版隨機選的是科技應用，而非一次解鎖整個互斥主題。
- **remake 已實作**：`applyStartingRandomTech` 已每次重算 `AvailableTopics`，以分離亂數流
  保證玩家／AI 決定性，並把選定應用寫入 `ChosenTech` 與 `ExplicitChoice`。
- **2026-08-25 補充**：`sub_FD335` 的應用級單次抽選與 `sub_FC845` 人類／共用估值鏈已閉合並接線；
  raw objective 與 AI personality／objective／theme 分支仍是明確留白。見
  [`starting-tech-application-audit-20260825.md`](starting-tech-application-audit-20260825.md)。

### 開局建築

- **已證實**：`Init_Homeworld_Colony2_ @ 0x13A3D` 依 `word_17D8AC` 的優先表過濾
  已知科技，並使用 `byte_13A3A = {3,5,9}` 作開局等級上限。
- **已證實**：另有 `min(ceil(2/3 population), cap)` 的人口上限。
- **remake 已實作**：`InitialBuildings` 與 `StartingBuildingCount` 已分別實作優先過濾、
  科技門與人口夾限，並由 `applyStartingBuildings` 寫進玩家與 AI 殖民地建築。

## 本輪結論

1. 開局建築與十九次隨機選擇已有垂直實作，不應再當作「整塊未做」。
2. 通用難度倍率是直接可移除的已知不符；本輪規格為
   [`docs/spec/newgame-opening-rules.md`](../spec/newgame-opening-rules.md)。
3. 整個新遊戲條目仍不能標成原版 parity 完成：完整 AI 難度分支與 AI 科技估值仍屬後續 RE。
