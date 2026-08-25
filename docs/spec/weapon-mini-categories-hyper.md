# 武器微型化分類與 Hyper 重複等級規格

1. 成本微型化維持共同階梯 `100/75/55/40/30/25%`。
2. 佔格微型化必須依原版 typed weapon category 選階梯：一般 `100/80/65/50/35/25%`、
   魚雷／特殊武器 `100/80/70/60/50/40%`、機庫 `100%`。
3. 特殊裝置使用自己的解鎖科技計算等級；一般裝置走第二階梯，增強引擎、強化船體、
   延伸油箱、重裝甲、部隊艙固定不縮小。成本仍走共同階梯。
4. `PlayerState.HyperAdvancedLevels` 保存八個 Hyper topic 的重複完成次數。舊 JSON 缺欄時，
   已完成 Hyper topic 視為一級，未完成視為零級。
5. 一般科技的微型化等級沿後繼主題計數；走到 Hyper topic 時直接加該領域重複等級，
   不可再把 `CompletedTopics` 的布林值重複加一次。
6. 完成 Hyper 研究時增加該 topic 等級並保留 `CompletedTopics=true`；研究選單在領域走到底後
   仍提供該 Hyper topic。成本為版本 profile 基礎值加已完成等級 `×10000`，詳見
   [`hyper-repeated-cost.md`](hyper-repeated-cost.md)。
7. 造艦 UI、是否超格、實際扣款、命令重播與改裝成本必須共用同一套分類與 Hyper 等級。
