package shell

// DiplomacyResultCode 是外交規則結果的穩定識別字。完整顯示句子由 UI 的外部
// JSON catalog 提供；規則層不得保存玩家可見翻譯。
type DiplomacyResultCode string

// DiplomacyResult 保存格式化外交回應所需的 typed 資料。未使用欄位維持零值。
type DiplomacyResult struct {
	Code      DiplomacyResultCode
	Enemy     string
	Amount    int
	Available int
	Detail    string
}

func diplomacyResult(code DiplomacyResultCode, enemy string) DiplomacyResult {
	return DiplomacyResult{Code: code, Enemy: enemy}
}

const (
	DiploResultFormalExists       DiplomacyResultCode = "formal_exists"
	DiploResultPeaceStrong        DiplomacyResultCode = "peace_strong"
	DiploResultPeaceWeak          DiplomacyResultCode = "peace_weak"
	DiploResultTradeExists        DiplomacyResultCode = "trade_exists"
	DiploResultTradeStarted       DiplomacyResultCode = "trade_started"
	DiploResultResearchExists     DiplomacyResultCode = "research_exists"
	DiploResultResearchStarted    DiplomacyResultCode = "research_started"
	DiploResultSpecialExists      DiplomacyResultCode = "special_exists"
	DiploResultSpecialFoodStarted DiplomacyResultCode = "special_food_started"
	DiploResultSpecialResStarted  DiplomacyResultCode = "special_research_started"
	DiploResultFormalConflict     DiplomacyResultCode = "formal_conflict"
	DiploResultNAPStarted         DiplomacyResultCode = "nonaggression_started"
	DiploResultAllianceStarted    DiplomacyResultCode = "alliance_started"
	DiploResultTributeExists      DiplomacyResultCode = "tribute_exists"
	DiploResultTribute5Started    DiplomacyResultCode = "tribute_5_started"
	DiploResultTribute10Started   DiplomacyResultCode = "tribute_10_started"
	DiploResultNoGiftTech         DiplomacyResultCode = "no_gift_tech"
	DiploResultNoGiftStar         DiplomacyResultCode = "no_gift_star"
	DiploResultNoTrade            DiplomacyResultCode = "no_trade"
	DiploResultTradeEnded         DiplomacyResultCode = "trade_ended"
	DiploResultNoResearch         DiplomacyResultCode = "no_research"
	DiploResultResearchEnded      DiplomacyResultCode = "research_ended"
	DiploResultNoFormal           DiplomacyResultCode = "no_formal"
	DiploResultFormalEnded        DiplomacyResultCode = "formal_ended"
	DiploResultNoTribute          DiplomacyResultCode = "no_tribute"
	DiploResultTributeEnded       DiplomacyResultCode = "tribute_ended"
	DiploResultNoSpecial          DiplomacyResultCode = "no_special"
	DiploResultSpecialEnded       DiplomacyResultCode = "special_ended"
	DiploResultThreatStrong       DiplomacyResultCode = "threat_strong"
	DiploResultThreatWeak         DiplomacyResultCode = "threat_weak"
	DiploResultCashInvalid        DiplomacyResultCode = "cash_invalid"
	DiploResultCashInsufficient   DiplomacyResultCode = "cash_insufficient"
	DiploResultCashAccepted       DiplomacyResultCode = "cash_accepted"
	DiploResultTechUnknown        DiplomacyResultCode = "tech_unknown"
	DiploResultTechKnown          DiplomacyResultCode = "tech_known"
	DiploResultTechAccepted       DiplomacyResultCode = "tech_accepted"
	DiploResultStarInvalid        DiplomacyResultCode = "star_invalid"
	DiploResultStarLastColony     DiplomacyResultCode = "star_last_colony"
	DiploResultStarNotOwned       DiplomacyResultCode = "star_not_owned"
	DiploResultStarAlreadyOwned   DiplomacyResultCode = "star_already_owned"
	DiploResultStarAccepted       DiplomacyResultCode = "star_accepted"
)
