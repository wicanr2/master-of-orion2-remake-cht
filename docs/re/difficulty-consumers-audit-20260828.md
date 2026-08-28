# 五級難度全域 consumer 完整索引（2026-08-28）

## 結論

`byte_199CB0 @ 0x199CB0` 的全部直接交叉參照已由 IDA Pro 9.4 重生：125 個指令站點，分屬
78 個 owner 函式。原版沒有一個可套遍全遊戲的「難度倍率」；每個子系統各自使用離散門檻、
整數加值、除數、機率或資料表。這份索引關閉「是否還有漏掉的直接難度 consumer」問題，但不把
每個大型 AI 子決策器的所有非難度欄位一併宣稱閉合。

本輪只補 RE，未修改 Go。編譯器 helper、C runtime 與平台 API 沒有被納入玩法分母。

## 證據身分

- 輸入：`Orion2.exe`
- 輸入 SHA-256：`7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- IDA 資料庫 SHA-256：`4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`
- 外部符號表 SHA-256：`f83116049964c14072547dd087a6316117cf62808c328cc7c087984467030b28`
- 工具：IDA Pro 9.4／IDAPython；位址均為 DOS/4GW LE object #1 的 IDA 線性位址
- 匯出：[`difficulty-consumers-ida-20260828.json`](evidence/difficulty-consumers-ida-20260828.json)
- 腳本：[`audit_difficulty_consumers.py`](../../tools/ida/audit_difficulty_consumers.py)

外部名稱只供導覽；JSON 每站同列保留原始位址、bytes、運算元、data refs、局部指令窗、owner
原名與完整 owner 反編譯。以下語意均經 raw 指令審查。

## 儲存、畫面與多人同步邊界

這一組只建立難度值 `0..4` 的來源與持久化，不是 AI 加成公式：

| 類別 | raw owner |
|---|---|
| 預設／新局選擇 | `sub_127E1 @ 0x127E1` 預設 0；`sub_CD435 @ 0xCD435` 讀寫新局暫存值 |
| 存檔 | `sub_10E2F @ 0x10E2F` 從 game settings block 的 byte 212 載入 |
| 熱座／主畫面 | `sub_628E2 @ 0x628E2`、`sub_1049B @ 0x1049B`、`main__0 @ 0x10057` |
| 舊多人同步 | `sub_F5A9F @ 0xF5A9F`、`sub_F74CD @ 0xF74CD`、`sub_F79F6 @ 0xF79F6`、`sub_FB7E5 @ 0xFB7E5` |
| 分數紀錄 | `sub_9F712 @ 0x9F712` 把難度寫入高分項目 |

多人函式只證實難度必須是鎖步／快照的一部分；舊傳輸與 Windows／網路 API 的內部行為不屬
remake 或玩法 RE 範圍。

## 玩家可見 consumer 分群

### 開局、種族、科技與領袖

- `sub_150FB @ 0x150FB`：難度 3／4 才依 difficulty、AI 種族與 raw table `0x180184`
  改動 AI 客製種族配置。
- `sub_589D6 @ 0x589D6`：AI personality 初始化使用
  `clamp(Random(10)+1-difficulty,1,10)`；難度 3 對三個 raw 權重各加 3，難度 4 各加 10。
- `sub_FC845 @ 0xFC845`：真人科技估值基準為 `25*difficulty²+50`，指定類別再乘 4。
- `sub_FD335 @ 0xFD335`：難度非 0 時啟用 AI application 選擇的額外候選處理。
- `sub_9781D @ 0x9781D`、`sub_97D59 @ 0x97D59`、`sub_C7ADA @ 0xC7ADA`：領袖招募／選擇／
  彈窗在低難度採不同門檻或提示路徑；完整招募公式仍以領袖回合鏈文件為準。

### 經濟、人口、殖民地與破產

- `sub_DE280 @ 0xDE280`：AI 的 food／industry／research 共用職務產出讀五級 raw table
  `byte_DD4D7[difficulty]`；它是在 `Colony_Job_Production_` 共用 subtotal 中加入，不是 Go 舊版
  任意浮點倍率。
- `sub_DEE1B @ 0xDEE1B`：AI 工業另加入
  `(base*difficulty+4)/8` 的有號整數式，位置在污染／建造 derived 欄位建立鏈內。
- `sub_E03F1 @ 0xE03F1`：AI 有機人口 BC 先把 `byte_DD4E6[difficulty]` 加入 quarter factor，
  再以整座有機人口聚合並按 `/4` 取整；這訂正舊文件把捨入點列為未知的敘述。
- `sub_E1839 @ 0xE1839`：AI owner 的 population growth 百分點直接加 `difficulty`。
- `sub_E2000 @ 0xE2000`：AI command deficit 每點維護係數為 `12-difficulty`；真人固定 10。
- `sub_D0B08 @ 0xD0B08`：AI 殖民地建築候選以 `score*(6-difficulty) < budget` 過濾；真人
  代入 4。`sub_D6ED4 @ 0xD6ED4` 在 Tutor 完全略過一段 queue 清理／重選。
- `sub_D6AD4 @ 0xD6AD4`：有額外農夫需求時，以難度作 `Random(10)` 門檻並向 AI cache 加 5。
- `sub_D5D99 @ 0xD5D99`、`sub_D5E19 @ 0xD5E19`：原版明名為 Freighter／Ship Upgrade
  cheats；分別依難度補貨運量與使用 `byte_D575C[difficulty]` 影響升級抽樣。
- `sub_ED908 @ 0xED908`、`sub_EDB35 @ 0xEDB35`、`sub_EDDF7 @ 0xEDDF7`：AI 破產出售研究
  建築、一般建築或艦艇後，除正常退款外另加 `difficulty` BC。

`sub_D2AEA @ 0xD2AEA`／`sub_D2CAE @ 0xD2CAE` 的 proximity／colony worth 在 Tutor 關閉、
Average 起完整啟用，是 AI 選星與殖民地估值的難度門；其完整 worth 公式仍由 AI 決策器列管理。

### 艦隊移動與攻擊決策

| raw owner | 難度效果 |
|---|---|
| `sub_D896F @ 0xD896F` | 新殖民船目的地：難度 `>=2` 放寬；難度 1 只對指定 personality 放寬 |
| `sub_DA99C @ 0xDA99C` | transport 行動：Tutor 關閉部分敵情路徑，Average 依 personality，Normal 起完整 |
| `sub_DAFD4 @ 0xDAFD4`、`sub_DB47E @ 0xDB47E` | Tutor 對未探索星目的與一般探索呼叫採特殊限制 |
| `sub_DBB9F @ 0xDBB9F` | opportunity attack 只在難度 `>1` 啟用 |
| `sub_CFAE5 @ 0xCFAE5` | mean ship strength 的 difficulty-adjusted 評估；原始指令站點已收錄，完整艦力欄由 AI fleet slice 管理 |

### 間諜與地面戰

- `sub_100D19 @ 0x100D19`：Tutor 直接不執行 AI 間諜配置；其餘難度 1..4 有不同 Agent／Spy
  配額與分配門檻。
- `sub_1014A4 @ 0x1014A4`：當攻方是 AI、守方是真人時，spy resolution 差值直接加
  `difficulty-2`。這是攻方向單一注入，不是「AI attack 與 defense 共同加一次」。舊
  `ai-spy-difficulty-audit-20260826.md` 的共同攻防強推論已被此直接 consumer 推翻。
- `sub_EC15C @ 0xEC15C`：一般 AI 地面加成為 `difficulty-2`；owner >= 8 的安塔蘭側為
  `2*difficulty-4`；真人為 0。
- `sub_ED260 @ 0xED260`：真人殖民地對 AI 舊主的 rebellion 權重另加
  `4*difficulty-8`。
- `sub_E87D2 @ 0xE87D2`：難度 `>=3` 時 AI 入侵另有先決條件成立的自動決策分支。
- `sub_ECF41 @ 0xECF41`：AI 佔領殖民地後擄獲科技的兩段嘗試量分別為
  `5*difficulty+115` 與 `5*difficulty+55`；真人固定 125／65。

### 隨機事件、星系與安塔蘭

- `sub_2230A @ 0x2230A`：事件排程 delta、Tutor 可選事件、事件強度、超新星與 stasis 倒數都
  分別讀難度；例如 stasis／部分倒數為 `Random(5)+10-difficulty`，不是共用倍率。
- `sub_2448F @ 0x2448F`：海盜活動強度按五級 switch 調整。
- `sub_22F5C @ 0x22F5C`：安塔蘭受害者候選對真人帝國依五級 switch 調整權重。
- `sub_8CB93 @ 0x8CB93`：星系特殊物數量／機率參數為低三難度 5、Hard 8、Impossible 12。
- `sub_5514C`、`sub_55161`、`sub_5542C`、`sub_55738`、`sub_55B12`、`sub_55F67`
  （`0x5514C..0x561C5`）：各級安塔蘭艦依難度加入 raw special 9／21／38；小型艦在 Hard 以上
  另加 raw 9。這些是 loadout 分支，不是 hull HP 倍率。
- `sub_60362 @ 0x60362`、`sub_60B59 @ 0x60B59`：Normal 起改用較強的飛彈／光束改造挑選分支。
- `sub_645EC @ 0x645EC`：每 25 回合資源 pulse，Hard／Impossible 分別採 150%／200%，其餘 100%。
- `sub_63FF0 @ 0x63FF0`：Hard／Impossible 第一次建出 Titan 後把各艦成本降到 90%。
- `sub_647D7 @ 0x647D7`：小型艦淘汰回合為低難度 100、Hard `12500/150=83`、Impossible
  `12500/200=62`。
- `sub_6481B @ 0x6481B`：中型艦淘汰回合為低難度 199、Hard `20000/150=133`、Impossible 100。

### 議會與外交

全部直接 consumer 為：

`sub_16021 @ 0x16021`、`sub_252A7 @ 0x252A7`、`sub_2552D @ 0x2552D`、
`sub_25C7C @ 0x25C7C`、`sub_25DF1 @ 0x25DF1`、`sub_2670A @ 0x2670A`、
`sub_26BBD @ 0x26BBD`、`sub_26FBA @ 0x26FBA`、`sub_2736E @ 0x2736E`、
`sub_277CF @ 0x277CF`、`sub_27A3D @ 0x27A3D`、`sub_4E3B5 @ 0x4E3B5`、
`sub_501CA @ 0x501CA`、`sub_51078 @ 0x51078`、`sub_53EDB @ 0x53EDB`、
`sub_544A1 @ 0x544A1`。

主要可見公式如下：

- AI 對 AI 條約嘗試週期使用 `250-40*difficulty`；部分提案檢定用 `difficulty+1`。
- 科技交換反應包含 `-50*difficulty+150`；good proposal score 包含 `+4*difficulty`。
- 和平門檻包含 `90-15*difficulty-20*warCount`。
- 一般宣戰候選含 `30*(difficulty²+1)`、`2*difficulty+1` 除數及 `+25*difficulty` 分數。
- Hard／Impossible 改變真人參與戰爭的正式狀態、treaty hatred 與 surrender fallback。
- `Change_Relations_` 只在 AI↔AI、百回合後且負變動時除以 `floor(difficulty/2)+1`。
- 議會在真人候選且難度 >2 時，以有號向零截斷計算 `6*score/(difficulty+6)`。
- sneak attack 與 unprovoked message 使用數個 `difficulty`、`difficulty+1`、`50*difficulty`
  的獨立機率／評分項；完整非難度欄位已分別留在外交與 AI 外交證據列。

這些站點證實難度會深入 AI 外交，而不是調整一個通用 relation multiplier。

## 證據分級與停止線

- **已證實**：125 個直接站點／78 個 owner 的完整集合、五級 raw 值的存讀同步，以及上述直接
  門檻與算術。JSON 中所有 owner 都保留完整原始函式與 callsite 導覽。
- **強推論**：少數 raw table 元素的玩家顯示名稱；程式只證實索引與消費位置。
- **未知但分列管理**：每個大型 AI owner 的其餘非難度評分欄、全域 PRNG 位元序列、1.50 profile
  差異。它們不再是「難度 consumer 漏查」，而是 AI／外交／事件等各自的子系統缺口。
- **排除**：`memset`、格式化、存檔／網路 API、編譯器 stack／exception helper 的內部行為。

## 對 remake 的影響

remake 已接 Growth、Food／Prod／Res、BC、Command Deficit、一般 AI ground bonus 與部分 AI
決策門，但 Spy 共同攻防注入已被原版直接 consumer 反證；RE gate 關閉後必須建立 READY spec，
改成只在「AI 攻擊真人」的 resolution 差值加入 `difficulty-2`。其餘缺口應回到對應 AI／外交／
事件／安塔蘭列，不再建立任何通用難度倍率。
