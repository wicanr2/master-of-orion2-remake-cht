package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// 領袖每回合真的要付錢——這是這一項修的那個洞。
//
// 回歸保護：領袖維護費已接進 EndTurn 的單次帝國經濟結算；這裡另測純計費 helper。
func TestLeadersCostMoneyEveryTurn(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	// 新開一局**沒有領袖**(第 14 項(地表道路)改成原版的雇用制,不再自帶),所以要自己放幾位。
	s.Leaders = demoLeaders()
	cost := s.LeaderUpkeepTotal()
	if cost <= 0 {
		t.Fatalf("有領袖就該有維護費,得到 %d", cost)
	}
}

// 正對照:一個領袖都沒有就不扣錢。
//
// 少了這條,「每回合固定扣一筆」的實作也會讓上面那條通過。
func TestNoLeadersNoUpkeep(t *testing.T) {
	s := NewDemoSession()
	s.Leaders = nil
	if got := s.LeaderUpkeepTotal(); got != 0 {
		t.Errorf("沒有領袖不該有維護費,得到 %d", got)
	}
}

// 有 Megawealth 的領袖免維護費(openorion2 `leaderMaintenanceCost` 的第一個分支)。
func TestMegawealthLeaderIsFree(t *testing.T) {
	rich := Leader{Name: "測試富豪", Skill: "巨富", Level: 3, Ship: false, Tier: 1,
		Skills: []LeaderSkill{{ID: int(gamedata.SKILL_MEGAWEALTH), Tier: 1}}}
	if !leaderHasMegawealth(rich) {
		t.Fatal("應判定為有 Megawealth")
	}
	if got := leaderUpkeepCost(rich); got != 0 {
		t.Errorf("有 Megawealth 應免維護費,得到 %d", got)
	}

	// 反面:同等級但技能不同的領袖要付錢——否則上面那個 0 只是「所有人都免費」。
	poor := Leader{Name: "測試軍官", Skill: "指揮官", Level: 3, Ship: true, Tier: 1,
		Skills: []LeaderSkill{{ID: int(gamedata.SKILL_ORDNANCE), Tier: 1}}}
	if got := leaderUpkeepCost(poor); got <= 0 {
		t.Errorf("沒有 Megawealth 的領袖應付費,得到 %d", got)
	}
}

// 維護費與雇用費走**同一條公式**:等級越高、雇用越貴的,維護也越貴。
//
// 兩邊用不同基準的話,同一位領袖會出現「雇用時算貴、維護時算便宜」這種對不起來的狀況。
func TestUpkeepTracksHireCost(t *testing.T) {
	s := NewDemoSession()
	low := Leader{Name: "低階", Skill: "指揮官", Level: 1, Ship: true, Tier: 1}
	high := Leader{Name: "高階", Skill: "指揮官", Level: 5, Ship: true, Tier: 1}
	if s.MercHireCost(high) <= s.MercHireCost(low) {
		t.Fatalf("測試前提不成立:高階雇用費應較高(%d vs %d)", s.MercHireCost(high), s.MercHireCost(low))
	}
	if leaderUpkeepCost(high) < leaderUpkeepCost(low) {
		t.Errorf("雇用費較高者維護費不該較低:%d vs %d", leaderUpkeepCost(high), leaderUpkeepCost(low))
	}
}

// 端到端:回合結算真的有呼叫到(先前那個洞正是「函式寫好了但沒有呼叫端」)。
func TestEndTurnChargesLeaderUpkeep(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	// 新開一局**沒有領袖**(第 14 項(地表道路)改成原版的雇用制,不再自帶),所以要自己放幾位。
	s.Leaders = demoLeaders()
	s.Player.BC = 100000 // 拉高,免得被其他收支淹沒
	s.EndTurn()
	if s.LastLeaderUpkeep <= 0 {
		t.Errorf("回合結算應扣領袖維護費,LastLeaderUpkeep=%d", s.LastLeaderUpkeep)
	}
	if s.LastLeaderUpkeep != s.LeaderUpkeepTotal() {
		t.Errorf("扣的金額應等於總額 %d,得到 %d", s.LeaderUpkeepTotal(), s.LastLeaderUpkeep)
	}
}
