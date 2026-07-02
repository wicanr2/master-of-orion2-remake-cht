package engine

import "github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"

// RunResearchPhase 執行一回合的研究階段:把帝國本回合總研究點(researchPoints,通常是
// 各殖民地 ColonyOutput.Research 加總)灌入玩家目前研究中的主題(ps.ResearchTopic),
// 判斷該主題本回合是否累積達成花費(gamedata.ResearchChoiceFor 查得的 Cost)而完成。
//
// 本階段只回答「主題是否研究完成」,不選擇完成後要解鎖 ResearchChoice.Choices 中哪一項
// 科技——那是玩家(或 AI)決策,屬於後續另一步驟;此函式為純函式,不含 RNG。
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

	cost := gamedata.ResearchChoiceFor(ps.ResearchTopic).Cost

	if cost == 0 {
		ps.CompletedTopics[ps.ResearchTopic] = true
		return ps, true
	}

	ps.ResearchProgress += researchPoints

	if ps.ResearchProgress >= cost {
		ps.CompletedTopics[ps.ResearchTopic] = true
		ps.ResearchProgress -= cost // 保留溢出點數(見上方註解)
		return ps, true
	}

	return ps, false
}
