package shell

import (
	"math/rand"
	"reflect"
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
// (1010)=10。
//
// 2026-07-11 起 AI 母星有 ColonyBuildings(海軍陸戰隊營+星基共 2 棟,見 homeworldBuildings),
// hits 先摧毀建築才扣人口(見 orbital_bombardment.go BombardColony「建築吸收」段落)——本測試
// 用的是預設 Profile15(BombardmentBuildingBonusHits=0),每棟建築消耗 1 hit:
//
//	hits=10 → 摧毀 2 棟建築耗 2 hits → 餘 8 hits 才進人口
//	popLoss = GroundBombardPopulationLoss(8, LARGE_PLANET) = 8*6/7(整數除法)= 6
//	Population 8-6=2(未被扣光,建築確實吸收保護了部分人口——這正是本子系統要驗證的行為)
//	RemainingHits(建築+人口都扣完後的餘數)= 8-6 = 2
//
// 舊版(無 ColonyBuildings 資料模型時)hits=10 直接全數進人口,popLoss=10*6/7=8=扣光,是本測試
// 修改前的斷言;現在因為建築先吸收了 2 hits,人口損傷從 8 降到 6,不再全滅。
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
	startBuildings := len(s.AIPlayers[aiIdx].ColonyBuildings[colonyIdx])
	if startBuildings != 2 {
		t.Fatalf("測試前提錯誤:預期 AI 母星開局有 2 棟建築(海軍陸戰隊營+星基),got %d", startBuildings)
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
	const wantBuildingsDestroyed = 2 // 2 棟建築各耗 1 hit(Profile15,BombardmentBuildingBonusHits=0)
	if res.BuildingsDestroyed != wantBuildingsDestroyed {
		t.Fatalf("BuildingsDestroyed 應為 %d(hits 足夠摧毀全部 2 棟建築),got %d", wantBuildingsDestroyed, res.BuildingsDestroyed)
	}
	if res.BuildingsRemaining != 0 {
		t.Fatalf("BuildingsRemaining 應為 0(2 棟建築皆已摧毀),got %d", res.BuildingsRemaining)
	}
	const wantPopulationLost = 6 // 建築吸收後餘 8 hits → GroundBombardPopulationLoss(8, LARGE_PLANET)=8*6/7=6
	if res.PopulationLost != wantPopulationLost {
		t.Fatalf("PopulationLost 應為 %d(建築先吸收 2 hits,餘 8 hits 才扣人口),got %d", wantPopulationLost, res.PopulationLost)
	}
	const wantRemainingHits = 2 // 餘 8 hits 扣完 6 人口損傷後的剩餘
	if res.RemainingHits != wantRemainingHits {
		t.Fatalf("RemainingHits 應為 %d,got %d", wantRemainingHits, res.RemainingHits)
	}
	if got := s.AIPlayers[aiIdx].Colonies[colonyIdx].Population; got != startPop-wantPopulationLost {
		t.Fatalf("殖民地人口應扣減到 %d(建築吸收保護,未被扣光),got %d", startPop-wantPopulationLost, got)
	}
	if got := len(s.AIPlayers[aiIdx].ColonyBuildings[colonyIdx]); got != 0 {
		t.Fatalf("AI 殖民地建築 map 應清空(2 棟皆已摧毀),got %d 棟", got)
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

// --- 2026-07-11:AI 殖民地建築資料模型 + 軌道防禦建築吸收轟炸 新增測試 ---

// TestBombardColony_BuildingAbsorbsBeforePopulation 驗證 hits 不足以打完建築時,建築先吸收、
// 完全不扣人口(demonstrate「軌道防禦」保護人口的核心行為)。1 輪齊射(自訂 RuleProfile 只跑
// 1 輪)、1 艘保證滿傷艦(101 傷害)→ TotalDamage=101 → hits=1。AI 母星有 2 棟建築(海軍陸戰隊
// 營+星基),Profile15 語意(BombardmentBuildingBonusHits=0)每棟耗 1 hit,故這 1 hit 全部
// 用來摧毀 1 棟建築,沒有餘數進人口。
func TestBombardColony_BuildingAbsorbsBeforePopulation(t *testing.T) {
	s, starIdx := newFleetAtAIHomeSession(t)
	s.RaceCombatPct = 0
	s.RuleProfile = gamedata.RuleProfile{BombardmentVolleys: 1, BombardmentBuildingBonusHits: 0}
	s.Ships = []Ship{deterministicBombardShip()}

	aiIdx, colonyIdx, ok := s.findAIColonyByStar(starIdx)
	if !ok {
		t.Fatal("應找得到 AI 母星的殖民地模型")
	}
	startPop := s.AIPlayers[aiIdx].Colonies[colonyIdx].Population

	res := s.BombardColony(starIdx)
	if !res.Ok {
		t.Fatalf("前置條件應齊備,got Reason=%q", res.Reason)
	}
	if res.Hits != 1 {
		t.Fatalf("測試前提錯誤:預期 hits=1(101 傷害/100),got %d", res.Hits)
	}
	if res.BuildingsDestroyed != 1 {
		t.Fatalf("BuildingsDestroyed 應為 1(1 hit 恰好摧毀 1 棟),got %d", res.BuildingsDestroyed)
	}
	if res.BuildingsRemaining != 1 {
		t.Fatalf("BuildingsRemaining 應為 1(2 棟建築摧毀 1 棟,剩 1 棟),got %d", res.BuildingsRemaining)
	}
	if res.PopulationLost != 0 {
		t.Fatalf("hits 全部耗在建築上,PopulationLost 應為 0,got %d", res.PopulationLost)
	}
	if got := s.AIPlayers[aiIdx].Colonies[colonyIdx].Population; got != startPop {
		t.Fatalf("人口不應變動(建築吸收保護),got %d want %d", got, startPop)
	}
}

// TestBombardColony_RemainingHitsAfterBuildingsGoToPopulation 驗證 hits 打完全部建築後,
// 剩餘的 hits 才進人口損傷(順序:建築先、人口後)。4 輪齊射 × 1 艦 × 101 傷害 = 404 傷害 →
// hits=4。2 棟建築各耗 1 hit(Profile15 語意)→ 摧毀 2 棟耗 2 hits,餘 2 hits → popLoss=
// GroundBombardPopulationLoss(2, LARGE_PLANET)=2*6/7(整數除法)=1。
func TestBombardColony_RemainingHitsAfterBuildingsGoToPopulation(t *testing.T) {
	s, starIdx := newFleetAtAIHomeSession(t)
	s.RaceCombatPct = 0
	s.RuleProfile = gamedata.RuleProfile{BombardmentVolleys: 4, BombardmentBuildingBonusHits: 0}
	s.Ships = []Ship{deterministicBombardShip()}

	aiIdx, colonyIdx, ok := s.findAIColonyByStar(starIdx)
	if !ok {
		t.Fatal("應找得到 AI 母星的殖民地模型")
	}

	res := s.BombardColony(starIdx)
	if !res.Ok {
		t.Fatalf("前置條件應齊備,got Reason=%q", res.Reason)
	}
	if res.Hits != 4 {
		t.Fatalf("測試前提錯誤:預期 hits=4(404 傷害/100),got %d", res.Hits)
	}
	if res.BuildingsDestroyed != 2 {
		t.Fatalf("BuildingsDestroyed 應為 2(4 hits 足夠摧毀全部 2 棟),got %d", res.BuildingsDestroyed)
	}
	const wantPopLoss = 1 // 餘 2 hits → GroundBombardPopulationLoss(2, LARGE_PLANET) = 2*6/7 = 1
	if res.PopulationLost != wantPopLoss {
		t.Fatalf("建築摧毀完後餘 hits 應扣人口 %d,got %d", wantPopLoss, res.PopulationLost)
	}
	if got := s.AIPlayers[aiIdx].Colonies[colonyIdx].Population; got != 8-wantPopLoss {
		t.Fatalf("殖民地人口應為 %d,got %d", 8-wantPopLoss, got)
	}
}

// TestBombardColony_BombardmentBuildingBonusHits_VersionDifference 驗證 #7
// (RuleProfile.BombardmentBuildingBonusHits)真的接線影響「同一批 hits 能摧毀幾棟建築」:
// 1.3(bonus=1,每棟耗 2 hits)vs 1.5(bonus=0,每棟耗 1 hit)在完全相同的 hits 輸入下,摧毀
// 的建築數不同——3 輪齊射 × 1 艦 × 101 傷害 = 303 傷害 → hits=3(兩個 profile 的
// BombardmentVolleys 刻意設成相同值,只讓 BombardmentBuildingBonusHits 這個變數不同,分離
// 出單一差異來源)。
func TestBombardColony_BombardmentBuildingBonusHits_VersionDifference(t *testing.T) {
	build := func(bonus int) *GroundBombardResult {
		s, starIdx := newFleetAtAIHomeSession(t)
		s.RaceCombatPct = 0
		s.RuleProfile = gamedata.RuleProfile{BombardmentVolleys: 3, BombardmentBuildingBonusHits: bonus}
		s.Ships = []Ship{deterministicBombardShip()}
		res := s.BombardColony(starIdx)
		return &res
	}
	res15 := build(0) // 1.5:每棟耗 1 hit(GroundPlanetHitsPerBuilding=1 + bonus 0)
	res13 := build(1) // 1.3:每棟耗 2 hits(GroundPlanetHitsPerBuilding=1 + bonus 1)

	if res15.Hits != 3 || res13.Hits != 3 {
		t.Fatalf("測試前提錯誤:兩者 hits 應皆為 3(BombardmentVolleys 相同、只有 bonus 不同),got 1.5=%d 1.3=%d", res15.Hits, res13.Hits)
	}
	if res15.BuildingsDestroyed != 2 {
		t.Fatalf("1.5(bonus=0)hits=3 應摧毀 2 棟建築(3/1=3,但只有 2 棟可摧毀),got %d", res15.BuildingsDestroyed)
	}
	if res13.BuildingsDestroyed != 1 {
		t.Fatalf("1.3(bonus=1)hits=3 每棟耗 2 hits,只夠摧毀 1 棟(3/2=1,餘 1 hit 不夠摧毀第 2 棟),got %d", res13.BuildingsDestroyed)
	}
	if res15.BuildingsDestroyed == res13.BuildingsDestroyed {
		t.Fatalf("同一批 hits(3)在 1.3/1.5 下應摧毀不同棟數的建築,got 皆為 %d(BombardmentBuildingBonusHits 未真正接線?)", res15.BuildingsDestroyed)
	}
}

// TestBombardColony_NilColonyBuildingsRegressionSafe 驗證 ColonyBuildings[colonyIdx]==nil
// (模擬加入本欄位前的舊存檔解碼結果)時,行為與「加建築吸收機制之前」逐位元一致:hits 全部
// 直接進人口損傷,不會 panic、不會誤判成有建築可摧毀。10 輪 × 1 艦 × 101 傷害 = 1010 →
// hits=10 → popLoss=GroundBombardPopulationLoss(10, LARGE_PLANET)=10*6/7(整數除法)=8=扣光
// (與加本子系統前的 TestBombardColony_ReducesPopulationDeterministically 舊斷言完全一致)。
func TestBombardColony_NilColonyBuildingsRegressionSafe(t *testing.T) {
	s, starIdx := newFleetAtAIHomeSession(t)
	s.RaceCombatPct = 0
	s.Ships = []Ship{deterministicBombardShip()}

	aiIdx, colonyIdx, ok := s.findAIColonyByStar(starIdx)
	if !ok {
		t.Fatal("應找得到 AI 母星的殖民地模型")
	}
	s.AIPlayers[aiIdx].ColonyBuildings[colonyIdx] = nil // 模擬舊存檔沒有這個欄位的解碼結果

	res := s.BombardColony(starIdx)
	if !res.Ok {
		t.Fatalf("前置條件應齊備,got Reason=%q", res.Reason)
	}
	if res.BuildingsDestroyed != 0 {
		t.Fatalf("nil 建築 map 不應有任何建築被摧毀,got %d", res.BuildingsDestroyed)
	}
	if res.BuildingsRemaining != 0 {
		t.Fatalf("nil 建築 map 的 BuildingsRemaining 應為 0,got %d", res.BuildingsRemaining)
	}
	if res.PopulationLost != 8 {
		t.Fatalf("nil 建築時 hits 應全部進人口(回歸行為與加此機制前一致),PopulationLost 應為 8,got %d", res.PopulationLost)
	}
	if got := s.AIPlayers[aiIdx].Colonies[colonyIdx].Population; got != 0 {
		t.Fatalf("人口應扣光到 0,got %d", got)
	}
}

// --- 2026-07-11:防禦方反擊(軌道基地/飛彈基地打玩家艦隊)新增測試 ---
//
// 這幾個測試刻意用 RuleProfile{BombardmentVolleys: 0} 讓 fleetBombardDamage 跑 0 輪
// (TotalDamage=0 → hits=0),確保「建築吸收」那段不會消耗任何 hits、不摧毀任何建築——
// 這樣才能直接掌控 aiPlayer.ColonyBuildings[colonyIdx] 的內容,乾淨測試反擊本身的行為,
// 不必和上面「建築吸收」的隨機組合糾纏在一起。

// retaliationTestSetup 建一個「AI 母星只有指定防禦建築、玩家艦隊固定」的轟炸情境,
// 回傳 session 與目標星索引,供以下反擊測試共用。
//
// 2026-07-11(#14,space 預算模型改版)誠實記錄:這裡刻意沿用零值 RuleProfile
// (SatelliteBeamArcCostPct 隨之為 0%),且 AI 母星(newFleetAtAIHomeSession 建出的
// buildDemoAIOpponents)開局 CompletedTopics 沒有任何武器科技主題,retaliationAttackers 內部
// bestUnlockedWeaponValue 因此固定落到雷射 fallback(Value=4,space=10,見該函式註解)。舊版
// (改版前)反擊戰力是固定 tier 4/8/16,新公式在這組「零 arc-cost + 雷射 fallback」輸入下實際
// 算出 5(星基,fit=25*4/20)/10(戰鬥站,fit=50*4/20)/24(星辰要塞,fit=120*4/20)——比舊值
// 略高,但下方各測試斷言檢查的是「是否觸發反擊」「單艦是否被一發擊沉」「星辰要塞反擊是否不弱於
// 星基」這類質性行為,不是斷言精確 atk 數字,重新驗證(go test)後這些質性結果全部維持不變,
// 故不需要改動既有斷言——版本相依(1.3 vs 1.5 arc-cost 不同)與科技效果(解鎖電漿砲後基地
// 變強)兩項新行為的精確數字驗證見 satellite_defense_test.go 直接呼叫 retaliationAttackers
// 的單元測試,不在這裡用 BombardColony 端對端斷言重複驗證。
func retaliationTestSetup(t *testing.T, buildings map[string]bool, ships []Ship) (*GameSession, int, int, int) {
	t.Helper()
	s, starIdx := newFleetAtAIHomeSession(t)
	s.RuleProfile = gamedata.RuleProfile{BombardmentVolleys: 0, BombardmentBuildingBonusHits: 0}
	s.Ships = ships
	aiIdx, colonyIdx, ok := s.findAIColonyByStar(starIdx)
	if !ok {
		t.Fatal("應找得到 AI 母星的殖民地模型")
	}
	s.AIPlayers[aiIdx].ColonyBuildings[colonyIdx] = buildings
	return s, starIdx, aiIdx, colonyIdx
}

// TestBombardColony_RetaliationWithSurvivingStarBase 驗證:AI 母星有存活星基時,轟炸觸發
// DefenderRetaliated=true,且用固定 rng 種子(s.Turn/starIdx 固定)得到確定結果。
func TestBombardColony_RetaliationWithSurvivingStarBase(t *testing.T) {
	s, starIdx, _, _ := retaliationTestSetup(t, map[string]bool{"星基": true}, []Ship{deterministicBombardShip()})
	s.Turn = 3 // 固定種子:此輪對單艦低 HP 目標星基一發命中擊沉(見探測腳本掃過 Turn 0-29 記錄)

	res := s.BombardColony(starIdx)
	if !res.Ok {
		t.Fatalf("前置條件應齊備,got Reason=%q", res.Reason)
	}
	if !res.DefenderRetaliated {
		t.Fatalf("存活星基應觸發反擊,got DefenderRetaliated=false")
	}
	// 只有 1 艘攻方艦(deterministicBombardShip,body=1、hp=3),星基反擊(atk 用驅逐艦級=4)
	// 對固定種子 rng(s.Turn=1, starIdx)解算,結果應可重現;實測見下方斷言。
	if res.AttackerShipsLost != 1 {
		t.Fatalf("固定種子下星基反擊 AttackerShipsLost 應為 1(單艦低 HP,星基 atk=4 一發應可擊沉),got %d", res.AttackerShipsLost)
	}
	if len(s.Ships) != 0 {
		t.Fatalf("唯一一艘攻方艦被反擊擊沉後 s.Ships 應清空,got %d 艘", len(s.Ships))
	}
}

// TestBombardColony_NoRetaliationWithoutDefensiveBuildings 驗證:殖民地無任何防禦建築(本次
// 轟炸把防禦建築全炸掉,或本來就沒有)時,完全不觸發反擊——DefenderRetaliated=false、
// AttackerShipsLost=0,且 s.Ships 逐位元不變(回歸,見設計說明「無存活防禦建築時完全不呼叫
// 反擊解算」)。
func TestBombardColony_NoRetaliationWithoutDefensiveBuildings(t *testing.T) {
	ships := []Ship{deterministicBombardShip(), deterministicBombardShip()}
	s, starIdx, _, _ := retaliationTestSetup(t, map[string]bool{}, ships)
	beforeShips := append([]Ship(nil), s.Ships...)

	res := s.BombardColony(starIdx)
	if !res.Ok {
		t.Fatalf("前置條件應齊備,got Reason=%q", res.Reason)
	}
	if res.DefenderRetaliated {
		t.Fatalf("無防禦建築不應觸發反擊,got DefenderRetaliated=true")
	}
	if res.AttackerShipsLost != 0 {
		t.Fatalf("無反擊時 AttackerShipsLost 應為 0,got %d", res.AttackerShipsLost)
	}
	if len(s.Ships) != len(beforeShips) {
		t.Fatalf("無反擊時 s.Ships 艘數不應變動,got %d want %d", len(s.Ships), len(beforeShips))
	}
	for i := range beforeShips {
		if !reflect.DeepEqual(s.Ships[i], beforeShips[i]) {
			t.Fatalf("無反擊時 s.Ships[%d] 不應變動,got %+v want %+v", i, s.Ships[i], beforeShips[i])
		}
	}
}

// TestBombardColony_StarFortressRetaliationStrongerThanStarBase 驗證基地分級火力遞增序
// (星基 < 戰鬥站 < 星辰要塞,對應手冊「Star Fortress 比 Battlestation 更強」定性描述):
// 同樣的玩家艦隊、同樣的 rng 種子輸入下,星辰要塞反擊擊沉的艦數應 >= 星基反擊擊沉的艦數。
func TestBombardColony_StarFortressRetaliationStrongerThanStarBase(t *testing.T) {
	fleet := func() []Ship {
		return []Ship{
			{Name: "護衛艦一號", Class: "護衛艦", Weapon: "無", Armor: "無裝甲", Shield: "無護盾"},
			{Name: "護衛艦二號", Class: "護衛艦", Weapon: "無", Armor: "無裝甲", Shield: "無護盾"},
			{Name: "護衛艦三號", Class: "護衛艦", Weapon: "無", Armor: "無裝甲", Shield: "無護盾"},
		}
	}
	sBase, idxBase, _, _ := retaliationTestSetup(t, map[string]bool{"星基": true}, fleet())
	sBase.Turn = 5
	sFortress, idxFortress, _, _ := retaliationTestSetup(t, map[string]bool{"星辰要塞": true}, fleet())
	sFortress.Turn = 5

	resBase := sBase.BombardColony(idxBase)
	resFortress := sFortress.BombardColony(idxFortress)
	if !resBase.Ok || !resFortress.Ok {
		t.Fatalf("前置條件應齊備,got base Reason=%q fortress Reason=%q", resBase.Reason, resFortress.Reason)
	}
	if !resBase.DefenderRetaliated || !resFortress.DefenderRetaliated {
		t.Fatalf("兩者皆應觸發反擊,got base=%v fortress=%v", resBase.DefenderRetaliated, resFortress.DefenderRetaliated)
	}
	if resFortress.AttackerShipsLost < resBase.AttackerShipsLost {
		t.Fatalf("星辰要塞(atk=16)反擊應不弱於星基(atk=4),got fortress=%d < base=%d",
			resFortress.AttackerShipsLost, resBase.AttackerShipsLost)
	}
}

// TestBombardColony_RetaliationClearsFleetCargoWhenFleetWiped 驗證:反擊把玩家艦隊整支
// 擊沉時,FleetMarines/FleetTanks 要跟著歸 0(容量隨艦隊歸零,不能留著「艦隊已消失卻還載運
// 陸戰隊/戰車營」的不合理狀態),且不出負數。用星辰要塞(atk=16,最高階)打一艘低 HP 單艦,
// 確保單輪齊射足以擊沉。
func TestBombardColony_RetaliationClearsFleetCargoWhenFleetWiped(t *testing.T) {
	s, starIdx, _, _ := retaliationTestSetup(t, map[string]bool{"星辰要塞": true}, []Ship{deterministicBombardShip()})
	s.Turn = 3
	s.FleetMarines = 5
	s.FleetTanks = 2

	res := s.BombardColony(starIdx)
	if !res.Ok {
		t.Fatalf("前置條件應齊備,got Reason=%q", res.Reason)
	}
	if !res.DefenderRetaliated {
		t.Fatalf("星辰要塞應觸發反擊,got DefenderRetaliated=false")
	}
	if len(s.Ships) != 0 {
		t.Fatalf("測試前提:唯一一艘攻方艦應被星辰要塞反擊擊沉,got %d 艘存活(需要調整測試前提)", len(s.Ships))
	}
	if s.FleetMarines != 0 {
		t.Fatalf("艦隊清空後 FleetMarines 應歸 0,got %d", s.FleetMarines)
	}
	if s.FleetTanks != 0 {
		t.Fatalf("艦隊清空後 FleetTanks 應歸 0,got %d", s.FleetTanks)
	}
}
