# 玩家玩法 RE 剩餘盤點與垂直追溯（2026-08-30）

## 盤點口徑

本盤點只計會改變原版玩家可見玩法、且尚缺輸入／規則／consumer 垂直證據的 RE。以下不計入：

- compiler helper、C runtime、Windows API 與 DOS 硬體內部；
- 已閉合 RE、但 remake 尚未實作或尚未 CONFORMED 的項目；
- raw 名稱、逐字文案、sprite／逐像素與 1.50 profile 差異等非 1.31 gameplay gate；
- 沒有玩家 consumer 的欄位與只剩資料模型近似的實作工作。

依此口徑，41 列 parity matrix 並不是「41 項 RE 未完成」。目前剩 **7 個主要 RE 群組**；每組
仍需拆成數個窄切片，約 **15–20 條**可執行任務。這是工作量區間，不是完成百分比。

## 七個主要群組

| 群組 | 尚缺的玩家玩法證據 | 不應混入的已完成／實作工作 |
| --- | --- | --- |
| 隨機事件剩餘 record | raw type 2／6、其餘持續 record、AI 殖民地／艦隊／外交複合回寫與版本衝突 | 已閉合排程、Lucky、事件 4／5／8／9／13／18／27 不重開 |
| 原版 AI 回合殘餘 | `Compute_AI_Data_` 未表示欄位、殖民地配置、建造／艦隊 target state、全域 PRNG 與 Uncreative status | 已閉合 raw profile、科技 trait category、常態 application 抽選不重開 |
| AI 生產與艦艇產品 | 其餘建築分數區、帝國配額、支援／戰鬥艦產品、未解 raw mods | 已閉合 40 個建築分數與六艦體設計庫不重開 |
| 安塔蘭週期入侵 | owner 8 逐座標航行、完整快速／戰術 record、固定防禦與戰後 consumer | 五種怪獸 blueprint 與要塞既有證據分列，不以實作近似冒充 |
| 事件怪獸移動／戰術殘餘 | 途中逐座標截擊、Plasma Flux 對在途飛彈、快速反擊目標與戰術 AI | 已閉合藍圖、艦艇擴散、戰機傷亡、Caustic Slime、轟炸與分裂不重開 |
| AI 協議與政府演化 | AI↔AI 原版協議矩陣、正式終止／關係回寫與 AI 政體升級 | 玩家可表示的逐回合貿易／研究協議公式已閉合 |
| 狀態播報與投降殘餘 | 33／35 的 1.31 觸發 caller、投降 `+0x717`／`sub_27A3D` 接收者評分 | 29–35 record 與 34 資產移交 consumer 已閉合 |

## 非阻塞 oracle／命名清單

- raw 6／4／7 profile 候選的正式玩家名稱與客製外觀 raw27 來源；數字權重與 consumer 已閉合。
- battle record `+0x24／+0x4B` enum、star `+0x38` 前哨站 writer。
- 事件 1.50 差異、raw GNN 精確呈現時序與 JIMTEXT 按鈕文案。
- 原版 sprite／逐像素／動畫 timing；除非會改變 hitbox、規則時序或玩家輸入 gate，不阻塞玩法 RE。

## 三條高風險垂直鏈審查

本輪建立 [`remake-traceability.tsv`](remake-traceability.tsv) 與
[`scripts/check-re-traceability.sh`](../../scripts/check-re-traceability.sh)，先納入最容易誤判的三輪：

1. `0x589D6`：RE 已閉合、spec READY、source 部分實作、驗證不足。四候選初始權重與
   Ship Defense／Attack index 有已證實差異，不能因總和測試綠而標完整。
2. `0xFC845`：RE 已閉合、spec READY、category／共同估值 source 已部分接線；但上游 profile
   仍不符合，因此整條只有 PARTIAL／INTERNAL，不能標 CONFORMED。
3. `0x62C70／0xE5832`：RE 已閉合、spec READY；Advanced Civilization 全圖分配與 Money 初始
   國庫都沒有 source／test，明確標 MISSING，不因 `TechLevels` 可往返就誤判完成。

後續每閉合一條 RE，先追加 immutable key 與 evidence backlink，再更新 spec、source、test 與
verification state。README／WORKLIST 只能消費這份當前鏈，不從歷史 `[x]` 推斷實作完成。
