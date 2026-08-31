package main

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

func TestTacticalOriginalSkeletonRegionsStayInsideCanvas(t *testing.T) {
	if combatControlDeckY != 351 || combatControlDeckY+129 != moo2ScreenH {
		t.Fatalf("COMBAT 控制甲板必須固定在 (0,351,640,129)：y=%d screen=%d", combatControlDeckY, moo2ScreenH)
	}
	if gcX0 < 0 || gcY0 < 0 || gcX0+gcCols*gcCW > moo2ScreenW || gcY0+gcRows*gcCH > combatControlDeckY {
		t.Fatalf("隱形戰術格位不得跨入控制甲板：(%d,%d)-(%d,%d)",
			gcX0, gcY0, gcX0+gcCols*gcCW, gcY0+gcRows*gcCH)
	}
	if tacticalShipInfoY < combatControlDeckY || tacticalSystemsY < combatControlDeckY ||
		tacticalShipInfoX+tacticalShipInfoW > moo2ScreenW || tacticalSystemsX+tacticalSystemsW > moo2ScreenW {
		t.Fatal("選艦／Systems 資訊必須完全位於原版控制甲板")
	}
	if tacticalPortraitX < tacticalShipInfoX || tacticalPortraitX+59 > tacticalShipInfoX+tacticalShipInfoW ||
		tacticalPortraitY < combatControlDeckY || tacticalPortraitY+60 > moo2ScreenH {
		t.Fatal("59×60 CMBTSHP portrait 必須留在原版左側選艦框")
	}
}

func TestCombatPlanetAssetMappingMatchesTrilarOracle(t *testing.T) {
	asset, palette, ok := combatPlanetAssetIndices(gamedata.PlanetClimate(5), gamedata.PlanetSize(2))
	if !ok || asset != 32 || palette != 35 {
		t.Fatalf("Trilar III CMBTPLNT 映射錯誤：asset=%d palette=%d ok=%v", asset, palette, ok)
	}
}

func TestCombatMonsterAssetMappingUsesMonsterArchive(t *testing.T) {
	wants := map[gamedata.SpaceMonster]int{
		gamedata.MonsterGuardian: 7, gamedata.MonsterEel: 8, gamedata.MonsterCrystal: 9,
		gamedata.MonsterAmoeba: 10, gamedata.MonsterHydra: 11, gamedata.MonsterDragon: 12,
	}
	for kind, want := range wants {
		if got, ok := combatMonsterAsset(kind); !ok || got != want {
			t.Fatalf("怪物 %d MONSTER.LBX asset=(%d,%v)，want %d", kind, got, ok, want)
		}
	}
	if _, ok := combatMonsterAsset(gamedata.MonsterNone); ok {
		t.Fatal("MonsterNone 不得誤映射到怪物 sprite")
	}
}

func TestTacticalShipFreeCoordinateOverridesGridOnlyForRendering(t *testing.T) {
	ship := shell.CombatShip{Col: 1, Row: 2, ScreenX: 412, ScreenY: 133, ScreenPositionKnown: true}
	if x, y := tacticalShipScreenCenter(ship); x != 412 || y != 133 {
		t.Fatalf("自由座標未優先：%.0f,%.0f", x, y)
	}
	ship.ScreenPositionKnown = false
	x, y := tacticalShipScreenCenter(ship)
	gx, gy, gw, gh := cellRect(1, 2)
	if x != float64(gx+gw/2) || y != float64(gy+gh/2) {
		t.Fatalf("格位 fallback 錯誤：%.0f,%.0f", x, y)
	}
}

func TestTacticalMinimapRawProjectionMatchesDeployShipsOracle(t *testing.T) {
	tests := []struct {
		x, y         int
		wantX, wantY float32
	}{
		{10, 34, 519, 412.8}, // 行星 raw deployment；圖示圓心另加 y=5
		{21, 35, 532.2, 414}, // Star Base raw deployment；圖示圓心另加 y=3
		{25, 34, 537, 412.8}, // 第一艘 Frigate
		{25, 36, 537, 415.2}, // 第二艘 Frigate
		{55, 34, 573, 412.8}, // Amoeba raw deployment；怪物圖示中心另有 (-2,+3)
	}
	for _, tt := range tests {
		gotX, gotY := tacticalMinimapRawPoint(tt.x, tt.y)
		if gotX != tt.wantX || gotY != tt.wantY {
			t.Fatalf("minimap(%v,%v)=(%v,%v)，want (%v,%v)", tt.x, tt.y, gotX, gotY, tt.wantX, tt.wantY)
		}
	}
}

func TestTacticalOrbitalBaseUsesMediumSelectionRing(t *testing.T) {
	if got := tacticalSelectionRingClass(shell.CombatShip{OrbitalBase: true, SizeClass: gamedata.SHIP_DOOMSTAR}); got != 1 {
		t.Fatalf("Star Base 選艦環 class=%d，want COMBAT#33 對應的 1", got)
	}
	if got := tacticalSelectionRingClass(shell.CombatShip{SizeClass: gamedata.SHIP_BATTLESHIP}); got != 2 {
		t.Fatalf("一般 Battleship 選艦環 class=%d，want 2", got)
	}
}

func TestTacticalMinimapShipPointKeepsRawDeploymentSeparateFromScreenCenter(t *testing.T) {
	base := shell.CombatShip{
		DeployX: 21, DeployY: 35, DeployPositionKnown: true,
		ScreenX: 340, ScreenY: 201, ScreenPositionKnown: true, OrbitalBase: true,
	}
	if x, y := tacticalMinimapShipPoint(base); x != 532.2 || y != 417 {
		t.Fatalf("Star Base raw minimap=(%v,%v)，want (532.2,417)", x, y)
	}
	amoeba := shell.CombatShip{
		DeployX: 55, DeployY: 34, DeployPositionKnown: true,
		MonsterKind: gamedata.MonsterAmoeba,
	}
	if x, y := tacticalMinimapShipPoint(amoeba); x != 571 || y != 415.8 {
		t.Fatalf("Amoeba raw minimap=(%v,%v)，want (571,415.8)", x, y)
	}
}

func TestTacticalTargetSurvivesEnemySliceCompactionByID(t *testing.T) {
	tactical := &tacticalScreen{
		enemy: []shell.CombatShip{
			{TacticalID: 11, HP: 0},
			{TacticalID: 22, HP: 10},
			{TacticalID: 33, HP: 10},
		},
		target: 2,
	}
	tactical.compactEnemyCasualties()
	if tactical.target != 1 || tactical.enemy[tactical.target].TacticalID != 33 {
		t.Fatalf("敵艦切片壓縮後目標必須依 TacticalID 續接：target=%d enemy=%+v", tactical.target, tactical.enemy)
	}

	tactical.enemy[tactical.target].HP = 0
	tactical.compactEnemyCasualties()
	if tactical.target != 0 || tactical.enemy[tactical.target].TacticalID != 22 {
		t.Fatalf("目前目標被擊沉後應安全選第一艘存活敵艦：target=%d enemy=%+v", tactical.target, tactical.enemy)
	}
}

func TestTacticalWeaponRowsRemainInOriginalDeck(t *testing.T) {
	for i := 0; i < 8; i++ {
		r := tacticalWeaponSlotRect(i)
		if r[0] < 108 || r[0]+r[2] > 266 || r[1] < combatControlDeckY || r[1]+r[3] > moo2ScreenH {
			t.Fatalf("武器列 %d 越出原版 WEAPONS／SPECIALS 區：%v", i, r)
		}
		if barButtonHit(r[0]+2, r[1]+r[3]/2) >= 0 {
			t.Fatalf("武器列 %d 與控制鈕熱區重疊：%v", i, r)
		}
	}
}

func TestTacticalFighterLaunchAdapterStaysInSpecialsDeck(t *testing.T) {
	x, y, w, h := launchRect()
	if x < 108 || x+w > 266 || y < combatControlDeckY || y+h > moo2ScreenH {
		t.Fatalf("戰機出擊 adapter 必須留在 SPECIALS 控制甲板：(%d,%d,%d,%d)", x, y, w, h)
	}
	for i := 0; i < 8; i++ {
		r := tacticalWeaponSlotRect(i)
		if x < r[0]+r[2] && x+w > r[0] && y < r[1]+r[3] && y+h > r[1] {
			t.Fatalf("戰機出擊 adapter 與武器列 %d 重疊", i)
		}
	}
}
