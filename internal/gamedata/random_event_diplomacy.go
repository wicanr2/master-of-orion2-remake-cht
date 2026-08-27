package gamedata

// OriginalDiplomaticIncidentRelation 對應事件 4／5 呼叫 Change_Relations_ @ 0x4E3B5
// 的玩家可見結果。current 使用原版 -100..100 尺度。Determine_Event_
// 先要求受害帝國對第二帝國的正式狀態為 4／5；但 Change_Relations_
// 在 0x4E75C..0x4E764 讀取反向狀態，該值 >=4 就直接返回，不寫回關係分數。
// remake 目前的正式條約是對稱的，因此事件成立時必然保留現值。
func OriginalDiplomaticIncidentRelation(current, eventID int, policy ForeignPolicy) (int, bool) {
	if current < -100 || current > 100 || policy < DIPLO_LIMITED_WAR {
		return current, false
	}
	switch eventID {
	case 4, 5:
		return current, true
	default:
		return current, false
	}
}
