# NPC 國力矩陣 `+0x5EC` 靜態稽核（2026-08-28）

## 證據身分

- 輸入：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- IDA 資料庫 SHA-256：
  `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`；IDA Pro 9.4
  只分析一次性副本，位址均為 IDA linear EA、DOS/4GW LE object #1。
- 原始函式、bytes、operand 與交叉參照匯出：
  [`evidence/npc-treaty-negotiations-ida-20260827.json`](evidence/npc-treaty-negotiations-ida-20260827.json)。

## 已證實：資料形狀與唯一直接 writer

`+0x5EC` 的直接站點共九處：四個讀取函式 `sub_27A3D／sub_4F59B／sub_500CF／sub_51DCE`，
以及唯一直接 writer `sub_D3D34 @ 0xD3D34..0xD574D`。後者函式 bytes SHA-256
`1574ebdb562d0547e3d83e9e44dc966e6ebfb3ff0b6186399424b812d7343aab`。

`0xD42EA..0xD4309` 對每個帝國記錄從 `+0x5EC` 起清零 32 bytes，因此不是單一國力值，而是
八個方向性 `dword`。`0xD4334..0xD43BB` 再掃描所有 stride `0x81` 艦艇記錄：owner byte `+0x63`
小於 8 時，對每個正常帝國槽把該艦在 `dword_1AA1DC` 的對應 `word` 累加到
`ownerRecord + 0x5EC + 4*observer`。

## 已證實：逐艦 producer

在同一個 `sub_D3D34` 前段，每艘存活／有效艦對各帝國槽呼叫 `sub_5EF17`，結果寫入每艦 16 bytes
的八個 `word`。相關函式身分：

- `sub_5EF17 @ 0x5EF17..0x5EF33`，bytes SHA-256
  `c3753e02ce04bc80748ca694758bc4b518457d0579fa72bae2c23aa9b8786185`；
- `sub_5EF4B @ 0x5EF4B..0x5F2C3`，bytes SHA-256
  `d624036078d35f84ae6a44408661db3a33754d0378278c1d2e3d860a32af7b8c`；
- `sub_5F871 @ 0x5F871..0x5F8BF`，bytes SHA-256
  `6d862c7137f34b8127a265537ae6bab90299e68553a9b6e28e653bf5d88e7da9`。

`sub_5EF4B` 掃每艦八個武器槽，讀武器 ID、數量、改造 flags、武器類別與基礎值；處理重／點防
改造、戰機類特殊換算、命中率下限／上限、船員／傷害修正及觀察者相關防禦值，最後形成方向性的
逐艦武器效能。`sub_5F871` 提供由帝國科技／特殊 owner 決定的扣減輸入。這些資料流證明
`+0x5EC` 不是艦體數量、建造成本或全銀河無條件可見的 `FleetStrength`。

2026-08-28 由同一 `.i64` 副本再匯出 `sub_5EE27` 十一個直接讀取的 signed word。依 raw bit
`0x0002／0004／0020／0080／0100／0200／0400／0800／1000／2000／4000`，倍率增量依序為
`+100／-50／+25／+50／+100／+100／+25／+25／+25／+300／+50`；乘完後逐槽封頂 64000。
原始位址、bytes 與值已加入同一 evidence JSON。這是國力評分倍率，不可拿造艦佔格成本表替代。

同輪進一步證實兩個 observer 輸入不可混用：`sub_5F871` 依 `sub_5679E` 的最佳電腦 0..5
查 `word_17F6C1`，扣減依序為 `0／1／3／5／7／10`；`sub_5EAE9` 則依 `sub_56726` 的最佳
引擎 0..7 查速度 `0／6／8／10／12／14／16／2`，乘 5 後才加 empire `+2213` 艦防與
`+2236` 跨維度的 20。純規則已拆成 `ObserverWeaponReduction` 與 `ObserverDefense` 兩欄。

`sub_5F2F6` 的一般艦分支會把 `sub_58387` 與 `sub_58425` 相加。IDA 已追回兩者共用的 hull
基礎表 `4／10／30／50／80／150` 與 armor raw 0..7 百分比
`-100／0／100／300／500／700／900／0`；`sub_582BF` 的 special raw 27 只把 structure 乘三，
`sub_58425` 的 raw 14 只把 armor 乘三。之後才扣 ship `+125` structure damage 與 `+123`
armor damage。remake 已依此建立獨立耐久 producer，未再借用 `shipMaxHP`。

`sub_54D80／sub_54E5B` 寫給 `sub_5EF4B` 的 attack 輸出也已閉合：computer 表
`0／25／50／75／100／125`，可用的 special raw 5 戰鬥掃描器加 50，再加逐艦 Weaponry 軍官、
crew offense `0／15／30／50／75` 與 owner `+2214` 種族艦攻；此路徑不讀 ComputerDamage。

## 已證實：消費公式

`sub_500CF @ 0x500CF` 對指定方向使用 `+0x5EC`：

```text
ratio = min(800, 100 * (sourcePower + 1) / (targetPower + 1))
ratio /= 2 for each source third-party war
```

該比例直接供 NPC 納貢與宣戰決策。公式已於 remake 接入；尚未對齊的是 producer 的逐艦方向值。

## 結論與停止條件

- **已證實**：資料形狀、唯一直接 writer、逐艦逐觀察者累加、主要武器槽輸入及 ratio consumer。
- **已完成資料鏈**：原版 `.GAM` 的八槽武器、computer、size、armor、shield、base combat speed、
  五個 damaged-special bytes、分離損傷與 crew level 已映射到 typed `Ship`，並通過 JSON 往返測試。
- **已完成純規則**：`OriginalNPCShipPower` 已接八槽限制、戰機／突擊艇／脈衝星轉換、HV／PD、
  觀察者扣減、光束命中 10..100、飛彈／魚雷 75%、炸彈歸零、十一個 raw modifier、電腦落後與
  `sub_5EED4` 耐久壓縮；非法 raw 輸入失敗即關閉。
- **已完成新造艦輸入切片**：可回查 `OrigWeaponByTech` 的武器會寫 raw ID；已有 consumer 證據的
  HV／PD／CO／NR／AF／ARM／FST／OVR 寫 raw bits，其他未閉合改造讓 mount 維持不可供國力
  producer 消費。交付 AI 艦同時保存 computer、size、armor、shield 與 base combat speed。
- **已完成一般 AI producer**：`originalAIPowerMatrix` 每次由實艦重建 owner×observer 方向矩陣，
  不作權威存檔；納貢與一般宣戰共同消費該矩陣。觀察者科技造成方向差、八槽、受損與零艦隊
  均有純規則／shell 測試。
- **相容邊界**：只有非零 `FleetStrength`、卻完全沒有實艦 raw 資料的舊存檔或精簡測試才回退
  hull-only 純量，並明標 `exact=false`。特殊 owner／要塞分支不在一般 AI 艦隊本切片。
