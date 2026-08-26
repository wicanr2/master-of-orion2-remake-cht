# Enemy Moves 設定規格

## 資料與規則

1. `GameSettings.EnemyMoves` 是唯一開關，預設關閉並隨 `.GAM`／JSON 設定往返。
2. 規則層提供 typed 的可見敵方移動查詢，至少包含 AI 索引、起點星、目的星與剩餘 ETA。
3. 候選必須同時滿足：
   - `FleetETA > 0`；
   - 起點、目的地索引有效且不同；
   - 設定開啟；
   - 起點與目的地都在玩家目前可見星圖內。全知沿用 `VisibleStars()` 的全可見結果。
4. 查詢不得修改 AI 艦隊、RNG、回合狀態或存檔內容。

## 星圖呈現

1. 航線在星球 sprite 前繪製，避免壓住星球與星名。
2. 使用固定敵方色的細線與一個沿線循環的 marker，動畫只依 `sceneBuilder.animTick` 與穩定 AI 索引；不寫入 session。
3. 設定關閉或查詢為空時不產生任何航線繪製。
4. 不新增玩家文案；未來若需 tooltip／圖例，必須先加入 `assets/i18n/ui.json`，Go 只保存文案鍵。

## 證據邊界

原版設定 byte 沒有已閉合的直接 gameplay xref。remake 只實作 help 明示的「看見敵方移動」契約；線色、marker 形狀與 timing 是視覺近似，不能寫成原版精確動畫。

## 驗收

- 規則測試覆蓋開關、ETA、索引、霧區與全知。
- UI 測試覆蓋有效／越界航線的純幾何、marker 位移與安全邊界；實際畫布由 headless 畫廊驗證。
- `scripts/test-ebiten.sh ./...` 通過，中文 `-gamegallery` 35/35 非空並抽看星圖、SETTINGS 入口及狀態指紋。
