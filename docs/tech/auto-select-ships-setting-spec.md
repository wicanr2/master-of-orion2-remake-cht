# 自動選取艦艇設定串接規格

依據：`docs/re/auto-select-ships-setting-audit-20260827.md`。

## 行為

1. `GameSettings.AutoSelectShips` 開啟時，從星圖進入艦隊畫面須預先選取目前艦隊的全部艦艇。
2. 玩家切換到另一支我方艦隊時，須以新艦隊重新建立選取集合，不得沿用上一支艦隊的切片索引。
3. 設定關閉時，進入或切換艦隊均以空集合開始。
4. 自動選取只在「新進入／切換艦隊」初始化一次。玩家以 ALL 或逐艦點擊取消後，單純重建畫面不得再次強制選回。
5. 拆分艦隊成功後，目前艦隊的索引與艦數可能改變；選取集合須重新初始化，不能保留已失效索引。

## 架構與文案

- 選取集合仍是 `sceneBuilder.shipPick`，設定值由 `EffectiveGameSettings()` 讀取，以保留舊 JSON 的原版預設遷移。
- 本功能不新增提示句。SETTINGS 標籤既有 `gamesettings.option.auto_select_ships` 仍由 `assets/i18n/ui.json` 提供；Go 不新增中文或英文顯示字串。
- 艦艇選取標記由 `ui.json` 的 `fleet.selection.*` 提供，兩態使用同寬 ASCII，避免
  `✔` 在 runtime fallback 字型顯示成缺字方框。
- `Auto Select Colony` 不在本規格內，需另以其三個原版讀取端完成證據鏈。

## 驗收

- 開啟／關閉設定各測進入艦隊畫面。
- 兩支不同艦數的艦隊測切換，確認沒有殘留越界索引。
- 開啟設定後手動全不選，再重建同一畫面，確認保持全不選。
- 拆分後重新初始化目前艦隊。
- 執行完整 `go test ./...`；若畫廊有艦隊頁，抽查勾選標記與文字安全框。
