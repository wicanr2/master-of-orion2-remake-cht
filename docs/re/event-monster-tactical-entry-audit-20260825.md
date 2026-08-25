# 事件怪物格子戰術入口稽核（2026-08-25）

## 證據

- 輸入與位址契約同 `event-monster-colony-battle-audit-20260825.md`：`Orion2.exe` SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`，IDA Pro 9.4，
  IDA linear、DOS/4GW LE object #1。
- `Search_For_Battles_ @ 0xE9D62` 對 owner 8 與一般帝國走同一 `Do_1_Combat_ @ 0xE938C`；
  `sub_E6A0C @ 0xE6A0C` 收集同星停泊艦艇，戰後再由 `sub_E87D2 @ 0xE87D2` 依 owner/type
  分派怪物消費端。這證實怪物不是只能走獨立快速傷害腳本。
- `Load_Combat_Ship_ @ 0x4954A` 對 raw type 10..14 讀同一 99-byte design 並建立一般 combat
  record；結構、裝甲、戰速與逐槽武器已由 `event-monster-blueprints-audit-20260825.md` 閉合。

## 現況反例

`cmd/moo2/interactive.go` 的星系動作 `attackmonster` 直接呼叫 `GameSession.AttackMonster`，完成
六回合抽象解算後返回星系畫面；`newTacticalScreen` 只會呼叫 `StartCombat(PrimaryEnemyName())`。
因此玩家正常路徑無法把怪物送入既有格子戰術，是垂直鏈缺口，不是單純缺測試。

## 邊界

- 原版怪物 sprite 的 loader picture 8..12 已證實；picture 與 remake CMBTSHP asset index 的
  逐像素對照仍未完成，先使用同 raw picture 作可追溯 adapter，不宣稱 sprite parity。
- 戰術移動 AI 仍屬近似；後續已閉合 raw `0x4000` 為 OVR，並接入 Dragon Breath 每格
  -15。原始未知狀態的勘誤與證據見 `event-monster-dragon-raw4000-audit-20260825.md`。
