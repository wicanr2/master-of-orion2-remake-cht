# AI 對 AI 實艦戰鬥、傷亡與回寫稽核（2026-08-28）

## 證據邊界

- 輸入：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- IDA 資料庫：`Orion2.exe.i64`，SHA-256
  `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- 工具：IDA Pro 9.4／IDAPython；位址為 IDA linear、DOS/4GW LE object #1。
- 15 個根函式的原始名稱、bytes、指令、caller／callee 與外部符號導覽名見
  [`evidence/ai-ai-ship-battle-ida-20260828.json`](evidence/ai-ai-ship-battle-ida-20260828.json)。
  本文以原始指令為證據；反編譯稿與外部符號名稱只供導覽。

## 已證實：戰鬥排程與實艦清單

1. `Search_For_Battles_ @ 0xE9D62..0xEA27A` 先從每艘 129-byte ship record 的 owner、type、
   status 與 star 建立逐星 side mask；殖民地 owner 也加入同一逐星表。它洗牌有戰鬥的星系，
   再由 `Pick_Star_And_Attacker_ @ 0xE84A5` 逐場取星系與攻擊 side。這條排程沒有 AI 對 AI的
   抽象旁路。
2. 攻擊 side 小於 8 且不是 local human 時，`Get_AI_Target_ @ 0xE7DCA` 從該星現存 side、
   方向情報與強度表選 defender；候選可以是另一個 AI。選定後與真人、事件怪獸共用
   `Do_1_Combat_ @ 0xE938C`。
3. `Get_Combat_Ships_ @ 0xE6A0C` 掃描全部實際 ship records，只收
   `ship+0x65 == star`、`ship+0x64 == 0` 且 owner 相符的 ship ID。攻方從陣列前端加入，守方
   從 500 項陣列尾端反向加入，並各自回傳精確數量。故原版 AI 對 AI 的參戰者是實際 ship ID，
   不是 `FleetStrength` 比例或另擲一個「損失艦數」。

## 已證實：AI 對 AI 固定走戰略解算

4. `Russ_Combat_ @ 0xE7343` 先讀全域戰鬥模式；若設定為格子戰術（raw 0），但攻方不是真人，
   它會把 local selector 改為「守方是否不是真人」。AI 對 AI 時雙方均非真人，因此 selector
   成為 1，固定呼叫 `Strategic_Combat_ @ 0x40148`。只有至少一側是真人且相應 gate 成立時，
   才可能呼叫 `Tactical_Combat_ @ 0x47939`。
5. 因此 AI 對 AI 不需要另找格子戰術 AI 的自動操作 oracle；原版把這類戰鬥強制交給 33-byte
   combatant 的戰略解算。`Russ_Combat_` 將結果正規化為攻方勝 1、平局 0、攻方敗 -1，交回
   `Do_1_Combat_` 做撤退、非戰鬥艦與殖民地後續處理。

## 已證實：傷亡不是隨機選艦

6. `Strategic_Combat_` 以兩側 ship ID 清單建立 33-byte combatants；每筆一般艦 combatant
   保存原始 ship ID。回合內被標成 destroyed 的 combatant，戰後直接取該 ID 呼叫
   `Kill_Ship_ @ 0xA163A`。沒有從艦隊倖存清單再做 reservoir sampling、依艦級比例刪艦或
   隨機挑一艘代替實際被摧毀者。
7. 倖存一般艦會回寫 `ship+0x7D`：戰略模式直接寫 combatant `+0x1F`；格子戰術才按剩餘結構
   與 `Get_Ship_Structure_` 換算。AI 對 AI固定使用前者。若艦上有軍官，死亡、回人才庫或轉派
   依軍官狀態與剩餘同 owner 艦處理，並非忽略 AI 軍官。
8. `Kill_Ship_` 先處理艦圖與已指派軍官，再把該 ship record 的 `+0x65` 寫 -1、`+0x64`
   寫 5。這是實艦死亡的持久回寫；不會只扣帝國抽象戰力。
9. 敗方非戰鬥艦由 `Kill_Noncombat_Ships_ @ 0xE6B44` 逐 ID 呼叫 `Kill_Ship_`。需要撤退時，
   `Determine_Retreat_Ships_ @ 0xE6CAA` 把同 owner、同 star、active 的每艘實艦 status 寫 9；
   hyperspace flux 且無跨維度能力時則逐艦死亡。`Process_Retreating_Ships_ @ 0xE6E52` 隨後把
   status 9 艦群尋路到合法星系，成功者改成航行狀態並調整目的編碼，無路者逐艦死亡。

## 格子戰術交叉驗證

`End_Of_Combat_ @ 0x4B184` 同樣讓 destroyed 的 313-byte combat ship 依其原始 ship ID 呼叫
`Kill_Ship_`，倖存者回寫戰損與戰後 XP。這證明兩種解算都採「實際被摧毀 combatant → 原始
ship ID」契約；但它不是 AI 對 AI 的可達分支，不能拿來要求 remake 為 AI 對 AI開啟格子戰術。

## 強推論與未知

- `Get_AI_Target_` 內部的兩組暫存強度陣列與 `sub_E76B2／sub_E78A7` 已確認直接決定 defender，
  但每個暫存槽正式名稱尚未逐欄命名；這不影響本切片的參戰實艦、死亡與回寫結論。
- ship status raw 5／9 的正式 UI 字串尚未由文字資產閉合；其 producer、consumer 與資料流已證實。
- remake 現行 AI 對 AI 仍以 `FleetStrength` 比例先算損失，再刪除若干實艦；它不是原版算法。
  依 RE-first gate，本輪只登記差異，不修改 Go／Ebitengine 行為。
