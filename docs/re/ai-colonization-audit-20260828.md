# AI 殖民主鏈靜態稽核

日期：2026-08-28

## 證據契約

- 輸入 `Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- 正式 `Orion2.exe.i64`，SHA-256
  `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- 工具：IDA Pro 9.4，映像 `ida-pro-9.4-idapython:py312-v1`；位址基準為 DOS/4GW LE image
  的 IDA linear EA。
- 非破壞性匯出：
  [`evidence/ai-colonization-ida-20260828.json`](evidence/ai-colonization-ida-20260828.json)。
  正式資料庫未改名、未加型別、未寫回。

## 已證實

1. `All_AI_Colonize_` 的 raw 函式是 `sub_E67F6 @ 0xE67F6`。它每回合反向掃描所有存活 AI，
   對每位呼叫 `sub_E65F8 @ 0xE65F8..0xE67F6`；不存在固定五回合免費殖民分支。
2. `sub_E65F8` 反向掃描 113-byte 星系記錄，只處理 owner 等於目前 AI，且
   `sub_FDA3F()` 或 `sub_FDAA7(ai,1,star)` 至少一個非零的來源。前者的成功下游在建殖民地後
   呼叫 `sub_145EA(sourceColony,11)`，與 raw 11 Colony Base 的既有移除鏈一致；後者下游解除
   指派領袖並呼叫 `sub_A163A` 移除實際艦艇。原版因此需要一次性的 Colony Base 或 Colony Ship，
   不是只看回合數便產生殖民地。
3. 函式只從該星系五個行星槽中挑 `planet colony index == -1` 的未殖民行星，
   以 `sub_D27A7` 估值並取嚴格較高者；找不到時不建立殖民地。
4. `sub_E5EB3 @ 0xE5EB3..0xE6071` 建立新殖民地、寫入一名起始人口、處理原住民、重算殖民地，
   最後重算星系狀態。這與既有玩家／AI 共用 typed 建構器的玩家可見結果一致。

## Remake 訂正與剩餘邊界

- 舊 `advanceAI` 每五回合呼叫 `aiExpand`，而 `aiExpand` 不查也不消耗任何殖民來源，會讓 AI
  無限免費擴張。精確 AI 職務器接線後，這會快速製造大量缺糧殖民地並把母星工業壓成零；
  它是舊擴張代理缺陷，不是職務公式應放寬的理由。
- 新局 AI 現保存一艘 Average 開局殖民船；`aiExpand` 沒有實際殖民船時直接停止，成功後移除該艦，
  因此不會在下一個五回合週期再次免費殖民。
- **強推論近似：**目前只有一支 AI 主力艦隊位置，`aiExpand` 尚未把殖民船拆成獨立航線；目標仍沿用
  remake 的全圖行星價值選擇與五回合觸發。原版逐星來源、實際抵達時序、Colony Base 同星系來源及
  多支殖民艦仍待下一個窄切片，不能把本輪稱為完整 `All_AI_Colonize_` parity。

## 同輪開局資料勘誤

AI 母星先前先建立固定 Normal-G 殖民地，再由 `syncAIRaceEngineFields` 寫入種族原生重力；Low-G／
High-G AI 因而在自己的母星承受非原生重力懲罰。現由 `applyHomeworldRaceTraits` 同步母星與行星的
原生重力，並在第一回合職務排序前重建 owner population profile。這是玩家可見的開局產出修正；
不以軍力測試反推產出常數。
