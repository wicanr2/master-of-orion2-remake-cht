package shell

import (
	"math/rand"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// fixedRoll 回傳一個永遠擲出同一個值的 GroundRoll(用比例夾在合法範圍內)。
func fixedRoll(v int) gamedata.GroundRoll {
	return func(n int) int {
		if n <= 0 {
			return 0
		}
		if v >= n {
			return n - 1
		}
		return v
	}
}

// TestBoardingCaptureKillsEntireCrew 手冊:「their object is to kill off the entire
// defending crew. If they are successful, ownership of that ship changes.」
func TestBoardingCaptureKillsEntireCrew(t *testing.T) {
	// 攻方戰力壓倒性領先 → 守軍應該全滅、判定奪船。
	p := BoardingParty{Intent: BoardingCapture, Marines: 8, Strength: 500, HitsToKill: 2}
	d := BoardingDefense{Marines: 5, Strength: 0, HitsToKill: 2}
	res := ResolveBoarding(p, d, fixedRoll(0))
	if !res.Captured {
		t.Fatalf("壓倒性優勢應該奪船成功:%+v", res)
	}
	if res.DefenderSurvived != 0 {
		t.Errorf("奪船成功時守軍應該全滅,實剩 %d", res.DefenderSurvived)
	}
	if res.AttackerSurvived <= 0 {
		t.Errorf("攻方應該有倖存者,實剩 %d", res.AttackerSurvived)
	}
}

// TestBoardingSecurityStationsHelpDefender 保安站的 +20 要真的讓守方變強。
//
// 用同一組雙方、同一顆種子跑兩次,只差保安站——這樣測的是**那個加成**,
// 不是「守方本來就贏」。
func TestBoardingSecurityStationsHelpDefender(t *testing.T) {
	run := func(sec bool) BoardingResult {
		rng := rand.New(rand.NewSource(20260808))
		p := BoardingParty{Intent: BoardingCapture, Marines: 12, Strength: 30, HitsToKill: 2}
		d := BoardingDefense{Marines: 12, Strength: 30, HitsToKill: 2, SecurityStations: sec}
		return ResolveBoarding(p, d, func(n int) int {
			if n <= 0 {
				return 0
			}
			return rng.Intn(n)
		})
	}
	plain, withSec := run(false), run(true)
	if withSec.DefenderSurvived <= plain.DefenderSurvived {
		t.Errorf("裝了保安站的守軍倖存 %d,應該多於沒裝的 %d(+%d 戰力)",
			withSec.DefenderSurvived, plain.DefenderSurvived,
			gamedata.ShipSecurityStationsDefenseBonus)
	}
}

// TestBoardingRaidDestroysSystemsNotStructure 手冊:「Marines cannot do damage directly to
// the armor or structure of a target ship.」——突襲只算系統,而且結果裡沒有任何傷害欄位
// 可以拿去扣血,這條測的是「突襲真的有拆到東西」。
func TestBoardingRaidDestroysSystemsNotStructure(t *testing.T) {
	p := BoardingParty{Intent: BoardingRaid, Marines: 8, Strength: 500, HitsToKill: 2}
	d := BoardingDefense{Marines: 5, Strength: 0, HitsToKill: 2}
	res := ResolveBoarding(p, d, fixedRoll(0))
	if res.SystemsDestroyed < gamedata.BoardingRaidSystemsMin {
		t.Errorf("突襲應該至少拆掉 %d 個系統,實得 %d", gamedata.BoardingRaidSystemsMin, res.SystemsDestroyed)
	}
	if res.Captured {
		t.Error("突襲不該判定成奪船——那是另一個 Intent")
	}
	// 奪船那一側同樣的兵力對比不該產生 SystemsDestroyed。
	p.Intent = BoardingCapture
	if cap := ResolveBoarding(p, d, fixedRoll(0)); cap.SystemsDestroyed != 0 {
		t.Errorf("奪船不該拆系統,實得 %d", cap.SystemsDestroyed)
	}
}

// TestBoardingNoMarinesIsNoOp 沒有人可以派就什麼都不發生(而且不能 panic)。
func TestBoardingNoMarinesIsNoOp(t *testing.T) {
	res := ResolveBoarding(BoardingParty{Marines: 0}, BoardingDefense{Marines: 7}, fixedRoll(0))
	if res.Captured || res.Rounds != 0 || res.DefenderSurvived != 7 {
		t.Errorf("零登艦隊應該什麼都不做:%+v", res)
	}
}

// TestShipMarineComplement 艦上陸戰隊 = 手冊 p.121 的 Marines 欄,部隊艙翻倍。
func TestShipMarineComplement(t *testing.T) {
	cases := []struct {
		class string
		pods  bool
		want  int
	}{
		{"巡防艦", false, 5},
		{"戰艦", false, 20},
		{"末日之星", false, 50},
		{"戰艦", true, 40}, // 部隊艙:手冊「doubling the number of Marines on board a ship」
	}
	for _, c := range cases {
		sh := Ship{Class: c.class}
		if c.pods {
			sh.Special = "部隊艙"
		}
		if got := ShipMarineComplement(sh); got != c.want {
			t.Errorf("%s(部隊艙=%v)的陸戰隊=%d,期望 %d", c.class, c.pods, got, c.want)
		}
	}
}

// TestBoardingComponentsReachCombatShip 兩個新元件要真的變成 CombatShip 上的旗標
// ——「元件表有」不等於「效果有接」(第 72 項(元件表有≠效果有接))。
func TestBoardingComponentsReachCombatShip(t *testing.T) {
	for _, name := range []string{"突擊艇", "保安站", "傳送器"} {
		s := NewDemoSession()
		s.Fleet().Ships = append(s.Fleet().Ships, Ship{
			Name: "測試艦", Class: "戰艦", Weapon: "雷射砲", Special: name})
		player, _ := s.StartCombat("測試敵人")
		last := player[len(player)-1]
		if last.Marines != 20 {
			t.Errorf("戰艦的艦上陸戰隊=%d,期望 20", last.Marines)
		}
		switch name {
		case "突擊艇":
			if !last.AssaultShuttles {
				t.Error("突擊艇沒有接到格子戰術的 CombatShip 上")
			}
		case "保安站":
			if !last.SecurityStations {
				t.Error("保安站沒有接到格子戰術的 CombatShip 上")
			}
		case "傳送器":
			if !last.Transporters {
				t.Error("傳送器沒有接到格子戰術的 CombatShip 上")
			}
		}
		cs, _ := s.mkPlayerCombatantsIndexed()
		c := cs[len(cs)-1]
		if c.marines != 20 {
			t.Errorf("快速結算的 combatant 陸戰隊=%d,期望 20", c.marines)
		}
		if name == "突擊艇" && !c.assaultShuttles {
			t.Error("突擊艇沒有接到快速結算的 combatant 上")
		}
		if name == "保安站" && !c.securityStations {
			t.Error("保安站沒有接到快速結算的 combatant 上")
		}
	}
}

// TestAssaultShuttleFighterStats 突擊艇的戰機三欄要是手冊值:速度 6、血量 3、不開火。
func TestAssaultShuttleFighterStats(t *testing.T) {
	if got := FighterBaseSpeed(FighterAssaultShuttle); got != gamedata.AssaultShuttleBaseSpeed {
		t.Errorf("突擊艇基礎速度=%d,期望 %d", got, gamedata.AssaultShuttleBaseSpeed)
	}
	if got := FighterBaseHits(FighterAssaultShuttle); got != gamedata.AssaultShuttleBaseHits {
		t.Errorf("突擊艇基礎血量=%d,期望 %d", got, gamedata.AssaultShuttleBaseHits)
	}
	if got := FighterShots(FighterAssaultShuttle); got != 0 {
		t.Errorf("突擊艇不開火,射擊次數應為 0,實得 %d", got)
	}
	if got := FighterKindName(FighterAssaultShuttle); got != "突擊艇" {
		t.Errorf("型別名=%q,期望「突擊艇」", got)
	}
}

// TestQuickBoardingHappensOnce 快速結算的登艦一場一次(手冊:突擊艇放完人就漂在那裡)。
func TestQuickBoardingHappensOnce(t *testing.T) {
	atk := []combatant{{hp: 100, atk: 300, def: 0, wmin: 1, wmax: 1, kind: WeaponKindBeam,
		shots: 1, assaultShuttles: true}}
	def := []combatant{{hp: 1000, atk: 0, def: 0, marines: 20, shipIdx: -1}}
	rng := rand.New(rand.NewSource(1))
	battleShot(&atk[0], &def, rng)
	afterFirst := def[0].marines
	if afterFirst >= 20 {
		t.Fatalf("第一次登艦應該讓守軍減員,實剩 %d", afterFirst)
	}
	if !atk[0].boarded {
		t.Fatal("登艦之後 boarded 應為真")
	}
	battleShot(&atk[0], &def, rng)
	if def[0].marines != afterFirst && def[0].marines > 0 {
		t.Errorf("第二次不該再登艦一次:%d → %d", afterFirst, def[0].marines)
	}
}

// TestShipBoardingReach 突擊艇不必貼身,沒有的必須相鄰。
func TestShipBoardingReach(t *testing.T) {
	shuttle := CombatShip{AssaultShuttles: true}
	plain := CombatShip{}
	for _, d := range []int{1, 3, 8} {
		if !ShipBoardingReach(shuttle, d) {
			t.Errorf("有突擊艇在 %d 格應該登得了", d)
		}
	}
	if !ShipBoardingReach(plain, 1) {
		t.Error("沒有突擊艇但貼身,應該登得了")
	}
	if ShipBoardingReach(plain, 2) {
		t.Error("沒有突擊艇又不貼身,不該登得了")
	}
}

// TestTransporterReachRequiresDisabledFacing 對照手冊「12 squares — if the shield
// facing the attacking ship is disabled」與 Hard Shields 的例外。
func TestTransporterReachRequiresDisabledFacing(t *testing.T) {
	att := CombatShip{Name: "傳送艦", Transporters: true, Col: 0, Row: 0}
	def := CombatShip{Name: "目標", ShieldReduction: gamedata.DamageShieldReductionClassI,
		SizeClass: gamedata.SHIP_FRIGATE, Col: 6, Row: 0}
	facing := ShieldFacingForShot(att, def)
	def.EnsureShieldFacings()
	if ShipBoardingReachAgainst(att, def, gamedata.TransporterRangeSquares) {
		t.Fatal("護盾分面尚未擊穿時,12 格傳送器不應可登艦")
	}
	if got := def.ShieldFacingHP[facing]; got != gamedata.DamageShieldCapacityForShipClass(
		gamedata.DamageShieldReductionClassI, gamedata.SHIP_FRIGATE) {
		t.Fatalf("初始分面容量=%d,與手冊容量不符", got)
	}
	def.ApplyShieldDamage(facing, gamedata.DamageShieldCapacityForShipClass(
		gamedata.DamageShieldReductionClassI, gamedata.SHIP_FRIGATE)-1)
	if ShipBoardingReachAgainst(att, def, gamedata.TransporterRangeSquares) {
		t.Fatal("分面尚餘 1 點時不應可傳送")
	}
	def.ApplyShieldDamage(facing, 1)
	if !ShipBoardingReachAgainst(att, def, gamedata.TransporterRangeSquares) {
		t.Fatal("面向攻擊方的分面歸零後,12 格內應可傳送")
	}
	if ShipBoardingReachAgainst(att, def, gamedata.TransporterRangeSquares+1) {
		t.Fatal("超過 12 格不應可傳送")
	}
	hard := def
	hard.HardShield = true
	if ShipBoardingReachAgainst(att, hard, gamedata.TransporterRangeSquares) {
		t.Fatal("硬化護盾即使分面歸零也應阻擋傳送器")
	}
	noShield := def
	noShield.ShieldReduction = 0
	noShield.ShieldFacingsInitialized = false
	if !ShipBoardingReachAgainst(att, noShield, 6) {
		t.Fatal("無護盾目標的指定分面應視為失效")
	}
}

// TestShipBoardingPartySize 突擊艇一次只送一個中隊(4 人),貼身登艦傾巢而出。
func TestShipBoardingPartySize(t *testing.T) {
	if got := ShipBoardingPartySize(CombatShip{Marines: 20, AssaultShuttles: true}); got != 4 {
		t.Errorf("突擊艇一次送 %d 個單位,期望 4(中隊 4 架 × 每架 1 人)", got)
	}
	if got := ShipBoardingPartySize(CombatShip{Marines: 3, AssaultShuttles: true}); got != 3 {
		t.Errorf("艦上只有 3 人時送 %d,不該超過實際人數", got)
	}
	if got := ShipBoardingPartySize(CombatShip{Marines: 20}); got != 20 {
		t.Errorf("貼身登艦送 %d,期望全員 20", got)
	}
}

// TestShipBoardingAttackUpdatesBothSides 解算要就地改雙方人數,奪船要立旗標。
//
// 用壓倒性兵力對零守軍,結果是確定的——這測的是**接線**(誰的欄位有被改到),
// 不是機率(那由 ResolveBoarding 自己的測試涵蓋)。
func TestShipBoardingAttackUpdatesBothSides(t *testing.T) {
	s := NewDemoSession()
	att := CombatShip{Name: "我艦", Marines: 20}
	def := CombatShip{Name: "敵艦", Marines: 1}
	res := s.ShipBoardingAttack(&att, &def, BoardingCapture, fixedRoll(99))
	if !res.Captured || !def.Captured {
		t.Fatalf("20 打 1 應該奪船:res=%+v def.Captured=%v", res, def.Captured)
	}
	if def.Marines != 0 {
		t.Errorf("奪船後守軍應歸零,實得 %d", def.Marines)
	}
	if att.Marines > 20 {
		t.Errorf("攻方人數只會減不會增,實得 %d", att.Marines)
	}
}

// TestShipBoardingAttackShuttlesAreOneWay 突擊艇送出去的人不會回來。
func TestShipBoardingAttackShuttlesAreOneWay(t *testing.T) {
	s := NewDemoSession()
	att := CombatShip{Name: "我艦", Marines: 20, AssaultShuttles: true}
	def := CombatShip{Name: "敵艦", Marines: 1}
	s.ShipBoardingAttack(&att, &def, BoardingCapture, fixedRoll(99))
	if att.Marines != 16 {
		t.Errorf("突擊艇送出一個中隊(4)之後艦上應剩 16,實得 %d", att.Marines)
	}
}

// TestHomeworldShipNamesAreEnglishOriginals 開局那三艘的名字要存**英文原文**,
// 中文由翻譯器在建艦當下產生——否則英文模式的艦隊畫面會永遠掛著中文艦名。
//
// 這是第 84 項(名稱池雙語化)的同一條原則:名稱是資料,語言是呈現。
// (放在 boarding_test.go 是因為登艦戰是這一輪同時做的;shell 沒有專屬的命名測試檔。)
func TestHomeworldShipNamesAreEnglishOriginals(t *testing.T) {
	raw := homeworldShips(nil) // nil 翻譯器 = 拿到未翻譯的原文
	want := []string{"Pathfinder", "Vanguard I", "Vanguard II"}
	for i, sh := range raw {
		if sh.Name != want[i] {
			t.Errorf("第 %d 艘的原文名 = %q,期望 %q", i, sh.Name, want[i])
		}
	}
	// 翻譯器接上去之後要真的變中文(這條才是「有沒有接線」)。
	zh := homeworldShips(func(en string) string {
		return map[string]string{"Pathfinder": "拓荒號", "Vanguard I": "先驅一號", "Vanguard II": "先驅二號"}[en]
	})
	if zh[0].Name != "拓荒號" {
		t.Errorf("接上翻譯器後第 0 艘 = %q,期望「拓荒號」", zh[0].Name)
	}
}
