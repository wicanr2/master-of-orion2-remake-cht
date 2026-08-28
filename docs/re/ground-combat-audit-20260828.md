# 一般地面入侵、解算與戰後回寫稽核（2026-08-28）

## 問題與證據契約

舊文件已正確摘出 `Ground_Combat_Round_` 的 d100 比較，但 parity matrix 仍只列強推論；一般
入侵的運輸艦兵力、55-byte record、軍官、AI 決策、勝敗與殖民地回寫未形成同一條證據鏈，且
`Resolve_Invasion_Troops_` 曾被誤寫成相鄰的 `Resolve_Rebellion_Troops_`。本輪只閉合 RE 與
登記 remake 差異，不撰寫 READY spec、不修改玩法。

- 輸入：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`；IDA `.i64`
  SHA-256 `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- 工具：IDA Pro 9.4、`tools/ida/audit_ground_combat.py`；位址均為 IDA linear，
  DOS/4GW LE object #1。
- 外部導航符號：`symbols_fixed.tsv`，SHA-256
  `f83116049964c14072547dd087a6316117cf62808c328cc7c087984467030b28`；名稱不作證據。
- 可重生證據：[`evidence/ground-combat-ida-20260828.json`](evidence/ground-combat-ida-20260828.json)。
- `memset_`、`qmemcpy`、`sprintf_` 等 runtime helper 只作邊界，不進 RE／remake 分母。

## 玩家入口與 55-byte record

`Search_For_Battles_ → Do_1_Combat_ @ 0xE938C → Do_Attacker_Beat_Colony_Stuff_ @ 0xE87D2`
形成太空戰後的殖民地處理入口。真人攻方可由 colony selection popup 選轟炸、摧毀、入侵等
結果；AI 則經勝率預測、運輸艦 gate 與轟炸／入侵分支。`Invade_ @ 0xED48B..0xED59D` 以
`Get_Invasion_Info_ @ 0xEC831..0xEC97C` 建立雙方各 55 bytes 的記錄：

| offset | 長度 | 已證實用途 |
| ---: | ---: | --- |
| `+0x00` | 2 | 本方 owner |
| `+0x02` | 8 | 四類型攻擊力 `word[4]` |
| `+0x0A` | 8 | 四類型剩餘數 `word[4]` |
| `+0x12` | 4 | 四類型耐受命中數 `byte[4]` |
| `+0x16` | 1 | 目前類型，4 為全滅 |
| `+0x17` | 1 | 目前類型累積命中 |
| `+0x18/+0x19` | 1+1 | 本輪受傷／死亡類型，`0xFF` 為無 |
| `+0x1A` | 2 | colony index |
| `+0x1C` | 4 | 參戰 ship index list 指標 |
| `+0x20` | 2 | ship count |
| `+0x22` | 2 | 對方 owner |
| `+0x24` | 19 | 本方地面戰加成 block |

原版一般入侵攻方只掃 ship list 中 `ship+0x11 == 2` 的 type 2 transport，每艘把
**type 1 陸戰隊增加 4**；不建立攻方 type 0 裝甲。守方則由殖民地 `+0x132` 裝甲、`+0x130`
陸戰隊及 `Colony_N_Militia_` 民兵填 type 0／1／2。這由 `Get_Invasion_Info_`、
`Player_Owns_Transports_ @ 0xE8776`、`Unload_Transports_ @ 0xED6B7` 三條獨立 consumer 交叉證實，
不是單一反編譯器變數名推測。remake 允許艦隊另載戰車並作攻方 type 0，屬明確自製模型。

## 加成 block 與四兵種

`Compute_Player_Ground_Combat_Bonuses_ @ 0xEC15C` 產生 19-byte block；
`Compute_Ground_Combat_Info_ @ 0xEC3CE` 的共同攻擊基礎為：

```text
Anti-Grav + best armor + Personal Shield + best rifle
+ race ground modifier + Low-G(-10) + difficulty + officer
+ defending Subterranean(+10)
```

type 0 裝甲另加 `10 + Battleoids(+10)`，耐受另加 `1 + Battleoids(+1)`；type 1 陸戰隊另加
Powered Armor `+10/+1`；type 2 民兵攻擊 `-10`；type 3 叛軍攻擊 `-20` 且共同科技基礎取另一方
block。High-G 令所有類型基礎耐受由 1 變 2；Low-G 與 High-G 是互斥分支。真人難度項為 0，
AI 為 `difficulty-2`，owner>=8 為 `2*difficulty-4`。raw 科技 ID、三張 best-tech 表、race trait
offset 與 consumer 均已交叉閉合，舊「加成 block 尚未逐欄解出」不再成立。

軍官有兩條不同公式：

- 殖民地 Commando：一般 `2×(level+1)`，進階 `3×(level+1)`；
- 攻方艦隊 Commando：在所有參戰艦軍官中取最大值，一般 `5×(level+1)`，進階
  `floor((15×level+16)/2)`。

兩者只寫 block `+0x10` 的 word；`+0x12` 另保存實際 officer ID 供畫面。以上為**已證實**。

## 民兵、兵營生產與上限

`Colony_N_Militia_ @ 0xEC61E` 只計 race slot `<8` 且 prisoner bit `0x04` 未設的人口，再整除 5；
所以不能以總 `Population/5` 精確代表多種族／被征服殖民地。

`Produce_Ground_Military_ @ 0xE3616` 由每回合殖民地套用鏈呼叫：Marine Barracks raw 22 與
Armor Barracks raw 2 各有獨立五回合 counter；低於兵營上限時每五回合加一，高於行星上限時
每回合減一。Marine 上限取 `min(currentPopulation, planetCapacity)`，無 Warlord 再除 2；Armor
對應 `/2` 後，無 Warlord再除 2，即一般為四分之一；結果至少 1。`player+0x8BD` 已由獨立
trait／領袖 consumer 證實為 Warlord。這關閉舊文件的 raw flag 未知。

`Unload_Transports_` 在陸戰隊未達上限時逐艘消耗 type 2 transport、每艘增加四名陸戰隊，最後
更新帝國統計；再次證實原版的運輸單位是整艘 transport，而不是 remake 艦隊上的獨立 troop pool。

## 一輪解算、AI 預測與 RNG 重播

`Ground_Combat_Round_ @ 0xEC4FE..0xEC601` 每輪各做一次 `Random_(100)`：

```text
A = attackA[currentType] + Random(100)
B = attackB[currentType] + Random(100)
if A <= B: A 累積一次命中
if A >= B: B 累積一次命中
```

兩個條件獨立，平手雙方受傷。累積值**等於**耐受值時才減一單位；該類型歸零後前進至下一個
非零類型。`Resolve_Ground_Combat_ @ 0xEC601` 只是一個迴圈 wrapper，直到任一 currentType >=4；
一般 `Invade_`、叛亂與畫面動畫則直接逐輪呼叫同一 round helper。

`AI_Invasion_Will_Succeed_ @ 0xCF26E` 複製同一對 record 執行十次完整解算，至少七次攻方仍有兵
才判成功；若戰鬥狀態另有 incoming transports，先每艘加四名、再扣兩名作預測調整。這證明
`Resolve_Ground_Combat_` 的直接 caller 雖只有 AI predictor，算法仍由實際入侵與 UI 逐輪共用。

`Invade_` 先保存全域亂數 seed，再完成真實解算；真人畫面以同一開始 seed 重播逐輪動畫，畫面
結束後恢復已解算完的 ending seed，避免動畫把 PRNG 再消費一次。remake 使用自製「回合＋星系」
seed，沒有保存原版 save global seed，故只對齊分支公式，不具逐骰 parity。

## 戰後回寫

- **攻方全滅（包括同輪雙方全滅）**：`Invade_` 將守方剩餘 type 0／1 寫回殖民地裝甲／陸戰隊，
  並殺死參戰的 type 2 transports；殖民地 owner 不變。
- **攻方勝利**：先以 conquest-report 指標呼叫 `Change_Colony_Ownership_ @ 0xECF41`，因此一般
  入侵可走科技擄獲；其 `Change_Pop_Ownership_` 保留人口數但重標 owner／prisoner、處理領袖、
  首都、三棟統治建築、佇列與衍生 cache。接著呼叫正確的
  `Resolve_Invasion_Troops_ @ 0xECE05..0xECECA`：新殖民地先有一名陸戰隊，再依攻方剩餘陸戰隊
  與 transport 分組處理，最後夾至 `Colony_Infantry_Limit_`。相鄰的
  `Resolve_Rebellion_Troops_ @ 0xECECA` 只屬叛亂，不能再混稱入侵回寫。

原版沒有以守方存活地面單位覆蓋平民人口。remake 保留人口這一點方向正確，但現行運輸艦、
攻方戰車、transport 全滅／存活與勝利後陸戰隊回寫模型不同。

## 閉合結論與 remake 邊界

- **已證實**：玩家／AI入口、transport gate、55-byte record、四兵種、19-byte bonus、兩種軍官
  公式、兵營生產與上限、民兵、兩次 RNG、AI 7/10 預測、動畫 seed 重播、勝敗與 ownership 回寫。
- **已知 remake 偏差**：獨立 Marine／Tank pool 取代 type 2 transports；允許攻方裝甲；民兵以
  總人口除 5；軍官以帝國清單代理位置；自製 seed；失敗時 transport 與勝利後駐軍回寫不同。
- **仍未知但不阻止本列 RE 閉合**：原版 save global seed 的序列校準、1.50 可選 Commando 參數
  對 1.31 binary 的變更，以及 UI 每個 sprite frame 的逐幀 parity；後者屬視覺列。
- **停止線**：runtime memcpy／格式化、畫面平台 helper 與 Win95／DOS 服務不進完成分母。
