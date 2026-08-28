# 殖民地研究基礎與總產出稽核（2026-08-25；2026-08-28 重審）

## 問題與證據契約

2026-08-25 的初稿只沿 `DFE77 → DFF74` 判讀，因而錯稱研究建築只提供殖民地固定 RP；
2026-08-28 依 RE-first gate 重新匯出直接上游 `DFDC6`，證實同一批建築另有每位科學家的加成。
本文件保留可回查的錯誤成因，但只陳述重審後的現行結論。

- 輸入：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- 工具：IDA Pro 9.4、`tools/ida/audit_colony_turn_chain.py`；位址均為 IDA linear，
  DOS/4GW LE object #1。
- 外部導航符號：`symbols_fixed.tsv`，SHA-256
  `f83116049964c14072547dd087a6316117cf62808c328cc7c087984467030b28`；名稱不作證據。
- 可重生證據：
  [`evidence/colony-turn-chain-ida-20260828.json`](evidence/colony-turn-chain-ida-20260828.json)。
- `memset_`、`sprintf_`、`Format_Quotient_` 與字串複製只服務 cache／報表；依 runtime
  停止線保留 callsite，不納入玩法分母。

## owner 每科學家環境基礎：`Colony_Research_Per_Scientist_ @ 0xDFDC6..0xDFE77`

函式從全域基準 3 開始，依固定順序加入殖民地建築與 planet special：

| raw operand | 原版 ID／語意 | 每科學家 RP |
| --- | ---: | ---: |
| 基準立即數 `0xDFDC7` | 銀河基準 | +3 |
| `colony+0x13A` | raw 4 Astro University | +1 |
| `colony+0x154` | raw 30 Planetary Supercomputer | +2 |
| `colony+0x149` | raw 19 Galactic Cybernet | +3 |
| `colony+0x159` | raw 35 Research Laboratory | +1 |
| `planet+0x0F == 10` | Ancient Artifacts | +2 |
| `planet+0x0F == 11` | Orion special | +5 |

building raw ID 由 building base `+0x136` 與獨立 `OrigBuildingID` 對映閉合；planet special
raw 10／11 則由 12 項 typed enum、生成／發現 consumer 與手冊交叉驗證。Ancient Artifacts
因此使一般基準 3 變成 5；Orion special 是在相同 accumulator 上另加 5，不可誤寫成取代值。

`Pre_Import_Computing_ @ 0xE1D97..0xE1DA7` 以 owner 呼叫此函式，把 low byte 寫入
`colony+0xDF`。所以 `+0xDF` 是 owner 的環境／建築每科學家快取，尚未加入逐 race research
trait、Heightened Intelligence、重力、士氣、政體或領袖修正。

## 混合人口基礎：`sub_DFE77 @ 0xDFE77..0xDFF74`

`DE280(job 2)` 透過 dispatch 對每名科學家呼叫 `DFE77`：

- owner 且不要求 breakdown 時可直接使用 `colony+0xDF`；外族 slot 會重跑 `DFDC6`。
- 一般 slot `<8` 加 `player[slot]+0x8A3` 的研究 trait raw 值。
- 只有 race slot 等於 colony owner 時，若 `player+0x16B == 3` 再加 1。
  科技狀態陣列基址是 `player+0x117`，`0x16B-0x117=84`，與受版控
  `TECH_HEIGHTENED_INTELLIGENCE = 84` 及研究 application status `3` 共同閉合。
- Android slot 8 固定加 3；Natives slot 9 不加研究。
- 十格 cache 以 colony、owner slot、race slot 為鍵；`memset_` 只重設 cache sentinel。

這個逐人口基礎再由已閉合的 `DE280` 同時套重力、prisoner、士氣、封建／民主系研究修正、
Science Leader 與 AI 難度。`DE280` 最後以 `(rawSum+10)/20` 取得整數研究人口產出。

## 研究總產出：`Colony_Research_Production_ @ 0xDFF74..0xE03C3`

normal caller 是 `Pre_Import_Computing_ @ 0xE1DFF`。人口 `colony+0x0A == 0` 或
`Event_Check_Space_Anomaly_ @ 0x2341E` 命中時，`colony+0xEB=0`。正常路徑先呼叫
`DE280(colony, job=2)`，再加入四棟建築的**殖民地固定 RP**：

| Colony offset | 原版 building ID | 建築 | 固定 RP |
| --- | ---: | --- | ---: |
| `+0x159` | 35 | Research Laboratory | +5 |
| `+0x154` | 30 | Planetary Supercomputer | +10 |
| `+0x149` | 19 | Galactic Cybernet | +15 |
| `+0x13C` | 6 | Autolab | +30 |

```text
colony+0xEB = DE280(job=2)
              + 5×[Research Laboratory]
              + 10×[Planetary Supercomputer]
              + 15×[Galactic Cybernet]
              + 30×[Autolab]
```

Research Laboratory、Planetary Supercomputer、Galactic Cybernet 同時出現在 `DFDC6` 與
`DFF74`，所以它們確實同時提供逐科學家 `+1／+2／+3` 與固定 `+5／+10／+15`；這不是重複
計算。Astro University 只提供每科學家 `+1`，Autolab 只提供固定 `+30`。

## 勘誤與 remake 邊界

- **推翻舊斷言**：2026-08-25 初稿稱四個研究建築旗標只由 `DFF74` 讀取，並據此把 remake
  的逐科學家加成判作重複。`DFDC6 @ 0xDFDD6..0xDFE35` 的四個直接 building consumer 足以
  否定該結論；錯因是切片漏掉 `DFE77` 的直接上游，而不是原版語意模糊。
- **已證實**：owner 快取 `+0xDF`、四個逐科學家建築加成、兩種 planet special、逐 race
  trait、Heightened Intelligence、Android／Natives、空殖民地／anomaly gate、四個固定建築
  加成及 `+0xEB` writer。
- optional 598-byte breakdown 的逐格 UI 文案索引未全數命名，但不改變研究總產出，不列為
  玩法公式缺口。
- remake 是否已同時、且僅一次接上上述兩層效果，須在 RE gate 關閉後依 READY spec 與正常
  玩家路徑驗證；既有綠色測試不能反向證明原版公式。
