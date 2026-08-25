package gamedata

// OriginalDiplomaticIncidentRelation 對應事件 4／5 呼叫 Change_Relations_ @ 0x4E3B5
// 時可達的 raw signed-byte 關係鏈。current 使用原版 -100..100 尺度；事件 4／5
// 的基礎增量分別是 -100／+100。這兩項事件只在正式狀態 4／5 時成立。
func OriginalDiplomaticIncidentRelation(current, eventID, actorGovernment int,
	targetCharismatic bool, policy ForeignPolicy) (int, bool) {
	if current < -100 || current > 100 || policy < DIPLO_LIMITED_WAR {
		return current, false
	}
	delta := 0
	switch eventID {
	case 4:
		delta = -100
	case 5:
		delta = 100
	default:
		return current, false
	}

	if delta > 0 {
		if current < 0 {
			delta *= 2
			if current+delta > 10 {
				delta = 10 - current
			}
		} else {
			delta /= current/25 + 1
		}
	} else if current > 0 {
		delta = delta*2 - current
	} else {
		delta /= current/-25 + 1
	}

	// 原版 actor government raw enum 與 MoraleGovernmentType 同為 0..7。
	if actorGovernment == 4 && delta < 0 {
		delta *= 2
	} else if actorGovernment == 0 {
		if delta < 0 {
			delta = delta * 3 / 2
		} else if delta > 0 {
			delta = delta * 3 / 4
		}
	}
	if targetCharismatic {
		if delta > 0 {
			delta *= 2
		} else if delta < 0 {
			delta /= 2
		}
	}

	result := current + delta
	if result < -100 {
		result = -100
	} else if result > 100 {
		result = 100
	}
	// Change_Relations_ 對 raw 正式狀態 4／5 的最後約束。
	if result > -25 {
		result = -25
	}
	return result, true
}
