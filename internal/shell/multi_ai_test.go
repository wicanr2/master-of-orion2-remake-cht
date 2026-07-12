package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
)

// TestNewDemoSessionBuildsThreeAIOpponents 驗證 2026-07-11 由 1 AI 擴為 3 AI 的骨架本身:
// 3 個 AI 對手、各有不同母星索引(不重疊、也不是玩家母星星 0)、名稱互異,且
// PlayerSpies(玩家對每個 AI 對手的間諜數,平行 AIPlayers)長度同步為 3、預設皆 0。
func TestNewDemoSessionBuildsThreeAIOpponents(t *testing.T) {
	s := NewDemoSession()

	if len(s.AIPlayers) != 3 {
		t.Fatalf("NewDemoSession 應建 3 個 AI 對手,got %d", len(s.AIPlayers))
	}

	seenStars := map[int]bool{0: true} // 星 0 = 玩家母星,AI 母星不應撞這顆
	seenNames := map[string]bool{}
	for i, a := range s.AIPlayers {
		if len(a.ColonyStars) != 1 {
			t.Fatalf("AI[%d] 應恰有 1 顆已知母星索引,got %d", i, len(a.ColonyStars))
		}
		star := a.ColonyStars[0]
		if seenStars[star] {
			t.Fatalf("AI[%d] 母星索引 %d 與其他帝國(含玩家母星)重疊", i, star)
		}
		seenStars[star] = true

		if star < 0 || star >= len(s.Stars) {
			t.Fatalf("AI[%d] 母星索引 %d 超出星系範圍(共 %d 星)", i, star, len(s.Stars))
		}
		if s.Stars[star].Owner != 2 {
			t.Fatalf("AI[%d] 母星(星 %d)Owner 應為 2(AI),got %d", i, star, s.Stars[star].Owner)
		}

		if a.Name == "" {
			t.Fatalf("AI[%d] 名稱不應為空", i)
		}
		if seenNames[a.Name] {
			t.Fatalf("AI[%d] 名稱 %q 與其他 AI 對手重複,應各自不同種族/名稱", i, a.Name)
		}
		seenNames[a.Name] = true

		if len(a.Colonies) != 1 {
			t.Fatalf("AI[%d] 開局應恰有 1 座母星殖民地,got %d", i, len(a.Colonies))
		}
		if a.OwnedStars != 1 {
			t.Fatalf("AI[%d] 開局 OwnedStars 應為 1,got %d", i, a.OwnedStars)
		}
		if a.Decider == nil {
			t.Fatalf("AI[%d] 應有 Decider(AI 性格決策器)", i)
		}
	}

	if len(s.PlayerSpies) != len(s.AIPlayers) {
		t.Fatalf("PlayerSpies 長度應與 AIPlayers 平行,got %d vs %d", len(s.PlayerSpies), len(s.AIPlayers))
	}
	for i, n := range s.PlayerSpies {
		if n != 0 {
			t.Fatalf("PlayerSpies[%d] 開局應為 0,got %d", i, n)
		}
	}
}

// TestAllAIOpponentsExpandAndBuildIndependently 是 TestAIBuildsAndExpands(ai_behavior_test.go)
// 的 N=3 版本:驗證「每個」AI 對手(不只 AIPlayers[0])都會隨回合推進獨立造艦、擴張版圖,不是
// 只有第一個 AI 在動、其餘兩個停滯(若只有 EndTurn 迴圈第一個索引在動,代表 generalize 不完整)。
func TestAllAIOpponentsExpandAndBuildIndependently(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	if len(s.AIPlayers) < 3 {
		t.Fatalf("需要 3 個 AI 對手,got %d", len(s.AIPlayers))
	}

	startFleet := make([]int, len(s.AIPlayers))
	startColonies := make([]int, len(s.AIPlayers))
	for i, a := range s.AIPlayers {
		startFleet[i] = a.FleetStrength
		startColonies[i] = len(a.Colonies)
	}

	for i := 0; i < 40; i++ {
		s.EndTurn()
	}

	for i, a := range s.AIPlayers {
		if a.FleetStrength <= startFleet[i] {
			t.Errorf("AI[%d](%s)軍力應成長:%d → %d", i, a.Name, startFleet[i], a.FleetStrength)
		}
		if len(a.Colonies) <= startColonies[i] {
			t.Errorf("AI[%d](%s)殖民地數應增加(擴張):%d → %d", i, a.Name, startColonies[i], len(a.Colonies))
		}
		if len(a.Colonies) != len(a.ColonyStars) {
			t.Errorf("AI[%d] Colonies/ColonyStars 平行陣列長度須一致,got %d vs %d", i, len(a.Colonies), len(a.ColonyStars))
		}
	}
}

// TestCouncilNWayVotingSecondAIWins 驗證議會 N 帝國計票不是「只看 AIPlayers[0]」的殘留假設:
// 讓 AIPlayers[1](非索引 0)人口壓倒性領先,預期 PendingCouncilElection.EnemyName 是
// AIPlayers[1].Name,而不是誤指向 AIPlayers[0]。
func TestCouncilNWayVotingSecondAIWins(t *testing.T) {
	s := NewDemoSession()
	if len(s.AIPlayers) < 3 {
		t.Fatalf("需要至少 3 個 AI 對手,got %d", len(s.AIPlayers))
	}
	settleHalfGalaxy(s)
	s.PlayerColonies[0].Population = 1
	s.AIPlayers[0].Colonies[0].Population = 1
	s.AIPlayers[1].Colonies[0].Population = 100 // 第二個 AI(索引1)壓倒性領先
	s.AIPlayers[2].Colonies[0].Population = 1

	s.EndTurn()

	if s.Victory.Over {
		t.Fatalf("AI 當選不應自動結束遊戲,應等玩家回應,got Victory=%+v", s.Victory)
	}
	if s.PendingCouncilElection == nil {
		t.Fatal("AIPlayers[1] 達 2/3 多數應留下待回應選舉")
	}
	if s.PendingCouncilElection.EnemyName != s.AIPlayers[1].Name {
		t.Fatalf("當選者應為 AIPlayers[1](%q),got %q", s.AIPlayers[1].Name, s.PendingCouncilElection.EnemyName)
	}
}

// TestCouncilNWayVotingNoneReachesMajorityWithThreeAI 驗證 3 個 AI 人口分散、無人(含玩家)
// 單獨達 2/3 多數時流會——確保 generalize 後的判定不會因為「多個 AI 人口加總」被誤判成某一方
// 達標(見 advanceCouncil 註解:每個帝國各自的票獨立跟總票數比較,不把 AI 們的人口灌成一票)。
func TestCouncilNWayVotingNoneReachesMajorityWithThreeAI(t *testing.T) {
	s := NewDemoSession()
	if len(s.AIPlayers) < 3 {
		t.Fatalf("需要至少 3 個 AI 對手,got %d", len(s.AIPlayers))
	}
	settleHalfGalaxy(s)
	// 玩家 40、三個 AI 各 20,總票 100。每個帝國各自的票數都遠低於 2/3(≈67),沒有人達標。
	// 這個案例同時排除「誤把 AI 們的人口加總當一票」的舊二元邏輯殘留:若真的誤用「AI 陣營合計
	// (60)」跟玩家(40)比,60/100 一樣不到 2/3,兩種算法在這組數字下都應該流會——之所以還是
	// 值得測,是確認 generalize 後每個帝國真的各自獨立判定,不是巧合算出同一個結論。
	s.PlayerColonies[0].Population = 40
	s.AIPlayers[0].Colonies[0].Population = 20
	s.AIPlayers[1].Colonies[0].Population = 20
	s.AIPlayers[2].Colonies[0].Population = 20

	s.EndTurn()

	if s.Victory.Over {
		t.Fatalf("票數分散,任何一方都不到 2/3 多數,不應分出勝負,got Victory=%+v", s.Victory)
	}
	if s.PendingCouncilElection != nil {
		t.Fatalf("不應有一方達標,不該留下待回應選舉,got %+v", s.PendingCouncilElection)
	}
	if s.CouncilMeetings != 1 {
		t.Fatalf("應已召開第1屆(流會),got CouncilMeetings=%d", s.CouncilMeetings)
	}
}

// TestCouncilRequiresManualMinimumThreeExtantRaces 驗證移除 councilMinExtantRacesOverride
// 後,councilEligible 真的改用手冊字面門檻 gamedata.CouncilMinExtantRaces(3):只有玩家 + 1
// 個 AI(共 2 個存續帝國)時,即使半數銀河已殖民,議會依手冊原文「3 個以上存續種族」也不應成立
// ——這不是 regression,是訂正回手冊原意(先前的 councilMinExtantRacesOverride=2 才是「資料
// 模型限制下的權宜近似」,見已刪除的該常數註解)。
func TestCouncilRequiresManualMinimumThreeExtantRaces(t *testing.T) {
	s := NewDemoSession()
	s.AIPlayers = s.AIPlayers[:1] // 只留 1 個 AI,模擬存續帝國數退回「玩家+1 AI」=2
	s.PlayerSpies = s.PlayerSpies[:1]
	settleHalfGalaxy(s)

	if s.councilEligible() {
		t.Fatalf("只有 2 個存續帝國(未達手冊 3 的門檻),議會不應成立")
	}
	s.EndTurn()
	if s.CouncilMeetings != 0 {
		t.Fatalf("議會未成立時不應召開,got CouncilMeetings=%d", s.CouncilMeetings)
	}
}

// TestAdvanceCouncilNWayLogicWorksAtManualMinimum 驗證 advanceCouncil 的 N 帝國計票邏輯在
// 「剛好達到手冊門檻」(玩家 + 2 個 AI = 3 個存續帝國)時正確運作,不是只在 NewDemoSession
// 預設的 3 AI(共 4 帝國)下才碰巧正確——3 帝國是能讓議會成立的最小規模,故意在這個邊界值上
// 重跑一次「某 AI 達 2/3 多數」的判定,確保 generalize 沒有隱藏「至少 4 帝國」的假設。
func TestAdvanceCouncilNWayLogicWorksAtManualMinimum(t *testing.T) {
	s := NewDemoSession()
	s.AIPlayers = s.AIPlayers[:2] // 玩家 + 2 個 AI = 3 個存續帝國,剛好達手冊門檻
	s.PlayerSpies = s.PlayerSpies[:2]
	settleHalfGalaxy(s)
	s.PlayerColonies[0].Population = 1
	s.AIPlayers[0].Colonies[0].Population = 1
	s.AIPlayers[1].Colonies[0].Population = 100 // 第二個 AI 壓倒性領先

	s.EndTurn()

	if s.Victory.Over {
		t.Fatalf("AI 當選不應自動結束遊戲,應等玩家回應")
	}
	if s.PendingCouncilElection == nil {
		t.Fatal("AI 達 2/3 多數應留下待回應選舉")
	}
	if s.PendingCouncilElection.EnemyName != s.AIPlayers[1].Name {
		t.Fatalf("當選者應為 AIPlayers[1](%q),got %q", s.AIPlayers[1].Name, s.PendingCouncilElection.EnemyName)
	}

	s.RespondToCouncilElection(true)
	if !s.Victory.Over || s.Victory.Reason != engine.VictoryHighCouncil || s.Victory.Winner != s.AIPlayers[1].Name {
		t.Fatalf("接受後應以 engine.VictoryHighCouncil 結束,Winner=AIPlayers[1],got Over=%v Reason=%v Winner=%q",
			s.Victory.Over, s.Victory.Reason, s.Victory.Winner)
	}
}

// TestCouncilSwingVotesElectFrontrunner 驗證忠實搖擺票機制:玩家是票數領先者但自身票不足 2/3,
// 兩個對玩家友好(關係達 councilSwingVoteMinRelation)的 AI 把票投給玩家,推玩家越過 2/3 當選。
// 直接呼叫 advanceCouncil(不走 EndTurn),避免 advanceAI 每回合重算 AIOpponent.Relation 沖掉
// 這裡刻意設定的友好關係——比照 TestConquestVictory* 直呼引擎函式的既有測試手法。
//
// 場景:玩家 40、AI0 30、AI1 20、AI2 10,總 100(2/3≈67)。候選人=玩家(40)+ AI0(30)。
// AI1/AI2 對玩家友好(Relation=20)、對 AI0 中立(AIRelations 預設 0)→ 兩者搖擺票(20+10)投玩家
// → 玩家 70/100 達 2/3 當選。玩家單靠自身 40 票達不到,證明搖擺票確實改變了選舉結果。
func TestCouncilSwingVotesElectFrontrunner(t *testing.T) {
	s := NewDemoSession()
	settleHalfGalaxy(s)
	s.PlayerColonies[0].Population = 40
	s.AIPlayers[0].Colonies[0].Population = 30
	s.AIPlayers[1].Colonies[0].Population = 20
	s.AIPlayers[2].Colonies[0].Population = 10
	s.ensureAIRelations()
	s.AIPlayers[1].Relation = 20 // AI1 對玩家友好
	s.AIPlayers[2].Relation = 20 // AI2 對玩家友好

	s.advanceCouncil()

	if !s.Victory.Over {
		t.Fatalf("友好 AI 的搖擺票應把玩家推過 2/3 當選,got Victory=%+v LastCouncil=%q", s.Victory, s.LastCouncil)
	}
	if s.Victory.Reason != engine.VictoryHighCouncil || s.Victory.Winner != "player" {
		t.Fatalf("應為玩家以議會選舉當選,got Reason=%v Winner=%q", s.Victory.Reason, s.Victory.Winner)
	}
}

// TestCouncilNeutralAISAbstainNoWinner 是上一個測試的對照組:同樣的人口分布(玩家 40 領先),
// 但兩個搖擺 AI 對玩家「中立」(Relation=0,未達友好門檻)→ 依手冊棄權,不投給任一候選人 →
// 玩家維持 40 票,達不到 2/3,流會。證明搖擺票的關鍵是「友好門檻」,不是「領先者自動吸走中立票」。
func TestCouncilNeutralAISAbstainNoWinner(t *testing.T) {
	s := NewDemoSession()
	settleHalfGalaxy(s)
	s.PlayerColonies[0].Population = 40
	s.AIPlayers[0].Colonies[0].Population = 30
	s.AIPlayers[1].Colonies[0].Population = 20
	s.AIPlayers[2].Colonies[0].Population = 10
	s.ensureAIRelations() // 全部關係預設 0(中立),未達 councilSwingVoteMinRelation

	s.advanceCouncil()

	if s.Victory.Over {
		t.Fatalf("中立 AI 應棄權,玩家 40 票達不到 2/3,不應分出勝負,got Victory=%+v", s.Victory)
	}
	if s.PendingCouncilElection != nil {
		t.Fatalf("無人達 2/3,不該留下待回應選舉,got %+v", s.PendingCouncilElection)
	}
}
