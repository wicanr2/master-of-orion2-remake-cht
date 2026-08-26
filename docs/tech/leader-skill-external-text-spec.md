# 領袖技能外部文案規格

依據：`docs/re/leader-skill-display-text-audit-20260827.md`。

## 資料契約

1. `internal/gamedata/leader_skills.json` 每筆必須包含唯一數字 `id`、穩定 `key`、`zh` 與 `en`；缺欄、重複 ID 或無效 JSON 應在程式啟動時失敗即關閉（fail-closed），不得靜默產生錯誤技能名稱。
2. Go 規則層只以 `LeaderSkills` ID 判斷效果。`ZH`／`EN` 僅供顯示及舊中文存檔標籤相容，不得新增以翻譯文字決定玩法的分支。
3. JSON 以 `go:embed` 嵌入 `gamedata` 套件，確保三平台單一執行檔不依賴工作目錄；文案來源仍是可獨立編輯、檢查與版本控制的 JSON，而非 Go 字面值。
4. 無技能領袖的類型通稱使用 `ui.json` 的 `leader.type.*` 鍵；通稱不得與任何真實技能譯名相同。

## 驗證

- 27 個列舉技能都必須有完整、唯一的外部資料。
- captain／administrator 的所有候選技能 ID 都能查到名稱。
- 中英文招募標籤保持既有輸出；無技能 fallback 不與 Commando 衝突。
- 完整 Go 測試及至少一條領袖招募／管理畫面路徑不得出現 JSON 鍵值裸露。
