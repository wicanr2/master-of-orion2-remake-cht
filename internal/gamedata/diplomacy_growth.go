package gamedata

// originalBaseRelationLowBytes 是 Orion2.exe 的 byte_180ED4（IDA linear EA
// 0x180ED4）14×14 word 表之 signed low byte。每列是觀察者種族、每欄是
// 目標種族；high byte 全為符號延伸。原始 392 bytes SHA-256：
// 05e582491a6173319e8d57d0751d24dd5629f73f00dc3d7c2a5b979e39efa831。
var originalBaseRelationLowBytes = [14][14]int8{
	{0, 0, -24, 24, 0, 0, -12, 0, -24, 0, 12, 0, 24, 0},
	{0, 0, -12, 24, 0, 0, 0, -12, 0, 0, 0, 0, 0, 0},
	{-24, -12, 0, 0, 12, 0, 0, 12, 0, 0, 0, 12, 0, 0},
	{24, 24, 0, 0, -24, 0, 0, 0, 24, -12, 0, -12, 0, 0},
	{0, 0, 12, -24, 0, 0, 0, 12, -12, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{-12, 0, 0, 0, 0, 0, 0, 0, -12, 0, -12, 12, 24, 0},
	{0, 0, 12, 0, 12, 0, 0, 0, 12, 0, 0, -12, -12, 0},
	{-24, 0, 0, 24, -12, 0, -12, 12, 0, 0, 0, 0, 12, 0},
	{0, 0, 0, -12, 0, 0, 0, 0, 0, 0, 0, 0, 12, 0},
	{12, 0, 0, 0, 0, 0, -12, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 12, -12, 0, 0, 12, -12, 0, 0, 0, 0, 0, 0},
	{24, 0, 0, 0, 0, 0, 24, -12, 12, 12, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
}

// OriginalBaseRelation 回傳 sub_4D78E 初始化 player+0x61F 時使用的原版
// 種族配對目標。自訂種族或非法原版索引不可猜值，回傳 ok=false。
func OriginalBaseRelation(observerRace, targetRace int) (value int, ok bool) {
	if observerRace < 0 || observerRace >= len(originalBaseRelationLowBytes) ||
		targetRace < 0 || targetRace >= len(originalBaseRelationLowBytes[0]) {
		return 0, false
	}
	return int(originalBaseRelationLowBytes[observerRace][targetRace]), true
}

// OriginalRelationChangeInput 是 Change_Relations_ @ 0x4E3B5 中關係分數
// 寫回公式的 typed 輸入。本函式只表示 policy<4、無特殊 grievance
// reason 的可達切片；快取／抱怨 record 另由後續規格處理。
type OriginalRelationChangeInput struct {
	CurrentRaw        int
	BaseDelta         int
	ActorGovernment   int
	TargetCharismatic bool
	Policy            ForeignPolicy
	BothAI            bool
	RelativeTurn      int
	Difficulty        int
}

// OriginalChangeRelationScore 套用 `Change_Relations_ @ 0x4E3B5` 的已證實
// raw signed-byte 分數公式。成功時回傳 -100..100；ok=false 表示輸入
// 不在本 typed 切片可表示的範圍，不得猜測補值。
func OriginalChangeRelationScore(in OriginalRelationChangeInput) (int, bool) {
	if in.CurrentRaw < -100 || in.CurrentRaw > 100 || in.BaseDelta == 0 ||
		in.Policy < DIPLO_NONE || in.Policy >= DIPLO_LIMITED_WAR {
		return in.CurrentRaw, false
	}
	delta := in.BaseDelta
	if delta > 0 {
		if in.CurrentRaw < 0 {
			delta *= 2
			if in.CurrentRaw+delta > 10 {
				delta = 10 - in.CurrentRaw
			}
		} else {
			delta /= in.CurrentRaw/25 + 1
		}
	} else if in.CurrentRaw > 0 {
		delta *= 2
	} else {
		delta /= in.CurrentRaw/-25 + 1
	}

	// actor government raw enum 來自 player+0x27。
	if in.ActorGovernment == 4 && delta < 0 {
		delta *= 2
	} else if in.ActorGovernment == 0 {
		if delta < 0 {
			delta = delta * 3 / 2
		} else if delta > 0 {
			delta = delta * 3 / 4
		}
	}
	if in.TargetCharismatic {
		if delta > 0 {
			delta *= 2
		} else {
			delta /= 2
		}
	}
	if in.BothAI && in.RelativeTurn > 100 {
		if delta > 0 {
			delta *= 2
		} else if delta < 0 {
			divisor := in.Difficulty/2 + 1
			if divisor < 1 {
				divisor = 1
			}
			delta /= divisor
		}
	}

	result := in.CurrentRaw + delta
	if result < -100 {
		result = -100
	} else if result > 100 {
		result = 100
	}
	if in.Policy != DIPLO_ALLIANCE && result > 65 {
		result = 65
	}
	return result, true
}

// OriginalWarBlockadeGrievance 對映 Change_Relations_ reason raw 7、policy 4/5、
// target personality 100 的 +0x6BF 特例。它回傳寫入積怨表的 delta/4（負數向下取整），
// 不改一般 +0x617 關係分數。
func OriginalWarBlockadeGrievance(in OriginalRelationChangeInput) (int, bool) {
	if in.CurrentRaw < -100 || in.CurrentRaw > 100 || in.BaseDelta >= 0 || in.BaseDelta < -5 ||
		(in.Policy != DIPLO_LIMITED_WAR && in.Policy != DIPLO_WAR) {
		return 0, false
	}
	delta := in.BaseDelta
	if in.CurrentRaw > 0 {
		delta *= 2
	} else {
		delta /= in.CurrentRaw/-25 + 1
	}
	if in.ActorGovernment == 4 {
		delta *= 2
	} else if in.ActorGovernment == 0 {
		delta = delta * 3 / 2
	}
	if in.TargetCharismatic {
		delta /= 2
	}
	if delta >= 0 {
		return 0, false
	}
	quarter := delta / 4
	if delta%4 != 0 {
		quarter--
	}
	return quarter, true
}

// OriginalDiplomacyGrowthTreatyInput 是 Diplomacy_Growth_ @ 0x4DD6B 首個
// pair loop 中已閉合的條約欄位。TributeMode 是 raw +0x63F 的 0/1/2。
type OriginalDiplomacyGrowthTreatyInput struct {
	CurrentRaw        int
	FormalPolicy      ForeignPolicy
	TradeActive       bool
	ResearchActive    bool
	TributeMode       int
	ActorGovernment   int
	TargetCharismatic bool
}

// OriginalDiplomacyGrowthTreatyRelation 依 0x4DE07..0x4DFB7 的原始順序
// 消費條約與 Random_(n)。roll 必須回傳 1..n；非法回傳失敗即關閉。
func OriginalDiplomacyGrowthTreatyRelation(in OriginalDiplomacyGrowthTreatyInput,
	roll func(int) int) (int, bool) {
	if in.CurrentRaw < -100 || in.CurrentRaw > 100 || roll == nil ||
		in.FormalPolicy < DIPLO_NONE || in.FormalPolicy > DIPLO_TOTAL_WAR ||
		in.TributeMode < 0 || in.TributeMode > 2 {
		return in.CurrentRaw, false
	}
	current := in.CurrentRaw
	checkedRoll := func(n int) (int, bool) {
		value := roll(n)
		return value, value >= 1 && value <= n
	}
	apply := func(base int) bool {
		next, ok := OriginalChangeRelationScore(OriginalRelationChangeInput{
			CurrentRaw: current, BaseDelta: base, ActorGovernment: in.ActorGovernment,
			TargetCharismatic: in.TargetCharismatic, Policy: in.FormalPolicy,
		})
		if ok {
			current = next
		}
		return ok
	}
	percentGate := func(maxDelta int) bool {
		percent, ok := checkedRoll(100)
		if !ok {
			return false
		}
		if percent > 100-current {
			return true
		}
		delta, ok := checkedRoll(maxDelta)
		return ok && apply(delta)
	}

	if in.FormalPolicy == DIPLO_NON_AGGRESSION && !percentGate(3) {
		return in.CurrentRaw, false
	}
	if in.TradeActive && !percentGate(3) {
		return in.CurrentRaw, false
	}
	if in.ResearchActive && !percentGate(3) {
		return in.CurrentRaw, false
	}
	if in.FormalPolicy == DIPLO_ALLIANCE {
		delta, ok := checkedRoll(5)
		if !ok || !apply(delta) {
			return in.CurrentRaw, false
		}
	}
	if in.TributeMode == 1 && !percentGate(3) {
		return in.CurrentRaw, false
	}
	if in.TributeMode == 2 && !percentGate(8) {
		return in.CurrentRaw, false
	}
	return current, true
}

// OriginalDiplomacyRelationDriftInput 對應 Diplomacy_Growth_ 的
// 0x4E11D..0x4E276。Locked 是 observer+2*target+0x737 的非零鎖定字。
type OriginalDiplomacyRelationDriftInput struct {
	CurrentRaw int
	TargetRaw  int
	Policy     ForeignPolicy
	Locked     bool
}

// OriginalDiplomacyRelationDrift 依原版擲骰順序，讓 current 以 0 或 1 向
// target 靠近；policy>=4 另把雙方關係壓到最多 -90。roll 必須回傳 1..n。
func OriginalDiplomacyRelationDrift(in OriginalDiplomacyRelationDriftInput,
	roll func(int) int) (int, bool) {
	if in.CurrentRaw < -100 || in.CurrentRaw > 100 || in.TargetRaw < -100 ||
		in.TargetRaw > 100 || in.Policy < DIPLO_NONE || in.Policy > DIPLO_TOTAL_WAR || roll == nil {
		return in.CurrentRaw, false
	}
	checked := func(n int) (int, bool) {
		v := roll(n)
		return v, v >= 1 && v <= n
	}
	draw, ok := checked(105)
	if !ok {
		return in.CurrentRaw, false
	}
	step := 0
	absCurrent := in.CurrentRaw
	if absCurrent < 0 {
		absCurrent = -absCurrent
	}
	if draw > absCurrent {
		quarter, valid := checked(4)
		if !valid {
			return in.CurrentRaw, false
		}
		if quarter == 1 {
			coin, valid := checked(2)
			if !valid {
				return in.CurrentRaw, false
			}
			step = coin - 1
		}
	}
	current := in.CurrentRaw
	if !in.Locked && step != 0 {
		if in.TargetRaw < current {
			current--
			if current < in.TargetRaw {
				current = in.TargetRaw
			}
		} else if in.TargetRaw > current {
			current++
			if current > in.TargetRaw {
				current = in.TargetRaw
			}
		}
	}
	if in.Policy >= DIPLO_LIMITED_WAR && current > -90 {
		current = -90
	}
	return current, true
}
