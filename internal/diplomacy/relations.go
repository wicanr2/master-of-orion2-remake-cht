// Package diplomacy 是外交關係系統的【設計性重建】。
//
// ⚠ 重要:這不是原版 MOO2 行為。MOO2 的外交關係每回合如何升降,官方手冊未給、
// 英語社群也公認未破解(見 docs/tech/community-mechanics-findings.md)。本套件在使用者
// 授權下,做一套「合理但非原版」的關係設計,供 remake 有可運作的外交層。
//
// 【原版資料 vs 本套件設計】的界線:
//   - 原版(權威):17 級關係「名稱量表」(FEUD…HARMONY)——來自遊戲資料 BILLTEXT，已翻譯於
//     assets/i18n/misc.tsv;本套件的 RelationLevel 名稱與之對齊,顯示時可經 i18n 翻成中文。
//   - 本套件設計(非原版):數值分數(RelationScore)、分數→等級的對映、各事件的調整值、
//     每回合往中立漂移的速率——這些數字都是設計選擇,原版實際值未知。
package diplomacy

// RelationLevel 是 17 級外交關係(由最敵對到最友好)。名稱對齊 MOO2 原版資料(misc.tsv 的英文 key),
// 顯示時用該英文名經 i18n 翻譯。
type RelationLevel int

const (
	RelationFeud RelationLevel = iota // 世仇(最敵對)
	RelationHate
	RelationDiscord
	RelationTroubled
	RelationTense
	RelationRestless
	RelationWary
	RelationUnease
	RelationNeutral // 中立(分數 0 附近)
	RelationRelaxed
	RelationAmiable
	RelationCalm
	RelationAffable
	RelationPeaceful
	RelationFriendly
	RelationUnity
	RelationHarmony // 和睦(最友好)
	relationLevelCount
)

// relationLevelNames 是各級對應的英文 key(與 misc.tsv 一致,供 i18n 翻譯顯示)。
var relationLevelNames = [relationLevelCount]string{
	"FEUD", "HATE", "DISCORD", "TROUBLED", "TENSE", "RESTLESS", "WARY", "UNEASE",
	"NEUTRAL", "RELAXED", "AMIABLE", "CALM", "AFFABLE", "PEACEFUL", "FRIENDLY", "UNITY", "HARMONY",
}

// Name 回傳該級的英文 key(顯示時經 i18n.Translate 轉中文)。
func (l RelationLevel) Name() string {
	if l < 0 || int(l) >= len(relationLevelNames) {
		return "NEUTRAL"
	}
	return relationLevelNames[l]
}

// 關係分數範圍(設計):[-RelationScoreMax, +RelationScoreMax],0 為中立。
const RelationScoreMax = 100

// RelationLevelForScore 把分數對映到 17 級(設計:對稱分佈,0 落在 NEUTRAL 帶中央)。
// 每級寬度 = 2*Max/17,NEUTRAL(index 8)含 0。
func RelationLevelForScore(score int) RelationLevel {
	if score < -RelationScoreMax {
		score = -RelationScoreMax
	}
	if score > RelationScoreMax {
		score = RelationScoreMax
	}
	// 把 [-Max,+Max] 線性映到 [0,16],四捨五入。
	n := int(relationLevelCount) // 17
	// shifted 到 [0, 2*Max],再縮到 [0, n-1]。
	idx := (score + RelationScoreMax) * (n - 1) / (2 * RelationScoreMax)
	if idx < 0 {
		idx = 0
	}
	if idx >= n {
		idx = n - 1
	}
	return RelationLevel(idx)
}

// IsAtWar 依等級判定是否處於敵對狀態(設計:WARY 以下視為敵對傾向)。
func (l RelationLevel) IsHostile() bool { return l < RelationWary }

// IsFriendly 依等級判定是否友好(設計:AFFABLE 以上視為友好)。
func (l RelationLevel) IsFriendly() bool { return l > RelationAffable }
