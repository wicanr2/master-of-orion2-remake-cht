# 事件怪物同星多群組稽核（2026-08-25）

## 證據身分

- 輸入：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- 工具：IDA Pro 9.4；IDA linear、DOS/4GW LE object #1
- 本輪唯讀匯出腳本：`tools/ida/audit_monster_same_star_selection.py`
- 主要函式：`Search_For_Battles_ @ 0xE9D62`、`sub_E8194 @ 0xE8194`、
  `sub_E84A5 @ 0xE84A5`、`sub_E8029 @ 0xE8029`

## 已證實

1. `Search_For_Battles_ @ 0xE9E68..0xE9EA0` 對 raw owner/type `>=8`，以
   `1 << (owner/type-8)` 寫入星系 bitset；type 10..14 是彼此獨立的 side bit，不會因位於
   同一星系便合併成單一怪物種類。
2. `sub_FE9B5 @ 0xFE9B5` 先以 Fisher–Yates 形狀洗牌有戰鬥的星系清單；
   `sub_E84A5 @ 0xE84A5` 逐項消費該清單，呼叫 `sub_E8194` 從該星尚未清除的
   owner/type bit 做 reservoir sampling，隨即清除選中的 bit。外層
   `while (word_1AA3E8 > 0)` 會持續處理，並非只回傳一個 side 後捨棄其餘項目。
3. side 已選定後，`sub_E82B4 @ 0xE82B4` 才建立可交戰的艦隊／殖民地候選；
   `sub_E8029` 只在已選 side `>= 8` 時從這些候選選對手。它不是怪獸 side selector。
4. 因位元遮罩以 raw owner/type 為鍵，同星不同種類怪獸是分開的戰鬥 side；同 raw type
   的多艘 active ship record 則屬同一 side，戰鬥資料仍逐艦保存，不會覆蓋彼此。
5. 原版 active ship record 逐艦保存 owner/type、所在星與狀態；多艘同星不會覆蓋彼此。
   remake 的 `[]MonsterGuard` 形狀足以保存多群組，缺口位於只回傳單一值的查找／顯示／刪除 API。

## 近似與未知

- 原版 RNG 決定的是全銀河自動戰鬥排程內的星系與 side 處理順序；remake 的「攻擊怪獸」是
  玩家對目前星系主動發起一次戰鬥，沒有等價排程語境。此介面轉接採 `Monsters` 穩定順序選
  第一個停泊群組，確保鎖步重播與戰後回寫指向同一筆 record；不冒稱原版排程順序。
- 本切片不把不同種類合併進同一格子戰術畫面；這與原版 side bit 分離證據一致。

## 推翻的舊假設

- 「每顆星最多只會有一個 `MonsterGuard`」不成立：多個 owner 8 航行 record 可在不同回合
  抵達同一星，且原版以 side bit 分開保存。
- 「`sub_E8029` reservoir sampling 選出同星怪獸 side」不成立：實際 side 抽樣在
  `sub_E8194`；`sub_E8029` 位於 side 已選定後的對手選擇層。
- 只用 `StarIndex` 刪除怪物不安全：同目的星可能同時有航行中與已停泊 record；刪除必須至少
  限定 `TransitETA == 0` 且只移除本次選中的群組。
