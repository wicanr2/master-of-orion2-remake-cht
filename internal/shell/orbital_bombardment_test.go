package shell

import (
	"math/rand"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// deterministicBombardShip 建一艘保證「必中 + 固定滿傷」的艦艇,供轟炸傷害測試排除 rng
// 不確定性:atk = 1(body,"偵察艦") + WeaponAttack(100) = 101 >= 99,依
// gamedata.CombatClassicToHit/DamageForHit 的「netAttack>=99 恆命中、恆滿傷(max_dmg)」分支
// (combat.go/damage.go 已註解的手冊 [2] 規則),每發傷害固定 = wmax = atk = 101,不受 roll
// 影響,可手算驗證。
func deterministicBombardShip() Ship {
	return Ship{Name: "測試艦", Class: "偵察艦", Weapon: "電漿砲", Armor: "無裝甲", Shield: "無護盾", Special: "無", WeaponAttack: 100}
}

// TestBombardColony_PreconditionsChecked 驗證前置條件缺一都會被擋下(Ok=false),且不消耗
// 任何狀態,比照 TestInvadeColony_PreconditionsChecked 的既有模式。
func TestBombardColony_PreconditionsChecked(t *testing.T) {
	s, starIdx := newFleetAtAIHomeSession(t)

	if res := s.BombardColony(-1); res.Ok {
		t.Fatalf("無效星索引不應允許轟炸,got Ok=true")
	}

	// 艦隊尚未抵達。
	s.FleetAtStar = 0
	s.FleetETA = 3
	if res := s.BombardColony(starIdx); res.Ok {
		t.Fatalf("艦隊未抵達不應允許轟炸,got Ok=true")
	}

	// 目標星非敵方(玩家自己的母星)。
	s.FleetAtStar = 0
	s.FleetETA = 0
	if res := s.BombardColony(0); res.Ok {
		t.Fatalf("非敵方星不應允許轟炸,got Ok=true")
	}

	// 已抵達敵方星,但艦隊沒有艦艇。
	s.FleetAtStar = starIdx
	s.FleetETA = 0
	s.Ships = nil
	if res := s.BombardColony(starIdx); res.Ok {
		t.Fatalf("艦隊無艦艇不應允許轟炸,got Ok=true")
	}
}

// TestBombardColony_UnmodeledExpansionStarRejected 驗證 aiExpand 產生的「有 Owner 旗標、
// 無實際殖民地模型」的星不可轟炸(與 InvadeColony 同款簡化限制)。
func TestBombardColony_UnmodeledExpansionStarRejected(t *testing.T) {
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
	s.Ships = []Ship{deterministicBombardShip()}

	res := s.BombardColony(target)
	if res.Ok {
		t.Fatalf("無殖民地模型的星不應允許轟炸,got Ok=true(Reason=%q)", res.Reason)
	}
}

// TestBombardColony_ReducesPopulationDeterministically 用保證命中+固定滿傷的艦隊驗證整條
// 轟炸換算鏈:10 輪 × 1 艦 × 101 傷害 = 1010 總傷害 → hits=gamedata.GroundBombHitsFromDamage
// (1010)=10 → 扣減殖民地人口(母星預設 Population=8,不足 10,故全數扣光,RemainingHits=2)。
func TestBombardColony_ReducesPopulationDeterministically(t *testing.T) {
	s, starIdx := newFleetAtAIHomeSession(t)
	s.RaceCombatPct = 0
	s.Ships = []Ship{deterministicBombardShip()}

	aiIdx, colonyIdx, ok := s.findAIColonyByStar(starIdx)
	if !ok {
		t.Fatal("應找得到 AI 母星的殖民地模型")
	}
	startPop := s.AIPlayers[aiIdx].Colonies[colonyIdx].Population
	if startPop != 8 {
		t.Fatalf("測試前提錯誤:預期母星預設人口 8,got %d(playerHomeworldColony 是否變動?)", startPop)
	}

	res := s.BombardColony(starIdx)
	if !res.Ok {
		t.Fatalf("前置條件應齊備,got Reason=%q", res.Reason)
	}

	const wantTotalDamage = 10 * 101 // 10 輪 × 1 艦 × 每發固定 101 傷害
	if res.TotalDamage != wantTotalDamage {
		t.Fatalf("TotalDamage 應為固定值 %d(必中+滿傷,無 rng 不確定性),got %d", wantTotalDamage, res.TotalDamage)
	}
	wantHits := gamedata.GroundBombHitsFromDamage(wantTotalDamage)
	if wantHits != 10 {
		t.Fatalf("測試前提錯誤:預期 hits=10,got %d", wantHits)
	}
	if res.Hits != wantHits {
		t.Fatalf("Hits 應等於 GroundBombHitsFromDamage(TotalDamage)=%d,got %d", wantHits, res.Hits)
	}
	if res.PopulationLost != startPop {
		t.Fatalf("hits(%d) > 起始人口(%d),應扣光全部人口,got PopulationLost=%d", res.Hits, startPop, res.PopulationLost)
	}
	if res.RemainingHits != res.Hits-startPop {
		t.Fatalf("RemainingHits 應為 hits-人口扣減量=%d,got %d", res.Hits-startPop, res.RemainingHits)
	}
	if got := s.AIPlayers[aiIdx].Colonies[colonyIdx].Population; got != 0 {
		t.Fatalf("殖民地人口應扣減到 0(不應變負),got %d", got)
	}
}

// TestBombardColony_PopulationNeverNegative 驗證即使轟炸傷害遠超過殖民地人口所需 hits,
// Population 也只會夾在 0,不會扣成負數。
func TestBombardColony_PopulationNeverNegative(t *testing.T) {
	s, starIdx := newFleetAtAIHomeSession(t)
	// 3 艘保證滿傷艦艇,傷害遠超過母星人口(8)所需的 hits。
	s.Ships = []Ship{deterministicBombardShip(), deterministicBombardShip(), deterministicBombardShip()}

	aiIdx, colonyIdx, ok := s.findAIColonyByStar(starIdx)
	if !ok {
		t.Fatal("應找得到 AI 母星的殖民地模型")
	}

	res := s.BombardColony(starIdx)
	if !res.Ok {
		t.Fatalf("前置條件應齊備,got Reason=%q", res.Reason)
	}
	if got := s.AIPlayers[aiIdx].Colonies[colonyIdx].Population; got != 0 {
		t.Fatalf("人口不應扣成負數,應夾在 0,got %d", got)
	}
	if res.RemainingHits <= 0 {
		t.Fatalf("傷害遠超過人口所需 hits,RemainingHits 應 > 0,got %d", res.RemainingHits)
	}
}

// TestBombardColony_DoesNotCaptureStarOrColony 驗證轟炸不會佔領星(手冊:轟炸只削弱/殺人口,
// 佔領仍要靠 InvadeColony 的陸戰隊/戰車入侵),Owner 與 AI 殖民地清單皆不變動。
func TestBombardColony_DoesNotCaptureStarOrColony(t *testing.T) {
	s, starIdx := newFleetAtAIHomeSession(t)
	s.Ships = []Ship{deterministicBombardShip()}
	beforeAIColonies := len(s.AIPlayers[0].Colonies)

	res := s.BombardColony(starIdx)
	if !res.Ok {
		t.Fatalf("前置條件應齊備,got Reason=%q", res.Reason)
	}
	if s.Stars[starIdx].Owner != 2 {
		t.Fatalf("轟炸不應改變星 Owner,got %d", s.Stars[starIdx].Owner)
	}
	if len(s.AIPlayers[0].Colonies) != beforeAIColonies {
		t.Fatalf("轟炸不應增減 AI 殖民地筆數,got %d want %d", len(s.AIPlayers[0].Colonies), beforeAIColonies)
	}
}

// TestBombardColony_Deterministic 驗證同回合、同星索引、同艦隊輸入下,rng 種子化使結果可重現
// (比照 TestInvadeColony_Deterministic 的既有模式)。
func TestBombardColony_Deterministic(t *testing.T) {
	build := func() (*GameSession, int) {
		s, starIdx := newFleetAtAIHomeSession(t)
		s.Turn = 9
		s.Ships = []Ship{{Name: "測試艦", Class: "巡防艦", Weapon: "核飛彈", WeaponAttack: 40}}
		return s, starIdx
	}
	s1, idx1 := build()
	s2, idx2 := build()
	r1 := s1.BombardColony(idx1)
	r2 := s2.BombardColony(idx2)
	if r1 != r2 {
		t.Fatalf("相同輸入的轟炸解算應可重現,got %+v vs %+v", r1, r2)
	}
}

// TestFleetBombardDamage_NoShipsZeroDamage 驗證沒有艦艇時總傷害為 0(邊界情況,雖然
// BombardColony 本身已擋在「無艦艇」的前置條件,這裡直接測 fleetBombardDamage 本身的邊界)。
func TestFleetBombardDamage_NoShipsZeroDamage(t *testing.T) {
	s := &GameSession{}
	rng := rand.New(rand.NewSource(1))
	if got := s.fleetBombardDamage(rng); got != 0 {
		t.Fatalf("無艦艇時總傷害應為 0,got %d", got)
	}
}
