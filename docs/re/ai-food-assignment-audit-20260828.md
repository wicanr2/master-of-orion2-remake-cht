# AI 追加農夫與食物運輸靜態稽核

日期：2026-08-28

## 證據契約

- 輸入 `Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- IDA 資料庫為既有事件怪物研究 checkpoint，SHA-256
  `587e0665e6fa115b2262efcf3e02197a2dc1dd3319a983283ba8afd56ff62867`；它不是文件其他切片所用
  `4a0179...` 正式 checkpoint，因此名稱只供導覽，結論以 raw bytes／位址／operand 為準。
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

以上 1–5 為已證實。`sub_D68CB` 的 raw comparator 與 `sub_D682A` 的
`food-currentJobOutput` 主形狀亦已證實；packed colonist 額外 `+1000` 分支所讀欄位的完整
玩家語意、`player+0x36/+0x40` 的正式名稱，以及 `player+0x38` 全尺度仍未知。

## Remake 對映與停止線

`engine.ApplyOriginalAIJobs` 現在於工業／研究平衡後執行追加農夫 pass。它逐一試算將普通人口
改為農夫造成的半食物增益，只接受正增益候選，直到帝國 `TotalFoodHalf>=0`。由於 remake 尚未
保存 `player+0x38` 的 typed 運輸容量，另要求每個未封鎖殖民地不留在本地饑荒；這等價於
「沒有可證明的運輸容量時不假設食物能憑空運送」，屬強推論的失敗即關閉近似，不宣稱原版
精確殖民地選擇／freighter parity。

正常 200 回合測試確認此切片消除 AI 人口 8→1 死亡螺旋，且非 Creative AI 會完成科技並寫入
application 擇一。這是 remake 玩家路徑驗證，不把它升格為原版同 seed／同回合 oracle。
