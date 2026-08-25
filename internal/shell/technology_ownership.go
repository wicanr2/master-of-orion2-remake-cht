package shell

import (
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// playerStateKnowsTech 是所有逐 application 消費端的共同判定。GrantedTechs 表示
// 原版 callback／多來源取得的額外 application；其餘維持既有主題完成相容語意。
func playerStateKnowsTech(ps engine.PlayerState, topic gamedata.ResearchTopic, tech gamedata.Technology) bool {
	if ps.GrantedTechs != nil && ps.GrantedTechs[tech] {
		return true
	}
	if ps.CompletedTopics == nil || !ps.CompletedTopics[topic] {
		return false
	}
	if ps.ExplicitChoice == nil || !ps.ExplicitChoice[topic] {
		return true
	}
	return ps.ChosenTech != nil && ps.ChosenTech[topic] == tech
}

// grantTechnologyApplication 授予一個明確 application。若同一主題已有另一個主要
// 選擇，不覆蓋舊選擇，而放入 GrantedTechs，對映原版逐 application 狀態陣列。
func grantTechnologyApplication(ps *engine.PlayerState, topic gamedata.ResearchTopic, tech gamedata.Technology) {
	if ps == nil || playerStateKnowsTech(*ps, topic, tech) {
		return
	}
	if ps.CompletedTopics == nil {
		ps.CompletedTopics = make(map[gamedata.ResearchTopic]bool)
	}
	if !ps.CompletedTopics[topic] {
		ps.CompletedTopics[topic] = true
		if ps.ChosenTech == nil {
			ps.ChosenTech = make(map[gamedata.ResearchTopic]gamedata.Technology)
		}
		if ps.ExplicitChoice == nil {
			ps.ExplicitChoice = make(map[gamedata.ResearchTopic]bool)
		}
		ps.ChosenTech[topic] = tech
		ps.ExplicitChoice[topic] = true
	} else {
		if ps.GrantedTechs == nil {
			ps.GrantedTechs = make(map[gamedata.Technology]bool)
		}
		ps.GrantedTechs[tech] = true
	}
	applyTechnologyGrantCallbacks(ps, tech)
}

// applyTechnologyGrantCallbacks 對映 sub_E4204 的玩家可見特殊授予。其他分支只是
// 原版 raw 快取重算，remake 由來源狀態動態計算，不在此複製。
func applyTechnologyGrantCallbacks(ps *engine.PlayerState, tech gamedata.Technology) {
	if ps == nil || tech != gamedata.TECH_BATTLEOIDS ||
		playerStateKnowsTech(*ps, gamedata.TOPIC_ASTRO_ENGINEERING, gamedata.TECH_ARMOR_BARRACKS) {
		return
	}
	if ps.GrantedTechs == nil {
		ps.GrantedTechs = make(map[gamedata.Technology]bool)
	}
	ps.GrantedTechs[gamedata.TECH_ARMOR_BARRACKS] = true
}

func applyResearchTopicGrantCallbacks(ps *engine.PlayerState, topic gamedata.ResearchTopic) {
	if ps == nil {
		return
	}
	for _, tech := range gamedata.ResearchChoicesForTopic(topic) {
		if playerStateKnowsTech(*ps, topic, tech) {
			applyTechnologyGrantCallbacks(ps, tech)
		}
	}
}
