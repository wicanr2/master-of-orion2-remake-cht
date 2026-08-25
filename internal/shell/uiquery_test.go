package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// 同化進度查得出來,而且**會隨回合遞減**。
//
// 同化 raw 進度的 ETA 必須由 UI 查詢路徑實際消費
// ——「一個只在背景默默跑的機制對玩家等於不存在」。那支函式抽出來之後一直沒有呼叫端。
func TestAssimilationRemainingTurnsCountsDown(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	if _, ok := s.AssimilationRemainingTurns(0); ok {
		t.Fatal("沒有未同化人口時不該回報進度")
	}

	markColonyConquered(&s.PlayerColonies[0], 1)
	first, ok := s.AssimilationRemainingTurns(0)
	if !ok || first <= 0 {
		t.Fatalf("有未同化人口時應回報還需幾回合,得到 %d ok=%v", first, ok)
	}

	s.advanceAssimilation()
	second, ok := s.AssimilationRemainingTurns(0)
	if !ok {
		t.Fatal("同化一回合之後還沒完成,應仍回報進度")
	}
	if second >= first {
		t.Errorf("推進一回合後剩餘回合數應下降:%d → %d", first, second)
	}
}

// 邊界:索引越界不 panic,也不謊報有進度。
func TestAssimilationRemainingTurnsGuards(t *testing.T) {
	s := NewDemoSession()
	for _, i := range []int{-1, len(s.PlayerColonies), 9999} {
		if _, ok := s.AssimilationRemainingTurns(i); ok {
			t.Errorf("索引 %d 不該回報進度", i)
		}
	}
}

// 艦員摘要取**最低**那一艘——艦隊戰力由最弱的那條線決定,報最高的會讓玩家高估自己。
func TestFleetCrewSummaryReportsTheWeakestShip(t *testing.T) {
	s := NewDemoSession()
	ships := s.Fleet().Ships
	if len(ships) < 2 {
		t.Skip("這局的艦隊不足兩艘,測不出最低值")
	}
	for i := range ships {
		ships[i].CrewXP = gamedata.CrewXPForLevel(gamedata.CrewLevelCount-2, s.RaceWarlord())
	}
	ships[len(ships)-1].CrewXP = 0 // 最後一艘是新兵

	lv, _, ok := s.FleetCrewSummary()
	if !ok {
		t.Fatal("有參戰艦時應回報摘要")
	}
	if lv != gamedata.CrewLevelForXP(0, s.RaceWarlord()) {
		t.Errorf("應回報最低的那一艘(新兵),得到等級 %d", lv)
	}
}

// 沒有參戰艦時回 ok=false(支援艦不算)。
func TestFleetCrewSummaryIgnoresSupportShips(t *testing.T) {
	s := NewDemoSession()
	s.Fleet().Ships = []Ship{{Name: "殖民船", Class: ColonyShipClass}}
	if _, _, ok := s.FleetCrewSummary(); ok {
		t.Error("只有支援艦時不該回報艦員摘要")
	}
	s.Fleet().Ships = nil
	if _, _, ok := s.FleetCrewSummary(); ok {
		t.Error("沒有船時不該回報艦員摘要")
	}
}
