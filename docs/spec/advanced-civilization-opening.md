# Advanced Civilization 開局規格

狀態：CONFORMED

RE-TRACE: dos-orion2-1.31:0x62C70

RE-TRACE: dos-orion2-1.31:0xE5832

## 範圍

本規格只定義 Advanced Civilization 的額外殖民地全圖分配、平衡與 Money 初始國庫。
開局科技發放、共同母星修改、原始亂數演算法及沒有玩家可見效果的 starting-ships 空迴圈
不在本規格內。

## 原版契約

1. 每名玩家額外行星額度為
   `trunc((trunc(starCount/2)*10)/playerCount)`，再做 `trunc((value+9)/10)-1`。
2. 候選只取距離不超過 9、非 raw star type 9、有效 raw planet type 3、無 owner 且未被
   其他候選占用的行星；每星最多五顆、每玩家最多 360 筆。
3. worth 使用未殖民行星 base、proximity bonus 與母星星系 67% 加項，按 final worth 降冪。
4. 玩家順序先隨機化，再 round-robin 選取；距離門檻為
   `max(difficulty+9,trunc(maxDistance/10))`，每玩家後續平衡表最多 20 筆。
5. 非最佳玩家逐顆最多六次提升 raw `+0x05／+0x08／+0x0A`，直到平均至少為最佳玩家 90%。
6. raw special 4／5／10 另做最多 100 次再分配；不得覆寫 Artifacts Homeworld 母星。
7. 最終依 owner table 初始化全部殖民地。
8. Advanced Civilization 的初始國庫為 `(MoneyRaw+2)*100`；標準 raw `-1／0／+1` 對應
   `100／200／300 BC`。

## typed／存檔與失敗行為

- 分配器輸入必須明示 star／planet records、玩家母星、難度、Money raw 與可重播亂數源。
- 資產或資料索引越界失敗即關閉，不產生部分 owner table。
- 結果必須經一般新遊戲入口建立 colonies，並由既有 JSON／多人 snapshot 往返；不得只留
  direct-entry 或測試專用注入。

## 驗收

- 額度奇偶、候選距離 9／10、owner 衝突、20／360 上限及 90% 門檻邊界測試。
- 固定 RNG 驗證 round-robin、每顆六次上限與 special 100 次停止條件。
- Money 三個標準 raw 的初始國庫測試。
- 正常新遊戲以 Advanced Civilization 啟動，存檔重載後殖民地 owner、planet 屬性與國庫一致。

原版證據見
[`../re/advanced-civilization-planets-audit-20260830.md`](../re/advanced-civilization-planets-audit-20260830.md)
與
[`../re/ai-trait-profile-tech-homeworld-audit-20260830.md`](../re/ai-trait-profile-tech-homeworld-audit-20260830.md)。
