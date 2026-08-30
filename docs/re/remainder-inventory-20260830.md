# 玩家玩法 RE 剩餘盤點與垂直追溯（2026-08-30）

## 盤點口徑

本盤點只計會改變原版玩家可見玩法、且尚缺輸入／規則／consumer 垂直證據的 RE。以下不計入：

- compiler helper、C runtime、Windows API 與 DOS 硬體內部；
- 已閉合 RE、但 remake 尚未實作或尚未 CONFORMED 的項目；
- raw 名稱、逐字文案、sprite／逐像素與 1.50 profile 差異等非 1.31 gameplay gate；
- 沒有玩家 consumer 的欄位與只剩資料模型近似的實作工作。

依此口徑，41 列 parity matrix 並不是「41 項 RE 未完成」。本表原列七群；2026-08-30 起逐群
以 IDA 與既有垂直鏈重審，已閉合者直接標明，不再以舊群組數推估剩餘量。

## 七個主要群組

| 群組 | 尚缺的玩家玩法證據 | 不應混入的已完成／實作工作 |
| --- | --- | --- |
| 隨機事件剩餘 record | **已閉合**：`sub_23DFE` 的 2／6 已訂正為事件 16／24 的 record 狀態；29 種隨機事件的 1.31 建立／主要 consumer 與玩家、熱座、AI typed 回寫已有逐事件證據 | 1.50 二進位差異與 GNN 逐幀呈現是獨立 profile／UI gate，不再列入 1.31 玩法 RE |
| 原版 AI 回合殘餘 | **已閉合**：三張 cache 指標的完整直接 consumer census 已完成；殖民地職務、建造／艦隊 target 分流、Uncreative 狀態與 PRNG 停止線均有明確歸屬 | cache 配置內部、無玩家 consumer 欄位及 Watcom qsort 等價 partition 不納入 remake；source 仍為部分接線 |
| AI 生產與艦艇產品 | **已閉合**：48 個 building ID 的完整分數／零分、三張 quota、D10EE 六 case、Agent／貨運艦／Transport／Marine Barracks、支援 pseudo-product → ship slot 與 role 0..4 藍圖均有證據 | remake 仍為部分接線；raw mods 若沒有獨立玩家 consumer 不再列為生產 RE |
| 安塔蘭週期入侵 | **已閉合**：owner 8 共用 129-byte ship 的逐座標 route；抵達後進共同 battle side／快速或戰術鏈；`sub_E87D2` 戰後依艦級呼叫 `sub_6485F` 歸還 offensive pool，再刪除 deployed ship record | remake 的 ETA、聚合 combatant 與殖民地防禦投影仍是明示近似，不冒充資料模型一致 |
| 事件怪獸移動／戰術殘餘 | **已閉合**：途中只共用一般逐座標航行，抵達星系才搜尋戰鬥，沒有獨立「途中截擊」consumer；Plasma Flux 對 26-byte 飛行物、快速戰鬥入口／目標與戰術特殊武器 consumer 均有證據 | remake 同步飛彈沒有原版在途 record，快速目標與戰術 AI 是可重播近似，屬實作模型限制 |
| AI 協議與政府演化 | **已閉合**：ordered AI pair 的條約／協議／納貢門檻、宣戰／停戰、方向關係記憶與政府科技寫入 1／3／5／7 及其玩家 consumer 均有垂直證據 | 玩家方向矩陣與未消費 raw 欄不由此冒稱完成；remake 部分接線另由 trace state 表示 |
| 狀態播報與投降殘餘 | **已閉合**：29–35 record 與 dispatcher、29／30／31／32／34 caller、34 觸發及延後資產 consumer 已閉合；完整 direct-xref census 證實 33／35 在 1.31 沒有 caller，因此停止於未使用 setter 契約 | remake 33／35 與投降接收者選擇保留 trigger approximation，不反推不存在的 1.31 caller |

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

## 七群收斂結論

2026-08-30 七群的原版 1.31 玩家玩法 RE gate 已全部閉合。彙總證據、停止線與「RE 已閉合但
remake 仍為部分接線／近似」的逐群區分，見
[`seven-group-closure-audit-20260830.md`](seven-group-closure-audit-20260830.md)。這不會自動把
trace ledger 中的 `PARTIAL`、`MISSING` 或 `INTERNAL` 升格為 `CONFORMED`。
