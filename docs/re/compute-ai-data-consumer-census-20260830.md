# `Compute_AI_Data_` producer／consumer census（2026-08-30）

## 證據契約

- 輸入：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- 資料庫：`Orion2.exe.i64`，SHA-256
  `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`；正式檔唯讀，分析一次性副本。
- 工具：IDA Pro 9.4；IDA linear，DOS/4GW LE object #1。
- 可重生腳本：`tools/ida/audit_compute_ai_data_consumers.py`。
- 非破壞性匯出：
  [`evidence/compute-ai-data-consumers-ida-20260830.json`](evidence/compute-ai-data-consumers-ida-20260830.json)。

## 已證實的生命週期

`Next_Turn_Calc_ @ 0x136B3` 在 AI 外交、殖民地決策、移動與研究之前呼叫
`Compute_AI_Data_ @ 0xD3D34..0xD574D`；唯一直接 caller 是該回合主鏈。`Deallocate_AI_Data_
@ 0xD574D` 在殖民地與帝國變更套用後釋放暫存資料。這些 cache 不是存檔狀態，也不是需要在
remake 複製位元配置的遊戲資料格式；需要重製的是有玩家 consumer 的 typed 輸入與規則。

## 三張 cache 指標的完整直接 consumer census

IDA 對三個全域指標的全部直接 xref 顯示：

| 指標 | 直接 consumer（保留 raw 位址） | 玩家玩法分類 |
| --- | --- | --- |
| `dword_1AA1E4 @ 0x1AA1E4` | `sub_D3D34`、`sub_D9F85 @ 0xD9F85` | 星系／帝國艦隊摘要；入侵截擊 |
| `dword_1AA1EC @ 0x1AA1EC` | `sub_D0036`、`sub_D0B08`、`sub_D0D2F`、`sub_D10EE`、`sub_D2CAE`、`sub_D7171`、`sub_DA99C` | 建築分數與抽選、帝國產品／購買、殖民地 worth、領袖指派、運輸 |
| `dword_1AA1F8 @ 0x1AA1F8` | `sub_D0036`、`sub_D10EE`、`sub_D2AEA`、`sub_D7171`、`sub_D896F`、`sub_DA99C` | 距離／外交／艦隊 ETA 壓力、帝國產品、proximity worth、殖民船目的地、運輸 |

除 producer 本體外沒有其他直接指標 owner。指標經暫存器完成的間接欄位讀寫，已分別由建造
分數、殖民地職務、食物運輸、殖民、領袖、艦隊目標與入侵專題沿 consumer 追回；不能因 IDA
直接 xref 沒列出每個 `cache+N` 就宣稱沒有 consumer。

## 群組 2 閉合結論

- 殖民地配置已有逐人口職務排序、封鎖分流、帝國工業／研究平衡及後置補農夫證據。
- 建造與艦隊 target state 並非「AI 回合未知欄位」；它們分屬群組 3 生產規則及已閉合的
  殖民、食物運輸、玩家目標、機會攻擊與星際移動專題。
- Uncreative 的單一 application 狀態在 `sub_E4204 → sub_E408F` 建立可選集合時已閉合；
  常態 AI 研究只是消費同一 application 狀態。舊文件把它列成 AI 回合待解是分類錯誤。
- 原版全域 PRNG 的逐位元位置沒有必要成為玩法 RE gate。remake 必須保存自己的決定性 stream
  並維持已證實的同一決策內呼叫順序，但不得宣稱與原版存檔 seed 位元一致。
- cache 中沒有已證實玩家 consumer 的配置／釋放細節，以及 qsort 等價項的 Watcom runtime
  partition，依編譯器輔助與無 consumer 停止線排除，不再計入 remake RE。

因此「原版 AI 回合」群組的 1.31 玩家玩法 RE 已閉合；remake 仍有多個明確實作差異，需由
各 READY spec 逐條接線，不能把本結論寫成 source 已 CONFORMED。
