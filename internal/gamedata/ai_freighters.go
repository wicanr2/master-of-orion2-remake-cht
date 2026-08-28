package gamedata

// OriginalAIFreighterFleetGain 對映 sub_D6AD4 @ 0xD6D4D..0xD6D76。
// 只有本回合曾進入 +0x38<=0 的運輸壓力分支，且 Random(10)<=difficulty 時增加五艘。
func OriginalAIFreighterFleetGain(pressure bool, difficulty, roll10 int) (int, bool) {
	if difficulty < 0 || difficulty > 4 || roll10 < 1 || roll10 > 10 {
		return 0, false
	}
	if pressure && roll10 <= difficulty {
		return FreighterFleetShipsPerBuild, true
	}
	return 0, true
}
