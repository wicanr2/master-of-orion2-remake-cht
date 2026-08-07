package engine

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// pollutionColony 是一個「產能高到一定會污染」的殖民地:20 個工人 × 3 產能 = 60,
// 中型星的污染容忍值只有 6。
func pollutionColony() ColonyState {
	return ColonyState{
		Population: 10, PopMax: 16,
		Farmers: 0, Workers: 20, Scientists: 0,
		FoodPerFarmer: 2, IndustryPerWorker: 3, ResearchPerScientist: 2,
		FlatFood:        20, // 免得饑荒把成長那一段吃掉,污染這條線才是主角
		PlanetSize:      gamedata.MEDIUM_PLANET,
		PlanetGravity:   gamedata.NORMAL_G,
		MineralRichness: gamedata.ABUNDANT,
	}
}

// 環保官降的是**會產生污染的產能**,不是產能本身:毛工業一點都不能少,
// 少的是清理污染扣掉的那部分。
func TestEnvironmentalistCutsPollutionNotProduction(t *testing.T) {
	base := RunColonyTurn(pollutionColony())

	cs := pollutionColony()
	cs.PollutionReductionPercent = 50
	eco := RunColonyTurn(cs)

	if eco.GrossIndustry != base.GrossIndustry {
		t.Errorf("毛工業不該被環保官改動:%d → %d", base.GrossIndustry, eco.GrossIndustry)
	}
	if eco.PollutingProduction >= base.PollutingProduction {
		t.Errorf("致污染產能應下降:%d → %d", base.PollutingProduction, eco.PollutingProduction)
	}
	if eco.PollutionCleanupCost >= base.PollutionCleanupCost {
		t.Errorf("清理成本應跟著下降:%d → %d", base.PollutionCleanupCost, eco.PollutionCleanupCost)
	}
	if eco.NetIndustry <= base.NetIndustry {
		t.Errorf("淨工業應上升(少扣清理成本):%d → %d", base.NetIndustry, eco.NetIndustry)
	}
	// 淨工業的增加額**恰好**等於清理成本的減少額——沒有別的東西被動到。
	if got, want := eco.NetIndustry-base.NetIndustry,
		base.PollutionCleanupCost-eco.PollutionCleanupCost; got != want {
		t.Errorf("淨工業多出 %d,但清理成本只少 %d——差額代表動到了別的東西", got, want)
	}
}

// 正對照:減幅 0 時一切原封不動。少了這條,「其實根本沒接上」也會讓上面那支通過
// (兩組都跑 base 路徑)。
func TestZeroReductionChangesNothing(t *testing.T) {
	base := RunColonyTurn(pollutionColony())
	cs := pollutionColony()
	cs.PollutionReductionPercent = 0
	if got := RunColonyTurn(cs); got != base {
		t.Errorf("減幅 0 應與沒有環保官完全相同\n有: %+v\n無: %+v", got, base)
	}
}

// 環保官不碰食物/研究/成長——它只在污染那條鏈上。
func TestEnvironmentalistTouchesOnlyPollution(t *testing.T) {
	cs := pollutionColony()
	cs.Farmers, cs.Scientists = 5, 5
	base := RunColonyTurn(cs)

	cs.PollutionReductionPercent = 60
	eco := RunColonyTurn(cs)

	if eco.Food != base.Food {
		t.Errorf("食物不該變:%d → %d", base.Food, eco.Food)
	}
	if eco.Research != base.Research {
		t.Errorf("研究不該變:%d → %d", base.Research, eco.Research)
	}
}

// 手冊 p.90 的算術:污染處理器留 1/2、大氣更新器留 1/4,兩者並存留 1/8 = 1/2 × 1/4。
// 環保官是同一條相乘鏈上的另一個乘數,所以要**疊在查表結果之上**,不是取代它。
func TestEnvironmentalistMultipliesOnTopOfTheBuildings(t *testing.T) {
	withPP := pollutionColony()
	withPP.PollutionProcessor = true
	half := RunColonyTurn(withPP)

	both := withPP
	both.PollutionReductionPercent = 50
	got := RunColonyTurn(both)

	// 60 產能 → 污染處理器留 30 → 環保官 50% → 15
	if want := half.PollutingProduction / 2; got.PollutingProduction != want {
		t.Errorf("污染處理器(%d)再打五折應是 %d,得到 %d",
			half.PollutingProduction, want, got.PollutingProduction)
	}
	// 正對照:沒有污染處理器時是從全額打折,兩者不該相同——否則代表建築被無視了。
	noPP := pollutionColony()
	noPP.PollutionReductionPercent = 50
	if RunColonyTurn(noPP).PollutingProduction == got.PollutingProduction {
		t.Error("有無污染處理器算出同一個致污染產能——建築那一層被吃掉了")
	}
}

// 減幅 100% 時完全沒有污染,清理成本歸零(不是負的)。
func TestFullReductionRemovesPollutionEntirely(t *testing.T) {
	cs := pollutionColony()
	cs.PollutionReductionPercent = 100
	out := RunColonyTurn(cs)
	if out.PollutingProduction != 0 {
		t.Errorf("減幅 100%% 應無致污染產能,得到 %d", out.PollutingProduction)
	}
	if out.PollutionCleanupCost != 0 {
		t.Errorf("清理成本應為 0,得到 %d", out.PollutionCleanupCost)
	}
	if out.NetIndustry != out.GrossIndustry {
		t.Errorf("沒有污染時淨工業應等於毛工業:%d vs %d", out.NetIndustry, out.GrossIndustry)
	}
}

// 容忍寬容種族(矽晶生物)本來就不花產能清污染——環保官對它們沒有可省的東西。
// 這條確認新加的那一步不會意外把「已經是 0」變成別的值。
func TestTolerantRaceIsUnaffected(t *testing.T) {
	cs := pollutionColony()
	cs.TolerantRace = true
	base := RunColonyTurn(cs)
	cs.PollutionReductionPercent = 75
	eco := RunColonyTurn(cs)
	if eco.PollutionCleanupCost != 0 || base.PollutionCleanupCost != 0 {
		t.Errorf("容忍種族清理成本應恆為 0:無環保官 %d、有環保官 %d",
			base.PollutionCleanupCost, eco.PollutionCleanupCost)
	}
	if eco.NetIndustry != base.NetIndustry {
		t.Errorf("淨工業不該變:%d → %d", base.NetIndustry, eco.NetIndustry)
	}
}
