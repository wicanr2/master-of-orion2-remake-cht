// spy.go 是「間諜(Spying)最小可玩迴圈」的殼層(shell)膠合層:把 gamedata/spy.go 已備妥的
// 機率公式(SpySlotBonus/SpyEffectiveThreshold/SpyRollChance/SpyVsSpy*,移植自手冊
// MANUAL_150.html「Notes on Spying」)接到活的對局狀態——訓練間諜、逐對手任務、每回合
// 諜報結算與 SpyVsSpy 判定。範圍與依據見 docs/tech/spy-system.md,重點摘要:
//
//   - 已做「偷科技(STEAL)」、「破壞(SABOTAGE)」與「隱匿(HIDE)」。HIDE 依手冊給的
//     SpyVsSpy +20 結算,不走偷科技判定；SABOTAGE 依原版 `0x1014A4` 的 70 門檻，
//     接到原版 `0x10130A` 已證實的「從殖民地已建建築中按建造成本加權抽一項並清除」
//     行為。IDA 已追回 raw `sub_1014A4` 的 packed relationship byte、三段式 slot helper、
//     亂數／兩張 score table 的使用位置與 70／90 分支；table 項目與上游欄位語意仍未命名，
//     因此 remake 以可保存的 AB／DB／E 近似完成玩家可感知判定，不宣稱 raw score parity。
//   - PlayerSpies 與 PlayerSpyMissions 逐 AI 對手平行保存、可存檔；原版三顆任務鈕的左右
//     語意尚未由反組譯確認,所以 remake 以明確標籤的 STEAL/SABOTAGE/HIDE 循環控制呈現
//     已證實的任務效果。
//   - 防禦方 Agent(手冊區分 Spy 攻擊 vs Agent 防守,各自累計 slot bonus)由
//     GameSession.DefensiveAgents / AIOpponent.DefensiveAgents 追蹤；成功的
//     Spy-vs-Spy 擊殺會真正扣除一名 Agent，舊存檔缺欄位時退回 0。
//   - SABOTAGE／STEAL 共用結構化 spyMissionScore，明列 slot、科技、政府、種族／
//     領袖與有效門檻；原版未命名 raw score record 仍標為未知，不把近似輸入冒充原版欄位。
//   - AI 的「已知科技」沿用既有 engine.PlayerState.CompletedTopics/ChosenTech——但 AI 目前
//     只有初始研究主題會被 RunResearchPhase 完成,advanceResearch()(推進到下一個未完成
//     主題)只接了玩家,AI 研究主題不會自動往下推進(既有限制,非本輪引入)。這代表 AI 可偷的
//     科技池長期而言很小,是誠實反映 AI 科技模型目前的抽象程度,不是假裝精確。
package shell

import (
	"fmt"
	"sort"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/ai"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// spyTrainCostBC 訓練 1 名間諜的 BC 成本。手冊(GAME_MANUAL.pdf p.70 Ships & Spies)只說明
// 間諜是透過殖民地建造佇列「訓練」出來的,像建築/艦艇一樣消耗產能與時間("Training a spy is
// unlike constructing a building or a ship, but it takes quite a lot of work..."),但沒有
// 給出具體成本數字(Item Info 面板才會顯示,原始資料不可得,本 remake 也還沒有殖民地佇列的
// 「間諜」建造選項)。這裡直接用 BC 簡化訓練流程(逐殖民地建造佇列整合留待完整 UI),成本量級
// 比照最低艦體(巡防艦 18 BC,見 session.go ShipCost)抓一個 remake 拍板值,不是手冊精確數字。
const spyTrainCostBC = 30

// spyMaxSlots 是手冊對每個對手的 Spies 與帝國共用 Defensive Agents 的上限。
// 這個上限是資料規則；訓練成本與 AI 的週期政策仍是 remake 近似。
const spyMaxSlots = 63

// spyMaintenancePerSpyBC 已由 Orion2.exe sub_1026CF @ 0x1026CF 與
// Compute_Player_Maintenance_ @ 0xE25B0..0xE25CF 證實：每個目標槽取低 6 位的間諜數後加總。
const spyMaintenancePerSpyBC = 1

func (s *GameSession) totalSpyMaintenance() int {
	total := 0
	for _, spies := range s.PlayerSpies {
		if spies > 0 {
			total += spies * spyMaintenancePerSpyBC
		}
	}
	return total
}

// spyMaxTopic 是 gamedata 研究主題的最大合法值(techtree.go researchChoices 陣列長度 83,
// 索引 0..82;TOPIC_HYPER_SOCIOLOGY=82 是常數表最後一項)。spyStealOptions 用它當迴圈上界,
// 避免呼叫 gamedata.ResearchChoiceFor 時索引越界 panic。
const spyMaxTopic = gamedata.TOPIC_HYPER_SOCIOLOGY

// TrainSpy 讓玩家花 spyTrainCostBC 訓練一名間諜派駐到 AIPlayers[targetIdx]。
// BC 不足或 targetIdx 越界回 false(不扣款、不增加間諜數)。
func (s *GameSession) TrainSpy(targetIdx int) bool {
	s.recordPlayerCommand(PlayerCommand{Name: CmdTrainSpy, Args: []int{targetIdx}})
	if targetIdx < 0 || targetIdx >= len(s.AIPlayers) {
		return false
	}
	if s.Player.BC < spyTrainCostBC {
		return false
	}
	s.ensurePlayerSpies()
	if s.PlayerSpies[targetIdx] >= spyMaxSlots {
		return false
	}
	s.Player.BC -= spyTrainCostBC
	s.PlayerSpies[targetIdx]++
	return true
}

// TrainDefensiveAgent 讓玩家花固定 BC 訓練一名帝國級防守 Agent。
// 原版完整建造佇列成本仍未追回；這個 API 沿用 TrainSpy 的 remake 成本尺度，
// 但把 Agent slot 與進攻間諜分開，讓防守加成真正進入每次任務結算。
func (s *GameSession) TrainDefensiveAgent() bool {
	s.recordPlayerCommand(PlayerCommand{Name: CmdTrainAgent})
	if s.Player.BC < spyTrainCostBC || s.DefensiveAgents >= spyMaxSlots {
		return false
	}
	s.Player.BC -= spyTrainCostBC
	s.DefensiveAgents++
	return true
}

// DismissDefensiveAgent 解除一名防守 Agent，不退款。
func (s *GameSession) DismissDefensiveAgent() bool {
	s.recordPlayerCommand(PlayerCommand{Name: CmdDismissAgent})
	if s.DefensiveAgents <= 0 {
		return false
	}
	s.DefensiveAgents--
	return true
}

// ensurePlayerSpies 確保 PlayerSpies 長度跟上 AIPlayers(新對局/AI 數變動時延遲初始化,
// 比照 popAccum 的既有 lazy-init 慣例,見 advancePopulation)。
func (s *GameSession) ensurePlayerSpies() {
	for len(s.PlayerSpies) < len(s.AIPlayers) {
		s.PlayerSpies = append(s.PlayerSpies, 0)
	}
	s.ensurePlayerSpyMissions()
}

// psKnowsTech 判定 ps 是否「知道」某個特定 Technology(隸屬 topic)。規則與
// ground_invasion.go 的 componentUnlockedFor/groundEquipTechOwned 完全一致(主題已完成、
// 但未明確抉擇 → 視為該主題全部選項皆解鎖;已明確抉擇 → 僅所選項),只是脫離「元件」語境,
// 供間諜偷科技判定「對方已知、我方未知」時共用同一套主題/抉擇規則,不另立一套邏輯。
func psKnowsTech(ps engine.PlayerState, topic gamedata.ResearchTopic, tech gamedata.Technology) bool {
	return playerStateKnowsTech(ps, topic, tech)
}

// spyStealOption 是一個「可偷」的科技候選:defender 已知、attacker 未知。
type spyStealOption struct {
	Topic gamedata.ResearchTopic
	Tech  gamedata.Technology
}

// spyStealOptions 列出 attacker 可從 defender 偷到的科技(defender 已知、attacker 未知),
// 依 Topic、Tech 由小到大排序(可重現,供固定 rng 挑選索引的單元測試)。
//
// 手冊依據(GAME_MANUAL.pdf p.174-175 Espionage 段):「your Spy goes off into the colonies
// of another race...tries to steal technologies you have yet to gain」——偷來的科技必須是
// 攻擊方尚未擁有的;間諜是潛入對方殖民地行動,邏輯上只能偷到對方已經擁有的科技,故用
// defender.CompletedTopics 當來源池,這正是硬門檻查核時確認的「偷對方已知、我方未知」規則。
// TOPIC_STARTING_TECH 這類 Cost=0、無 Choices 的填充主題(見 techtree.go researchChoices)
// 不是「科技」,略過。
func spyStealOptions(attacker, defender engine.PlayerState) []spyStealOption {
	var out []spyStealOption
	if defender.CompletedTopics == nil && defender.GrantedTechs == nil {
		return nil
	}
	for topic := gamedata.ResearchTopic(0); topic <= spyMaxTopic; topic++ {
		if !defender.CompletedTopics[topic] {
			continue
		}
		choice := gamedata.ResearchChoiceFor(topic)
		if len(choice.Choices) == 0 {
			continue // 填充主題(如起始科技),無科技可偷
		}
		if choice.ResearchAll {
			for _, tech := range choice.Choices {
				if !psKnowsTech(attacker, topic, tech) {
					out = append(out, spyStealOption{Topic: topic, Tech: tech})
				}
			}
			continue
		}
		tech, ok := defender.ChosenTech[topic]
		if !ok {
			continue
		}
		if !psKnowsTech(attacker, topic, tech) {
			out = append(out, spyStealOption{Topic: topic, Tech: tech})
		}
	}
	for tech, granted := range defender.GrantedTechs {
		if !granted {
			continue
		}
		topic, ok := gamedata.OrigTechTopic(tech)
		if ok && !psKnowsTech(attacker, topic, tech) {
			out = append(out, spyStealOption{Topic: topic, Tech: tech})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Topic != out[j].Topic {
			return out[i].Topic < out[j].Topic
		}
		return out[i].Tech < out[j].Tech
	})
	if len(out) > 1 {
		unique := out[:1]
		for _, option := range out[1:] {
			last := unique[len(unique)-1]
			if option != last {
				unique = append(unique, option)
			}
		}
		out = unique
	}
	return out
}

// applyTechTheft 讓 attacker 偷到 opt 這項科技。首次取得該 Topic 時保存成主要明確選擇；
// 若同主題已有另一項，改存 GrantedTechs 而不覆蓋舊科技。授予後共用原版特殊 callback。
//
// 已知限制(不修正,記錄於此):若 opt.Topic 剛好是 attacker 正在研究的 ResearchTopic,偷竊會
// 直接把該主題標記完成,但不動 ResearchProgress——已經投入該主題的研究點數會變成「投給一個
// 已經完成的主題」而無處可去,下回合 advanceResearch() 會把主題推進到下一項,那些點數就此浪費。
// 這是最小迴圈的邊界情況,影響有限(只在偷到「正好在研究中」的科技時發生),故不在本輪額外處理。
func applyTechTheft(ps *engine.PlayerState, opt spyStealOption) {
	grantTechnologyApplication(ps, opt.Topic, opt.Tech)
}

// spyAttackerBonus 算出「攻擊方(派間諜出去偷科技的一方)」的 attacker bonus(AB,見
// gamedata.SpyEffectiveThreshold 的定義)。目前只接上 SpySlotBonus——手冊 Spy Bonuses 表中
// 唯一「有明確人數 → 加成對照表」的項目(見 gamedata/spy.go 檔頭)。
//
// ⚠ **2026-08-08(第 58 項(擋門理由過期三個月))訂正過。** 這段原本寫著三項加成一律回 0,理由是
// 「無種族間諜特性強度資料、**無逐科技模型可查是否擁有 spy.go 列的 5 項科技**、
// AIOpponent 無政府型態欄位」。
//
// 中間那條**當時可能成立,現在不成立**:`groundEquipTechOwned` 已經是三個系統共用的
// 判定(生物武器、地面裝備、進階政體),而那 5 項科技在 `enums.go` 都有常數。
// 科技加成因此接上了,攻守兩側都算(手冊那張表兩欄同值)。
//
// 仍然回 0 的兩項,理由各自不同:
//   - **種族間諜特性**:`enums.go` 的 `TRAIT_SPYING` 只標記「有沒有」,沒有 -3/+3/+6 的
//     強度分級,而 `Races` 表也沒有這一欄。**資料不存在**,不是沒接。
//   - **政府**:手冊只給 Defense 欄(攻擊方本來就沒有政府加成),所以這裡不需要它。
//     防守側見 `spyDefenderBonus`。
func spyAttackerBonus(ps engine.PlayerState, spyCount, raceBonus int) int {
	return gamedata.SpySlotBonus(spyCount) + spyTechBonusFor(ps) + raceBonus
}

// spyTechBonusFor 加總這一方已擁有的間諜相關科技加成(手冊 Spy Bonuses 表的 Technology 列)。
//
// ⚠ **「加總」是這裡的讀法,手冊沒有明說。** 手冊列的是 5 項**互不相關**的科技
// (神經掃描器、隱形衣、心靈學…),不是同一件事的三個階——取最佳的話,研究第二項就完全
// 沒有意義。這與領袖技能那邊的處理刻意不同:那裡手冊明寫「Megawealth 與 Researcher 是
// 累加的」,反面暗示其餘取最佳(見 gamedata/leader_skill_apply.go)。這裡沒有那句話。
func spyTechBonusFor(ps engine.PlayerState) int {
	total := 0
	for _, tech := range []gamedata.Technology{
		gamedata.TECH_NEURAL_SCANNER, gamedata.TECH_TELEPATHIC_TRAINING,
		gamedata.TECH_CYBERSECURITY_LINK, gamedata.TECH_STEALTH_SUIT, gamedata.TECH_PSIONICS,
	} {
		topic, ok := gamedata.OrigTechTopic(tech)
		if !ok || !groundEquipTechOwned(ps, topic, tech) {
			continue
		}
		total += gamedata.SpyTechnologyBonus(tech)
	}
	return total
}

// spyDefenderBonus 算出「防守方(被偷科技的一方)」的 defender bonus(DB)。手冊區分 Spy
// (攻擊,逐對手指派)與 Agent(防守,不分對手、全體共用)兩種 slot；公開的舊 wrapper
// 保留 0 Agent 語意，實際任務一律走 spyDefenderBonusWithAgents／spyMissionScore。
//
// ⚠ **2026-08-08(第 58 項(擋門理由過期三個月))起不再恆為 0。** Agent 人數、科技與政府
// 都可接上:
//   - **科技**:`spyTechBonusFor`,攻守兩側同一套(手冊那張表兩欄同值)。
//   - **政府**:`gamedata.SpyGovernmentDefenseBonus`,手冊只給 Defense 欄。
//     govBonus 由呼叫端算好傳入——**只有玩家有政府型態**,`AIOpponent` 沒有這個欄位
//     (原版是 `[player+0x89F]`,見第 54 項(三個寫入端)),所以 AI 當防守方時呼叫端傳 0。
//     那是資料模型的缺口,不是規則沒接。
func spyDefenderBonus(ps engine.PlayerState, govBonus, raceBonus int) int {
	return spyDefenderBonusWithAgents(ps, govBonus, raceBonus, 0)
}

// playerSpyGovernmentDefenseBonus 回傳玩家目前政府型態的防諜加成。
//
// 走 `assimilationGovernment()` 而不是直接看 `s.Government`:那支會把「研究出進階政體科技」
// 算進去(邦聯/帝國/聯邦/銀河統一),而手冊的防諜表對基本型與進階型給的是**不同的值**
// ——只看 s.Government 會讓研究出帝國的獨裁玩家永遠拿獨裁那一格。
//
// 兩個列舉的編號相同不是巧合:原版只有一個 `[player+0x89F]` 欄位(第 54 項(三個寫入端)),
// Go 這邊分成好幾個列舉是歷史。`spy_government_test.go` 把這件事釘住。
func (s *GameSession) playerSpyGovernmentDefenseBonus() int {
	return gamedata.SpyGovernmentDefenseBonus(gamedata.SpyGovernmentType(s.assimilationGovernment()))
}

// spyVsSpyOutcome 是 SpyVsSpy(間諜互殺)判定結果。
type spyVsSpyOutcome struct {
	AttackerKilled bool
	DefenderKilled bool
}

// resolveSpyVsSpy 用 gamedata.SpyVsSpyAttackerBonus/SpyVsSpyDefenderBonus 算出雙方淨值後,
// 直接比較手冊給出的擊殺門檻(±80,見 gamedata.SpyVsSpyDefenderKillThreshold/
// SpyVsSpyAttackerKillThreshold)。
//
// 手冊原文(MANUAL_150.html Spy vs Spy):「At +80 a defender is killed, and at -80 an
// attacker[is killed]」——只給了淨值門檻,沒給 SpyRollChance 那套「T + 骰子」機率公式的
// 對應版本(gamedata/spy.go 檔頭已標 TODO:T 基準值不明)。故這裡忠實只做「淨值是否跨過 ±80
// 門檻」的確定性判定,不臆造機率或 lucky roll 加成(手冊提到 lucky roll 也能在門檻內造成
// 擊殺,此簡化模型不含,TODO)。
//
// 現行 remake 的 ab/db 含 SpySlotBonus 與已接上的科技、玩家政府、玩家種族間諜加成；
// AI 的種族／政府與防守 Agent 數量仍沒有資料欄位,故對 AI 防守側相應部分傳 0。
// 單元測試仍用直接構造的 ab/db 數值驗證 ±80 門檻與 HIDE +20,避免把目前簡化的
// 正常遊戲數值誤當成原版完整 Agent 模型。
func resolveSpyVsSpy(ab, db int, attackerHide bool) spyVsSpyOutcome {
	attackerB := gamedata.SpyVsSpyAttackerBonus(ab, attackerHide)
	defenderB := gamedata.SpyVsSpyDefenderBonus(db)
	net := attackerB - defenderB
	var out spyVsSpyOutcome
	if net >= gamedata.SpyVsSpyDefenderKillThreshold {
		out.DefenderKilled = true
	}
	if net <= gamedata.SpyVsSpyAttackerKillThreshold {
		out.AttackerKilled = true
	}
	return out
}

// spySabotageCandidate 是原版 SABOTAGE 抽選池的一個可重現候選。
//
// 原版 `Steal_App` @ 0x10130A 逐殖民地掃描 49 個建築槽，使用原版建築表
// `off_17EB3D` 每列 +8 的建造成本作為權重；Go 的 ColonyBuildings 是 map，不能直接
// 依 map 迭代順序抽選，故保存原版編號並在抽選前排序。`OriginalID != 9` 是原版函式
// `v6 != 9` 的已證實保留槽位；`OrigBuildingID` 的對照仍保留其原始定位，不在此改名。
type spySabotageCandidate struct {
	colonyIdx  int
	originalID int
	name       string
	weight     int
}

// spySabotageCandidates 建立 SABOTAGE 的候選池。
//
// 已知中文建築名但尚未能對回原版 49 槽的 map key 會被保守略過；這避免把 remake
// 特殊行動或未知槽位誤當成原版 `Steal_App` 可以清除的建築。這個保守轉譯是強推論，
// 不代表原版未知槽位一定不能被破壞。
func spySabotageCandidates(colonyBuildings []map[string]bool) []spySabotageCandidate {
	var candidates []spySabotageCandidate
	for colonyIdx, built := range colonyBuildings {
		if len(built) == 0 {
			continue
		}
		for _, building := range gamedata.Buildings {
			if !built[building.NameZH] {
				continue
			}
			originalID, ok := gamedata.OrigBuildingID[building.NameEN]
			weight, weightOK := gamedata.OriginalBuildingProductionCost(originalID)
			if !ok || !weightOK || originalID == 9 {
				continue
			}
			candidates = append(candidates, spySabotageCandidate{
				colonyIdx:  colonyIdx,
				originalID: originalID,
				name:       building.NameZH,
				weight:     weight,
			})
		}
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].colonyIdx != candidates[j].colonyIdx {
			return candidates[i].colonyIdx < candidates[j].colonyIdx
		}
		if candidates[i].originalID != candidates[j].originalID {
			return candidates[i].originalID < candidates[j].originalID
		}
		return candidates[i].name < candidates[j].name
	})
	return candidates
}

// sabotageRandomBuilding 依原版 SABOTAGE 的已證實資料結構選取並移除一棟建築。
//
// 原版 `0x10130A` 的 `toggle_flag(total)` 亂數細節尚未完全解出；此處只將已證實的
// 建造成本權重映射到可保存、可測試的 Intn(total) 抽選。回傳殖民地索引、中文建築名
// 與是否真的移除；沒有可破壞候選時不改任何 map。
func sabotageRandomBuilding(rng rollSource, colonyBuildings []map[string]bool) (int, string, bool) {
	candidates := spySabotageCandidates(colonyBuildings)
	totalWeight := 0
	for _, candidate := range candidates {
		totalWeight += candidate.weight
	}
	if totalWeight <= 0 {
		return 0, "", false
	}
	pick := rng.Intn(totalWeight)
	for _, candidate := range candidates {
		if pick < candidate.weight {
			delete(colonyBuildings[candidate.colonyIdx], candidate.name)
			return candidate.colonyIdx, candidate.name, true
		}
		pick -= candidate.weight
	}
	return 0, "", false
}

// spyStealAttempt 對單一方向(attacker 派 spyCount 個間諜偷 defender 的科技)跑一次 STEAL +
// SpyVsSpy 判定,回傳要記進 LastEspionage 的訊息(可能 0~2 則)、attacker 間諜是否被擊殺、
// 以及偷到的科技是否套用到了 *attackerPS(呼叫端已把 attackerPS 指到正確的 engine.PlayerState)。
// attackerName/defenderName 純供訊息文字使用。
// rollSource 是「會擲骰的東西」——`*randStream`(可存檔的流,見 randstream.go)與
// `*rand.Rand`(測試裡直接餵的臨時流)都滿足它。收介面而不是收具體型別,
// 是為了讓既有測試不必跟著換掉。
type rollSource interface {
	Intn(n int) int
	Float64() float64
}

func spyMissionAttempt(rng rollSource, mission SpyMission, attackerPS *engine.PlayerState, defenderPS engine.PlayerState,
	spyCount int, attackerName, defenderName string,
	defenderGovBonus, attackerRaceBonus, defenderRaceBonus int) (messages []string, attackerSpyKilled bool) {
	return spyMissionAttemptWithBuildings(rng, mission, attackerPS, defenderPS, spyCount,
		attackerName, defenderName, defenderGovBonus, attackerRaceBonus, defenderRaceBonus, nil)
}

// spyMissionResult 是一次任務的結算結果；DefenderAgentKilled 讓呼叫端能真正
// 消費防守 Agent，而不是只在摘要顯示「擊殺」文字。
type spyMissionResult struct {
	Messages            []string
	AttackerSpyKilled   bool
	DefenderAgentKilled bool
	TechStolen          bool
	Score               spyMissionScore
}

// spyMissionAttemptWithAgents 是含防守 Agent 數量的相容任務入口。保留舊回傳形狀，
// 新的 runtime 走下方 result 版本取得完整分數與 Agent 消費結果。
func spyMissionAttemptWithAgents(rng rollSource, mission SpyMission, attackerPS *engine.PlayerState, defenderPS engine.PlayerState,
	spyCount int, attackerName, defenderName string,
	defenderGovBonus, attackerRaceBonus, defenderRaceBonus, defenderAgents int,
	defenderBuildings []map[string]bool) (messages []string, attackerSpyKilled bool) {
	result := spyMissionAttemptWithAgentsResult(rng, mission, attackerPS, defenderPS, spyCount,
		attackerName, defenderName, defenderGovBonus, attackerRaceBonus, defenderRaceBonus,
		defenderAgents, defenderBuildings)
	return result.Messages, result.AttackerSpyKilled
}

func spyMissionAttemptWithAgentsResult(rng rollSource, mission SpyMission, attackerPS *engine.PlayerState, defenderPS engine.PlayerState,
	spyCount int, attackerName, defenderName string,
	defenderGovBonus, attackerRaceBonus, defenderRaceBonus, defenderAgents int,
	defenderBuildings []map[string]bool) spyMissionResult {
	if attackerPS == nil {
		return spyMissionResult{}
	}
	result := spyMissionResult{Score: calculateSpyMissionScore(mission, *attackerPS, defenderPS,
		spyCount, defenderAgents, defenderGovBonus, attackerRaceBonus, defenderRaceBonus)}
	mission = result.Score.Mission
	if mission == SpyMissionHide {
		result.Messages = append(result.Messages, fmt.Sprintf(
			"%s 的間諜在 %s 執行隱匿任務", attackerName, defenderName))
		outcome := resolveSpyVsSpy(result.Score.AttackerBonus, result.Score.DefenderBonus, true)
		if outcome.AttackerKilled {
			result.AttackerSpyKilled = true
			result.Messages = append(result.Messages, fmt.Sprintf("%s 的一名間諜在 %s 被反間諜擊殺", attackerName, defenderName))
		}
		if outcome.DefenderKilled && defenderAgents > 0 {
			result.DefenderAgentKilled = true
			result.Messages = append(result.Messages, fmt.Sprintf(
				"%s 的隱匿間諜在 %s 的 Spy vs Spy 判定中擊殺一名防守 Agent", attackerName, defenderName))
		}
		return result
	}

	if rng.Float64() < result.Score.SuccessChance {
		if mission == SpyMissionSabotage {
			colonyIdx, building, ok := sabotageRandomBuilding(rng, defenderBuildings)
			if !ok {
				result.Messages = append(result.Messages, fmt.Sprintf(
					"%s 的間諜潛入 %s 得手,但沒有可破壞的已建建築", attackerName, defenderName))
			} else {
				result.Messages = append(result.Messages, fmt.Sprintf(
					"%s 的間諜在 %s 的第 %d 殖民地破壞了%s", attackerName, defenderName, colonyIdx+1, building))
			}
		} else {
			opts := spyStealOptions(*attackerPS, defenderPS)
			if len(opts) == 0 {
				result.Messages = append(result.Messages, fmt.Sprintf(
					"%s 的間諜潛入 %s 得手,但對方已無%s尚未擁有的科技可偷", attackerName, defenderName, attackerName))
			} else {
				pick := opts[rng.Intn(len(opts))]
				applyTechTheft(attackerPS, pick)
				result.TechStolen = true
				result.Messages = append(result.Messages, fmt.Sprintf(
					"%s 的間諜從 %s 偷得科技:%s", attackerName, defenderName, gamedata.TechnologyName(pick.Tech)))
			}
		}
	}

	outcome := resolveSpyVsSpy(result.Score.AttackerBonus, result.Score.DefenderBonus, false)
	if outcome.AttackerKilled {
		result.AttackerSpyKilled = true
		result.Messages = append(result.Messages, fmt.Sprintf("%s 的一名間諜在 %s 被反間諜擊殺", attackerName, defenderName))
	}
	if outcome.DefenderKilled && defenderAgents > 0 {
		result.DefenderAgentKilled = true
		result.Messages = append(result.Messages, fmt.Sprintf("%s 的間諜在 %s 擊殺一名防守 Agent", attackerName, defenderName))
	}
	return result
}

// spyMissionAttemptWithBuildings 是可選殖民地建築狀態的任務結算。
//
// 保留不帶建築參數的 spyMissionAttempt，讓既有測試與 AI 的 STEAL 預設保持相容；玩家
// 對 AI 的 SABOTAGE 呼叫端才傳入 defenderBuildings。SABOTAGE 的直接刪除效果已由原版
// `Add_Building` @ 0x145EA 的 `buildingFlags = 0` 寫入證實，命中率沿用手冊公式與
// 原版 `0x1014A4` 的 action threshold=70；raw score 的 table／亂數語意仍留在
// oracle 層，remake 不為了補一個不可查證的欄位而改壞可重播的 AB／DB／E 模型。
func spyMissionAttemptWithBuildings(rng rollSource, mission SpyMission, attackerPS *engine.PlayerState, defenderPS engine.PlayerState,
	spyCount int, attackerName, defenderName string,
	defenderGovBonus, attackerRaceBonus, defenderRaceBonus int,
	defenderBuildings []map[string]bool) (messages []string, attackerSpyKilled bool) {
	return spyMissionAttemptWithAgents(rng, mission, attackerPS, defenderPS, spyCount,
		attackerName, defenderName, defenderGovBonus, attackerRaceBonus, defenderRaceBonus, 0, defenderBuildings)
}

// spyMissionAttemptWithBuildingsLegacy 是舊實作的名稱保留點，避免未來需要比較
// 0 Agent 與有 Agent 的結果時把兩份任務 switch 再複製一次。
func spyMissionAttemptWithBuildingsLegacy(rng rollSource, mission SpyMission, attackerPS *engine.PlayerState, defenderPS engine.PlayerState,
	spyCount int, attackerName, defenderName string,
	defenderGovBonus, attackerRaceBonus, defenderRaceBonus int,
	defenderBuildings []map[string]bool) (messages []string, attackerSpyKilled bool) {
	return spyMissionAttemptWithAgents(rng, mission, attackerPS, defenderPS, spyCount,
		attackerName, defenderName, defenderGovBonus, attackerRaceBonus, defenderRaceBonus, 0, defenderBuildings)
}

// spyStealAttempt 保留原本測試與呼叫端的 STEAL 封裝。
func spyStealAttempt(rng rollSource, attackerPS *engine.PlayerState, defenderPS engine.PlayerState,
	spyCount int, attackerName, defenderName string,
	defenderGovBonus, attackerRaceBonus, defenderRaceBonus int) (messages []string, attackerSpyKilled bool) {
	return spyMissionAttempt(rng, SpyMissionSteal, attackerPS, defenderPS, spyCount,
		attackerName, defenderName, defenderGovBonus, attackerRaceBonus, defenderRaceBonus)
}

func spyDefenderBonusWithAgents(ps engine.PlayerState, govBonus, raceBonus, agentCount int) int {
	return gamedata.SpySlotBonus(agentCount) + spyTechBonusFor(ps) + govBonus + raceBonus
}

// spyMissionScore 是一次諜報／SABOTAGE 判定的完整分數拆解。
//
// 原版 `sub_100A83` 先把科技、種族、心靈感應、最佳領袖與政府寫成每帝國攻防兩表；
// `sub_1014A4` 再於逐對手任務加入 Spies／Agents slot helper。remake 不需要保存短命的
// runtime 表，但此結構保留同一分層，讓每一項只計一次。
type spyMissionScore struct {
	Mission                 SpyMission
	BaseThreshold           int
	AttackerSpies           int
	AttackerSlotBonus       int
	AttackerTechnologyBonus int
	AttackerRaceLeaderBonus int
	AttackerBonus           int
	DefenderAgents          int
	DefenderAgentSlotBonus  int
	DefenderTechnologyBonus int
	DefenderGovernmentBonus int
	DefenderRaceLeaderBonus int
	DefenderBonus           int
	EffectiveThreshold      int
	SuccessChance           float64
}

func spyMissionBaseThreshold(mission SpyMission) int {
	switch normalizedSpyMission(mission) {
	case SpyMissionSabotage:
		return gamedata.SpyThresholdSabotage
	case SpyMissionSteal:
		return gamedata.SpyThresholdSteal
	default:
		return 0
	}
}

// calculateSpyMissionScore 把 AB／DB 的每一個來源攤平，再計算 E／p。
// attackerRaceLeaderBonus 與 defenderRaceLeaderBonus 是呼叫端已合併的種族＋
// 領袖 bonus；這樣可保留既有測試入口的相容簽名，同時讓 runtime 分數不再是黑箱。
func calculateSpyMissionScore(mission SpyMission, attackerPS, defenderPS engine.PlayerState,
	spyCount, defenderAgents, defenderGovBonus, attackerRaceLeaderBonus, defenderRaceLeaderBonus int) spyMissionScore {
	mission = normalizedSpyMission(mission)
	attackerSlot := gamedata.SpySlotBonus(spyCount)
	attackerTech := spyTechBonusFor(attackerPS)
	defenderSlot := gamedata.SpySlotBonus(defenderAgents)
	defenderTech := spyTechBonusFor(defenderPS)
	threshold := spyMissionBaseThreshold(mission)
	score := spyMissionScore{
		Mission: mission, BaseThreshold: threshold,
		AttackerSpies: spyCount, AttackerSlotBonus: attackerSlot,
		AttackerTechnologyBonus: attackerTech, AttackerRaceLeaderBonus: attackerRaceLeaderBonus,
		DefenderAgents: defenderAgents, DefenderAgentSlotBonus: defenderSlot,
		DefenderTechnologyBonus: defenderTech, DefenderGovernmentBonus: defenderGovBonus,
		DefenderRaceLeaderBonus: defenderRaceLeaderBonus,
	}
	// 呼叫端已把原版共同基底中的種族／心靈感應與最佳領袖壓成各方向 bonus；
	// 透過同一個原版兩表 adapter 組合科技與政府，再於消費端加入 slot。
	attackEmpire, _ := gamedata.OriginalSpyEmpireBonuses(attackerTech, attackerRaceLeaderBonus, 0, 0, 0, 0)
	_, defenseEmpire := gamedata.OriginalSpyEmpireBonuses(defenderTech, defenderRaceLeaderBonus, 0, 0, 0, defenderGovBonus)
	score.AttackerBonus = attackerSlot + attackEmpire
	score.DefenderBonus = defenderSlot + defenseEmpire
	if threshold > 0 {
		score.EffectiveThreshold = gamedata.SpyEffectiveThreshold(threshold, score.DefenderBonus, score.AttackerBonus)
		score.SuccessChance = gamedata.SpyRollChance(score.EffectiveThreshold)
	}
	return score
}

// aiSpyMission 是 remake 的 AI 任務政策：原版 AI 的完整間諜／防守策略尚未由
// oracle 還原，所以只把已存在的 personality 差異接到可玩的任務效果，不宣稱這是
// 原版逐格策略。冷酷、好戰、排外會優先破壞；反覆無常每回合交替；重信譽、和平與
// 失信仍偷科技。
func aiSpyMission(personality ai.Personality, turn int) SpyMission {
	switch personality {
	case ai.PersonalityXenophobic, ai.PersonalityRuthless, ai.PersonalityAggressive:
		return SpyMissionSabotage
	case ai.PersonalityErratic:
		if turn%2 == 0 {
			return SpyMissionSabotage
		}
	}
	return SpyMissionSteal
}

// advanceEspionage 每回合結算玩家 ↔ 各 AI 對手之間的間諜行動。玩家側依
// PlayerSpyMissions 執行 STEAL/SABOTAGE/HIDE；AI 依 aiSpyMission 選擇 STEAL 或 SABOTAGE。
// 呼叫時機:EndTurn 已完成玩家與所有 AI 本回合的研究/經濟結算之後,讓「偷到的科技」
// 判定用的是本回合最新的 CompletedTopics/ChosenTech(見 EndTurn 呼叫點註解)。
func (s *GameSession) advanceEspionage() {
	s.LastEspionage = nil
	s.ensurePlayerSpies()
	if s.spyRand == nil {
		s.spyRand = newRandStream(s.EventSeed*2654435761 + 7)
	}
	// 刺客是每位領袖各自擲一次的獨立行動，必須排在任務前，讓本回合
	// 被刺殺的防守 Agent 立即影響後續的 Spy／Agent 判定。
	s.advanceLeaderAssassinActions()

	for i := range s.AIPlayers {
		a := &s.AIPlayers[i]

		// 玩家 → AI:依逐對手任務執行。
		if s.PlayerSpies[i] > 0 {
			result := spyMissionAttemptWithAgentsResult(s.spyRand, s.SpyMissionFor(i), &s.Player, a.Player,
				s.PlayerSpies[i], "我方", a.Name, 0,
				s.raceSpyBonusForActions()+leaderEmpireSkillBonus(s.Leaders, gamedata.SKILL_SPYMASTER),
				aiRaceSpyBonus(*a)+leaderEmpireSkillBonus(a.Leaders, gamedata.SKILL_TELEPATH),
				a.DefensiveAgents, a.ColonyBuildings)
			s.LastEspionage = append(s.LastEspionage, result.Messages...)
			if result.AttackerSpyKilled && s.PlayerSpies[i] > 0 {
				s.PlayerSpies[i]--
			}
			if result.DefenderAgentKilled && a.DefensiveAgents > 0 {
				a.DefensiveAgents--
			}
			if result.TechStolen {
				s.UpdatePlayerShipDesignsAfterTech()
			}
		}

		// AI → 玩家:依性格執行 STEAL/SABOTAGE + SpyVsSpy(對稱處理;AI 已知科技集長期而言
		// 很小,見檔頭說明)。玩家的 ColonyBuildings 是 SABOTAGE 的防守目標。
		if a.Spies > 0 {
			// 玩家當防守方:政府加成算得出來。
			result := spyMissionAttemptWithAgentsResult(s.spyRand, aiSpyMission(a.Personality, s.Turn), &a.Player, s.Player, a.Spies, a.Name, "我方",
				s.playerSpyGovernmentDefenseBonus(),
				aiRaceSpyBonus(*a)+leaderEmpireSkillBonus(a.Leaders, gamedata.SKILL_SPYMASTER),
				s.raceSpyBonusForActions()+leaderEmpireSkillBonus(s.Leaders, gamedata.SKILL_TELEPATH),
				s.DefensiveAgents, s.ColonyBuildings)
			s.LastEspionage = append(s.LastEspionage, result.Messages...)
			if result.AttackerSpyKilled && a.Spies > 0 {
				a.Spies--
			}
			if result.DefenderAgentKilled && s.DefensiveAgents > 0 {
				s.DefensiveAgents--
			}
		}
	}
}
