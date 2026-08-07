package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// FighterKind 的檔頭一直寫著「手冊 p.127 的四種」,而底下只有兩個。
//
// 這條測試釘住「註解說幾種、程式碼就有幾種」——先前那種不一致沒有任何測試會抓到。
// (突擊艇仍然沒有,理由寫在型別註解裡:登艦戰機制不存在,加一個不做事的型別只是把洞藏起來。)
func TestBomberKindIsFullyDefined(t *testing.T) {
	cases := []struct {
		kind              FighterKind
		name              string
		speed, hits, shot int
	}{
		{FighterInterceptor, "攔截機", gamedata.CombatFighterBaseSpeedInterceptor,
			gamedata.FighterHitsInterceptor, gamedata.FighterShotsInterceptor},
		{FighterHeavy, "重戰機", gamedata.CombatFighterBaseSpeedHeavyFighter,
			gamedata.FighterHitsHeavyFighter, gamedata.FighterShotsHeavyFighter},
		{FighterBomber, "轟炸機", gamedata.CombatFighterBaseSpeedBomber,
			gamedata.FighterHitsBomber, gamedata.FighterShotsBomber},
	}
	for _, c := range cases {
		if got := FighterKindName(c.kind); got != c.name {
			t.Errorf("型別 %d 的名稱應為 %s,得到 %s", c.kind, c.name, got)
		}
		if got := FighterBaseSpeed(c.kind); got != c.speed {
			t.Errorf("%s 的基礎速度應為 %d,得到 %d", c.name, c.speed, got)
		}
		if got := FighterBaseHits(c.kind); got != c.hits {
			t.Errorf("%s 的血量應為 %d,得到 %d", c.name, c.hits, got)
		}
		if got := FighterShots(c.kind); got != c.shot {
			t.Errorf("%s 的射擊次數應為 %d,得到 %d", c.name, c.shot, got)
		}
	}
}

// 轟炸機的三個數字逐字對手冊。
//
// 手冊:「Bombers … carry **one bomb** … They move at **speed 8** … and can take
// **4 damage** …」——速度 8、血量 4、一次出手。
func TestBomberStatsMatchTheManualProse(t *testing.T) {
	if gamedata.CombatFighterBaseSpeedBomber != 8 {
		t.Errorf("速度應為 8,得到 %d", gamedata.CombatFighterBaseSpeedBomber)
	}
	if gamedata.FighterHitsBomber != 4 {
		t.Errorf("血量應為 4,得到 %d", gamedata.FighterHitsBomber)
	}
	if gamedata.FighterShotsBomber != 1 {
		t.Errorf("射擊次數應為 1,得到 %d", gamedata.FighterShotsBomber)
	}
	// 轟炸機比攔截機慢、比較耐打、出手次數較少——三個關係都該成立。
	if gamedata.CombatFighterBaseSpeedBomber >= gamedata.CombatFighterBaseSpeedInterceptor {
		t.Error("轟炸機應比攔截機慢")
	}
	if gamedata.FighterHitsBomber <= gamedata.FighterHitsInterceptor {
		t.Error("轟炸機應比攔截機耐打")
	}
	if gamedata.FighterShotsBomber >= gamedata.FighterShotsInterceptor {
		t.Error("轟炸機的出手次數應少於攔截機")
	}
}

// 轟炸機庫裝得上,而且兩條戰鬥路徑都認得。
func TestBomberBayIsEquippableAndReachesBothPaths(t *testing.T) {
	var found *Component
	for i := range SpecialOptions {
		if SpecialOptions[i].Name == "轟炸機庫" {
			found = &SpecialOptions[i]
		}
	}
	if found == nil {
		t.Fatal("轟炸機庫應出現在 SpecialOptions")
	}
	topic, ok := gamedata.OrigTechTopic(gamedata.TECH_BOMBER_BAYS)
	if !ok || found.Tech != topic {
		t.Errorf("主題應為執行檔的 %v,得到 %v(ok=%v)", topic, found.Tech, ok)
	}
	if found.UnlockTech != gamedata.TECH_BOMBER_BAYS {
		t.Errorf("解鎖科技對不上:%v", found.UnlockTech)
	}

	s := NewDemoSession()
	s.Fleet().Ships = append(s.Fleet().Ships, Ship{
		Name: "轟炸母艦", Class: "戰艦", Weapon: "雷射砲", Special: "轟炸機庫",
	})
	// 格子戰術:能派出轟炸機隊。
	player, _ := s.StartCombat("測試敵人")
	last := player[len(player)-1]
	if !last.Bay || last.BayKind != FighterBomber {
		t.Errorf("格子戰術應能派出轟炸機隊,得到 Bay=%v Kind=%v", last.Bay, last.BayKind)
	}
	// 快速結算:母艦戰力加成。
	plain := NewDemoSession()
	plain.Fleet().Ships = append(plain.Fleet().Ships, Ship{Name: "普通艦", Class: "戰艦", Weapon: "雷射砲"})
	pc, _ := plain.mkPlayerCombatantsIndexed()
	bc, _ := s.mkPlayerCombatantsIndexed()
	fatk, fhp := gamedata.FighterBomberBayCombatContribution()
	if got := bc[len(bc)-1].atk - pc[len(pc)-1].atk; got != fatk {
		t.Errorf("快速結算的攻擊加成應為 %d,得到 %d", fatk, got)
	}
	if got := bc[len(bc)-1].hp - pc[len(pc)-1].hp; got != fhp {
		t.Errorf("快速結算的 HP 加成應為 %d,得到 %d", fhp, got)
	}
}

// 轟炸機的炸彈**算得進**艦隊戰,而艦載炸彈算不進去——載具不同,規則不同。
func TestBomberBombsCountAgainstShipsButShipBombsDoNot(t *testing.T) {
	fatk, _ := gamedata.FighterBomberBayCombatContribution()
	if fatk <= 0 {
		t.Error("轟炸機對艦攻擊應為正(手冊:can attack either a planet **or a ship**)")
	}
	// 反面:艦載炸彈在艦隊戰完全不開火(第 126 項),那條規則不該被這一項動到。
	if weaponKindByName("核彈") != WeaponKindBomb {
		t.Error("核彈仍應是艦載炸彈類別")
	}
}
