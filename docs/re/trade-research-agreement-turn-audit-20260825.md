# 貿易／研究協議逐回合處理稽核（2026-08-25）

## 問題與舊模型

`internal/shell/treaty.go` 原以 `(target-current)/5` 推進目前收益，並在整數商為零時
強制移動一點。這能產生平滑曲線，但沒有原版指令證據；`WORKLIST.md` 也仍把協議成長、
上限與雙方時序列為未知。本輪以正常回合入口往下追，不以既有 Go helper 反推原版。

## 輸入與工具

- 輸入：`Orion2.exe`
- SHA-256：`7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- 資料庫：正式 `Orion2.exe.i64` 的一次性使用者副本
- 工具：IDA Pro 9.4，IDAPython image `ida-pro-9.4-idapython:py312-v1`
- 位址基準：IDA linear；DOS/4GW LE object #1
- 可重生探針：`tools/ida/audit_trade_research_agreements.py`
- 證據等級：除特別註明者外，本頁指令與資料流均為**已證實**；Go 的獨立亂數流仍是
  remake 可重播對應，不宣稱原版 PRNG 狀態逐位元一致。

## 位址勘誤

兩份舊外部符號名稱把相鄰函式混在一起。原始函式邊界與 caller 證明：

| raw 位址 | 真正角色 | 直接證據 |
|---|---|---|
| `sub_101E77 @ 0x101E77..0x101EE3` | 每回合處理所有貿易／研究協議 | `Next_Turn_Calc_` 於 `0x13733` 直接呼叫；掃 `+0x62F/+0x637`，再呼叫兩個逐方向 helper |
| `sub_101EE3 @ 0x101EE3..0x101F82` | 建立雙向貿易協議 | 寫雙方 `+0x62F=1`、`+0x5A4=-base`、`+0x5B4=goal` |
| `sub_101F82 @ 0x101F82..0x102037` | 建立雙向研究協議 | 寫雙方 `+0x637=1`、`+0x5C6=-base`、`+0x5D6=goal` |

因此 `0x101EE3` 不能再稱為 `Process_Trade_And_Research_Agreements_`。本頁保留 raw
名稱與位址，不用推測性改名覆蓋資料庫。

## 建立協議

### 貿易

`sub_101EE3` 先以 `sub_101B3C` 計算雙方較小的 `player+0x5A2` 再除以二：

```text
base = min(left.tradeBase, right.tradeBase) / 2
left.current[right]  = -base       // +0x5A4
right.current[left]  = -base
left.goal[right]     = tradeGoal(left, right)   // +0x5B4
right.goal[left]     = tradeGoal(right, left)
left.active[right]   = 1           // +0x62F
right.active[left]   = 1
```

### 研究

`sub_101F82` 對 `player+0x5C4` 做同樣的 `min/2`，寫入 `+0x5C6/+0x5D6/+0x637`。
負起點是 `-base`，不是負目標值；政府與種族／領袖修正只進 goal helper。

## 每回合逐方向公式

貿易 `sub_101D53 @ 0x101D53` 與研究 `sub_101DE5 @ 0x101DE5` 是同形函式，欄位分別為
`+0x5A4/+0x5B4` 與 `+0x5C6/+0x5D6`：

```text
goal = recomputeGoal(empire, opponent)
current = storedCurrent

if current < goal:
    quotient  = trunc(goal / 5)
    remainder = goal % 5
    bonus = 0
    if remainder != 0:
        bonus = (Random(5) <= remainder) ? 1 : 0  // Random 為 1..5
    current = min(goal, current + quotient + bonus)
else:
    current = goal

storedGoal = goal
storedCurrent = current
```

重要後果：

- 每回合增加的是 **goal/5**，不是剩餘差距的五分之一。
- `goal % 5 != 0` 時，每個方向、每種協議各消耗一次 `Random(5)`；可整除時不抽亂數。
- 初始 `-base` 且 goal=`base` 的基準情況，第 5 回合到 0，第 10 回合到上限。
- current 已高於新 goal 時，該回合立即寫回 goal，不做五回合漸降。
- 沒有在這兩個 helper 內自動終止協議；active flag 為零才略過，終止由其他外交路徑負責。

## 全局順序與回寫

`sub_101E77` 以 player record `0xEA9` stride 由高槽位往低槽位掃描，每個方向先處理
`+0x62F` 貿易，再處理 `+0x637` 研究。對玩家 0 與 AI 槽位而言，原版先結算高槽 AI
方向，最後才以玩家 0 為 outer record、由高 AI 槽往低槽結算玩家方向。結束後呼叫
`sub_101A42`：清空 `+0x5A2/+0x5C4`，掃 packed colonist，依 owner 重新累加兩組協議基數。

remake 沒有完整原版 player record 與 AI↔AI 協議矩陣；本輪可精確承接玩家↔AI 的可表示
子集，採「AI 方向由高索引到低索引，再玩家方向由高索引到低索引」，每方向內貿易先於研究。
AI↔AI 原版協議仍屬 AI 外交切片，不藉此冒稱閉合。

## Remake 映射與剩餘限制

- `TreatyState.PlayerTradeValue/AITradeValue` 對應雙方向 `+0x5A4` current。
- `PlayerResearchValue/AIResearchValue` 對應雙方向 `+0x5C6` current。
- goal 每回合由現有 typed 帝國人口與政府／Trader helper 重算，不另持久化原版 cache。
- 新增獨立、可存檔的 agreement RNG；抽取條件與順序對齊，但 PRNG 演算法非原版逐位元重現。
- AI 目前政府仍固定 Dictatorship，是既存已明示近似；不阻擋本輪公式與時序閉合。
- `SpecialTradeState` 是 remake 擴充，不混入本輪原版普通協議結論。
- 外交畫面的單行「第 N 回合／目前 BC 或 RP」摘要是 remake 可觀察性轉接，並非上述欄位
  位址能證實的原版逐句文案。規則層現只輸出 typed `TreatySummaryPart`；名稱、格式與分隔符
  由 `assets/i18n/ui.json` 提供，不把顯示字串反寫成原版證據。
