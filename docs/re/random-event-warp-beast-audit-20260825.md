# 隨機事件 26「超空間獸」靜態稽核（2026-08-25）

## 證據身分

- 輸入：`Orion2.exe`（1.31），SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- 工具：IDA Pro 9.4、`ida-pro-9.4-idapython:py312-v1`；分析既有 `.i64` 的 `/tmp` 副本。
- 位址基準：IDA linear、DOS/4GW LE object #1。
- 可重生匯出器：`tools/ida/audit_event_warp_beast.py`；本輪 JSON：
  `/tmp/moo2-event-monsters-work/random-event-warp-beast.json`（不加入 Git）。
- 匯出保留原始函式名、位址、bytes、運算元與 callers；附加語意只供導覽。

## 已證實

1. `Determine_Event_ @ 0x2230A` 事件 26 建立分支只把 record age `+0x02` 清零；帝國槽由
   通用受害者抽選寫入 record `+0x00`。事件 26 的固定 record 範圍為
   `0x19AC8E..0x19AC96`。
2. `sub_206A2 @ 0x2123F..0x212BE` 只在 state 2 消費事件 26。age `>4` 才擲
   `Random(20)==1`，age `>20` 強制結束；若尚未結束，便以 record `+0x00` 的帝國槽呼叫
   `sub_100618 @ 0x100618`，之後再呼叫 `sub_22D57` 抽下一個帝國槽並回寫 `+0x00`。
3. 上述逐回合 consumer 在呼叫 `sub_100618` 前沒有額外的百分比命中骰。因此舊 remake
   的 `Random(100)<=20` 不是原版規則。
4. `sub_100618` 只掃 owner 等於參數且 ship status `+0x64==1` 的船，以
   `Random(candidateCount)==1` 做 reservoir sampling。沒有合格船時不刪艦。
5. helper 會保存被選船的原始索引與艦隊欄位，必要時把同位置、同 owner、同艦隊的另一艘船
   接成替代鏈首；可直接刪除時呼叫 `sub_941C6 @ 0x941C6` 清除 combat-ship／leader 連結，
   `sub_8A4C4 @ 0x8A4C4` 重算衍生資料，再把 ship status 寫成 8。玩家可見結果是被選航行艦
   消失，不是依艦體戰力挑最弱艦。
6. `sub_21371 @ 0x21371` 對事件 26／27 共用新聞狀態分派；事件結束 record 不再執行刪艦。

## 強推論與 remake 停止線

- ship status 1 依完整選艦 caller 與移動資料流判為「航行中」是強推論；原始位址與
  `+0x64` 運算元已保留，不把導覽名稱當證據。
- 原版用 0x81-byte 單艦鏈、0x3B-byte combat-ship record 與 leader slot 表示艦隊鏈首替換；
  remake 使用 `Fleet{Ships, ETA}` 切片，沒有同構 raw record。remake 以「目標帝國全部
  `ETA>0` 艦隊中的每艘船均勻抽一艘並移除」重現玩家可見結果，標為資料模型近似，不逐位元
  仿造鏈首中間狀態。
- 1.50 binary 未取得，不能排除版本差異。

## 推翻的舊斷言

- 「每回合固定 20% 才抓船」：錯；原版 active consumer 沒有這顆骰。
- 「只看目前選取的玩家艦隊」：錯；原版按 record 指定帝國掃全部合格船，並逐回合重抽帝國。
- 「抓走最弱艦」：錯；原版在合格航行船間 reservoir sampling。

