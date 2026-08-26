# 殖民地軌道轟炸畫面 IDA 稽核（2026-08-26）

## 輸入與位址契約

- 輸入：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- IDA 資料庫：`Orion2.exe.i64`，SHA-256
  `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- 工具：IDA Pro 9.4，`metapc`。本文位址皆為 IDA DOS/4GW image 線性位址。
- 非破壞性匯出器：`tools/ida/audit_colony_bombing_screen_ui.py`；保留 raw `sub_*`、
  函式邊界、bytes、呼叫與資料參照。

## 已證實

- `sub_B4D02 @ 0xB4D02..0xB4D47` 是轟炸結果的外層入口：保存輸入指標、
  組出內部紀錄、呼叫 `sub_B494B`，並在結束後更新玩家狀態。
- `sub_B4800 @ 0xB4800..0xB48EB` 是逐幀繪製／推進回呼。它以
  `Print_Centered` 在 `(319,10)` 畫行星名標題，再呼叫 `sub_B46CF` 推進炸彈佇列；
  佇列清空後進入 `sub_B4B51`。
- `sub_B4606 @ 0xB4606..0xB46CF` 以 `0x0F` 為炸彈紀錄 stride，未啟用記錄
  只在 `Random_(5) == 1` 時啟用。`sub_B43A6` 掃描 `0x31`（49）個殖民地目標槽。
- `sub_B435A @ 0xB435A..0xB4379` 載入 `COLONY.LBX#1` 作為炸彈動畫。
- `sub_B8EFB @ 0xB8EFB..0xB8F5F` 的直接呼叫點全部屬於地面戰資產 loader
  `sub_B6D51`；沒有軌道轟炸呼叫點。因此舊註解不得用它證明轟炸背景的精確換色。

## 強推論與 remake 近似

- `COLONY.LBX#8` 是 640×480、6 幀累積的建築格地面，且 49 格與原版佇列的
  49 個目標吻合；但本次靜態呼叫鏈未找到 `#8` 的直接 loader，故只列強推論。
- remake 將這張格面的紫色三階以 55%／75%／100% 映射到守方旗色。
  這是視覺近似，不是 `sub_B8EFB` 的轟炸精確還原。
- 原版為逐幀落彈並在完成後自動進入結果；remake 以可重現彈著點呈現一次性
  戰報，並保留「繼續」按鈕。

## 未知／非阻塞

- `COLONY.LBX#8` 的原版直接 loader、精確 palette provider 與換色表仍未閉合。
- 此留白不影響玩家從正常星系畫面發動轟炸、看到戰果並返回星圖。
