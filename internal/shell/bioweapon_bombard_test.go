package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// newBioBombardSession 準備一個「轟炸打不死整個殖民地」的場景。
//
// 生物武器是在一般轟炸傷害**之後**再殺人口,所以人口若已被打到 0,不論有沒有生物武器
// 結果都一樣——那種場景測不出東西。這裡把人口拉高到轟炸打不完,確保轟炸後還有人可殺。
// bioTestPopulation 要大過「20 艘滿傷艦一波轟炸打得掉的人口」,否則人口先被清空,
// 生物武器那一段永遠是 0——測試會綠,但什麼都沒驗到。
const bioTestPopulation = 1000

func newBioBombardSession(t *testing.T, ships int) (*GameSession, int, *engine.ColonyState) {
	t.Helper()
	s, starIdx := newFleetAtAIHomeSession(t)
	s.RaceCombatPct = 0
	s.Fleet().Ships = nil
	for i := 0; i < ships; i++ {
		s.Fleet().Ships = append(s.Fleet().Ships, deterministicBombardShip())
	}
	aiIdx, colonyIdx, ok := s.findAIColonyByStar(starIdx)
	if !ok {
		t.Fatal("應找得到 AI 母星的殖民地模型")
	}
	colony := &s.AIPlayers[aiIdx].Colonies[colonyIdx]
	colony.Population = bioTestPopulation
	return s, starIdx, colony
}

// giveBioWeapon 讓玩家擁有指定的生物武器(走 groundEquipTechOwned 的同一條判定路徑:
// 主題完成 + 明確抉擇選中該科技)。
func giveBioWeapon(t *testing.T, s *GameSession, tech gamedata.Technology) {
	t.Helper()
	topic, ok := gamedata.OrigTechTopic(tech)
	if !ok {
		t.Fatalf("%s 應查得到所屬研究主題", gamedata.TechnologyName(tech))
	}
	if s.Player.ExplicitChoice == nil {
		s.Player.ExplicitChoice = map[gamedata.ResearchTopic]bool{}
	}
	if s.Player.ChosenTech == nil {
		s.Player.ChosenTech = map[gamedata.ResearchTopic]gamedata.Technology{}
	}
	s.Player.CompletedTopics[topic] = true
	s.Player.ExplicitChoice[topic] = true
	s.Player.ChosenTech[topic] = tech
}

// 沒有生物武器就一顆孢子都不投——正對照。
//
// 少了這條,「無條件每次轟炸都額外殺人口」的實作也會讓下面那條通過。
func TestBombard_NoBioWeaponNoExtraKills(t *testing.T) {
	s, starIdx, _ := newBioBombardSession(t, 20)
	res := s.BombardColony(starIdx)
	if !res.Ok {
		t.Fatalf("前置條件應齊備,got Reason=%q", res.Reason)
	}
	if res.BioWeaponKills != 0 {
		t.Errorf("沒有生物武器科技時 BioWeaponKills 應為 0,得到 %d", res.BioWeaponKills)
	}
	if res.PopulationLost >= bioTestPopulation {
		t.Fatalf("測試前提不成立:人口應該還有剩(才測得出生物武器),PopulationLost=%d", res.PopulationLost)
	}
}

// 有生物滅絕者 → 一般轟炸傷害**之外**再殺人口,而且那一份記在 BioWeaponKills 裡。
//
// 兩局種子相同(rng 由 Turn+星索引決定,而生物武器是在所有傷害擲骰之後才擲),
// 所以差額可以直接歸因給生物武器,不是「兩局隨機性不同」。
func TestBombard_BioWeaponKillsOnTopOfBombardment(t *testing.T) {
	base, starIdx, _ := newBioBombardSession(t, 20)
	baseRes := base.BombardColony(starIdx)

	s, _, colony := newBioBombardSession(t, 20)
	giveBioWeapon(t, s, gamedata.TECH_BIOTERMINATOR)
	popBefore := colony.Population
	res := s.BombardColony(starIdx)

	if res.BioWeaponKills <= 0 {
		t.Fatalf("20 莢 × 20%% 應至少殺到一個人口單位,得到 %d", res.BioWeaponKills)
	}
	if res.PopulationLost != baseRes.PopulationLost+res.BioWeaponKills {
		t.Errorf("PopulationLost 應為「一般轟炸 %d + 生物武器 %d」= %d,得到 %d",
			baseRes.PopulationLost, res.BioWeaponKills,
			baseRes.PopulationLost+res.BioWeaponKills, res.PopulationLost)
	}
	if got := popBefore - colony.Population; got != res.PopulationLost {
		t.Errorf("殖民地實際減少的人口(%d)應等於回報的 PopulationLost(%d)", got, res.PopulationLost)
	}
}

// 屏障護盾把生物武器擋成 0,而**其他傷害照常**。
//
// 手冊那句是「cannot enter the planet's atmosphere」——完全擋掉,不是減傷,所以這裡要求
// 恰好 0,不是「比較少」。
func TestBombard_BarrierShieldBlocksBioWeaponEntirely(t *testing.T) {
	s, starIdx, colony := newBioBombardSession(t, 20)
	giveBioWeapon(t, s, gamedata.TECH_BIOTERMINATOR)

	aiIdx, colonyIdx, _ := s.findAIColonyByStar(starIdx)
	s.AIPlayers[aiIdx].ColonyBuildings[colonyIdx][gamedata.BuildingPlanetaryBarrierShield] = true

	res := s.BombardColony(starIdx)
	if res.BioWeaponKills != 0 {
		t.Errorf("屏障護盾應完全擋住生物武器,BioWeaponKills 得到 %d", res.BioWeaponKills)
	}
	// 一般轟炸傷害沒被擋光——否則上面那個 0 只是「根本沒打到」的假陽性。
	if res.PopulationLost <= 0 {
		t.Errorf("屏障護盾只擋生物武器,一般轟炸仍應造成人口損失,得到 %d", res.PopulationLost)
	}
	if colony.Population >= bioTestPopulation {
		t.Errorf("殖民地人口應有實際減少,得到 %d", colony.Population)
	}
}

// 屏障護盾若在**這一次**轟炸中被拆掉,孢子仍然進不去。
//
// 這守的是實作上的一個陷阱：一般傷亡候選池可能在同輪摧毀屏障護盾；若等到孢子那一步
// 才查 buildings，這輪已開始的投放會被錯誤改判成未受阻擋。
func TestBombard_BarrierShieldDestroyedThisTurnStillBlocks(t *testing.T) {
	s, starIdx, _ := newBioBombardSession(t, 20)
	giveBioWeapon(t, s, gamedata.TECH_BIOTERMINATOR)

	aiIdx, colonyIdx, _ := s.findAIColonyByStar(starIdx)
	s.AIPlayers[aiIdx].ColonyBuildings[colonyIdx][gamedata.BuildingPlanetaryBarrierShield] = true

	res := s.BombardColony(starIdx)
	if s.AIPlayers[aiIdx].ColonyBuildings[colonyIdx][gamedata.BuildingPlanetaryBarrierShield] {
		t.Skip("這波轟炸沒把護盾拆掉,這條測試的前提不成立")
	}
	if res.BioWeaponKills != 0 {
		t.Errorf("護盾是這一波才被拆的,本次轟炸的孢子仍該被擋掉,得到 %d", res.BioWeaponKills)
	}
}

// 生物滅絕者(20%)平均殺得比死亡孢子(10%)多。
//
// 單一種子的一局比大小會被隨機性帶著走,所以這裡跑多局(每局換 Turn = 換種子)比總數。
func TestBombard_BioTerminatorOutkillsDeathSpores(t *testing.T) {
	const rounds = 60
	tally := func(tech gamedata.Technology) int {
		total := 0
		for turn := 1; turn <= rounds; turn++ {
			s, starIdx, _ := newBioBombardSession(t, 20)
			s.Turn = turn
			giveBioWeapon(t, s, tech)
			total += s.BombardColony(starIdx).BioWeaponKills
		}
		return total
	}
	spores := tally(gamedata.TECH_DEATH_SPORES)
	terminator := tally(gamedata.TECH_BIOTERMINATOR)
	if spores <= 0 {
		t.Fatalf("%d 局下來死亡孢子應該殺得到人,得到 %d", rounds, spores)
	}
	if terminator <= spores {
		t.Errorf("生物滅絕者(20%%,共 %d)應殺得比死亡孢子(10%%,共 %d)多", terminator, spores)
	}
}
