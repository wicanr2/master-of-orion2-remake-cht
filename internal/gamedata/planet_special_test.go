package gamedata

import (
	"math/rand"
	"testing"
)

// planet_special_test.go:行星特殊物產的資料表與效果查詢。
// 期望值全部標明來源(原版權重表 / 手冊逐字 / 反組譯指令),不是從實作反推。

// 權重表總和恰為 100 是「12 欄真的對應 SpecialType 0..11」的獨立佐證——欄數錯開一格就湊不出來。
func TestPlanetSpecialWeightsSumTo100(t *testing.T) {
	sum := 0
	for _, w := range planetSpecialWeights {
		sum += w
	}
	if sum != 100 {
		t.Errorf("權重總和 = %d,預期 100(原版 _planet_special_weighted_chance)", sum)
	}
	// 原版把「無特殊物產」的權重放在索引 0,佔 64%——大部分行星是平凡的。
	if planetSpecialWeights[NoSpecial] != 64 {
		t.Errorf("無特殊物產權重 = %d,預期 64", planetSpecialWeights[NoSpecial])
	}
	// 索引 9(BadSpecial2)與 11(獵戶座)的權重是 0:不由一般骰表產生。
	if planetSpecialWeights[BadSpecial2] != 0 || planetSpecialWeights[OrionSpecial] != 0 {
		t.Error("BadSpecial2 / 獵戶座的出現權重應為 0(不由一般骰表產生)")
	}
}

// RollPlanetSpecial 必須落在合法範圍,且權重 0 的種類永遠不會被骰出來。
func TestRollPlanetSpecialRespectsZeroWeights(t *testing.T) {
	r := rand.New(rand.NewSource(20260806))
	counts := map[PlanetSpecial]int{}
	for i := 0; i < 20000; i++ {
		sp := RollPlanetSpecial(r)
		if sp < NoSpecial || sp > OrionSpecial {
			t.Fatalf("骰出越界值 %d", sp)
		}
		counts[sp]++
	}
	if counts[BadSpecial2] != 0 || counts[OrionSpecial] != 0 {
		t.Errorf("權重 0 的種類被骰出來了:BadSpecial2=%d 獵戶座=%d",
			counts[BadSpecial2], counts[OrionSpecial])
	}
	// 64% 的權重下,2 萬次裡「無」應該佔絕大多數(寬鬆界,只抓權重表接錯的情形)。
	if counts[NoSpecial] < 11000 {
		t.Errorf("無特殊物產只出現 %d/20000,權重表可能沒接上", counts[NoSpecial])
	}
}

// 手冊逐字:寶石礦 +10 BC/回合、金礦 +5 BC/回合。
func TestSpecialIncomePerTurn(t *testing.T) {
	cases := []struct {
		sp   PlanetSpecial
		want int
	}{
		{GemDeposits, 10}, {GoldDeposits, 5},
		{Natives, 0}, {NoSpecial, 0}, {AncientArtifacts, 0},
	}
	for _, c := range cases {
		if got := SpecialIncomePerTurn(c.sp); got != c.want {
			t.Errorf("SpecialIncomePerTurn(%d) = %d,want %d", c.sp, got, c.want)
		}
	}
}

// 反組譯 Do_System_Discoveries_At_Star_:太空殘骸 add 0x32(50)、海盜藏寶 add 0x64(100)。
// 手冊只說「is added to your treasury」不給金額,這兩個數字只有反組譯給得出來。
func TestSpecialDiscoveryBC(t *testing.T) {
	if got := SpecialDiscoveryBC(SpaceDebris); got != 50 {
		t.Errorf("太空殘骸 = %d BC,want 50", got)
	}
	if got := SpecialDiscoveryBC(PirateCache); got != 100 {
		t.Errorf("海盜藏寶 = %d BC,want 100", got)
	}
	// 海盜藏寶正好是太空殘骸的兩倍——與 AI 估值加分(金礦 1280 / 寶石礦 2560)同款的整數倍關係。
	if SpecialDiscoveryBC(PirateCache) != 2*SpecialDiscoveryBC(SpaceDebris) {
		t.Error("海盜藏寶應為太空殘骸的兩倍")
	}
	for _, sp := range []PlanetSpecial{NoSpecial, GoldDeposits, Natives, SplinterColony} {
		if got := SpecialDiscoveryBC(sp); got != 0 {
			t.Errorf("SpecialDiscoveryBC(%d) = %d,want 0", sp, got)
		}
	}
}

// 反組譯 Make_New_Colony_Or_Outpost_:原住民分支寫 3 個額外人口單位,且只有它會被消耗。
func TestSpecialColonizeEffects(t *testing.T) {
	if NativePopulationUnits != 3 {
		t.Errorf("NativePopulationUnits = %d,want 3(原版 colony+0x10..0x1C,stride 4)", NativePopulationUnits)
	}
	if got := SpecialExtraPopulationOnColonize(Natives); got != NativePopulationUnits {
		t.Errorf("原住民額外人口 = %d,want %d", got, NativePopulationUnits)
	}
	for _, sp := range []PlanetSpecial{NoSpecial, GoldDeposits, SplinterColony, AncientArtifacts} {
		if got := SpecialExtraPopulationOnColonize(sp); got != 0 {
			t.Errorf("SpecialExtraPopulationOnColonize(%d) = %d,want 0", sp, got)
		}
		if SpecialConsumedOnColonize(sp) {
			t.Errorf("special %d 不該在殖民後被消耗(原版只清原住民)", sp)
		}
	}
	if !SpecialConsumedOnColonize(Natives) {
		t.Error("原住民應在殖民後消耗(原版 [planet+0Fh]=0)")
	}
}

// 五種「抵達星系觸發」的分類要與原版 dispatch 的 case 一致(2/3/7/8/10),不多不少。
func TestSpecialIsSystemDiscovery(t *testing.T) {
	want := map[PlanetSpecial]bool{
		SpaceDebris: true, PirateCache: true, SplinterColony: true,
		LostHero: true, AncientArtifacts: true,
	}
	for sp := NoSpecial; sp <= OrionSpecial; sp++ {
		if got := SpecialIsSystemDiscovery(sp); got != want[sp] {
			t.Errorf("SpecialIsSystemDiscovery(%d) = %v,want %v", sp, got, want[sp])
		}
	}
}

// 手冊逐字:原住民 +2 食物/農夫、遠古文物 5 研究/科學家(取代基準 3)。
func TestSpecialProductionEffects(t *testing.T) {
	if got := SpecialFoodPerFarmerBonus(Natives); got != 2 {
		t.Errorf("原住民食物加成 = %d,want 2", got)
	}
	if got := SpecialFoodPerFarmerBonus(GoldDeposits); got != 0 {
		t.Errorf("金礦不該有食物加成,得 %d", got)
	}
	if got := SpecialResearchPerScientist(AncientArtifacts); got != 5 {
		t.Errorf("遠古文物研究 = %d,want 5", got)
	}
	if got := SpecialResearchPerScientist(AncientArtifacts); got <= ResearchPerScientistNorm {
		t.Errorf("遠古文物(%d)應高於一般基準(%d)", got, ResearchPerScientistNorm)
	}
}

// AI 估值的特殊物產加分只判 4/5/10/11 四個 case(原版硬編),且金礦:寶石礦 = 1:2,
// 與手冊的 5:10 BC 相符——兩個獨立來源對同一組相對關係一致。
func TestAISpecialBonusMatchesManualRatio(t *testing.T) {
	gold, gem := aiSpecialBonus(int(GoldDeposits)), aiSpecialBonus(int(GemDeposits))
	if gold != 1280 || gem != 2560 {
		t.Fatalf("AI 加分 金礦=%d 寶石=%d,want 1280/2560", gold, gem)
	}
	if gem != 2*gold {
		t.Error("AI 加分的金礦:寶石比應為 1:2")
	}
	if SpecialIncomePerTurn(GemDeposits) != 2*SpecialIncomePerTurn(GoldDeposits) {
		t.Error("手冊收入的金礦:寶石比應為 1:2,與 AI 加分同比")
	}
}

// 遠古文物送幾項:原版 `Random_(4)/4 + 1`,而 Random_ 回 1..n(見 planet_special.go 訂正說明)。
// roll 1-3 → 1 項、roll 4 → 2 項,也就是 25% 的機率拿到 2 項。
func TestArtifactFreeTechCount(t *testing.T) {
	want := map[int]int{1: 1, 2: 1, 3: 1, 4: 2}
	for roll, w := range want {
		if got := ArtifactFreeTechCount(roll); got != w {
			t.Errorf("ArtifactFreeTechCount(%d) = %d,want %d", roll, got, w)
		}
	}
	// 越界的 roll 夾回合法範圍,不回奇怪的數字。
	if got := ArtifactFreeTechCount(0); got != 1 {
		t.Errorf("roll 0 應夾成 1 → 1 項,實得 %d", got)
	}
	if got := ArtifactFreeTechCount(99); got != 2 {
		t.Errorf("roll 99 應夾成 4 → 2 項,實得 %d", got)
	}
}
