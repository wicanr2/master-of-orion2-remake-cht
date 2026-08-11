package gamedata

// original_diplomacy_oracle.go 保存 IDA 從原版外交／特殊貿易相關 raw table 讀出的
// 完整 16-word 快照。這些陣列的語意只在有直接讀取端的位置升格；尚未能把所有欄位
// 對回 remake 的外交選項時，呼叫端必須保留 raw 位址與「未知」標籤。
//
// 輸入：`Orion2.exe.i64` SHA-256
// `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`，IDA Pro 9.4，
// IDA 線性位址。資料與位址索引也同步記在 docs/re/oracle-static-ida-20260811.md。

const (
	OriginalTradeAgreementValuesAddress = 0x18105C
	OriginalDiplomacyPersonalityAddress = 0x180DB8
	OriginalGiftResponseValuesAddress   = 0x180CCC
	OriginalTradeEventValuesAddress     = 0x181070
	OriginalDiplomacyOracleWordCount    = 16
)

// OriginalTradeAgreementValues 對應 raw 0x18105C；sub_5232E 的 mode 1/2
// 直接以政府索引讀這張表。陣列是 signed word，不把後半的 0／保留列丟掉。
var OriginalTradeAgreementValues = [...]int{5, 10, 20, 5, 50, 40, 5, 0, 0, 0, 7, 10, 1, 1, 0, 0}

// OriginalDiplomacyPersonalityValues 是 raw 0x180DB8 的 signed word 快照。
var OriginalDiplomacyPersonalityValues = [...]int{100, 50, 50, 100, 50, 20, 0, 0, 0, 0, -520, -515, -515, -515, -6, -1}

// OriginalGiftResponseValues 是 raw 0x180CCC 的 signed word 快照。
var OriginalGiftResponseValues = [...]int{-50, -20, -20, 0, 20, 30, -70, 0, 40, 30, 20, 5, 0, -10, 50, 0}

// OriginalTradeEventValues 是 raw 0x181070 的 signed word 快照。
var OriginalTradeEventValues = [...]int{7, 10, 1, 1, 0, 0, 5, 0, -10, -5, -3, 0, 20, 20, -10, 0}

func originalDiplomacyTableValue(table []int, government int) (int, bool) {
	if government < 0 || government >= len(table) {
		return 0, false
	}
	return table[government], true
}

// OriginalTradeStartRelationDelta 對應 sub_5232E 的已證實 mode 分支：mode 1
// 是 raw government value × 3/2，mode 2 是 ×2。其餘 mode 不在該讀取端定義。
func OriginalTradeStartRelationDelta(government, mode int) (int, bool) {
	value, ok := originalDiplomacyTableValue(OriginalTradeAgreementValues[:], government)
	if !ok {
		return 0, false
	}
	switch mode {
	case 1:
		return value * 3 / 2, true
	case 2:
		return value * 2, true
	default:
		return 0, false
	}
}

// OriginalTradeAgreementGoalPercent 對應 sub_101BA4 的政府／特性段，未納入
// 活動領袖加成。bonusTrait 是 raw +0x8B7 非零（神級商人）旗標。
func OriginalTradeAgreementGoalPercent(government int, bonusTrait bool) (int, bool) {
	if government < 0 || government > 15 {
		return 0, false
	}
	percent := 100
	switch government {
	case 4:
		percent = 150
	case 5:
		percent = 175
	}
	if bonusTrait {
		percent += 50
	}
	return percent, true
}

// OriginalLeaderExperienceBucket 對應 sub_93D4B @ 0x93D4B 的 raw 經驗分桶。
// Warlord（raw player +0x8BD）只在經驗 >=1000 時把最高桶由 4 提升為 5；
// leaderID=0x42 是原版特殊領袖，sub_94951 @ 0x94951 對它強制不讀該旗標。
func OriginalLeaderExperienceBucket(experience int, warlord bool, leaderID int) int {
	switch {
	case experience < 60:
		return 0
	case experience < 150:
		return 1
	case experience < 300:
		return 2
	case experience < 500:
		return 3
	case experience >= 1000 && warlord && leaderID != 0x42:
		return 5
	default:
		return 4
	}
}

// OriginalSpecialTradeLeaderBonus 對應 sub_101BA4 掃描活動領袖時的兩列 raw 表：
// Trader tier 1 = (experienceBucket+1)*10，tier 2 = (experienceBucket+1)*15。
// 呼叫端應只傳入 raw status < 3 的活動領袖；這裡保留 leaderID 以維持 0x42
// 特殊領袖的 Warlord 分支差異。
func OriginalSpecialTradeLeaderBonus(experience, traderTier int, warlord bool, leaderID int) int {
	bucket := OriginalLeaderExperienceBucket(experience, warlord, leaderID)
	switch traderTier {
	case 1:
		return (bucket + 1) * 10
	case 2:
		return (bucket + 1) * 15
	default:
		return 0
	}
}

// OriginalTradeAgreementGoalPercentWithLeader 是 sub_101BA4 的完整倍率鏈：
// 基礎政府／raw +0x8B7 特性，再加活動 Trader 領袖掃描所得的最大 bonus。
func OriginalTradeAgreementGoalPercentWithLeader(government int, bonusTrait bool, leaderBonus int) (int, bool) {
	percent, ok := OriginalTradeAgreementGoalPercent(government, bonusTrait)
	if !ok {
		return 0, false
	}
	if leaderBonus > 0 {
		percent += leaderBonus
	}
	return percent, true
}
