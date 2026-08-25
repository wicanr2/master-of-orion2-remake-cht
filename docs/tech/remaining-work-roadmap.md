# 剩餘工作路線圖(2026-08-10 收尾版；歷史決策菜單)

> 目的:把「完整移植」剩餘工作,按**阻塞類型**分類、標依賴與建議順序,讓下一步只需一個決策就能啟動。
> 這不是現況清單(現況見 `HONEST-STATUS.md`),是**決策賦能的排序菜單**。

## 一、當前位置(這 session 14 輪的成果)

**中文化維度:完成**(4 專有名詞池 13+829+672+104 + 22 UI TSV,無未翻字串源)。

**移植維度:核心 gameplay 系統多數已從手冊/openorion2 錨定接線**——殖民地(重力/士氣/礦產/建築18棟/地形改造/污染/成長/饑荒/維護)、艦隊(指揮點數)、戰鬥(beam/飛彈/球狀分流、地面戰陸戰隊+戰車+軌道轟炸)、勝利(征服+議會+**安塔蘭反攻,三條路徑全接,能贏一局**)、領袖技能、間諜最小迴圈、外交關係核心。

> **安塔蘭勝利路徑已接(2026-07-11 追加)**:次元傳送門建築(`gamedata.Buildings` 早已存在,
> 本輪只補「建成後解鎖反攻」流程)+ `GameSession.AssaultAntares()`(戰鬥沿用 `ResolveBattle`
> 同款 `battleVolley` 解算,防禦方改用 2026-08-08 反組譯解出的 `Intruder ×3`／`Interdictor ×2`／
> `Harbinger ×7`＋1 座星際要塞切片,並以 remake 的艦級／要塞代理消費)+ `advanceAntaranVictory`(`EndTurn`
> 偵測)。**艦隊組成與標準艦非空武器槽已證實；精確總火力、要塞設計、艦級映射與即時戰鬥敵方武器
> 消費仍有近似／未知**。詳見 `docs/re/antaran-defense-fleet.md` 與 `docs/tech/victory-conditions.md` 第 4 節。

> **武器改造(mod)系統已接(2026-08-09)**:手冊(`GAME_MANUAL.pdf` p.115-118)的 8 個
> 光束/通用 mod 與 ECCM/EMG/MV/魚雷 ENV/OVR 已接進
> `ShipDesignSpaceUsedWithMods`/`DesignCostWithMods`(佔格/成本)、`ResolveShotWithMods`／
> `ResolveMissileShotWithMods`(命中/傷害,`battleVolley` 快速結算與 `fireRound` 格鬥畫面共用)、
> 艦艇設計畫面依武器類型顯示適用改造 chip。MIRV 逐彈頭判定；無改造武器的舊入口回歸不變。
> ARM/FST、魚雷 NR 與敵我戰機消費端已接入 remake 的設計／存檔／快速結算／格子戰術路徑；
> 原版完整攔截器／AMR／逐艦戰機 blueprint 仍待證，raw `0x0800`／`0x1000` 的分級保留。詳見
> `docs/tech/weapon-mods.md`。

> **狀態警告（2026-08-12）**：本文件以下內容是舊工程路線圖，不再是原版忠實度的完成依據。
> 「remake 消費端已接」與「原版玩法已對齊」曾被錯誤合併；最新現況以
> [`../re/parity-re-audit-20260812.md`](../re/parity-re-audit-20260812.md)、
> [`../re/parity-matrix.tsv`](../re/parity-matrix.tsv) 與 `WORKLIST.md` 頂端活表為準。

**已被推翻的舊判斷**：先前把許多問題歸類為 `gamedata/` 死碼未接，並以接上 remake 消費端後
宣稱「已挖到見底」。IDA Pro 重新抽樣證實，這種探針無法衡量原版 call graph 覆蓋，且議會、
轟炸、外交與 AI 已存在直接模型差異，因此不得再作為原版 parity 證據。

## 二、剩餘工作:按阻塞類型分類

> 「解鎖後可否自驅」= 一旦你給了對應決策/資源,我能否不再需要你就做完(帶硬門檻+驗證)。

### A. 需你 playtest 驗證(先做,成本最低,解鎖 income 死碼)
| 項目 | 說明 | 解鎖後 |
|---|---|---|
| ~~**開局經濟平衡**~~ **已完成 headless 基線(2026-08-10)** | 本 session 數處改了開局狀態:士氣 +10→**0**(手冊忠實)、指揮點數(先前誤判**-20 BC/回合為忠實**,2026-07-11 同日已修復——用真實存檔 `SAVE10.GAM` oracle 反推補上帝國基礎供給 `gamedata.CommandPointsBase=5`,開局供給=5+1(星基)=6≥3(需求),不再超支)、20 回合固定無事件探針為 BC 50→264、人口 8→11、士氣 0% 全程、食物 0→1、工業 6→8、研究 6 維持 | 這是可重播的平衡基線，不是玩家主觀回報；若日後玩家認為太苦，再另開調校工作，不在本輪改收入公式 |

> **income 死碼接線已完成(2026-07-11,同日稍後)**:政府 money 加成(Democracy/Federation)已接進
> `RunEmpireTurn`,demo(Dictatorship)no-op;運輸艦維護費當時已接線但 remake 無 Freighter 艦種,
> 恆 0 no-op——**此缺口已於同日晚些補上(#4)**:新增「運輸艦隊」建造選項後,玩家側
> `ActiveFreighters` 會真的變非 0,維護費隨之生效,並補上 1.3/1.5 版本現金加成差異,詳見
> `docs/tech/moo2-formulas-reference.md`「運輸艦淨現金版本差異」節;AI 對手仍未接同一建造流程,
> 恆 0 no-op。士氣對收入的調整**判定為刻意不接**(收入已從士氣調整過的產出換算出來,再套一次會
> 雙重計算,見 `docs/tech/moo2-formulas-reference.md`「士氣對收入的影響」節)。20 回合 BC 軌跡探針
> 確認接線前後一致(101→130),無 regression。詳見 `docs/HONEST-STATUS.md` 2026-07-11 收入死碼段落。

### B. 需你授權方向的基礎設施(大工程,選錯白做,故等你點)
| 項目 | 依賴 | 自驅度 |
|---|---|---|
| **多 AI 對手 + 真星系拓殖** | **多 AI 對手數量已接**(2026-07-11:`NewDemoSession` 由 1 個 AI 擴為 3 個,各不同母星/種族名/`ai.Profile` 性格,議會門檻 `gamedata.CouncilMinExtantRaces` 真值可達、`advanceCouncil` generalize 為逐帝國計票,見 `docs/tech/victory-conditions.md`)。**拓殖部分已接**(2026-07-11:`shell.GameSession.ColonizeStar`,玩家可用殖民船在無主適居星建立新殖民地,起始人口/PopMax 公式對手冊+openorion2 核實,詳見 `docs/tech/colonization.md`)。**AI 側殖民地模型已接**(`aiExpand` 使用共用 `newColonyFromStar`,佔星時建真 `engine.ColonyState`)。**2026-08-11 可選強化已接**：AI 選星讀行星價值／距離，議會讀 AI↔AI 關係做兩候選人與第三方搖擺票；`EnableAIVsAI` 另讓 AI 彼此依關係交戰、停戰／結盟、貿易／研究並以抽象艦隊佔領殖民地，詳見 `docs/tech/ai-to-ai.md` | remake 路徑已完成；原版逐艦 blueprint、正式 AI 接受門檻與外交細表仍是 oracle 差異 |
| **戰機/航母系統** | 玩家／敵方中隊、突擊艇、最弱護盾面，以及 ID 31 第二組 `1..4 / 4..16 / 2..7`、`sub_3AD57 @ 0x3AD57` 的 1..100／95／40／`max-min+1` 式與相鄰 `sub_3AC20 @ 0x3AC20` 的直接插值式已分開接入；敵方五級 blueprint writer／非空武器槽與要塞四槽、`sub_6EE8E @ 0x6EE8E` 容量 divisor 中間式也已由同一輸入 IDA 追回 | 兩份 raw 函式外部名稱衝突、`sub_3DF8D` 部分 runtime 欄位、raw flag 正式玩法名稱、live tech 導出的當下數量與 DOSBox 逐值 oracle仍待補；remake 命中／傷害／要塞火力消費已完成有證據邊界 |
| ~~**艦艇／殖民地領袖指派**~~ **已接(2026-08-10)** | `save.Ship.Officer`、艦隊／`LEADERS` 分頁、殖民地領袖任職／改派／解除／解雇與 JSON／熱座保存已接；同一輸入 IDA 已證實 `sub_10E2F`／`sub_1160B` 對 `dword_1930DC` 讀寫 `0x3B×0x43` 全局區塊；2026-08-11 新增原版 `.GAM` 匯入、RawExperience、活動 Trader 最大加成與 `RawStatus=4`／30 回合清理切片，並接 `RawETA=0` 的 `sub_E2AB1` 六槽鏈之玩家可感知殖民地計算近似 callback | `sub_E1D59`／`sub_DF8F0`／`sub_E2710` 的 raw 設計／帝國欄位仍是 oracle 差異 |
| ~~**飛彈/魚雷專屬 mod(ARM/ECCM/EMG/FST/MV/OVR)**~~ **remake 已接** | ECCM/EMG/MV、魚雷 ENV/OVR/NR、ARM/FST PD 與兩條戰鬥路徑已接 | 只剩原版完整攔截器／AMR／戰機 oracle；不再是 remake 小工程 |
| **火線角(Firing Arc:Fwd Ext/Back Ext/360 Degree)** | 設計／成本／保存與格子戰術方向切片已接；原版快速結算維持抽象模型 | 手冊 p.127-128 的 +25%/+25%/+50% 與原版 `Relative_Bearing`／`Move_Ship`／部署朝向已進 `Ship.Arc`、`CombatShip.Facing`、玩家／敵方格子射擊；`QGet_Target`／`Strategic_Combat` 快速 call graph 未讀取 arc。方向切片已結案，剩 sprite 幀號對照未知 |

### C. 需你定互動設計的 UI(引擎層多已備,缺玩法介面)
| 項目 | 現況 |
|---|---|
| ~~完整間諜／外交畫面~~ **remake 核心已接** | RACES 內嵌訓練、逐對手 STEAL/SABOTAGE/HIDE、餽贈、特殊貿易、正式條約、貿易／研究、納貢、終止與回合收益均已接；AI→玩家 SABOTAGE 也已接 remake 性格政策與玩家建築池；原版 49 筆 SABOTAGE 建築成本表、raw slot helper、結構化 AB/DB 分數與 Agent 實際消費已接 | 原版兩張 score table 的上游填值、特殊槽位／防守策略、特殊貿易尚未對回的 raw 上游與 AI 接受門檻仍未知 |
| ~~完整領袖招募／任職畫面~~ **remake 核心已接** | 艦艇 HIRE／POOL／DISMISS 與殖民地領袖 `LEADERS` 分頁均已接；原版 `.GAM` 全局領袖區塊讀寫、remake `ImportGAM` 與 `RawStatus=4`／30 回合清理均已接；`RawStatus=1` ETA 到期會刷新殖民地領袖增量與士氣 | `sub_E1D59`／`sub_DF8F0`／`sub_E2710` 的 raw 設計／帝國欄位與一般在職領袖旅行／任期逐值規則仍未知 |
| ~~完整外交畫面~~ **remake 核心已接** | 關係、提案、餽贈、特殊貿易、協議、終止、存檔與英文模板均已接 | 原版精確表與 AI 門檻列為 oracle 差異 |

### D. 需外部 oracle(我無法自給)
| 項目 | 需要 |
|---|---|
| 地面戰核心解算校驗 | `InvadeColony` 已接一條有部分 IDA 證據的 remake 路徑；`Ground_Combat_Round_ @ 0xEC4FE` 的四類型、亂數、傷亡游標與兵力回寫仍須閉合 record layout、caller setup、AI 裝甲與戰後人口 consumer。這是核心 parity 工作，不再列為低優先或發行非阻塞項。 |
| 飛彈速度分支 | 既有 raw ledger 已記錄 FST `+4` 與 `0x12/0x13/0x14` 的分支；本輪 IDA Pro `.i64` 交叉檢查把外部名稱／object-offset 對應的衝突另列於 [`oracle-static-ida-20260811.md`](../re/oracle-static-ida-20260811.md)，不影響 remake 已完成的 FST 垂直切片；不再把它當成本輪 remake 阻塞 |
| 逐畫面按鍵像素對齊 | 熱區多為估計座標 → **原版截圖**；議會／安塔蘭幀序已接，精確停留時間仍待 runtime 對照 |
| 戰術艦型 sprite 對照(#12) | `CMBTSHP` 已接 IDA 證實的 `45*playerColor+rawPicture`（raw 0..43；44 為 sentinel）與 `.GAM` raw picture；remake 已接移動後固定 tick timer 近似；只剩原版 20 幀停留需 **DOSBox oracle** |

### E. 低價值精修(不急,列此備查)
- 外交離散事件 `EventDelta`（已追回 `0x22D57` 極值排除／差平方權重與 `0x586D4` 0x200 反覆減半，但 remake 尚缺 raw event score record）、戰鬥／殖民地引擎爆炸連鎖傷害（已追回 `0x3868F`／`0x39985`／`0x40C2A`／`0x494A8` 純公式，raw record／flag 下游仍未知；已證實與隨機事件 8 無關）、alt-to-hit 命中變體；地面戰實機 global seed 校驗與事件 record 映射同列低優先精修。

## 三、建議順序(2026-08-10 收尾後)

1. **維持抽樣回歸**：新資料版本或 UI 改動後重跑四個高風險入口與 `-gamegallery`，不要求每輪完整遊戲測試。
2. **可選原版 oracle**：IDA Pro 靜態批次已追回敵方五級 blueprint raw writer／槽位、`sub_3AD57` 戰機命中／傷害式、相鄰 `sub_3AC20` 直接插值式、要塞 raw flag／seed／容量 divisor、外交比較常數與 `.GAM` 全局 `0x3B×0x43` 讀寫對稱，並記錄於 [`oracle-static-ida-20260811.md`](../re/oracle-static-ida-20260811.md)；若取得可啟動的 `VESA.COM`／DOSBox 環境，再只補兩個 raw 函式的外部名稱解析、live tech／逐彈逐槽 runtime 值、ARM/FST raw bit 正式名稱、外交特殊表、任命／任期與逐值差異，不回頭挖與遊戲無關的反組譯內部功能。
3. **外部音訊驗收**：在有音訊輸出的桌面逐曲聽 `STREAM`／`STREAMHD`，確認場景曲目與音量；Docker 技術檢查已完成。
4. **發行維護**：重新產出 Linux／Windows 包，保留 macOS CI universal 產物並做雜湊驗證；不把正版 LBX、音樂或未授權字型放進套件。

## 四、我的工作模式(已驗證有效,續用)
- 每塊帶**硬門檻**:先萃取手冊/openorion2 權威值,找到才建、找不到就停不准猜。
- 機械/移植/翻譯派 Sonnet subagent 實作,Opus 逐項對手冊核實+抓過度宣稱+docker build/vet/test 驗證+0 容器殘留。
- 每輪更新 `HONEST-STATUS.md`(清過期斷言)+ 本路線圖 + WORKLIST,推送 GitHub。
- 不做低價值 churn、不無方向建大 infra、不碰未驗證平衡——那不是真進度。
