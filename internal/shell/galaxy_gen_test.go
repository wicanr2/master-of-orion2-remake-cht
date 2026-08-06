package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// TestGalaxyGenerationDistribution 是星圖生成的分布回歸測試。
//
// 為什麼要有:2026-08-06 以前 remake 用 rand.Intn(7) 均勻擲光譜、把光譜當氣候索引,
// 星圖的統計性質與原版差很遠。換成原版骰表(gamedata/galaxygen.go)之後,這個測試把
// 「換回均勻亂數」「表被讀反」這類 regression 擋下來——單看單一星系看不出分布壞掉。
//
// 各項的期望值來自原版權重乘上「非母星佔比」:24 星中 4 顆是母星(玩家 1 + AI 3),
// 母星強制黃星,所以其餘光譜的實測比例約為原版權重的 20/24。
func TestGalaxyGenerationDistribution(t *testing.T) {
	const runs, starsPerRun = 120, 24
	total := runs * starsPerRun

	spec := map[int]int{}
	clim := map[gamedata.PlanetClimate]int{}
	farmable := 0
	for seed := 0; seed < runs; seed++ {
		s := NewDemoSession()
		s.SetupNewGame(starsPerRun, int64(seed), 3)
		for i, st := range s.Stars {
			spec[st.Spectral]++
			p := s.Planets[i]
			if p.NoPlanet {
				continue
			}
			clim[p.ClimateID]++
			if gamedata.ClimateFoodPerFarmer(p.ClimateID) > 0 {
				farmable++
			}
		}
	}

	pct := func(n int) float64 { return float64(n) / float64(total) * 100 }

	// 非母星光譜:實測應落在「原版權重 × 20/24」附近(容忍 ±3 個百分點的抽樣誤差)。
	nonHomeShare := 20.0 / 24.0
	for _, sc := range []gamedata.SpectralClass{gamedata.Blue, gamedata.White, gamedata.Red, gamedata.BlackHole} {
		want := float64(gamedata.StarClassWeights[sc][gamedata.GalaxyAverage]) * nonHomeShare
		if got := pct(spec[int(sc)]); got < want-3 || got > want+3 {
			t.Errorf("光譜 %d 比例 %.1f%%,預期 %.1f%%±3(原版權重 %d%% × 非母星佔比)",
				sc, got, want, gamedata.StarClassWeights[sc][gamedata.GalaxyAverage])
		}
	}

	// 宜居星必須是少數但不能絕跡——原版的核心張力就是「好行星要搶」。
	if p := pct(farmable); p < 25 || p > 45 {
		t.Errorf("可農作行星比例 %.1f%% 落在合理區間 25–45%% 之外", p)
	}
	// Gaia 在一般星系只有宜居帶 2% 的權重,整體應該遠低於 2%。
	if p := pct(clim[gamedata.GAIA]); p > 2 {
		t.Errorf("蓋亞行星比例 %.1f%% 過高(原版只在宜居帶有 2%% 權重)", p)
	}
	// 十種氣候都要出得來:舊生成器只產得出 7 種,少掉的正是玩家最在意的那幾種。
	for c := gamedata.TOXIC; c <= gamedata.GAIA; c++ {
		if clim[c] == 0 {
			t.Errorf("氣候 %s 在 %d 顆星中一次都沒出現", climateDisplayName(c), total)
		}
	}
}

// TestHomeStarsAlwaysHabitable 驗證玩家與 AI 母星所在的星一定生得出可農作行星
// (原版母星恆為宜居世界;交給機率骰會出現 Toxic 母星)。
func TestHomeStarsAlwaysHabitable(t *testing.T) {
	for seed := 0; seed < 50; seed++ {
		s := NewDemoSession()
		s.SetupNewGame(24, int64(seed), 3)
		homes := []int{0}
		for _, ai := range s.AIPlayers {
			homes = append(homes, ai.ColonyStars...)
		}
		for _, idx := range homes {
			if idx < 0 || idx >= len(s.Planets) {
				continue
			}
			p := s.Planets[idx]
			if p.NoPlanet {
				t.Fatalf("seed %d:母星 %d 沒有行星", seed, idx)
			}
			if gamedata.ClimateFoodPerFarmer(p.ClimateID) <= 0 {
				t.Fatalf("seed %d:母星 %d 的氣候 %s 無法農作", seed, idx, p.Climate)
			}
		}
	}
}

// TestOldSavePlanetsBackfillIDs 驗證 2026-08-06 之前的存檔(只有顯示字串、Gen=0)
// 載入後會被回填成 enum,拓殖不會因為讀不到 ID 而走錯路徑。
func TestOldSavePlanetsBackfillIDs(t *testing.T) {
	old := []Planet{
		{Name: "舊星 I", Climate: "海洋", Gravity: "低", Mineral: "豐富", Size: "大型"},
		{Name: "舊星 II", Climate: "地獄", Gravity: "高", Mineral: "貧瘠", Size: "小型"},
	}
	restorePlanetIDs(old)
	if old[0].Gen != planetGenVersion {
		t.Errorf("回填後 Gen 應為 %d,got %d", planetGenVersion, old[0].Gen)
	}
	if old[0].ClimateID != gamedata.OCEAN || old[0].GravityID != gamedata.LOW_G ||
		old[0].MineralID != gamedata.RICH || old[0].SizeID != gamedata.LARGE_PLANET {
		t.Errorf("第一顆行星回填錯誤:%+v", old[0])
	}
	// 「地獄」是舊生成器對黑洞星系的填充詞,應回填成 TOXIC(見 climateDisplayToGamedata)。
	if old[1].ClimateID != gamedata.TOXIC {
		t.Errorf("「地獄」應回填成 TOXIC,got %v", old[1].ClimateID)
	}
}
