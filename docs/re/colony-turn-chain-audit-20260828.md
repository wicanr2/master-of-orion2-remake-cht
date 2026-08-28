# 殖民地回合套用與兩遍重算稽核（2026-08-28）

## 證據契約

- 輸入：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- 工具：IDA Pro 9.4、`tools/ida/audit_colony_turn_chain.py`；位址均為 IDA linear，
  DOS/4GW LE object #1。
- 外部符號導航：`symbols_fixed.tsv`，SHA-256
  `f83116049964c14072547dd087a6316117cf62808c328cc7c087984467030b28`。
  名稱只供導航，語意以指令、caller、欄位與 consumer 分級。
- 可重生證據：
  [`evidence/colony-turn-chain-ida-20260828.json`](evidence/colony-turn-chain-ida-20260828.json)。
- `memset_`、`memmove_`、`Assert_Settings_`／`nullsub_14` 等 C runtime／build helper 只保留
  callsite，不納入 remake 或玩法完成分母。

## 結論：套用一次，衍生狀態重算兩次

`Next_Turn_Calc_ @ 0x136B3` 的殖民地相關外層順序為：

```text
0x13742 Apply_All_Player_Changes_
0x13747 Apply_All_Colony_Changes_       ← 玩家可見變更只在此套用一次
...
0x13765 Event_Twiddle_
0x1376A Lose_Out_Of_Range_Ships_
0x1376F Do_Colony_Calculations_         ← 第一遍衍生狀態重算
0x13774 Compute_Blockades_
0x13779 Move_Settlers_
0x1377E Do_Colony_Calculations_         ← 吸收封鎖／人口遷移後再重算
```

因此兩次 `Do_Colony_Calculations_` 不是兩次人口成長、兩次建造或兩次收入。真正的殖民地狀態
套用發生在較早的 `Apply_All_Colony_Changes_`；後段兩遍是同一份 derived cache 的重新計算。

## `Apply_All_Colony_Changes_` 的一次性套用鏈

`sub_E3FDC @ 0xE3FDC..0xE401D` 正向掃描所有 361-byte 殖民地 record，只處理
`owner != -1 && colony+0x06 == 0`，逐座呼叫 `sub_E3F6E @ 0xE3F6E..0xE3FDC`。

`sub_E3F6E` 的原始順序為：

1. 以殖民地索引呼叫 `Event_Check_Space_Anomaly_ @ 0x2341E`；命中即跳過整座本回合套用。
2. `colony+0x12F == 4` 時改成 3；這是既有同化狀態轉換欄，精確名稱沿用同化專題證據。
3. `colony+0xDC < 255` 時加一。
4. 若 `colony+0x0A > 0`，依序呼叫：
   `Apply_Colony_Pop_Growth_ @ 0xE2DCA` → `Apply_Assimilation_ @ 0xE3456` →
   `Produce_Ground_Military_ @ 0xE3616` → `Apply_Production_ @ 0xE36DF`。
5. 人口為零則不走上述四項，只呼叫 `Ghost_Colony_Stuff_ @ 0xE3E9A`。

這一段證實人口成長早於同化，同化早於陸軍與建造；它們不會因後段 `E2B31` 呼叫兩次而重複。
各 helper 內部公式仍由人口、同化、地面部隊與建造專題分別判定，不因外層順序閉合而一併升格。

## `Do_Colony_Calculations_ @ 0xE2B31` 的六階段

函式範圍為 `0xE2B31..0xE2D09`；除開頭無效果的 settings/build helper 外，玩法順序固定：

1. **逐殖民地 pre-import 重算**（`0xE2B50..0xE2B71`）：反向掃描 active 殖民地，呼叫
   `Pre_Import_Computing_ @ 0xE1D59`。
2. **逐帝國分配 imports**（`0xE2B73..0xE2B89`）：player slot 由 0 遞增，呼叫
   `Pass_Out_Imports_ @ 0xDF8F0(player, 0)`。
3. **逐殖民地人口預測與專業**（`0xE2B8B..0xE2BCA`）：再次反向掃描 active 殖民地，依序呼叫
   `Colony_Pop_Grows_ @ 0xE1839`、`Colony_Specialty_ @ 0xE1E1F`。
4. **逐帝國彙總**（`0xE2BCC..0xE2BEE`）：player slot 由 0 遞增，呼叫
   `Update_Player_Stats_ @ 0xE2710`。
5. **重建三個 player raw cache**（`0xE2BF0..0xE2CFD`）：
   - `player+0x5E8`：清零後計數 83 個 topic 中 status `3` 的項目；
   - `player+0x5EA`：清零後掃 active ship record，owner `<8`、type `<5`，以 `ship+0x10`
     作 bit index 累加；原版使用 `add` 而不是 `or`，精確高層語意仍未知；
   - `player+0x60C`：清零後計數 owner 的 active 殖民地中 `colony+0x0A >= 5` 的數量。
6. 返回；函式本身不呼叫人口套用、同化、地面部隊或建造套用 helper。

`Colony_Specialty_ @ 0xE1E1F` 另可精確描述：讀 `colony+0xE7/+0xE9/+0xEB` 三個 signed word，
保存最大值及其 0..2 索引；若最大值嚴格大於三值總和除以二（Watcom signed `/2`），把索引寫入
`colony+0x0B`，否則寫 `-1`。名稱只供導航，公式與欄位為已證實。

## pre-import 與帝國彙總邊界

`Pre_Import_Computing_ @ 0xE1D59` 依序呼叫 14 個玩家可見 producer：

```text
食物／農夫 → 環境 → 工業／工人 → 研究／科學家 → 士氣
→ 工業維護 → 工業產出 → 食物產出 → 食物維護 → 食物複製機
→ 研究產出 → 工業轉稅 → BC 產出 → BC 維護
```

開頭的 `memset_` 只清 derived 區，不另算玩法。`Pass_Out_Imports_ @ 0xDF8F0` 在帝國尺度分配
食物／貨運 imports，內部會在分配改變後重跑 `Colony_Pop_Grows_` 與 `Colony_Specialty_`；
這些是快取重算，不是套用人口。`Update_Player_Stats_ @ 0xE2710` 再聚合人口、食物、工業、研究、
BC、維護、納貢、領袖及事件分流；其已閉合欄位沿用既有專題文件，尚未逐欄證實者不因本輪升格。

## 其他 caller 與停止線

`E2B31` 另有兩個 callsite `0xE587C`、`0xE5ADD`，都位於
`Twiddle_Initial_Homeworlds_ @ 0xE5832..0xE5AE7`：第一次在建立母星與更新艦艇航程後，第二次在
建立開局艦後。這交叉支持它是可重入的 derived-state rebuild，而不是「每呼叫一次就過一回合」。

封鎖與殖民者移動本體仍是獨立玩法：`Compute_Blockades_` 的 `memset_` 只作暫存清理，
`Move_Settlers_` 的 `memset_／memmove_` 只作記錄操作；它們的玩家可見選擇、亂數與回寫不可因
runtime helper 被排除。第二遍 `E2B31` 已證實消費它們改變後的 colony／player records。

## 完成邊界

- **已證實**：一次性套用鏈、兩遍 derived rebuild、六階段順序、active record gates、三個 raw
  cache producer、殖民地專業公式、開局兩個額外 caller。
- **未一併閉合**：14 個 pre-import producer 與 `Update_Player_Stats_` 的每一條公式、封鎖選擇、
  殖民者遷移規則，以及 `+0x5EA` bit 累加的高層 consumer 語意。它們應按玩家影響另開窄切片。
- **remake 判定**：目前 `RunEmpireTurn`／shell helpers 可玩，但尚未依此精確 phase ordering 做
  同狀態對照；RE-first gate 下只記錄證據，不在本輪改 Go。
