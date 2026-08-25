# 事件怪物殖民地戰鬥與戰後回寫稽核（2026-08-25）

## 證據身分

- 輸入：`Orion2.exe`（1.31），SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- 工具：IDA Pro 9.4、`ida-pro-9.4-idapython:py312-v1`。
- 位址基準：IDA linear、DOS/4GW LE object #1。
- 可重生匯出器：`tools/ida/audit_monster_colony_battle.py`；JSON 位於
  `/tmp/moo2-monster-route-20260825/monster-colony-battle-v3.json`，不加入 Git。
- 本文只把原始函式名當定位；語意由指令、欄位讀寫及 caller/callee 資料流分級。

## 已證實

1. `Search_For_Battles_ @ 0xE9D62` 把停泊 owner/type `>=8` 寫入每星 owner bitset；選到
   owner 8 一方時呼叫 `sub_E8029 @ 0xE8029` 取得殖民地／艦隊對手，接著與一般帝國相同呼叫
   `Do_1_Combat_ @ 0xE938C`，最後呼叫 `sub_E6AA2 @ 0xE6AA2` 更新星系 presence bitset。
2. `sub_E8029` 對 raw type 13（太空鰻）在 `0xE8039` 立即返回，兩側索引維持 `-1`；因此太空鰻
   不建立殖民地戰鬥。其他事件怪物會在該星的殖民地候選中 reservoir sampling；若同時有前哨
   與一般殖民地候選，優先一般殖民地，沒有才退回另一池。
3. `Do_1_Combat_` 以 `sub_E6A0C @ 0xE6A0C` 收集同星停泊艦艇，owner 8 與一般帝國使用同一
   combat core。戰後唯一直接呼叫 `sub_E87D2 @ 0xE87D2`；owner/type 8..14 進入其
   `0xE8853` 怪物／安塔蘭分支，而不是一般帝國的登陸、俘虜或外交分支。
4. 怪物分支建立 0x49-byte 殖民地戰果 record：`+0x0F` 保存 raw type，`+0x10` 保存殖民地索引。
   raw type 10（變形蟲）另把 `word +0x12=20`、`word +0x15=0`。其後所有 type 10..14 都呼叫
   `sub_4267B @ 0x4267B`，再呼叫 `sub_DD2F2 @ 0xDD2F2` 消費戰略殖民地傷害。
5. `sub_4267B` 依戰果 record 找到殖民地所在星，再掃 500 個 ship slot，只收同星、停泊、
   `ship+0x11==0` 且 raw type 與 record 相同的艦艇。戰略模式呼叫
   `Strategic_Bombardment_ @ 0x4257E`，將回傳總傷害除以 40 後寫入 `word +0x03`，並把
   `word +0x06` 清零。
6. `Strategic_Bombardment_` 固定執行三外圈；因此同星同 type 的所有存活怪物共同轟炸，不是
   每隻怪物各自另開一場。傷害 `/40` 後進入既有 `sub_DD2F2/sub_DCEBD` 候選池，會回寫一般
   建築、陸戰隊、戰車、建造進度與人口。
7. 怪物分支在戰果 `+0x51` 非零時呼叫 `sub_DCDAC` 移除殖民地；另一條玩家可見分支亦串接
   `sub_E1CED`、`sub_EC97C`、`sub_E2A70` 清除殖民地與平行資料。原版不是把人口留在 0 的
   空殖民地，也不是直接扣受害帝國 BC。
8. raw owner/type 8 的安塔蘭殘艦在 `0xE9304` 另有回收／刪船鏈；事件怪物 10..14 不走該回收
   分支，存活者會保留在星系，供後續 `sub_DB8D8` 再選目標。

## 近似與未知

- **已證實藍圖：**五個 loader 的艦體、武器槽、裝甲、結構、戰速與低位 HV／PD 已於後續
  `event-monster-blueprints-audit-20260825.md`、`event-monster-weapon-runtime-audit-20260825.md`
  閉合並形成 typed blueprint。後續已將 Dragon `0x4000` 閉合為 OVR，怪物專用特殊武器
  亦另有逐效果稽核；本句原先的未知狀態已由後續證據推翻。
- **近似：**remake 的固定防禦 combatant 已接軌道基地、飛彈基地、地面砲台、戰機基地及行星
  恆星轉換器，但其戰機傷害與部分艦體配置仍是既有明示近似。不得把此切片宣稱成精確戰術 parity。
- **未知：**`sub_4267B` 的 type 10 `20/0` 特殊欄位在戰術／戰略 core 的完整下游效果，以及
  `sub_DB6D2` 戰後下一目標評分表，仍須獨立切片。

## 推翻的舊結論

- 「怪物抵達後只盤據、等待玩家主動攻擊」：錯；除太空鰻外會進一般戰鬥並轟炸殖民地。
- 「事件怪物可直接用固定人口傷害」：錯；原版固定三外圈後 `/40`，再走共用隨機候選池。
- 「人口歸零後保留空殖民地」：錯；戰後鏈有明確殖民地移除 consumer。
