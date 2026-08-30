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

## 第四輪：五項 signed 經濟 trait 的 raw profile 權重

三組候選的 raw ID 對應仍採上一節的 6／4／7 組。初始權重依序為
`[1,2,1,2,2,1]`、`[2,1,2,1]`、`[2,2,1,2,1,2,3]`，再套下表：

| trait 條件 | 6 候選組變更 | 4 候選組 | 7 候選組 |
| --- | --- | --- | --- |
| Population `==50` | ID5、ID2 **設為** 11 | 無 | 無 |
| Population `==100` | ID5、ID2 `+100` | 無 | 無 |
| Farming `==1` | ID5、ID4、ID3 各 `+3` | 無 | 無 |
| Farming `==2` | ID5、ID4、ID3 各 `+100` | 無 | 無 |
| Industry `==1` | ID5、ID4、ID3 各 `+3` | 無 | 無 |
| Industry `==2` | ID5、ID4、ID3 各 `+10` | 無 | 無 |
| Science `==1` | ID4 `+3`、ID3 `+10` | 無 | 無 |
| Science `==2` | ID4 `+10`、ID3 `+100` | 無 | 無 |
| Money | 無 direct site | 無 | 無 |

Population 第一級是覆寫初值，不是加 10；這是 `0x58A9F..0x58AB0` 的 raw `mov 0xB`。
其餘位於 `0x58AB7..0x58B64`。函式只比對列出的精確正值，負值不在 profile 初始化直接加權；
負值仍會在玩家經濟與科技估值 consumer 生效。以上是**已證實** raw 權重。

## 第五輪：其餘 trait 的 raw profile 權重

| trait 條件 | 6 候選組變更 | 4 候選組變更 | 7 候選組變更 |
| --- | --- | --- | --- |
| Ship Defense `==20` | ID1 `+10` | ID1 `+10` | ID3 `+100` |
| Ship Defense `==40` | ID1 `+100` | ID1 `+100` | ID3 `+1000` |
| Ship Attack `==25` | ID1 `+10` | ID0 `+100` | ID4 `+3` |
| Ship Attack `==50` | ID1 `+100` | ID0 `+1000` | ID4 `+10` |
| Ground Combat `==10` | ID2 `+10` | 無 | ID1 `+100` |
| Ground Combat `==20` | ID2 `+100` | 無 | ID1 `+1000` |
| Spying `==10／20` | ID0 `+3／+10` | 無 | 無 |
| Low-G | ID4 `+10`、ID5／ID3 各 `+3` | 無 | 無 |
| High-G | ID2 `+10` | 無 | ID1 `+1000` |
| Aquatic | ID4／ID5／ID3 各 `+3` | 無 | 無 |
| Subterranean | ID4 `+10`、ID5／ID3 各 `+3` | 無 | 無 |
| Large Homeworld | ID4 `+3`、ID5 `+10` | 無 | 無 |
| Rich Homeworld `==1` | ID4 `+10` | 無 | 無 |
| Charismatic | ID0 `+100` | 無 | 無 |
| Creative | ID3 `+1000` | 無 | 無 |
| Tolerant | ID4 `+100` | 無 | 無 |
| Fantastic Traders | ID0 `+100` | 無 | 無 |
| Stealthy Ships | ID1 `+100` | 無 | 無 |
| Warlord | ID2 `+10` | 無 | 無 |
| Trans-Dimensional | ID2 `+10`、ID1 `+13` | 無 | ID3 `+100` |

Cybernetic／Lithovore 已列於第一輪。Poor Homeworld、Repulsive、Uncreative、Telepathic、Lucky、
Omniscience 在此函式沒有 direct profile 加權；不能把「沒有 direct site」解讀成整體 AI 無效果，
它們各自仍有外交、研究、事件或地圖 consumer。表格對應 `0x58B6D..0x58D8B`，證據等級為
**已證實**；raw ID 正式語意名稱仍未知。

## 第六輪：完整 trait × raw tech category 覆寫表

`0xFCA88..0xFCDE9` 是 `Calc_Tech_Value_` 的 raw category 分派。下表列出所有會讀
`player+0x89F..+0x8BD` 的 category branch；「不變」表示保留 category table 或較早 profile
分支產生的 `ecx`。

| raw category | trait 條件 | category multiplier `ecx` |
| ---: | --- | ---: |
| 0 | Farming `<0`／`>0` | 100／10 |
| 0 | Lithovore；否則 Cybernetic | 1；20 |
| 1 | Industry `<0` | 100 |
| 2 | Science `!=0` | 100 |
| 3 | Money `<0`／`>0` | 100／20 |
| 4 | Industry `>0` | 100 |
| 4 | Tolerant | 1（最後覆寫） |
| 6 | Subterranean | 20 |
| 6 | Population `<0`／`>0` | 100／5（在 Subterranean 後覆寫） |
| 12 | Spying `!=0` | 50 |
| 12 | `trunc(GovernmentRaw/2)==2` | 50 |
| 16 | Ground Combat `<0` | 20 |
| 18 | Ship Defense `<0` | 50 |
| 25 | Ship Attack `<0` | 100 |
| 27 | Ship Attack `>0` | 100 |
| 28 | Ship Defense `>0` | 100 |
| 37 | Stealthy Ships | 1 |
| 40 | `trunc(GovernmentRaw/2)==3` | 1 |

另有兩個 tech application raw ID 特例，不走 category 表：ID 5 且 Telepathic 時 `ecx=1`；
ID 131 在 High-G 時 `ecx=1`，否則 Low-G 時 `ecx=50`。位址為 `0xFCDE9..0xFCE46`。
表內 signed division 依 Watcom `cdq; sub eax,edx; sar eax,1` 向零截斷。這閉合全部 25 個
trait direct site 的輸入、優先序與倍率；後續 tier、已研究邊際遞減、profile、cap 與完成度仍
照 `Calc_Tech_Value_` 共同鏈處理，不能把表內倍率當最終科技 worth。

## 2026-08-30 更新後邊界

- 三組 raw profile 的所有 trait direct site，以及 `Calc_Tech_Value_` 的所有 trait direct site
  均已閉合。
- 仍未知的是 raw profile ID 的正式名稱、tech category 的玩家名稱與非 trait 共同估值表名稱；
  原始數字、控制流與玩家／AI consumer 不因名稱未知而失效。
- 這三輪仍屬 RE-only；remake profile／科技估值接線須待完整 RE gate 後建立 READY spec。
