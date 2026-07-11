# 剩餘工作路線圖(2026-07-11,14 輪死碼接線後)

> 目的:把「完整移植」剩餘工作,按**阻塞類型**分類、標依賴與建議順序,讓下一步只需一個決策就能啟動。
> 這不是現況清單(現況見 `HONEST-STATUS.md`),是**決策賦能的排序菜單**。

## 一、當前位置(這 session 14 輪的成果)

**中文化維度:完成**(4 專有名詞池 13+829+672+104 + 22 UI TSV,無未翻字串源)。

**移植維度:核心 gameplay 系統多數已從手冊/openorion2 錨定接線**——殖民地(重力/士氣/礦產/建築18棟/地形改造/污染/成長/饑荒/維護)、艦隊(指揮點數)、戰鬥(beam/飛彈/球狀分流、地面戰陸戰隊+戰車+軌道轟炸)、勝利(征服+議會+**安塔蘭反攻,三條路徑全接,能贏一局**)、領袖技能、間諜最小迴圈、外交關係核心。

> **安塔蘭勝利路徑已接(2026-07-11 追加)**:次元傳送門建築(`gamedata.Buildings` 早已存在,
> 本輪只補「建成後解鎖反攻」流程)+ `GameSession.AssaultAntares()`(戰鬥沿用 `ResolveBattle`
> 同款 `battleVolley` 解算,防禦方戰力用保守預設,見下)+ `advanceAntaranVictory`(`EndTurn`
> 偵測)。**母星防禦艦隊戰力手冊/openorion2 均無精確數字**(手冊只用「awe-inspiring」定性描述),
> 保守預設為 6 艘末日之星等級戰力,待考證。詳見 `docs/tech/victory-conditions.md` 第 4 節。

> **武器改造(mod)系統已接(2026-07-11 第三輪追加)**:手冊(`GAME_MANUAL.pdf` p.115-118)8 個
> 光束/通用 mod(HV/PD/AF/CO/AP/ENV/NR/SP)逐字核對佔格/成本/命中/傷害數字,接進
> `ShipDesignSpaceUsedWithMods`/`DesignCostWithMods`(佔格/成本)、`ResolveShotWithMods`(命中/
> 傷害,`battleVolley` 快速結算與 `fireRound` 格鬥畫面共用)、艦艇設計畫面 8 個 mod 勾選 chip。
> 無 mod 武器逐位元回歸不變。詳見 `docs/tech/weapon-mods.md`。

**關鍵洞察**:先前多輪誤判為「RE-gated 需 DOSBox」的東西,絕大多數是**前面 session 已移植進 `gamedata/` 卻沒接進遊戲迴圈的死碼**。接死碼(重力/士氣/勝利/飛彈/間諜…)是本 session 的主線,已挖到見底。

## 二、剩餘工作:按阻塞類型分類

> 「解鎖後可否自驅」= 一旦你給了對應決策/資源,我能否不再需要你就做完(帶硬門檻+驗證)。

### A. 需你 playtest 驗證(先做,成本最低,解鎖 income 死碼)
| 項目 | 說明 | 解鎖後 |
|---|---|---|
| **開局經濟平衡** | 本 session 數處改了開局狀態:士氣 +10→**0**(手冊忠實)、指揮點數(先前誤判**-20 BC/回合為忠實**,2026-07-11 同日已修復——用真實存檔 `SAVE10.GAM` oracle 反推補上帝國基礎供給 `gamedata.CommandPointsBase=5`,開局供給=5+1(星基)=6≥3(需求),不再超支,詳見 `docs/HONEST-STATUS.md`/`docs/tech/moo2-formulas-reference.md`「指揮評等供需」節)、領袖研究 **+25**(正 bonus 抵銷)。指揮點數這項死結已解,剩士氣 0 起跳+領袖 +25 這兩項待你 playtest 判斷是否偏苦 | 你回報「正常/太苦」→ 我接 income 死碼(政府 money bonus、morale 對收入的影響、貨運維護 `income.go`,均已移植) |

> **income 死碼接線已完成(2026-07-11,同日稍後)**:政府 money 加成(Democracy/Federation)已接進
> `RunEmpireTurn`,demo(Dictatorship)no-op;運輸艦維護費已接線但 remake 無 Freighter 艦種,
> 恆 0 no-op;士氣對收入的調整**判定為刻意不接**(收入已從士氣調整過的產出換算出來,再套一次會
> 雙重計算,見 `docs/tech/moo2-formulas-reference.md`「士氣對收入的影響」節)。20 回合 BC 軌跡探針
> 確認接線前後一致(101→130),無 regression。詳見 `docs/HONEST-STATUS.md` 2026-07-11 收入死碼段落。

### B. 需你授權方向的基礎設施(大工程,選錯白做,故等你點)
| 項目 | 依賴 | 自驅度 |
|---|---|---|
| **多 AI 對手 + 真星系拓殖** | **多 AI 對手數量已接**(2026-07-11:`NewDemoSession` 由 1 個 AI 擴為 3 個,各不同母星/種族名/`ai.Profile` 性格,議會門檻 `gamedata.CouncilMinExtantRaces` 真值可達、`advanceCouncil` generalize 為逐帝國計票,見 `docs/tech/victory-conditions.md`)。**拓殖部分已接**(2026-07-11:`shell.GameSession.ColonizeStar`,玩家可用殖民船在無主適居星建立新殖民地,起始人口/PopMax 公式對手冊+openorion2 核實,詳見 `docs/tech/colonization.md`)。**AI 側殖民地模型已接**(2026-07-11 追加:`aiExpand` 改用共用函式 `newColonyFromStar`,佔星時建真 `engine.ColonyState`,不再只標旗標——AI 經濟隨擴張成長,見 `docs/HONEST-STATUS.md` 同日追加段落) | 剩 **AI 選星策略**(現為星圖索引順序,非距離/資源導向)與 **AI 對 AI 互動**(3 個 AI 目前只各自獨立對玩家造艦/擴張/外交,彼此不打仗不外交,也沒有「候選人限定票數最高兩位+第三方外交搖擺票」的議會規則,需要先補 AI 對 AI 的關係模型)。給方向後可自驅 |
| **戰機/航母系統** | 戰鬥基礎設施 | 解鎖 `combat.go` CombatFighter* 死碼 + 戰機庫建築。自驅度中 |
| **艦艇軍官指派** | 需「軍官→艦艇」指派模型 | 解鎖 `ShipBeamAttackWithOfficer` 死碼(openorion2 `sptr->officer` 有對應)。小工程 |
| **飛彈/魚雷專屬 mod(ARM/ECCM/EMG/FST/MV/OVR)** | 武器改造(光束 8 個 mod 已於 2026-07-11 接線,見下方新增段落) | 手冊 p.115-116 已有精確數字,待飛彈解算(`ResolveMissileShot`)先建 mod 掛鉤機制。小工程 |
| **火線角(Firing Arc:Fwd Ext/Back Ext/360 Degree)** | 艦艇設計基礎設施 | 手冊 p.127-128 已有精確數字(+25%/+25%/+50%),與武器改造平行、獨立的機制,尚未接線。小工程 |

### C. 需你定互動設計的 UI(引擎層多已備,缺玩法介面)
| 項目 | 現況 |
|---|---|
| 完整間諜畫面 | 引擎最小迴圈已接(訓練→偷科技);缺逐對手分配/任務(STEAL/SABOTAGE/HIDE)選單。SABOTAGE 手冊無數值需先定 |
| 完整領袖招募畫面 | 技能效果已接;缺招募池/指派 UI |
| 完整外交畫面 | 關係核心已接;缺提案/條約/宣戰互動(畫面美術已重建,見 `diplomat-lbx-layout.md`) |

### D. 需外部 oracle(我無法自給)
| 項目 | 需要 |
|---|---|
| 地面戰核心解算校驗 | `ResolveGroundBattle` 的 d100+force 沿用一代(1oom)借用結構,force 值用 MOO2 手冊表但結構未對 MOO2 實機核實 → **DOSBox oracle** |
| 飛彈速度定案 | `missile.go` 手冊公式與附表自相矛盾 → **DOSBox oracle** |
| 逐畫面按鍵像素對齊 | 熱區多為估計座標 → **原版截圖** |
| 戰術艦型 sprite 對照(#12) | CMBTSHP 資產→艦級對照不在 openorion2 碼 → **DOSBox oracle** |

### E. 低價值精修(不急,列此備查)
- 外交離散事件 `EventDelta`(現用軍力差漂移替代)、引擎爆炸連鎖傷害、alt-to-hit 命中變體。

## 三、建議順序(投報比排序)

1. **你 playtest 開局平衡**(5 分鐘,零我方成本)→ 解鎖 income + 校準方向。
2. **多 AI + 真星系拓殖**(最大「能玩一局」解鎖)——你給設計方向,我自驅建。
3. 依你興趣:戰機/航母 或 飛彈 mod/火線角(安塔蘭勝利、光束武器改造 mod 已於 2026-07-11 接線,不再是待選項)。
4. UI 畫面(間諜/領袖/外交)——你定互動玩法後我接引擎+最小 UI,逐步補全畫面。
5. 待你安排 DOSBox oracle / 截圖 → 校驗地面戰·飛彈·像素對齊·sprite。

## 四、我的工作模式(已驗證有效,續用)
- 每塊帶**硬門檻**:先萃取手冊/openorion2 權威值,找到才建、找不到就停不准猜。
- 機械/移植/翻譯派 Sonnet subagent 實作,Opus 逐項對手冊核實+抓過度宣稱+docker build/vet/test 驗證+0 容器殘留。
- 每輪更新 `HONEST-STATUS.md`(清過期斷言)+ 本路線圖 + WORKLIST,推送 GitHub。
- 不做低價值 churn、不無方向建大 infra、不碰未驗證平衡——那不是真進度。
