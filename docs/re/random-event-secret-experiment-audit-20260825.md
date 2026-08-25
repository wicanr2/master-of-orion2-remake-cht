# 隨機事件 18「秘密實驗」靜態稽核（2026-08-25）

## 證據身分

- 輸入：`Orion2.exe`
- SHA-256：`7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- 工具：IDA Pro 9.4、`tools/ida/audit_event_secret_experiment.py`
- 位址基準：IDA linear；DOS/4GW LE object #1
- 方法：唯讀 `.i64` 函式邊界、交叉參照與指令匯出；未改名、未改資料庫。

## 已證實：建立與目標

1. `Determine_Event_ @ 0x2230A` 的事件 18 不具有個別建立公式；它走一般好事件的帝國
   目標選擇鏈。事件 record 的 `+0x00` byte 是被選帝國槽。
2. 本事件沒有殖民地、星系或艦隊次目標，也不是跨回合持續事件。

## 已證實：科技完成與研究回寫

`sub_206A2` 的事件 18 分支位於 `0x20ECA..0x20F64`：

1. `0x20ED0..0x20EE8` 先以事件 `+0x00` 帝國槽乘 `0xEA9`，讀取
   `player+0x321` 的目前研究 field；`0x20EEC..0x20EF8` 將該 word 保存至事件
   `+0x03`，供後續新聞顯示。
2. `0x20EFF..0x20F1C` 再取得同一帝國與 field，以 `ebx=1` 呼叫
   `sub_E4410 @ 0xE4410`。既有研究稽核已證實該函式是研究 field 完成／application
   授予鏈：一般 field 授予已選 application；固定全解或 Creative 分支授予全部；
   Hyper-Advanced field 增加等級。
3. `0x20F21..0x20F40` 將目標帝國 `player+0x1EB` 的研究進度 dword 清為 0；
   `0x20F45..0x20F5F` 將 `player+0x321` 的目前研究 field byte 清為 0。
4. 分支中沒有加上固定 RP、回合數或隨機 RP 的算式。因此 remake 原有的
   `80 + Turn` 是無原版證據的舊實作，必須移除。

## 已證實：展示資料

- `sub_21371` 的事件 18 分支位於 `0x2171D..0x21730`，讀取事件 `+0x03` 保存的
  field 作為 GNN 訊息參數。remake 應在完成前保存主題，訊息不得在清空目前主題後
  退化成未知科技。

## Remake 投影與證據界線

- `player+0x321` 投影為 `PlayerState.ResearchTopic`，`player+0x1EB` 投影為
  `PlayerState.ResearchProgress`；科技完成結果沿用既有 `CompletedTopics`、
  `ChosenTech`、`ExplicitChoice`、`GrantedTechs` 與 `HyperAdvancedLevels`。
- 熱座席位與 AI 都必須套用同一規則；這是原版 player 槽陣列到 remake 分離狀態的
  型別投影，不改變事件語意。
- `sub_E4410` 內部的所有快取重算與 Win95 顯示 callback 不逐一翻譯；remake 只執行已
  接線的玩家可見科技 callback 與自動艦型更新。此為可玩性停止線，不冒稱逐指令 parity。
- field 為 0 時原版仍進入完成呼叫後清空狀態；靜態證據不足以證明會授予有效科技。
  remake 保留事件成功與訊息，但不得捏造科技或 RP。
