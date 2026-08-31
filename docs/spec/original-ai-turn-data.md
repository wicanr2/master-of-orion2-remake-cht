# 原版 AI 回合資料規格

規格狀態：READY

RE-TRACE: dos-orion2-1.31:0xD3D34

逆向證據：[`compute-ai-data-consumer-census-20260830.md`](../re/compute-ai-data-consumer-census-20260830.md)

訂正標記：`CORRECTION-20260830-AI-CACHE-CONSUMERS`

## 實作契約

1. AI 回合依 `Next_Turn_Calc_` 已證實順序建立當回合衍生資料，再執行外交、殖民地決策、
   移動、殖民、研究、經濟套用與戰鬥；不可把 cache 當持久存檔格式。
2. typed 資料只須表示有玩家 consumer 的輸入：殖民地人口／產出、科技、外交方向狀態、
   星系距離與造訪、艦隊位置／目的地／ETA、指揮額度及帝國經濟。
3. 殖民地職務使用原版逐人口邊際排序與帝國層反覆重算；群組模型遺失的完全等價 qsort
   次序必須標成 deterministic reconstruction。
4. 常態研究重用 application 級估值與既有 Uncreative 單項狀態。remake 亂數流須可存檔、
   可重播並維持同一決策的消耗順序，但不冒稱原版全域 PRNG 位元相同。
5. 生產、殖民、運輸、領袖與艦隊 target 的詳細公式由各自 spec 管理；本規格只鎖定共用資料
   邊界與回合順序，不複製過期的「未表示 cache 欄位」待辦。

## 目前 source／test 狀態

- 已接：常態 application 選擇、原版殖民地職務與補農夫、typed 殖民／食物運輸／玩家目標
  producer，以及決定性可存檔亂數流。
- 已接：`originalAITurnData` 在每位 AI 的職務分配前建立、帝國結算後自然釋放；它不是
  `GameSession`／`AIOpponent`／`sessionSnapshot` 欄位。職務排序與最終經濟共用同一份難度、
  事件、封鎖與 `colony+0xDD` typed 輸入；職務回寫只合併持久欄位，暫態加成不跨回合累積。
  缺逐種族 profile 或未知食物欄位時仍走既有有界 fallback，且不改寫原版沒有 runtime writer
  的 AI 稅率。
- 部分接線：整體 AI 經濟、移動、建造、外交與戰鬥仍混合原版規則與明示近似；須依各子系統
  trace row 判定，不能由本規格 READY 升格成 CONFORMED。
