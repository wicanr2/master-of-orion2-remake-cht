# 殖民地 REFIT 外部文案與安全框規格

證據來源：[`docs/re/refit-ui-text-audit-20260827.md`](../re/refit-ui-text-audit-20260827.md)。

## 文案與資料邊界

- `internal/shell` 回傳 typed `RefitError`：穩定錯誤碼及必要參數，不回傳玩家語言句子。
- `refit.go` 只保存 `refit.*` 文案鍵、格式參數、typed 艦艇／元件資料與幾何。
- 艦級與元件名稱沿用既有外部 catalog；不得複製到 `ui.json`。
- remake 的「自動最佳模板」必須在玩家文案中明說，避免誤導為原版設計庫已完成。

## 主要鍵值與格式契約

| 鍵 | 參數 |
|---|---:|
| `refit.title`、`refit.subtitle`、`refit.empty` | 0 |
| `refit.candidate.summary` | 6 個 `%s` |
| `refit.preview.source_target` | 2 個 `%s` |
| `refit.preview.detail` | 4 個 `%s`、1 個 `%d` |
| `refit.preview.scrap_warning`、`refit.preview.select_prompt` | 0 |
| `refit.button.queue`、`refit.button.return` | 0 |
| `refit.error.no_upgrade` | 1 個 `%s` 艦名 |
| 其餘 `refit.error.*` | 0 |

英文與繁中的格式參數數量與順序必須一致。

## 錯誤碼

`RefitError.Code` 只允許穩定技術值：`no_colony`、`select_ship`、`colony_missing`、
`fleet_missing`、`fleet_not_parked`、`ship_missing`、`facility_required`、`no_upgrade`、
`queue_full`、`enqueue_failed`。未知錯誤以 `refit.error.unknown` 安全 fallback，不把 `err.Error()`
直接顯示給玩家。

## 邏輯座標與溢出策略

畫布固定 640×480：

| 欄位 | 安全框 | 策略 |
|---|---:|---|
| 標題 | `(24,14,592,32)` | 單行置中省略；18px bitmap 字墨高為 32px |
| 副標題 | `(24,46,592,16)` | 單行置中省略；與標題框相接但不重疊 |
| 列表第 i 列 | `(28,76+24i,584,22)` 內縮 6 | 單行省略 |
| 空清單提示 | `(32,180,576,32)` | 單行置中省略 |
| 預覽來源／目標 | `(38,330,556,20)` | 單行省略 |
| 預覽元件／成本 | `(38,352,556,34)` | 最多兩行，末行省略 |
| 報廢警告 | `(38,388,556,16)` | 單行省略 |
| 未選提示 | `(38,348,556,24)` | 單行省略 |
| 兩顆按鈕 | 各自 126×28 熱區內縮 4 | 中心與熱區共用 |
| 底部錯誤 | `(30,458,580,20)` | 單行省略，不越出 480 |

## 驗收

- `refit.go` 不含 `.tr(`、直接字型繪製或玩家句子；`production_controls.go` 的 REFIT 錯誤分支
  不含中文玩家訊息。
- typed 錯誤到 `ui.json` 的映射完整，未知錯誤有安全 fallback。
- 雙語 catalog、格式參數、最長候選列、預覽與按鈕均通過實際 bitmap font 安全框測試。
- 既有預覽、排入、取消報廢、Star Base 門檻、保存與命令重播測試保持通過。
- 本切片不宣稱單頁 adapter 與原版兩階段 REFIT popup 像素相同。
