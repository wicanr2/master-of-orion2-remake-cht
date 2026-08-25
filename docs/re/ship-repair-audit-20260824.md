# 艦艇靠港修復靜態稽核（2026-08-24）

## 證據契約

- 原版輸入：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- IDA 資料庫：`Orion2.exe.i64`，分析前 SHA-256
  `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`；探針只操作暫存副本。
- 工具：IDA Pro 9.4／IDAPython，SDK 940；本文位址均為 IDA linear address。
- 探針：`tools/ida/audit_ship_repair.py`。匯出 raw name、位址、bytes、operand、caller、
  direct callee 與反編譯導覽文字，不改名、不套型別、不寫回原資料庫。

## 回合呼叫與逐艦篩選

**已證實**：`sub_580F5 @ 0x580F5` 的唯一直接 caller 是
`sub_136B3 @ 0x136B3`（`Next_Turn_Calc_`）內的 `0x13797`。它先建立八位玩家的
drive 對照，接著由 `word_199994` 控制，逐筆掃描 `dword_197F9C` 的 0x81-byte Ship record。

已驗證 `.GAM` Ship layout 可把篩選偏移對回：

- `+0x11`：`Design.Type`。
- `+0x63`：`Owner`。
- `+0x64`：`Status`。
- `+0x65`：`Star`。

對一般玩家船（`Owner <= 8`），`0x58199..0x581D6` 必須同時通過：

1. `Star < 72`。
2. `Status == 0`。
3. `Star > -1`。
4. `Design.Type == 0`，即 `COMBAT_SHIP`。
5. `sub_1276F0(starRecord + 0x38, Owner) != 0`。

`sub_1276F0 @ 0x1276F0` 是純 bit test：讀取 `base[owner>>3]` 的 `owner&7` 位元。
這個 star `+0x38` 玩家位元被靠港修復 caller 當成「該玩家在此星有可修復據點」消費，
證據等級為**已證實的消費語意**；其所有寫入端及殖民地／前哨站各自如何設定位元尚未在
本輪完整追查。因此「殖民地一定設定」由函式用途與玩家可見行為支撐為**強推論**；
「前哨站也設定」目前只有手冊補給站描述與 remake 模型支撐，仍是**強推論**，不能寫成
這份 IDA 已直接證實。

`Owner > 8` 會從 `0x58199` 直接跳到修復呼叫，略過一般玩家船篩選。這是特殊／非玩家
record 的原版處理，已證實控制流；remake 沒有同尺度的怪物／安塔蘭逐艦持久損傷，故本輪
不為它虛構對映。

## 完整修復寫回

**已證實**：通過條件後，`0x581DB` 呼叫 `sub_581F3 @ 0x581F3`。後者對同一
0x81-byte Ship record 執行：

- `+0x6E` ShieldDamage = 0。
- `+0x6F` DriveDamage = 0。
- `+0x70` ComputerDamage = 0。
- `+0x7B` ArmorDamage = 0。
- `+0x7D` StructureDamage = 0。
- 五個 `+0x76..+0x7A` DamagedSpecials bytes 清零。
- 八個武器槽把 `WorkingCount (+0x1F + 8*i)` 恢復成 `MaxCount (+0x1E + 8*i)`。

所以原版靠港行為確實是**全系統完整修復**，不是每回合按比例修復。remake 的 `Ship.Damage`
只保留裝甲／結構合併後的單一損傷量；把它清零是目前資料模型能做到的玩家可見對映，但不代表
逐武器、特殊裝置、護盾、引擎與電腦損傷已完成。

## 舊斷言勘誤與 remake 邊界

- 舊 `repair.go` 說明沒有記錄 `Design.Type == COMBAT_SHIP` 門檻，導致 remake 也會修復
  殖民船／前哨船；這與 `0x581B5..0x581B9` 直接矛盾，本輪修正。
- `ETA > 0` 對映原版 `Status != 0` 是**強推論**：兩者都表示不處於可停靠的靜止狀態，
  但 remake 只有 ETA 航行模型，沒有 raw Status enum。
- 前哨站仍暫列玩家基地，證據等級降為強推論；若日後追出 star `+0x38` 寫入端不含前哨站，
  必須移除，不用手冊「補給站」文字硬撐修復語意。
