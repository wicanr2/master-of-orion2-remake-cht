# MOO2 逆向工程完整性第二輪審查（2026-08-13）

## 結論

目前 MOO2 的逆向工程**不完整**，而且缺口不只是一批尚未追回的常數。專案已累積大量資產格式、
畫面座標、資料表、存檔欄位與局部公式證據，但原版核心玩家機制尚未普遍形成「原版入口 → caller
與資料欄位 → RNG／分支 → 回寫 → Go 消費端 → 正常玩家路徑測試」的垂直鏈。

現況比較準確的描述是：

- 資產、畫面與部分資料表逆向較深。
- 若干局部公式已由 IDA 靜態證實。
- 核心 4X 回合系統的覆蓋不均，許多 Go 功能仍是手冊形狀加 remake 模型。
- 原版 AI、事件、研究完成、安塔蘭戰後消費等仍有未閉合子系統；間諜、艦隊航行、殖民／前哨站
  與殖民地完整回合外層已於 2026-08-28 閉合，其他子公式仍依矩陣分列追查。

本文件不換算「完成百分比」。函式數、符號數與矩陣列數只能量測研究表面，不能代表語意完成度。

## 輸入、工具與位址基準

- 原版輸入：`Orion2.exe`，2,644,842 bytes。
- SHA-256：`7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- 主要工具：IDA Pro 9.4；既有成功稽核使用 `ida-pro-9.4-ver3:py312-v1`。
- 位址基準：IDA 線性位址，`min_ea=0x10000`、`max_ea=0x1D5CD0`。
- 原版符號索引：8,589 筆，其中函式名稱 4,201 筆。
- 一次性 IDA 自動分析：辨識 5,397 個函式；未套外部符號前只有 259 個非 `sub_` 函式名。
- 本輪證據來源：2026-08-12 一次性 `.i64` 原始輸出、原版外部符號索引、目前 Go 呼叫端、
  受版控 RE 文件與工作樹。攤平 `.asm` 只用來顯示既有指令文字，不作 call graph 完整性證據。

## 回合主鏈完整性

`Next_Turn_Calc_ @ 0x136B3` 已於 2026-08-28 由 IDA 資料庫重新匯出，確認為 **52 個**直接
`call` 指令。完整順序、callsite、target、三個條件 gate、
逐玩家招募迴圈與平台停止線見
[`next-turn-chain-audit-20260828.md`](next-turn-chain-audit-20260828.md)。以下只保留重要子系統摘要：

- `Antaran_Invasion_Check_ @ 0x63D92`
- `Compute_AI_Data_ @ 0xD3D34`
- `NPC_Diplomacy_ @ 0x252A7`
- `Diplomacy_Growth_ @ 0x4DD6B`
- `Do_AI_Leaders_ @ 0xD7439`
- `All_Colony_AI_ @ 0xD6F67`
- `Move_All_AI_ @ 0xDBB29`
- `All_AI_Colonize_ @ 0xE67F6`
- `All_AI_Tech_Select_ @ 0xDCA69`
- `Process_Trade_And_Research_Agreements_ @ 0x101E77`
- `Move_All_Ships_Toward_Stars_ @ 0xFFEEA`
- `Resolve_Spies_ @ 0x10192B`
- `Apply_All_Player_Changes_ @ 0xE4F49`
- `Apply_All_Colony_Changes_ @ 0xE3FDC`
- `Search_For_Battles_ @ 0xE9D62`
- `Determine_Event_ @ 0x2230A`
- `Do_All_Ships_XP_Check_ @ 0x14A27`
- `Do_Colony_Calculations_ @ 0xE2B31`（出現兩次）
- `Compute_Blockades_ @ 0xE5097`
- `Check_All_Rebellions_ @ 0xED44A`
- `Repair_Ships_At_Colonies_ @ 0x580F5`
- `Compute_Contacts_ @ 0xEB192`
- `Check_For_Council_Meeting_ @ 0x168AF`
- `Random_Officer_Check_ @ 0x97A66`
- `Record_History_ @ 0x10208A`

Go 的 `GameSession.EndTurn` 也有完整可玩流程，但它採另一組系統切割與順序；其中包含多個自訂
adapter、固定週期和 remake AI。原版 52 個 call 的外層順序已閉合；各子函式的輸入、狀態建立與
回寫仍由 parity matrix 的獨立列追查。因此「EndTurn 很長／功能很多」不是各子系統已對齊的證據。

## 覆蓋面盤點

原版具名函式以玩家行為關鍵字做重疊分群，得到下列導航規模：

| 導航分群 | 命中具名函式數 |
|---|---:|
| 回合、殖民地與經濟 | 243 |
| 科技與研究 | 97 |
| 星系、行星與艦隊 | 254 |
| 戰鬥、武器、飛彈與地面戰 | 262 |
| 外交、間諜與領袖 | 210 |
| 事件、勝利與安塔蘭 | 107 |
| 種族與開局 | 37 |

這些分群會互相重疊，也包含 UI 與 helper，不能相加或換算百分比。它們只證明原版玩家系統的
導航面遠大於第一輪 14 列矩陣。第二輪已把矩陣擴為 31 個系統列；新增項目仍多為「未知」或
「已證實模型不同」，所以矩陣擴充代表誠實盤點增加，不代表逆向完成度增加。

## 第二輪找到的直接反例

### 隨機事件

原版 `Determine_Event_ @ 0x2230A` 是 `Next_Turn_Calc_` 的直接子系統，另有
`Setup_Next_Event_ @ 0x21371` 與各事件檢查函式。本文件寫成時，Go 使用每回合 30% 與最多
重抽八次；2026-08-25 已由 `random-event-schedule-audit-20260825.md` 閉合並取代為前五次保護、
五級 delta 公式、`Random(512)` 與最多五個 0..28 候選。多個事件效果、全銀河目標與持續 record
仍未閉合，因此本節的「完整事件系統未逆完」判定仍成立，但 30% 不再是現行程式。

### 安塔蘭週期入侵

原版每回合先呼叫 `Antaran_Invasion_Check_ @ 0x63D92`，並有資源、攻守艦隊建造、目標選擇與部署
函式。2026-08-25 已由 `antaran-periodic-invasion-audit-20260825.md` 推翻固定第 20 回合、每
15 回合與直接扣人口／BC 的舊 remake：科技延遲、25 回合加速資源、攻守建艦、readiness、
`Random(200)`、Lucky 目標權重、殖民星抽樣、最多五艘、全局熱座一次與 pending ETA 均已接。
owner 8 的原版中途座標、完整快速／戰術 battle record 及殖民地戰後消費仍未逐欄閉合，故本節
維持「部分對齊」，但舊固定腳本不再是現行程式。

### 間諜

原版 `Resolve_Spies_ @ 0x10192B` 是回合主鏈直接子系統。2026-08-28 已閉合 packed pair、
self Agent pool、100 工業訓練、每名 1 BC 維護、三任務碼、AI 收回／重配及 RACES 寫回。
判定：原版 1.31 玩家可見 RE 已閉合；Go 的 30 BC 直訓、AI 週期免費補人與任務政策仍是
明確不同的 remake 模型。見 `spy-turn-policy-audit-20260828.md`。

### 艦隊航行

原版 `Move_All_Ships_Toward_Stars_ @ 0xFFEEA` 每回合推進每艘 ship 的實際座標。2026-08-28
已閉合 30 座標／parsec、逐步投影、ETA 衍生值、拆分／聚合、中繼、動態截擊、星雲、黑洞、
曲速場干擾器、gate 與抵達 consumer。判定：原版 1.31 玩家可見 RE 已閉合；Go 仍以整段 ETA
且玩家／AI 分裂的資料模型表示航行，明確不對齊。見 `fleet-interstellar-movement-audit-20260828.md`。

### 殖民與前哨站

2026-08-28 已從正常玩家、回合 dispatch 與 AI 入口閉合資格、owner gate、候選衝突、殖民船／
前哨船／Colony Base 消耗、半價退款、共用建立器、人口／Native、前哨站升級 Marine Barracks、
derived callback 與玩家通知。判定：原版 1.31 玩家可見 RE 已閉合；Go 仍只有一般殖民主要路徑，
屬部分對齊。見 `colonization-full-audit-20260828.md`。

### 研究

科技樹、成本表、抉擇 UI 與元件解鎖已接通。2026-08-25 已閉合
`Check_For_Research_Breakthrough_ @ 0xE44E0` 的突破率、擲骰與成功清零；
`Colony_Research_Production_ @ 0xDFF74` 亦已閉合研究人口 dispatch 及四棟研究建築的固定
5／10／15／30 RP。其餘人口修正與應用授予副作用仍未完整閉合，因此資料與操作完整度仍不等於
整條研究回合忠實度。

### 原版 AI

原版回合主鏈分別執行 AI 資料建立、殖民地 AI、移動、殖民、科技選擇、領袖與搜尋戰鬥。Go
`internal/ai` 與 `advanceAI` 明示不是原版 MOO2 AI。判定：現有 AI 是可玩的重製模型，不是原版 AI
逆向成果。

## 工具與證據缺口

2026-08-12 的 `re_audit_core.idc` 對 far call 使用 `get_operand_value`，在輸出中得到 55 個
`0xFFFFFFFFFFFFFFFF`，但原始指令仍保留 `call sub_...` 位址。2026-08-28 已由
`audit_next_turn_chain.py` 直接讀 IDA code operand 與函式邊界，正式輸出包含 52 個有效 target、
callsite bytes 與來源雜湊；舊 IDC 查詢不再作完成證據。

## 完整性閘門與建議順序

RE 完整性必須至少通過三層：

1. **回合主鏈覆蓋**：52 個直接呼叫逐一映射到子系統矩陣列或明示為 runtime／平台邊界排除項。
2. **系統垂直閉合**：每列完成 caller、欄位、RNG、分支、回寫、平行玩家／AI 路徑與存檔。
3. **獨立驗證**：同輸入與邊界值的 IDA 靜態測試；靜態不足才做原版動態 oracle，再走正常玩家路徑。

建議依玩家影響排序：

1. 回合主鏈、殖民地完整回合、研究、維護與人口。
2. 原版 AI、外交、條約、間諜與領袖。
3. 艦隊航行、封鎖、接觸、殖民、事件與安塔蘭週期入侵。
4. 快速／戰術戰鬥、地面戰、轟炸與行星防禦。
5. 議會、勝利、歷史、分數與客製種族跨系統消費端。

在上述核心矩陣閉合前，專案只能稱為「可玩 remake 預覽」。
