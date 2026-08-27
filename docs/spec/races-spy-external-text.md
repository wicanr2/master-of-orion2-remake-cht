# RACES 種族／間諜外部文案與安全版面規格

## 文案與分層

- `races()` 的固定句、格式模板、任務名、Agent 控制及轉場名稱全部來自 `ui.json`。
- `SpyMission`、AI 關係分數與相容存檔中的中文態勢只作 typed／legacy state；UI 負責轉成目前語言。
- 規則層不得再提供 `SpyMissionLabel` 或 `AIRelationName` 這類玩家顯示句子。
- AI 彼此關係項目及清單分隔符亦由 JSON 組合，Go 只保存對手與分數。

## 幾何

- 四條帝國資訊列沿用 `racesInfoLineRect`，相鄰列至少相隔 1px，且不得越過關係滑桿或間諜按鈕列。
- 三個任務文字框由 `racesSpySlotRect` 內縮 3×2px 推導，與命中框共用中心。
- Agent 狀態、訓練及解散位於右下 BONUSES 面板，文字框由各自可見／命中矩形推導；不得占用
  左欄第 4 個帝國槽，底緣不得超過外交按鈕列起點 `y=418`。
- 單行值確定性省略；正常雙語任務名、Agent 數值與按鈕文案必須完整落框。

## 驗證

- 雙語最長帝國名、態勢、關係摘要、63 名間諜／Agent 及六位數 BC 使用 runtime 字型驗證。
- `races()` source slice不得有 `.tr(`；`internal/shell` 不得再保存三種任務或五級 AI 關係顯示名。
- 本規格不證明原版三槽左右語意或 callback。
