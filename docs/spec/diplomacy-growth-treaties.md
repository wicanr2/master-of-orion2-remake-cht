# 每回合外交關係成長規格

**狀態：CONFORMED（玩家↔AI 可表示切片）**  
證據：[`../re/diplomacy-growth-treaty-audit-20260827.md`](../re/diplomacy-growth-treaty-audit-20260827.md)。

## 範圍

本規格實作 `Diplomacy_Growth_ @ 0x4DD6B` 的互不侵犯、同盟、貿易、
研究與納貢關係成長、它們可達的 `Change_Relations_` raw 公式，以及
`sub_4D78E` 建立的種族目標表與逐回合漂移。

不含 `+0x72F` 未知欄位、`sub_DBCC8` 領袖／實力事件、方向不對稱條約、
`+0x806/+0x80E` 計時器、完整 reason 快取／抱怨矩陣與原版 PRNG 位元序列。

## `Change_Relations_` 分數公式

1. 正 delta 且 current `<0` 時乘 2，但最多一步推到 `+10`；current `>=0`
   時除以 `current/25+1`。
2. 負 delta 且 current `>0` 時乘 2；current `<=0` 時除以
   `current/-25+1`。
3. actor government raw 4 遇負 delta 乘 2；raw 0 遇負 delta 乘 `3/2`，
   遇正 delta 乘 `3/4`，使用朝零整數除法。
4. target Charismatic 使正 delta 乘 2、負 delta 除 2。
5. 結果夾到 `-100..100`；非同盟正上限另夾到 65。

## 條約成長順序

每條關係邊依序處理：

1. policy 1：`roll(100) <= 100-current` 時套 `roll(3)`。
2. trade active：同一守門，通過時套 `roll(3)`。
3. research active：同一守門，通過時套 `roll(3)`。
4. policy 2：無百分比守門，直接套 `roll(5)`。
5. tribute mode 1／2：同一守門，通過時分別套 `roll(3)`／`roll(8)`。

每一步立即消費更新後的 current；無對應條約不擲骰。現行單邊模型只消費
玩家→AI 的 `PlayerTribute`，反向需待方向矩陣擴充。

## 種族目標與漂移

1. 目標取 `byte_180ED4[observerRace][targetRace]` 的 signed low byte。
   現行投影以 AI 種族為 observer、玩家種族為 target；自訂種族採中立 0
   的失敗即關閉投影。
2. 每條邊先擲 `roll(105)`；只有結果大於 `abs(current)` 才擲 `roll(4)`，
   等於 1 時再擲 `roll(2)-1` 得到 step 0／1。
3. 未鎖定時以 step 向 target 靠近且不越過。正式狀態 `>=4` 時，把高於
   `-90` 的 current 壓到 `-90`。
4. 一回合先處理全部條約邊，再處理全部目標漂移邊，不逐 AI 交錯。

## Remake 資料與玩家路徑

- `AIOpponent.Relation` 維持供 UI／AI 消費的 `-40..40` 投影。
- 存檔另保存 raw current、raw target 與各自 Known。舊存檔以能往返保存
  每個整數顯示刻度的向外取整建立 raw；外部玩法改寫顯示值時重建 raw。
- 使用可存檔的獨立確定性亂數流；這只保證 remake 重播，不宣稱原版 PRNG
  位元一致。
- 正式戰爭是權威狀態；不能因 `-90 raw` 投影為 `-36` 而降級成敵視。

## 驗收

- 純規則測試覆蓋條約順序、百分比邊界、政體、Charismatic、65 上限、
  14×14 表方向、漂移擲骰順序、鎖定與戰爭 -90。
- `EndTurn` 正常路徑集中執行條約成長與目標漂移；正式戰爭可沿正常回合
  路徑形成戰爭態勢與突襲。
- 存檔往返保留 raw current、raw target、Known 與亂數流位置。
