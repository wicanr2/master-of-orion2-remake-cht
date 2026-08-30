# AI trait profile、科技估值與 Advanced Civilization 初始國庫稽核（2026-08-30）

## 證據契約

- 輸入 `Orion2.exe` SHA-256：
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- IDA 資料庫 `Orion2.exe.i64` SHA-256：
  `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- 外部符號表 `symbols_fixed.tsv` SHA-256：
  `f83116049964c14072547dd087a6316117cf62808c328cc7c087984467030b28`。
- 工具：IDA Pro 9.4、`tools/ida/audit_ai_trait_profiles_homeworld.py`；位址均為
  `Orion2.exe.i64` 的 IDA linear address、DOS/4GW LE object #1。
- 可重生證據：
  [`evidence/ai-trait-profiles-homeworld-ida-20260830.json`](evidence/ai-trait-profiles-homeworld-ida-20260830.json)。

正式 `.i64` 唯讀掛載並複製至 `/tmp` 分析。匯出保留原始函式邊界、bytes、operand、caller、
trait direct sites 與外部導航符號；外部名稱不取代 raw 位址或證據等級。Watcom distant tail、
`Random`／除法 runtime 與平台 API 不納入玩法分母。

## 第一輪：Cybernetic／Lithovore 的 NPC profile 權重

`Init_NPC_Personalities_Objectives_Themes_ @ 0x589D6` 建立三組權重，候選數依序為 6、4、7。
每組若總和大於 1000，就把全部權重以 signed `/2` 重複縮小，直到總和不超過 1000；接著以
`Random(sum)` 落入累積區間。結果分別寫 `player+0x28`、`player+0x205`、`player+0x206`。
`player+0x28 == 100` 的真人 sentinel 會保留，不被第一組抽選覆寫。

下表只使用 raw profile 組與候選 ID。函式外部名稱雖列出 Personalities／Objectives／Themes，
但目前沒有獨立 consumer 足以把三組逐一固定到這三個英文語意；因此不以名稱升格事實。

| trait | 6 候選組 | 4 候選組 | 7 候選組 |
| --- | --- | --- | --- |
| Cybernetic `+0x8B0` | raw ID 4 `+3` | raw ID 0 `+10` | raw ID 5 `+1000` |
| Lithovore `+0x8B1` | raw ID 2 `+10`、ID 3 `+3`、ID 4 `+3` | 無 | 無 |

Cybernetic 分支位於 `0x58CCB..0x58CE0`；Lithovore 位於 `0x58CE9..0x58CFB`。以上是**已證實**
的抽選權重與寫入；候選 ID 的玩家可見名稱仍為未知，不從區域變數名或二手符號猜填。

## 第二輪：`Calc_Tech_Value_` 的 raw category 0 trait 覆寫

`Calc_Tech_Value_ @ 0xFC845` 先由 13-byte tech application record 的 `+3` 取 raw category，
由 category table 建立 `ecx` 類別倍率，再接 tier 基礎值、profile、已研究邊際遞減、上限扣抵、
開局加成與主題完成度。因此本節的 `ecx` 是完整估值中的**類別倍率**，不是最終回傳值。

raw category 0 由 `0xFCB10 → 0xFCB69` 進入下列順序：

```text
if FarmingRaw < 0: categoryMultiplier = 100
if FarmingRaw > 0: categoryMultiplier = 10
if Lithovore:      categoryMultiplier = 1
else if Cybernetic: categoryMultiplier = 20
```

trait 覆寫位於 `0xFCBB4..0xFCBD3`；Lithovore 先判斷，成立時直接走 `ecx=1`，所以即使不合法的
食物科技已由另一個 gate 排除，這裡仍保留其通用估值契約。只有 Lithovore 不成立時才檢查
Cybernetic。這閉合兩項 trait 在此 category 的精確倍率與優先序；其他 raw category 的 trait
分支仍由完整科技估值專題管理，不以本切片概括為全 trait 已閉合。

## 第三輪：Money 與 Advanced Civilization 初始國庫

`Twiddle_Initial_Homeworlds_ @ 0xE5832` 先執行共同星圖／殖民地初始化；只有
`byte_199CB5 == 2`（開局文明等級 Advanced Civilization）才進 `0xE58EE` 後段。它逐玩家讀取
signed `player+0x8A4` Money raw，於 `0xE5954..0xE596C` 寫：

```text
player+0x32 = trunc((MoneyRaw + 2) * 400 / 4)
            = (MoneyRaw + 2) * 100
```

標準 Money raw `-1／0／+1` 因而給 Advanced Civilization 開局國庫 `100／200／300 BC`。
乘數必為 100 的整數，沒有額外捨入差異。這是**初始國庫** consumer，不是修改母星 size、
氣候、礦產或 special；後者由 `Modify_Home_Worlds_` 與 Advanced Civilization 行星平衡器分別
管理。`Twiddle_Initial_Homeworlds_` 的其餘艦艇、接觸與開局配置不依賴其他 trait direct byte，
故 Aquatic／Subterranean 的母星效果不應在此函式另開缺口。

## 閉合與 remake 邊界

- **已證實**：兩項 trait 的三組 raw profile 權重、縮放與抽選結果欄位；raw category 0 的
  Lithovore／Cybernetic 科技倍率與優先序；Advanced Civilization 的 Money 初始國庫公式。
- **未知**：三組 profile 候選的正式玩家語意名稱；其他 trait／category 的完整 AI profile 與
  科技估值表。未知不推翻本輪 raw 權重與公式。
- remake 的 profile、科技選擇與 Advanced Civilization 全圖開局仍是較窄模型；依全域
  RE-first gate，本輪只補證據與差異，不建立未經 READY spec 的 Go 實作。
