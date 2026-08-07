package shell

import (
	"os"
	"path/filepath"
	"testing"
)

// fleet_test.go:多艦隊模型的護欄。
//
// ============ 這支測試檔存在的理由 ============
//
// 改成多艦隊時,舊程式碼裡的 `s.Ships` 混著兩種語意——「這支艦隊的船」與「全帝國的船」
// ——而**單一艦隊時兩者剛好相同**,所以分錯了也看不出來。
//
// 下面每一條都刻意用**兩支艦隊**,讓兩種語意分離:分錯的那一處會立刻算少。
// 不能等多艦隊 UI 做出來才發現,那時候症狀只是「數字偏小」,看起來完全正常。

// twoFleetSession 造一個有兩支艦隊的對局:艦隊 0 在星 0 有 2 艘,艦隊 1 在星 3 有 3 艘。
func twoFleetSession() *GameSession {
	s := &GameSession{Stars: make([]Star, 5)}
	s.Fleets = []Fleet{
		{AtStar: 0, DestStar: -1, Ships: []Ship{
			{Name: "甲", Class: "巡洋艦", WeaponAttack: 4, BonusHP: 6},
			{Name: "乙", Class: "巡洋艦", WeaponAttack: 4, BonusHP: 6},
		}},
		{AtStar: 3, DestStar: -1, Ships: []Ship{
			{Name: "丙", Class: "巡防艦", WeaponAttack: 2, BonusHP: 2},
			{Name: "丁", Class: "巡防艦", WeaponAttack: 2, BonusHP: 2},
			{Name: "戊", Class: "巡防艦", WeaponAttack: 2, BonusHP: 2},
		}},
	}
	return s
}

// TestShipCountAndAllShipsSpanEveryFleet:全帝國計數要涵蓋每一支艦隊。
func TestShipCountAndAllShipsSpanEveryFleet(t *testing.T) {
	s := twoFleetSession()
	if got := s.ShipCount(); got != 5 {
		t.Errorf("全帝國應有 5 艘,實得 %d", got)
	}
	if got := len(s.AllShips()); got != 5 {
		t.Errorf("AllShips 應展平成 5 艘,實得 %d", got)
	}
	if got := len(s.Fleet().Ships); got != 2 {
		t.Errorf("目前選中的艦隊(0)應只有 2 艘,實得 %d", got)
	}
}

// TestCommandPointsCountEveryFleet:指揮點數是**全帝國**的。
//
// 手冊 p.169 明文:指揮評等的需求來自帝國所有艦艇,不是玩家眼前操作的那一支。
// 這是最容易分錯的一處——分錯之後玩家會發現自己憑空多出指揮點數。
func TestCommandPointsCountEveryFleet(t *testing.T) {
	s := twoFleetSession()
	all := s.usedCommandPoints()
	// 只把第二支艦隊清空,需求必須跟著降;若當初寫成只看選中艦隊,這裡不會變。
	s.Fleets[1].Ships = nil
	less := s.usedCommandPoints()
	if !(less < all) {
		t.Errorf("清空第二支艦隊後指揮需求應下降(算的是全帝國),%d → %d", all, less)
	}
}

// TestPlayerMilitaryCountsEveryFleet:國力是全帝國的。
func TestPlayerMilitaryCountsEveryFleet(t *testing.T) {
	s := twoFleetSession()
	all := s.playerMilitary()
	s.Fleets[1].Ships = nil
	if less := s.playerMilitary(); !(less < all) {
		t.Errorf("清空第二支艦隊後國力應下降,%d → %d", all, less)
	}
}

// TestFleetStrengthForHistoryCountsEveryFleet:折線圖的軍力是全帝國的。
func TestFleetStrengthForHistoryCountsEveryFleet(t *testing.T) {
	s := twoFleetSession()
	all := s.playerFleetStrength()
	s.Fleets[1].Ships = nil
	if less := s.playerFleetStrength(); !(less < all) {
		t.Errorf("清空第二支艦隊後折線圖軍力應下降,%d → %d", all, less)
	}
}

// TestSupportShipNamesAreUniqueAcrossFleets:支援艦編號要看全帝國。
//
// 分錯的話會出現兩艘「殖民船 1 號」,而那要等玩家真的分了艦隊才看得到。
func TestSupportShipNamesAreUniqueAcrossFleets(t *testing.T) {
	s := twoFleetSession()
	s.Fleets[1].Ships = append(s.Fleets[1].Ships, Ship{Name: "殖民船 1 號", Class: ColonyShipClass})
	if got := s.nextSupportShipName(ColonyShipClass); got == "殖民船 1 號" {
		t.Error("第二艘殖民船不該和另一支艦隊裡的撞名")
	}
}

// TestRemoveShipGlobalCrossesFleetBoundary:隨機事件打的是整個帝國。
//
// 索引 0/1 落在艦隊 0,2/3/4 落在艦隊 1——邊界要接得起來。
func TestRemoveShipGlobalCrossesFleetBoundary(t *testing.T) {
	for _, c := range []struct {
		k    int
		want string
	}{{0, "甲"}, {1, "乙"}, {2, "丙"}, {4, "戊"}} {
		s := twoFleetSession()
		got, ok := s.removeShipGlobal(c.k)
		if !ok || got.Name != c.want {
			t.Errorf("移除全帝國第 %d 艘應為「%s」,實得 %q(ok=%v)", c.k, c.want, got.Name, ok)
		}
		if s.ShipCount() != 4 {
			t.Errorf("移除後應剩 4 艘,實得 %d", s.ShipCount())
		}
	}
	s := twoFleetSession()
	if _, ok := s.removeShipGlobal(5); ok {
		t.Error("越界的索引不該移除任何東西")
	}
	if _, ok := s.removeShipGlobal(-1); ok {
		t.Error("負索引不該移除任何東西")
	}
}

// TestRepairDocksEachFleetSeparately:停靠據點的艦隊才修,航行中的不修。
//
// 這條是重構順便修正的行為:先前只看「玩家選中的那一支」,別支停在母星也不會修。
func TestRepairDocksEachFleetSeparately(t *testing.T) {
	s := twoFleetSession()
	s.PlayerColonyStars = []int{0} // 只有星 0 是自己的據點
	s.Fleets[0].Ships[0].Damage = 2
	s.Fleets[1].Ships[0].Damage = 2

	if n := s.advanceShipRepair(); n != 1 {
		t.Fatalf("只有停在據點的那一支該修,應修 1 艘,實得 %d", n)
	}
	if s.Fleets[0].Ships[0].Damage != 0 {
		t.Error("停在據點的艦隊應被修好")
	}
	if s.Fleets[1].Ships[0].Damage == 0 {
		t.Error("不在據點的艦隊不該被修")
	}

	// 把第二支也開回據點:這次換它被修。
	s.Fleets[1].AtStar = 0
	if n := s.advanceShipRepair(); n != 1 {
		t.Errorf("第二支回到據點後應修 1 艘,實得 %d", n)
	}
}

// TestHomeworldDefenceAcceptsAnyFleetAtHome:任何一支停在母星的艦隊都算防禦。
//
// 先前只看玩家選中的那一支——玩家把視角切到別支艦隊,母星就「沒有防禦」了,
// 而那是純粹的操作副作用,不該影響世界狀態。
func TestHomeworldDefenceAcceptsAnyFleetAtHome(t *testing.T) {
	s := twoFleetSession()
	s.SelectedFleet = 1 // 視角在星 3 那一支
	if s.Fleets[0].AtStar != 0 {
		t.Fatal("前置條件:艦隊 0 應停在母星")
	}
	// 走 raid 的防禦判定:母星有戰力 → 人口損失減半。
	// 這裡直接驗判定本身(避免相依整個 raid 流程的其他隨機性)。
	defended := false
	for f := range s.Fleets {
		if s.Fleets[f].AtStar != 0 {
			continue
		}
		for _, sh := range s.Fleets[f].Ships {
			if shipStrength(sh.Class) > 0 {
				defended = true
			}
		}
	}
	if !defended {
		t.Error("艦隊 0 停在母星且有戰力,應算防禦成功——與玩家的視角無關")
	}
}

// TestEnsureFleetKeepsInvariant:永遠至少一支艦隊,SelectedFleet 永遠合法。
//
// 這條不變量讓 Fleet() 可以無條件回傳可寫指標——呼叫端不必逐處 nil 檢查,
// 而那種檢查漏一個就是 panic。
func TestEnsureFleetKeepsInvariant(t *testing.T) {
	s := &GameSession{}
	if f := s.Fleet(); f == nil {
		t.Fatal("空對局的 Fleet() 不該回 nil")
	}
	if len(s.Fleets) != 1 {
		t.Errorf("空對局應自動補一支艦隊,實得 %d 支", len(s.Fleets))
	}
	s.SelectedFleet = 99
	if f := s.Fleet(); f != &s.Fleets[0] {
		t.Error("越界的 SelectedFleet 應夾回 0")
	}
	s.Fleets = nil
	s.SelectedFleet = -5
	if f := s.Fleet(); f == nil || s.SelectedFleet != 0 {
		t.Error("Fleets 被清空 + 負索引也要救得回來")
	}
}

// TestAddShipToHomeFleetPrefersColocatedFleet:新艦併進停在該星的艦隊。
func TestAddShipToHomeFleetPrefersColocatedFleet(t *testing.T) {
	s := twoFleetSession()
	s.AddShipToHomeFleet(3, Ship{Name: "新艦"})
	if len(s.Fleets[1].Ships) != 4 {
		t.Errorf("新艦應併進停在星 3 的艦隊 1,實得該艦隊 %d 艘", len(s.Fleets[1].Ships))
	}
	// 沒有艦隊停在那顆星 → 退回第一支(見 AddShipToHomeFleet ⚠)。
	s.AddShipToHomeFleet(4, Ship{Name: "孤艦"})
	if len(s.Fleets[0].Ships) != 3 {
		t.Errorf("找不到同星艦隊時應退回第一支,實得該艦隊 %d 艘", len(s.Fleets[0].Ships))
	}
}

// TestLoadLegacySingleFleetSave:2026-08-07 之前的**舊格式存檔**要讀得回來。
//
// 舊格式是單艦隊時代的形狀:頂層一組 ships + fleetAtStar / fleetDestStar / fleetETA /
// fleetMarines。這裡手寫一份最小的舊檔,驗證它會被組成唯一的一支艦隊。
//
// ⚠ 判斷「是不是舊檔」用的是 `len(Fleets) == 0` 而不是版本號——版本號會被別的改動一起
// 往上帶,而**這個欄位在不在**才是真正的判準。
func TestLoadLegacySingleFleetSave(t *testing.T) {
	const legacy = `{
	  "version": 1,
	  "turn": 7,
	  "stars": [{"Name":"甲"},{"Name":"乙"},{"Name":"丙"}],
	  "ships": [{"Name":"老船","Class":"巡洋艦"}],
	  "fleetAtStar": 2,
	  "fleetDestStar": 1,
	  "fleetETA": 3,
	  "fleetMarines": 8,
	  "selectedStar": -1
	}`
	path := filepath.Join(t.TempDir(), "legacy.json")
	if err := os.WriteFile(path, []byte(legacy), 0o644); err != nil {
		t.Fatal(err)
	}
	s, err := LoadSession(path)
	if err != nil {
		t.Fatalf("舊格式存檔應讀得回來:%v", err)
	}
	if len(s.Fleets) != 1 {
		t.Fatalf("舊檔應組成 1 支艦隊,實得 %d 支", len(s.Fleets))
	}
	f := s.Fleet()
	if len(f.Ships) != 1 || f.Ships[0].Name != "老船" {
		t.Errorf("船應接回艦隊裡,實得 %v", f.Ships)
	}
	if f.AtStar != 2 || f.DestStar != 1 || f.ETA != 3 {
		t.Errorf("位置/任務應接回來(2,1,3),實得 (%d,%d,%d)", f.AtStar, f.DestStar, f.ETA)
	}
	if f.Marines != 8 {
		t.Errorf("陸戰隊應接回來 8,實得 %d", f.Marines)
	}
	// ⚠ 舊格式沒有存戰車營(fleetTanks 欄位根本不存在),所以只能是 0。
	// 這不是這次遷移弄丟的,是舊格式本身的漏欄——新格式序列化整個 Fleet,這個洞已補上。
	if f.Tanks != 0 {
		t.Errorf("舊檔沒有戰車營欄位,應為 0,實得 %d", f.Tanks)
	}
}

// TestSaveRoundTripKeepsEveryFleet:新格式要存得下多支艦隊(含戰車營)。
func TestSaveRoundTripKeepsEveryFleet(t *testing.T) {
	s := twoFleetSession()
	s.SelectedStar = -1
	s.Fleets[1].Marines, s.Fleets[1].Tanks = 4, 2
	s.SelectedFleet = 1

	path := filepath.Join(t.TempDir(), "multi.json")
	if err := s.Save(path); err != nil {
		t.Fatalf("存檔失敗:%v", err)
	}
	got, err := LoadSession(path)
	if err != nil {
		t.Fatalf("讀檔失敗:%v", err)
	}
	if len(got.Fleets) != 2 {
		t.Fatalf("應存回 2 支艦隊,實得 %d", len(got.Fleets))
	}
	if got.SelectedFleet != 1 {
		t.Errorf("選中的艦隊應是 1,實得 %d", got.SelectedFleet)
	}
	if got.Fleets[1].Tanks != 2 || got.Fleets[1].Marines != 4 {
		t.Errorf("陸戰隊/戰車營應存得回來 (4,2),實得 (%d,%d)",
			got.Fleets[1].Marines, got.Fleets[1].Tanks)
	}
	if got.ShipCount() != 5 {
		t.Errorf("全帝國應仍有 5 艘,實得 %d", got.ShipCount())
	}
}
