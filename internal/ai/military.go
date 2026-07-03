// military.go:AI 決定殖民地下一個生產項目的優先類別【設計性重建,非原版 MOO2】。
//
// ⚠ 重要:MOO2 官方未公開生產排程 AI 的決策邏輯,社群也公認未逆向出完整規則
// (見 docs/tech/community-mechanics-findings.md)。本檔是在使用者授權下設計的一套
// 合理但非原版的生產優先啟發式,權重/門檻/判斷次序均為設計選擇。
package ai

// BuildPriority 是 AI 決定殖民地下一個生產項目的類別【設計】。
type BuildPriority int

const (
	BuildColonyInfrastructure BuildPriority = iota // 殖民建設(工廠/實驗室等)
	BuildWarships                                  // 戰艦
	BuildColonyShip                                // 殖民船(擴張)
	BuildDefenses                                  // 防禦(飛彈基地/星際基地)
)

// DecideBuildPriority 依 AI 性格、威脅與擴張機會決定生產優先【設計啟發式】。
//
// 判斷次序(由高到低,前者成立即回傳,不再往下看):
//  1. 受敵對勢力威脅(threatenedByEnemy):生存優先——好戰性格(IndustryWeight>ResearchWeight)
//     傾向主動造戰艦迎戰;其餘性格(平衡/科學/防禦傾向)選擇造防禦設施固守。
//  2. 基礎建設未完成(!infrastructureComplete):不論性格,先把殖民地基礎建設補齊,
//     否則後續工業/研究產出都受限,優先度高於擴張與默認軍備生產。
//  3. 有可殖民目標(hasColonizableTarget)且性格傾向擴張(expansionist 或 balanced,
//     以 Profile.Name 判斷):把握擴張機會造殖民船。好戰/科學性格不在此列——
//     好戰性格默認寧可造艦,科學性格默認寧可持續建設。
//  4. 其餘情況依性格 fallback:好戰性格造戰艦(持續備戰),其餘(科學/平衡等)
//     持續投入殖民建設。
func DecideBuildPriority(p Profile, threatenedByEnemy, hasColonizableTarget, infrastructureComplete bool) BuildPriority {
	warlike := p.IndustryWeight > p.ResearchWeight // 設計:沿用 diplomacy.go 的好戰判斷準則

	switch {
	case threatenedByEnemy:
		if warlike {
			return BuildWarships
		}
		return BuildDefenses
	case !infrastructureComplete:
		return BuildColonyInfrastructure
	case hasColonizableTarget && isExpansionOriented(p):
		return BuildColonyShip
	default:
		if warlike {
			return BuildWarships
		}
		return BuildColonyInfrastructure
	}
}

// isExpansionOriented 判斷性格是否傾向「有機會就擴張」【設計】:
// 以 Profile.Name 對齊 economy.go 定義的 ProfileExpansionist / ProfileBalanced。
// 好戰性格(aggressive)偏好直接備戰、科學性格(scientific)偏好持續研究建設,
// 兩者即使有可殖民目標也不視為擴張優先。
func isExpansionOriented(p Profile) bool {
	return p.Name == "expansionist" || p.Name == "balanced"
}
