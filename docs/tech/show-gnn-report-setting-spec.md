# Show GNN Report 重製規格

## 路由規則

1. 事件存在且 `ShowGNNReport=true`：進入 GNN 快報畫面。
2. 事件存在且 `ShowGNNReport=false`：略過 GNN 快報，強制進入一般回合摘要。
3. 第二條優先於 `EndOfTurnSummary=false`；特殊事件不可因兩個選項同時關閉而消失。
4. 星系勘查回報不屬於 GNN，仍可使用事件面板呈現。
5. 事件與勘查同時存在且 GNN 關閉：畫面顯示勘查，摘要保存事件訊息。

## 資料與文字

- 不新增玩家可見 Go 字串。GNN 台標、事件內容與 SETTINGS 說明分別沿用 `ui.json`、`event.json` 與結構化雙語 `EventReport`。
- 不修改事件結算、事件亂數或存檔資料；本切片只修正結算後的呈現路由。

## 驗證

- 純路由測試涵蓋設定開／關、摘要開／關、事件與勘查組合。
- `currentReport` 測試確認 GNN 關閉時不選事件，但仍可選勘查。
- 完整 Ebitengine 測試與中文 headless 畫廊抽樣不得退化。

