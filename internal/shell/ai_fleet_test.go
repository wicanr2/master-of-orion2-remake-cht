package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/ai"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// AI 艦隊開局停在自己的母星——而且**不能靠零值**。
//
// `FleetStar` 的零值 0 是合法的星索引,所以初始化要看 `FleetPosSet`。
// 這一條同時涵蓋舊存檔(解出來 FleetPosSet=false)。
func TestAIFleetStartsAtItsHomeworld(t *testing.T) {
	s := NewDemoSession()
	for i := range s.AIPlayers {
		a := s.AIPlayers[i]
		if a.FleetPosSet {
			t.Errorf("AI %d 開局不該已經設過位置(應由 advanceAIFleets 初始化)", i)
		}
		if got, want := aiFleetStar(a), a.ColonyStars[0]; got != want {
			t.Errorf("AI %d 未初始化時應退回母星 %d,得到 %d", i, want, got)
		}
	}
	s.advanceAIFleets()
	for i := range s.AIPlayers {
		a := s.AIPlayers[i]
		if !a.FleetPosSet {
			t.Errorf("AI %d 跑過一回合後位置應已設定", i)
		}
		if a.FleetStar != a.ColonyStars[0] {
			t.Errorf("AI %d 應停在母星 %d,得到 %d", i, a.ColonyStars[0], a.FleetStar)
		}
	}
}

// 突襲不再是瞬移的:艦隊要**先飛過去**,抵達那一回合才打得到。
//
// 這是第 47 項(AI艦隊移動)的核心行為改變,也是玩家「看得到它來」的依據。
func TestAIRaidRequiresTheFleetToArriveFirst(t *testing.T) {
	s := newRaidTestSession(t)
	s.advanceAIFleets() // 初始化位置 + 決定出兵

	launched := -1
	for i := range s.AIPlayers {
		if s.AIPlayers[i].FleetETA > 0 {
			launched = i
			break
		}
	}
	if launched < 0 {
		t.Fatal("條件全滿足時應該有 AI 派出艦隊")
	}
	eta := s.AIPlayers[launched].FleetETA
	if eta < 1 {
		t.Fatalf("ETA 應至少 1 回合,得到 %d", eta)
	}

	// 在途期間一次都不能打到。
	for k := 0; k < eta; k++ {
		s.advanceAIRaids()
		if s.LastRaidReport != nil {
			t.Fatalf("艦隊還在路上(第 %d/%d 回合)就打到了:%+v", k+1, eta, s.LastRaidReport)
		}
		s.Turn++
		s.advanceAIFleets()
	}
	// 抵達了才算數。
	if s.AIPlayers[launched].FleetETA != 0 {
		t.Fatalf("%d 回合後應已抵達,ETA 還是 %d", eta, s.AIPlayers[launched].FleetETA)
	}
	if _, ok := s.aiFleetAtPlayerColony(launched); !ok {
		t.Fatal("抵達後應被判定為「停在玩家殖民地上空」")
	}
	a := s.AIPlayers[launched]
	if a.Treaty.FormalPolicy != gamedata.DIPLO_WAR || !a.WantsAudience || a.AudienceReason != AudienceReasonWar {
		t.Fatalf("AI 艦隊抵達玩家殖民星後未建立正式宣戰：policy=%d audience=%v/%q",
			a.Treaty.FormalPolicy, a.WantsAudience, a.AudienceReason)
	}
}

// 停在原地的艦隊**不能每回合都打**。
//
// 這一條守的是第 47 項(AI艦隊移動)差點引入的一個真 bug:願打門檻(態勢/軍力/性格)搬到了出發那一刻
// 之後,結算端若不保留間隔守門,一支抵達後就停著不動的艦隊會每一回合結算一次突襲。
func TestParkedAIFleetDoesNotRaidEveryTurn(t *testing.T) {
	s := newRaidTestSession(t)
	parkAIFleetsAtPlayerColony(s)
	// 只留一個 AI,免得「一回合最多一次突襲」的限制混淆計數。
	s.AIPlayers = s.AIPlayers[:1]

	raids := 0
	for k := 0; k < aiRaidInterval; k++ {
		s.advanceAIRaids()
		if s.LastRaidReport != nil {
			raids++
		}
		s.Turn++
	}
	if raids != 1 {
		t.Errorf("%d 回合內停在原地的艦隊應只打 1 次,實際打了 %d 次", aiRaidInterval, raids)
	}
}

// 阿提米絲系統網現在打得到 AI 了——這條缺口從第 38 項(行星護盾等三棟)記到第 47 項(AI艦隊移動)。
func TestArtemisNetNowHitsArrivingAIFleets(t *testing.T) {
	s := newRaidTestSession(t)
	s.advanceAIFleets() // 初始化位置

	star := s.PlayerColonyStarIndex(0)
	s.ColonyBuildings[0][testBuildingByRawID(t, 3).NameZH] = true

	i := 0
	before := s.AIPlayers[i].FleetStrength
	// 手動把艦隊設成「下一回合抵達那顆星」,跳過航程。
	s.AIPlayers[i].FleetDestStar, s.AIPlayers[i].FleetETA = star, 1
	// advanceAIFleets **回傳**抵達清單;s.LastAIArrivals 是 EndTurn 才填的。
	arrivals := s.advanceAIFleets()

	if s.AIPlayers[i].FleetStrength >= before {
		t.Errorf("進入雷區的 AI 艦隊戰力應下降:%d → %d", before, s.AIPlayers[i].FleetStrength)
	}
	found := false
	for _, arr := range arrivals {
		if arr.StarIdx == star && arr.Mines != nil {
			found = true
			if arr.Mines.TotalDamage <= 0 {
				t.Errorf("水雷回報的傷害應 > 0,得到 %d", arr.Mines.TotalDamage)
			}
		}
	}
	if !found {
		t.Errorf("抵達紀錄裡應有一筆帶水雷的:%+v", arrivals)
	}
}

// 正對照:沒建阿提米絲的星不會扣戰力。少了這條,「無條件扣戰力」也會讓上面那支通過。
func TestArrivingAtAStarWithoutTheNetCostsNothing(t *testing.T) {
	s := newRaidTestSession(t)
	s.advanceAIFleets()
	star := s.PlayerColonyStarIndex(0)
	delete(s.ColonyBuildings[0], testBuildingByRawID(t, 3).NameZH)

	i := 0
	before := s.AIPlayers[i].FleetStrength
	s.AIPlayers[i].FleetDestStar, s.AIPlayers[i].FleetETA = star, 1
	arrivals := s.advanceAIFleets()

	if s.AIPlayers[i].FleetStrength != before {
		t.Errorf("沒有雷區不該扣戰力:%d → %d", before, s.AIPlayers[i].FleetStrength)
	}
	for _, arr := range arrivals {
		if arr.StarIdx == star && arr.Mines != nil {
			t.Errorf("沒有雷區不該有水雷回報:%+v", arr.Mines)
		}
	}
}

// 航速:AI 沒有任何引擎科技時退回核融引擎(階 1),不是 0。
//
// 退回 0 的話 ETA 會被夾成 1,整個移動模型形同虛設——而且**畫面上完全看不出來**,
// 玩家只會覺得 AI 好像會瞬移。
func TestAIFleetSpeedFallsBackToNuclearDrive(t *testing.T) {
	s := NewDemoSession()
	for i := range s.AIPlayers {
		if sp := aiFleetSpeedParsecs(s.AIPlayers[i]); sp < 1 {
			t.Errorf("AI %d 航速應至少 1 秒差距/回合,得到 %d", i, sp)
		}
	}
}

// ETA 隨距離變長——不是所有目標都「一回合到」。
func TestAIFleetETAGrowsWithDistance(t *testing.T) {
	s := NewDemoSession()
	a := s.AIPlayers[0]
	home := a.ColonyStars[0]
	near, far, nearD, farD := -1, -1, 1<<30, -1
	for idx := range s.Stars {
		if idx == home {
			continue
		}
		d := s.ParsecsBetweenStars(home, idx)
		if d < nearD {
			near, nearD = idx, d
		}
		if d > farD {
			far, farD = idx, d
		}
	}
	if near < 0 || far < 0 || nearD == farD {
		t.Skip("這張星圖上找不到距離差異夠大的兩顆星")
	}
	nearETA, farETA := s.aiFleetETATo(a, home, near), s.aiFleetETATo(a, home, far)
	if nearETA < 1 || farETA < 1 {
		t.Fatalf("ETA 應至少 1:近 %d 遠 %d", nearETA, farETA)
	}
	if farETA < nearETA {
		t.Errorf("遠的目標 ETA(%d)不該比近的(%d)短", farETA, nearETA)
	}
	// 同一顆星回 0(呼叫端當成「不用飛」)。
	if got := s.aiFleetETATo(a, home, home); got != 0 {
		t.Errorf("原地不動應回 0,得到 %d", got)
	}
}

// 和平主義 AI 從不派艦隊——出發那一端的守門要真的擋得住。
func TestPacifistAINeverLaunchesAFleet(t *testing.T) {
	s := newRaidTestSession(t)
	for i := range s.AIPlayers {
		s.AIPlayers[i].Personality = ai.PersonalityPacifist
	}
	for turn := 0; turn < 30; turn++ {
		s.Turn = aiRaidGraceTurns + turn
		s.advanceAIFleets()
		if aiFleetLaunched(s) {
			t.Fatalf("第 %d 回合和平主義 AI 派出了艦隊", s.Turn)
		}
	}
}

// 艦隊位置要進存檔——不存的話,一支飛了八回合快到玩家家門口的艦隊,
// 存一次檔就瞬移回母星。這種 bug 只有存讀檔的玩家會遇到,測試不寫就抓不到。
func TestAIFleetPositionSurvivesSaveLoad(t *testing.T) {
	s := newRaidTestSession(t)
	s.advanceAIFleets() // 初始化位置 + 派出艦隊

	launched := -1
	for i := range s.AIPlayers {
		if s.AIPlayers[i].FleetETA > 0 {
			launched = i
			break
		}
	}
	if launched < 0 {
		t.Fatal("測試前提不成立:應該有 AI 派出艦隊")
	}
	want := s.AIPlayers[launched]

	got := s.snapshot().restore().AIPlayers[launched]
	if got.FleetStar != want.FleetStar || !got.FleetPosSet {
		t.Errorf("艦隊位置沒存到:want star=%d set=%v,got star=%d set=%v",
			want.FleetStar, want.FleetPosSet, got.FleetStar, got.FleetPosSet)
	}
	if got.FleetDestStar != want.FleetDestStar || got.FleetETA != want.FleetETA {
		t.Errorf("航線沒存到:want dest=%d eta=%d,got dest=%d eta=%d",
			want.FleetDestStar, want.FleetETA, got.FleetDestStar, got.FleetETA)
	}
}
