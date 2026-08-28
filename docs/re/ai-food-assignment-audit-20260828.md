# AI 追加農夫與食物運輸靜態稽核

日期：2026-08-28

## 證據契約

- 輸入 `Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- IDA 資料庫為正式 checkpoint，SHA-256
  `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- 工具：IDA Pro 9.4，映像 `ida-pro-9.4-idapython:py312-v1`；位址基準為 DOS/4GW LE image
  的 IDA linear EA。
- 非破壞性匯出：
  [`evidence/ai-food-assignment-ida-20260828.json`](evidence/ai-food-assignment-ida-20260828.json)。
  正式／暫存資料庫都未改名；語意另在本文件分級。

## 推翻的舊結論

先前 [`ai-colony-jobs-audit-20260828.md`](ai-colony-jobs-audit-20260828.md) 只追到
`sub_D61E7`、`sub_D652C` 與 `sub_D66B3`，便把職務鏈視為完成。200 回合正常新局反證顯示，
非 Creative AI 的母星與新殖民地會全部縮到人口 1、全為工人，沒有科學家。這不是研究選擇器
失效，而是漏掉職務平衡後的追加農夫 consumer。

## 已證實控制流

1. `sub_D6E1D @ 0xD6E1D..0xD6ED4` 先執行封鎖／未封鎖職務、帝國工業／研究平衡與建造選擇，
   再呼叫 `sub_D6D80`；有未封鎖殖民地時於 `0xD6EBD` 呼叫
   `sub_D6AD4 @ 0xD6AD4..0xD6D80`。因此追加農夫是同一玩家回合的正式下游，不是 UI helper。
2. `sub_D6AD4` 只收未封鎖候選；它要求 `colony+0xDD>0`、目前農夫低於
   `colony+0xE0`，以 `sub_D68CB @ 0xD68CB..0xD6A00` 排序，並以
   `sub_D682A @ 0xD682A..0xD68CB` 建立轉農夫的邊際分數。
3. `sub_D6AD4 @ 0xD6BB3` 呼叫 `sub_E2D09` 重算食物／運輸，接著同時檢查
   signed `player+0xB0` 與 `player+0x38`。兩者任一為負且仍有候選時，函式會選殖民地並呼叫
   `sub_D6A00 @ 0xD6A00..0xD6AD4`，然後回到重算迴圈。
4. `sub_D6A00` 清除候選人口的舊職務 bits、寫入農夫職務、增加殖民地 AI record 的農夫計數，
   呼叫 `sub_E1D59` 重算殖民地，必要時重新排序剩餘候選。這直接證明原版不會在
   `sub_D66B3` 把所有人口分成工人／科學家後就停止。
5. `sub_DF8F0 @ 0xDF8F0..0xDFDC6` 在 `0xDFA34..0xDFA39` 將帝國食物產出減需求寫入
   `player+0xB0`；在 `0xDFA8C..0xDFA90` 或 `0xDFD17..0xDFD4A` 寫
   `player+0x3E/+0x38`。它會按殖民地缺糧與可運輸條件調整 `colony+0xF3`，因此
   `+0x38` 是食物運輸壓力／餘額，不是第二份食物總量。

6. `sub_DF8F0` 證實 `player+0x36` 為貨運艦總數、`+0x40` 為在途殖民者數、`+0x3E` 為
   已運食物、`+0x38` 為帶正負號的貨運艦餘額；可供運糧容量是
   `TotalFreighters-5*SettlersFreighted`。未封鎖殖民地以 `colony+0xE7/+0xEF/+0xF3`
   表示食物產出、需求與輸送量。
7. `sub_E2000 @ 0xE23A6..0xE23C2` 證實維護費使用量：`+0x38<=0` 時取總數，
   `+0x38>0` 時取總數減餘額，最後無條件捨去除以二。
8. `sub_D6AD4 @ 0xD6D57..0xD6D76` 只在運輸壓力成立時呼叫 `Random(10)`；
   `Random(10)<=difficulty` 便直接對 `player+0x36` 加 5，不經一般建造佇列。
9. `sub_D5795 @ 0xD5795..0xD58AD` 先把每筆 13-byte AI 職務記錄清零；
   `sub_D652C @ 0xD660C..0xD660F` 在逐殖民地前段職務完成後，把 `colony+0x08`
   的低 byte 保存至 `AI record+7`。既有清污成本稽核已由 `sub_DEE1B` producer 證實
   `colony+0x08` 是本回合污染清除成本，不是人口或未消費工業。
10. `sub_D682A @ 0xD682A..0xD68CB` 對目前工人計算 `food-worker`，若後續
    `sub_D66B3` 帝國職務平衡重算後的 signed `colony+0x08` 高於保存的低 byte，再加 1000；
    非工人候選計算 `food-research`。`sub_D68CB @ 0xD68CB..0xD6A00` 依此分數降冪，
    同分依 food、raw job、race slot 升冪；跨殖民地只比較各自首候選分數，同分保留較高
    colony index。
11. 本輪補匯出先前只留交叉參照、未留函式本體的 `sub_D66B3 @ 0xD66B3..0xD682A`。
    `0xD6787..0xD679D` 明確讀 signed `player+0xAC`（研究）乘第二權重、讀
    `player+0xAA`（工業）乘第一權重，前者低於後者時選最大 `AI record+8` 改科學家，
    否則選最小 `+0x0A` 改工人。基值 10／18、late-tech 與 personality 分支均與既有
    producer 稽核一致。

以上 1–11 為已證實。早先把 `AI record+7` 暫稱未知玩家欄位的結論已由 producer 與既有
`colony+0x08` 寫入鏈推翻；其低 byte 截斷是原始指令行為，remake 不改成完整整數快照。

原版 `sub_D63A6／sub_D6315` 對同種族、同三職產出的人口回傳相等；最終次序依逐人口
4-byte 陣列原順序與 Watcom `qsort` partition。remake 的 `PopulationGroup` 只保留職務計數，
無法重建人口身分順序；依專案「編譯器 helper 不深挖」停止線，這個等價類別內次序列為
**未知／非阻塞**，不得宣稱逐位元相同。

## Remake 對映與停止線

`engine.OriginalFoodTransport` 現按上述欄位重算跨殖民地運糧與帶正負號餘額；
`ApplyOriginalAIJobsWithTransport` 在帝國缺糧或運輸受限時逐一試算增加農夫，並於運輸容量耗盡時
只保留本地仍缺糧的殖民地候選。`RunEmpireTurn` 保存兩個衍生欄位並依實際使用艦數收維護費；
shell 的精確 AI 職務 consumer 只在壓力成立時擲骰，成功便增加 5 艘。`.GAM` importer 亦保存
四個原版欄位。追加農夫現也保存前段清污成本低 byte，使用已證實的
`food-currentJobOutput`、`+1000` 與可表示的 tie-break。等價 `qsort` 類別採確定性重建：
優先維持既有工人，若群組折疊仍把全帝國研究壓成 0，保留一名邊際損失最小的科學家。
這兩項是 **remake approximation**，只補遺失的人口排列資訊，不冒稱原版 raw tie-break。

正常 200 回合測試確認此切片消除 AI 人口 8→1 死亡螺旋，且非 Creative AI 會完成科技並寫入
application 擇一。這是 remake 玩家路徑驗證，不把它升格為原版同 seed／同回合 oracle。
