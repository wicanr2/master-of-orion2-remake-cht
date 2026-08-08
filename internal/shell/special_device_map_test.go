package shell

import (
	"sort"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// TestSpecialDeviceMapCoverage 把「哪幾項對不上原版特殊裝置表」釘死。
//
// 這條測試的用途不是驗證對照表是對的,是**擋住它安靜地退化**:新增一個特殊系統卻忘了
// 加對照,它會落回 5% 估計佔格 + Component.Cost,而畫面上完全看不出來。
// 名單變了就重看一次原因(見 special_device_map.go 檔頭的兩類),確認之後再改這裡。
func TestSpecialDeviceMapCoverage(t *testing.T) {
	want := []string{
		// ① 原版歸在**武器表**(stride 0x1C),不在特殊裝置表裡。
		// **佔格已經走武器表的真值**(見 specialDeviceSpaceFor 的退路①),
		// 所以這份名單現在的意思是「不在特殊裝置表」,不再等於「落回估計值」。
		"反飛彈火箭", "停滯力場", "牽引光束", "突擊艇", "戰機庫", "轟炸機庫", "重戰機庫",
		// ② 原版兩張表都沒有(電腦在原版是獨立槽,不佔特殊系統位)——只有它還吃 5% 估計。
		"戰鬥電腦",
	}
	var got []string
	for _, c := range SpecialOptions {
		if c.Name == "無" {
			continue
		}
		if _, ok := specialDeviceByName[c.Name]; !ok {
			got = append(got, c.Name)
		}
	}
	sort.Strings(got)
	sort.Strings(want)
	if len(got) != len(want) {
		t.Fatalf("未對上原版表的元件:%v\n期望:%v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("未對上原版表的元件:%v\n期望:%v", got, want)
			break
		}
	}
}

// TestSpecialDeviceMapPointsAtRealRows 對照表裡的每一項在原版表裡都要查得到,
// 而且解鎖科技要與元件清單那一列自己寫的 UnlockTech 相同。
//
// 這是**兩份獨立來源的對撞**:元件清單的 UnlockTech 是先前一項一項從執行檔的 tech→topic
// 表查來的,原版特殊裝置表的科技欄是這一輪才抽的。對不上代表其中一邊抄錯。
func TestSpecialDeviceMapPointsAtRealRows(t *testing.T) {
	for _, c := range SpecialOptions {
		dev, ok := specialDeviceByName[c.Name]
		if !ok {
			continue
		}
		st, ok := gamedata.SpecialDeviceStatsFor(dev)
		if !ok {
			t.Errorf("%s 對到 SPEC=%d,但原版表裡沒有這一筆", c.Name, dev)
			continue
		}
		// 重生程序在 remake 是「抽象(種族特性),proxy 待重設計」,UnlockTech 留 0
		// ——它與自動修復共用科技 23,不是漏抄,略過比對。
		if c.UnlockTech == 0 {
			continue
		}
		if st.Tech != c.UnlockTech {
			t.Errorf("%s 的解鎖科技:元件清單=%d,原版表=%d", c.Name, c.UnlockTech, st.Tech)
		}
	}
}

// TestBattlePodsAddsSpace 戰鬥艙是唯一佔格為負的系統:裝上去之後「已用空間」變小。
//
// 手冊逐字:「Battle Pods are strap-on bays that add equipment space without increasing
// the hull size.」——remake 只有一個特殊系統槽,所以它換來的是**塞得下更大的武器**。
func TestBattlePodsAddsSpace(t *testing.T) {
	pods := -1
	for i, c := range SpecialOptions {
		if c.Name == "戰鬥艙" {
			pods = i
		}
	}
	if pods < 0 {
		t.Fatal("SpecialOptions 裡找不到戰鬥艙")
	}
	// 同一把武器,裝戰鬥艙 vs 不裝任何特殊系統。
	const cl = "戰艦"
	none := ShipDesignSpaceUsed(cl, 1, 0, 0, 0)
	with := ShipDesignSpaceUsed(cl, 1, 0, 0, pods)
	if with >= none {
		t.Errorf("裝了戰鬥艙的已用空間 %d 應該**小於**沒裝的 %d", with, none)
	}
	if want := none - gamedata.ShipHullSpace(gamedata.SHIP_BATTLESHIP)/2; with != want {
		t.Errorf("戰艦裝戰鬥艙的已用空間=%d,期望 %d(少掉艦體空間的一半)", with, want)
	}
}

// TestSpecialCostScalesWithHull 原版的系統成本隨艦體等級變動,不是單一數字
// ——這是接上原版表之後最明顯的行為差異,釘住免得日後重構退回 Component.Cost。
func TestSpecialCostScalesWithHull(t *testing.T) {
	hs := -1
	for i, c := range SpecialOptions {
		if c.Name == "硬化護盾" {
			hs = i
		}
	}
	if hs < 0 {
		t.Fatal("SpecialOptions 裡找不到硬化護盾")
	}
	frigate := DesignCost("巡防艦", 0, 0, 0, hs) - ShipCost("巡防艦")
	doom := DesignCost("末日之星", 0, 0, 0, hs) - ShipCost("末日之星")
	if frigate != 10 || doom != 250 {
		t.Errorf("硬化護盾成本:巡防艦=%d(期望 10)、末日之星=%d(期望 250)", frigate, doom)
	}
}

// TestMegafluxersExpandsHullSpace 巨型通量器把**可用**空間 ×125/100(截斷)。
// 沒研究出來時 HullSpaceFor 應與 gamedata.ShipHullSpace 相同。
func TestMegafluxersExpandsHullSpace(t *testing.T) {
	s := NewDemoSession()
	if got, want := s.HullSpaceFor("巡防艦"), gamedata.ShipHullSpace(gamedata.SHIP_FRIGATE); got != want {
		t.Fatalf("開局(未研究巨型通量器)巡防艦可用空間=%d,期望 %d", got, want)
	}
	if s.Player.CompletedTopics == nil {
		s.Player.CompletedTopics = map[gamedata.ResearchTopic]bool{}
	}
	s.Player.CompletedTopics[gamedata.TOPIC_HIGH_ENERGY_DISTRIBUTION] = true
	if s.Player.ChosenTech == nil {
		s.Player.ChosenTech = map[gamedata.ResearchTopic]gamedata.Technology{}
	}
	s.Player.ChosenTech[gamedata.TOPIC_HIGH_ENERGY_DISTRIBUTION] = gamedata.TECH_MEGAFLUXERS
	if got := s.HullSpaceFor("巡防艦"); got != 31 {
		t.Errorf("研究出巨型通量器後巡防艦可用空間=%d,期望 31(25×125/100 截斷)", got)
	}
	if got := s.HullSpaceFor("末日之星"); got != 1500 {
		t.Errorf("研究出巨型通量器後末日之星可用空間=%d,期望 1500", got)
	}
}
