# 地面戰畫面、文案與輸入稽核（2026-08-26）

## 問題與結論

本輪只核對玩家看得到的地面戰畫面，不重開已閉合的傷亡公式。舊文件把外部符號
`Colony_Combat_Screen_ @ 0xB771D` 當完整畫面函式；IDA 資料流證明該位址只是地面單位記錄
helper。完整生命週期由 `sub_B7491` 建立，`sub_B7289` 負責逐幀繪製、動畫與戰鬥回合。

證據分級如下：

- **已證實**：`sub_B7491 @ 0xB7491..0xB771D` 是地面戰設定與生命週期。它載入資產、清空
  `word_19EB94` 記錄區、註冊 `sub_B7289` 繪製 callback，並在戰鬥完結後清理返回；caller 是
  `sub_E85F7 @ 0xE866E` 及 `sub_E87D2 @ 0xE8C76/0xE927D`。
- **已證實**：`sub_B7289 @ 0xB7289..0xB7491` 每幀呼叫攻／守面板
  `sub_B8BC7`／`sub_B8C8B`，再於 `0xB73AD..0xB73B7` 以 `(319,10)` 畫殖民地標題；
  `0xB7439` 呼叫 `sub_EC4FE` 推進地面戰回合。原版不是戰後等待按鈕的靜態報告。
- **已證實**：`sub_B771D @ 0xB771D..0xB774D` 以 side `×0x3E8`、slot `×0x19` 掃 40 筆
  記錄的 `+5` 狀態 byte 並計數；caller 只有動畫／記錄 helper。它不是畫面入口。外部名稱只保留
  作歷史索引，不再作語意證據。
- **已證實**：`sub_B6D51 @ 0xB6D51..0xB7185` 載 `COLONY.LBX` 及多組
  `COLGCBT.LBX`，並對 15 個動畫資產呼叫 `sub_B8EFB` 換玩家色。現行 remake 使用
  `COLGCBT.LBX#0/#5/#10/#15/#21` 的戰後定格，並分別以 typed 攻／守旗色替換單位 ramp；
  戰後定格本身是有標註的呈現近似。
- **已證實**：`sub_B8EFB @ 0xB8EFB..0xB8F5F` 對 raw color 2 直接返回；其他顏色以
  `byte_1828D3[color×8]` 替換 index `0xC0..0xC7`，並以
  `byte_182913[color×8]` 的前四 byte 替換 `0xE8..0xEB`。兩張 8 色 raw 表已由 IDA
  匯出並逐值寫入 `groundcombat.go`，不再用三個 RGB 近似。
- **已證實**：攻／守面板函式分別是 `sub_B8BC7 @ 0xB8BC7..0xB8C86` 與
  `sub_B8C8B @ 0xB8C8B..0xB8D4A`；兩者呼叫 `sub_B896D @ 0xB896D..0xB8BC7`
  列兵力。該 helper 在 `0xB8A0D`／`0xB8B2F` 使用寬度 `0x105=261`，與
  `COLGCBT.LBX#21` 的 261×149 相符。
- **已證實**：`sub_B88B2 @ 0xB88B2..0xB896D` 的落點使用 `Random(50)`；既有
  `(base + random - 20)` 與 Y 上限證據維持有效。
- **remake 近似**：原版戰鬥完成後自動離開；remake 只保存一次解算結果，無法重播原始即時
  動畫，因此保留戰後結果與「繼續」出口。此文案與輸入不可宣稱原版逐像素／逐時序 parity。
- **未知**：`COLGCBT` 無內嵌 palette，現借 `COLBLDG.LBX#0`；原版實際沿用的殖民地 palette
  仍未由 writer／consumer 閉合。這不阻塞安全 fallback。

## 可回查輸入

- 原始執行檔：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- IDA 資料庫：`Orion2.exe.i64`，SHA-256
  `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- 工具：IDA Pro 9.4，映像 `ida-pro-9.4-idapython:py312-v1`；位址皆為 DOS/4GW 映像的
  IDA 線性位址。
- 非破壞性匯出：[`tools/ida/audit_ground_combat_screen_ui.py`](../../tools/ida/audit_ground_combat_screen_ui.py)。
  腳本輸出原始名稱、函式邊界、caller、bytes 與指令，不修改 `.i64`。

## Remake 對應與停止線

本輪只把固定玩家文案移到 `assets/i18n/ui.json`、補雙軸文字安全框、保留 typed 戰果值與
現有真資產／fallback。原版四兵種、即時動畫、palette 與 global RNG 若沒有新的玩家可見阻塞，
不因文案遷移重新開 RE；玩法解算證據仍以 `ground-combat-algorithm.md` 為準。
