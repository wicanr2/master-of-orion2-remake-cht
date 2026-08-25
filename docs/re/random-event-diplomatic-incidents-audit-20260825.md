# 隨機事件 4／5「外交暗殺／外交聯姻」靜態稽核（2026-08-25）

## 證據身分

- 輸入：`Orion2.exe`
- SHA-256：`7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- 工具：IDA Pro 9.4、`tools/ida/audit_event_diplomatic_incidents.py`
- 位址基準：IDA linear；DOS/4GW LE object #1
- 方法：唯讀 `.i64` 函式、caller、原始跳表 dword 與指令匯出；未改名、未修改資料庫。

## 已證實：第二帝國目標

1. `Determine_Event_ @ 0x2230A` 在事件 4／5 共用分支 `0x225DC..0x22618`，以一般事件
   已選出的受害帝國槽呼叫 `sub_23B7D @ 0x23B7D`。
2. `sub_23B7D` 依帝國槽順序掃描，候選必須：
   - `player+0x24 == 0`，即帝國仍存續；
   - 與受害帝國不同；
   - 受害帝國 `player+0x584` 的接觸 bitset 對該槽為 1；
   - 每遇一個候選便以 `Random(candidateCount)==1` 替換目前答案，形成 reservoir sampling。
3. helper 回傳後，`0x225FC..0x22612` 才要求受害帝國對第二帝國的 `player+0x627 >= 4`。
   專案既有 raw enum 已證實 4／5 分別為有限戰爭與戰爭。因此事件只在交戰中的兩帝國間成立；
   但和平接觸者並非事前移出 reservoir：若 helper 抽到和平帝國，本次事件候選直接失敗，
   不會在剩餘戰爭對象中重抽。這項先抽後驗順序會影響事件成立率與亂數消費。
4. `Determine_Event_` 原始跳表第 4／5 格都到 `0x229FC`，把第二帝國槽寫入事件 record
   `+0x03`。不是在效果結算時重新任選一個 AI。

## 已證實：關係效果與新聞參數

1. `sub_206A2 @ 0x206A2` 的 consumer 原始跳表第 4 格到 `0x209BE`，第 5 格到
   `0x209F1`。兩者都以事件 `+0x00` 的受害帝國為 actor、record `+0x03` 為另一帝國，呼叫
   `Change_Relations_ @ 0x4E3B5`；額外 reason／payload 參數都是 0。
2. 傳入 `EAX` 的原始基礎關係變化：
   - 事件 4：`0xFFFFFF9C = -100`；
   - 事件 5：`0x00000064 = +100`。
3. `Change_Relations_` 的事件可達資料流會依目前 signed-byte 關係值調整增量，再套 actor
   政體：raw 4 對負值 ×2，raw 0 對負值 ×3/2、正值 ×3/4；第二帝國若有
   `player+0x8B3` 魅力旗標，正值 ×2、負值 ÷2。最後把關係夾在 `-100..100`。
4. 正式狀態 4／5 時，結算後關係若高於 `-25`，`0x4E987..0x4E9A3` 強制寫回 `-25`。
   因此外交聯姻會大幅改善交戰關係，但不會自行終止戰爭或使分數跨過 `-25`。
5. `sub_21371 @ 0x21371` 的 case 4 把 record `+0x03` 放入新聞參數；case 5 亦使用同欄，
   並另放常數 100。EVENTMSG 的原文同時引用兩帝國，remake 不應只顯示模糊的「某個 AI」。

## Remake 投影與剩餘未知

- remake 現有 `Relation`／`AIRelations` 是 `-40..40` 的正規化模型，原版欄位是 signed byte
  `-100..100`。本切片採可逆尺度 `raw = normalized×5/2`、`normalized = raw×2/5`，在 raw
  尺度執行上述事件可達公式後才轉回；這是資料模型投影，不是 raw byte 逐值保存。
- 玩家／AI 正式戰爭可由 `Treaty.FormalPolicy` 表示，AI／AI 由 `AIPolicies` 表示。熱座真人
  彼此目前沒有外交關係矩陣，因此該配對不能成為第二目標；這是明示資料模型缺口。
- remake 沒有逐帝國接觸 bitset；正常對局把每條可表示的外交關係邊視為已接觸，先納入
  reservoir，再依正式狀態驗證。這保留和平對象可使候選失敗的順序，但「可表示邊＝已接觸」
  仍是強推論。
- 1.50 二進位尚未取得，本頁只證實 1.31 行為。
