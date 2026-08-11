package shell

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func TestGalaxyKeepsEnglishStarNameAlongsideLocalizedDisplayName(t *testing.T) {
	localized, _ := genGalaxy(24, 77, 3, galaxyAgeSetting, func(string) string { return "中文星名" })
	english, _ := genGalaxy(24, 77, 3, galaxyAgeSetting, nil)
	for i := range localized {
		if localized[i].Name == "" || localized[i].NameEN == "" {
			t.Fatalf("星 %d 缺少顯示名或英文原名:%+v", i, localized[i])
		}
		if localized[i].Name != "中文星名" {
			t.Errorf("星 %d 顯示名未套用注入翻譯器:%q", i, localized[i].Name)
		}
		if localized[i].NameEN != english[i].Name || english[i].NameEN != english[i].Name {
			t.Errorf("星 %d 英文原名不穩定: localized=%q english=%q", i, localized[i].NameEN, english[i].Name)
		}
	}
}

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
			// ⚠ 不能寫 `s.Planets[idx]` —— Planets 自 2026-08-07 起不再與 Stars 平行。
			pi := s.PlanetAt(idx)
			if pi < 0 {
				continue
			}
			p := s.Planets[pi]
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

// --- 行星類別(Gen 2:_orbit_to_satellite_type)---

// 母星恆為一般行星:交給類別骰有機會生出「母星是氣態巨星」這種開局就死的局面。
func TestGenPlanetsHomeStarsAlwaysHabitable(t *testing.T) {
	for seed := int64(0); seed < 20; seed++ {
		galaxy, aiHomes := genGalaxy(24, seed, 3, galaxyAgeSetting, nil)
		homes := demoHomeStarSet(aiHomes)
		ps := genPlanets(galaxy, rand.New(rand.NewSource(seed+1)), rand.New(rand.NewSource(seed+5)), galaxyAgeSetting, homes)
		for i := range galaxy {
			if !homes[i] {
				continue
			}
			pi := representativePlanet(galaxy, ps, i)
			if pi < 0 {
				t.Fatalf("seed %d:母星 %d 挑不到代表行星", seed, i)
			}
			if ps[pi].TypeID != gamedata.HABITABLE {
				t.Fatalf("seed %d:母星 %d 的類別是 %d,應為一般行星", seed, i, ps[pi].TypeID)
			}
			if ps[pi].NoPlanet {
				t.Fatalf("seed %d:母星 %d 沒有行星", seed, i)
			}
		}
	}
}

// 代表行星的挑法:系統裡只要有一般行星,代表行星就必須是它(對齊原版「一個系統有一顆可殖民
// 的行星就能殖民」),而不是隨機挑到氣態巨星就整顆星報銷。
func TestGenPlanetsPrefersHabitableRepresentative(t *testing.T) {
	nonHab, withHabSibling := 0, 0
	for seed := int64(0); seed < 40; seed++ {
		galaxy, aiHomes := genGalaxy(24, seed, 3, galaxyAgeSetting, nil)
		ps := genPlanets(galaxy, rand.New(rand.NewSource(seed+1)), rand.New(rand.NewSource(seed+5)), galaxyAgeSetting, demoHomeStarSet(aiHomes))
		for _, p := range ps {
			if p.NoPlanet || p.TypeID == gamedata.HABITABLE {
				continue
			}
			nonHab++
			for _, b := range p.SystemBodies {
				if b.Type == gamedata.HABITABLE {
					withHabSibling++
					break
				}
			}
		}
	}
	if nonHab == 0 {
		t.Fatal("40 個星圖裡一顆不可殖民的星都沒有,類別骰可能沒接上")
	}
	if withHabSibling != 0 {
		t.Errorf("有 %d 顆星的代表行星不可殖民、但同系其實有一般行星(挑法錯了)", withHabSibling)
	}
	t.Logf("不可殖民的星共 %d 顆(佔 40×24 = 960 顆星)", nonHab)
}

// 同系天體不包含代表行星本身(避免兩份資料要同步),且軌道不重複。
func TestGenPlanetsSystemBodiesExcludeRepresentative(t *testing.T) {
	for seed := int64(0); seed < 20; seed++ {
		galaxy, aiHomes := genGalaxy(24, seed, 3, galaxyAgeSetting, nil)
		ps := genPlanets(galaxy, rand.New(rand.NewSource(seed+1)), rand.New(rand.NewSource(seed+5)), galaxyAgeSetting, demoHomeStarSet(aiHomes))
		for i, p := range ps {
			seen := map[int]bool{p.Orbit: true}
			for _, b := range p.SystemBodies {
				if b.Orbit == p.Orbit {
					t.Fatalf("seed %d 星 %d:同系天體與代表行星同軌道 %d", seed, i, b.Orbit)
				}
				if seen[b.Orbit] {
					t.Fatalf("seed %d 星 %d:軌道 %d 重複", seed, i, b.Orbit)
				}
				seen[b.Orbit] = true
				if b.Orbit < 0 || b.Orbit > 4 {
					t.Fatalf("seed %d 星 %d:軌道 %d 越界", seed, i, b.Orbit)
				}
			}
		}
	}
}

// 氣態巨星/小行星帶不能殖民(手冊 p.55「colonies can only survive on a solid planet」),
// 而且拒絕的理由要講得出是哪一類,不是一句「不支援」。
func TestColonizeRejectsNonHabitable(t *testing.T) {
	for _, tp := range []gamedata.PlanetType{gamedata.GAS_GIANT, gamedata.ASTEROIDS} {
		s := NewDemoSession()
		target := -1
		for i := range s.Stars {
			if s.Stars[i].Owner == 0 && i < len(s.Planets) {
				target = i
				break
			}
		}
		s.Planets[target] = Planet{
			Name: "測試星 I", Gen: planetGenVersion, TypeID: tp,
			ClimateID: gamedata.TERRAN, GravityID: gamedata.NORMAL_G,
			MineralID: gamedata.ABUNDANT, SizeID: gamedata.MEDIUM_PLANET,
		}
		s.Fleet().Ships = append(s.Fleet().Ships, Ship{Class: ColonyShipClass})
		s.Fleet().AtStar, s.Fleet().ETA = target, 0

		res := s.ColonizeStar(target)
		if res.Ok {
			t.Errorf("類別 %d 不該可以殖民", tp)
		}
		if want := planetTypeDisplayName(tp); !strings.Contains(res.Reason, want) {
			t.Errorf("類別 %d 的拒絕理由 %q 應含 %q", tp, res.Reason, want)
		}
		// 被拒絕時不該消耗殖民船,也不該改變星的歸屬。
		if s.findColonyShipIndex() < 0 {
			t.Errorf("類別 %d:拒絕拓殖卻消耗了殖民船", tp)
		}
		if s.Stars[target].Owner != 0 {
			t.Errorf("類別 %d:拒絕拓殖卻改了星的歸屬", tp)
		}
	}
}

// AI 也不該把氣態巨星/小行星帶排進拓殖目標(原版 AIPlanetValue 開頭就是 type != HABITABLE → 0)。
func TestAIPlanetValueZeroForNonHabitable(t *testing.T) {
	s := NewDemoSession()
	target := -1
	for i := range s.Stars {
		if s.Stars[i].Owner == 0 && i < len(s.Planets) {
			target = i
			break
		}
	}
	base := Planet{
		Name: "測試星 I", Gen: planetGenVersion, TypeID: gamedata.HABITABLE,
		ClimateID: gamedata.TERRAN, GravityID: gamedata.NORMAL_G,
		MineralID: gamedata.ABUNDANT, SizeID: gamedata.LARGE_PLANET,
	}
	s.Planets[target] = base
	if got := s.aiPlanetValue(0, target); got <= 0 {
		t.Fatalf("一般行星的 AI 估值應 > 0,實得 %d", got)
	}
	for _, tp := range []gamedata.PlanetType{gamedata.GAS_GIANT, gamedata.ASTEROIDS} {
		p := base
		p.TypeID = tp
		s.Planets[target] = p
		if got := s.aiPlanetValue(0, target); got != 0 {
			t.Errorf("類別 %d 的 AI 估值應為 0,實得 %d", tp, got)
		}
	}
}
