# 自繪畫面 → 原版佈局重建計畫

> 日期:2026-07-10。目的:標出哪些畫面目前是「我自繪佈局」(非原版 UI),需比照種族選擇畫面的做法重建成原版佈局,以達「操作畫面跟原版一樣」。
> 已用原版美術 + 量測熱區的畫面(主選單/星系/殖民地/種族選擇/自訂點數/命名旗色/研究抉擇/軍官…)不在此列。

## 待重建的自繪畫面

### 1. 外交對談(diplomacy,`cmd/moo2/interactive.go` diplomacyScreen)
- **現況**:用 `DIPLOMAT.LBX#29`(議事大廳)當背景 + 自繪深色選項按鈕(提議和平/貿易/威脅)。功能可用但**非原版佈局**。
- **原版**:顯示**該種族使節的動畫肖像 + 其專屬房間** + 對話文字 + 提議選項(原版 UI 樣式)。
- **重建所需(已初步調查)**:
  - `DIPLOMAT.LBX` 共 **39 個 asset**;0–6 為小圖示(已渲染確認),使節肖像在中後段(15/20/25/28 等**需調色盤鏈**才能渲染,推測 palette provider = asset 0,同 room#29 的解法 `loadDiplomatRoom`)。
  - **工具已備**:`cmd/moo2` 加了 `-palasset <n>`(用另一資產當調色盤)+ `-accum`(多幀 delta 累積)渲染旗標,可 `moo2 -lbx diplomat.lbx -asset 15 -palasset 0 -accum -shot out.png` 渲染需調色盤鏈的動畫資產。
  - **⚠ 關鍵障礙(2026-07-10 深入查明)**:DIPLOMAT.LBX 的大圖資產(13+,如 15/21/27)是**王座廳場景**(中央有使節/領袖figure),但**解碼有雜點缺陷**:
    - asset 0–12 有內嵌調色盤(provider 候選);13+ 無,需借調色盤。
    - asset 15 換 palette provider 0/10/12 **都仍雜點**(中央 figure 區尤重)→ **問題不在 palette 選擇**,而在這批**多幀 delta 動畫資產的 delta 解碼 / 動態調色盤**未被 `internal/lbx` 乾淨還原。現況 diplomacy 用的 room#29 中央也有同款雜點(見 `diplo_cur.png`)。
    - **→ 真正的下一步是解決 `internal/lbx` 對 DIPLOMAT 動畫資產的解碼**(可能是 color-cycling 動畫調色盤,或 delta 幀重建的邊角情形),對照 openorion2 `lbx.cpp` 影像解碼再逐位元組核。這是**解碼器層級的 RE**,不是畫面佈局工作。解決後才能乾淨顯示使節,再談重建佈局 + 逐族資產對應。
  - 渲染工具已備(`-lbx ... -palasset N -accum`),供這項解碼 RE 的視覺驗證。
  - **雜點根因確認(2026-07-10 徹底查明)**:`AccumulatedRGBA` 把全部 38 幀已寫像素疊起——**動畫中會動的中央使節/能量在各幀位置不同,全疊=散點**(靜態廊柱不動故乾淨)。但改渲染**單一 frame 0** 也不乾淨:結構較清楚,惟**天空/穹頂區變白噪點**(疑該區 index 對 palette 0 映射錯,或稀疏星場)。→ **兩種渲染法都不乾淨**,換 palette provider 也不解。結論:DIPLOMAT 動畫資產是**動態調色盤(color-cycling)+ delta 幀**的困難組合,忠實靜態呈現需**聚焦解碼 RE**(逐幀 + palette cycling 分析,對照 openorion2 逐位元組),**非快速修法**。
  - **現況最佳解**:diplomacy 畫面目前用的 **room#29(accumulated)是目前最乾淨的可用資產**(比 15/21/27 少雜點),故先維持;deeper 忠實(逐族使節 + 乾淨動畫)待上述解碼 RE。
  - 重建:依對手種族選對應使節肖像(動畫)+ 房間 + 對話框 + 提議按鈕(位置量測自原版)。
- **接線**:種族關係畫面「報告」→ diplomacy;已接 `bgmDiplo` 場景音樂。

### 2. 格子戰術戰鬥(tactical,tacticalScreen)
- **現況**:自繪星空底 + 格線 + 艦艇 token + HP 條;戰鬥數學**已接 gamedata 真公式**(命中/傷害/過盾/過甲,見 `gameplay-systems-status.md`)。
- **原版**:`COMBAT.LBX` 的戰場背景 + 艦艇 sprite(`CMBTSHP.LBX`)+ 原版 UI 控制列。
- **重建**:換原版戰場背景 + 真艦艇 sprite;保留已忠實的戰鬥數學。屬中型工作(sprite 對應 + 佈局)。

## 建議做法(比照種族選擇畫面的成功流程)

1. 渲染原版對應畫面/資產(需調色盤鏈者用鏈),**Read 圖判讀**佈局與元件位置(渲染出的原版美術即 oracle,不需外部截圖)。
2. 建 origScreen,載入原版背景 + sprite,依讀圖量測設定元件/熱區座標。
3. headless 導覽渲染驗證;逐步對齊。

## 現況小結

操作畫面「跟原版一樣」的剩餘主要是這兩個自繪畫面的重建(diplomacy 較大,需使節資產調查;tactical 中型)。其餘 overlay 畫面多已用原版美術 + 量測熱區。像素級微調為長尾。
