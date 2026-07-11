package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
)

// buildDimensionalPortal 讓母星(PlayerColonies[0])記為已建成次元傳送門,不經過建造佇列
// (供測試直接構造前置條件已滿足的 GameSession,比照其餘測試檔手法)。
func buildDimensionalPortal(s *GameSession) {
	if s.ColonyBuildings == nil {
		s.ColonyBuildings = make([]map[string]bool, len(s.PlayerColonies))
	}
	if s.ColonyBuildings[0] == nil {
		s.ColonyBuildings[0] = make(map[string]bool)
	}
	s.ColonyBuildings[0][dimensionalPortalBuildingName] = true
}

// strongDoomStarFleet 回傳一支戰力遠高於 antaranHomeFleetDefense 保守預設的末日之星艦隊
// (滿裝 WeaponOptions/ArmorOptions/ShieldOptions 表最高階元件),供「玩家應該打贏」的測試
// 案例使用。
func strongDoomStarFleet(n int) []Ship {
	out := make([]Ship, n)
	for i := range out {
		out[i] = Ship{
			Name: "反攻艦隊", Class: "末日之星",
			Weapon: "電漿砲", WeaponAttack: 200,
			Armor: "精金裝甲", Shield: "第十級護盾",
		}
	}
	return out
}

func TestCanAssaultAntaresRequiresPortalAndFleet(t *testing.T) {
	s := NewDemoSession()
	if s.CanAssaultAntares() {
		t.Fatalf("尚未建次元傳送門,不應允許反攻")
	}

	buildDimensionalPortal(s)
	if !s.CanAssaultAntares() {
		t.Fatalf("已建次元傳送門且艦隊非空,應允許反攻")
	}

	s.Ships = nil
	if s.CanAssaultAntares() {
		t.Fatalf("艦隊為空,不應允許反攻(手冊:select a fleet)")
	}
}

// TestAssaultAntaresBlockedWithoutPortal 驗證沒有次元傳送門時 AssaultAntares 直接擋下,
// 不消耗艦隊、不觸發戰鬥、不誤判勝利。
func TestAssaultAntaresBlockedWithoutPortal(t *testing.T) {
	s := NewDemoSession()
	shipsBefore := len(s.Ships)

	res, ok := s.AssaultAntares()

	if ok {
		t.Fatalf("尚未建次元傳送門,AssaultAntares 應回傳 ok=false")
	}
	if res.Enemy != "" || res.PlayerWon || res.PlayerStart != 0 || res.EnemyStart != 0 || res.Log != nil {
		t.Fatalf("前置條件不滿足時應回傳零值 BattleResult,got %+v", res)
	}
	if len(s.Ships) != shipsBefore {
		t.Fatalf("前置條件不滿足不應消耗艦隊:before=%d after=%d", shipsBefore, len(s.Ships))
	}
	if s.Victory.Over {
		t.Fatalf("不應誤判勝利")
	}
	if s.AntaranHomeworldConquered {
		t.Fatalf("不應誤設 AntaranHomeworldConquered")
	}
}

// TestAssaultAntaresBlockedWithEmptyFleet 驗證已建傳送門但艦隊為空時仍被擋下。
func TestAssaultAntaresBlockedWithEmptyFleet(t *testing.T) {
	s := NewDemoSession()
	buildDimensionalPortal(s)
	s.Ships = nil

	_, ok := s.AssaultAntares()

	if ok {
		t.Fatalf("艦隊為空,AssaultAntares 應回傳 ok=false")
	}
	if s.Victory.Over {
		t.Fatalf("不應誤判勝利")
	}
}

// TestAssaultAntaresBlockedWhenEventsDisabled 驗證手冊「This strategy is not available if you
// disabled Antaran Attacks when setting up your game」——DisableEvents=true 時一併關閉反攻路徑。
func TestAssaultAntaresBlockedWhenEventsDisabled(t *testing.T) {
	s := NewDemoSession()
	buildDimensionalPortal(s)
	s.DisableEvents = true

	_, ok := s.AssaultAntares()

	if ok {
		t.Fatalf("DisableEvents=true(關閉安塔蘭攻擊)時不應允許反攻")
	}
}

// TestAssaultAntaresWeakFleetLoses 驗證艦隊太弱(遠不如 antaranHomeFleetDefense 保守預設)時
// 戰敗,不誤判勝利,且套用了艦隊損失。
func TestAssaultAntaresWeakFleetLoses(t *testing.T) {
	s := NewDemoSession()
	buildDimensionalPortal(s)
	s.Ships = []Ship{{Name: "偵察艦1", Class: "偵察艦"}} // 戰力遠不如 6 艘末日之星等級的防禦艦隊

	res, ok := s.AssaultAntares()

	if !ok {
		t.Fatalf("前置條件已滿足,AssaultAntares 應回傳 ok=true(結果為戰敗,非前置擋下)")
	}
	if res.PlayerWon {
		t.Fatalf("單艘偵察艦不應能擊敗安塔蘭母星防禦艦隊")
	}
	if s.Victory.Over {
		t.Fatalf("戰敗不應誤判勝利")
	}
	if s.AntaranHomeworldConquered {
		t.Fatalf("戰敗不應設定 AntaranHomeworldConquered")
	}
	if s.LastBattle == nil || s.LastBattle.PlayerWon {
		t.Fatalf("LastBattle 應記錄本次戰敗結果,got %+v", s.LastBattle)
	}
}

// TestAssaultAntaresStrongFleetWinsAndVictoryDetected 是本輪的核心驗證:已建次元傳送門 + 強
// 艦隊反攻成功 → AntaranHomeworldConquered=true;下一次 advanceAntaranVictory(EndTurn 呼叫)
// 偵測到並設定 s.Victory.Over=true、Reason=engine.VictoryAntaran、Winner="player"。
func TestAssaultAntaresStrongFleetWinsAndVictoryDetected(t *testing.T) {
	s := NewDemoSession()
	buildDimensionalPortal(s)
	s.Ships = strongDoomStarFleet(8) // 8 艘末日之星,遠強於 antaranHomeFleetDefense(6 艘同級)

	res, ok := s.AssaultAntares()

	if !ok {
		t.Fatalf("前置條件已滿足,AssaultAntares 應回傳 ok=true")
	}
	if !res.PlayerWon {
		t.Fatalf("8 艘末日之星艦隊應能擊敗 antaranHomeFleetDefense(6 艘同級),log=%v", res.Log)
	}
	if !s.AntaranHomeworldConquered {
		t.Fatalf("戰勝後應設定 AntaranHomeworldConquered=true")
	}
	if s.Victory.Over {
		t.Fatalf("AssaultAntares 本身不應直接設定 s.Victory——那是 advanceAntaranVictory(EndTurn)的職責")
	}

	s.advanceAntaranVictory()

	if !s.Victory.Over {
		t.Fatalf("advanceAntaranVictory 應偵測 AntaranHomeworldConquered 並結束遊戲")
	}
	if s.Victory.Reason != engine.VictoryAntaran {
		t.Fatalf("勝利路徑應為 engine.VictoryAntaran,got %v", s.Victory.Reason)
	}
	if s.Victory.Winner != "player" {
		t.Fatalf("安塔蘭勝利的勝者應為 player,got %q", s.Victory.Winner)
	}
}

// TestAdvanceAntaranVictoryNoOpWithoutConquest 驗證未攻陷母星時 advanceAntaranVictory 不誤判。
func TestAdvanceAntaranVictoryNoOpWithoutConquest(t *testing.T) {
	s := NewDemoSession()
	s.advanceAntaranVictory()
	if s.Victory.Over {
		t.Fatalf("尚未攻陷安塔蘭母星,不應判定勝利")
	}
}

// TestEndTurnVictoryPriorityExterminationBeatsAntaran 驗證 EndTurn 的呼叫順序(見 session.go
// EndTurn:advanceConquestVictory → advanceAntaranVictory → advanceCouncil)與
// engine.CheckVictory 文件記載的「滅絕 → 安塔蘭 → 議會」優先序一致:兩個條件同時成立時,
// 殲滅勝利先判定並鎖定 Victory.Over,安塔蘭偵測不會覆蓋它。
func TestEndTurnVictoryPriorityExterminationBeatsAntaran(t *testing.T) {
	s := NewDemoSession()
	for i := range s.AIPlayers {
		s.AIPlayers[i].Colonies = nil // 觸發殲滅勝利
	}
	s.AntaranHomeworldConquered = true // 同時也滿足安塔蘭勝利條件

	s.advanceConquestVictory()
	s.advanceAntaranVictory()

	if s.Victory.Reason != engine.VictoryExtermination {
		t.Fatalf("兩條件同時成立時應優先判定殲滅勝利,got %v", s.Victory.Reason)
	}
}
