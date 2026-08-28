# 殖民地每工人工業公式稽核（2026-08-28）

## 證據契約

- 輸入：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- 工具：IDA Pro 9.4、`tools/ida/audit_colony_turn_chain.py`；位址均為 IDA linear，
  DOS/4GW LE object #1。
- 外部符號導航：`symbols_fixed.tsv`，SHA-256
  `f83116049964c14072547dd087a6316117cf62808c328cc7c087984467030b28`。
  名稱只供導航，結論以原始指令、資料表、caller 與既有 building／technology ID 證據為準。
- 可重生證據：
  [`evidence/colony-turn-chain-ida-20260828.json`](evidence/colony-turn-chain-ida-20260828.json)。
- `memset_` 等 C runtime helper 只保留 callsite；其內部行為不納入玩法 RE 或 remake 範圍。

## 已證實公式

`sub_DEC95 @ 0xDEC95..0xDED47`（外部導航名
`Colony_Industry_Per_Worker_`）先以 `colony+0x02` 取得行星，再以行星 `+0x0A`
的礦產 class 索引 `byte_DD4B5 @ 0xDD4B5`。該表的原始 bytes 為
`01 02 03 05 08`，依既有礦產 ordinal 對應：

| 礦產 | 每工人基礎工業 |
| --- | ---: |
| Ultra Poor | 1 |
| Poor | 2 |
| Abundant | 3 |
| Rich | 5 |
| Ultra Rich | 8 |

之後依殖民地 building flag 與擁有者科技逐項相加：

| 原始欄位 | raw ID／條件 | 效果 |
| --- | --- | ---: |
| `colony+0x13A` | building 4，Astro University | +1 |
| `colony+0x13D` | building 7，Automated Factory | +1 |
| `colony+0x15A` | building 36，Robo Mining Plant | +2 |
| `colony+0x142` | building 12，Deep Core Mine | +3 |
| `player(owner)+0x183 == 3` | technology 108，Microlite Construction 已取得 | +1 |

building record base `colony+0x136`、上述 raw ID 與名稱的對應，另由
[`ai-colony-build-selection-audit-20260826.md`](ai-colony-build-selection-audit-20260826.md)
及 `OrigBuildingID` 資料交叉支持；不是只靠相對位址猜名。科技 status base 為
`player+0x117`，所以 `+0x183` 對應 topic 108；typed enum 亦為
`TECH_MICROLITE_CONSTRUCTION`。公式可寫成：

```text
IndustryPerWorker = mineralBase[mineralClass]
                  + AstroUniversity
                  + AutomatedFactory
                  + 2 * RoboMiningPlant
                  + 3 * DeepCoreMine
                  + MicroliteConstruction
```

各旗標與科技項成立時取 1，否則取 0。正常 `Pre_Import_Computing_ @ 0xE1D59`
呼叫時不要求 breakdown，回傳值低 byte 寫入 `colony+0xDE`；若傳入非零 breakdown
指標，原函式會分別寫出基礎值及各加成，該路徑支持上述加總拆項。

## 與混合種族產出的下游鏈

`Colony_Empire_Base_Industry_Produced_ @ 0xDED47..0xDEE1B` 接受 colony、owner slot、
colonist slot 與可選 breakdown。colonist slot 等於 owner 且未要求 breakdown 時，直接使用
`colony+0xDE` 快取；其餘情況重呼叫 `sub_DEC95`，再加 colonist-specific modifier：

- 一般 player slot `< 8`：signed `player[slot]+0x8A2`；
- Android slot 8：固定 `+3`；
- Natives slot 9：`+0`。

這與 [`mixed-race-colonist-production-audit-20260825.md`](mixed-race-colonist-production-audit-20260825.md)
的 packed colonist slot、重力與 trait 證據閉合成完整每工人鏈。函式內用 `memset_` 清除
十格暫存 cache 只是實作細節，不是需要在 remake 重製的 gameplay。

## 證據等級與完成邊界

- **已證實**：五級礦產表、四座建築加成、Microlite Construction 加成、owner 快取欄位、
  非 owner／Android／Natives 的下游 modifier 與兩個 caller。
- **強推論**：`colony+0xDE` 的高層名稱採既有結構語意；原始欄位、writer 與 consumer
  已證實，但名稱本身不取代位址證據。
- **未涵蓋**：總工業產出的工人人數、重力、士氣、污染、維護、封鎖與貨運等後續公式。
  它們屬其他 pre-import producer／consumer，不能因本切片閉合而一併宣稱精確。
- **remake gate**：本輪只補 RE 知識庫，不修改 Go。既有 typed 欄位能表達這些加成，仍須等
  玩家玩法 RE 分母閉合後，才依規格檢查同狀態 phase ordering 與兩條實際消費路徑。
