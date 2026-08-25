# 隨機事件 18「秘密實驗」規格

證據來源：[`../re/random-event-secret-experiment-audit-20260825.md`](../re/random-event-secret-experiment-audit-20260825.md)。

## 事件效果

1. 事件沿用一般好事件的帝國目標選擇，不另選殖民地、星系或艦隊。
2. 先保存目標帝國目前的 `ResearchTopic`，再立即完成該主題，不檢查成本、研究進度或
   一般突破骰。
3. 一般主題授予目前已選的 application；`ResearchAll` 與 Creative 授予全部；
   Uncreative 沿用研究開始時已固定的 application；Hyper-Advanced 主題增加一級。
4. 完成後把 `ResearchProgress` 與 `ResearchTopic` 清為 0，並清除暫存 application／
   pending-choice 狀態。事件本身不得自動選下一個研究主題。
5. 執行既有科技授予 callback 與玩家／AI 自動艦型更新，效果必須可由存檔自然往返。
6. 若 `ResearchTopic==0`，事件仍可顯示並清理研究暫存狀態，但不授予科技或 RP。

## 訊息與驗收

- 繁中與英文訊息都使用完成前保存的主題名稱；無有效主題時明說實驗未產生可用科技。
- 抽樣測試至少覆蓋：玩家已選 application、Creative 全解、AI、熱座非目前席位、
  Hyper-Advanced、無主題，以及研究進度／主題清零。
- 測試應明確防止舊 `80 + Turn` RP 公式回歸。
