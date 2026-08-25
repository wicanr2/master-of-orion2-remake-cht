package engine

import "github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"

// RunResearchPhase 執行一回合的研究階段:把帝國本回合總研究點(researchPoints,通常是
// 各殖民地 ColonyOutput.Research 加總)灌入玩家目前研究中的主題(ps.ResearchTopic),
// 判斷該主題本回合是否累積達成花費(gamedata.ResearchChoiceFor 查得的 Cost)而完成。
//
// 完成時處理 MOO2 的「每主題數科技間抉擇」:
//   - ResearchAll 主題(如 TOPIC_CHEMISTRY):全部 Choices 一次記入 ChosenTech。
//   - 單一選項主題:直接記入該項。
//   - 多選項主題:授予研究開始前已寫入 ResearchApplication 的項目。
//
// 舊存檔／純 engine caller 若沒有預選，才保留突破後 PendingChoice 相容分支。
//
// 一般呼叫使用固定最低擲骰 1，供純規則測試與舊 adapter 使用；正常遊戲必須走
// RunResearchPhaseWithRoller 注入可保存亂數流。原版不是達成本立即完成：只有累積值
// 嚴格超過成本才按超額比例擲突破率，成功後進度直接清零，不保留溢出。
//
// ResearchTopic==0 對應原版 player+0x321==0 的「沒有有效研究 field」；sub_E44E0
// 會直接返回，不累積、不完成。起始科技由開局初始化授予，不由每回合突破鏈補發。
func RunResearchPhase(ps PlayerState, researchPoints int) (PlayerState, bool) {
	return RunResearchPhaseWithRoller(ps, researchPoints, func(int) int { return 1 })
}

// RunResearchPhaseWithRoller 對映 Check_For_Research_Breakthrough_ @ 0xE44E0。
// roller 僅在本回合研究產出 >0 且累積值嚴格超過成本時呼叫，回傳 1..max；nil 表示
// 不得突破。這讓 shell 注入可存檔亂數流，也保留 engine 的可重播性。
func RunResearchPhaseWithRoller(ps PlayerState, researchPoints int, roller func(max int) int) (PlayerState, bool) {
	if ps.CompletedTopics == nil {
		ps.CompletedTopics = make(map[gamedata.ResearchTopic]bool)
	}
	if ps.HyperAdvancedLevels == nil {
		ps.HyperAdvancedLevels = make(map[gamedata.ResearchTopic]int)
		// 舊 JSON 只有 CompletedTopics；已完成的 Hyper topic 等價於第一級。
		for topic, completed := range ps.CompletedTopics {
			if completed && gamedata.IsHyperAdvancedTopic(topic) {
				ps.HyperAdvancedLevels[topic] = 1
			}
		}
	}

	choice := gamedata.ResearchChoiceFor(ps.ResearchTopic)
	cost := choice.Cost
	if ps.ResearchTopic == 0 {
		return ps, false
	}
	// 原版 Player_Research_Cost_ @ 0xE1E96：第一級讀版本基礎成本，後續每個已完成
	// level byte 再加 10000。HyperAdvancedLevels 正是同一八 byte 的 remake 對應。
	if gamedata.IsHyperAdvancedTopic(ps.ResearchTopic) {
		base := cost
		if ps.HyperAdvancedResearchCost > 0 {
			base = ps.HyperAdvancedResearchCost
		}
		cost = gamedata.HyperAdvancedRepeatedCost(base, ps.HyperAdvancedLevels[ps.ResearchTopic])
	}

	if cost <= 0 {
		return ps, false
	}

	ps.ResearchProgress += researchPoints

	chance := gamedata.ResearchBreakthroughChance(cost, ps.ResearchProgress)
	if researchPoints > 0 && chance > 0 && roller != nil &&
		gamedata.ResearchBreakthroughSucceeded(chance, roller(100)) {
		ps.CompletedTopics[ps.ResearchTopic] = true
		if gamedata.IsHyperAdvancedTopic(ps.ResearchTopic) {
			ps.HyperAdvancedLevels[ps.ResearchTopic]++
		}
		ps.ResearchProgress = 0
		recordCompletion(&ps, ps.ResearchTopic, choice)
		return ps, true
	}

	return ps, false
}

// ForceCompleteResearchTopic 對映不經成本與突破骰、直接完成目前 field 的事件／授予鏈。
// 回傳完成前的 topic，供呼叫端在原版清空 player+0x321 後仍能顯示及執行 callback。
// topic==0 時不捏造科技，但仍清除研究暫存狀態。
func ForceCompleteResearchTopic(ps PlayerState) (PlayerState, gamedata.ResearchTopic, bool) {
	topic := ps.ResearchTopic
	ps.ResearchProgress = 0
	ps.ResearchTopic = 0
	if topic == 0 {
		ps.ResearchApplication = 0
		ps.HasResearchApplication = false
		ps.PendingChoice = 0
		ps.HasPendingChoice = false
		return ps, topic, false
	}
	if ps.CompletedTopics == nil {
		ps.CompletedTopics = make(map[gamedata.ResearchTopic]bool)
	}
	if ps.HyperAdvancedLevels == nil {
		ps.HyperAdvancedLevels = make(map[gamedata.ResearchTopic]int)
		for completedTopic, completed := range ps.CompletedTopics {
			if completed && gamedata.IsHyperAdvancedTopic(completedTopic) {
				ps.HyperAdvancedLevels[completedTopic] = 1
			}
		}
	}
	ps.CompletedTopics[topic] = true
	if gamedata.IsHyperAdvancedTopic(topic) {
		ps.HyperAdvancedLevels[topic]++
	}
	recordCompletion(&ps, topic, gamedata.ResearchChoiceFor(topic))
	return ps, topic, true
}

// recordCompletion 在主題完成時才把目前 application 寫入已擁有科技。
func recordCompletion(ps *PlayerState, topic gamedata.ResearchTopic, choice gamedata.ResearchChoice) {
	if ps.ChosenTech == nil {
		ps.ChosenTech = make(map[gamedata.ResearchTopic]gamedata.Technology)
	}
	if len(choice.Choices) == 0 {
		return // 純填充主題(如起始科技),無科技可記
	}
	if choice.ResearchAll {
		// 全解:記第一項為代表(ChosenTech 為單值;全解語意由 ResearchAll 旗標表達)。
		ps.ChosenTech[topic] = choice.Choices[0]
		ps.HasResearchApplication = false
		ps.ResearchApplication = 0
		return
	}
	if ps.HasResearchApplication && technologyInChoices(ps.ResearchApplication, choice.Choices) {
		ps.ChosenTech[topic] = ps.ResearchApplication
		if len(choice.Choices) > 1 {
			if ps.ExplicitChoice == nil {
				ps.ExplicitChoice = make(map[gamedata.ResearchTopic]bool)
			}
			ps.ExplicitChoice[topic] = true
		}
		ps.HasResearchApplication = false
		ps.ResearchApplication = 0
		ps.PendingChoice = 0
		ps.HasPendingChoice = false
		return
	}
	ps.ChosenTech[topic] = choice.Choices[0] // 舊狀態相容：先保留第一項，等待一次改選
	if len(choice.Choices) > 1 {
		ps.PendingChoice = topic
		ps.HasPendingChoice = true
	}
}

func technologyInChoices(tech gamedata.Technology, choices []gamedata.Technology) bool {
	for _, candidate := range choices {
		if candidate == tech {
			return true
		}
	}
	return false
}

// SelectResearchApplication 選定目前 topic 突破後要取得的 application，但不提前解鎖。
func SelectResearchApplication(ps PlayerState, topic gamedata.ResearchTopic, tech gamedata.Technology) (PlayerState, bool) {
	if ps.ResearchTopic != topic || ps.CompletedTopics[topic] {
		return ps, false
	}
	choice := gamedata.ResearchChoiceFor(topic)
	if !technologyInChoices(tech, choice.Choices) {
		return ps, false
	}
	ps.ResearchApplication = tech
	ps.HasResearchApplication = true
	ps.PendingChoice = 0
	ps.HasPendingChoice = false
	return ps, true
}

// ApplyResearchChoice 僅供舊存檔的「突破後 PendingChoice」相容路徑。
// 成功則更新 ChosenTech 並清除 PendingChoice;非法選項或無待決則原樣返回 false。
func ApplyResearchChoice(ps PlayerState, tech gamedata.Technology) (PlayerState, bool) {
	if !ps.HasPendingChoice {
		return ps, false
	}
	choice := gamedata.ResearchChoiceFor(ps.PendingChoice)
	valid := false
	for _, t := range choice.Choices {
		if t == tech {
			valid = true
			break
		}
	}
	if !valid {
		return ps, false
	}
	if ps.ChosenTech == nil {
		ps.ChosenTech = make(map[gamedata.ResearchTopic]gamedata.Technology)
	}
	if ps.ExplicitChoice == nil {
		ps.ExplicitChoice = make(map[gamedata.ResearchTopic]bool)
	}
	ps.ChosenTech[ps.PendingChoice] = tech
	ps.ExplicitChoice[ps.PendingChoice] = true // 標記此主題已明確抉擇(元件解鎖改科技層級)
	ps.HasPendingChoice = false
	return ps, true
}
