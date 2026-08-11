package shell

import (
	"encoding/json"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func TestWeaponArcOptionsAndDefaults(t *testing.T) {
	if got := DefaultWeaponArc("雷射"); got != gamedata.ARC_FWD {
		t.Fatalf("光束武器預設應為前向,得到 %d", got)
	}
	if got := DefaultWeaponArc("核飛彈"); got != gamedata.ARC_360 {
		t.Fatalf("飛彈預設應為 360,得到 %d", got)
	}
	if got := WeaponArcOptionsForWeapon("質子魚雷"); len(got) != 1 || got[0] != gamedata.ARC_360 {
		t.Fatalf("魚雷應只有 360 選項,得到 %#v", got)
	}
	if got := WeaponArcOptionsForWeapon("雷射"); len(got) != 5 {
		t.Fatalf("光束武器應有五種弧,得到 %#v", got)
	}
	if got := NormalizeWeaponArc("核飛彈", gamedata.ARC_FWD); got != gamedata.ARC_360 {
		t.Fatalf("飛彈的無效前向弧應修正成 360,得到 %d", got)
	}
}

func TestCycleWeaponArc(t *testing.T) {
	if got := cycleWeaponArc("雷射", gamedata.ARC_FWD); got != gamedata.ARC_FWD_EXT {
		t.Fatalf("前向後應循環到前向延伸,得到 %d", got)
	}
	if got := cycleWeaponArc("雷射", gamedata.ARC_360); got != gamedata.ARC_FWD {
		t.Fatalf("360 後應循環回前向,得到 %d", got)
	}
	if got := cycleWeaponArc("核飛彈", gamedata.ARC_360); got != gamedata.ARC_360 {
		t.Fatalf("飛彈只有 360,循環後仍應為 360,得到 %d", got)
	}
}

func TestShipDesignArcChangesSpaceAndCost(t *testing.T) {
	weapon := 0
	for i, c := range WeaponOptions {
		if c.Name == "雷射" {
			weapon = i
			break
		}
	}
	baseSpace := ShipDesignSpaceUsedWithModsAndArc("護衛艦", weapon, 0, 0, 0, nil, gamedata.ARC_FWD)
	extSpace := ShipDesignSpaceUsedWithModsAndArc("護衛艦", weapon, 0, 0, 0, nil, gamedata.ARC_FWD_EXT)
	fullSpace := ShipDesignSpaceUsedWithModsAndArc("護衛艦", weapon, 0, 0, 0, nil, gamedata.ARC_360)
	if baseSpace != 10 || extSpace != 12 || fullSpace != 15 {
		t.Fatalf("雷射火線角佔格 = %d/%d/%d, want 10/12/15", baseSpace, extSpace, fullSpace)
	}
	baseCost := DesignCostWithModsAndArc("護衛艦", weapon, 0, 0, 0, nil, gamedata.ARC_FWD)
	extCost := DesignCostWithModsAndArc("護衛艦", weapon, 0, 0, 0, nil, gamedata.ARC_FWD_EXT)
	fullCost := DesignCostWithModsAndArc("護衛艦", weapon, 0, 0, 0, nil, gamedata.ARC_360)
	if extCost <= baseCost || fullCost <= extCost {
		t.Fatalf("火線角成本應遞增: base=%d ext=%d full=%d", baseCost, extCost, fullCost)
	}
}

func TestShipArcJSONRoundTrip(t *testing.T) {
	input := Ship{Name: "弧測試艦", Weapon: "雷射", Arc: gamedata.ARC_BACK_EXT}
	raw, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	var got Ship
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	if got.Arc != input.Arc {
		t.Fatalf("火線角 JSON round-trip = %d, want %d", got.Arc, input.Arc)
	}
}

func TestBuildShipWithArcStoresSelectedArc(t *testing.T) {
	s := NewDemoSession()
	s.Player.BC = 10000
	weapon := 0
	for i, c := range WeaponOptions {
		if c.Name == "雷射" {
			weapon = i
			break
		}
	}
	if !s.BuildShipWithModsAndArc("護衛艦", weapon, 0, 0, 0, nil, gamedata.ARC_BACK_EXT) {
		t.Fatal("應能建造含火線角的護衛艦")
	}
	got := s.Fleet().Ships[len(s.Fleet().Ships)-1]
	if got.Arc != gamedata.ARC_BACK_EXT {
		t.Fatalf("建造後 Ship.Arc=%d, want %d", got.Arc, gamedata.ARC_BACK_EXT)
	}
}

func TestWeaponArcAllowsCombatShotUsesFacingAndGridPosition(t *testing.T) {
	attacker := CombatShip{WeaponName: "雷射", WeaponArc: gamedata.ARC_FWD, Col: 1, Row: 1, Facing: 0}
	front := CombatShip{Col: 2, Row: 1}
	back := CombatShip{Col: 0, Row: 1}
	if !WeaponArcAllowsCombatShot(attacker, front) {
		t.Fatal("朝右的前向武器應能攻擊右側目標")
	}
	if WeaponArcAllowsCombatShot(attacker, back) {
		t.Fatal("朝右的前向武器不應能攻擊左側目標")
	}
	attacker.Facing = 8
	if !WeaponArcAllowsCombatShot(attacker, back) {
		t.Fatal("轉向左後,前向武器應能攻擊左側目標")
	}
}

func TestWeaponArcAllowsCombatShotPreservesAllRoundAndUnknownFixtures(t *testing.T) {
	attacker := CombatShip{WeaponArc: gamedata.ARC_360, Col: 3, Row: 3, Facing: 0}
	if !WeaponArcAllowsCombatShot(attacker, CombatShip{Col: 0, Row: 3}) {
		t.Fatal("360 度武器不應受朝向限制")
	}
	// 測試或舊存檔直接建構、沒有武器名稱與 arc 的 CombatShip 不應突然變成無法開火。
	unknown := CombatShip{Col: 1, Row: 1}
	if !WeaponArcAllowsCombatShot(unknown, CombatShip{Col: 0, Row: 1}) {
		t.Fatal("未知 arc 的舊 CombatShip 應保留相容行為")
	}
}
