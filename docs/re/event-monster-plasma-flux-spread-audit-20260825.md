# Plasma Flux 範圍傷害稽核（2026-08-25）

## 證據身分

- 輸入：`Orion2.exe`
- SHA-256：`7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- 工具：IDA Pro 9.4；IDA linear、DOS/4GW LE object #1
- 資料庫：`moo2-Orion2-consumer-work.i64` 的一次性副本
- 唯讀匯出：`tools/ida/audit_plasma_flux_spread.py`
- 資產交叉驗證：完整版 `CMBTSFX.LBX`，asset 6 解碼為 192×192、12 幀。

## 勘誤

舊文件把 `sub_289C4 @ 0x289C4` 的 ID 44 分支當成命中傷害消費端。交叉參照證明它的
callers 位於戰術 AI 評估鏈；函式按武器類別累加估值，並未寫入目標 combat record。
其中「平均傷害 × 4 × 目標尺寸」只能說明 AI 評分，不是 runtime 傷害公式。

## 已證實 runtime

1. `sub_38B5E @ 0x38B5E` 對特殊武器類別呼叫 `sub_3A82F @ 0x3A82F`。
2. `sub_3A82F @ 0x3ABF1..0x3ABF9` 對 ID 44 設 `ecx=2`，把射手、座標與該槽有效數量
   傳入 `sub_ADE18 @ 0xADE18`，之後清空本次可射數量。
3. `sub_ADE18 @ 0xADFB2..0xAE0CD` 讀取 effect type 2 的 `CMBTSFX.LBX` asset 6，令
   半徑平方為 `(width/2)^2`；它枚舉 combat record 1..`word_1998C0-1`，排除射手、死亡／
   非 active 狀態與 `sub_3E563` 排除者，沒有 owner 過濾，所以友軍與敵軍都可能受害。
4. 實際 asset 6 寬 192，故半徑為 96 像素。原版座標與格位換算使用 20 像素，因此不是
   舊待辦所稱固定六格，而是約 4.8 格；判定直接比較歐氏距離平方，邊界包含等號。
5. `0xAE793..0xAE9FD` 對 type 2 先把該槽每門武器的隨機傷害加總並套射手 ordnance
   百分比；對每艘目標以 `100-distanceSquared/radiusSquared*100` 衰減，結果最低夾 1。
6. `0xAEA1E..0xAEAD1` 依目標 combat record `+0x25 + 1` 重複擲值；每段先擲
   `Random(attenuated)`，若結果 `<=1` 就取 1，否則再擲一次同範圍並累加。最後按射手到
   目標的護盾朝向呼叫 `sub_39985`，不是把尺寸直接乘在單一固定傷害上。

## 飛行物分支

1. `sub_ADE18 @ 0xAE61E..0xAE88C` 掃描全部 300 筆、每筆 26-byte 的飛行物 record。
   weapon category 2 直接排除；category 4 先擲 `Random(2)`，結果 1 時略過整筆，與手冊
   「戰機有 50% 機率避開球形武器」互證。
2. 其餘 active record 仍使用相同 96px 歐氏半徑與距離平方衰減。effect type 2 的中心
   基礎傷害同樣聚合 ID 44 mount 全部門數。
3. `0xAE7F9..0xAE887` 計算
   `killChance = 25 × attenuatedDamage / sub_3E095(weaponID, record+2, record+0x11)`；
   接著對 record `+0x0F` 的每一架各擲一次 `Random(100)`，小於等於機率就將數量減一。
4. `sub_3BB3D @ 0x3BB3D` 只是讀取 26-byte record `+2`。`sub_3E095 @ 0x3E095`
   對 ID 28／29／30／31 分別以 6／10／8／4 乘裝甲係數再除 200，對應四種戰機的
   3／5／4／2 基礎單架耐久；remake 的 `FighterSquadron.HPEach` 可作同構分母。

## 資料模型邊界

- 原版用精確像素中心；remake 只保存格位，故以格中心 `20×(col,row)` 計算同一平方距離，
  屬可重播的資料模型近似。
- remake 的戰機有持久場上中隊，可接同一消費端；飛彈則在射擊函式內同步解算，沒有原版
  26-byte 在途 record。故可閉合戰機，但不能在不重做飛彈飛行狀態模型時宣稱在途飛彈 parity。
