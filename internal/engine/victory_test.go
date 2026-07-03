package engine

import "testing"

func TestCheckExterminationOneSurvivor(t *testing.T) {
	// 4 個玩家,只剩 index 2 存活 → 該玩家滅絕勝利。
	alive := []bool{false, false, true, false}
	ok, winner := CheckExtermination(alive)
	if !ok {
		t.Fatal("只剩一名存活玩家應判定滅絕勝利")
	}
	if winner != 2 {
		t.Errorf("勝者 index = %d,預期 2", winner)
	}
}

func TestCheckExterminationMultipleSurvivors(t *testing.T) {
	// 仍有 2 個以上玩家存活 → 尚未滅絕勝利。
	alive := []bool{true, false, true}
	ok, winner := CheckExtermination(alive)
	if ok {
		t.Error("仍有多名存活玩家不應判定滅絕勝利")
	}
	if winner != -1 {
		t.Errorf("未達成時 winner = %d,預期 -1", winner)
	}
}

func TestCheckExterminationNoSurvivors(t *testing.T) {
	// 全滅(例如互相同歸於盡)→ 沒有勝者,也不算滅絕勝利。
	alive := []bool{false, false, false}
	ok, winner := CheckExtermination(alive)
	if ok {
		t.Error("全滅不應判定滅絕勝利(無存活玩家可稱王)")
	}
	if winner != -1 {
		t.Errorf("winner = %d,預期 -1", winner)
	}
}

func TestCheckExterminationEmpty(t *testing.T) {
	ok, winner := CheckExtermination(nil)
	if ok || winner != -1 {
		t.Errorf("空輸入應為 (false, -1),實際 (%v, %d)", ok, winner)
	}
}

func TestCheckHighCouncilExactlyTwoThirds(t *testing.T) {
	// 總票 30,得票 20 → 20*3=60 == 30*2=60,剛好達 2/3 門檻,應通過。
	if !CheckHighCouncil(20, 30) {
		t.Error("剛好 2/3 得票應判定通過(20/30)")
	}
	// 總票 3,得票 2 → 2*3=6 == 3*2=6,同樣剛好 2/3。
	if !CheckHighCouncil(2, 3) {
		t.Error("剛好 2/3 得票應判定通過(2/3)")
	}
}

func TestCheckHighCouncilOneVoteShort(t *testing.T) {
	// 總票 30,得票 19 → 19*3=57 < 30*2=60,差一點點未達門檻。
	if CheckHighCouncil(19, 30) {
		t.Error("19/30 未達 2/3 門檻,不應通過")
	}
	// 總票 3,得票 1 → 1*3=3 < 3*2=6。
	if CheckHighCouncil(1, 3) {
		t.Error("1/3 未達 2/3 門檻,不應通過")
	}
}

func TestCheckHighCouncilInvalidTotal(t *testing.T) {
	if CheckHighCouncil(5, 0) {
		t.Error("totalVotes<=0 是無效輸入,不應判定通過")
	}
	if CheckHighCouncil(0, -1) {
		t.Error("totalVotes 為負是無效輸入,不應判定通過")
	}
}

func TestCheckAntaranVictory(t *testing.T) {
	if CheckAntaranVictory(false) {
		t.Error("母星未攻陷不應判定安塔蘭勝利")
	}
	if !CheckAntaranVictory(true) {
		t.Error("母星已攻陷應判定安塔蘭勝利")
	}
}

func TestCheckVictoryExtermination(t *testing.T) {
	// 滅絕優先於其他條件同時成立的情況也應正確回報滅絕。
	alive := []bool{true, false, false}
	cond, winner := CheckVictory(alive, 1, true, 2, 30, 30)
	if cond != VictoryExtermination || winner != 0 {
		t.Errorf("CheckVictory = (%v, %d),預期 (VictoryExtermination, 0)", cond, winner)
	}
}

func TestCheckVictoryAntaran(t *testing.T) {
	alive := []bool{true, true, true} // 多方存活,滅絕條件不成立
	cond, winner := CheckVictory(alive, 1, true, -1, 0, 30)
	if cond != VictoryAntaran || winner != 1 {
		t.Errorf("CheckVictory = (%v, %d),預期 (VictoryAntaran, 1)", cond, winner)
	}
}

func TestCheckVictoryHighCouncil(t *testing.T) {
	alive := []bool{true, true, true}
	cond, winner := CheckVictory(alive, -1, false, 2, 20, 30) // 20/30 = 2/3
	if cond != VictoryHighCouncil || winner != 2 {
		t.Errorf("CheckVictory = (%v, %d),預期 (VictoryHighCouncil, 2)", cond, winner)
	}
}

func TestCheckVictoryNone(t *testing.T) {
	alive := []bool{true, true, true}
	cond, winner := CheckVictory(alive, -1, false, 2, 19, 30) // 未達 2/3
	if cond != VictoryNone || winner != -1 {
		t.Errorf("CheckVictory = (%v, %d),預期 (VictoryNone, -1)", cond, winner)
	}
}
