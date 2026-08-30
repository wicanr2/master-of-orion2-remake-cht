# 逆向研究紀錄

## 目標

- 遊戲／版本：`Master of Orion II`，1.31 與 patch 1.5。
- 平台：DOS；重製執行環境為 Linux amd64、Go／Ebiten。
- 原始輸入：私有正版資料與存檔位於 `/home/anr2/moo2-private-build/`，只讀掛載，禁止進入儲存庫。
- 可重生工具：`moo2-ebiten:latest`、Go `go1.25.12 linux/amd64`、`git 2.39.5`、Docker + Xvfb。
- 可寫研究區：儲存庫文件與測試；一次性解出或分析產物使用容器內 `/tmp`。

## 研究紀律

每一項規則分成「已證實、強推論、假說、未知」，並保留原始檔名、位址／偏移、工具與位址基準。
執行檔反組譯立即數優先於手冊，手冊優先於 `openorion2`；單元測試只能證明重製內部一致，
不能單獨證明原版一致。詳見 [`CONTEXT.md`](../CONTEXT.md) 與 [`docs/re/01-gap-report.md`](re/01-gap-report.md)。

## 本輪問題

「在暫停新玩法實作後，如何先讓 parity matrix 的每個玩家可見系統具備 raw 位址、輸入狀態、
規則／資料表、玩家 consumer、證據等級與未知邊界，並排除 compiler helper、C runtime 與
Windows API 內部實作？」間諜、艦隊移動、殖民、五級難度、戰略轟炸與 AI 外交切片已閉合；
AI 外交目前另有 `+0x60F..+0x88F` 的固定逐欄契約，分開 ordered-pair、每帝國純量、pending
payload 與只有初始化／清除端的未知欄。後續切片一律由 parity matrix 尚未閉合的玩家玩法列
選取，不再因看到單一 raw offset 就直接開 spec 或修改 Go。

## 已知主張與證據

| 主張 | 來源／證據 | 狀態 | 重製對映 |
|---|---|---|---|
| ECCM 使飛彈干擾機率減半 | `GAME_MANUAL.pdf` p.115–116、p.123；`gamedata.MissileJamChance` | 已證實 | `ResolveMissileShotWithMods` |
| MIRV 為四枚彈頭，AMR 只摧毀一枚 | 手冊 p.115–116、既有 AMR 說明；`MissileDefenses` 逐彈頭骰 | 已證實 | `WeaponModMissileWarheadCount` + shell 解算 |
| EMG 先過盾再直接傷結構 | 手冊「direct structure」與既有過盾／過甲管線相接 | 強推論 | `WeaponModMissileArmorPiercing` |
| 魚雷 OVR +50%、ENV 四倍 | 手冊 p.115–116；魚雷分類目前只有「質子魚雷」 | 已證實（適用性有模型邊界） | `WeaponModMissileDamageMultiplier` |
| ARM／FST 的完整效果 | FST 的速度 +4 已由 `sub_3CD21` 證實；同一分支另確認 raw `0x12/0x13`=20、`0x14`=24 且不加 FTL，`sub_3DFE0` 有速度乘 5 的 OCV-like 下游。`sub_3E095` 的 Dcv 分支與 `sub_3A0B9` 的攔截 quotient/remainder 鏈已追回；ARM 對 raw high-byte bit `0x08` 仍是強推論，完整原版攔截器順序與所有映射仍未知 | FST 速度分支已證實；ARM raw／Dcv 消費鏈強推論；原版全路徑未知 | remake 已提供 ARM/FST 設計選項、成本／存檔與快速／格子戰術的 PD 垂直切片；不宣稱原版全路徑完成；詳見 [`docs/re/weapon-mod-flags.md`](re/weapon-mod-flags.md) |
| 魚雷 NR 的射程衰減 | 手冊有規則，但目前飛彈傷害為固定值、沒有該消費端 | 未知 | 不進實際傷害公式 |
| 原版逐艦保存軍官 | `internal/save/entities.go:436-473` 的 `save.Ship.Officer int16`；`openorion2/src/gamestate.cpp:2365-2405` 只在 `sptr->officer >= 0` 時疊 Weaponry／Helmsman | 已證實（結構／次級原碼） | `Ship.OfficerName` + `officer_assignment.go` |
| 軍官改派時一人只能在一艘船上 | 原版 `Ship.Officer` 是逐艦單值；remake 的唯一性是依該結構的強推論 | 強推論 | `AssignOfficerToShip` 先清除全帝國舊指派 |
| `.GAM` 數字 `Officer` ID 與重製領袖 ID | `gamestate.h:27,1277`、`gamestate.cpp:1724-1725,1870-1871,2372-2401`；IDA `sub_10E2F @ 0x10E2F`／`sub_1160B @ 0x1160B` 直接讀寫 `dword_1930DC` 的 `0x3B×0x43`；完整雜湊與限制見 `docs/re/officer-ids.md` | 已證實（原版索引與全局 `.GAM` 讀寫）；HERODATA 對應為強推論 | `herodata.Leader.ID` → `shell.Leader.ID` → `Ship.OfficerID`；JSON 仍保留 `OfficerName` 作舊格式回退，重製原生 importer 尚未實作 |
| 導航／工程師／戰鬥加成改讀逐艦軍官 | 手冊「assigned officer」語意、既有 engine officer wrapper 與 `CombatShip` 消費端 | 已接（重製內部已驗證） | 航行、修復、快速結算、格子戰術 |
| `HIRE` 可挑指定待僱傭兵 | 手冊軍官管理說明（`help.tsv`：HIRE 模式與待僱佇列）；既有 `MercPool`／`MercHireCost` | 已接（重製內部已驗證） | `HireMercAt` + 軍官畫面 HIRE 模式；未宣稱原版候選排序完全一致 |
| `POOL` 解除艦艇任職 | 手冊軍官管理說明（`help.tsv`：moves the officer into the Officer Pool）；remake 以空 `Ship.OfficerName` 表示未指派 | 已接（重製設計／內部已驗證） | `ReturnShipOfficerToPool`；沒有虛構原版旅行計時 |
| `DISMISS` 解雇艦艇軍官 | 手冊軍官管理說明（`help.tsv`：officers may be dismissed）；`save.Ship.Officer` 逐艦欄位 | 已接（艦艇軍官範圍） | `DismissShipOfficer` 清除指派並移出 `Leaders`；殖民地領袖因缺少反向加成資料仍拒絕 |
| 原版每個武器保存火線角 | `internal/save/entities.go:373-390` 的 `ShipWeapon.Arc`；`openorion2/src/gamestate.h:676-683` 與 `gamestate.cpp:759-781` 的原始值／顯示索引 | 已證實（結構／次級原碼） | `Ship.Arc`，目前單武器重製模型保存一個弧 |
| Fwd Ext／Back Ext／360 的設計代價 | `docs/knowledge-base/manual-cht/03-combat.md:295-304`，手冊 p.127–128 | 已證實（手冊） | `WeaponArcCostPercent`、`WeaponArcAdjustedValue`；+25%／+25%／+50% |
| 火線角如何限制目前格子戰術的合法射擊方向 | 原版 `Relative_Bearing @ 0x32AD1`、`Relative_Bearing_XY @ 0x32A20`、`Move_Ship @ 0x3F5F1`、`Ship_Can_Deploy_At @ 0x49043`；完整位址／雜湊見 `docs/re/weapon-arcs.md` | 已證實（格子戰術方向鏈）；快速路徑不消費射界為強推論 | `CombatShip.Facing`、原始 bearing mask 與格子戰術玩家／敵方射擊已接；`battleVolley` 保留原版抽象模型，詳見快速 call graph 章節 |
| 戰機從目標最弱護盾面攻擊 | `GAME_MANUAL.pdf` p.157；重製已有 `CombatShip.ShieldFacingHP` 四面容量 | 已證實（手冊＋既有資料模型） | `FighterDamageAtWeakestShield` 選最低容量分面並扣除護盾，戰機武器傷害仍標近似 |

### 2026-08-30 匿蹤裝置與 Telepathic 下游兩輪

| 主張 | 來源／證據 | 證據等級 | 重製對映 |
|---|---|---|---|
| raw 6／23 使用 combat ship `+0x40` 狀態機，Phasing 另有 `+0x41=10` | `Init_Special_Devices_ @ 0x4C9F6`、`Fire_Ship_ @ 0x38B5E`、`Init_Ship_For_Start_Of_Turn_ @ 0x42B70`、`Do_Combat_Turn_ @ 0x42F7F` | 已證實 | `internal/shell/cloak.go` 已有同等玩家規則；本輪未改 Go |
| Phasing `state=4` 不可選取／改鎖／防禦射擊；Cloaking `state=1` 提供 +80 與飛彈 50% miss | `Select_Ship_To_Target_ @ 0x2A46A`、`Retarget_Missile_ @ 0x3DDD8`、`Ship_Specials_Defensive_Bonus_ @ 0x36A63`、`Resolve_Missile_ @ 0x3D2DF` | 已證實 | 既有格子規則已接；快速結算仍維持分離模型 |
| raw 31 不進 cloak 戰術狀態機 | 全庫 `+0x40／+0x41` direct-operand census、`Init_Special_Devices_`；`cloak-state-offsets-ida-20260830.json` | 已證實資料流邊界 | 只保留戰略 concealment，不虛構格子效果 |
| Telepathic action type 2 進俘獲解算，且跳過俘獲後 `+0xB0` 戰鬥失能旗標 | `Resolve_Ai_Boarding_ @ 0x2BF73`、`Resolve_Capture_ @ 0x37DA8`、`Capture_Ship_ @ 0x38312` 與 38 個 0x139-stride consumer | 已證實 | 現行 Telepathic 擄獲能力需依此契約抽測，不把 type 2 寫成無條件成功 |
| Telepathic 敵殖民地 worth 權重位移為 `(+1,-1)`，最終雙表加權除 6 | `Enemy_Colony_Worth_To_Player_ @ 0xD8D11..0xD8DE0` | 已證實（raw 表欄名未知） | 尚不因未知欄名新增猜測性實作 |
| Telepathic 對 raw tech 5 Alien Management Center 把估值設為 1 | `Calc_Tech_Value_ @ 0xFC845` 的入口 `edx→var_4` 與 `0xFCDE9..0xFCE46`；受版控 enum raw 5 對照 | 已證實 | AI／外交科技估值證據，不是額外玩家能力 |
| 心控合法性的兩次相同將領 helper 沒有可見副作用 | `Player_Has_Leader_With_General_Skill_At_Star_ @ 0xC6052` 與唯讀 `Officer_Has_General_Skill_ @ 0x9467D` | 已證實 | remake 可合併成一次查詢 |

### 2026-08-09 外交現金餽贈切片

| 主張 | 來源／證據 | 證據等級 | 重製對映 |
|---|---|---|---|
| 指定金額現金是外交餽贈選項 | `help.tsv` 第 600 筆；`Get_Human_Offer_Gift_Message @0x186D3` 的 `a3=5` 分支；`Diplomacy_Offer_Money @0x1D565` | 強推論（介面／raw 分支互相支持） | `GameSession.OfferCashGift`；畫面暫提供 10 BC |
| 現金交易的國庫方向 | `An @0x1DEF8` case 13；`.asm` 同段對 `word_19A192` 的玩家減額／對手加額 | 已證實 | `s.Player.BC -= amount`、`ai.Player.BC += amount`，不足時失敗即關閉（fail-closed） |
| 現金餽贈改善外交評估 | **2026-08-28 勘誤：** `sub_539D9` 唯一 caller 是 `Diplomacy_Demand_`，不屬 cash gift；此舊列不能證明餽贈接受或關係公式 | 原版公式未知；資源轉移方向已證實 | remake 的 `Relation +5` 仍是強推論正規化；詳見 [`docs/re/diplomacy-gifts.md`](re/diplomacy-gifts.md) |

### 2026-08-09 戰機主要目標與 FST runtime 下游切片

| 主張 | 來源／證據 | 證據等級 | 重製對映 |
|---|---|---|---|
| 戰機主要目標仍有效時不應每回合改追最近艦 | `GAME_MANUAL.pdf` p.157 明定只有主要目標不再有效且仍有射擊時才自動重選；原版 `Orion2.exe.asm` 的 `sub_2A46A`（raw `0x2A46A`）在目標評估掃描中以 `sub_3E563`（raw `0x3E563`）排除戰機記錄，再進 `sub_2A239`。候選優先分數本身仍有未定參數。 | 已證實（行為／排除端）；候選排序未知 | `FighterSquadron.TargetName` 保存有效主要目標；`fighterTarget` 只在失效時重選，測試鎖定「維持／重選」兩個邊界。 |
| FST raw flag 會進入飛彈 runtime 速度 | 同一份 `Orion2.exe.asm`：`sub_3C892`（raw `0x3C892`）把設計槽 `+0x56` 複製到 runtime `+0x11`，呼叫 `sub_3CD21`（raw `0x3CD21`）；後者 flag `0x10` 執行 `+4`，另有 `0x14` 的 24 速／不加 FTL 分支，結果寫入 runtime `+0x16`／`+0x13`。`sub_3DFE0`（raw `0x3DFE0`）再使用該結果並乘 5，且由 `0x2B7CC`／`0x3AD57` 呼叫。 | 已證實（欄位／呼叫／算術）；完整攔截器消費仍未知 | remake 以 `MissileBeamDefenseOf` 接入標準飛彈 PD 垂直切片；詳見 [`docs/re/weapon-mod-flags.md`](re/weapon-mod-flags.md)。 |

### 2026-08-09 戰機接戰前 PD 切片

| 主張 | 來源／證據 | 證據等級 | 重製對映 |
|---|---|---|---|
| 近迫防禦要在戰機接戰前先開火 | `assets/i18n/help.json` 第 402 筆（原版 help 編號 #402）明定：PD 尚未開火時，會在飛彈命中艦艇或戰機接戰前開火。 | 已證實（手冊／help） | `cmd/moo2/tacticalfighter.go:advanceSquadrons` 在 `StepToward` 到達接戰距離後、戰機攻擊前呼叫 `ResolvePointDefenseFighterShot`，並共用 `CombatShip.PointDefenseSpent`。 |
| 戰機的 Beam Defense 公式 | `docs/knowledge-base/manual-cht/03-combat.md` p.157；`gamedata.CombatFighterBeamDefense` 已保存 `5*Speed + racial Ship Defense + Fighter Pilot + Helmsman`。 | 已證實（公式）；Helmsman 原版接線仍未知 | `CombatShip`／`FighterSquadron` 帶入參戰種族 Ship Defense 與 `SKILL_FIGHTER_PILOT` 最大值；Helmsman 明示傳 0。 |
| PD／戰機的逐彈命中／消費完整等價 | raw `sub_3AD57 @0x3AD57` 只證實 fighter／beam 分支與 `Weapon_In_Range` 呼叫；相鄰 `sub_3AC20 @0x3AC20` 的直接傷害式已追回，但完整 raw record 參數、敵方戰機政策與 runtime 邊界尚未追回。 | 未知 | 只接玩家格子戰術的順序、共用 PD 武器改造與兩條戰機受傷消費；測試證明重製內部閉環，不升格為原版全路徑。 |

## 重製結果

- `Ship.Mods` 經既有 JSON 存檔保存；設計成本／佔格、快速結算與格子戰術共用武器適用性過濾。
- 舊 `ResolveMissileShot` 入口保留，無改造呼叫行為不變；新入口為 `ResolveMissileShotWithMods`。
- UI 依武器類型顯示已接線改造，避免選到無消費端的代碼。
- 軍官管理切片：HIRE 由玩家點選待僱候選，POOL 解除艦艇任職，DISMISS 移除艦艇軍官並清理 `Ship.OfficerName`／`Ship.OfficerID`；殖民地領袖沒有誤用艦艇頁解雇流程。原版 `_leaders[]` 來源序號已另存；同一輸入 IDA 已證實整個 `0x3B×0x43` 全局區塊的 `.GAM` 讀寫，重製原生 importer 仍未實作。
- 火線角切片：設計畫面可循環選擇；成本／佔格、建造判定、JSON 與戰鬥資料共用同一個 `Ship.Arc`；格子戰術再以 `CombatShip.Facing` 套用原版 bearing mask；原版快速 call graph 不消費射界，`battleVolley` 維持抽象路徑。
- 戰術戰機切片：玩家中隊已接最弱護盾分面、「主要目標有效時維持、失效才重選」與戰機接戰前 PD 順序；敵方五級 blueprint raw writer／非空武器槽已由同一輸入 IDA 證實，敵方戰機下游命中／傷害仍未知。

## 重現命令

```sh
docker run --rm --network bridge --memory 2g --cpus 2 --pids-limit 512 \
  -u 1000:1000 -v /home/anr2/moo2:/workspace:ro moo2-ebiten:latest \
  sh -lc 'cd /workspace; export PATH=/usr/local/go/bin:$PATH; \
  export GOMODCACHE=/tmp/moo2-gomod GOPATH=/tmp/moo2-gopath GOCACHE=/tmp/moo2-gocache; \
  mkdir -p "$GOMODCACHE" "$GOPATH" "$GOCACHE"; \
  xvfb-run -a -s "-screen 0 1024x768x24" go test -buildvcs=false ./...'
```

## 已更新的過期參考

- [x] `WORKLIST.md`
- [x] `docs/tech/weapon-mods.md`
- [x] `docs/knowledge-base/manual-cht/03-combat.md`
- [x] `docs/re/01-gap-report.md` 第 102 項
- [x] 程式碼註解與測試
- [x] 艦隊 → LEADERS → 軍官列 → 指派／改派／解除；JSON／熱座席位保存
- [x] 艦艇設計 → 火線角循環選擇 → 成本／佔格／建造判定 → JSON／戰鬥資料傳遞 → 格子戰術方向合法性；快速結算依原版 call graph 維持抽象，不新增空間射界

## 2026-08-09 間諜 HIDE 任務切片

| 主張 | 來源／證據 | 狀態 | 重製對映 |
|---|---|---|---|
| HIDE 會讓攻擊方在 Spy vs Spy 得到 +20 | `internal/gamedata/spy.go` 的 `SpyVsSpyAttackerBonus`；手冊 `MANUAL_150.html`「Spy vs Spy」 | 已證實 | `spyMissionAttempt(..., SpyMissionHide, ...)` 傳 `hide=true` |
| HIDE 不同時執行 STEAL | `GAME_MANUAL.pdf` p.174-175 將 Espionage、Sabotage、Hide 以互斥 `or` 描述 | 強推論 | HIDE 跳過偷科技擲骰，只保留 Spy vs Spy |
| 間諜任務逐對手保存 | 原版 `gamestate.h` 的每對手 `spies[]` 任務位元；RACES 任務鈕資料表 | 已證實（資料形狀）；remake UI 控制為重製設計 | `PlayerSpies` + `PlayerSpyMissions`、JSON、熱座席位 |
| 原版三顆任務鈕的左右順序 | `Orion2.exe.asm` 只證實 x 偏移 0/+76/+149，未追到各自任務碼 | 未知 | 不硬套原版位置；RACES 提供明確標籤的 STEAL/SABOTAGE/HIDE 循環 |
| HIDE 擊殺防守 Agent 的實際扣除 | 手冊有「kill enemy agents」；本輪已補上 remake／AI 的獨立防守 Agent 數量與 `advanceEspionage` 回寫 | remake 已完成；原版 raw 消費細節未知 | Spy-vs-Spy 擊殺防守方時實際扣 1 個 Agent，不誤扣 AI 進攻 Spy |

## 2026-08-09 SABOTAGE 建築破壞切片（反組譯勘誤）

| 主張 | 證據 | 證據等級 | 重製對映 |
|---|---|---|---|
| 任務碼 `1` 是 SABOTAGE，成功門檻為 70 | 原版執行檔 `N_Spies_Bonus` raw `0x1014A4`：mission `0` 走 `Allocate_AI_Spies`，mission `1` 比較 `0x46` 後呼叫 `Steal_App`，mission `2` 只進共用 Spy-vs-Spy 加成；輸入 SHA256 `7AE2AC2E5904CA330009AF2827279D889906B0B9B7A8854C38EB707A56E955B5` | 已證實 | `SpyMissionSabotage`、`gamedata.SpyThresholdSabotage=70` |
| SABOTAGE 目標來自殖民地已建建築槽 | 原版 `Steal_App` raw `0x10130A`：逐殖民地、49 槽讀 `colony + 0x136` 旗標，排除原始條件 `v6 != 9` 不成立的槽 | 已證實 | `spySabotageCandidates` 只收 `ColonyBuildings` 中能對回 `OrigBuildingID` 的建築，未知 key 保守略過 |
| 目標依建造成本加權 | 同一函式讀 `dword_17EB45`，每 19-byte 建築表列的 `+8` 成本累加到抽選總量；`internal/gamedata/buildings.go` 的 `ProductionCost` 已由 `off_17EB3D + 8` 對回 | 已證實（資料來源）；抽選邊界映射為強推論 | 候選按殖民地／原版 ID 排序後以 `Intn(totalWeight)` 抽選 |
| 成功會清除選定建築 | `Steal_App` 呼叫 `Add_Building` raw `0x145EA`；該函式對選定殖民地的 `+0x136 + buildingID` 寫入 `0` | 已證實 | 從防守方 `AIOpponent.ColonyBuildings` 刪除選定中文建築名 |
| SABOTAGE 與 STEAL／HIDE 共用間諜互殺結算 | `0x1014A4` 的 mission 分支在行動後回到共用區塊；手冊 `GAME_MANUAL.pdf` p.174-175 也把三種任務列為同一個 mission 的互斥選項 | 強推論 | 玩家 → AI 使用 `spyMissionAttemptWithBuildings`；命中後仍跑 `resolveSpyVsSpy` |
| remake 的中文 map 與原版 49 槽完全一對一 | 原版存檔／建築表與現有 `OrigBuildingID` 對照；特殊／未知槽尚未全部命名 | 未知 | 不把未知 key 當成可破壞候選，避免把推測升格成效果 |
| 原版 `toggle_flag` 的亂數邊界與完整分數來源 | `0x10130A`／`0x1014A4` 的資料流仍有未解析欄位；反組譯只足以確認權重與 70 門檻 | 未知 | 沿用手冊 `SpyRollChance` 封閉解；以 `Intn(totalWeight)` 做可重現近似，文件與註解明示等級 |

### 2026-08-11 SABOTAGE raw score／領袖 callback／CMBTSHP timer 勘誤

這段是對上面歷史快照的追加勘誤，不覆寫舊推導。輸入仍為 `Orion2.exe.i64` SHA-256
`4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`，工具為 IDA Pro 9.4，
位址為 IDA 線性位址，探針為 `tools/ida/consumer_closure.idc`。

| 新結論 | IDA 證據 | 等級 | remake 對映 |
|---|---|---|---|
| raw slot helper 已追回 | `sub_101483 @ 0x101483`：`n<=5 → 2n`、`6..10 → n+5`、`n>10 → floor((n-10)/2)+15` | 已證實 | `gamedata.OriginalSpyScoreHelper`；`SpySlotBonus` 夾限後直接使用 |
| relationship byte 的分層 | `sub_1026CF @ 0x1026CF` 取 `record+0xE57+other` 的低 `0x3F`；`sub_1026F1 @ 0x1026F1` 取同 byte `>>6`；`sub_10278D @ 0x10278D` 保留 `0xC0` 再寫低 6 位 | 已證實 raw 位元；高 2 位語意未知 | `OriginalSpyRelationshipCount/Mode` 只保留 raw fixture，不改 JSON schema |
| SABOTAGE／Spy-vs-Spy raw score 形狀 | `sub_1014A4 @ 0x1014A4` 使用 helper、`Random(100)` 兩次、`word_1ACE78`／`word_1ACE7A`、raw context 修正，並分支 60／70／80／90／±80；兩表 raw bytes 為 `0xFF` 初始化 | 已證實算術形狀；上游填值／欄位語意未知 | `spyMissionScore` 以可保存 AB／DB／E 與 `T=70` 近似完成，不宣稱 raw parity |
| ETA callback 並非單一 UI call | `sub_E2AB1 @ 0xE2AB1` 掃六個 `+0x48..+0x52` 槽，命中時呼叫 `sub_E1D59`／`sub_DF8F0`，尾端呼叫 `sub_E2710`；後者寫回多個 empire derived 欄位 | 已證實 callback 鏈；raw 設計／帝國欄位語意未全命名 | `applyLeaderETACallback` 撤銷／重套殖民地領袖增量並刷新所有殖民地士氣，保留領袖任職 |
| CMBTSHP timer 仍不可由目前靜態鏈定值 | `sub_30062`／`sub_30631`／`sub_31F25`／`sub_3F628` 可證實 loader、draw、input loop、heading，但未見 frame counter／clock／timer 寫入 | 未知 | `CMBTSHPFrameAtTick` 固定 4 tick/frame、`[0,1,2,1]`；採可重播近似 |

獨立文字 oracle：`assets/i18n/help.json` 的 `Sabotage` 說明直接寫「tries to destroy buildings」；
它與 raw `0x145EA` 的旗標清除互相支持「建築破壞」解釋，但不額外證明特殊槽位或精確選取機率。

## 2026-08-09 逐對手外交入口切片

| 主張 | 來源／證據 | 狀態 | 重製對映 |
|---|---|---|---|
| RACES 會列出多個已接觸勢力 | 原版 RACES 列欄位與既有 `racesRelationBars`／AI 對手資料 | 已證實（列與資料）；單列 REPORT 選取語意未完整追出 | 每個 AI 列的 remake 明確熱區 |
| 和平／貿易／威脅操作是目前可執行的 remake 外交入口 | `internal/shell/session.go:DiplomacyResponse` 與現有 `diplomacyScreen`；操作會調整 `AIOpponent.Relation` | 已接（remake 內部路徑） | 逐 AI 列 → `diplomacyWith` → 三個操作 → RACES |
| 原版正式條約的持續狀態與數值 | 手冊／當時的反組譯搜尋尚未形成可重現的條約消費端；`internal/diplomacy` 自身註記為設計性重建 | 當時結論，已由下方勘誤更新 | 見下方 raw code／欄位讀寫證據；仍不宣稱完整數值對齊 |

## 2026-08-09 正式外交條約與協議切片（勘誤與新增證據）

> 上一節「不新增狀態」是當時工具搜尋尚未把既有 `Orion2.exe.asm` 的 call site
> 與欄位讀寫串起來的暫時結論。後續在同一份原版輸入（SHA256
> `7AE2AC2E5904CA330009AF2827279D889906B0B9B7A8854C38EB707A56E955B5`）確認
> `sub_17D5B`、`sub_180D4`、`sub_101E77` 與 `sub_101EE3`／`sub_101F82` 的證據，
> 因此推翻「完全未知」；不推翻原先對完整數值係數仍缺 oracle 的保留。

| 主張 | 來源／證據 | 證據等級 | 重製對映 |
|---|---|---|---|
| 正式外交狀態有和平、互不侵犯、同盟三種可提議項目 | `sub_17D5B` 依序把 raw code `7/8/3/1/2` 送進 `sub_5265E`；code `3` 直接呼叫 `sub_524FB Declare_Peace_`；code `1/2` 呼叫 `sub_5232E Start_Treaty_`；`help.tsv` 與既有 openorion2 `ForeignPolicy` enum 對應語意 | 已證實（狀態與 call site）；code 1/2 的 Alliance／Non-Aggression 語意為強推論 | `TreatyState.FormalPolicy` 使用 `gamedata.ForeignPolicy`，外交畫面可提議／終止並保存 |
| 貿易／研究協議是獨立且可並存的旗標 | `sub_180D4` 讀 `[+0x62F]`、`[+0x637]`；`sub_101E77` 分別呼叫 `sub_101D53` 與 `sub_101DE5`；建立函式是 `sub_101EE3`／`sub_101F82` | 已證實（原始位址與讀寫端） | `TradeActive`／`ResearchActive` 分開建模，兩者可同時存在 |
| 協議收益依雙方較低的協議基準，從負值投入期逐步趨近目標 | `sub_101B3C`／`sub_101C93` 取兩方基準較小值後除以二；建立時目前值設為負基準；`sub_101D53`／`sub_101DE5` 以 5 回合差額逼近；`help.tsv` 明確寫第 5 回合損益兩平、至少 10 回合達最大 | 已證實（流程）；以 remake 總人口映射 `+0x5A2/+0x5C4` 為強推論 | `TreatyState` 保存雙方目前值／回合；`engine.EmpireOutput` 把收益納入 BC／RP |
| 神級商人貿易條約收益 +25% | `help.tsv` 條目「Fantastic Traders ... +25% bonus ... trade treaties」 | 已證實（手冊文字） | 貿易目標對玩家／AI 分別套用 +25% |
| 原版完整目標係數 | `sub_101BA4` 仍讀政府／特質／貿易項目表；`sub_101CC5` 仍讀研究相關政體欄位；目前沒有安全的完整欄位映射與 oracle | 未知 | 不把 remake 的人口基準、確定性逼近與既有特性以外的係數宣稱為原版精確值 |
| 正式和平／互不侵犯／同盟阻止 AI 突襲 | 原版 help 對互不侵犯／同盟的不得攻擊語意；remake 的 `aiRaidWilling` 以 `TreatyState.BlocksOffensive` 守門 | 原版語意已證實；重製接線已驗證 | `treaty_test.go` 驗證正式狀態、攻勢守門、回合收益與 JSON round-trip |

### 2026-08-09 條約目標倍率勘誤（保留前述歷史結論）

> 上表的「神級商人 +25%」保留為當時依 `help.tsv` 建立的歷史結論，不刪除其
> 證據來源。後續把原版欄位基準與特性編號對齊後，得到更直接的執行檔證據；因此
> remake 從本次起採用執行檔數值，並把百科文字衝突標為待查證，而不是把兩者
> 假裝成同一個數字。

| 主張 | 來源／證據 | 證據等級 | 重製對映 |
|---|---|---|---|
| `+0x8B7` 是神級商人特性，原版貿易目標加 50 個百分點 | `race_traits.go` 的原版特性陣列從 `+0x89F` 起算 31 格；`TRAIT_FANTASTIC_TRADERS = 24`，所以 `+0x89F + 24 = +0x8B7`。`Orion2.exe.asm` `sub_101BA4`（輸入 SHA256 `7AE2AC2E5904CA330009AF2827279D889906B0B9B7A8854C38EB707A56E955B5`，IDA 2026、LE 32-bit OS/2/DOS 位址空間）在 `loc_101BDC` 讀 `[ecx+8B7h]`，非零即 `add esi, 32h`。 | 已證實（欄位偏移、特性編號、原始指令互相吻合） | `treatyTarget` 對神級商人加 50 個百分點；`treaty_test.go` 固定驗證獨裁 150%、民主 200%。`help.tsv` 的 +25% 與原始指令衝突，保留為待查證文字差異。 |
| 原版貿易政府倍率 | `sub_101BA4`：`[+0x89F] == 4` → 150，`== 5` → 175，其餘 100；現有 `MoraleGovernmentType`／`OrigRaceTrait`／政府編號測試已釘住 4=Democracy、5=Federation。 | 已證實（原始分支與現有編號交叉驗證） | `treatyTradeGovernmentPercent` 套用 Democracy 150%、Federation 175%；AI 尚無目前政府欄位，保守沿用既有 Dictatorship 模型。 |
| 原版研究政府倍率 | `sub_101CC5`：政府 0→50%、1→75%、2/3→100%、4→150%、5→175%，其餘已知值回基準；同一份原始輸入與 `+0x89F` 政府編號證據。 | 已證實（逐分支代入） | `treatyResearchGovernmentPercent` 接上 Feudalism、Confederation、Dictatorship、Imperium、Democracy、Federation、Unification／Galactic Unification；AI 政體仍固定為 remake 既有 Dictatorship 邊界。 |
| 完整條約目標仍非全數解出 | `sub_101BA4` 在倍率後還掃描 `dword_1930DC + 0xF71` 特殊貿易表；`sub_101B3C` 的 `sub_2346E` 版本／位元遮罩與 `+0x5A2/+0x5C4` 完整消費端尚未形成可重現模型。 | 強推論／未知 | 不加入特殊表、`+0x8B7` 以外的臆測修正；一次性餽贈、原版 SABOTAGE 的完整分數／特殊槽位、AI 防守 Agent raw policy 仍列為 oracle 缺口；remake 的結構化分數與 Agent 消費已完成。 |

### 2026-08-09 ARM／FST／NR 飛彈改造證據盤點

| 主張 | 來源／證據 | 證據等級 | 重製對映 |
|---|---|---|---|
| FST 的原版速度加成是 +4；完整攔截器 Beam Defense 消費鏈仍未證實 | `Orion2.exe.asm` `sub_3CD21`（原版輸入 SHA256 同上；IDA Pro 9.4，LE 32-bit OS/2/DOS 位址空間）在 `loc_3CE40` 對函式旗標 `0x10` 執行 `add edx, 4`；`sub_3DFE0`（原始名稱 `Fighter_Ocv_`）再以速度乘 5 形成 OCV-like 下游值。`help.tsv` 條目也寫速度 +4、成本／佔格 +25%。 | 速度分支已證實；完整原版防禦消費端未知 | `gamedata.MissileSpeedOf`／`MissileBeamDefenseOf` 已保留公式；FST 已進設計 UI 與快速／格子戰術 PD 垂直切片，但不宣稱原版攔截器全路徑完成。 |
| ARM 的手冊效果是摧毀所需傷害 ×2，且與 raw Dcv 分支相符 | 手冊 p.115-116；`Missile_Dcv_` @ `0x3E095` 的 high-byte bit `0x08` 使 `4/8/12/16` 等 Dcv 分支加倍；`Weapon_In_Range_` @ `0x3A0B9` 以 Dcv 與攔截傷害的 quotient/remainder 計算擊落數。原始符號仍未直接命名該 bit 為 ARM。 | 手冊效果已知；raw／消費鏈強推論 | `gamedata.MissileInterceptionDurabilityForRawFlags` 使用 raw `0x0800`；ARM 已進設計 UI、成本／存檔與兩條 PD 垂直切片，原版完整攔截器／AMR 路徑仍未知。 |
| 光束 NR 已有原版與重製消費端，魚雷 NR 尚未有 | 原版完整改造函式 `sub_394F7` 呼叫射程／命中輔助 `sub_39434`；後者在改造旗標 `0x20` 成立時跳過 `word_17D867` 的射程衰減表；remake `ResolveBeamShot` 目前依 `DamageDissipationPenalty` 接入，魚雷仍走固定彈頭傷害。 | 光束已證實／已接；魚雷未知 | 光束 NR 已可在遠距離看到固定傷害；魚雷 NR 保持待做，不把兩條路徑混為一談。 |

### 2026-08-09 週期納貢條約切片

| 主張 | 來源／證據 | 證據等級 | 重製對映 |
|---|---|---|---|
| 原版關係資料保存雙向納貢模式 | 同一份 `Orion2.exe.asm`（SHA256 `7AE2AC2E5904CA330009AF2827279D889906B0B9B7A8854C38EB707A56E955B5`；IDA 2026；LE 32-bit OS/2/DOS 位址空間）中，`sub_194C5` 的固定納貢選項呼叫 `sub_52049`；該函式在關係矩陣 `[row+col*2+0x63F]` 寫入 raw mode，並同步另一方向的關係資料。`sub_180D4` 讀取同一欄位以顯示／終止納貢狀態。 | 已證實（位址、讀寫端） | `TreatyState.PlayerTribute`／`AITribute` 以 raw mode 1／2 保存方向；JSON round-trip 由既有 AI `Treaty` 快照承接。 |
| 固定納貢選項是 5%／10% 的週期納貢 | `assets/i18n/misc.json` 的 `GIVING \\x82% TRIBUTE`／`RECEIVING \\x82% TRIBUTE`；`sub_194C5` 固定傳入 `EBX=1`／`EBX=2`，原版顯示路徑依 raw mode 1／2 分別傳 5／10。 | 已證實 | 外交畫面提供「給予 5% 進貢／給予 10% 進貢」與「終止進貢」；`TreatySummary` 顯示付出／收取方向。 |
| 納貢成本從付款方收入計算，並扣除先前納貢成本 | `sub_E1FC7` 讀取付款方收入欄位 `+0xAE` 與既有總納貢成本 `+0xE78`，再執行 `mode*(income-existing_cost)/20`，結果由 `sub_E2000`／`sub_E2710` 累加到帝國經濟；負值在原版路徑被夾為 0。 | 已證實（公式）；`+0xAE` 對應 remake gross 收入為強推論 | `tributeCost` 採同一 raw mode／20 公式；`EndTurn` 在雙方 `RunEmpireTurn` 後移轉 BC，玩家摘要暴露 `EmpireOutput.TributeCost`。 |
| 一次性金錢／科技／星系餽贈等同於週期納貢 | `sub_194C5` 後續選項另呼叫 `sub_19F26`、`sub_1C479`、`sub_1D237`；其支付、選科技／星系與接受條件尚未形成完整可重現模型。 | 未知 | 不把一次性餽贈誤接成週期納貢；目前只接已證實的固定 5%／10% 路徑。 |

## 2026-08-09 安塔蘭終局防禦艦級消費切片

| 主張 | 來源／證據 | 證據等級 | 重製對映 |
|---|---|---|---|
| 終局防禦艦不是同級清單，而是 `Intruder ×3`、`Interdictor ×2`、`Harbinger ×7`，另加一座星際要塞 | 原版 `Load_Antaran_Defense_Fleet_`（`Orion2.exe` raw `0x4D18E`；`func_names.txt` 的原始函式定位）逐尺寸呼叫 `Load_Combat_Antaran_Ship_`，上限表 `_n_max_antaran_def_ships` 位於 raw `0x181746`，值為 `{0,0,3,2,7,0,0,0,0}`；要塞由同一載入器之後獨立追加 | 已證實（數量／尺寸）；星際要塞的 remake 戰力仍是代理 | `antaranDefenseUnit` 保存原版名稱、`CombatShipClass` 與 `Fortress`；終局戰不再把所有防禦單位退化成零值 Frigate |
| 球形武器在終局戰應按安塔蘭艦體級數計算 | 手冊球形傷害規則與重製 `battleShot` 的 `sizeClass` 消費端；原版五級尺寸到重製六級階梯的對照仍是相對順序映射：Large→Battleship、Huge→Titan、Titan→Doom Star | 已證實規則；艦級階梯對照為強推論 | `AssaultAntares` 將 `antaranDefenseUnit.CombatClass` 傳入終局戰 `combatant.sizeClass`；要塞沿用既有 Doom Star 等量代理 |
| 安塔蘭標準戰機艙與非空武器槽 | 即時戰鬥 loader `sub_55738`（Intruder）在 ID `31` 槽寫入數量 `3`；`sub_55F67`（Harbinger）在 ID `31` 槽寫入數量 `6`；`sub_55B12`（Interdictor）沒有 ID `31` 槽。其餘非空槽的 ID／數量／`+0x56` raw flags 也依 `0x0B` 槽距保存，完整表見 [`docs/re/antaran-defense-fleet.md`](re/antaran-defense-fleet.md)。`OrigWeaponTable` 將 ID `31` 定義為 `WeaponCatFighterBay`。先前所列 `0x5565D`／`0x55E16` 是 `Init_Ship_Designs` 的 compact strategic design 分支，並非這兩艘的即時戰鬥 loader；勘誤見 [`docs/re/antaran-defense-fleet.md`](re/antaran-defense-fleet.md)。 | 已證實（槽位、數量與 raw flags；不等於敵方精確命中／傷害公式） | `antaranDefenseUnit` 保存 `CombatLoaderRaw`、`WeaponSlots`、`FighterBayWeaponID` 與 `FighterBayCount`；終局快速戰鬥只沿用玩家既有 `FighterBayCombatContribution`，其他武器槽不猜火力。 |
| 安塔蘭其餘武器槽與戰機精確解算 | `Init_Ship_Designs`（raw `0x5514C`）、`Load_*_Antaran_Design` 與 `Load_Antaran_Star_Fortress_` 確實有更多武器／特殊槽，但即時戰鬥 loader 的欄位語意、完整多武器分配、敵方 `Fighter_Ocv_` 呼叫參數仍有未解析反編譯區域 | 未知 | 不把 ID 4／11／13／24／37 或暫定傷害轉成單一敵方武器；原版艦身方向與精確戰機命中／傷害仍保留缺口 |

## 2026-08-11 同一輸入 IDA 深度勘誤

本輪以 `Orion2.exe.i64`（SHA-256
`4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`）為主資料庫，
工具為 IDA Pro 9.4／Hex-Rays `9.4.0.260610`，位址基準為 IDA 線性位址；只在 Docker
複製資料庫探測，沒有執行原版。舊 `orion2_all.c`（SHA-256
`c2b5c30701019c0cc58763eb29c2abddb55eb551e1e7e52f68070d629e694505`）對應較舊
`ORION95.EXE`，不再作本輪 `Orion2.exe` 的主證據。

- 敵方：`sub_4D10E` 的五分支 dispatcher 與 `sub_55161`／`sub_5542C`／`sub_55738`／
  `sub_55B12`／`sub_55F67` 五個 raw writer 已追回；`dword_192864 + 0x139*index`、
  名稱、尺寸值與非空 `(ID, quantity, raw flags)` 已記錄。`Intruder`／`Interdictor`／
  `Harbinger` remake map 依同一資料同步，Interdictor／Harbinger 第 4 槽修為 raw flags `4`。
- 外交：`sub_53146` 的五類 proposal、完整輸入、modifier 消耗與四桶，以及 `sub_539D9`
  九類 demand 的 payload、雙門檻與成功 consumer 已於 2026-08-28 閉合。後者不是餽贈接受器；
  完整證據見 `docs/re/diplomacy-response-audit-20260828.md`。
- `.GAM`：`sub_10E2F` 以 `rb`／版本 `0xE0` 讀取 `dword_1930DC` 的 `0x3B×0x43`，
  `sub_1160B` 以 `wb`／同版本同形狀寫出，`sub_1307F` 逐筆消費 67 筆；原版全局匯入鏈
  已證實，重製原生 importer 仍未實作。

本勘誤覆蓋本檔較早把「敵方精確 blueprint、外交門檻、`.GAM` 全局匯入」統列為未知的
快照；戰機下游命中／傷害、要塞完整火力、特殊外交表與任命／任期規則仍保留為未完成 oracle。

## 2026-08-11 GAM／戰機／星際要塞 remake 接線勘誤

上一節記錄的是 IDA oracle 的中途狀態；本節補記已完成的 remake 消費端，避免把
「尚未取得逐值 runtime oracle」誤讀成「功能尚未接上」。原版證據仍以本檔上方的
輸入雜湊、IDA 版本與線性位址為準，沒有改寫歷史結論。

| 項目 | 已接入的 remake 行為 | 證據邊界 |
|---|---|---|
| `.GAM` | `internal/shell.ImportGAM`／`LoadGAMSession` 依 `0xE0` 版本 magic 進入原版固定記錄解析；`LoadSession` 自動分流，載入畫面也會探測同槽 `SAVE1.GAM`～`SAVE10.GAM`，讀入後另存 remake JSON。星系、行星、殖民地／前哨站、玩家／AI、外交旗標、67 筆領袖、艦隊、建築與建造佇列會轉成可玩的 `GameSession`。 | 研究完成 byte、未知特殊槽 bit、任命／任期下游仍以報告註記，不猜測語意。真實 `SAVE10.GAM` 已抽樣匯入。 |
| 敵方戰機 | 原版 Fighter Bays ID `31` 第二組傷害範圍 `Interceptor=1..4`、`Heavy=4..16`、`Bomber=2..7`，以及 `sub_3AD57 @ 0x3AD57` 的攻防修正、`roll <= 95`、40 命中門檻與相鄰 `sub_3AC20 @ 0x3AC20` 直接插值式已分開接入；逐架結果會進最弱護盾面、裝甲、結構傷害鏈。 | `sub_3DF8D`／`sub_3DFE0` 的部分欄位語意、兩份外部函式名稱與 raw record 逐彈參數仍未知；這不再阻塞下游命中／傷害功能。 |
| 星際要塞 | `sub_4D18E @ 0x4D18E` 追回的四槽已保存為 seed／raw／cap：`375/2/99`、`187/0/198`、`187/4/198`、`375/2/99`；class 6 讀址 `0x17F69C=900`、`P=750` 與直接火力消費已接入，快速戰鬥採 full-cap policy。 | `sub_6EE8E @ 0x6EE8E` divisor、raw 2/4 百分比表與 live tech 分支已追回；raw flags 正式命名與 live tech 導出的當下數量仍標證據等級，不宣稱逐位元原版總戰力已完全重建。 |

本輪抽樣驗證：Docker 內 `go test ./internal/gamedata ./internal/shell` 通過；以真實
`SAVE10.GAM` 通過 `TestImportGAMFixture`，讀出 `(Auto Save)`、36 顆恆星、82 顆行星、
1 個殖民地、0 個前哨站、4 個 AI、3 艘艦艇與 67 筆領袖。`cmd/moo2` 的 GUI 目標測試
依使用者要求不作完整回歸；先前 Xvfb 全套啟動在 GUI 初始化階段無界等待，未被當成
邏輯測試通過。

## 2026-08-11 Fire／要塞 runtime 參數深度追查（第二次勘誤）

本次沿用上節 `Orion2.exe.i64` SHA-256
`4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`、IDA Pro 9.4／
Hex-Rays `9.4.0.260610`、IDA 線性位址與非破壞性 IDC。結果不是 runtime 執行：

- `0x3AC20`／`0x3AD57` 的外部名稱索引互相衝突，研究記錄改以 raw function 位址為主。
  `sub_3AC20` 的候選特殊武器式與 `sub_3AD57` 的戰機對艦式分開記錄，不再把相鄰函式
  當成同一路徑。
- `sub_3AD57` 已追回 raw/min/max 取值、`sub_3DF8D`／`sub_3DFE0` OCV 呼叫、
  `sub_1247A0(100)` 的 1..100 隨機、`R<=95` 修正、40 命中門檻、
  `floor((M-40)*(max-min+1)/60)` 與 `DI==0` signed half；`var_18` 只可能 0／0x40，
  所以 `test var_18,4` 不可達。完整欄位名稱與 live target 狀態仍未知。
- `sub_4D18E` 的 class 6 table 讀址是 `0x17F69C`，不是舊探針誤算的 `0x17F6F6`；
  raw value 是 900。`P=750`、四槽 seed `375/187/187/375`、raw `2/0/4/2`、
  cap `99/198/198/99` 與超過 99 的拆槽已證實。
- `sub_6EE8E` 回傳 divisor；`sub_6A636` 的 raw 2/4 百分比為 200/50，
  `sub_6A406` 對四槽 raw 0/2/4 為 0，`sub_6F11C` 對 ID 4/11 靜態為 B=1，
  `sub_6E60E` 的 B=1 base 為 T=2→700、T=3→600、其他→400。ID 4/11 的 K=15/30，
  但 T 仍由 `sub_6E70A` 讀 live tech 狀態，故只列可重算 bucket，不冒充開局數量。

證據細節、原始運算元、雜湊與 probe 入口已集中到
[`docs/re/oracle-static-ida-20260811.md`](re/oracle-static-ida-20260811.md)；這輪不再
深挖與遊戲行為無關的反組譯內部功能。

## 2026-08-11 地面戰／事件／爆炸／CMBTSHP／外交尾項

本輪沿用同一份 `Orion2.exe.i64`（SHA-256
`4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`）與 IDA Pro 9.4，
在既有 Docker IDA 副本中追加 `sub_94951 @ 0x94951`／`sub_93D4B @ 0x93D4B` 探針。

- 地面戰 `0xEC4FE` 的兩次 `Random(100)+currentType*2+2`、平手雙命中、AI 守方
  裝甲營／兵力上限與 `0xECECA` 只寫地面部隊欄已核對；remake 改用原版 LCG 的
  rejection-sampling adapter，保留守方存活兵力與 captured population，不把未證實
  的人口 consumer 反推成公式。
- 事件 `0x22D57` 的好／壞極值排除與差平方權重、`0x586D4` 的 `0x200` 反覆減半已
  進純 oracle 測試；後續 2026-08-25 已證實 event record 為 36 槽、隨機候選為 0..28，
  並以 `sub_2230A` 原版排程取代 30%。raw score `+0xA6`、全銀河目標與 36-case dispatcher
  的完整 record consumer 仍未對回 remake。
- 爆炸 `0x3868F`／`0x39985`／`0x40C2A`／`0x494A8` 的亂數範圍、每步減 20、type 7
  四分之一、resistance、engine potential 與 raw `0x14` 分支已進純 oracle；戰略
  `Ship.Damage` 與原版 record 的完整 consumer 尚未證實，不接錯欄位。
- CMBTSHP 精確映射已由 `sub_30062` 固定為 `45*color+rawPicture`（0..43，44 sentinel）；
  原版 20 幀 timer 仍未知，但 remake 已接移動後固定 tick 的顯示近似。這修正了先前把
  艦級 fallback 寫成原版 picture 對照的過期描述。
- `sub_101BA4`／`sub_94951` 追回活動 Trader 的完整最大加成表：經驗 bucket 0..5，
  tier1=`×10`、tier2=`×15`；GAM importer 新保留 raw Experience，Treaty target
  實際消費。`Steal_App @ 0x10130A` 的 49 筆 `+8` 建造成本表、slot 9 skip、累積
  `Random(total)` 也已接入；remake 的結構化分數／防守 Agent 消費已完成，原版 raw
  分數／特殊槽政策仍未知。

文件同步：`README.md`、`WORKLIST.md`、`docs/tech/ground-combat-algorithm.md`、
`docs/tech/spy-system.md` 與
`docs/re/special-trade-sabotage-leader-eta-20260811.md`。

## 2026-08-27 編譯器／runtime 排除指紋

使用 IDA Pro 9.4／IDAPython，對 repo 外唯讀保存的 DOS 與 Win95 正式資料庫建立非破壞性
inventory 與代表函式輸出。DOS `Orion2.exe`（SHA-256 `7ae2ac2e…55e955b5`）已證實為
Watcom C/C++ 32-bit runtime＋DOS/4GW；Win95 `ORION95.EXE`（SHA-256
`6e19afdc…6e275185`）已證實為 Microsoft Visual C++ runtime 家族，PE linker 為 4.20。
精確編譯器小版本沒有足夠證據，維持未知。

本輪以位址／邊界、bytes SHA-256、指令形狀、IDA library 辨識與 caller role 建立 Watcom
`__CHK`／`__STK`／`__STKOVERFLOW_`／`__chk8087_`，以及 Microsoft
`__alloca_probe`／C++／SEH handler 的排除 pattern。這些 helper 不計入玩法 parity，Go 也不
照譯；只有 caller 把其結果轉成玩家可見分支時才重新開啟 caller 切片。另確認舊
`symbols_ea.tsv` 在 runtime 區段仍有一格錯位，後續一律用 `symbols_fixed.tsv` 並以 bytes／xref
獨立錨定。完整證據見 `docs/re/compiler-runtime-fingerprint-20260827.md`。

## 2026-08-28 `Change_Relations_` caller／reason 閉合

以 IDA Pro 9.4 的正式 `.i64` 副本重新匯出 `Change_Relations_`、`Reward_Check_`、
`Determine_Diplomacy_Messages_` 與 `Determine_Bad_Message_`，保存 30 個直接 callsite、
17 個 caller 及四個根函式 xref。原始指令證實 `Reward_Check_` 唯一可達 delta 是固定
`+200`；`Random(15)+15 = 16..30` 區塊被恆真的 `cmp ax,ax／jle` 擋住，屬不可達 dead code。
所有可達 delta／reason／payload、pending record 與每回合訊息 consumer 已閉合；多人函式只列
透傳邊界，不重複計入玩法 producer。完整分級與勘誤見
`docs/re/change-relations-callers-audit-20260828.md`。

## 2026-08-28 銀河議會完整投票 score

以正式 `.i64` 副本補匯出議會九個 root、五個外交 response producer、四個解約 writer 及
七組位移索引。`Vote_Check_` 的 `+0x617／+0x6D7／+0x7EE／+0x827` 已分別閉合為目前關係、
reputation、違約積怨與 NPC proposal response memory；`sub_78398` raw 資料流則回傳
star record #0 第一座有效殖民地的 owner，外部名稱 `Orion_Owner_` 只作語意導覽。
高難度真人 signed scaling 與完整 score consumer 已閉合，詳見
`docs/re/council-voting-audit-20260824.md`。

## 2026-08-28 領袖回合、ETA callback 與 AI 任命評分

以正式 `.i64` 副本匯出 24 個 gameplay root、八類共 1,379 個領袖欄位 operand site 與兩個 AI heap cache
pointer slot 的 producer／consumer。`0x934CF` 已訂正為 `Decrement_Officer_ETA_`，真正的
`Deassign_Officer_` 是相鄰 `0x933F2`；ETA 到 0 的 callback 以 star＋owner 重算同星系合格
殖民地、imports 與帝國彙總，不是單一領袖刷新。AI dispatch、艦長五欄支配排序、管理領袖
九組技能評分與 cache producer 均已閉合；`D7078` 只做同位置艦長重派。compiler／runtime helper
只作邊界定位，沒有納入玩法分母。詳見 `docs/re/leader-turn-chain-audit-20260828.md`。

## 2026-08-28 人口成長完整 runtime 與回寫

以 IDA Pro 9.4 重新匯出 `Colony_Pop_Grows_ @ 0xE1839`、
`Apply_Colony_Pop_Growth_ @ 0xE2DCA`、事件／科技／領袖／人口上限 helper 及 star #0 reserved
flag 的初始化／改名保存鏈。1.31 的完整 factor、Cloning Center 逐族截斷、prisoner 衰減、
滅絕每回合刪除、負池→正池、容量與新人口職務回寫均已閉合。特殊 `+150` consumer 來自
star #0 `+0x0E` bit 3；正式 1.31 binary data 為 0 且無直接 writer，標準新局不生效，
patch/profile 正式設定名維持強推論。數學／亂數 helper 只保存玩家可見契約，不逆向內部。
詳見 `docs/re/population-growth-runtime-audit-20260828.md`。

## 2026-08-28 戰略戰鬥模式與戰機駐防 120 分支

以 IDA Pro 9.4 全域匯出 `byte_199CB4` 的 52 個直接 operand site 與 39 個 owner，並逐一保留
raw 指令、caller 與資料流。`Russ_Combat_` 的 selector、新遊戲反相 toggle、預設值、存檔及
多人同步共同證實值 1 選戰略戰鬥解算，值 0 選格子戰術。因此 Fighter Garrison 的固定 120
是戰略模式，40／0、40／24、32／24 權重公式是值 0 分支；舊命名已訂正。C runtime 與網路
平台 helper 只保存持久化／同步邊界，不納入玩法分母。詳見
`docs/re/strategic-combat-mode-audit-20260828.md`。

## 2026-08-28 Fighter Garrison 格子戰術鏈

IDA Pro 9.4 已分開 `Qload_Ships_` 的 33-byte 快速結算記錄與
`Load_Tactical_Colony_` 的 313-byte 格子戰術 combatant。Fighter Garrison 會建立兩個
fighter slot，依科技得到 10／6／4 中隊，再由 `Fire_Missile_` 以每中隊四架展開；部署、
發射位置、返航回收與 raw building ID47 被摧毀後的清槽／持久化邊界均已閉合。個別損失不寫回
colony；C runtime、排序 helper 與視聽平台內部不納入玩法分母。詳見
`docs/re/fighter-garrison-tactical-audit-20260828.md`。

## 2026-08-28 AI 機會攻擊多艦隊搜尋

以 IDA Pro 9.4／IDAPython 匯出並逐指令審查 `sub_1FD80 → sub_DBC5C → sub_DBB9F →
sub_D94B3 → sub_D8ED2`，另納入 hostility、enemy-colony worth、可達敵星與殖民地太空／地面
防禦 helper。搜尋已閉合為逐 ship 航路、十回合候選、抵達前防守造艦、攻防剪枝與正常模式
加權抽選；對真人的機會攻擊只評估不移動。重要勘誤：`+0x837／+0x847／+0x857` 是 planet ID、
enemy-colony worth、攻防壓力，不是舊概稱的 star／score／pressure。compiler helper、C runtime
與平台 API 未納入本切片。詳見 `docs/re/opportunity-attack-search-audit-20260828.md`。

## 2026-08-28 AI 對 AI 實艦戰鬥與傷亡回寫

以 IDA Pro 9.4／IDAPython 匯出 15 個 gameplay roots，閉合 `Search_For_Battles_ →
Do_1_Combat_ → Get_Combat_Ships_ → Russ_Combat_ → Strategic_Combat_` 及死亡／撤退回寫。
AI pair 使用實際 ship ID，雙方皆非真人時固定走戰略解算；destroyed combatant 直接刪除原 ship，
倖存艦回寫戰損，沒有 FleetStrength 比例後再抽損失艦的旁路。格子戰術 `End_Of_Combat_` 只作
相同 ship-ID 契約的交叉驗證，不是 AI↔AI 可達分支。詳見
`docs/re/ai-ai-ship-battle-audit-20260828.md`。

## 2026-08-28 客製種族 31-byte 全庫直接消費端普查

以 IDA Pro 9.4 對全部函式搜尋玩家陣列基址與 `+0x89F..+0x8BD` raw operand，確認 runtime
固定為 31 個 signed byte；Poor Homeworld 是 index 15 的 `-1`，不是額外 index 31。共找回
137 個直接 owner 位址，受雜湊約束的 `symbols_fixed.tsv` 在 exact address 命中 136 個；證據
同時保留 IDA 原始 `sub_xxx`、函式邊界、bytes、運算元與外部符號名稱，沒有破壞性改名。
本普查只升格欄位身分與直接 consumer 存在性，不把函式名稱升格成公式證據，也不涵蓋間接
指標／衍生欄位 consumer。已同步訂正 `+0x8B8` 為 Telepathic，並保留其叛亂 prisoner 分支的
精確語意仍待資料流審查。詳見 `docs/re/custom-race-trait-consumer-census-20260828.md`。

## 2026-08-28 Telepathic 四條玩家可見鏈

對 `player+0x8B8` 的 13 個直接 owner 與 caller 逐指令審查：`Compute_Spy_Bonuses_` 證實
Telepathic raw 1 對攻守間諜基礎加 10；`Diplomacy_Test_` 對 treaty proposal accumulator 加 25。
心控合法性要求攻方 Telepathic、同星系至少 raw size 2 艦艇、兩次 general skill 0 helper 均失敗、
目標非 Telepathic；`Do_Attacker_Beat_Colony_Stuff_` 證實無 transports 時也可開啟同一接管路徑，
`Change_Pop_Ownership_` 則跳過異族 prisoner bit。戰術 AI 的登艦 action type、自毀權重與俘獲後
raw flag，以及 AI 選殖民地／worth 調整亦已定位；五個 raw 下游仍誠實保留未知。詳見
`docs/re/telepathic-trait-audit-20260828.md`。

## 2026-08-28 Omniscience／Stealthy Ships 與 operand 掃描勘誤

`Player_Is_Omniscient_` 已證實等於 Galactic Lore 領袖或種族 Omniscience，15 個 caller 覆蓋
星圖、Fleet、行星、摘要與玩家清單；Fleet 掃描 UI、`Star_Owner_` 可見性過濾、艦隊 stack
過濾、部分 discovery report 與隱藏熱鍵已分開閉合。Stealthy Ships 在 `Compute_AI_Data_` 與
逐艦 stealth device 共用「不要直接標為可見」fallback，並令 `Add_Specials_To_Design_` 跳過 raw
special 31 Stealth Field。

同時訂正全庫掃描證據等級：`Remove_Non_Detected_Ships_` 有位於 `0xF4D5E`、跳回 `0x5DC77`
的 Watcom distant tail chunk，IDA owner 與線性最近符號會互相衝突。因此 137 個數字是 candidate
owner 位址，不是 137 個已閉合公式；未人工確認 register provenance 的命中不得升格。詳見
`docs/re/omniscience-stealthy-ships-audit-20260828.md`。

## 2026-08-29 Stealthy Ships 遠端尾區塊資料流閉合

專用 IDAPython 匯出四個 `Remove_Non_Detected_Ships_` chunks、所有跨 chunk 入口與基本區塊邊。
`0x5DC5D..0xF4D66` 已證實由 `ship[0x63]` owner、`0xEA9` player stride 與 `_player` base 讀取
`+0x8BB`，因此該命中由強推論升為已證實。原版 scanner raw level 為 2／4／6／9，殖民地與
己方艦艇以 `ceil(Range/30)` 建立最大 detector margin；Stealthy Ships、Stealth Field、
Cloaking Device、Phasing Cloak 各加 500 concealment，最後只有 concealment 不大於 margin 的
艦艇留在可見 stack。三個 scanner 科技位移與 colony flags 的正式名稱仍不靠二手 enum 猜填。
證據見 `docs/re/evidence/stealthy-distant-tail-ida-20260828.json`。

## 2026-08-29 Omniscience discovery report 與一次性發現編碼

IDA 證實 `Make_Ship_Arrive_At_Star_` 在首次抵達時同步設定 `star+0x33` 造訪 bit 與
`star+0x34` 待報告 bit。`Do_System_Discoveries_At_Star_` 則把 Space Debris、Pirate Cache、
Splinter Colony、Lost Hero、Ancient Artifacts 編為 `star+0x28 = 50／60／70／80／90 + player`；
`Draw_Special_` 以十位還原類型、個位確認 recipient。Omniscience 會清掉普通 `+0x34` 報告而
不顯示，但一次性發現仍進 popup 並保留獎勵。`Reset_One_Shot_System_Special_` 另含
`0xF4F74` Watcom distant tail，完整 chunks 已保留。證據見
`docs/re/evidence/omniscience-discovery-report-ida-20260829.json`。

## 2026-08-29 Stealthy Ships AI profile 與科技 category

`Init_NPC_Personalities_Objectives_Themes_` 的六候選第一組已閉合 stack 權重順序；Stealthy
Ships 對候選 1 加 100，抽選結果寫 `player+0x28`。`Calc_Tech_Value_` 的命中不是單一科技：
category `0x25` 預設 multiplier 5，非人類 Stealthy AI 覆寫成 1；原始 212×13 tech-item 表
證實該類別包含 Cloaking Device、Displacement Device、Phasing Cloak、Stealth Field
（ID 38／53／126／172）。人類分支會在後面重建 multiplier。證據見
`docs/re/evidence/stealthy-profile-tech-ida-20260829.json`。

## 2026-08-29 Stealthy Ships 與匿蹤裝置戰鬥邊界

全庫 direct census 確認 `player+0x8BB` 只有五個站點，快速／戰略與格子戰術載入器均未直接
讀取；`Ship_Has_Stealth_Device_ @ 0x5D3DB` 的唯一 caller 是 AI 可見性資料 producer。
`Qload_Ships_ @ 0x416CF` 雖遍歷 39 個特殊裝置，但 raw 6 Cloaking Device、23 Phasing Cloak、
31 Stealth Field 的 `0x17EF0C + id*0x2F` effect type／signed value 都是 `0／0`，不向 33-byte
快速結算記錄提供通用數值加成。`Load_Combat_Ship_ @ 0x4954A` 則在
`0x49634..0x4963E` 把五個設計 bitfield bytes 從 `+0x17／+0x76` 複製到格子記錄
`+0x4C／+0xB2`。因此 trait 與三裝置的戰鬥等價被強推論否定；下一切片只追裝置 bitfield 的
格子間接 consumer。完整 raw bytes、函式、caller 與輸入雜湊見
`docs/re/evidence/stealth-battle-boundary-ida-20260829.json`。

## 2026-08-28 Ship Attack／Defense／Trans-Dimensional 多路徑消費端

Ship Attack／Defense 的 signed raw 值已分別在格子戰術 OCV／DCV、戰機 OCV、飛彈／戰機
命中下游、戰術 AI、`Qload_Ships_` 快速／戰略 record 及艦艇／衛星估值找到直接 consumer。
Trans-Dimensional 已證實格子 current speed +4、三組飛彈 base speed +4、估值 DCV +20、FTL
+2，並在 `Ships_Try_To_Move_To_`、AI 移動、機會攻擊、外交威脅與特殊宣戰繞過 Hyperspace
Flux gate。舊 gap report 稱 `+0x8BC` 未知的歷史敘述已追加勘誤。strategic record 欄名、bonus
輸出型別、撤退完整 gate、科技 raw ID、AI profile 與 RNG 序列仍保留未知。詳見
`docs/re/ship-racial-bonuses-transdimensional-audit-20260828.md`。

## 2026-08-28 Repulsive／Charismatic 多系統公式與 gate

兩項 trait 已拆成不同尺度：議會 Charismatic +40／Repulsive -100、treaty proposal +50／-50；
Charismatic 使關係正向 delta ×2、負向 delta 向零 ÷2，科技交換兩路 +50；領袖出現 accumulator
為 +5／-10，候選 score 則 +10／向零 ÷2；同化 ×2／÷2；AI talker priority 為 3／2／1。
Repulsive 另切換單機與正式網路外交 choice、阻止 AI↔AI treaty negotiation，並改走專用訊息
分支。`NPC_To_Human_Diplomacy_ @ 0x26990` 位於 IDA 誤併的大函式後段，已同時保留 raw 位址、
IDA owner 與外部符號。choice／message 完整表、advanced officer／AI leader flags、sneak-attack
完整權重、profile／talker raw gate 與 RNG 仍待。詳見
`docs/re/repulsive-charismatic-trait-audit-20260828.md`。

## 2026-08-28 Government 主要玩家消費鏈

由 `player+0x89F` 出發人工審查 IDA 候選與 caller，閉合八個政體編號、四項進階科技寫入、
封建族生產成本、三職務產出、士氣、殖民地 BC、帝國指揮點、間諜防禦、同化及貿易／研究
協議公式。AI 評分、產品合法性 raw 表與 occupation policy 仍獨立保留，不由函式名猜填；
C runtime、Watcom helper 與平台 API 明確排除於玩法分母。詳見
`docs/re/government-trait-audit-20260828.md`。

## 2026-08-28 經濟與環境種族特性整併

將原先散落於人口成長、食物、工業、研究、BC、地戰與 AI 殖民地資料的證據，以 31-byte
trait offset 重新串成垂直鏈。Population／Farming／Industry／Science／Money 的 raw 尺度，
Aquatic／Subterranean／Tolerant 的環境效果，以及 Fantastic Traders 三個不同 BC consumer
已閉合；開局母星、AI profile／估值與 Cybernetic／Lithovore 保留為後續窄切片。詳見
`docs/re/economic-environment-race-traits-audit-20260828.md`。

## 2026-08-28 Cybernetic／Lithovore 多路徑審查

IDA 證實 Cybernetic 除了食物／工業各半外，格子戰術每回合還有獨立 `+10%` 維修率，並在
戰後與 33-byte 戰略解算清除持久損傷；這訂正了舊文件只寫「戰後全修」的缺漏。Lithovore
則閉合零食物、初始 worker、AI 糧食行星 bypass，以及 Android Farmers、Biomorphic Fungi、
Food Replicators、Hydroponic Farm、Soil Enrichment、Subterranean Farms 六項科技合法性 gate。
詳見 `docs/re/cybernetic-lithovore-trait-audit-20260828.md`。

## 2026-08-30 Advanced Civilization 全圖分配

IDA 已閉合額外行星額度、距離／owner 候選 gate、未殖民行星 worth 與 proximity、隨機玩家
輪流選取、最佳玩家 90% 平衡、每顆最多六次屬性提升、special 4／5／10 再分配，以及最後
`Init_Colony_` 回寫。`Assign_Advanced_Civilization_Starting_Ships_ @ 0x63848` 在此 binary 是
無玩家可見效果的空迴圈；runtime helper 不納入玩法分母。詳見
`docs/re/advanced-civilization-planets-audit-20260830.md`。

## 2026-08-30 Cybernetic／Lithovore AI 行星價值

`Uncolonized_Planet_Worth_To_Player_ @ 0xD27A7` 已證實食物項為 Lithovore `0`、Cybernetic
`food*75`、一般 `food*150`。`Colony_Worth_To_Player_ @ 0xD2CAE` 的產出項分別為 Lithovore
`6*(industry+research)`、Cybernetic `4*(food+industry+research)`、一般
`3*(industry+research)+6*food`。NPC profile 與科技估值是不同資料流，未冒充本輪已閉合。
詳見 `docs/re/cybernetic-lithovore-trait-audit-20260828.md`。

## 2026-08-28 母星、重力、研究與 Warlord traits

IDA 直接寫入證實 Large 母星 size raw 3、Rich／Poor 礦產 raw 3／1、Artifacts special raw 10，
並閉合 Low／Normal／High-G 母星 raw 0／1／2、逐 race 產出、Low-G 地戰 -10、High-G 耐受及
轟炸門檻。Creative／Uncreative 的研究時序與 Warlord 艦員／領袖 +1、每殖民地 command +2、
兵營容量兩倍亦整併成單一證據鏈。詳見
`docs/re/homeworld-gravity-research-warlord-traits-audit-20260828.md`。

## 2026-08-28 Spying trait 完整垂直鏈

將 `player+0x8A8` 的 signed `-10/+10/+20` 接回共同攻守表、AI spy 生產需求與零庫存
bootstrap gate，再與 packed Agent／Spy pool、100 industry 訓練、每名 1 BC、63 上限、AI
留守／外派及三任務解算整併。確認難度只在 AI 攻真人的最終差值加入，不進共同攻防表。
詳見 `docs/re/spying-race-trait-audit-20260828.md`。
