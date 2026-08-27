# AI 殖民地職務分配靜態稽核

日期：2026-08-28

## 證據契約

- 輸入 `Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- 正式 `Orion2.exe.i64`，SHA-256
  `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- 工具：IDA Pro 9.4，映像 `ida-pro-9.4-idapython:py312-v1`；位址為
  DOS/4GW LE image 的 IDA linear EA。
- 非破壞性匯出：
  [`evidence/ai-colony-jobs-ida-20260828.json`](evidence/ai-colony-jobs-ida-20260828.json)；
  保留原始 `sub_` 名、函式邊界、bytes、運算元與 caller。

`ida-pro-9.4-idapython:locked-v1` 本輪因 `Python3TargetDLL` 與 license
環境未載入而退出 1，沒有輸出證據；依正對照改用已驗證的
`py312-v1` 後，同一份 `.i64` 成功產生 189,241-byte JSON。前一次空輸出
不作為「函式不存在」或玩法結論。

## 已證實的控制流

1. `sub_D6E1D @ 0xD6E1D` 先分流封鎖殖民地
   `sub_D61E7 @ 0xD61E7` 與未封鎖殖民地 `sub_D652C @ 0xD652C`，
   處理完才呼叫帝國級 `sub_D66B3 @ 0xD66B3`。
2. 兩條殖民地路徑都先呼叫 `sub_D5FE1 @ 0xD5FE1`；它掃描
   `colony+0x0C` 起的 4-byte colonist records，計數當前職務，並以
   `qsort_` 排序可重新指派的人口。
3. `sub_D652C` 以 `(population+4)/5` 建立一個最低需求，呼叫
   `sub_E1D59` 重算殖民地輸出，並逐個修改 colonist record
   的 2-bit job 欄位；不是只寫轉換後的 farmers/workers/scientists 總數。
4. `sub_DDFD3 @ 0xDDFD3` 以 raw job 索引呼叫
   `funcs_DDFF0 @ 0x18355C`；三個 table target 依序是已獨立閉合的
   食物 `sub_DE0C6`、工業 `sub_DED47`、研究 `sub_DFE77`，因此 raw
   `0/1/2` 分別是農夫／工人／科學家。
5. `sub_D648A @ 0xD648A` 對候選範圍兩端分別以科學家 raw `2`
   與工人 raw `1` 呼叫 `sub_DDFD3`，將研究−工業邊際差寫入每個 13-byte
   殖民地 AI 記錄 `+8` 與 `+0x0A`。沒有可指派人口時分別寫
   `0x8001` 與 `0x7FFF`作為 sentinel。
6. `sub_D66B3` 先以 `sub_E2D72` 重算帝國，然後從所有
   未封鎖殖民地選最大 `+8` 候選，把一名人口改為 job raw `1`；
   每次都呼叫 `sub_D648A` 與 `sub_E2A70` 重算。到達帝國停止條件後，
   再以最小 `+0x0A` 候選逐名改為 job raw `2`。
7. 帝國停止條件直接讀 `player+0xAA` 與 `player+0xAC`。權重基值
   為 `10` 與 `18`；`player+0xB2 < 0` 時後者加
   `isqrt(-signed(player+0xB2))`，late-tech `player+0x59D` 使前者歸零，
   raw personality 3 使前者 `+2`，raw personality 4 使後者 `+2`。

上述第 1–7 點是原始指令、caller 與資料流可回查的已證實結論。

## 對 remake 的直接反證

`internal/engine.ApplyAIEconomy` 目前逐殖民地呼叫
`Decider.ColonyJobs`，每個殖民地一次得出三種總數。`RemakeDecider` 則先餵飽全體，
再依設計性 Industry／Research weights 切分餘數。原版是單人口排序、
邊際產出、全帝國迭代與每步重算；因此不能藉由讓
`NewDecider(ModeOriginal)` 回傳同一個逐殖民地 helper 就宣稱原版 AI。

## 待閉合

- `sub_D5FA9`、`sub_D614D`、`sub_D6315`、`sub_D63A6` 的完整排序鍵，
  特別是異族、android、士氣、重力與種族產出差異。
- `player+0xAA/+0xAC` 的寫入端與尺度，以及 `sub_D66B3` 兩個循環
  的 signed／unsigned 比較邊界。
- 現有 aggregate `ColonyState` 能否保留排序所需的逐種族資料；
  若不足，應先擴充 typed population groups，不得以平均值偽裝 exact。
