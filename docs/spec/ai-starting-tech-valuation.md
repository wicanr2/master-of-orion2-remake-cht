# AI 開局科技估值規格

狀態：CONFORMED

RE-TRACE: dos-orion2-1.31:0xFC845

CORRECTION-20260830-PROFILE-UPSTREAM：`aiProfileCategoryValue` 已涵蓋目前證實的 trait
category direct-site 表；其 `OriginalAITechProfile` 上游 `0x589D6` 的四候選初值與 Ship
Defense／Ship Attack index 已於 2026-08-30 訂正並通過逐格測試。已知上游阻塞解除，這條
remake 可表示的開局／常態 AI 估值鏈升為 CONFORMED；原版全域 PRNG 位元序仍不在此聲明內。

## 輸入

- 原版種族索引與 31 格已展開 runtime 特性；轉換順序見
  [`ai-profile-trait-conversion-order.md`](ai-profile-trait-conversion-order.md)。
- 難度 raw `0..4`。
- AI 獨立、可重播的開局亂數流。
- 已知科技、對手已知科技、當下可研究主題與每回合研究點。

## 契約

1. 先抽 raw27，再依種族特性與難度建立 raw 6／4／7 profile；不得以中文性格名或 remake AI profile 代替 raw 值。
2. 每組權重總和超過 1000 時反覆除 2，然後各消費一次加權抽選。
3. 每個科技應用以 category 靜態值開始，順序套用 raw4、raw7、raw6、種族特性與特定 tech 覆寫，再進入 `sub_FC845` 共用後段。完整 trait direct-site 表見
   [`../re/ai-trait-profile-tech-homeworld-audit-20260830.md`](../re/ai-trait-profile-tech-homeworld-audit-20260830.md)。
4. 依主題成本／研究點得到視野分數後，難度大於 0 時再套 raw6 最高分門檻。
5. 候選篩選完畢只做一次應用級加權抽選，選中應用同時決定完成主題。
6. 不可用浮點難度倍率或 `TechCategoryWeight` fallback 冒充已知 AI profile；只有無法對到原版種族的舊存檔才安全回退。

## 驗收

- 全零特性的 6／4／7 基礎權重總和為 9／6／13；raw27=0 時七項總和為 16。
- raw profile 固定時，category switch 的 20／50／100 樣本與 IDA 立即數相同。
- 同開局種子产生同一組 AI 應用；不同 AI 亂數流不共用消費位置。
- 正常開局、存檔往返與網路快照後的 `ChosenTech`／`ExplicitChoice` 不遺失。

證據見 [`../re/ai-starting-tech-profile-audit-20260825.md`](../re/ai-starting-tech-profile-audit-20260825.md)。
