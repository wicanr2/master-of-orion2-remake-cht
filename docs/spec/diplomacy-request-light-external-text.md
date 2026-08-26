# 外交會談請求燈外部文案與版面規格

## 原版錨點與轉接邊界

- 第一盞燈左緣固定 x=506、y=5；後續燈依已畫數乘共同寬度向左排列。
- remake 維持 22×16 方塊與來意色作可操作轉接；這組尺寸、色彩與來意 glyph 不宣稱原版
  精確視覺。原版逐種族動畫資產來源未閉合。

## 文案

- 四種 glyph 由 `assets/i18n/ui.json` 提供：
  `audience.reason.war.glyph`、`trade`、`alliance`、`unknown`。
- `internal/shell` 只保存 ASCII reason code；Go UI 不得內嵌玩家可見的中英文 glyph。
- 未知或空 reason 必須安全落到 `unknown`，不可顯示 catalog key。

## 幾何與操作

- `audienceLightRect(n)` 同時是面板、點擊熱區與文字安全框的來源。
- glyph 使用 `textSafeRect` 單行置中，水平內縮 2px；目前點陣 glyph 字墨高正好 16px，與
  轉接燈同高，故垂直不再假設不存在的 1px 內縮。超寬時截斷，不得越過 22×16 外框。
- 請求順序沿用 `AudienceRequests()`；點第 n 盞必須進入相同索引的對手並清除該請求。

## 驗證

- IDA 匯出可重生遮罩讀取、第一個 x、y、共同寬度與 frame wrap 指令。
- 測試釘住第一盞左緣 506、相鄰方塊向左緊貼、點擊索引、雙語 catalog 與 glyph 安全框。
- 靜態測試禁止 `audience.go` 出現 `.tr(`、直接字型繪製與固定 glyph 字面值。
