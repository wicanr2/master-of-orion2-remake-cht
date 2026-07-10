package engine

// leader.go 是軍官(Leader)雇用的最小引擎機制——只做「BC 夠不夠、扣款」這件事,不碰任何
// UI/畫面(招募畫面留待後續,見 gamedata/officer.go 檔頭與 docs/tech/leader-officer-skills.md)。
// hireCost 一律由呼叫端以 gamedata.LeaderHireCost(skillValue, expLevel, modifier) 算好傳入,
// modifier 則以 gamedata.LeaderHireModifier(該玩家已受雇領袖的 Famous 加成) 算好傳入——
// 本函式本身不重算任何技能公式,只做金流判斷,對照 openorion2
// LeaderListView::selectSlot(officer.cpp:866-889)「money < hire_cost 就拒絕」的判斷邏輯
// (實際扣款與領袖狀態轉換在原碼是 STUB「Not implemented yet」,remake 這裡把「扣款」這一半
// 做成可用的純函式,領袖狀態轉換仍留給呼叫端/未來 UI)。

// HireLeader 嘗試以 cost 雇用一名軍官:currentBC 足夠才扣款成功。cost<0 視為 0(對照
// gamedata.LeaderHireCost 已保證下限 0,這裡對外部傳入值仍做防禦)。
func HireLeader(currentBC int, cost int) (newBC int, ok bool) {
	if cost < 0 {
		cost = 0
	}
	if currentBC < cost {
		return currentBC, false
	}
	return currentBC - cost, true
}
