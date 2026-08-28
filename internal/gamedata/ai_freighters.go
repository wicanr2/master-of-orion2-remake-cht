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

// OriginalAIFreighterFleetBuildQuota 對映 sub_CFCB6 → sub_CF3BD：raw type 1 且
// status 1／2 的航行殖民船數，扣除 player+0x38 的貨運艦餘額後，每不足 5 艘形成
// 一個 -15 貨運艦隊產品配額。原版使用 (4-diff)/5；此處寫成等價的正數無條件進位。
func OriginalAIFreighterFleetBuildQuota(surplusFreighters, movingColonyShips int) (int, bool) {
	if movingColonyShips < 0 {
		return 0, false
	}
	deficit := movingColonyShips - surplusFreighters
	if deficit <= 0 {
		return 0, true
	}
	return (deficit + FreighterFleetShipsPerBuild - 1) / FreighterFleetShipsPerBuild, true
}
