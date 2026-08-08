package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// 領袖每回合真的要付錢——這是這一項修的那個洞。
//
// `gamedata.LeaderMaintenanceCost` 寫好了、有測試,但**零個生產端呼叫**:
// 一局跑 300 回合量覆蓋率是 0.0%。所以 remake 的領袖雇用一次之後永久免費。
func TestLeadersCostMoneyEveryTurn(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	// 新開一局**沒有領袖**(第 15 項改成原版的雇用制,不再自帶),所以要自己放幾位。
	s.Leaders = demoLeaders()
	cost := s.LeaderUpkeepTotal()
	if cost <= 0 {
		t.Fatalf("有領袖就該有維護費,得到 %d", cost)
	}

	s.Player.BC = 10000
	before := s.Player.BC
	paid := s.advanceLeaderUpkeep()

	if paid != cost {
		t.Errorf("扣款額應為 %d,得到 %d", cost, paid)
	}
	if s.Player.BC != before-cost {
		t.Errorf("國庫應從 %d 扣到 %d,得到 %d", before, before-cost, s.Player.BC)
	}
}

// 正對照:一個領袖都沒有就不扣錢。
//
// 少了這條,「每回合固定扣一筆」的實作也會讓上面那條通過。
func TestNoLeadersNoUpkeep(t *testing.T) {
	s := NewDemoSession()
	s.Leaders = nil
	s.Player.BC = 500
	if got := s.advanceLeaderUpkeep(); got != 0 {
		t.Errorf("沒有領袖不該扣錢,扣了 %d", got)
	}
	if s.Player.BC != 500 {
		t.Errorf("國庫不該變動,得到 %d", s.Player.BC)
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

// 付不出來時只扣到 0,不會把國庫扣成負的、也不會反向加錢。
//
// ⚠ 反向加錢不是假想的:session.go 對戰損 bcLoss 有一段註解記著同一個坑
// ——「BC 為負時若只判斷 bcLoss > BC 會把損失夾成負值,`BC -= bcLoss` 反而變成加錢」。
func TestUpkeepNeverGoesNegativeOrPaysBack(t *testing.T) {
	s := NewDemoSession()
	// 新開一局**沒有領袖**(第 15 項改成原版的雇用制,不再自帶),所以要自己放幾位。
	s.Leaders = demoLeaders()
	s.Player.BC = 1
	s.advanceLeaderUpkeep()
	if s.Player.BC < 0 {
		t.Errorf("國庫不該被扣成負的,得到 %d", s.Player.BC)
	}

	s.Player.BC = -50
	before := s.Player.BC
	if got := s.advanceLeaderUpkeep(); got != 0 {
		t.Errorf("國庫已經是負的就不該再扣,扣了 %d", got)
	}
	if s.Player.BC != before {
		t.Errorf("國庫不該變動(尤其不該變多):%d → %d", before, s.Player.BC)
	}
}

// 端到端:回合結算真的有呼叫到(先前那個洞正是「函式寫好了但沒有呼叫端」)。
func TestEndTurnChargesLeaderUpkeep(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	// 新開一局**沒有領袖**(第 15 項改成原版的雇用制,不再自帶),所以要自己放幾位。
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
