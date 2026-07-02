package gamedata

// 間諜(Spying)機率公式與查表,移植自 moo2_patch1.5/MANUAL_150.html「Notes on Spying」段
// (p.113,含 Spy Bonuses / Assassins / Roll Chance / Spy vs Spy 四小節)。openorion2 未實作
// 間諜邏輯,本檔無原始碼可對照,一律以手冊原文數字為準;手冊沒有給精確公式的項目一律標
// `TODO 手冊未明列` 並保留範圍常數,不臆測填數字。
//
// 擲骰(roll)本身不在此實作:SpyRollChance 只回傳「攻擊方成功機率 p」這個決定性數值,
// 呼叫端要模擬 AR/DR 兩顆 1..100 骰子的話,手冊原文機制是:
//   AR = 100(幸運骰,必定成功,若為 stealing 還會順便嫁禍)
//   或 AR - DR > E 才成功
// p 已經是這個機制的封閉解機率(見 SpyRollChance 註解),上層可直接拿 p 跟 rand.Float64() 比較,
// 不必真的擲兩顆骰子。

// SpySlotBonus 每個間諜/防諜 slot 依派駐人數(0-63)換算出的加成。
// 手冊原文(Spy Bonuses):「The first five spies add 2 points each, and spies 6 to 10 add 1
// point each. Subsequently, each pair of spies adds 1 point to the bonus. So with spy 11 the
// bonus is still +15 while spy 12 brings it up to +16. The maximum bonus is +41 for 62 or 63
// spies.」spyCount 超出 [0,63] 會夾在範圍內(手冊:「You can train up to 63 defensive agents
// and 63 spies per opponent」)。
func SpySlotBonus(spyCount int) int {
	if spyCount < 0 {
		spyCount = 0
	}
	if spyCount > 63 {
		spyCount = 63
	}
	switch {
	case spyCount <= 5:
		return 2 * spyCount
	case spyCount <= 10:
		return 10 + (spyCount - 5)
	default:
		return 15 + (spyCount-10)/2
	}
}

// SpyGovernmentType 政府型態,僅用於 SpyGovernmentDefenseBonus 查表。
// 手冊此表沒有對應到既有的通用列舉(enums.go 只有 RaceTrait/ForeignPolicy,無 Government),
// 故在此另建、加 Spy 前綴,避免與其他檔案之後新增的通用 Government 型別撞名。
type SpyGovernmentType int

const (
	SpyGovFeudalism SpyGovernmentType = iota
	SpyGovConfederation
	SpyGovDictatorship
	SpyGovImperium
	SpyGovDemocracy
	SpyGovFederation
	SpyGovUnification
	SpyGovGalacticUnification
)

// 手冊原文(Spy Bonuses 表,Government 列,Defense/Offense 兩欄):
//
//	Feudalism            0   -
//	Confederation        0   -
//	Dictatorship        10   -
//	Imperium            15   -
//	Democracy          -10   -
//	Federation         -10   -
//	Unification         15   -
//	Galactic Unification 15  -
//
// Offense 欄全為「-」,即政府型態不提供攻擊加成,故只有 Defense 查表函式。
var spyGovernmentDefenseBonusTable = map[SpyGovernmentType]int{
	SpyGovFeudalism:           0,
	SpyGovConfederation:       0,
	SpyGovDictatorship:        10,
	SpyGovImperium:            15,
	SpyGovDemocracy:           -10,
	SpyGovFederation:          -10,
	SpyGovUnification:         15,
	SpyGovGalacticUnification: 15,
}

// SpyGovernmentDefenseBonus 政府型態對防諜(Defense)的加成;手冊未列的值一律回 0。
func SpyGovernmentDefenseBonus(gov SpyGovernmentType) int {
	return spyGovernmentDefenseBonusTable[gov]
}

// SpyRaceTraitLevel 種族創建畫面「間諜」特性的三檔強度(Race Picks -3/+3/+6)。
// enums.go 的 RaceTrait.TRAIT_SPYING 只標記「有此特性」,不含強度分級,故另建列舉對應手冊三檔加成。
type SpyRaceTraitLevel int

const (
	SpyRaceTraitMinus3 SpyRaceTraitLevel = iota // Spy -3 picks
	SpyRaceTraitPlus3                           // Spy +3 picks
	SpyRaceTraitPlus6                           // Spy +6 picks
)

// SpyRaceTraitBonus 種族間諜特性的 Defense/Offense 加成(手冊表中兩欄同值)。
// 手冊原文:「Spy -3 picks -10 -10 / Spy +3 picks 10 10 / Spy +6 picks 20 20」。
func SpyRaceTraitBonus(level SpyRaceTraitLevel) int {
	switch level {
	case SpyRaceTraitMinus3:
		return -10
	case SpyRaceTraitPlus3:
		return 10
	case SpyRaceTraitPlus6:
		return 20
	default:
		return 0
	}
}

// SpyTelepathicRaceBonus TRAIT_TELEPATHIC(enums.go)種族的間諜 Defense/Offense 加成(兩欄同值)。
// 手冊原文:「Telepathic 10 10」。
const SpyTelepathicRaceBonus = 10

// SpyTechnologyBonus 已知科技對間諜 Defense/Offense 的加成(手冊表中兩欄同值)。
// 手冊原文(Spy Bonuses 表,Technology 列):
//
//	Neural Scanner       10  10
//	Telepathic Training   5   5
//	Cyber Security Link  10  10
//	Stealth Suit         10  10
//	Psionics             10  10
//
// 對應 enums.go 的 Technology 常數;手冊未列的科技一律回 0(不誤加成)。
func SpyTechnologyBonus(tech Technology) int {
	switch tech {
	case TECH_NEURAL_SCANNER:
		return 10
	case TECH_TELEPATHIC_TRAINING:
		return 5
	case TECH_CYBERSECURITY_LINK:
		return 10
	case TECH_STEALTH_SUIT:
		return 10
	case TECH_PSIONICS:
		return 10
	default:
		return 0
	}
}

// 手冊原文(Spy Bonuses 表,Leaders 列):「Telepath 2 to 18 - / Spy Master - 2 to 18」。
// 只給範圍,未列 leader 技能等級 → 加成值的對應公式,故只保留上下限常數。
//
// TODO 手冊未明列 leader 技能等級對應到 2~18 加成的精確映射公式,待查證(可能需要對照存檔/
// 其他社群資料);目前不提供 SpyLeaderXxxBonus(skillLevel) 函式,避免臆測。
const (
	SpyLeaderTelepathDefenseMin  = 2
	SpyLeaderTelepathDefenseMax  = 18
	SpyLeaderSpyMasterOffenseMin = 2
	SpyLeaderSpyMasterOffenseMax = 18
)

// 手冊原文(Spy Bonuses 表,Assets 列):「Agents 0 to 41 / Spies 2 to 41」。
// 這是 SpySlotBonus(0..63) 實際落在遊戲內的觀察範圍:防守方(Agents)可以是 0 人;
// 攻擊方(Spies)至少要派 1 人才有效果(SpySlotBonus(1)=2),上限則是 SpySlotBonus(63)=41。
// 純粹重申 SpySlotBonus 的值域,不是獨立公式,故不另外建函式。
const (
	SpyAssetAgentsMin = 0
	SpyAssetAgentsMax = 41
	SpyAssetSpiesMin  = 2
	SpyAssetSpiesMax  = 41
)

// 手冊原文(Assassins):「The chance to assassinate an enemy spy depends on a leader's skill
// level and varies from +2% to +18% per turn.」只給範圍,未列 leader 技能等級 → 機率的對應公式。
//
// TODO 手冊未明列 leader 技能等級對應到 2%~18% 暗殺機率的精確映射公式,待查證。
const (
	SpyAssassinChanceMin = 0.02 // +2% / turn
	SpyAssassinChanceMax = 0.18 // +18% / turn
)

// 手冊原文(Roll Chance → Action Thresholds):
//
//	stealing is successful:                          80
//	stealing is successful and frame another race:    90
//	sabotage is successful:                           70
//	sabotage is successful and frame another race:    90
const (
	SpyThresholdSteal            = 80
	SpyThresholdStealAndFrame    = 90
	SpyThresholdSabotage         = 70
	SpyThresholdSabotageAndFrame = 90
)

// SpyEffectiveThreshold 有效門檻 E = T + DB - AB(手冊 Roll Chance → Definitions)。
// t = action threshold(見 SpyThresholdXxx 常數)、db = defender bonus、ab = attacker bonus。
func SpyEffectiveThreshold(t, db, ab int) int {
	return t + db - ab
}

// SpyRollChance 給定有效門檻 E,回傳攻擊方本回合成功的機率 p(0..1)。
// 手冊原文(Roll Chance → Formula):
//
//	Each turn both attacker and defender get a roll(100). A spying action is successful if:
//	AR = 100 (aka lucky roll)  或  AR - DR > E
//
//	E <= -100            : p = 1
//	-100 <= E <= -1       : p = 1 - (101+E)*(100+E)/2/10000
//	0 <= E <= 99          : p =      (99-E)*(98-E)/2/9900 + 0.01
//	99 <= E               : p = 0.01
//
// 這個 p 已經是「雙方各擲一顆 1..100 骰子、且 AR=100 必勝」規則的封閉解機率,呼叫端只要拿 p
// 跟 [0,1) 均勻亂數比較即可判定成功與否,不必真的模擬 AR/DR 兩顆骰子。
// RNG 擲骰、以及成功後是否嫁禍(frame,取決於是否為 AR=100 的幸運骰)不在此實作,留給上層。
func SpyRollChance(e int) float64 {
	switch {
	case e <= -100:
		return 1
	case e <= -1:
		return 1 - float64((101+e)*(100+e))/2/10000
	case e <= 99:
		return float64((99-e)*(98-e))/2/9900 + 0.01
	default:
		return 0.01
	}
}

// SpyVsSpyDefenderBonus Spy vs Spy(間諜互殺)判定時 defender 的加成。
// 手冊原文(Spy vs Spy):「the defender gets an extra +20 bonus」。
func SpyVsSpyDefenderBonus(db int) int {
	return db + 20
}

// SpyVsSpyAttackerBonus Spy vs Spy(間諜互殺)判定時 attacker 選擇 HIDE 指令的加成。
// 手冊原文(Spy vs Spy):「the attacker gets +20 if he has chosen HIDE」。
func SpyVsSpyAttackerBonus(ab int, hide bool) int {
	if hide {
		return ab + 20
	}
	return ab
}

// 手冊原文(Spy vs Spy):「At +80 a defender is killed, and at -80 an attacker[is killed], and
// both parties have the possibility to kill with lucky rolls, so both can lose a spy in the
// same turn.」
//
// TODO 手冊未明列 Spy vs Spy 判定用的 action threshold(T)基準值(steal/sabotage 各有 T=70/80/90,
// 互殺這節完全沒提 T 是多少),也未明列 ±80 這兩個門檻與 SpyEffectiveThreshold/SpyRollChance 的
// 精確對應公式(是直接比較 AR-DR、還是先套 SpyRollChance 曲線再看機率),待查證;
// 這裡只忠實保留手冊給出的加成規則與門檻數字,不臆測其餘部分。
const (
	SpyVsSpyDefenderKillThreshold = 80
	SpyVsSpyAttackerKillThreshold = -80
)
