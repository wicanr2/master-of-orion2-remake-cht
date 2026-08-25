# 議會人口票數換算規格

## 範圍

本規格只修正單一帝國「人口總數 → 基礎議會票數」的換算。候選人選擇、第三方投票、棄權、
三分之二門檻與議會排程仍是其他 RE 切片，不由本規格改動。

## 證據契約

- 原版輸入：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- 工具與位址：IDA Pro 9.4／IDAPython；DOS/4GW LE object #1 的 IDA 線性位址。
- 原始定位：`sub_15B90 @ 0x15B90`；外部符號索引別名 `Calc_Council_Vote_`。
- 證據等級：正人口換算公式與議會消費端均為「已證實」。完整指令與 caller 證據見
  [`parity-re-audit-20260812.md`](../re/parity-re-audit-20260812.md#議會票數)。

## 行為規格

`gamedata.CouncilVotes(population)` 必須遵守：

1. `population <= 0` 回傳 `0`，作為 remake 的失敗安全輸入政策。
2. `population > 0` 回傳 `ceil(population / 10)`，以整數運算實作。
3. 所有議會候選排序、總票數、搖擺票與 `CouncilStatus` 必須繼續共用此函式，不另建第二套公式。
4. 不改動人口儲存單位；目前 shell 以各殖民地 `Population` 加總後傳入，與原版帝國人口欄位的
   玩家可見尺度相同。若日後證據推翻此尺度，須先訂正 RE 文件與本規格，再修改程式。

## 驗收

- 純規則邊界：`-5→0`、`0→0`、`1→1`、`10→1`、`11→2`、`42→5`。
- 抽樣執行 `internal/gamedata` 與 `internal/shell` 議會測試，確認候選、分母與勝負測試已同步新尺度。
- 測試綠只表示 remake 符合本規格；原版證據來自上述 IDAPython 靜態鏈。
