package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// newFleetAtAIHomeSession 建一個新對局,並把玩家艦隊直接擺到 AI 母星上空(已抵達,ETA=0),
// 供入侵相關測試省去先跑 SendFleet/EndTurn 航行流程。回傳對局與 AI 母星的 Stars 索引。
func newFleetAtAIHomeSession(t *testing.T) (*GameSession, int) {
	t.Helper()
	s := NewDemoSession()
	s.DisableEvents = true
	if len(s.AIPlayers) == 0 || len(s.AIPlayers[0].ColonyStars) == 0 {
		t.Fatal("需至少一個有 ColonyStars 對映的 AI 對手")
	}
	starIdx := s.AIPlayers[0].ColonyStars[0]
	s.FleetAtStar = starIdx
	s.FleetDestStar = -1
	s.FleetETA = 0
	return s, starIdx
}

// TestInvadeColony_StrongAttackerWinsMost 驗證:玩家兵力(40 陸戰隊)+ 裝甲科技(精金裝甲,
// +25 force)遠強於 AI 母星駐軍(公式算出的 ~4 單位、無裝甲科技加成)時,入侵應高機率獲勝,
// 且勝利後星 Owner 轉 1、AI 殖民地從 AIPlayers[0].Colonies 移除、玩家新增一筆殖民地。
func TestInvadeColony_StrongAttackerWinsMost(t *testing.T) {
	const n = 100
	wins := 0
	for i := 0; i < n; i++ {
		s, starIdx := newFleetAtAIHomeSession(t)
		s.Turn = i + 1
		s.FleetMarines = 40
		s.Player.CompletedTopics[gamedata.TOPIC_MOLECULAR_CONTROL] = true // 精金裝甲 +25 force

		beforePlayerColonies := len(s.PlayerColonies)
		beforeAIColonies := len(s.AIPlayers[0].Colonies)
		beforeAIOwnedStars := s.AIPlayers[0].OwnedStars

		res := s.InvadeColony(starIdx)
		if !res.Ok {
			t.Fatalf("i=%d: 前置條件應齊備,InvadeColony 應可發動,got Reason=%q", i, res.Reason)
		}
		if res.AttackerWon != res.StarCaptured {
			t.Fatalf("i=%d: AttackerWon=%v 與 StarCaptured=%v 不一致", i, res.AttackerWon, res.StarCaptured)
		}
		if res.AttackerWon {
			wins++
			if s.Stars[starIdx].Owner != 1 {
				t.Fatalf("i=%d: 攻方勝後星 Owner 應轉 1,got %d", i, s.Stars[starIdx].Owner)
			}
			if len(s.PlayerColonies) != beforePlayerColonies+1 {
				t.Fatalf("i=%d: 玩家殖民地應 +1(%d→%d),got %d", i, beforePlayerColonies, beforePlayerColonies+1, len(s.PlayerColonies))
			}
			if len(s.AIPlayers[0].Colonies) != beforeAIColonies-1 {
				t.Fatalf("i=%d: AI 殖民地應 -1(%d→%d),got %d", i, beforeAIColonies, beforeAIColonies-1, len(s.AIPlayers[0].Colonies))
			}
			if s.AIPlayers[0].OwnedStars != beforeAIOwnedStars-1 {
				t.Fatalf("i=%d: AI OwnedStars 應 -1(%d→%d),got %d", i, beforeAIOwnedStars, beforeAIOwnedStars-1, s.AIPlayers[0].OwnedStars)
			}
			if s.PlayerOwnedStars() < 2 { // 母星(1)+ 新佔領星(1)
				t.Fatalf("i=%d: PlayerOwnedStars() 應至少為 2,got %d", i, s.PlayerOwnedStars())
			}
			if s.FleetMarines != res.AttackerSurvived {
				t.Fatalf("i=%d: FleetMarines 應回寫攻方存活數,got %d want %d", i, s.FleetMarines, res.AttackerSurvived)
			}
		} else {
			if s.Stars[starIdx].Owner != 2 {
				t.Fatalf("i=%d: 攻方敗,星 Owner 不應變動,got %d", i, s.Stars[starIdx].Owner)
			}
			if len(s.AIPlayers[0].Colonies) != beforeAIColonies {
				t.Fatalf("i=%d: 攻方敗,AI 殖民地不應變動", i)
			}
		}
	}
	rate := float64(wins) / n
	if rate <= 0.85 {
		t.Fatalf("兵力 40 + 精金裝甲 vs AI 母星駐軍(~4、無裝甲加成),%d 場攻方勝率 = %.2f,預期 > 0.85", n, rate)
	}
	t.Logf("強攻方 %d 場勝率 = %.2f", n, rate)
}

// TestInvadeColony_StrongDefenderWinsMost 驗證反向情境:守方兵力(30 駐軍,由灌高人口撐大
// GroundMarineBarracksCap 而來)+ 裝甲科技遠強於玩家(僅 1 陸戰隊、無裝甲科技)時,入侵應
// 高機率落敗,且星 Owner 與雙方殖民地清單皆不變動。
func TestInvadeColony_StrongDefenderWinsMost(t *testing.T) {
	const n = 50
	wins := 0
	for i := 0; i < n; i++ {
		s, starIdx := newFleetAtAIHomeSession(t)
		s.Turn = i + 1
		s.FleetMarines = 1

		aiIdx, colonyIdx, ok := s.findAIColonyByStar(starIdx)
		if !ok {
			t.Fatalf("i=%d: 應找得到 AI 母星的殖民地模型", i)
		}
		s.AIPlayers[aiIdx].Colonies[colonyIdx].Population = 60
		s.AIPlayers[aiIdx].Colonies[colonyIdx].PopMax = 80
		s.AIPlayers[aiIdx].Player.CompletedTopics[gamedata.TOPIC_MOLECULAR_CONTROL] = true // 精金裝甲 +25 force

		res := s.InvadeColony(starIdx)
		if !res.Ok {
			t.Fatalf("i=%d: 前置條件應齊備,got Reason=%q", i, res.Reason)
		}
		if res.AttackerWon {
			wins++
		} else {
			if s.Stars[starIdx].Owner != 2 {
				t.Fatalf("i=%d: 攻方敗,星 Owner 不應變動,got %d", i, s.Stars[starIdx].Owner)
			}
			if res.AttackerSurvived != 0 {
				t.Fatalf("i=%d: 攻方僅 1 單位、敗方存活理應為 0,got %d", i, res.AttackerSurvived)
			}
		}
	}
	rate := float64(wins) / n
	if rate >= 0.15 {
		t.Fatalf("玩家僅 1 陸戰隊 vs 守方 ~30 駐軍 + 精金裝甲,%d 場攻方勝率 = %.2f,預期 < 0.15", n, rate)
	}
	t.Logf("強守方 %d 場攻方勝率 = %.2f", n, rate)
}

// TestInvadeColony_Deterministic 驗證同回合、同星索引、同輸入狀態下,rng 種子化使結果可重現。
func TestInvadeColony_Deterministic(t *testing.T) {
	build := func() (*GameSession, int) {
		s, starIdx := newFleetAtAIHomeSession(t)
		s.Turn = 7
		s.FleetMarines = 6
		return s, starIdx
	}
	s1, idx1 := build()
	s2, idx2 := build()
	if idx1 != idx2 {
		t.Fatalf("兩次建立的目標星索引應相同,got %d / %d", idx1, idx2)
	}
	r1 := s1.InvadeColony(idx1)
	r2 := s2.InvadeColony(idx2)
	if r1 != r2 {
		t.Fatalf("相同輸入的入侵解算應可重現,got %+v vs %+v", r1, r2)
	}
}

// TestInvadeColony_PreconditionsChecked 驗證三個前置條件缺一都會被擋下(Ok=false),
// 且不會誤動任何狀態。
func TestInvadeColony_PreconditionsChecked(t *testing.T) {
	s, starIdx := newFleetAtAIHomeSession(t)

	// 條件 1:艦隊尚未抵達(仍在母星,FleetETA 尚未歸零)。
	s.FleetAtStar = 0
	s.FleetETA = 3
	if res := s.InvadeColony(starIdx); res.Ok {
		t.Fatalf("艦隊未抵達不應允許入侵,got Ok=true")
	}

	// 條件 2:已抵達但沒有載運陸戰隊。
	s.FleetAtStar = starIdx
	s.FleetETA = 0
	s.FleetMarines = 0
	if res := s.InvadeColony(starIdx); res.Ok {
		t.Fatalf("無陸戰隊不應允許入侵,got Ok=true")
	}

	// 條件 3:目標星非敵方(玩家自己的母星)。
	s.FleetMarines = 5
	s.FleetAtStar = 0
	if res := s.InvadeColony(0); res.Ok {
		t.Fatalf("非敵方星不應允許入侵,got Ok=true")
	}
}

// TestInvadeColony_UnmodeledExpansionStarRejected 驗證 aiExpand 產生的「有 Owner 旗標、
// 無實際殖民地模型」的星不可入侵(簡化限制,見 AIOpponent.ColonyStars 註解)。
func TestInvadeColony_UnmodeledExpansionStarRejected(t *testing.T) {
	s, home := newFleetAtAIHomeSession(t)
	target := -1
	for i, st := range s.Stars {
		if i != home && st.Owner == 0 {
			target = i
			break
		}
	}
	if target < 0 {
		t.Fatal("找不到可用的無主星做測試")
	}
	s.Stars[target].Owner = 2 // 模擬 aiExpand:只標 Owner,不建殖民地模型
	s.FleetAtStar = target
	s.FleetETA = 0
	s.FleetMarines = 10

	res := s.InvadeColony(target)
	if res.Ok {
		t.Fatalf("無殖民地模型的星不應允許入侵,got Ok=true(Reason=%q)", res.Reason)
	}
}

// TestAdvanceMarines_GrowsOverTurnsUpToCap 驗證 Marine Barracks 駐軍池隨回合成長,且不超過
// GroundMarineBarracksCap 上限。把殖民地人口灌高到 16(cap=min(8,10)=8,高於初始 4),觀察
// 駐軍池從 4 成長到 8。
func TestAdvanceMarines_GrowsOverTurnsUpToCap(t *testing.T) {
	s := NewDemoSession()
	s.PlayerColonies[0].Population = 16 // PopMax 沿用 playerHomeworldColony() 的 20

	wantCap := gamedata.GroundMarineBarracksCap(16, 20, false)
	if wantCap != 8 {
		t.Fatalf("測試前提錯誤:預期 cap=8,got %d(檢查 GroundMarineBarracksCap 公式是否變動)", wantCap)
	}

	s.advanceMarines() // age=0:初始 4 單位(未達 8 的上限)
	if s.PlayerColonyMarines[0] != 4 {
		t.Fatalf("首次 advanceMarines 後應為手冊初始值 4,got %d", s.PlayerColonyMarines[0])
	}

	for i := 0; i < 30; i++ {
		s.advanceMarines()
	}
	if s.PlayerColonyMarines[0] != wantCap {
		t.Fatalf("30+1 回合後駐軍應成長到上限 %d,got %d", wantCap, s.PlayerColonyMarines[0])
	}
}

// TestAdvanceMarines_RespectsCapWhenPopulationSmall 驗證人口偏低(母星預設 8/20)時,cap=4,
// 即使跑很多回合也不會超過(不像 popAccum 式成長無上限)。
func TestAdvanceMarines_RespectsCapWhenPopulationSmall(t *testing.T) {
	s := NewDemoSession() // Population=8, PopMax=20 → cap = min(4,10) = 4
	for i := 0; i < 100; i++ {
		s.advanceMarines()
	}
	if s.PlayerColonyMarines[0] != 4 {
		t.Fatalf("人口 8 時 cap 應為 4,100 回合後 got %d", s.PlayerColonyMarines[0])
	}
}

// TestAdvanceMarines_NoBarracksNoGrowth 驗證沒有海軍陸戰隊營的殖民地不會生成陸戰隊。
func TestAdvanceMarines_NoBarracksNoGrowth(t *testing.T) {
	s := NewDemoSession()
	s.ColonyBuildings[0] = map[string]bool{} // 移除海軍陸戰隊營標記
	for i := 0; i < 10; i++ {
		s.advanceMarines()
	}
	if s.PlayerColonyMarines[0] != 0 {
		t.Fatalf("無 Marine Barracks 的殖民地不應生成陸戰隊,got %d", s.PlayerColonyMarines[0])
	}
}

// TestLoadMarines_TransportCapacityLimits 驗證 LoadMarines 受 MarineTransportCapacity 節制,
// 未載運部分留在殖民地駐軍池。
func TestLoadMarines_TransportCapacityLimits(t *testing.T) {
	s := NewDemoSession()
	s.PlayerColonyMarines = []int{999} // 手動灌爆,測試上限節制(非正常遊戲數值)
	capacity := s.MarineTransportCapacity()
	if capacity <= 0 {
		t.Fatalf("預期新遊戲艦隊(3 艘起始艦)應有正的運力上限,got %d", capacity)
	}

	n := s.LoadMarines(0)
	if n != capacity {
		t.Fatalf("應恰好載運到運力上限,got %d want %d", n, capacity)
	}
	if s.FleetMarines != capacity {
		t.Fatalf("FleetMarines 應等於運力上限,got %d", s.FleetMarines)
	}
	if s.PlayerColonyMarines[0] != 999-capacity {
		t.Fatalf("殖民地駐軍池應扣除已載運數,got %d want %d", s.PlayerColonyMarines[0], 999-capacity)
	}

	// 運力已滿,再次載運應回 0,不再變動狀態。
	if again := s.LoadMarines(0); again != 0 {
		t.Fatalf("運力已滿時再次 LoadMarines 應回 0,got %d", again)
	}
}

// TestLoadMarines_LoadsAllWhenUnderCapacity 驗證殖民地駐軍量低於運力上限時,全數載運。
func TestLoadMarines_LoadsAllWhenUnderCapacity(t *testing.T) {
	s := NewDemoSession()
	for i := 0; i < 10; i++ {
		s.advanceMarines()
	}
	pool := s.PlayerColonyMarines[0]
	if pool <= 0 {
		t.Fatal("advanceMarines 後駐軍池應 > 0(母星開局即有 Marine Barracks)")
	}
	n := s.LoadMarines(0)
	if n != pool {
		t.Fatalf("駐軍量低於運力上限時應全數載運,got %d want %d", n, pool)
	}
	if s.PlayerColonyMarines[0] != 0 {
		t.Fatalf("載運後殖民地駐軍池應歸零,got %d", s.PlayerColonyMarines[0])
	}
}

// TestEndTurn_AdvancesMarineBarracks 驗證 EndTurn 有接上 advanceMarines(不需要另外手動呼叫)。
func TestEndTurn_AdvancesMarineBarracks(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	before := s.PlayerColonyMarines // nil(尚未跑過 EndTurn)
	if before != nil {
		t.Fatalf("EndTurn 前 PlayerColonyMarines 應為 nil(懶初始化),got %v", before)
	}
	s.EndTurn()
	if len(s.PlayerColonyMarines) == 0 || s.PlayerColonyMarines[0] <= 0 {
		t.Fatalf("EndTurn 後母星(有 Marine Barracks)應已生成陸戰隊,got %v", s.PlayerColonyMarines)
	}
}

// --- 裝甲營房(Armor Barracks)戰車營生成(死碼串接:GroundArmorBarracksUnits/Cap) ---

// TestAdvanceArmor_GrowsOverTurnsUpToCap 驗證 Armor Barracks 駐軍池隨回合成長,且不超過
// GroundArmorBarracksCap 上限。母星預設沒有裝甲營房(homeworldBuildings 只給海軍陸戰隊營+
// 星基),故先手動標記,再灌高人口(40/40)讓 cap=10(手冊初始 2,每 5 回合 +1)。
func TestAdvanceArmor_GrowsOverTurnsUpToCap(t *testing.T) {
	s := NewDemoSession()
	s.ColonyBuildings[0][armorBarracksBuildingName] = true
	s.PlayerColonies[0].Population = 40
	s.PlayerColonies[0].PopMax = 40

	wantCap := gamedata.GroundArmorBarracksCap(40, 40, false)
	if wantCap != 10 {
		t.Fatalf("測試前提錯誤:預期 cap=10,got %d(檢查 GroundArmorBarracksCap 公式是否變動)", wantCap)
	}

	s.advanceArmor() // age=0:初始 2 單位(未達 10 的上限)
	if s.PlayerColonyTanks[0] != 2 {
		t.Fatalf("首次 advanceArmor 後應為手冊初始值 2,got %d", s.PlayerColonyTanks[0])
	}

	for i := 0; i < 40; i++ {
		s.advanceArmor()
	}
	if s.PlayerColonyTanks[0] != wantCap {
		t.Fatalf("40+1 回合後戰車營應成長到上限 %d,got %d", wantCap, s.PlayerColonyTanks[0])
	}
}

// TestAdvanceArmor_NoBarracksNoGrowth 驗證沒有裝甲營房的殖民地不會生成戰車營(母星開局預設
// 就是這個情境,homeworldBuildings 沒有裝甲營房)。
func TestAdvanceArmor_NoBarracksNoGrowth(t *testing.T) {
	s := NewDemoSession()
	for i := 0; i < 10; i++ {
		s.advanceArmor()
	}
	if s.PlayerColonyTanks[0] != 0 {
		t.Fatalf("無 Armor Barracks 的殖民地不應生成戰車營,got %d", s.PlayerColonyTanks[0])
	}
}

// TestEndTurn_AdvancesArmorBarracks 驗證 EndTurn 有接上 advanceArmor(母星預設無裝甲營房,
// 故驗證「跑了但沒生成」——真正生成的成長曲線見 TestAdvanceArmor_GrowsOverTurnsUpToCap)。
func TestEndTurn_AdvancesArmorBarracks(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	if s.PlayerColonyTanks != nil {
		t.Fatalf("EndTurn 前 PlayerColonyTanks 應為 nil(懶初始化),got %v", s.PlayerColonyTanks)
	}
	s.EndTurn()
	if len(s.PlayerColonyTanks) != 1 || s.PlayerColonyTanks[0] != 0 {
		t.Fatalf("EndTurn 後應已懶初始化 PlayerColonyTanks(母星無裝甲營房,值應為 0),got %v", s.PlayerColonyTanks)
	}
}

// TestLoadTanks_SharesTransportCapacityWithMarines 驗證 LoadTanks 與 LoadMarines 共用同一個
// MarineTransportCapacity() 運力池(既有簡化,見 LoadTanks 註解),先載的一方先吃到運力。
func TestLoadTanks_SharesTransportCapacityWithMarines(t *testing.T) {
	s := NewDemoSession()
	s.PlayerColonyMarines = []int{999}
	s.PlayerColonyTanks = []int{999}
	capacity := s.MarineTransportCapacity()
	if capacity <= 0 {
		t.Fatalf("預期新遊戲艦隊應有正的運力上限,got %d", capacity)
	}

	nMarines := s.LoadMarines(0)
	if nMarines != capacity {
		t.Fatalf("陸戰隊應先吃滿運力上限,got %d want %d", nMarines, capacity)
	}
	nTanks := s.LoadTanks(0)
	if nTanks != 0 {
		t.Fatalf("運力已被陸戰隊吃滿,LoadTanks 應回 0,got %d", nTanks)
	}

	// 換個順序:先載一半陸戰隊,驗證戰車能吃掉剩餘運力(而非被完全排除)。
	s2 := NewDemoSession()
	half := capacity / 2
	s2.PlayerColonyMarines = []int{half}
	s2.PlayerColonyTanks = []int{999}
	if got := s2.LoadMarines(0); got != half {
		t.Fatalf("應載運全部 %d 陸戰隊,got %d", half, got)
	}
	wantTanks := capacity - half
	if got := s2.LoadTanks(0); got != wantTanks {
		t.Fatalf("戰車應吃掉剩餘運力 %d,got %d", wantTanks, got)
	}
	if s2.FleetMarines != half || s2.FleetTanks != wantTanks {
		t.Fatalf("FleetMarines/FleetTanks 應等於已載運數,got marines=%d tanks=%d", s2.FleetMarines, s2.FleetTanks)
	}
}

// --- Battleoids 升級(GroundBattleoidHitsToKill / GroundBattleoidCombatBonus) ---

// TestTankHitsToKillFor_BattleoidsUpgrade 驗證未研究 Battleoids 沿用 GroundTankHitsToKill,
// 已研究則改用固定 3 hits 的 GroundBattleoidHitsToKill(手冊 p.81,整批換裝、不再疊加)。
func TestTankHitsToKillFor_BattleoidsUpgrade(t *testing.T) {
	s := NewDemoSession()
	if got := tankHitsToKillFor(s.Player); got != gamedata.GroundTankHitsToKill(false) {
		t.Fatalf("未研究 Battleoids 應沿用 GroundTankHitsToKill(false)=%d,got %d", gamedata.GroundTankHitsToKill(false), got)
	}
	s.Player.CompletedTopics[gamedata.TOPIC_ASTRO_CONSTRUCTION] = true
	if got := tankHitsToKillFor(s.Player); got != gamedata.GroundBattleoidHitsToKill {
		t.Fatalf("已研究 Battleoids 應改用 GroundBattleoidHitsToKill=%d,got %d", gamedata.GroundBattleoidHitsToKill, got)
	}
}

// TestTankForceBonusFor_OnlyWhenTanksPresent 驗證 Battleoid 的 +10 相對加成只在 tankCount>0
// 時套用(0 輛戰車不該白拿加成)。
func TestTankForceBonusFor_OnlyWhenTanksPresent(t *testing.T) {
	s := NewDemoSession()
	s.Player.CompletedTopics[gamedata.TOPIC_ASTRO_CONSTRUCTION] = true
	if got := tankForceBonusFor(s.Player, 0); got != 0 {
		t.Fatalf("0 輛戰車不應套用 Battleoid 加成,got %d", got)
	}
	if got := tankForceBonusFor(s.Player, 3); got != gamedata.GroundBattleoidCombatBonus {
		t.Fatalf("有戰車且已升級 Battleoid,應套用 +%d 加成,got %d", gamedata.GroundBattleoidCombatBonus, got)
	}
	// 未升級 Battleoid 則即使有戰車也不套用。
	s2 := NewDemoSession()
	if got := tankForceBonusFor(s2.Player, 3); got != 0 {
		t.Fatalf("未升級 Battleoid,有戰車也不應套用加成,got %d", got)
	}
}

// --- 戰車納入 InvadeColony 攻方 GroundForce ---

// TestInvadeColony_TanksAloneCanInvade 驗證只有戰車、沒有陸戰隊時仍可發動入侵(guard 條件已
// 從「FleetMarines>0」放寬為「FleetMarines>0 或 FleetTanks>0」)。
func TestInvadeColony_TanksAloneCanInvade(t *testing.T) {
	s, starIdx := newFleetAtAIHomeSession(t)
	s.FleetMarines = 0
	s.FleetTanks = 5
	res := s.InvadeColony(starIdx)
	if !res.Ok {
		t.Fatalf("只有戰車、無陸戰隊,應仍可發動入侵,got Reason=%q", res.Reason)
	}
}

// TestInvadeColony_NoGroundForceRejected 驗證陸戰隊與戰車皆為 0 時,仍應被前置條件擋下。
func TestInvadeColony_NoGroundForceRejected(t *testing.T) {
	s, starIdx := newFleetAtAIHomeSession(t)
	s.FleetMarines = 0
	s.FleetTanks = 0
	res := s.InvadeColony(starIdx)
	if res.Ok {
		t.Fatalf("陸戰隊與戰車皆為 0 不應允許入侵,got Ok=true")
	}
}

// TestInvadeColony_TanksSplitSurvivorsConsistently 驗證陸戰隊+戰車混編入侵後,
// AttackerMarinesSurvived/AttackerTanksSurvived 的拆解與 AttackerSurvived 總數一致,且各自
// 不超過原始載運數,FleetMarines/FleetTanks 正確回寫拆分後存活數。
func TestInvadeColony_TanksSplitSurvivorsConsistently(t *testing.T) {
	const n = 30
	for i := 0; i < n; i++ {
		s, starIdx := newFleetAtAIHomeSession(t)
		s.Turn = i + 1
		s.FleetMarines = 3
		s.FleetTanks = 4

		res := s.InvadeColony(starIdx)
		if !res.Ok {
			t.Fatalf("i=%d: 前置條件應齊備,got Reason=%q", i, res.Reason)
		}
		if res.AttackerSurvived != res.AttackerMarinesSurvived+res.AttackerTanksSurvived {
			t.Fatalf("i=%d: AttackerSurvived(%d) 應等於陸戰隊+戰車拆解和(%d+%d)", i, res.AttackerSurvived, res.AttackerMarinesSurvived, res.AttackerTanksSurvived)
		}
		if res.AttackerTanksSurvived > 4 {
			t.Fatalf("i=%d: 戰車存活數不應超過原始載運數 4,got %d", i, res.AttackerTanksSurvived)
		}
		if res.AttackerMarinesSurvived > 3 {
			t.Fatalf("i=%d: 陸戰隊存活數不應超過原始載運數 3,got %d", i, res.AttackerMarinesSurvived)
		}
		if s.FleetMarines != res.AttackerMarinesSurvived || s.FleetTanks != res.AttackerTanksSurvived {
			t.Fatalf("i=%d: FleetMarines/FleetTanks 應回寫拆分後存活數,got marines=%d(want %d) tanks=%d(want %d)",
				i, s.FleetMarines, res.AttackerMarinesSurvived, s.FleetTanks, res.AttackerTanksSurvived)
		}
	}
}

// TestInvadeColony_TanksImproveWinRate 驗證加上戰車營確實提升攻方勝率(對照組:同樣 3 陸戰隊
// 但沒有戰車 vs 3 陸戰隊+12 戰車),證明坦克真的被納入了 GroundForce 解算,不是擺著沒用的死碼。
func TestInvadeColony_TanksImproveWinRate(t *testing.T) {
	const n = 100
	winRate := func(tanks int) float64 {
		wins := 0
		for i := 0; i < n; i++ {
			s, starIdx := newFleetAtAIHomeSession(t)
			s.Turn = i + 1
			s.FleetMarines = 3
			s.FleetTanks = tanks
			res := s.InvadeColony(starIdx)
			if !res.Ok {
				t.Fatalf("tanks=%d i=%d: 前置條件應齊備,got Reason=%q", tanks, i, res.Reason)
			}
			if res.AttackerWon {
				wins++
			}
		}
		return float64(wins) / n
	}
	withoutTanks := winRate(0)
	withTanks := winRate(12)
	if withTanks <= withoutTanks {
		t.Fatalf("加上 12 輛戰車應提升攻方勝率,got 無戰車=%.2f 有戰車=%.2f", withoutTanks, withTanks)
	}
	t.Logf("無戰車勝率=%.2f 有戰車(12)勝率=%.2f", withoutTanks, withTanks)
}
