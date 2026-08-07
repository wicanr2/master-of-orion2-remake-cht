package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// 運力**依艦體等級**,不是每艘一律 4。
//
// 先前是 `len(Ships) * 4`,所以一支偵察艦隊與一支末日之星艦隊的登陸能力完全相同
// ——而手冊 p.121 上兩者差 10 倍。
func TestMarineCapacityScalesWithHullClass(t *testing.T) {
	s := NewDemoSession()
	cases := []struct {
		class string
		want  int
	}{
		{"巡防艦", 5}, {"護衛艦", 5}, {"驅逐艦", 8}, {"巡洋艦", 12},
		{"戰艦", 20}, {"泰坦", 30}, {"末日之星", 50},
	}
	for _, c := range cases {
		s.Fleet().Ships = []Ship{{Name: "測試艦", Class: c.class}}
		if got := s.MarineTransportCapacity(); got != c.want {
			t.Errorf("%s 單艦運力應為 %d(手冊 p.121 Marines 欄),得到 %d", c.class, c.want, got)
		}
	}
}

// 混編艦隊逐艦累加,不是「艦數 × 某個代表值」。
func TestMarineCapacityAddsPerShip(t *testing.T) {
	s := NewDemoSession()
	s.Fleet().Ships = []Ship{
		{Name: "a", Class: "末日之星"}, {Name: "b", Class: "巡防艦"}, {Name: "c", Class: "巡洋艦"},
	}
	want := gamedata.ShipHullMarines(gamedata.SHIP_DOOMSTAR) +
		gamedata.ShipHullMarines(gamedata.SHIP_FRIGATE) +
		gamedata.ShipHullMarines(gamedata.SHIP_CRUISER)
	if got := s.MarineTransportCapacity(); got != want {
		t.Errorf("混編艦隊運力應逐艦累加為 %d,得到 %d", want, got)
	}
}

// 正對照:艦數相同但艦體不同,運力**必須不同**。
//
// 少了這條,「艦數 × 固定值」的舊實作也會讓上面兩條的某些情形通過。
func TestSameShipCountDifferentHullsGiveDifferentCapacity(t *testing.T) {
	s := NewDemoSession()
	s.Fleet().Ships = []Ship{{Name: "a", Class: "巡防艦"}, {Name: "b", Class: "巡防艦"}}
	small := s.MarineTransportCapacity()
	s.Fleet().Ships = []Ship{{Name: "a", Class: "末日之星"}, {Name: "b", Class: "末日之星"}}
	big := s.MarineTransportCapacity()
	if small == big {
		t.Errorf("同樣兩艘但艦體不同,運力不該相同(都得到 %d)", small)
	}
	if big <= small {
		t.Errorf("末日之星的運力應高於巡防艦:%d vs %d", big, small)
	}
}

// 空艦隊回 0,不 panic。
func TestMarineCapacityEmptyFleet(t *testing.T) {
	s := NewDemoSession()
	s.Fleet().Ships = nil
	if got := s.MarineTransportCapacity(); got != 0 {
		t.Errorf("空艦隊運力應為 0,得到 %d", got)
	}
}

// 載運不能超過運力——換了公式之後這條界限仍然守得住。
func TestLoadMarinesStillRespectsCapacity(t *testing.T) {
	s := NewDemoSession()
	s.Fleet().Ships = []Ship{{Name: "小艇", Class: "巡防艦"}} // 運力 5
	s.Fleet().Marines = 0
	// 駐軍池是延遲配置的(advanceMarines 才建),測試自己備一個滿的,
	// 否則這條會 skip——而會 skip 的測試等於沒有守住任何東西。
	s.PlayerColonyMarines = make([]int, len(s.PlayerColonies))
	s.PlayerColonyMarines[0] = 999
	s.LoadMarines(0)
	if got := s.Fleet().Marines; got > s.MarineTransportCapacity() {
		t.Errorf("載運 %d 超過運力 %d", got, s.MarineTransportCapacity())
	}
}
