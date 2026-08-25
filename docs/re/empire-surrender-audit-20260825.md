# 帝國投降與事件 34 靜態稽核（2026-08-25）

## 證據身分

- 輸入：`Orion2.exe`（1.31），SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- 工具：IDA Pro 9.4、`ida-pro-9.4-idapython:py312-v1`；分析既有 `.i64` 的 `/tmp` 副本。
- 位址基準：IDA linear、DOS/4GW LE object #1。
- 可重生匯出器：`tools/ida/audit_empire_surrender.py`；本輪 JSON：
  `/tmp/moo2-surrender-audit-20260825/empire-surrender-v3.json`（不加入 Git）。
- 下列語意均保留原始函式名與位址；1.50 binary 未取得。

## 已證實控制流

1. `sub_2670A @ 0x2670A` 只讓非真人帝國進入投降檢查。它要求一條正式戰爭態勢
   （raw 4／5）、接觸旗標與分數門檻，門檻含難度 `15×difficulty` 及其他戰爭數
   `20×count`。`sub_27A3D @ 0x27A3D` 選接收者；其完整估值仍未逐欄定名。
2. 成功時先以 `sub_E4D06 @ 0xE4D06` 寫投降關係，再以
   `sub_233D2 @ 0x233D2` 建立事件 34。事件 record 保存投降者與接收者。
3. `sub_E4D06` **不是資產轉移函式**。它只寫每個 `0xEA9`-byte player record 的
   `+0xE72`，並把原先指向投降者的鏈改指向新接收者。
4. `sub_E4DC9 @ 0xE4DC9` 是延後 consumer；它掃描所有 `+0xE72 != -1` 的帝國，再呼叫
   `sub_E4B5F @ 0xE4B5F` 執行一次完整接收。`Next_Turn_Calc_` 與議會勝利鏈都有 caller。

## `sub_E4B5F` 已證實資產契約

| 原版資料 | 原始證據 | 結果 |
|---|---|---|
| 領袖 | 掃全域領袖 record；owner 等於投降者時改成接收者，若 `+0x74 != -1` 先呼叫 `sub_933F2` | 接收者取得領袖，但全部解除殖民地／艦艇任命 |
| 科技應用 | 掃 83 項 `player+0x117+tech`，狀態 3 才處理；接收者未持有時呼叫 `sub_E4204` | 已知科技聯集，走正常科技 application callback |
| 殖民地 | 掃 500 筆 colony；owner 命中時呼叫 `sub_E481F @ 0xE481F` | 殖民地逐筆改 owner，人口忠誠 owner 位元、星系旗標與殖民地衍生值一併重算 |
| 間諜派駐 | `sub_1026CF`／`sub_10278D` 讀寫 `player+0xE57+target` 低六位 | 每個帝國原本派往投降者的間諜併入派往接收者的數量，上限 63 |
| 貨運艦 | `add [receiver+0x36], [surrenderer+0x36]` | 全數移交接收者 |
| 國庫 | `add [receiver+0x32], [surrenderer+0x32]` | 全數移交接收者 |
| 投降者艦艇 | `sub_E45FF @ 0xE45FF` 掃艦艇 owner 並逐艘呼叫 `sub_A163A` | **全部移除，不移交** |
| 外交／協議 | `sub_E45FF` 對所有帝國雙向清 `+0x627/+0x62F/+0x637/+0x63F` | 戰爭、正式條約、貿易與研究協議全部清除 |
| 帝國狀態 | `sub_E45FF` 寫 `player+0x24=1`，國庫／貨運艦清零 | 投降者出局；`+0xE72` 最後回復 `-1` |

`sub_8A4C4 @ 0x8A4C4` 只重算星系顯示／艦隊圖資料，不是艦艇移交。這也推翻了以
函式鄰接或名稱猜測「投降會把船送給勝方」的可能錯讀。

## 證據限制

- `sub_27A3D` 的接收者選擇含多組帝國評分、難度分支與 `sub_E4CD2` 防循環檢查；本輪只證實
  回傳合法接收者或 `-1`，沒有把每個未定名 player offset 冒稱為精確國力欄位。
- `sub_2670A` 的 raw `+0x717` 分數來源尚未閉合。remake 可先用已存在的正式戰爭、關係與
  可觀察國力做 trigger approximation，但資產結果不得近似成「所有東西直接合併」。
- 原版 player slot 固定；remake 也應保留 AI slice 索引並把投降者清成 inactive，不能移除
  slice 後讓存檔、事件 record、多人 fingerprint 與外交矩陣索引漂移。

## 勘誤

- `event-status-broadcasts-audit-20260825.md` 舊稱「完成帝國轉移 `sub_E4D06` 後建立播報」錯誤；
  正確順序是 `sub_E4D06` 寫 pending surrender → 事件 34 → `sub_E4DC9/sub_E4B5F` 延後轉移。
- `event-status-broadcasts.md` 舊規定「資產轉移後才播」與原版相反；應改為只有成功建立合法
  pending surrender 才播，下一個 surrender consumer 再完成轉移。
