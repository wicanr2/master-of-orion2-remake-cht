# 戰術戰機文字與版面證據稽核（2026-08-27）

## 輸入與定位

- 原版輸入：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- IDA 資料庫：`Orion2.exe.i64`，SHA-256
  `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- 位址基準：IDA linear，DOS/4GW LE image。
- 既有已審查錨點：`sub_3AC20 @ 0x3AC20`、`sub_3AD57 @ 0x3AD57`、
  `sub_3D2DF @ 0x3D2DF`；原始名稱衝突與公式證據詳見
  `docs/re/weapon-mod-flags.md`，本輪不以推測性改名覆蓋。

## 結論分級

- **已證實：**上述 raw 函式支撐戰機炸彈、光束與 runtime 分流；本輪不改玩法公式。
- **強推論：**出擊、返航、場上架數是玩家需要看見的既有狀態。
- **remake 適配：**`launchRect()`、24×16 中隊 token、固定 glyph 與摘要列均不是
  原版逐架動畫或控制列座標。它們只需在 640×480 畫布內提供穩定、可讀且不溢位的操作介面。
- **未知：**原版逐架 sprite 動畫的精確 frame timer 與原版控制列出擊熱區；不以本輪
  介面適配冒稱原版精確值。

## 本輪工具狀態

新增 `tools/ida/audit_tactical_fighter_runtime.py`，固定輸出原始定位、bytes、函式邊界及
直接呼叫。既有 `ida-pro-9.4-idapython:py312-v1` 映像本輪執行時回報 IDAPython 尚未設定
且授權不可用，因此沒有產生新證據匯出；這是工具鏈阻塞，不把它誤列為遊戲缺陷，也不把
既有結論升格。實作只依已審查文件與 remake 版面契約進行。
