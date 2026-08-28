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
5. `sub_FDA3F @ 0xFDA3F..0xFDAA7` 反掃來源星系五軌道，找到 owner 相同且
   `colony+0x141 != 0` 的殖民地便回傳；`sub_145EA(source,11)` 清除的正是
   `colony+0x136+11`。`sub_E65F8` 在 Colony Base 與 Colony Ship 同時存在時先走 base 分支。
6. `sub_D10EE` 不經 raw 11 的一般零分 scorer：在 expansion gate 成立、候選目標與來源行星
   climate byte 相同、來源 `population/8 + industry - pollution >= 13` 且尚無 base 時，直接把
   product raw word 寫成 11。原版建築表成本為 200 PP。
7. `sub_5E55F @ 0x5E766` 在建立新殖民地時直接寫 `colony+0x141=1`；新殖民地因此持有新的
   Colony Base，不是只有地表畫面虛擬顯示。

## Remake 訂正與剩餘邊界

- 舊 `advanceAI` 每五回合呼叫 `aiExpand`，而 `aiExpand` 不查也不消耗任何殖民來源，會讓 AI
  無限免費擴張。精確 AI 職務器接線後，這會快速製造大量缺糧殖民地並把母星工業壓成零；
  它是舊擴張代理缺陷，不是職務公式應放寬的理由。
- 新局 AI 現保存一艘 Average 開局殖民船；`aiExpand` 每回合只檢查 `aiFleetStar` 所在星系，
  `FleetETA>0` 時停止，成功後移除該艦。不再有五回合全圖瞬移殖民。
- 同星系選點已拆開 `sub_D27A7` 基礎價值與跨星 contextual 值：反向五軌道掃描的玩家可表示
  契約是「從軌道 4 反掃至 0，未殖民且基礎值嚴格較高才取代」；同分保留較高軌道。
- **強推論近似：**目前只有一支 AI 主力艦隊；跨星 adapter 把含 Colony Ship 的整支艦隊送往
  contextual 分數最高的合法星系，保存真實 ETA，抵達後下一回合才由已證實的本地 consumer
  建殖民地。原版多艦隊 route planner 及多支殖民艦仍待窄切片。
- Colony Base typed 鏈現依上述 producer／consumer 建模：AI 強制產品成本 200 PP、完工保存
  source flag、同星系殖民優先消耗 base，且新殖民地重新取得 base。尚未閉合的是 `sub_D10EE`
  expansion gate 的所有上游帝國陣列與真人 UI 專屬建造操作；目前只採已證實的同氣候與工業門檻。

## 同輪開局資料勘誤

AI 母星先前先建立固定 Normal-G 殖民地，再由 `syncAIRaceEngineFields` 寫入種族原生重力；Low-G／
High-G AI 因而在自己的母星承受非原生重力懲罰。現由 `applyHomeworldRaceTraits` 同步母星與行星的
原生重力，並在第一回合職務排序前重建 owner population profile。這是玩家可見的開局產出修正；
不以軍力測試反推產出常數。
