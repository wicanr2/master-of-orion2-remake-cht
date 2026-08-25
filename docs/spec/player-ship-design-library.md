# 玩家六艦體設計庫規格

## 證據邊界

`Auto_Design_Ship_ @ 0x616A5` 的 caller 已證實每位玩家以
`player*0xEA9 + 0x326 + hull*0x63` 定位六筆 99-byte 設計；hull raw 依序為 0..5。
`Update_Player_Ship_Designs_ @ 0x57112` 亦會把其中一筆交回自動設計函式。詳細證據、輸入雜湊、
IDA 版本與位址基準見 `docs/re/auto-design-ship-audit-20260824.md`。

已證實的是「每位玩家六筆、依艦體定位、可被更新」；八類 role 的完整選擇政策、名稱欄及
`+0x12..+0x16` 全部語意仍未閉合。本規格不得自行發明這些部分。

## Runtime 模型

`GameSession.ShipDesigns` 保存六筆 `ShipBlueprint`，索引與
`gamedata.CombatShipClass`／設計畫面六列一致。每筆至少保存：

- 艦體 key 與 raw Auto Design role；
- 目前相容 UI 使用的武器、裝甲、護盾、特殊裝置索引；
- 武器改造、火線角及彈架容量；
- 可往返的 `WeaponMounts` 與 `SpecialIDs`，為後續逐槽 UI／戰鬥保留資料形狀。

舊存檔沒有設計庫時，依讀檔後當前科技為六種艦體各產生一筆合法 mixed role 模板。這是
明示的存檔遷移政策，不宣稱是原版 role RNG。

## 玩家路徑

1. 第一次進入設計畫面時載入目前選中的艦體設計，預設為巡洋艦。
2. 點六個艦體列只切換設計；切換前先把目前編輯值寫回該筆 blueprint，不造船、不扣 BC。
3. 元件、改造、火線角與彈架每次改動後寫回目前 blueprint。
4. `CLEAR` 把目前設計回復為該艦體的合法 Auto Design 模板。
5. `CANCEL` 保存設計後返回艦隊，不建造。
6. `BUILD` 驗證目前艦體空間與 BC，成功後依目前 blueprint 建造並返回艦隊；失敗留在畫面顯示原因。

## 持久化與多人

- `ShipDesigns` 必須進 remake JSON 快照。
- 它是玩家側狀態，必須隨熱座席位切換；TCP 快照因此自然同步。
- 舊 JSON 的空／短 slice 由六艦體不變量補齊，不覆蓋已有設計。

## 本輪不冒稱完成

- UI 仍只編輯第一武器與第一特殊裝置；既有其餘 raw 槽必須保存，不得因切換畫面被清除。
- 快速結算與格子戰術尚未逐 mount 消費。
- `Update_Player_Ship_Designs_` 的原版科技更新／自動改裝政策與 AI 逐艦設計仍待後續 RE。

## 驗收

- 六筆設計初始化後艦體索引固定、全部合法。
- 修改其中一筆不污染其他艦體；存讀檔與熱座換席後仍保留。
- 點艦體不再直接造船；只有 `BUILD` 會扣款及增加艦艇。
- 全套 Go 回歸測試通過。
