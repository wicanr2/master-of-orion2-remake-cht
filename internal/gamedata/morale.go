package gamedata

// 殖民地士氣(Morale)可驗證公式與查表,移植自 moo2_patch1.5/GAME_MANUAL.pdf(繁中化工作用英文版
// 官方手冊;MANUAL_150.html 只含 1.5 版 changelog/選項說明,無 morale 章節)。openorion2 未實作
// morale 邏輯(gamestate.cpp 無對應計算),故以手冊原文數字為唯一權威來源。手冊沒有給精確數字的
// 項目(如 Spiritual Leader 領袖技能的 morale 加成、Tactics 技能)一律不移植,見檔尾 TODO 清單。
//
// 核心公式(p.65-66「Morale is in the top box」+ p.169-170「Morale」章節,兩處原文一致):
//
//	"Each smile represents a 10% bonus to all production (food, industry, research, and
//	income); each frown denotes a 10% penalty to all production."
//	"Every morale icon on the Colony screen represents a change of 10% in the total
//	production output of the colony."
//
// 即:殖民地最終產出(food/industry/research/income 皆同)= 基礎產出 *(100 + moralePercent)/100,
// moralePercent = 下列各來源(政府基礎值、首都淪陷懲罰、建築/成就加成、多種族懲罰…)加總後的百分點。
// 個別來源之間是否可以同時疊加,手冊沒有反例,依各段落原文採直接加總。

// MoraleGovernmentType 政府型態,僅用於本檔 morale 查表。
// enums.go 目前無通用 Government 型別(spy.go 已有相同理由的 SpyGovernmentType),故另建、
// 加 Morale 前綴,避免與其他檔案之後新增的通用 Government 型別撞名。
type MoraleGovernmentType int

const (
	MoraleGovFeudalism MoraleGovernmentType = iota
	MoraleGovConfederation
	MoraleGovDictatorship
	MoraleGovImperium
	MoraleGovDemocracy
	MoraleGovFederation
	MoraleGovUnification
	MoraleGovGalacticUnification
)

// 建築/成就固定加成(手冊原文皆為單一精確數字,非查表)。
const (
	// MoraleHoloSimulatorBonus Holo Simulator(建築):「increases a planet's morale by 20%」(p.95-96)。
	MoraleHoloSimulatorBonus = 20
	// MoralePleasureDomeBonus Pleasure Dome(建築):「increases colony morale by 30%」(p.97-98)。
	MoralePleasureDomeBonus = 30
	// MoraleVirtualRealityNetworkBonus Virtual Reality Network(成就,全帝國生效):
	// 「increases morale by 20% in every colony throughout the entire empire」(p.97-98)。
	MoraleVirtualRealityNetworkBonus = 20
	// MoralePsionicsBonus Psionics(成就)在特定政府下的 morale 加成,見 MoralePsionicsGovernmentBonus。
	MoralePsionicsBonus = 10
)

// MoraleProductionOutput 依士氣百分點計算最終產出(food/industry/research/income 共用同一公式)。
// 手冊(p.65-66、p.169-170):base *(100+moralePercent)/100,moralePercent 可為負(frown)。
func MoraleProductionOutput(base, moralePercent int) int {
	return base * (100 + moralePercent) / 100
}

// moraleGovernmentBaseTable 政府「未建 Barracks」時的基礎士氣百分點(不含首都淪陷懲罰)。
// 手冊原文(Imperial Policy > Government 段,p.165-167,與 p.21-22 敘述一致):
//
//	Feudal / Confederation:      Morale -20% at colonies without Barracks.
//	Dictatorship:                Morale -20% at colonies without Barracks.
//	Imperium:                    All colonies have a 20% Morale bonus. Morale still -20% at
//	                              colonies without Barracks.(兩者疊加,見 MoraleGovernmentBase)
//	Democracy / Federation:      未提及基礎 morale 效果 → 0。
//	Unification / Galactic Unification: 「The Morale of the race's populations cannot be
//	                              modified in any way」→ 不適用本公式,見 MoraleUnificationProductionBonus。
var moraleGovernmentBaseTable = map[MoraleGovernmentType]int{
	MoraleGovFeudalism:           -20,
	MoraleGovConfederation:       -20,
	MoraleGovDictatorship:        -20,
	MoraleGovImperium:            -20, // 「Morale still -20% at colonies without Barracks」,獨立疊加下方 +20% 帝國加成
	MoraleGovDemocracy:           0,
	MoraleGovFederation:          0,
	MoraleGovUnification:         0, // 不適用,見 MoraleUnificationProductionBonus
	MoraleGovGalacticUnification: 0,
}

// moraleImperiumBonus Imperium 額外的帝國全境 morale 加成,與有無 Barracks 無關,手冊:
// 「All colonies have a 20% Morale bonus.」(p.165-167、p.22)。
const moraleImperiumBonus = 20

// MoraleGovernmentBase 回傳政府基礎士氣百分點(不含首都淪陷懲罰、建築/成就加成)。
// hasBarracks = 該殖民地已建 Marine Barracks 或 Armor Barracks(手冊:兩者皆可解除懲罰,p.76-79)。
//
// Feudalism/Confederation/Dictatorship:無 Barracks 時 -20%,有則 0%。
// Imperium:固定 +20%,無 Barracks 時再疊加 -20%(有 Barracks 淨 +20%,無則淨 0%)。
// Democracy/Federation:恆 0%。Unification/Galactic Unification:不適用,回 0(見上方常數說明)。
func MoraleGovernmentBase(gov MoraleGovernmentType, hasBarracks bool) int {
	penalty := moraleGovernmentBaseTable[gov]
	ret := 0
	if !hasBarracks {
		ret += penalty
	}
	if gov == MoraleGovImperium {
		ret += moraleImperiumBonus
	}
	return ret
}

// moraleCapitalCapturedTable 首都被攻陷、新 Capitol 尚未建成前的士氣懲罰(手冊 Imperial Policy >
// Government 段,p.165-167,與各政府段落 p.21-22 原文一致):
//
//	Feudal/Confederation: 「Capture of the capital means anarchy until a new Capitol is
//	                       built. Morale is -50% during this anarchy.」
//	Dictatorship/Imperium:「Capture of the capital means -35% morale at all colonies until
//	                       a new Capitol is built.」
//	Democracy/Federation:「Capture of the capital causes/results in a -20% morale penalty
//	                       until a new Capitol is built.」
//	Unification/Galactic Unification:「Capture of the capital has no effect.」(無 Capitol 建築)
var moraleCapitalCapturedTable = map[MoraleGovernmentType]int{
	MoraleGovFeudalism:           -50,
	MoraleGovConfederation:       -50,
	MoraleGovDictatorship:        -35,
	MoraleGovImperium:            -35,
	MoraleGovDemocracy:           -20,
	MoraleGovFederation:          -20,
	MoraleGovUnification:         0,
	MoraleGovGalacticUnification: 0,
}

// MoraleCapitalCapturedPenalty 回傳首都被攻陷期間的士氣懲罰百分點(見 moraleCapitalCapturedTable)。
func MoraleCapitalCapturedPenalty(gov MoraleGovernmentType) int {
	return moraleCapitalCapturedTable[gov]
}

// MoraleMultiRacialPenalty 多種族殖民地(含未同化的被征服人口)士氣懲罰。手冊原文兩處一致
// (p.66-67「there is a 20% morale penalty on any multi-racial planet without an Alien
// Management Center」;p.92-93「The building also removes the 20% morale penalty from
// multi-racial colonies」)。hasAlienManagementCenter = 該殖民地已建 Alien Management Center。
func MoraleMultiRacialPenalty(hasAlienManagementCenter bool) int {
	if hasAlienManagementCenter {
		return 0
	}
	return -20
}

// MoralePsionicsGovernmentBonus Psionics(成就)的 morale 加成,僅特定政府適用。手冊(p.100-101):
// 「morale is raised by 10% throughout the empire if your government is a Dictatorship,
// Imperium, Feudalism, or Confederation.」(Democracy/Federation/Unification 系列未列入,回 0)
func MoralePsionicsGovernmentBonus(gov MoraleGovernmentType) int {
	switch gov {
	case MoraleGovFeudalism, MoraleGovConfederation, MoraleGovDictatorship, MoraleGovImperium:
		return MoralePsionicsBonus
	default:
		return 0
	}
}

// MoraleUnificationProductionBonus Unification 系政府不使用一般士氣公式(手冊:「Things that boost
// or lower Morale have no effect. The Morale of the race's populations cannot be modified in
// any way.」),改為固定的 food+industry 產出加成(注意:只作用於 food/industry,不含
// research/income,與 MoraleProductionOutput 的四項全用不同,呼叫端不可混用):
//
//	Unification:          Food and Industrial production are +50%.(p.166-167)
//	Galactic Unification: Food and Industrial production are +100%.(p.166-167)
//
// 其餘政府回 0(不適用本函式,一般政府請用 MoraleProductionOutput + MoraleGovernmentBase)。
func MoraleUnificationProductionBonus(gov MoraleGovernmentType) int {
	switch gov {
	case MoraleGovUnification:
		return 50
	case MoraleGovGalacticUnification:
		return 100
	default:
		return 0
	}
}

// 手冊有描述但未給精確公式/數字,故不移植,呼叫端需要時再補查證:
//   - Spiritual Leader(領袖 Administration Ability):「Raises the morale of all colonial
//     populations in the system.」(p.137)——只有文字描述,無百分比或數值。
//   - Tactics(領袖 Military Ability):手冊自陳「This skill is not implemented.」(p.137),
//     與 morale 無關但同段落,一併排除。
