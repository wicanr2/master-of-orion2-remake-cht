package shell

import "github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"

// leader_effects.go 收納「不是殖民地欄位、也不是單一艦艇欄位」的領袖技能消費端。
//
// 技能數值與累加規則仍以 gamedata.LeaderSkillBonus／LeaderSkillCombine 為唯一來源。
// 這裡只把已經存在的技能值接到 remake 已有的狀態模型；沒有對應模型的原版細節，
// 仍在呼叫端以註解標出，不用猜一個看似完整的替代規則。

// leaderEmpireSkillBonus 回傳帝國層技能的合成值。Captain 技能也能透過這個 helper
// 查詢，但呼叫端只會對「其規則單位是帝國」的技能使用它；逐艦技能仍必須走
// shipOfficerSkillBonus，不能把未指派的軍官套給整個帝國。
func leaderEmpireSkillBonus(leaders []Leader, skill gamedata.LeaderSkills) int {
	bonuses := make([]int, 0, len(leaders))
	for _, leader := range leaders {
		tier := leaderSkillTier(leader, int(skill))
		if tier <= 0 {
			continue
		}
		if bonus := gamedata.LeaderSkillBonus(int(skill), tier,
			leaderDisplayLevelToExpLevel(leader.Level)); bonus != 0 {
			bonuses = append(bonuses, bonus)
		}
	}
	return gamedata.LeaderSkillCombine(int(skill), bonuses)
}

// leaderAssassinBonuses 保留每位刺客各自的機率。手冊明寫「Assassin leaders all get
// a chance to act each turn」；這裡不能先用 LeaderSkillCombine，否則多位刺客只會
// 擲一次骰。
func leaderAssassinBonuses(leaders []Leader) []int {
	bonuses := make([]int, 0, len(leaders))
	for _, leader := range leaders {
		tier := leaderSkillTier(leader, int(gamedata.SKILL_ASSASSIN))
		if tier <= 0 {
			continue
		}
		if bonus := gamedata.LeaderSkillBonus(int(gamedata.SKILL_ASSASSIN), tier,
			leaderDisplayLevelToExpLevel(leader.Level)); bonus > 0 {
			bonuses = append(bonuses, bonus)
		}
	}
	return bonuses
}

// leaderFamousHireModifier 是原版「只取效果最強的 Famous」折扣。招募機率的原版
// 候選池／擲骰模型尚未存在，因此本輪只接上有明確數值的雇用費修正，不把固定的
// offer 週期冒充成原版機率。
func leaderFamousHireModifier(leaders []Leader) int {
	bonuses := make([]int, 0, len(leaders))
	for _, leader := range leaders {
		tier := leaderSkillTier(leader, int(gamedata.SKILL_FAMOUS))
		if tier <= 0 {
			continue
		}
		bonuses = append(bonuses, gamedata.LeaderSkillBonus(int(gamedata.SKILL_FAMOUS), tier,
			leaderDisplayLevelToExpLevel(leader.Level)))
	}
	return gamedata.LeaderHireModifier(bonuses)
}

// diplomacyRelationGain 把「外交點數」映射到 remake 目前唯一可觀察的外交結果：
// Relation。種族倍率仍先作用於原有提案值；Diplomat 的 +10 等固定點數再加上去。
// 這不是原版接受門檻的宣稱——remake 尚未保存那個獨立分數——但不會讓技能變成
// 只顯示在說明文字裡的空效果。
func (s *GameSession) diplomacyRelationGain(base int) int {
	if base <= 0 {
		return base
	}
	return base*(100+s.raceDiploBonusPct())/100 +
		leaderEmpireSkillBonus(s.Leaders, gamedata.SKILL_DIPLOMAT)
}

// leaderMegawealthBC 回傳本回合所有 Megawealth 領袖的固定帝國收入。Megawealth
// 是手冊列明的累加型技能；同時免維護費的既有規則仍由 leaderHasMegawealth 保留。
func leaderMegawealthBC(leaders []Leader) int {
	return leaderEmpireSkillBonus(leaders, gamedata.SKILL_MEGAWEALTH)
}

// advanceLeaderAssassinActions 在間諜任務前處理刺客的獨立行動。玩家刺客攻擊各 AI
// 的 DefensiveAgents；AI 若有刺客則反向攻擊玩家的共用 DefensiveAgents。每位刺客
// 每個目標各擲一次，且只在目標仍有 Agent 時消費亂數，保持沒有刺客時舊對局序列不變。
func (s *GameSession) advanceLeaderAssassinActions() {
	if s == nil || s.spyRand == nil {
		return
	}
	playerAssassins := leaderAssassinBonuses(s.Leaders)
	for i := range s.AIPlayers {
		ai := &s.AIPlayers[i]
		for _, chance := range playerAssassins {
			if ai.DefensiveAgents <= 0 {
				break
			}
			if s.spyRand.Intn(100) < chance {
				ai.DefensiveAgents--
				s.LastEspionage = append(s.LastEspionage,
					"我方刺客在"+ai.Name+"刺殺了一名防守 Agent")
			}
		}

		for _, chance := range leaderAssassinBonuses(ai.Leaders) {
			if s.DefensiveAgents <= 0 {
				break
			}
			if s.spyRand.Intn(100) < chance {
				s.DefensiveAgents--
				s.LastEspionage = append(s.LastEspionage,
					ai.Name+"的刺客刺殺了一名我方防守 Agent")
			}
		}
	}
}

// StarChartVisible 是 Galactic Lore 的「行星／星圖情報」消費端。它刻意與
// VisibleStars 分開：VisibleStars 仍是敵方艦隊偵測與戰爭迷霧規則，Galactic Lore
// 只讓星圖與行星列表立即顯示已知天體，不把 AI 抽象艦隊位置變成全知。
func (s *GameSession) StarChartVisible() []bool {
	if s == nil {
		return nil
	}
	visible := s.VisibleStars()
	if leaderEmpireSkillBonus(s.Leaders, gamedata.SKILL_GALACTIC_LORE) <= 0 {
		return visible
	}
	for i := range visible {
		visible[i] = true
	}
	return visible
}

// galacticLoreCombatBonus 回傳對太空怪獸／安塔蘭戰鬥的固定加成。一般 AI 艦隊不套用，
// 因為技能說明只列這兩類敵人。
func (s *GameSession) galacticLoreCombatBonus() int {
	if s == nil {
		return 0
	}
	return leaderEmpireSkillBonus(s.Leaders, gamedata.SKILL_GALACTIC_LORE)
}
