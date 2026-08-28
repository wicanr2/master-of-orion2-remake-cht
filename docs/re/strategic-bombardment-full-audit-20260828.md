# 戰略轟炸完整快速戰鬥與防禦鏈（2026-08-28）

## 結論

原版 1.31 的戰略轟炸鏈已從停泊艦搜尋、33-byte 快速 combatant 建立、固定艦體攻擊次數、
最佳武器欄、三外圈攻擊、命中／減傷／耐久、結果 record，到殖民地傷亡回寫閉合。

本輪同時訂正兩個舊假設：

1. 原版不按設計槽的實際武器門數逐門轟炸；它以 hull／record type 的固定 beam、missile、special
   attack count，搭配該艦已安裝系統中各類最佳 raw weapon。
2. `sub_DCEBD` 排除的八棟建築不是八個獨立戰鬥者。Capitol 不進任何轟炸傷亡候選；三層軌道基地
   是互斥的額外 33-byte record；Missile Base、Ground Batteries、Stellar Converter、Fighter
   Garrison 則是行星 record 內的武器／旗標。戰略轟炸本身不逐棟移除這七種防禦。

本輪只補 RE，未修改 Go。

## 證據身分

- 輸入：`Orion2.exe`
- 輸入 SHA-256：`7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- IDA 資料庫 SHA-256：`4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`
- 外部符號表 SHA-256：`f83116049964c14072547dd087a6316117cf62808c328cc7c087984467030b28`
- 工具：IDA Pro 9.4／IDAPython；DOS/4GW LE object #1 的 IDA 線性位址
- 匯出：[`strategic-bombardment-full-ida-20260828.json`](evidence/strategic-bombardment-full-ida-20260828.json)
- 腳本：[`audit_strategic_bombardment_full.py`](../../tools/ida/audit_strategic_bombardment_full.py)

外部名稱只供導覽；以下結論以 raw 位址、指令、資料表及交叉參照為準。

## 33-byte 快速 combatant

`sub_416CF @ 0x416CF` 建立 stride `0x21` 的快速 combatant。第一次由停泊艦索引建立攻方，
第二次 `DX=0` 時從 `word_199878` 指定 colony 建立 index 0 行星 record。主要欄位為：

| offset | 已證實用途 |
|---:|---|
| `+0x00` | owner／side raw byte |
| `+0x01` | record type；一般艦為 hull class，行星為 4，三層軌道基地為 6／7／8 |
| `+0x06` | destroyed／inactive state |
| `+0x09` | armor／damage-reduction raw type |
| `+0x0B` | 最佳 beam raw ID |
| `+0x0D` | 最佳 missile raw ID |
| `+0x0F` | 最佳 special-damage raw ID |
| `+0x13` | offense／damage bonus subtotal |
| `+0x15` | defense subtotal |
| `+0x17` | missile modifier subtotal |
| `+0x19` | missile damage percent modifier |
| `+0x1D` | maximum hits |
| `+0x1F` | accumulated damage |

一般艦的 installed-system bitsets會掃 39 種 raw item，依 category 1..7 各保留最高數值；因此
同類多門武器不會形成多個 attack slot。這是抽象快速模式，不是格子戰術的八武器槽模型。

## 固定 attack-count 表

`sub_57651 @ 0x57651` 直接回傳 `word_180030[18*recordType]`；missile 與 special consumer
分別讀同一 36-byte type record 的 `word_18002E` 與 `word_180034`。原始表如下：

| record type | 對應 | missile attacks | beam attacks | special attacks |
|---:|---|---:|---:|---:|
| 0 | hull 0 | 1 | 0 | 0 |
| 1 | hull 1 | 2 | 0 | 0 |
| 2 | hull 2 | 2 | 1 | 2 |
| 3 | hull 3 | 4 | 2 | 5 |
| 4 | hull 4／行星 record | 6 | 4 | 10 |
| 5 | hull 5 | 10 | 10 | 25 |
| 6 | Star Base record | 3 | 3 | 0 |
| 7 | Battlestation record | 6 | 6 | 0 |
| 8 | Star Fortress record | 10 | 10 | 0 |

因此舊待辦所稱「逐武器數量」的答案是：戰略模式沒有逐設計槽數量；攻擊次數由 record type
固定表決定。patch 1.50 的 bomb 5／10 當量是另一個版本規則，不能改寫這張 1.31 runtime 表或
三外圈次數。

## 三條攻擊鏈

### Beam：`sub_41F80 @ 0x41F80`

以 attacker type 取得固定 beam count，只有最佳 beam ID 與 count 都大於零才攻擊。命中基底使用
attacker offense 與 target defense；Antaran／科技分支可減 70／40／20。每次擲 `Random(100)`，
96..100 被改成 1000；通過後呼叫 `sub_40C2A`。目標 destroyed 時由 `sub_41E88` 從另一 side 的
未摧毀 record reservoir 選新目標。

### Missile：`sub_4221F @ 0x4221F`

固定 missile count 由 type 表取得，命中門檻為 `min(95, 40-(attacker missile defense-
target defense))` 的 raw 有號式；最佳 missile ID 決定 min/max damage，`+0x19` 再作百分比修正。
每發獨立擲骰、傷害插值、套用 `sub_40C2A`，目標摧毀後重選。

### Special：`sub_420C0 @ 0x420C0`

固定 special count 由 type 表取得；最佳 special ID 決定 min/max damage。每次直接做區間插值並
打向 record 0，不另作命中骰。

`sub_40C2A @ 0x40C2A` 對 raw armor type 7 使用 signed `/4`，其餘扣
`word_17F6C1[59*armorType]`；傷害下限 0，累積上限 30000，達 maximum hits 即設 destroyed。

## 行星與防禦建築如何進 record

行星本體固定是 type 4。`sub_42371 @ 0x42371` 以

`40 × (population + marines + tanks + count(non-orbital buildings except 8/40/41))`

建立 maximum hits。`sub_416CF` 再把下列防禦接到行星 record：

| raw building | record 效果 |
|---:|---|
| 23／24／28 | 分別把 defense subtotal 乘 30／20／10，並加入對應 flat 值 |
| 26 Missile Base | 設 defense mask bit 0，填 missile raw ID |
| 27 Ground Batteries | 設 bit 1，填 beam raw ID |
| 42 Stellar Converter | 設 bit 2，啟用 special 類防禦 |
| 47 Fighter Garrison | 設 bit 3，填 fighter／special raw 欄 |

若上述 mask 為零，行星 record state 設為 2；非零才保持 active。三層軌道基地不疊加：

- raw 40 Star Base → type 6；
- 否則 raw 8 Battlestation → type 7；
- 否則 raw 41 Star Fortress → type 8。

原始分支只會追加其中一個 record。這也說明 Capitol raw 9 只是傷亡排除項，沒有 combatant。

## 三外圈、結果 record 與持久回寫

`sub_4257E @ 0x4257E` 固定三個外圈，對 record 1..N−1 依序呼叫 beam、missile、special，
record 0 累積傷害超過 30000 時提前離開；最後以 signed `/40` 回傳。`sub_4267B @ 0x4267B`
把回傳值寫到 73-byte 結果 record `word +0x03`，並將 `word +0x06` 清零。

`sub_DD2F2 @ 0xDD2F2` 因 `+0x06=0`，只會反覆呼叫 `sub_DCEBD @ 0xDCEBD`。後者的候選池排除
`{8,9,26,27,40,41,42,47}`，所以正常戰略轟炸不逐棟移除 Capitol 或七種防禦建築；它只從
一般建築、marines、tanks、build progress 與 population 分配 `/40` 後的傷害點。當 colony 最後
人口被移除時，較高層 colony destruction 才會清理整座殖民地。

`sub_DD13E @ 0xDD13E` 只在結果 `+0x06` 非零的另一種特殊傷害池使用；`sub_4267B` 的戰略分支
明確把該欄清零，不能拿它補造七種防禦的摧毀結果。

格子戰術是另一條鏈：`sub_3A3C3 @ 0x3A3C3 → sub_3A19E @ 0x3A19E` 可從 raw
26／27／42／47 中抽一棟移除；三層軌道基地則是獨立 combatant。這個 tactical writeback 不可
套進 strategic bombardment。

## 證據分級與 remake 邊界

- **已證實**：record layout 的上述 consumer、fixed attack counts、最佳武器策略、三攻擊鏈、
  damage clamp、行星 hits、七種防禦的 record 形狀、三層軌道基地互斥、結果欄與一般傷亡回寫。
- **強推論**：raw building 23／24／28 的正式英文名稱；本文只以 raw ID 描述其數值效果。
- **未知但不阻塞玩法 RE**：快速戰鬥動畫／音效時序、原始 PRNG 位元序列、1.50 bomb 當量在
  三外圈內的精確 patch 實作。1.31 executable 的玩家規則已閉合。
- **排除**：allocator、格式化、音效 driver、Windows／DOS 平台與編譯器 helper 內部行為。

remake 目前仍以實際武器槽與 bomb 5／10 當量計算，和原版 1.31 的 fixed record-type 模型不同；
依 RE-first gate，本輪只登記差異，待所有 RE 列關閉後建立 READY spec。
