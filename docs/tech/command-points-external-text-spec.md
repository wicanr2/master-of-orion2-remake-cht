# 指揮點數視窗外部文案與版面規格

## 文案契約

固定玩家文案全部來自 `assets/i18n/ui.json` 的 `commandpoints.*` 鍵：標題、起始值、軌道基地、總計、已使用、淨餘、超額懲罰與關閉提示。Go 只保留鍵值、typed 數值與 `fmt.Sprintf` 參數；不得使用 `tr(中文, English)` 或內嵌可見標籤。

超額文案的第一個整數是赤字點數，第二個是 `deficit * gamedata.IncomeCommandOverflowCostPerPoint`；畫面不可自行重複 `10` 常數。

## 幾何與溢位政策

座標均為 640×480 邏輯座標，面板為 `(150,130,340,236)`：

- 標題框 `(168,140,304,24)`，以面板中心置中。
- 欄名框自 `x=168`開始、寬 `236`；數值框為 `x=408`、寬 `62`，兩者保留 4px 空隙。
- 超額懲罰框寬 `304`，超寬時顯示省略號。
- 關閉提示框 `(168,338,304,22)`，中心與面板中心一致。

所有列都必須經 `textSafeRect`量測、折斷後繪製，不得直接呼叫字型 `Draw`。反組譯證據分級見 [`docs/re/command-points-screen-ui-text-audit-20260826.md`](../re/command-points-screen-ui-text-audit-20260826.md)。
