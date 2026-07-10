package engine

import "github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"

// RunResearchPhase 執行一回合的研究階段:把帝國本回合總研究點(researchPoints,通常是
// 各殖民地 ColonyOutput.Research 加總)灌入玩家目前研究中的主題(ps.ResearchTopic),
// 判斷該主題本回合是否累積達成花費(gamedata.ResearchChoiceFor 查得的 Cost)而完成。
//
// 完成時處理 MOO2 的「每主題數科技間抉擇」:
//   - ResearchAll 主題(如 TOPIC_CHEMISTRY):全部 Choices 一次記入 ChosenTech。
//   - 單一選項主題:直接記入該項。
//   - 多選項主題:預設記入第一項並設 PendingChoice,讓玩家(或 AI)之後經 ApplyResearchChoice
//     改選其他項(不阻塞回合:先給預設,可覆寫)。
// 此函式為純函式,不含 RNG。
//
// 溢出處理:完成主題後 ResearchProgress -= cost,保留超出成本的點數帶到下一個主題,
// 不歸零浪費。這是 MOO2 系列常見的研究點結算慣例。無法從 openorion2 既有移植碼直接
// 查證此細節:researchProgress 欄位只出現在讀檔/初始化/UI 顯示
// (src/gamestate.cpp:1047 初始化為 0、1126 讀檔、1269 算「還差多少」的 helper、
// src/galaxy.cpp:1986 附近畫側欄 RP 顯示),openorion2 尚未重建「每回合累加+完成判定」
// 的回合處理邏輯,故無現成程式碼可查證溢出是否保留。若之後找到權威反證(如原版
// reverse-engineering 或社群 wiki 證實實際是歸零),再回頭修正本函式與此註解。
//
// cost == 0 的主題(如 TOPIC_STARTING_TECH,tech.cpp 研究表中純填充、帝國起始即視為
// 已完成的項目)視為「無需研究」:直接標記完成、ResearchProgress 不動(沒有花費可扣、
// 也不消耗本回合研究點去累加一個不存在的成本),回傳 true。
func RunResearchPhase(ps PlayerState, researchPoints int) (PlayerState, bool) {
	if ps.CompletedTopics == nil {
		ps.CompletedTopics = make(map[gamedata.ResearchTopic]bool)
	}

	choice := gamedata.ResearchChoiceFor(ps.ResearchTopic)
	cost := choice.Cost

	if cost == 0 {
		ps.CompletedTopics[ps.ResearchTopic] = true
		recordCompletion(&ps, ps.ResearchTopic, choice)
		return ps, true
	}

	ps.ResearchProgress += researchPoints

	if ps.ResearchProgress >= cost {
		ps.CompletedTopics[ps.ResearchTopic] = true
		ps.ResearchProgress -= cost // 保留溢出點數(見上方註解)
		recordCompletion(&ps, ps.ResearchTopic, choice)
		return ps, true
	}

	return ps, false
}

// recordCompletion 在主題完成時記錄「選定解鎖」的科技。ResearchAll/單選直接全記;
// 多選預設記第一項並開啟 PendingChoice(玩家可 ApplyResearchChoice 改選)。
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
		return
	}
	ps.ChosenTech[topic] = choice.Choices[0] // 預設第一項
	if len(choice.Choices) > 1 {
		ps.PendingChoice = topic
		ps.HasPendingChoice = true
	}
}

// ApplyResearchChoice 讓玩家/AI 把目前 PendingChoice 主題改選為 tech(須為該主題合法選項)。
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
