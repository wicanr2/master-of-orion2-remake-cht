package shell

import (
	"math"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/ai"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// TestAIBuildsAndExpands 驗證 AI 對手主動造艦(FleetStrength 成長)並擴張星圖(佔無主星)。
func TestAIBuildsAndExpands(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	if len(s.AIPlayers) == 0 {
		t.Fatal("需至少一個 AI 對手")
	}
	startFleet := s.AIPlayers[0].FleetStrength
	unownedStart := 0
	for _, st := range s.Stars {
		if st.Owner == 0 {
			unownedStart++
		}
	}

	for i := 0; i < 30; i++ {
		s.EndTurn()
	}

	if s.AIPlayers[0].FleetStrength <= startFleet {
		t.Fatalf("AI 應主動造艦累積軍力:%d → %d", startFleet, s.AIPlayers[0].FleetStrength)
	}
	if s.AIPlayers[0].OwnedStars == 0 {
		t.Fatal("AI 應擴張佔領無主星")
	}
	unownedEnd := 0
	for _, st := range s.Stars {
		if st.Owner == 0 {
			unownedEnd++
		}
	}
	if unownedEnd >= unownedStart {
		t.Fatalf("AI 擴張後無主星應減少:%d → %d", unownedStart, unownedEnd)
	}
	if s.AIPlayers[0].StanceName == "" {
		t.Fatal("AI 應有外交態勢")
	}
	t.Logf("AI 軍力 %d→%d、佔領 %d 星、態勢「%s」", startFleet, s.AIPlayers[0].FleetStrength, s.AIPlayers[0].OwnedStars, s.AIPlayers[0].StanceName)
}

// TestAIExpand_CreatesRealColony 驗證 aiExpand 佔領無主星時會建立真正的 engine.ColonyState
// (而不只是先前的「標旗標、無殖民地模型」簡化),且 Colonies/ColonyStars 兩個平行陣列同步、
// 不越界(見 AIOpponent.ColonyStars 欄位註解、colonization.go newColonyFromStar)。
func TestAIExpand_CreatesRealColony(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	// 這個測試要驗的是「擴張執行後會建出真的殖民地模型」,不是「AI 每回合都想擴張」——
	// 2026-08-06 起擴張前會先過性格的積極度判定(原版 _personality_expansion_chance:
	// 和平主義只有 30%),把性格固定成擴張率 100% 的冷酷無情,才測得到機制本身。
	s.AIPlayers[0].Personality = ai.PersonalityRuthless
	beforeColonies := len(s.AIPlayers[0].Colonies)
	beforeStars := len(s.AIPlayers[0].ColonyStars)

	s.aiExpand(0)

	if len(s.AIPlayers[0].Colonies) != beforeColonies+1 {
		t.Fatalf("aiExpand 後 AI Colonies 應 +1(%d→%d),got %d", beforeColonies, beforeColonies+1, len(s.AIPlayers[0].Colonies))
	}
	if len(s.AIPlayers[0].ColonyStars) != beforeStars+1 {
		t.Fatalf("aiExpand 後 AI ColonyStars 應 +1(%d→%d),got %d", beforeStars, beforeStars+1, len(s.AIPlayers[0].ColonyStars))
	}
	if len(s.AIPlayers[0].Colonies) != len(s.AIPlayers[0].ColonyStars) {
		t.Fatalf("Colonies/ColonyStars 平行陣列長度須一致,got %d vs %d", len(s.AIPlayers[0].Colonies), len(s.AIPlayers[0].ColonyStars))
	}
	newColony := s.AIPlayers[0].Colonies[len(s.AIPlayers[0].Colonies)-1]
	if newColony.Population != colonizeStartPopulation || newColony.PopMax < colonizeStartPopulation {
		t.Fatalf("新 AI 殖民地應有實際人口模型(Population=%d PopMax=%d),不應是零值旗標",
			newColony.Population, newColony.PopMax)
	}
	newStarIdx := s.AIPlayers[0].ColonyStars[len(s.AIPlayers[0].ColonyStars)-1]
	if s.Stars[newStarIdx].Owner != 2 {
		t.Fatalf("ColonyStars 記錄的星索引 %d,其 Star.Owner 應為 2(AI),got %d", newStarIdx, s.Stars[newStarIdx].Owner)
	}
}

// TestAIExpand_EconomyGrowsWithColonyCount 驗證 aiExpand 建立的新殖民地會被下一次
// engine.RunEmpireTurn 算進 AI 經濟——EndTurn 之後 AI 總淨工業/FleetStrength 應該因為擴張而
// 比「維持單一母星」時成長更快(對照修前恆定為初始母星產出、線性軍力成長)。
func TestAIExpand_EconomyGrowsWithColonyCount(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	// 同上:固定成擴張率 100% 的性格,讓「擴張 → 經濟成長」這條因果測得準,
	// 不受性格積極度的機率影響。
	for i := range s.AIPlayers {
		s.AIPlayers[i].Personality = ai.PersonalityRuthless
	}

	for turn := 0; turn < 10; turn++ {
		s.EndTurn()
	}
	if len(s.AIPlayers[0].Colonies) <= 1 {
		t.Fatalf("10 回合(含 2 次 aiExpand 時機:第5、10回合)後 AI 殖民地數應 >1,got %d", len(s.AIPlayers[0].Colonies))
	}
	fleetAfter10 := s.AIPlayers[0].FleetStrength

	for turn := 0; turn < 10; turn++ {
		s.EndTurn()
	}
	fleetAfter20 := s.AIPlayers[0].FleetStrength
	growthSecondDecade := fleetAfter20 - fleetAfter10
	growthFirstDecade := fleetAfter10 // 起始 FleetStrength=0

	if growthSecondDecade <= growthFirstDecade {
		t.Fatalf("殖民地數增加後,AI 軍力成長速度應加快(第11-20回合成長 %d 應大於第1-10回合成長 %d)",
			growthSecondDecade, growthFirstDecade)
	}
}

// TestAIExpand_NoOpWhenNoUnownedStars 驗證所有星都已有歸屬時,aiExpand 安全 no-op(不 panic、
// 不改動任何陣列長度)。
func TestAIExpand_NoOpWhenNoUnownedStars(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	for i := range s.Stars {
		if s.Stars[i].Owner == 0 {
			s.Stars[i].Owner = 1 // 全部標記已有歸屬,模擬版圖已滿
		}
	}
	beforeColonies := len(s.AIPlayers[0].Colonies)
	beforeStars := len(s.AIPlayers[0].ColonyStars)

	s.aiExpand(0) // 不應 panic

	if len(s.AIPlayers[0].Colonies) != beforeColonies || len(s.AIPlayers[0].ColonyStars) != beforeStars {
		t.Fatalf("無主星用完時 aiExpand 應 no-op,Colonies/ColonyStars 不應變動:%d→%d / %d→%d",
			beforeColonies, len(s.AIPlayers[0].Colonies), beforeStars, len(s.AIPlayers[0].ColonyStars))
	}
}

// TestAIStanceHostileWhenStrong 驗證 AI 遠強於玩家(玩家無艦隊)時,關係轉敵對。
func TestAIStanceHostileWhenStrong(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	s.Ships = nil                        // 玩家無軍力
	s.Difficulty = len(Difficulties) - 1 // 最高難度(倍率最高);用長度而非硬編索引,
	// 免得 Difficulties 增刪選項時這個測試靜默改測到別的難度(2026-08-07 補 Tutor 就踩過)。
	for i := 0; i < 40; i++ {
		s.EndTurn()
	}
	if s.AIPlayers[0].Relation >= 0 {
		t.Fatalf("AI 遠強於玩家時關係應轉負:%d", s.AIPlayers[0].Relation)
	}
	st := s.AIPlayers[0].StanceName
	if st != "宣戰" && st != "敵視" {
		t.Fatalf("AI 強勢時態勢應敵對(宣戰/敵視),實得「%s」", st)
	}
	t.Logf("AI 關係 %d、態勢「%s」", s.AIPlayers[0].Relation, st)
}

// TestAIPersonalitiesDiverge 驗證性格接線真的造成行為差異,不是只多了一個欄位。
//
// 2026-08-06 之前三個 AI 對手除了名字之外行為完全相同:profile 是手寫的、關係演化用固定
// 係數、擴張每回合必試。接上原版性格表(ai/personality_tables.go)後,擴張速度與外交走向
// 應該依性格分岔。
func TestAIPersonalitiesDiverge(t *testing.T) {
	grow := func(p ai.Personality) (colonies, relation int) {
		s := NewDemoSession()
		s.DisableEvents = true
		for i := range s.AIPlayers {
			s.AIPlayers[i].Personality = p
			s.AIPlayers[i].Relation = 0
		}
		// 20 回合:夠讓擴張速度分出高下,又不會兩邊都把 24 星的星圖佔滿而看不出差異
		// (60 回合時和平主義也會飽和,實測 7 vs 8 幾乎相同)。
		for turn := 0; turn < 20; turn++ {
			s.EndTurn()
		}
		return len(s.AIPlayers[0].Colonies), s.AIPlayers[0].Relation
	}
	pacifistColonies, pacifistRel := grow(ai.PersonalityPacifist)
	ruthlessColonies, ruthlessRel := grow(ai.PersonalityRuthless)

	// 擴張:冷酷無情(100%)應該明顯多於和平主義(30%)。
	if ruthlessColonies <= pacifistColonies {
		t.Errorf("冷酷無情的擴張應多於和平主義:%d vs %d 個殖民地", ruthlessColonies, pacifistColonies)
	}
	// 外交:和平主義的關係平衡點在友好側,不該跟冷酷無情一樣觸底。
	if pacifistRel <= ruthlessRel {
		t.Errorf("和平主義的關係應優於冷酷無情:%+d vs %+d", pacifistRel, ruthlessRel)
	}
	t.Logf("和平主義:%d 殖民地 關係%+d ／ 冷酷無情:%d 殖民地 關係%+d",
		pacifistColonies, pacifistRel, ruthlessColonies, ruthlessRel)
}

// TestAIPersonalityIsReproducible 驗證同一 seed 抽出同一組性格(存讀檔與重跑的前提)。
func TestAIPersonalityIsReproducible(t *testing.T) {
	a, b := NewDemoSession(), NewDemoSession()
	for i := range a.AIPlayers {
		if a.AIPlayers[i].Personality != b.AIPlayers[i].Personality {
			t.Errorf("AI %d 的性格不可重現:%v vs %v", i, a.AIPlayers[i].Personality, b.AIPlayers[i].Personality)
		}
	}
	// 存讀檔後性格要保留。
	restored := a.snapshot().restore()
	for i := range a.AIPlayers {
		if restored.AIPlayers[i].Personality != a.AIPlayers[i].Personality {
			t.Errorf("存讀檔後 AI %d 的性格跑掉:%v → %v",
				i, a.AIPlayers[i].Personality, restored.AIPlayers[i].Personality)
		}
	}
}

// TestAIPicksBestPlanetNotFirstAvailable 驗證 AI 擴張會挑估值最高的星,
// 而不是「掃到第一顆能殖民的就佔」(2026-08-06 之前的行為)。
func TestAIPicksBestPlanetNotFirstAvailable(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	s.AIPlayers[0].Personality = ai.PersonalityRuthless // 擴張率 100%,確保一定會出手

	// 找出所有無主星裡估值最高的那顆,以及索引最小的那顆。
	bestIdx, bestVal, firstIdx := -1, -1, -1
	for idx := range s.Stars {
		if s.Stars[idx].Owner != 0 {
			continue
		}
		if s.aiPlanetValue(0, idx) <= 0 {
			continue
		}
		if firstIdx < 0 {
			firstIdx = idx
		}
		if v := s.aiPlanetValue(0, idx); v > bestVal {
			bestIdx, bestVal = idx, v
		}
	}
	if bestIdx < 0 || bestIdx == firstIdx {
		t.Skip("這局的最佳星恰好就是索引最小的星,測不出差異")
	}

	s.aiExpand(0)
	got := s.AIPlayers[0].ColonyStars[len(s.AIPlayers[0].ColonyStars)-1]
	if got != bestIdx {
		t.Errorf("AI 應挑估值最高的星 %d(%d 分),got 星 %d(%d 分)",
			bestIdx, bestVal, got, s.aiPlanetValue(0, got))
	}
}

// TestAIAvoidsEnemyNeighborhood 驗證 AI 估值會避開敵方勢力範圍
// (原版 Compute_Contextual_Planet_Values_ 的 /(敵方鄰居數+2))。
func TestAIAvoidsEnemyNeighborhood(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true

	// 找一顆有價值的無主星,記下原始估值;把它的鄰居標成敵方後估值應下降。
	target := -1
	for i := range s.Stars {
		if s.Stars[i].Owner == 0 && s.aiPlanetValue(0, i) > 0 {
			target = i
			break
		}
	}
	if target < 0 {
		t.Skip("這局沒有可估值的無主星")
	}
	before := s.aiPlanetValue(0, target)

	// 把最近的一顆無主星標成敵方。
	nearest, nd := -1, 1e9
	for j := range s.Stars {
		if j == target || s.Stars[j].Owner != 0 {
			continue
		}
		d := math.Hypot(s.Stars[j].X-s.Stars[target].X, s.Stars[j].Y-s.Stars[target].Y)
		if d < nd {
			nearest, nd = j, d
		}
	}
	if nearest < 0 || nd > aiNeighborRadius {
		t.Skip("目標星附近沒有可標記成敵方的鄰星")
	}
	s.Stars[nearest].Owner = 2
	after := s.aiPlanetValue(0, target)
	if after >= before {
		t.Errorf("鄰近出現敵方後估值應下降:%d → %d", before, after)
	}
}

// TestAIPrefersNearbyWhenPlanetsEqual 驗證行星本身條件相同時,AI 偏好離自己近的
// (原版 Proximity_Worth_To_Player_ 的 120/distance)。
func TestAIPrefersNearbyWhenPlanetsEqual(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	home := s.AIPlayers[0].ColonyStars[0]

	// 找兩顆行星資料完全相同、但距離不同的無主星。
	nearIdx := -1
	var nearD, farD float64
	for i := range s.Stars {
		if s.Stars[i].Owner != 0 || i >= len(s.Planets) || s.Planets[i].NoPlanet {
			continue
		}
		d := math.Hypot(s.Stars[i].X-s.Stars[home].X, s.Stars[i].Y-s.Stars[home].Y)
		for j := i + 1; j < len(s.Stars); j++ {
			if s.Stars[j].Owner != 0 || j >= len(s.Planets) || s.Planets[j].NoPlanet {
				continue
			}
			if s.Planets[i].ClimateID != s.Planets[j].ClimateID ||
				s.Planets[i].SizeID != s.Planets[j].SizeID ||
				s.Planets[i].MineralID != s.Planets[j].MineralID ||
				s.Planets[i].GravityID != s.Planets[j].GravityID {
				continue
			}
			dj := math.Hypot(s.Stars[j].X-s.Stars[home].X, s.Stars[j].Y-s.Stars[home].Y)
			if d < dj {
				nearIdx, nearD, farD = i, d, dj
			} else {
				nearIdx, nearD, farD = j, dj, d
			}
			break
		}
		if nearIdx >= 0 {
			break
		}
	}
	if nearIdx < 0 {
		t.Skip("這局沒有兩顆行星條件完全相同的無主星")
	}
	// 排除鄰近效應的干擾:只比 proximity 那一項。
	nearProx := gamedata.AIProximityValue([]int{int(nearD * aiDistanceUnit)})
	farProx := gamedata.AIProximityValue([]int{int(farD * aiDistanceUnit)})
	if nearProx <= farProx {
		t.Errorf("近星的鄰近價值應高於遠星:%.3f→%d vs %.3f→%d", nearD, nearProx, farD, farProx)
	}
}
