package shell

import "testing"

// artemisNetSession 建一個「AI 母星有阿提米絲系統網、玩家艦隊正要飛過去」的局面。
func artemisNetSession(t *testing.T, ships []Ship) (*GameSession, int) {
	t.Helper()
	s, starIdx := newFleetAtAIHomeSession(t)
	aiIdx, colonyIdx, ok := s.findAIColonyByStar(starIdx)
	if !ok {
		t.Fatal("應找得到 AI 母星的殖民地模型")
	}
	artemis := testBuildingByRawID(t, 3)
	s.AIPlayers[aiIdx].ColonyBuildings[colonyIdx][artemis.NameZH] = true
	f := s.Fleet()
	f.Ships = ships
	f.AtStar = 0
	f.DestStar = starIdx
	f.ETA = 1
	return s, starIdx
}

func mineTestShip(name, class, shield string) Ship {
	return Ship{Name: name, Class: class, Weapon: "無武裝", Armor: "無裝甲", Shield: shield, Special: "無"}
}

// 末日之星的觸發機率是 100%——一定會踩到。
// 用它當測試主體是因為它把「機率」這個變數消掉了,剩下的都是可驗的算術。
func TestArtemisMines_DoomStarAlwaysTriggers(t *testing.T) {
	s, _ := artemisNetSession(t, []Ship{mineTestShip("末日一號", "末日之星", "無護盾")})
	s.advanceFleet()
	if s.LastArtemis == nil {
		t.Fatal("末日之星觸發機率 100%,應該踩到水雷")
	}
	if s.LastArtemis.ShipsHit != 1 {
		t.Errorf("應有 1 艘中招,得到 %d", s.LastArtemis.ShipsHit)
	}
	// 8–28 枚 × 每枚 20(無護盾)。
	if d := s.LastArtemis.TotalDamage; d < 8*20 || d > 28*20 {
		t.Errorf("傷害 %d 超出手冊範圍 %d–%d", d, 8*20, 28*20)
	}
}

// 正對照:同一顆星沒有水雷網時什麼都不該發生。
// 少了這條,「進任何星系都炸」也會讓上面那條通過。
func TestArtemisMines_NoNetNoStrike(t *testing.T) {
	s, starIdx := artemisNetSession(t, []Ship{mineTestShip("末日一號", "末日之星", "無護盾")})
	aiIdx, colonyIdx, _ := s.findAIColonyByStar(starIdx)
	delete(s.AIPlayers[aiIdx].ColonyBuildings[colonyIdx], testBuildingByRawID(t, 3).NameZH)
	s.advanceFleet()
	if s.LastArtemis != nil {
		t.Fatalf("沒有水雷網不該有結算,得到 %+v", s.LastArtemis)
	}
	if len(s.Fleet().Ships) != 1 || s.Fleet().Ships[0].Damage != 0 {
		t.Error("沒有水雷網時艦艇不該受損")
	}
}

// 護盾直接折抵每一枚的傷害,所以同樣的水雷數下高護盾的船受傷明顯較輕。
// 兩艘船同時進場,踩的是同一組亂數流,比的是「有沒有依護盾折抵」。
func TestArtemisMines_ShieldClassReducesDamage(t *testing.T) {
	bare, shielded := 0, 0
	// 跑多顆星取平均:單一顆星的水雷數是同一條亂數流上的兩個抽樣,
	// 個別比較可能被抽樣差異蓋過——比的是總量。
	for turn := 1; turn <= 20; turn++ {
		s, _ := artemisNetSession(t, []Ship{mineTestShip("裸船", "末日之星", "無護盾")})
		s.Turn = turn
		s.advanceFleet()
		if s.LastArtemis != nil {
			bare += s.LastArtemis.TotalDamage
		}
		s2, _ := artemisNetSession(t, []Ship{mineTestShip("盾船", "末日之星", "第十級護盾")})
		s2.Turn = turn
		s2.advanceFleet()
		if s2.LastArtemis != nil {
			shielded += s2.LastArtemis.TotalDamage
		}
	}
	if bare == 0 || shielded == 0 {
		t.Fatal("兩組都應該有傷害(末日之星必中)")
	}
	// 每枚 20 vs 10 —— 護盾組應該剛好是一半。
	if shielded*2 != bare {
		t.Errorf("第十級護盾應把傷害折半:裸船 %d、盾船 %d(盾船×2 應等於裸船)", bare, shielded)
	}
}

// 巡防艦的 20% 遠低於末日之星的 100%:同樣的局面下中招次數必須明顯較少。
func TestArtemisMines_SmallHullsMostlyGetThrough(t *testing.T) {
	hitFrigate, hitDoom := 0, 0
	const n = 60
	for turn := 1; turn <= n; turn++ {
		s, _ := artemisNetSession(t, []Ship{mineTestShip("小船", "巡防艦", "無護盾")})
		s.Turn = turn
		s.advanceFleet()
		if s.LastArtemis != nil {
			hitFrigate++
		}
		s2, _ := artemisNetSession(t, []Ship{mineTestShip("大船", "末日之星", "無護盾")})
		s2.Turn = turn
		s2.advanceFleet()
		if s2.LastArtemis != nil {
			hitDoom++
		}
	}
	if hitDoom != n {
		t.Errorf("末日之星 100%% 應每次都中,%d 次裡只中 %d 次", n, hitDoom)
	}
	if hitFrigate >= hitDoom {
		t.Errorf("巡防艦 20%% 應明顯少於末日之星:巡防 %d / 末日 %d", hitFrigate, hitDoom)
	}
}

// 被炸沉的船要從艦隊移除,而且結算要記下來。
func TestArtemisMines_DestroyedShipsLeaveTheFleet(t *testing.T) {
	// 巡防艦血量小(shipStrength 2 × 3 = 6),8 枚 × 20 = 160 遠超過 → 必沉。
	s, _ := artemisNetSession(t, []Ship{
		mineTestShip("必沉號", "末日之星", "無護盾"),
	})
	// 把它換成血少的艦體但保留 100% 觸發:直接手動確認「傷害 >= 血量就沉」。
	s.Fleet().Ships[0].Class = "末日之星"
	s.Fleet().Ships[0].Damage = shipMaxHP(s.Fleet().Ships[0]) - 1 // 只剩 1 點血
	s.advanceFleet()
	if s.LastArtemis == nil {
		t.Fatal("應有水雷結算")
	}
	if s.LastArtemis.ShipsLost != 1 {
		t.Fatalf("只剩 1 點血的船挨了水雷應沉,ShipsLost=%d", s.LastArtemis.ShipsLost)
	}
	if len(s.Fleet().Ships) != 0 {
		t.Errorf("沉掉的船應離開艦隊,還剩 %d 艘", len(s.Fleet().Ships))
	}
	if len(s.LastArtemis.LostNames) != 1 || s.LastArtemis.LostNames[0] != "必沉號" {
		t.Errorf("應記下沉掉的船名,得到 %v", s.LastArtemis.LostNames)
	}
}

// 決定性:同一回合、同一顆星,重跑必須得到一模一樣的結果(網路對戰的要求)。
func TestArtemisMines_Deterministic(t *testing.T) {
	run := func() ArtemisStrike {
		s, _ := artemisNetSession(t, []Ship{
			mineTestShip("甲", "戰艦", "第三級護盾"),
			mineTestShip("乙", "泰坦", "無護盾"),
			mineTestShip("丙", "巡洋艦", "第五級護盾"),
		})
		s.Turn = 7
		s.advanceFleet()
		if s.LastArtemis == nil {
			return ArtemisStrike{}
		}
		return *s.LastArtemis
	}
	a, b := run(), run()
	if a.ShipsHit != b.ShipsHit || a.TotalDamage != b.TotalDamage || a.ShipsLost != b.ShipsLost {
		t.Errorf("同樣的輸入應算出同樣的結果:%+v vs %+v", a, b)
	}
}

// 艦體與護盾的對照表本身。
func TestArtemisHullAndShieldMapping(t *testing.T) {
	if artemisShieldClass("第十級護盾") != 10 || artemisShieldClass("第一級護盾") != 1 {
		t.Error("護盾名字裡的「級」就是 shield class")
	}
	if artemisShieldClass("無護盾") != 0 || artemisShieldClass("不存在的護盾") != 0 {
		t.Error("無護盾與未知元件都應是 0——不假裝有防護")
	}
	// 偵察艦/殖民船不是原版的艦體等級,走 Frigate(見 artemis.go 的說明)。
	for _, c := range []string{"偵察艦", "殖民船", "巡防艦", "護衛艦"} {
		if got := artemisHullClass(c); got != 0 {
			t.Errorf("%s 應對到 Frigate(0),得到 %d", c, got)
		}
	}
}
