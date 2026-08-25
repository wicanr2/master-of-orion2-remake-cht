# 同化進度與 Charismatic／Repulsive 種族效果稽核（2026-08-25）

## 問題

`internal/gamedata/assimilation.go` 依手冊把 Charismatic 留成無效果參數，並以「每人口所需回合」保存同化進度。本次確認原版是否有精確倍率及持久化尺度。

## 證據契約

- 輸入：`Orion2.exe`
- SHA-256：`7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- 資料庫：既有 `Orion2.exe.i64` 的一次性可寫副本
- 工具：IDA Pro 9.4、`tools/ida/audit_assimilation_traits.py`
- 位址基準：IDA linear，DOS/4GW LE object #1
- 註記方式：不改函式名、全域名、位址、運算元或資料庫；附加語意僅在本受版控文件。

## 已證實

1. `sub_E3456 @ 0xE3456` 是逐殖民地同化 consumer；唯一直接 caller 為 `sub_E3FDC @ 0xE3FDC` 內的 `0xE3FBA`，而後者由回合鏈 `0x13747` 呼叫。
2. `0xE347E` 只處理 `colony+0x12F == 3`；`0xE348B` 讀 `colony+0x12E` 為既有進度，`0xE360B` 寫回低 byte。
3. `byte_DD4F5 @ 0xDD4F5` 的八 bytes 是 `1e 3c 1e 3c 3c 78 0c 10`，即八政體每回合進度 `30/60/30/60/60/120/12/16`。門檻在 `0xE353E` 為 `0xF0 = 240`，每跨一次門檻就清除一筆 packed colonist 的 prisoner bit。
4. `colony+0x137 != 0` 時，`0xE3501` 先把進度改為 120；這是既有 Alien Management Center consumer。
5. `0xE3516` 讀 player `+0x8B3`；非零時 `0xE351F add edx,edx`，所以 Charismatic 將本回合同化進度加倍。
6. 只有 Charismatic 為零才到 `0xE3523` 檢查 player `+0x8B2`；非零時 `0xE352C..0xE3533` 以有號除法把進度減半。因此 Charismatic 優先，兩項異常同存時不會先乘二再除二。
7. `0xE3538` 把本回合進度加到 `colony+0x12E`；`0xE35FC` 每成功同化一人口減 240，可在同回合跨越多次。

## 行為推導

- 基礎政體回合數為 `ceil(240/rate)`，恰好得到手冊的 `8/4/8/4/4/2/20/15`。
- Charismatic 的精確效果是「進度率 ×2」，不是籠統的回合數除二；銀河統一為 `ceil(240/32)=8` 回合。
- Repulsive 是「進度率 ÷2」；銀河統一為 `240/8=30` 回合。
- 異族管理中心也先套 120，再受種族分支影響：一般 2、Charismatic 1、Repulsive 4 回合。

## Remake 對應與剩餘限制

- 應把 `AssimilationProgress` 改成原版 0..239 進度點，逐回合加入精確 rate，跨 240 同化一人口。
- 舊 JSON 的同欄位是「已累積回合」，必須以讀檔當下的 rate 轉成進度點並保存格式版本；不可直接重解為 raw。
- remake 的人口群組不保存原版 packed colonist 順序，因此「清除哪一筆 prisoner」仍採既有可重播近似；同化人數、速率、餘數、士氣與叛亂 consumer 可精確接線。
