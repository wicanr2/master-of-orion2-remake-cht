# 多槽設計建造與快速戰鬥規格

日期：2026-08-24

## 證據邊界

- **已證實**：原版每筆 99-byte 設計自 `+0x1C` 起最多八個 8-byte 武器槽；每槽保存武器、
  數量、可用數量、射界、改造及彈藥。證據見
  [`../re/auto-design-ship-audit-20260824.md`](../re/auto-design-ship-audit-20260824.md)。
- **已證實**：單槽數量超過上限時，原版 helper 會拆成下一槽；因此槽與槽內數量都是玩家
  火力資料，不能只保存而不消費。
- **未知**：現行 `ShipWeaponMount.RawMods` 的每個 raw bit 尚未全部對回 remake 的字串改造；
  未解 raw bit 只保存，不猜成別的改造。
- **remake 限制**：既有戰鬥把艦體攻擊與武器傷害合成單一 `atk/wmin/wmax`。本規格沿用該
  已存在的單發公式，只把每個有效槽與 `WorkingCount` 展開成實際射擊，不冒稱原版逐武器
  命中／傷害已精確重建。

## 建造契約

1. `BuildShipDesign(hull)` 必須把所選 `ShipBlueprint.WeaponMounts` 與 `SpecialIDs` 深複製到
   新建 `Ship`；建造後再改 blueprint 不得污染既有艦艇。
2. 空槽（空名稱、`WorkingCount <= 0`）不產生射擊；舊 blueprint 沒有多槽資料時，依相容
   `Weapon/Arc/Ammo` 產生一槽，維持舊存檔行為。
3. 第一個有效槽繼續同步 `Ship.Weapon`、`WeaponAttack`、`Arc`、`WeaponAmmo`，供尚未遷移的
   UI 與戰術路徑安全回退。
4. 艦體、裝甲、護盾與特殊系統只計一次；每個有效武器槽先依自己的武器、彈架、射界、
   已解改造與科技微型化求單門成本／空間，再乘設計數量 `MaxCount`。
5. 已知武器但尾槽含尚未解碼的 `RawMods`，或 raw ID 無法對到已知武器時，BUILD 採
   失敗即關閉（fail-closed），不得把未知裝備當免費零成本。
6. UI 顯示、空間 gate、實際扣款與重播必須共用 blueprint 計算；多槽建造命令攜帶完整
   blueprint，不依賴遠端剛好具有相同的編輯暫存狀態。

## 玩家編輯契約

1. 六筆 hull blueprint 各自最多八個武器槽；畫面必須明確顯示目前槽與總槽數。
2. 選槽只切換編輯焦點；武器、射界、彈架與改造寫回目前槽。裝甲、護盾與特殊系統仍是
   整筆設計共用欄位。
3. 新增槽複製目前可表達的武器設定並把數量設為 1；刪除後至少保留一槽。第一槽改變時
   同步舊相容欄位，供舊存檔／尚未遷移顯示使用。
4. 槽內數量可在 1..99 調整；顯示與命中區域共用同一組 rectangle，所有中英文標籤必須
   落在 text-safe rectangle 內。

## 特殊裝置槽契約

1. 原版 special bitset 轉成最多八筆 `ShipSpecialMount{RawID, Name}`；可對回
   `gamedata.SpecialDevices` 的 raw ID 使用受版控名稱，未知 ID 保留 `原版特殊#N` 並標未知。
2. remake 原先放在特殊下拉、但原版屬武器／機庫／獨立電腦欄的項目以 `RawID=-1` 保存；
   這只是 typed remake 載體，不冒稱原版 special bit。
3. 第一筆繼續同步 `Ship.Special`／`ShipBlueprint.Special` 相容欄位；所有效果 consumer 必須
   透過 `shipHasSpecial` 搜尋完整 typed slots，舊存檔沒有 slots 時才回退相容欄位。
4. 每項特殊裝置的成本、空間與科技微型化各算一次；同一名稱不可重複安裝。未知 raw ID
   不能安全計價，BUILD 採失敗即關閉。
5. UI 可在最多八槽間切換、加入與移除；循環選項必須跳過其他槽已安裝的名稱。「無」只作
   空槽編輯值，不產生 raw ID、成本、空間或效果。
6. 建造、JSON、熱座、TCP、完整 blueprint command 與 `.GAM` importer 均保存 typed slots
   及原始 `SpecialIDs`；兩者衝突時 raw ID 不得被丟棄。

## 快速戰鬥契約

1. 一艘艦仍只有一份 HP、護盾、裝甲、軍官、特殊系統與主動權；不得把每個武器槽展開成
   多艘可被攻擊的假船。
2. 輪到該艦時，依槽順序逐槽處理；每槽射擊次數為 `max(0, WorkingCount)`，再乘既有的
   連射系統次數。
3. 每槽各自使用名稱、類型、傷害、射界與彈藥；目前無法解碼的 raw mods 不套效果。
4. 飛彈槽各自扣自己的彈藥。炸彈在艦隊戰仍不開火；空槽不消耗 RNG。
5. 相位匿蹤、儲能、登艦等艦級一次性狀態不得因槽數重複重置。

## 格子戰術契約

1. `StartCombat` 把 `Ship.WeaponMounts` 深複製到同一筆 `CombatShip`；沒有多槽資料的舊艦
   繼續使用相容武器欄位。
2. 只有槽數大於一或任一槽 `WorkingCount > 1` 時才走多槽派送；單槽路徑保持原有亂數消耗
   與畫面結果不變。
3. 每個槽使用自己的名稱、類型、射界、傷害與彈藥，`WorkingCount` 表示該槽同型武器數量；
   連射系統仍由既有 `TacticalShotsThisRound` 套在每一門武器上。
4. 儲能釋放、匿蹤解除、`Fired`、點防禦已用狀態與目標戰損都保留在同一筆 `CombatShip`，
   逐槽派送不得複製或還原這些艦級狀態。
5. 第一槽可沿用已解碼的 `Mods`；尾槽 raw mods 在位元對照完成前不套未證實效果。

## 驗證

- blueprint 建造後所有槽／特殊 ID 深複製往返；
- 兩個有效槽在快速齊射中都能造成傷害；第二槽設為零工作數時不開火；
- 兩個飛彈槽分別扣彈，不共用第一槽彈架；
- 格子戰術兩個有效槽均開火，射界封鎖／彈藥耗盡各自不污染另一槽；
- 格子戰術逐槽開火後匿蹤只解除一次、儲能只釋放一次；
- 兩槽成本／空間等於各槽單門結果乘數量後相加，再加一次艦體與共用設備；
- 未知 raw 武器／mods 被 BUILD 拒絕且不扣款；
- 多槽 BUILD 命令重播產生相同艦艇、扣款與槽資料；
- UI 可選槽、增刪槽與調整 1／99 邊界，所有控制文字通過幾何 containment；
- `.GAM` 已知／未知 special bits 均轉成 typed slots 並保留 raw ID；
- 同艦同時裝兩種特殊系統時，快速／格子戰術各自可觀察的效果都成立；
- 多特殊成本／空間逐項相加，重複名稱被拒絕，未知 raw ID fail-closed；
- 特殊槽 UI 增刪／選槽／去重與相容欄位同步通過測試；
- 單槽既有戰鬥測試維持逐項通過。

## 後續 gate

- 格子戰術玩家逐槽選擇與更細的逐槽畫面回饋；
- 戰後逐槽損傷／修復，以及 AI 逐艦多槽 blueprint。
