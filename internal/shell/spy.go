// spy.go 是「間諜(Spying)最小可玩迴圈」的殼層(shell)膠合層:把 gamedata/spy.go 已備妥的
// 機率公式(SpySlotBonus/SpyEffectiveThreshold/SpyRollChance/SpyVsSpy*,移植自手冊
// MANUAL_150.html「Notes on Spying」)接到活的對局狀態——訓練間諜、每回合諜報結算(偷科技)、
// SpyVsSpy 判定。範圍與依據見 docs/tech/spy-system.md,重點摘要:
//
//   - 只做「偷科技(STEAL)」,不做破壞(SABOTAGE)——手冊(GAME_MANUAL.pdf p.174-175
//     Espionage 段)只定性描述破壞效果「destroy some valuable piece of enemy property」,
//     沒給破壞對象/數值規則,標 TODO 不臆測。
//   - 逐對手分配 Espionage/Sabotage/Hide 任務選單延後,最小迴圈預設所有間諜對單一 AI 對手
//     做 STEAL(PlayerSpies 陣列結構本身已支援逐對手分配,只是目前唯一一個 AI 對手看不出差異)。
//   - 防禦方 Agent(手冊區分 Spy 攻擊 vs Agent 防守,各自累計 slot bonus)不獨立追蹤——
//     用 DB=0,這正好對應手冊原文「defenses against enemy spies are active...even with zero
//     defending agents」描述的「零 Agent」情境,不是遺漏。
//   - 種族/科技/政府 bonus 現行無對應資料(AIOpponent 無種族/政府欄位、無逐科技模型可查是否
//     擁有 spy.go 列的 5 項科技)→ 一律 0,見 spyAttackerBonus/spyDefenderBonus 註解,TODO。
//   - AI 的「已知科技」沿用既有 engine.PlayerState.CompletedTopics/ChosenTech——但 AI 目前
//     只有初始研究主題會被 RunResearchPhase 完成,advanceResearch()(推進到下一個未完成
//     主題)只接了玩家,AI 研究主題不會自動往下推進(既有限制,非本輪引入)。這代表 AI 可偷的
//     科技池長期而言很小,是誠實反映 AI 科技模型目前的抽象程度,不是假裝精確。
package shell

import (
	"fmt"
	"math/rand"

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

// spyMaintenancePerSpyBC 每個已訓練間諜每回合的維護費(BC)。
// engine.PlayerState.Maintenance 欄位註解已載明「間諜維護費本專案尚無可推導模型」——這裡給
// remake 佔位值,刻意不併入 totalBuildingMaintenance/Player.Maintenance(避免牽動既有經濟
// 測試的既定假設),改在 advanceEspionage 直接從 BC 扣。預設 0 間諜時扣款為 0,不影響任何
// 既有對局/測試。
const spyMaintenancePerSpyBC = 1

// spyMaxTopic 是 gamedata 研究主題的最大合法值(techtree.go researchChoices 陣列長度 83,
// 索引 0..82;TOPIC_HYPER_SOCIOLOGY=82 是常數表最後一項)。spyStealOptions 用它當迴圈上界,
// 避免呼叫 gamedata.ResearchChoiceFor 時索引越界 panic。
const spyMaxTopic = gamedata.TOPIC_HYPER_SOCIOLOGY

// TrainSpy 讓玩家花 spyTrainCostBC 訓練一名間諜派駐到 AIPlayers[targetIdx]。
// BC 不足或 targetIdx 越界回 false(不扣款、不增加間諜數)。
func (s *GameSession) TrainSpy(targetIdx int) bool {
	if targetIdx < 0 || targetIdx >= len(s.AIPlayers) {
		return false
	}
	if s.Player.BC < spyTrainCostBC {
		return false
	}
	s.ensurePlayerSpies()
	s.Player.BC -= spyTrainCostBC
	s.PlayerSpies[targetIdx]++
	return true
}

// ensurePlayerSpies 確保 PlayerSpies 長度跟上 AIPlayers(新對局/AI 數變動時延遲初始化,
// 比照 popAccum 的既有 lazy-init 慣例,見 advancePopulation)。
func (s *GameSession) ensurePlayerSpies() {
	for len(s.PlayerSpies) < len(s.AIPlayers) {
		s.PlayerSpies = append(s.PlayerSpies, 0)
	}
}

// psKnowsTech 判定 ps 是否「知道」某個特定 Technology(隸屬 topic)。規則與
// ground_invasion.go 的 componentUnlockedFor/groundEquipTechOwned 完全一致(主題已完成、
// 但未明確抉擇 → 視為該主題全部選項皆解鎖;已明確抉擇 → 僅所選項),只是脫離「元件」語境,
// 供間諜偷科技判定「對方已知、我方未知」時共用同一套主題/抉擇規則,不另立一套邏輯。
func psKnowsTech(ps engine.PlayerState, topic gamedata.ResearchTopic, tech gamedata.Technology) bool {
	if ps.CompletedTopics == nil || !ps.CompletedTopics[topic] {
		return false
	}
	if ps.ExplicitChoice == nil || !ps.ExplicitChoice[topic] {
		return true
	}
	return ps.ChosenTech != nil && ps.ChosenTech[topic] == tech
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
	if defender.CompletedTopics == nil {
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
	return out
}

// applyTechTheft 讓 attacker 偷到 opt 這項科技:標記該 Topic 為已完成、ChosenTech 記入偷到的
// 那一項、並標記 ExplicitChoice=true。語意比照 engine.ApplyResearchChoice「明確抉擇」:偷來的
// 只有那一項生效,不會像研究完成 ResearchAll/未明確抉擇主題時一樣讓同主題其餘選項也跟著解鎖
// (componentUnlockedFor 的判定規則,見 ground_invasion.go)。
//
// 已知限制(不修正,記錄於此):若 opt.Topic 剛好是 attacker 正在研究的 ResearchTopic,偷竊會
// 直接把該主題標記完成,但不動 ResearchProgress——已經投入該主題的研究點數會變成「投給一個
// 已經完成的主題」而無處可去,下回合 advanceResearch() 會把主題推進到下一項,那些點數就此浪費。
// 這是最小迴圈的邊界情況,影響有限(只在偷到「正好在研究中」的科技時發生),故不在本輪額外處理。
func applyTechTheft(ps *engine.PlayerState, opt spyStealOption) {
	if ps.CompletedTopics == nil {
		ps.CompletedTopics = make(map[gamedata.ResearchTopic]bool)
	}
	if ps.ChosenTech == nil {
		ps.ChosenTech = make(map[gamedata.ResearchTopic]gamedata.Technology)
	}
	if ps.ExplicitChoice == nil {
		ps.ExplicitChoice = make(map[gamedata.ResearchTopic]bool)
	}
	ps.CompletedTopics[opt.Topic] = true
	ps.ChosenTech[opt.Topic] = opt.Tech
	ps.ExplicitChoice[opt.Topic] = true
}

// spyAttackerBonus 算出「攻擊方(派間諜出去偷科技的一方)」的 attacker bonus(AB,見
// gamedata.SpyEffectiveThreshold 的定義)。目前只接上 SpySlotBonus——手冊 Spy Bonuses 表中
// 唯一「有明確人數 → 加成對照表」的項目(見 gamedata/spy.go 檔頭)。
//
// 種族特性(SpyRaceTraitBonus)/科技(SpyTechnologyBonus)/政府(SpyGovernmentDefenseBonus,
// 手冊本來就只給 Defense 欄,offense 無政府加成)三項現行 remake 無法從
// AIOpponent/engine.PlayerState 推導出對應資料(無種族間諜特性強度資料、無逐科技模型可查是否
// 擁有 spy.go 列的 5 項科技、AIOpponent 無政府型態欄位),一律回 0——TODO,待補上這些欄位後
// 在此函式接上,不臆造數字。
func spyAttackerBonus(spyCount int) int {
	return gamedata.SpySlotBonus(spyCount)
}

// spyDefenderBonus 算出「防守方(被偷科技的一方)」的 defender bonus(DB)。手冊區分 Spy
// (攻擊,逐對手指派)與 Agent(防守,不分對手、全體共用)兩種 slot,本 remake 目前完全沒有
// 追蹤 Agent 數(逐對手分配任務選單延後,見檔頭說明),故 DB 固定為 0——這正好對應手冊原文
// 「defenses against enemy spies are active...even with zero defending agents」描述的
// 「零 Agent」情境,不是遺漏,是誠實反映目前的簡化狀態。
//
// TODO:接上 Agent 訓練系統後,DB 應改用 gamedata.SpySlotBonus(agentCount) + 種族/科技/
// 政府加成(政府加成手冊只給 Defense 欄,屆時可直接用 gamedata.SpyGovernmentDefenseBonus)。
func spyDefenderBonus() int {
	return 0
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
// 現行 remake 的 ab/db 只含 gamedata.SpySlotBonus(間諜數換算),不含種族/科技/政府加成
// (spyAttackerBonus/spyDefenderBonus 已標 TODO 保守回 0)——SpySlotBonus 上限 41(63 名
// 間諜),SpyVsSpyDefenderBonus(0)=20 為基準,即使 ab 拉滿 41 也只有 net=41-20=21,遠不到
// ±80 門檻:透過目前正常遊戲流程幾乎不可能觸發擊殺,這是誠實反映「輔助加成未建置」的結果,
// 不是 bug——之後接上種族/科技/政府 bonus 才會讓門檻可及。單元測試改用直接構造的 ab/db 數值
// 驗證函式本身邏輯正確,不透過完整對局路徑(那條路徑目前確實走不到擊殺)。
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

// spyStealAttempt 對單一方向(attacker 派 spyCount 個間諜偷 defender 的科技)跑一次 STEAL +
// SpyVsSpy 判定,回傳要記進 LastEspionage 的訊息(可能 0~2 則)、attacker 間諜是否被擊殺、
// 以及偷到的科技是否套用到了 *attackerPS(呼叫端已把 attackerPS 指到正確的 engine.PlayerState)。
// attackerName/defenderName 純供訊息文字使用。
func spyStealAttempt(rng *rand.Rand, attackerPS *engine.PlayerState, defenderPS engine.PlayerState,
	spyCount int, attackerName, defenderName string) (messages []string, attackerSpyKilled bool) {
	ab := spyAttackerBonus(spyCount)
	db := spyDefenderBonus()

	e := gamedata.SpyEffectiveThreshold(gamedata.SpyThresholdSteal, db, ab)
	p := gamedata.SpyRollChance(e)
	if rng.Float64() < p {
		opts := spyStealOptions(*attackerPS, defenderPS)
		if len(opts) == 0 {
			messages = append(messages, fmt.Sprintf(
				"%s 的間諜潛入 %s 得手,但對方已無%s尚未擁有的科技可偷", attackerName, defenderName, attackerName))
		} else {
			pick := opts[rng.Intn(len(opts))]
			applyTechTheft(attackerPS, pick)
			messages = append(messages, fmt.Sprintf(
				"%s 的間諜從 %s 偷得科技:%s", attackerName, defenderName, gamedata.TechnologyName(pick.Tech)))
		}
	}

	outcome := resolveSpyVsSpy(ab, db, false) // 最小迴圈:間諜恆執行 STEAL,不下 HIDE 指令
	if outcome.AttackerKilled {
		attackerSpyKilled = true
		messages = append(messages, fmt.Sprintf("%s 的一名間諜在 %s 被反間諜擊殺", attackerName, defenderName))
	}
	return messages, attackerSpyKilled
}

// advanceEspionage 每回合結算玩家 ↔ 各 AI 對手之間的間諜行動(最小迴圈:只做 STEAL,見檔頭
// 說明)。呼叫時機:EndTurn 已完成玩家與所有 AI 本回合的研究/經濟結算之後,讓「偷到的科技」
// 判定用的是本回合最新的 CompletedTopics/ChosenTech(見 EndTurn 呼叫點註解)。
func (s *GameSession) advanceEspionage() {
	s.LastEspionage = nil
	s.ensurePlayerSpies()
	if s.spyRand == nil {
		s.spyRand = rand.New(rand.NewSource(s.EventSeed*2654435761 + 7))
	}

	for i := range s.AIPlayers {
		a := &s.AIPlayers[i]

		// 間諜維護費:opt-in,預設 0 間諜時扣款為 0,不影響既有對局/測試(見 spyMaintenancePerSpyBC)。
		if s.PlayerSpies[i] > 0 {
			s.Player.BC -= s.PlayerSpies[i] * spyMaintenancePerSpyBC
		}

		// 玩家 → AI:偷科技 + SpyVsSpy。
		if s.PlayerSpies[i] > 0 {
			msgs, killed := spyStealAttempt(s.spyRand, &s.Player, a.Player, s.PlayerSpies[i], "我方", a.Name)
			s.LastEspionage = append(s.LastEspionage, msgs...)
			if killed && s.PlayerSpies[i] > 0 {
				s.PlayerSpies[i]--
			}
		}

		// AI → 玩家:偷科技 + SpyVsSpy(對稱處理;AI 已知科技集長期而言很小,見檔頭說明)。
		if a.Spies > 0 {
			msgs, killed := spyStealAttempt(s.spyRand, &a.Player, s.Player, a.Spies, a.Name, "我方")
			s.LastEspionage = append(s.LastEspionage, msgs...)
			if killed && a.Spies > 0 {
				a.Spies--
			}
		}
	}
}
