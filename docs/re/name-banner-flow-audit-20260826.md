# 帝國命名／旗色流程逆向稽核（2026-08-26）

## 輸入與位址契約

- 輸入：`Orion2.exe`，SHA-256 `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- IDA 資料庫：`Orion2.exe.i64`，SHA-256 `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- 工具：IDA Pro 9.4、`ida-pro-9.4-idapython:py312-v1`。
- 位址空間：IDA 線性位址（DOS/4GW image）。非破壞性匯出工具為
  `tools/ida/audit_name_banner_flow.py`；工具不改名、不寫回 `.i64`，每條指令保留位址與 bytes。

## 已證實

1. 外部符號表的 `Request_Banner_Color_ @ 0xFBEE1` 不是單機旗色選單。IDA 顯示它由
   `sub_5C510` 的 `0x5C732`、`0x5C989`、`0x5CCC9` 呼叫，內部等待網路狀態、呼叫
   `sub_F6816` 接收訊息，並讀寫目前玩家記錄 `+37`。它只可標成「多人旗色同步要求」。
2. 外部符號表的 `Player_Name_ @ 0xE5E09` 也不是命名輸入畫面。原始指令以玩家記錄大小
   `0xEA9` 與另一張記錄大小 `0x3B` 掃描資料，測試 `[ebx+0x26]` 的 `0x20/0x10` 位元並
   計算 `×15`／`×10` 的最大值；沒有文字輸入、繪圖或字串 caller。舊名稱只能供導覽。
3. `RACEOPT.LBX` 共七個資產；asset 0 與 4 是 640×480 自訂種族框架，不是命名／旗色畫面。
   現行 remake 在 `nameflag.go` 使用 asset 0 當背景，是已證實的資產路由近似。
4. 旗色索引次序仍由既有雙重證據固定為
   red、yellow、green、silver、blue、brown、purple、orange；索引與 RGB 是資料，顯示名稱不是。

## 強推論與未知

- **強推論**：既有原版影片／畫面對照顯示命名與旗色是兩個獨立畫面，現行合併頁不是原版流程。
- **未知**：單機命名畫面與旗色選單的真正函式根、所用 LBX、八面旗幟圖的精確 asset id、熱區與
  轉場 caller。本輪沒有因外部符號名稱好看就把上述項目升格為已證實。
- `Add_Banner_Field_ @ 0xEFABA` 與 `Draw_Multi_Player_Banner_ @ 0xF27B7` 都載入
  `MULTIGM.LBX`，屬多人設定 UI；不能直接當單機新遊戲旗幟資產證據。

## Remake 決定

本輪只完成不依賴未知資產路由的安全垂直切片：把命名／旗色頁面的玩家文案與旗色名稱移至
`assets/i18n/ui.json`，玩法層只保留穩定旗色鍵、索引與 RGB。拆成兩畫面及真旗幟圖仍是未完成的
原版忠實工作，不以文字外部化冒充完成。
