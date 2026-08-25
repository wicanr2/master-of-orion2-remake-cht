# 隨機事件 27「曲速漏斗」靜態稽核（2026-08-25）

## 證據契約

- 輸入：`Orion2.exe` 1.31，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- 工具：IDA Pro 9.4；IDA linear，DOS/4GW LE object #1
- 唯讀匯出器：`tools/ida/audit_event_warp_funnel.py`
- 原始定位均保留；下列語意名稱只供導覽，不取代位址與運算元。

## 已證實

1. `sub_2230A @ 0x2230A` 在事件 27 與事件 26 都先呼叫
   `sub_23CED @ 0x23CED`。後者在目標帝國所有狀態為 0／1 的艦艇間做
   reservoir sampling；沒有候選艦時回傳 `-1`，事件失敗。
2. 事件 27 的 jump-table 分支 `0x22D2F` 把九位元組 record 寫成：
   state=`3`、age=`0`、`word +3`=所選全局艦艇索引。slot 27 的固定範圍是
   `0x19AC97..0x19AC9F`，通用基址為 `0x19ABA4 + eventID*9`。
3. `sub_206A2 @ 0x206A2` 的事件 27 分支 `0x212BE` 只在 active state=`2`
   時跳入共用生命週期：age 大於 4 後，每回合以 `sub_1247A0(20)==1`
   將 state 設為 5；age 大於 `0x14` 時無條件設為 5；最後 age 加一。
4. `sub_21371 @ 0x21371` 的事件 26／27 共用新聞狀態分派 `0x21902`。
5. `Move_All_Ships_Toward_Stars_ @ 0xFFEEA` 完整函式沒有讀取事件 record，
   也沒有呼叫事件 27 helper；它照常更新艦隊座標與航行倒數。
6. `sub_206A2` 的事件 27 分支沒有改寫所選艦艇 record、艦隊位置或航行倒數。
   所選艦艇索引只作事件目標／新聞資料；事件結束訊息亦明說船艦無損。

## 強推論與限制

- **強推論：**1.31 的事件 27 是「報告型持續事件」，玩家可見效果是受困／脫困
  新聞與持續時間，沒有實際停止艦隊航行。依據是建立端、完整逐回合 consumer、
  新聞 dispatcher、完整艦隊移動函式，以及事件 record 通用基址的直接交叉參照。
- IDA 直接交叉參照無法證明不存在所有可能的暫存器間接讀取；因此不把「整個程式
  絕無其他間接 consumer」升格成已證實。不過事件垂直鏈已足以界定 remake 行為，
  不再為玩家不可見 helper 擴大逆向範圍。
- 尚未以 1.31 DOSBox 動態 oracle 逐回合觀察艦隊；這不阻塞靜態證據支持的 remake。

## 對 remake 的訂正

舊 `gamedata.RandomEvents[27].Needs` 寫成「艦隊受困與脫困判定」，會暗示需要自行
凍結 ETA。這與 1.31 靜態鏈不符。remake 應保留原版的候選艦閘門、持續 record、
第 6 個 active turn 起 1/20 脫困及第 21 個 active turn 後強制結束，但不得創造
原版消費端未出現的停航效果。
