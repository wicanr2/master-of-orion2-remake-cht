package shell

import (
	"fmt"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// TreatyState 是玩家與單一 AI 對手之間的外交協議狀態。
//
// 原版證據：外交資料結構在 +0x627 儲存互不侵犯／同盟／和平的正式狀態，
// +0x62F 與 +0x637 分別是貿易、研究協議旗標；兩種協議的目前收益與目標
// 另存於 +0x5A4/+0x5B4 與 +0x5C6/+0x5D6。這裡保留同一個「一個正式
// 狀態 + 兩個可並存協議」的形狀，欄位名稱是 remake 語意，不取代原始位址。
//
// 原版特殊貿易表的可執行部分已追回：政府／神級商人倍率後，原版會掃描
// `0x3B×0x43` 領袖表，只納入同帝國、RawStatus<3 的 Trader，取最大的
// `(experience bucket+1)*10`／`*15` 加成。help.tsv 的 +25% 是過期敘述，
// 與原始指令衝突；未知只剩 raw player trait／特殊領袖欄位在不同 runtime 的
// 上游來源，不把食物／研究交換切片冒充原版 byte table。
type TreatyState struct {
	FormalPolicy   gamedata.ForeignPolicy `json:"formalPolicy,omitempty"`
	TradeActive    bool                   `json:"tradeActive,omitempty"`
	ResearchActive bool                   `json:"researchActive,omitempty"`
	// PlayerTribute／AITribute 對應原版關係矩陣 +0x63F 的兩個方向。
	// 原版固定選項寫入 raw mode 1 或 2，分別代表 5%／10% 的週期納貢；
	// 這裡保留 raw mode，而不是把它誤稱為百分比，方便直接回查反組譯。
	PlayerTribute TributeMode `json:"playerTribute,omitempty"`
	AITribute     TributeMode `json:"aiTribute,omitempty"`
	TradeTurns    int         `json:"tradeTurns,omitempty"`
	ResearchTurns int         `json:"researchTurns,omitempty"`

	// 負值起步是原版 Start_*_Agreement_ 把目前值初始化為負的基準目標之半；
	// 它讓百科所說的「第 5 回合損益兩平」有可觀察的狀態，而不是只顯示旗標。
	PlayerTradeValue    int               `json:"playerTradeValue,omitempty"`
	AITradeValue        int               `json:"aiTradeValue,omitempty"`
	PlayerResearchValue int               `json:"playerResearchValue,omitempty"`
	AIResearchValue     int               `json:"aiResearchValue,omitempty"`
	SpecialTrade        SpecialTradeState `json:"specialTrade,omitempty"`
}

// SpecialTradeKind 是外交畫面的特殊貿易選項。原版特殊貿易表的完整
// byte-to-row 對照尚未追回，這些名稱與收益是可保存、可測試的 remake
// 垂直切片，不冒充原版未解出的表格。
type SpecialTradeKind uint8

const (
	SpecialTradeNone SpecialTradeKind = iota
	SpecialTradeFoodForCredits
	SpecialTradeResearchExchange
)

// SpecialTradeState 是一份與普通貿易／研究協議可並存的特殊交換狀態。
type SpecialTradeState struct {
	Kind        SpecialTradeKind `json:"kind,omitempty"`
	Active      bool             `json:"active,omitempty"`
	Turns       int              `json:"turns,omitempty"`
	PlayerValue int              `json:"playerValue,omitempty"`
	AIValue     int              `json:"aiValue,omitempty"`
}

// TributeMode 是原版 +0x63F 的已解出固定納貢模式。
//
// 已證實：原版 sub_194C5 的前兩個納貢選項呼叫 sub_52049，分別傳入
// EBX=1／2；sub_E1FC7 以 mode×(收入−既有納貢成本)/20 算出 BC，
// 介面字串 GIVING／RECEIVING 也把兩者顯示為 5%／10%。
type TributeMode uint8

const (
	TributeNone TributeMode = iota
	TributeFivePercent
	TributeTenPercent
)

// TributePercent 回傳畫面與玩家可讀資料使用的實際百分比。
func (m TributeMode) TributePercent() int {
	switch m {
	case TributeFivePercent:
		return 5
	case TributeTenPercent:
		return 10
	default:
		return 0
	}
}

func validTributeMode(m TributeMode) bool {
	return m == TributeFivePercent || m == TributeTenPercent
}

// tributeCost 對應原版 sub_E1FC7：收入與既有納貢成本以 BC 整數表示時，
// raw mode 1／2 正好是 5%／10%。負收入不會再產生納貢成本；多個收款方
// 由呼叫端把前一筆成本傳入，保持原版「收入−既有納貢成本」的資料流。
func tributeCost(grossIncomeBC, existingTributeCost int, mode TributeMode) int {
	if !validTributeMode(mode) {
		return 0
	}
	base := grossIncomeBC - existingTributeCost
	if base <= 0 {
		return 0
	}
	return int(mode) * base / 20
}

// TreatyYield 是一回合協議結算給雙方的 BC／研究點。它是回合內暫態，不進存檔。
type treatyYield struct {
	PlayerBC       int
	PlayerResearch int
	AIBC           int
	AIResearch     int
}

func treatyBase(minPopulation int) int {
	if minPopulation < 0 {
		return 0
	}
	return minPopulation / 2
}

// treatyTradeGovernmentPercent 對應原版 sub_101BA4 的政府分支。
// 原始 +0x89F 是目前政府／種族特性陣列；政府編號與 remake 的
// MoraleGovernmentType 已由 RACESTUF.LBX、SAVE10 與既有 government tests 交叉釘住。
func treatyTradeGovernmentPercent(gov gamedata.MoraleGovernmentType) int {
	switch gov {
	case gamedata.MoraleGovDemocracy:
		return 150
	case gamedata.MoraleGovFederation:
		return 175
	default:
		return 100
	}
}

// treatyResearchGovernmentPercent 對應原版 sub_101CC5 的政府分支。
// 0/1/4/5 分別是 1/2、3/4、3/2、7/4；其餘已知政府走基準值。
func treatyResearchGovernmentPercent(gov gamedata.MoraleGovernmentType) int {
	switch gov {
	case gamedata.MoraleGovFeudalism:
		return 50
	case gamedata.MoraleGovConfederation:
		return 75
	case gamedata.MoraleGovDemocracy:
		return 150
	case gamedata.MoraleGovFederation:
		return 175
	default:
		return 100
	}
}

// treatyTarget 是目前可由證據支持的最小目標模型。
// 原版 sub_101B3C 先取雙方 +0x5A2/+0x5C4 的較小值再除以二；這裡以帝國

// 總人口作為 remake 對應欄位。原版 sub_101BA4 對 +0x89F 的政府值套用
// Democracy=150%、Federation=175%，再對 +0x8B7（特性 24，神級商人）
// 加上 50 個百分點；sub_101CC5 則依政府值套用研究倍率。原版另外還會
// 掃描特殊貿易表取更高值，該表與 remake 的特殊項目尚未完成映射，因此不在此
// 捏造。
func treatyTarget(base int, government gamedata.MoraleGovernmentType, fantasticTrader bool, trade bool) int {
	return treatyTargetWithLeader(base, government, fantasticTrader, trade, 0)
}

func treatyTargetWithLeader(base int, government gamedata.MoraleGovernmentType, fantasticTrader bool, trade bool, leaderBonus int) int {
	pct := treatyResearchGovernmentPercent(government)
	if trade {
		pct = treatyTradeGovernmentPercent(government)
		if originalPct, ok := gamedata.OriginalTradeAgreementGoalPercentWithLeader(int(government), fantasticTrader, leaderBonus); ok {
			pct = originalPct
		}
	}
	return base * pct / 100
}

// treatyApproach 對應原版協議處理器的五回合趨近方式；加入確定性的最小步進，
// 避免小帝國因整數除法永遠停在目標前。
func treatyApproach(current, target int) int {
	if current == target {
		return current
	}
	delta := (target - current) / 5
	if delta == 0 {
		if target > current {
			delta = 1
		} else {
			delta = -1
		}
	}
	next := current + delta
	if target > current && next > target {
		return target
	}
	if target < current && next < target {
		return target
	}
	return next
}

func (t *TreatyState) startFormal(policy gamedata.ForeignPolicy) bool {
	if policy != gamedata.DIPLO_NON_AGGRESSION && policy != gamedata.DIPLO_ALLIANCE && policy != gamedata.DIPLO_PEACE {
		return false
	}
	if t.FormalPolicy != gamedata.DIPLO_NONE {
		return false
	}
	t.FormalPolicy = policy
	return true
}

func (t *TreatyState) startTrade(minPopulation int) bool {
	if t.TradeActive {
		return false
	}
	base := treatyBase(minPopulation)
	t.TradeActive = true
	t.TradeTurns = 0
	t.PlayerTradeValue = -base
	t.AITradeValue = -base
	return true
}

func (t *TreatyState) startResearch(minPopulation int) bool {
	if t.ResearchActive {
		return false
	}
	base := treatyBase(minPopulation)
	t.ResearchActive = true
	t.ResearchTurns = 0
	t.PlayerResearchValue = -base
	t.AIResearchValue = -base
	return true
}

func validSpecialTradeKind(kind SpecialTradeKind) bool {
	return kind == SpecialTradeFoodForCredits || kind == SpecialTradeResearchExchange
}

func (t *TreatyState) startSpecialTrade(kind SpecialTradeKind, minPopulation int) bool {
	if !validSpecialTradeKind(kind) || t.SpecialTrade.Active {
		return false
	}
	base := treatyBase(minPopulation)
	t.SpecialTrade = SpecialTradeState{
		Kind: kind, Active: true, Turns: 0,
		PlayerValue: -base, AIValue: -base,
	}
	return true
}

func (t *TreatyState) endSpecialTrade() bool {
	if !t.SpecialTrade.Active {
		return false
	}
	t.SpecialTrade = SpecialTradeState{}
	return true
}

// startPlayerTribute 建立玩家→AI 的週期納貢。原版同一對帝國不能同時
// 維持兩個方向的納貢條約；目前 remake 只把已解出的固定 5%／10% 選項
// 接入，未解出的現金／科技／星系一次性餽贈不在此函式混用。
func (t *TreatyState) startPlayerTribute(mode TributeMode) bool {
	if !validTributeMode(mode) || t.PlayerTribute != TributeNone || t.AITribute != TributeNone {
		return false
	}
	t.PlayerTribute = mode
	return true
}

// startAITribute 供 AI 會談／測試接入已解出的另一方向資料形狀。
func (t *TreatyState) startAITribute(mode TributeMode) bool {
	if !validTributeMode(mode) || t.PlayerTribute != TributeNone || t.AITribute != TributeNone {
		return false
	}
	t.AITribute = mode
	return true
}

func (t *TreatyState) endFormal() bool {
	if t.FormalPolicy == gamedata.DIPLO_NONE {
		return false
	}
	t.FormalPolicy = gamedata.DIPLO_NONE
	return true
}

func (t *TreatyState) endTrade() bool {
	if !t.TradeActive {
		return false
	}
	t.TradeActive = false
	t.TradeTurns = 0
	t.PlayerTradeValue = 0
	t.AITradeValue = 0
	return true
}

func (t *TreatyState) endResearch() bool {
	if !t.ResearchActive {
		return false
	}
	t.ResearchActive = false
	t.ResearchTurns = 0
	t.PlayerResearchValue = 0
	t.AIResearchValue = 0
	return true
}

// endTribute 終止這一對帝國目前的納貢方向。原版 sub_51C02 清除雙向
// +0x63F，故一次操作同時清空玩家付出與 AI 付出的狀態。
func (t *TreatyState) endTribute() bool {
	if t.PlayerTribute == TributeNone && t.AITribute == TributeNone {
		return false
	}
	t.PlayerTribute = TributeNone
	t.AITribute = TributeNone
	return true
}

// BlocksOffensive 回報哪些正式狀態禁止攻勢。原版 help 明確寫出互不侵犯與同盟
// 不得互相攻擊；和平也沿用同一條「無敵對行動」的重製守門。
func (t TreatyState) BlocksOffensive() bool {
	switch t.FormalPolicy {
	case gamedata.DIPLO_NON_AGGRESSION, gamedata.DIPLO_ALLIANCE, gamedata.DIPLO_PEACE:
		return true
	default:
		return false
	}
}

func (t *TreatyState) advance(minPopulation int, playerGovernment, aiGovernment gamedata.MoraleGovernmentType, playerFantasticTrader, aiFantasticTrader bool, playerLeaderBonus, aiLeaderBonus int) treatyYield {
	base := treatyBase(minPopulation)
	var y treatyYield
	if t.TradeActive {
		// 相容舊存檔中若只有旗標、沒有目前值，補上原版相同的負起點。
		if t.TradeTurns == 0 && t.PlayerTradeValue == 0 && t.AITradeValue == 0 {
			t.PlayerTradeValue, t.AITradeValue = -base, -base
		}
		t.PlayerTradeValue = treatyApproach(t.PlayerTradeValue, treatyTargetWithLeader(base, playerGovernment, playerFantasticTrader, true, playerLeaderBonus))
		t.AITradeValue = treatyApproach(t.AITradeValue, treatyTargetWithLeader(base, aiGovernment, aiFantasticTrader, true, aiLeaderBonus))
		t.TradeTurns++
		y.PlayerBC, y.AIBC = t.PlayerTradeValue, t.AITradeValue
	}
	if t.ResearchActive {
		if t.ResearchTurns == 0 && t.PlayerResearchValue == 0 && t.AIResearchValue == 0 {
			t.PlayerResearchValue, t.AIResearchValue = -base, -base
		}
		t.PlayerResearchValue = treatyApproach(t.PlayerResearchValue, treatyTarget(base, playerGovernment, false, false))
		t.AIResearchValue = treatyApproach(t.AIResearchValue, treatyTarget(base, aiGovernment, false, false))
		t.ResearchTurns++
		y.PlayerResearch, y.AIResearch = t.PlayerResearchValue, t.AIResearchValue
	}
	if t.SpecialTrade.Active {
		value := base / 4
		if value < 1 {
			value = 1
		}
		t.SpecialTrade.Turns++
		switch t.SpecialTrade.Kind {
		case SpecialTradeFoodForCredits:
			t.SpecialTrade.PlayerValue += value
			t.SpecialTrade.AIValue += value
			y.PlayerBC += value
			y.AIBC += value
		case SpecialTradeResearchExchange:
			t.SpecialTrade.PlayerValue += value
			t.SpecialTrade.AIValue += value
			y.PlayerResearch += value
			y.AIResearch += value
		}
	}
	return y
}

func populationOfColonies(colonies []engine.ColonyState) int {
	total := 0
	for _, c := range colonies {
		total += c.Population
	}
	return total
}

func aiHasFantasticTrader(a AIOpponent) bool {
	idx := aiRaceIndex(a)
	if idx < 0 || idx >= len(Races) {
		return false
	}
	return gamedata.OrigRaceHasTrait(Races[idx].OrigIdx, gamedata.TRAIT_FANTASTIC_TRADERS)
}

// tradeAgreementLeaderBonus 對應 sub_101BA4 的活動領袖掃描。原版以 raw
// status<3 篩出同帝國領袖，再只看 CommonSkills 中 Trader 的 tier 1/2
// (raw +0x28 的 bit 0x04/0x08)，每位算一次、最後取最大值。GAM 匯入時
// RawExperience 是 +0x24 的真值；舊 JSON／demo 沒有它時，以顯示等級的
// bucket 起點作相容 fallback，並不宣稱那是原版精確經驗。
func tradeAgreementLeaderBonus(leaders []Leader, warlord bool) int {
	best := 0
	for _, leader := range leaders {
		if leader.RawStatus < 0 || leader.RawStatus >= 3 {
			continue
		}
		tier := leaderSkillTier(leader, int(gamedata.SKILL_TRADER))
		if tier <= 0 {
			continue
		}
		experience := leader.RawExperience
		if !leader.RawExperienceKnown {
			expBucket := leaderDisplayLevelToExpLevel(leader.Level)
			experience = [...]int{0, 60, 150, 300, 500}[expBucket]
		}
		bonus := gamedata.OriginalSpecialTradeLeaderBonus(experience, tier, warlord, leader.ID)
		if bonus > best {
			best = bonus
		}
	}
	return best
}

// aiTreatyGovernment 是目前 remake 的 AI 政體邊界。
// AIOpponent 尚未保存「目前政府」欄位，殖民地建立與 AI 經濟也都明確使用
// Dictatorship 作為保守預設；不把 RaceTrait[0] 的開局政體冒充成遊戲中的
// 目前政體。等 AI 政體研究／升級切片完成後，再由該欄位取代這個固定值。
func aiTreatyGovernment() gamedata.MoraleGovernmentType {
	return gamedata.MoraleGovDictatorship
}

// advanceTreaties 每個世界回合只推進一次所有玩家↔AI 協議。回傳切片與 AIPlayers
// 同索引，讓同一份協議結果分別餵給兩個帝國的經濟結算，不會把共享狀態推進兩次。
func (s *GameSession) advanceTreaties() []treatyYield {
	yields := make([]treatyYield, len(s.AIPlayers))
	playerPopulation := populationOfColonies(s.PlayerColonies)
	playerFantasticTrader := s.RaceFantasticTrader()
	playerGovernment := s.effectiveGovernment()
	playerLeaderBonus := tradeAgreementLeaderBonus(s.Leaders, s.RaceWarlord())
	aiGovernment := aiTreatyGovernment()
	for i := range s.AIPlayers {
		ai := &s.AIPlayers[i]
		aiPopulation := populationOfColonies(ai.Colonies)
		minPopulation := playerPopulation
		if aiPopulation < minPopulation {
			minPopulation = aiPopulation
		}
		yields[i] = ai.Treaty.advance(minPopulation, playerGovernment, aiGovernment, playerFantasticTrader, aiHasFantasticTrader(*ai), playerLeaderBonus, tradeAgreementLeaderBonus(ai.Leaders, aiRaceHasTrait(*ai, gamedata.TRAIT_WARLORD)))
	}
	return yields
}

// empireGrossBC 是納貢公式的 remake 對應欄位。原版 sub_E1FC7 讀取帝國
// 收入欄位 +0xAE，再扣該帝國已累積的 +0xE78 納貢成本；這裡以 engine
// 已拆出的稅收、餘糧、貿易品與協議收入總和承接 +0xAE。維護費不放入
// gross，因原版的納貢成本是在收入計算段另行累加。
func empireGrossBC(out engine.EmpireOutput) int {
	return out.TaxRevenue + out.FoodSurplusRevenue + out.TradeGoodsRevenue + out.TreatyIncomeBC
}

// applyTributeTransfers 在玩家與 AI 都完成本回合經濟後，依 +0x63F 的方向
// 移轉週期納貢。玩家可能同時向多個帝國納貢，後一筆會扣除前一筆成本，
// 對應原版 sub_E2710 逐對手累加 +0xE78 的順序；AI 目前只有玩家一個
// 可收款對象，因此各自從零成本開始計算。
func (s *GameSession) applyTributeTransfers(playerGross int, aiGross []int) {
	playerPaid := 0
	netTransfer := 0
	for i := range s.AIPlayers {
		t := &s.AIPlayers[i].Treaty
		if t.PlayerTribute != TributeNone {
			amount := tributeCost(playerGross, playerPaid, t.PlayerTribute)
			playerPaid += amount
			s.Player.BC -= amount
			s.AIPlayers[i].Player.BC += amount
			netTransfer -= amount
		}
	}
	for i := range s.AIPlayers {
		t := &s.AIPlayers[i].Treaty
		if t.AITribute == TributeNone || i >= len(aiGross) {
			continue
		}
		amount := tributeCost(aiGross[i], 0, t.AITribute)
		s.AIPlayers[i].Player.BC -= amount
		s.Player.BC += amount
		netTransfer += amount
	}
	// LastPlayerOutput 是畫面與測試讀取的玩家單回合摘要；Player.BC 在上方
	// 已先完成移轉，因此兩者必須一起更新，避免顯示值與可用國庫分叉。
	s.LastPlayerOutput.TributeCost = playerPaid
	s.LastPlayerOutput.NetBC += netTransfer
	s.LastPlayerOutput.Player.BC = s.Player.BC
}

// TreatyFor 回傳指定對手目前協議狀態的副本，避免 UI 不經外交回應與關係調整
// 就直接修改外交狀態。
func (s *GameSession) TreatyFor(enemy string) TreatyState {
	if ai := s.aiByDisplayName(enemy); ai != nil {
		return ai.Treaty
	}
	return TreatyState{}
}

// TreatyFormalName 回傳原版 ForeignPolicy 值的顯示名稱。
func TreatyFormalName(policy gamedata.ForeignPolicy, english bool) string {
	if english {
		switch policy {
		case gamedata.DIPLO_NON_AGGRESSION:
			return "Non-Aggression Pact"
		case gamedata.DIPLO_ALLIANCE:
			return "Alliance"
		case gamedata.DIPLO_PEACE:
			return "Peace Treaty"
		default:
			return "None"
		}
	}
	switch policy {
	case gamedata.DIPLO_NON_AGGRESSION:
		return "互不侵犯條約"
	case gamedata.DIPLO_ALLIANCE:
		return "同盟"
	case gamedata.DIPLO_PEACE:
		return "和平條約"
	default:
		return "無正式條約"
	}
}

// TreatySummary 是外交畫面目前協議的單行摘要。
func TreatySummary(t TreatyState, english bool) string {
	formal := TreatyFormalName(t.FormalPolicy, english)
	parts := make([]string, 0, 5)
	if t.FormalPolicy != gamedata.DIPLO_NONE {
		parts = append(parts, formal)
	}
	if t.PlayerTribute != TributeNone {
		if english {
			parts = append(parts, fmt.Sprintf("Paying tribute %d%%", t.PlayerTribute.TributePercent()))
		} else {
			parts = append(parts, fmt.Sprintf("進貢%d%%", t.PlayerTribute.TributePercent()))
		}
	}
	if t.AITribute != TributeNone {
		if english {
			parts = append(parts, fmt.Sprintf("Receiving tribute %d%%", t.AITribute.TributePercent()))
		} else {
			parts = append(parts, fmt.Sprintf("收取%d%%進貢", t.AITribute.TributePercent()))
		}
	}
	if t.TradeActive {
		if english {
			parts = append(parts, fmt.Sprintf("Trade %d turns (%d BC)", t.TradeTurns, t.PlayerTradeValue))
		} else {
			parts = append(parts, fmt.Sprintf("貿易第%d回合(%d BC)", t.TradeTurns, t.PlayerTradeValue))
		}
	}
	if t.ResearchActive {
		if english {
			parts = append(parts, fmt.Sprintf("Research %d turns (%d RP)", t.ResearchTurns, t.PlayerResearchValue))
		} else {
			parts = append(parts, fmt.Sprintf("研究第%d回合(%d RP)", t.ResearchTurns, t.PlayerResearchValue))
		}
	}
	if t.SpecialTrade.Active {
		if english {
			parts = append(parts, fmt.Sprintf("Special trade %d turns", t.SpecialTrade.Turns))
		} else {
			label := "特殊貿易"
			if t.SpecialTrade.Kind == SpecialTradeFoodForCredits {
				label = "食物換現金"
			} else if t.SpecialTrade.Kind == SpecialTradeResearchExchange {
				label = "研究交換"
			}
			parts = append(parts, fmt.Sprintf("%s第%d回合", label, t.SpecialTrade.Turns))
		}
	}
	if len(parts) == 0 {
		return formal
	}
	sep := "、"
	if english {
		sep = ", "
	}
	return joinTreatyParts(parts, sep)
}

func joinTreatyParts(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for _, part := range parts[1:] {
		result += sep + part
	}
	return result
}
