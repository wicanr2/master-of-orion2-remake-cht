# 隨機事件 28「蟲洞」靜態稽核（2026-08-25）

## 證據身分

- 輸入：`Orion2.exe`（1.31），SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- 工具：IDA Pro 9.4、`ida-pro-9.4-idapython:py312-v1`；分析既有 `.i64` 的 `/tmp` 副本。
- 位址基準：IDA linear、DOS/4GW LE object #1。
- 可重生匯出器：`tools/ida/audit_event_wormhole.py`；本輪 JSON：
  `/tmp/moo2-event-monsters-work/random-event-wormhole.json`（不加入 Git）。
- 匯出保留原始函式名、位址、bytes、運算元、callers 與 record 交叉參照。

## 已證實

1. `Determine_Event_ @ 0x22795..0x227B2` 對事件 28 呼叫 `sub_100519 @ 0x100519`，傳入
   一般好事件所選帝國槽及九位元組 stack 輸出 buffer；buffer byte 0 仍為 0 時候選失敗。
2. `sub_100519` 掃描該帝國全部 ship record，只接受 status `+0x64==1` 且 `+0x6D!=1` 的船，
   以 `Random(candidateCount)==1` 做 reservoir sampling。沒有候選時不修改航行狀態。
3. 抽中一艘後，helper 以 owner、status、`+0x65/+0x67/+0x69` 航程位置與 `+0x6C` 群組欄位
   找出同一支艦隊的全部船；每艘都呼叫 `sub_FFDDA @ 0xFFDDA`，不是只移動抽中的單艦。
4. 傳給 `sub_FFDDA` 的目的星為所選船 `word +0x65 - 500`。`sub_FFDDA` 立即把 status 寫 0、
   `+0x65` 寫目的星，並從 star record 回填 `+0x67/+0x69` 座標與抵達／探索旗標。因此事件
   是建立回合立即抵達，不是把剩餘時間改成 1 後等待下一回合。
5. stack buffer byte 0 寫成功旗標、dword `+1` 累計同群移動船數、dword `+5` 保存目的星；
   `Determine_Event_ @ 0x22D45..0x22871` 將數量與目的星寫入事件 28 record `+0x03/+0x05`，
   供 `sub_21371` 新聞顯示。固定 record 範圍為 `0x19ACA0..0x19ACA8`。

## 強推論與資料模型對映

- ship status 1、`dest+500` 與同群欄位的玩家語意由完整寫回鏈強力支持；原始運算元與位址均保留。
- `ship+0x6D==1` 與既有跨維度航行快取相符，標為強推論。remake 只有帝國層級
  `TRAIT_TRANS_DIMENSIONAL`，故以目標帝國具有跨維度特性時事件不適用來對映。
- remake 的一筆 `Fleet` 已是原版同位置／目的地／群組的集合，因此選擇權重應為合格艦艇數，
  抽中後把整筆 fleet 立即設為抵達。原版逐船座標中間值在 remake 無同構欄位，不另造假資料。
- 1.50 binary 未取得，不能排除版本差異。

## 推翻的舊斷言

- 「把目前選取艦隊 ETA 改為 1 即完成」：錯；原版立即抵達並可抽中該帝國任一合格航行艦隊。
- 「ETA<=1 不適用」：錯；原版 eligibility 看 ship status 與跨維度旗標，不讀 remake ETA 尺度。
- 「蟲洞只移動一艘船」：錯；抽船只用來選定群組，同群整支艦隊一起抵達。

