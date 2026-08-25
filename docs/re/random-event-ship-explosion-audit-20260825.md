# 隨機事件 8「艦船爆炸」靜態稽核（2026-08-25）

## 證據身分

- 輸入：`Orion2.exe`
- SHA-256：`7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- 工具：IDA Pro 9.4、`tools/ida/audit_event_ship_explosion.py`
- 位址基準：IDA linear；DOS/4GW LE object #1
- 方法：唯讀 `.i64` 函式邊界、交叉參照與原始指令匯出；未改名、未改資料庫。

## 已證實：建立與選艦

1. `Determine_Event_ @ 0x2230A` 在事件 8 eligibility 分支 `0x22619..0x22631`
   以目標帝國槽呼叫 `sub_23CED @ 0x23CED`；沒有合格艦時把候選改成 `-1`。
2. `sub_23CED` 掃描全局艦艇記錄 `0..word_199994-1`，每筆 stride `0x81`：
   - `ship+0x63` 必須等於目標帝國槽；
   - `ship+0x64` 只接受狀態 0 或 1；
   - 每遇一艘合格艦，候選數加一並擲 `Random(candidateCount)==1`，成立才替換目前候選。
     這是單趟 reservoir sampling，每艘合格艦等機率。
3. `Determine_Event_` 的事件 8 switch 分支把所選全局 ship index 寫入事件 record
   `+0x03`。沒有「帝國必須至少有兩艘」或「最後一艘免疫」判斷。

## 已證實：消費與軍官死亡

事件 consumer `sub_206A2` 的 case 8 位於 `0x20AD7..0x20B61`：

1. 以事件 `+0x03` ship index 讀 `ship+0x74` 的軍官 ID。若不是 `-1`，先把 ID
   保存到事件 `+0x05`，再呼叫 `sub_941C6 @ 0x941C6`。
2. `sub_941C6` 將該軍官 record `+0x39` 寫為 `0xFE`、`+0x37` ETA 寫 0、
   `+0x35` location 寫 `-1`；若原指派是艦艇，亦把該艦 `+0x74` 清為 `-1`。
   這是死亡／不可再使用狀態，不是返回人才庫。
3. 接著 consumer 以同一 ship index 呼叫 `sub_A163A @ 0xA163A`。該函式處理仍引用
   此艦的軍官記錄後，把 `ship+0x65` 寫 `-1`、`ship+0x64` 寫 5，完成單艦移除。
4. consumer 最後只把事件狀態寫成 5。分支沒有呼叫 `sub_3868F`、`sub_39985`、
   `sub_40C2A` 或其他爆炸傷害鏈，也沒有對鄰艦寫損傷。

## 已證實：新聞參數

- `sub_21371` 的事件 8 分支 `0x21579..0x215A6` 讀事件 `+0x05`。若有死亡軍官，
  會啟用額外新聞參數；無軍官時不啟用。remake 訊息應保留被毀艦名稱，並在有軍官時
  明示該軍官一併死亡。

## 勘誤與 remake 投影

- 舊文件把 `sub_3868F` 的戰鬥／殖民地引擎爆炸連鎖套到事件 8，並自行加入
  `Random(201)+74`、每步 `-20`、最多三艘鄰艦受損及「至少保留一艘」護欄。
  事件 8 的直接建立端與 consumer 都沒有這些呼叫或分支；該結論被本次 raw case 資料流推翻。
- remake 沒有原版全局 ship status 0／1 的逐艦欄位；正常存在於玩家 `Fleets[].Ships`
  或 AI `Ships` 的艦艇，是這兩個 active 狀態的 typed projection。這是強推論，不能宣稱
  raw status 逐值相同。
- remake 的領袖 slice 只保存仍可使用的已雇領袖，沒有死亡 record 顯示層；因此從 slice
  移除死亡軍官，是 status `0xFE` 的玩家可見投影。存檔自然保存此結果。
