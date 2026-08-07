package gamedata

// leader_skill_apply.go:**領袖技能怎麼疊**,以及 admin 技能的單位。
//
// ============ 手冊給了一條 remake 一直做錯的規則 ============
//
// 手冊 p.137「Applicability」:
//
//	The effects of the **Megawealth and Researcher** abilities are **cumulative**,
//	but **the rest are not**. If you assign more than one leader with the same ability
//	to the same fleet, the fleet gets the effect for that particular ability of
//	**the leader with the best applicable bonus**. The same goes for abilities affecting
//	your empire as a whole, with the exception of **Assassin**…
//
// remake 的 `applyLeaderColonyBonuses` 一直是**每個領袖都加一次**——兩個貿易家就加兩份。
// 照手冊只有 Megawealth 與 Researcher 該這樣,其餘一律**取最佳的那一個**。
//
// 刺客是第三個例外,但它的理由不同:手冊說「Assassin leaders all get a chance to act each
// turn」——那是**擲骰次數**變多,不是加成疊加。remake 沒有刺客系統,不在這裡處理。
//
// ============ admin 技能的單位是查出來的,不是猜的 ============
//
// 加成值來自 `baseSkillValues[2]`(openorion2 gamestate.cpp:75),**單位**來自
// `skillFormatStrings[2]`(officer.cpp:75)——兩張表在 openorion2 是分開的,
// 只看數值會不知道 10 是「10 點」還是「10%」。
//
//	code  技能               base   格式      單位
//	  0   Environmentalist   −10   "%+d%%"   百分比(降低會產生污染的產能)
//	  1   Farming Leader     +10   "%+d%%"   百分比(食物)
//	  2   Financial Leader   +10   "%+d%%"   百分比(收入)
//	  3   Instructor          +1   "%+d"     **固定點數**(每回合艦員經驗)
//	  4   Labor Leader       +10   "%+d%%"   百分比(工人產能)
//	  5   Medicine           +10   "%+d%%"   百分比(人口成長率)
//	  6   Science Leader     +10   "%+d%%"   百分比(研究)
//	  7   Spiritual Leader    +5   "%+d%%"   百分比(士氣)
//	  8   Tactics             +2   "%+d"     固定點數
//
// Instructor 那一格是 `"%+d"` 而不是百分比,正好對上手冊那句
// 「**Boosts the number of experience points earned each turn**」——是每回合多幾點,
// 不是多幾成。兩個獨立來源指到同一個語意。
//
// ============ Tactics:原版自己就沒實作 ============
//
// 手冊在 Tactics 那一條的**最後一句**寫著:
//
//	Improves the coordination of the military forces in the system, adding to the
//	Beam Attack and the strength of ground troops.  **This skill is not implemented.**
//
// 所以 remake 不做 Tactics **不是缺口**,是與原版一致。這句話值得記下來:
// 少了它,下一個盤點的人會把它列進「還沒接的技能」然後花時間去找它該有什麼效果。

// LeaderSkillCumulative 回報某個技能的加成是否可**累加**(多位領袖各加一份)。
//
// 手冊 p.137:只有 Megawealth 與 Researcher 是累加的,其餘取最佳者。
func LeaderSkillCumulative(skillID int) bool {
	return skillID == int(SKILL_MEGAWEALTH) || skillID == int(SKILL_RESEARCHER)
}

// LeaderSkillCombine 把同一個技能的多份加成合成一份。
//
// 累加型直接相加;其餘取**絕對值最大**的那一個——Environmentalist 的加成是負的
// (降低污染),取「數值最大」會挑到最弱的那個領袖。
func LeaderSkillCombine(skillID int, bonuses []int) int {
	if len(bonuses) == 0 {
		return 0
	}
	if LeaderSkillCumulative(skillID) {
		sum := 0
		for _, b := range bonuses {
			sum += b
		}
		return sum
	}
	best := bonuses[0]
	for _, b := range bonuses[1:] {
		if abs(b) > abs(best) {
			best = b
		}
	}
	return best
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// LeaderSkillTacticsIsUnimplementedInTheOriginal 是一個給讀程式碼的人看的常數。
//
// 手冊在 Tactics 那條的最後一句明寫「This skill is not implemented」——**原版自己就沒做**。
// remake 不做它與原版一致,不是缺口。
const LeaderSkillTacticsIsUnimplementedInTheOriginal = true
