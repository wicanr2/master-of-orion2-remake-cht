# 多人資訊面板文案契約稽核（2026-08-26）

## 問題與先前歧義

`cmd/moo2/netinfo.go` 已保存多人資訊面板的版面逆向筆記，但玩家可見的中英文標題、狀態與按鈕
仍以 `tr(中文, English)` 或字串常值內嵌於 Go。本輪要先確認這七個畫面確實是原版共用狀態面板，
避免把 remake 自製畫面誤當成原版七張獨立 UI，再把替換文案移至外部 JSON。

## 工具、輸入與位址空間

- 工具：IDA Pro 9.4，映像 `ida-pro-9.4-idapython:py312-v1`。
- 輸入：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- 資料庫：`Orion2.exe.i64`，SHA-256
  `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- 位址均為 IDA DOS/4GW image 線性位址。
- 可重生探針：`tools/ida/audit_netinfo_text_contract.py`。探針保留 raw 名稱、函式邊界、
  指令 bytes、caller 與 data/code refs，不改名也不寫回原始 `.i64`。

## 已證實控制流

1. `sub_F53D7 @ 0xF53D7..0xF54A0` 是共用 reload 主體。以下 wrapper 以 raw 資產編號
   進入它：
   - `sub_F552A @ 0xF552A`：`mov eax, 0Fh`；
   - `sub_F54CF @ 0xF54CF`：`mov eax, 17h`；
   - `sub_F551B @ 0xF551B`：`mov eax, 18h`；
   - `sub_F54BE @ 0xF54BE`：`mov eax, 19h`；
   - `sub_F54D9 @ 0xF54D9`：`mov eax, 1Ah`；
   - `sub_F53CB @ 0xF53CB`：`mov eax, 1Eh`；
   - `sub_F54A0 @ 0xF54A0`：`mov eax, 1Fh`。
2. `sub_F19C7 @ 0xF19C7` 是共用面板 draw；`sub_F2C8B @ 0xF2C8B` 是傳送／接收進度
   draw。`sub_F54D9 @ 0xF5506` 寫 `[window+0x10F]=0`，`sub_F54A0 @ 0xF54B6`
   寫 1，因此兩者不是兩張獨立畫面。
3. `sub_F0801 @ 0xF0801` 在 `0xF0838` 以 bytes `e8 73 49 02 00` 呼叫
   `sub_1151B0 @ 0x1151B0`。它是按鈕欄位建立鏈，不是「已加入人數」文字欄位。
4. `sub_F552A @ 0xF5559` 直接取 `MULTIGM.LBX`。英文標題、狀態欄與按鈕文案烘在該資產；
   executable 不提供 remake 可直接查詢的雙語字串表。

以上為**已證實**。資產中每個英文 glyph 的逐像素範圍仍屬量圖證據，不由本次函式鏈升格。

## Remake 對映與停止線

- `netInfoState` 繼續直接使用七個 raw 資產編號，不另造狀態序號。
- 中文模式以既有擦底疊字覆蓋烘字；英文模式可露出原圖，但自繪 fallback 與按鈕仍必須能從
  外部 catalog 取得英文，因此 JSON 同時保存語意鍵、`english` 與繁中 `value`。
- Go 只保存 `netinfo.*` 語意鍵。句子、標題與欄位標籤不得內嵌。
- Windows／網路 API 內部不影響此玩家可見文案契約，依專案停止線不再深挖。

## 驗證與剩餘限制

- JSON 解碼測試需覆蓋語意鍵的中英文查詢與缺鍵 fallback。
- `netinfo.go` 靜態測試需拒絕新增 `tr(...)` 及代表性玩家文案常值。
- 本輪不宣稱七個狀態都有正常多人流程觸發；那是多人狀態機的獨立驗證項目。

