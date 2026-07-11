package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// buildFreighterFleetOneAtATime 逐回合推進殖民地 0 的「運輸艦隊」建造,直到完工(或超過
// maxTurns 次仍未完工則 Fatal)。回傳完工那一回合 EndTurn 前後的 BC 差值(completedDeltaBC)、
// 該回合 RunEmpireTurn 本身算出的 NetBC(completedNetBC,不含完工當下的現金加成)。
// 之所以要逐回合跑而非一次跑到底,是因為每回合稅收/維護費本身就會動 BC(NetBC 非 0),若只比較
// 「建造前」與「完工後」的 BC 總差,會把「這幾回合正常經濟收支」與「完工當下的現金加成」混在一起,
// 驗證不出現金加成本身的金額——見 applySpecialAction 對「運輸艦隊」的實作註解(先算好本回合
// NetBC 併入 s.Player,advanceBuilds 才在後面加現金加成,兩者可用「這回合 NetBC vs BC 實際變化」
// 的差額分離出來)。
func buildFreighterFleetOneAtATime(t *testing.T, s *GameSession, maxTurns int) (completedDeltaBC, completedNetBC int) {
	t.Helper()
	action, ok := gamedata.SpecialActionByNameZH(gamedata.FreighterFleetActionName)
	if !ok {
		t.Fatal("找不到運輸艦隊的 SpecialAction 資料")
	}
	s.Builds[0] = ColonyBuild{Name: gamedata.FreighterFleetActionName, Progress: 0, Cost: action.ProductionCost}

	for i := 0; i < maxTurns; i++ {
		beforeBC := s.Player.BC
		s.EndTurn()
		if s.Builds[0].Name == "" { // 本回合完工(advanceBuilds 已清空 Builds[0])
			return s.Player.BC - beforeBC, s.LastPlayerOutput.NetBC
		}
	}
	t.Fatalf("運輸艦隊在 %d 回合內未完工(ProductionCost=%d)", maxTurns, action.ProductionCost)
	return 0, 0
}

// TestFreighterFleetBuildIncreasesActiveFreightersAndCash 驗證「運輸艦隊」完工後:
// ①ActiveFreighters += 5(手冊 p.168:每次建造 +5 艘)②完工當下 BC 的變化 = 該回合正常 NetBC +
// RuleProfile.FreightersCashBonus(本測試用 Profile13,固定回饋 5 BC)。
func TestFreighterFleetBuildIncreasesActiveFreightersAndCash(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	s.RuleProfile = gamedata.Profile13()
	if s.Player.CompletedTopics == nil {
		s.Player.CompletedTopics = map[gamedata.ResearchTopic]bool{}
	}
	s.Player.CompletedTopics[gamedata.TOPIC_NUCLEAR_FISSION] = true

	startFreighters := s.Player.ActiveFreighters

	deltaBC, netBC := buildFreighterFleetOneAtATime(t, s, 500)

	if got, want := s.Player.ActiveFreighters, startFreighters+gamedata.FreighterFleetShipsPerBuild; got != want {
		t.Errorf("ActiveFreighters = %d,want %d(起始 %d + 每批 %d 艘)", got, want, startFreighters, gamedata.FreighterFleetShipsPerBuild)
	}
	if want := netBC + gamedata.Profile13().FreightersCashBonus; deltaBC != want {
		t.Errorf("完工當回合 BC 變化 = %d,want %d(當回合 NetBC=%d + 現金加成 %d,Profile13)",
			deltaBC, want, netBC, gamedata.Profile13().FreightersCashBonus)
	}
	// Special 一次性行動不記入 ColonyBuildings(可重複建造,見 advanceBuilds 註解)。
	if s.ColonyBuildings[0][gamedata.FreighterFleetActionName] {
		t.Error("運輸艦隊不應記入 ColonyBuildings(需可重複建造)")
	}
}

// TestFreighterFleetCashBonusVersionDiff 驗證 1.3(FreightersCashBonus=5)與 1.5(=0)兩版
// 建造同一次運輸艦隊,完工當下「BC 變化 - 當回合 NetBC」(即現金加成本身)確實不同——這是
// diff 全量表 #4 的核心版本差異。
func TestFreighterFleetCashBonusVersionDiff(t *testing.T) {
	bonusOnly := func(profile gamedata.RuleProfile) int {
		s := NewDemoSession()
		s.DisableEvents = true
		s.RuleProfile = profile
		if s.Player.CompletedTopics == nil {
			s.Player.CompletedTopics = map[gamedata.ResearchTopic]bool{}
		}
		s.Player.CompletedTopics[gamedata.TOPIC_NUCLEAR_FISSION] = true
		deltaBC, netBC := buildFreighterFleetOneAtATime(t, s, 500)
		return deltaBC - netBC
	}

	got13 := bonusOnly(gamedata.Profile13())
	got15 := bonusOnly(gamedata.Profile15())

	if got13 != 5 {
		t.Errorf("Profile13 現金加成 = %d,want 5", got13)
	}
	if got15 != 0 {
		t.Errorf("Profile15 現金加成 = %d,want 0", got15)
	}
	if got13 == got15 {
		t.Error("1.3/1.5 現金加成應不同(diff 全量表 #4 核心差異),但兩者相等")
	}
}

// TestFreighterFleetMaintenanceAppliesNextTurn 驗證 ActiveFreighters 變非 0 後,維護費
// (每艘 0.5 BC/回合,gamedata.IncomeFreighterMaintenanceCost)確實從下一回合起反映在 NetBC——
// 完工當回合的 RunEmpireTurn 已跑在 advanceBuilds 之前,吃的是舊值,故維護費從下一回合才生效。
func TestFreighterFleetMaintenanceAppliesNextTurn(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	s.RuleProfile = gamedata.Profile15() // 現金加成=0,單純看維護費,不被現金加成干擾
	if s.Player.CompletedTopics == nil {
		s.Player.CompletedTopics = map[gamedata.ResearchTopic]bool{}
	}
	s.Player.CompletedTopics[gamedata.TOPIC_NUCLEAR_FISSION] = true

	buildFreighterFleetOneAtATime(t, s, 500)
	if s.Player.ActiveFreighters == 0 {
		t.Fatal("運輸艦隊應已完工,ActiveFreighters 不應為 0")
	}

	// 完工那次 EndTurn 的 LastPlayerOutput 是在 advanceBuilds 之前算的,不含新艦艇維護費。
	if got := s.LastPlayerOutput.FreighterMaintenanceCost; got != 0 {
		t.Errorf("完工當回合的 FreighterMaintenanceCost = %d,want 0(維護費下回合才生效)", got)
	}

	s.EndTurn() // 下一回合,這次 RunEmpireTurn 才吃到新的 ActiveFreighters

	want := gamedata.IncomeFreighterMaintenanceCost(s.Player.ActiveFreighters)
	if want == 0 {
		t.Fatal("測試前提錯誤:ActiveFreighters 算出的維護費不應為 0")
	}
	if got := s.LastPlayerOutput.FreighterMaintenanceCost; got != want {
		t.Errorf("下一回合 FreighterMaintenanceCost = %d,want %d(%d 艘 * 0.5 BC)", got, want, s.Player.ActiveFreighters)
	}
}

// TestFreighterFleetNoBuildNoRegression 驗證開局不建造運輸艦隊時,ActiveFreighters 維持 0、
// 維護費維持 0——確認本次改動對「沒用到這個功能的既有對局」零回歸。
func TestFreighterFleetNoBuildNoRegression(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	for i := 0; i < 20; i++ {
		s.EndTurn()
	}
	if s.Player.ActiveFreighters != 0 {
		t.Errorf("未建造運輸艦隊時 ActiveFreighters = %d,want 0", s.Player.ActiveFreighters)
	}
	if s.LastPlayerOutput.FreighterMaintenanceCost != 0 {
		t.Errorf("未建造運輸艦隊時 FreighterMaintenanceCost = %d,want 0", s.LastPlayerOutput.FreighterMaintenanceCost)
	}
}
