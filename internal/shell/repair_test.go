package shell

import "testing"

// repair_test.go:艦艇損傷與修復。
// 手冊逐字錨定的部分(20% / 完全修復 / 停在自家據點修好)逐條驗;remake 抽象化的部分只驗界限。

// 損傷會影響下一場戰鬥的血量——先前完全沒有這回事,倖存艦每場都滿血。
func TestShipDamageReducesCombatHP(t *testing.T) {
	s := NewDemoSession()
	s.Ships = []Ship{{Name: "戰艦", Class: "戰艦", Weapon: "死光", Armor: "無裝甲", Shield: "無護盾", Special: "無"}}
	full, _ := s.mkPlayerCombatantsIndexed()
	if len(full) != 1 {
		t.Fatalf("應有 1 艘參戰艦,實得 %d", len(full))
	}
	s.Ships[0].Damage = 5
	hurt, _ := s.mkPlayerCombatantsIndexed()
	if hurt[0].hp != full[0].hp-5 {
		t.Errorf("受損 5 點後血量應 %d,實得 %d", full[0].hp-5, hurt[0].hp)
	}
	// 損傷再重也留 1 點,不會出現「還沒開打就 0 血」。
	s.Ships[0].Damage = 99999
	broken, _ := s.mkPlayerCombatantsIndexed()
	if broken[0].hp < ShipDamageFloorHP {
		t.Errorf("血量下限應為 %d,實得 %d", ShipDamageFloorHP, broken[0].hp)
	}
	// maxHP 要有值,否則戰鬥中的自動修復算不出「已受損多少」。
	if full[0].maxHP <= 0 {
		t.Error("maxHP 未設定,戰鬥中自動修復會失效")
	}
}

// 原版 Repair_Ships_At_Colonies_ → Repair_Ship_Full_:停在自家據點的船**完全修復**。
func TestShipsFullyRepairAtOwnBase(t *testing.T) {
	s := NewDemoSession()
	s.Ships = []Ship{{Name: "戰艦", Class: "戰艦", Damage: 7}}
	s.FleetAtStar, s.FleetETA = s.PlayerColonyStarIndex(0), 0

	if n := s.advanceShipRepair(); n != 1 {
		t.Errorf("停在母星應修好 1 艘,實得 %d", n)
	}
	if s.Ships[0].Damage != 0 {
		t.Errorf("原版是完全修復,不是逐回合慢慢修;剩餘損傷 %d", s.Ships[0].Damage)
	}
}

// 不在自家據點就修不了(航行中、或停在別人的星)。
func TestNoRepairAwayFromBase(t *testing.T) {
	s := NewDemoSession()
	s.Ships = []Ship{{Name: "戰艦", Class: "戰艦", Damage: 7}}

	// 航行中。
	s.FleetAtStar, s.FleetETA = s.PlayerColonyStarIndex(0), 3
	if n := s.advanceShipRepair(); n != 0 {
		t.Errorf("航行中不該修復,實修 %d 艘", n)
	}
	// 停在無主星。
	away := -1
	for i := range s.Stars {
		if !s.starIsPlayerBase(i) {
			away = i
			break
		}
	}
	if away < 0 {
		t.Skip("找不到非玩家據點的星")
	}
	s.FleetAtStar, s.FleetETA = away, 0
	if n := s.advanceShipRepair(); n != 0 {
		t.Errorf("停在非自家據點不該修復,實修 %d 艘", n)
	}
	if s.Ships[0].Damage != 7 {
		t.Errorf("損傷不該變動,實為 %d", s.Ships[0].Damage)
	}
}

// 前哨站也是據點(手冊 p.119:前哨站是艦隊的補給站)。
func TestOutpostCountsAsRepairBase(t *testing.T) {
	s := NewDemoSession()
	target := -1
	for i := range s.Stars {
		if s.Stars[i].Owner == 0 && !s.StarGuardedByMonster(i) && i < len(s.Planets) && !s.Planets[i].NoPlanet {
			target = i
			break
		}
	}
	if target < 0 {
		t.Skip("找不到可建前哨站的星")
	}
	s.Ships = []Ship{{Class: OutpostShipClass}, {Name: "戰艦", Class: "戰艦", Damage: 4}}
	s.FleetAtStar, s.FleetETA = target, 0
	if res := s.BuildOutpost(target); !res.Ok {
		t.Fatalf("建前哨站失敗:%s", res.Reason)
	}
	if n := s.advanceShipRepair(); n != 1 {
		t.Errorf("停在自家前哨站應修好 1 艘,實得 %d", n)
	}
}

// 手冊 p.82:裝有自動修復元件的船「completely repaired after every battle」。
func TestAutoRepairUnitFullyRepairsAfterBattle(t *testing.T) {
	s := NewDemoSession()
	s.Ships = []Ship{
		{Name: "修復艦", Class: "戰艦", Weapon: "死光", Armor: "無裝甲", Shield: "無護盾", Special: "自動修復", Damage: 9},
		{Name: "普通艦", Class: "戰艦", Weapon: "死光", Armor: "無裝甲", Shield: "無護盾", Special: "無", Damage: 9},
	}
	s.repairAfterBattle()
	if s.Ships[0].Damage != 0 {
		t.Errorf("裝自動修復的船戰後應完全修復,剩餘損傷 %d", s.Ships[0].Damage)
	}
	if s.Ships[1].Damage != 9 {
		t.Errorf("沒裝自動修復的船不該被修,損傷 %d → %d", 9, s.Ships[1].Damage)
	}
}

// 手冊 p.82:每個戰鬥回合修復「20% of the ship's armor and structural damage」。
func TestAutoRepairInCombatPercent(t *testing.T) {
	if got := autoRepairInCombat(100); got != 20 {
		t.Errorf("100 點損傷應修 20 點(手冊 20%%),實得 %d", got)
	}
	if got := autoRepairInCombat(0); got != 0 {
		t.Errorf("沒有損傷就不用修,實得 %d", got)
	}
	// 小損傷至少修 1 點,免得整數捨去讓它永遠修不完。
	if got := autoRepairInCombat(3); got != 1 {
		t.Errorf("3 點損傷應至少修 1 點,實得 %d", got)
	}
	// 修復量不會超過損傷本身。
	if got := autoRepairInCombat(2); got > 2 {
		t.Errorf("修復量 %d 超過損傷 2", got)
	}
}

// 打完一場硬仗之後,倖存艦要帶著傷回來(而不是每場都滿血)。
func TestBattleLeavesSurvivorsDamaged(t *testing.T) {
	s := NewDemoSession()
	s.Turn = 40 // 敵方艦隊隨回合變強,確保打得起來
	s.Ships = []Ship{
		{Name: "戰艦一", Class: "戰艦", Weapon: "死光", Armor: "無裝甲", Shield: "無護盾", Special: "無"},
		{Name: "戰艦二", Class: "戰艦", Weapon: "死光", Armor: "無裝甲", Shield: "無護盾", Special: "無"},
		{Name: "戰艦三", Class: "戰艦", Weapon: "死光", Armor: "無裝甲", Shield: "無護盾", Special: "無"},
	}
	// 跑多場(不同回合 = 不同亂數種子)。remake 的戰鬥是「集火第一艘」,所以只有
	// 「戰鬥結束時正被集火的那艘」會帶傷回來——單場可能剛好是 0,多場一定有。
	everDamaged := false
	for turn := 40; turn < 80 && !everDamaged; turn += 3 {
		s.Turn = turn
		for i := range s.Ships {
			s.Ships[i].Damage = 0
		}
		s.ResolveBattle("測試敵人")
		for _, sh := range s.Ships {
			if sh.Damage > 0 {
				everDamaged = true
			}
			if sh.Damage < 0 {
				t.Errorf("%s 的損傷為負:%d", sh.Name, sh.Damage)
			}
			if max := shipMaxHP(sh); sh.Damage > max-ShipDamageFloorHP {
				t.Errorf("%s 的損傷 %d 超過上限 %d", sh.Name, sh.Damage, max-ShipDamageFloorHP)
			}
		}
		if len(s.Ships) == 0 {
			break
		}
	}
	if !everDamaged {
		t.Error("跑了十幾場都沒有任何船帶傷回來——損傷寫回可能沒接上,等於死碼")
	}
}

// 損傷要能存檔往返(否則讀檔後全艦隊自動變成滿血)。
func TestShipDamageSurvivesSaveLoad(t *testing.T) {
	s := NewDemoSession()
	s.Ships = []Ship{{Name: "戰艦", Class: "戰艦", Damage: 6}}
	restored := s.snapshot().restore()
	if len(restored.Ships) != 1 {
		t.Fatalf("讀檔後艦艇數 %d,want 1", len(restored.Ships))
	}
	if restored.Ships[0].Damage != 6 {
		t.Errorf("讀檔後損傷 %d,want 6", restored.Ships[0].Damage)
	}
}

// 損傷百分比是艦隊列表唯一的顯示出口:算錯或沒夾住,玩家看到的就是「損傷 3200%」。
func TestShipDamagePercentClamps(t *testing.T) {
	sh := Ship{Name: "戰艦", Class: "戰艦", Weapon: "死光", Armor: "無裝甲", Shield: "無護盾", Special: "無"}
	if p := ShipDamagePercent(sh); p != 0 {
		t.Errorf("完好的船應為 0%%,實得 %d%%", p)
	}
	max := shipMaxHP(sh)
	sh.Damage = max / 2
	if p := ShipDamagePercent(sh); p < 40 || p > 60 {
		t.Errorf("半血應在 40-60%% 之間,實得 %d%%", p)
	}
	// 損傷值再大也不能超過 100%——ShipDamage 會先夾到「最大血量 − 1」。
	sh.Damage = 99999
	if p := ShipDamagePercent(sh); p >= 100 {
		t.Errorf("損傷百分比應恆 < 100%%(船還活著),實得 %d%%", p)
	}
}
