package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// TestAntaresRaidsScheduleAndEscalate 釘住一般科技等級的第一個資源 pulse；原版不是
// 固定第 20 回合、每 15 回合直接扣 BC／人口。
func TestAntaresRaidsScheduleAndEscalate(t *testing.T) {
	s := NewDemoSession()
	s.TechLevel, s.TechLevelSet = TechLevelDefault, true
	for s.Turn = 1; s.Turn <= 124; s.Turn++ {
		s.advanceAntares()
	}
	if s.AntaranInvasion.OffensiveResource != 0 || s.AntaranInvasion.DefensiveResource != 0 {
		t.Fatalf("一般科技第 124 回合前不應有資源：%+v", s.AntaranInvasion)
	}
	s.Turn = 126 // elapsed=125：一般科技延遲 100 後的第一個 25 回合 pulse
	s.advanceAntares()
	if !s.AntaranInvasion.Initialized {
		t.Fatal("安塔蘭全局狀態未初始化")
	}
	if got := gamedata.OriginalAntaranWeightedStrength(s.AntaranInvasion.OffensiveShips, s.AntaranInvasion.Costs) +
		s.AntaranInvasion.OffensiveResource; got == 0 {
		t.Fatalf("第一個 pulse 應已投入攻擊資源或建艦：%+v", s.AntaranInvasion)
	}
}

// TestAntaresDefenseReducesDamage 保留舊測試名稱以便歷史搜尋；現行斷言是抵達後必須
// 進快速戰鬥，不能沿用舊腳本直接扣國庫。
func TestAntaresDefenseReducesDamage(t *testing.T) {
	s := NewDemoSession()
	beforeBC := s.Player.BC
	raid := AntaranRaidFleet{TargetKind: eventEmpirePlayer.String(), StarIndex: 0,
		Ships: [5]int{1, 0, 0, 0, 0}}
	s.resolveAntaranRaid(raid)
	if s.Player.BC != beforeBC {
		t.Fatalf("安塔蘭戰鬥不得沿用舊腳本直接扣 BC：%d → %d", beforeBC, s.Player.BC)
	}
	if s.LastAntaranNotice == nil || s.LastAntaranNotice.Kind != AntaranNoticeBattle {
		t.Fatalf("抵達戰鬥應有型別化報告：%+v", s.LastAntaranNotice)
	}
}

func TestAntaranInvasionStateSaveRoundTrip(t *testing.T) {
	s := NewDemoSession()
	s.AntaranInvasion = AntaranInvasionState{
		Initialized: true, OffensiveResource: 7, DefensiveResource: 3,
		OffensiveShips: [5]int{1, 2, 3, 0, 1}, DeployedShips: [5]int{0, 1, 1, 0, 1},
		OffensiveMax: gamedata.AntaranOffensiveMax, DefensiveMax: gamedata.AntaranDefensiveMax,
		Costs: gamedata.AntaranShipCosts, Readiness: 9,
		Pending: []AntaranRaidFleet{{TargetKind: "ai", TargetIndex: 0, StarIndex: 1, ETA: 2,
			Ships: [5]int{0, 0, 1, 0, 1}}},
	}
	path := t.TempDir() + "/antaran.json"
	if err := s.Save(path); err != nil {
		t.Fatal(err)
	}
	got, err := LoadSession(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.AntaranInvasion.OffensiveResource != 7 || got.AntaranInvasion.Readiness != 9 ||
		len(got.AntaranInvasion.Pending) != 1 || got.AntaranInvasion.Pending[0].ETA != 2 {
		t.Fatalf("安塔蘭全局 record 往返遺失：%+v", got.AntaranInvasion)
	}
}

func TestHotseatIdleSeatDoesNotAdvanceGlobalAntarans(t *testing.T) {
	s := NewDemoSession()
	if len(s.AIPlayers) < 1 || s.SetupHotseat(2) != 2 {
		t.Skip("需要第二個熱座帝國")
	}
	s.initAntaranInvasionState()
	s.AntaranInvasion.OffensiveResource = 17
	s.AntaranInvasion.Readiness = 4
	s.advanceSeatEmpire()
	if s.AntaranInvasion.OffensiveResource != 17 || s.AntaranInvasion.Readiness != 4 {
		t.Fatalf("席位私有結算不得重跑全局安塔蘭狀態：%+v", s.AntaranInvasion)
	}
}

func TestAntaranLaunchDeploysAtMostFiveShips(t *testing.T) {
	s := NewDemoSession()
	s.AntaranInvasion = AntaranInvasionState{
		Initialized: true, OffensiveShips: gamedata.AntaranOffensiveMax,
		OffensiveMax: gamedata.AntaranOffensiveMax, DefensiveMax: gamedata.AntaranDefensiveMax,
		Costs: gamedata.AntaranShipCosts, Readiness: 199,
	}
	if len(s.AIPlayers) == 0 || len(s.AIPlayers[0].Colonies) == 0 {
		t.Skip("需要人口不同的第二帝國作目標")
	}
	s.AIPlayers[0].Colonies[0].Population = s.PlayerColonies[0].Population + 20
	s.Turn = 2
	s.advanceAntares()
	if len(s.AntaranInvasion.Pending) != 1 {
		t.Fatalf("readiness=199 應必定部署一支艦隊：%+v", s.AntaranInvasion)
	}
	n := 0
	for _, count := range s.AntaranInvasion.Pending[0].Ships {
		n += count
	}
	if n < 1 || n > 5 {
		t.Fatalf("原版每次部署 1..5 艘，got %d：%v", n, s.AntaranInvasion.Pending[0].Ships)
	}
}

func TestAntaranRaidCanTargetAI(t *testing.T) {
	s := NewDemoSession()
	if len(s.AIPlayers) == 0 || len(s.AIPlayers[0].ColonyStars) == 0 {
		t.Skip("沒有 AI")
	}
	s.initAntaranInvasionState()
	s.AIPlayers[0].FleetStrength = 100
	before := s.AIPlayers[0].FleetStrength
	s.resolveAntaranRaid(AntaranRaidFleet{TargetKind: eventEmpireAI.String(), TargetIndex: 0,
		StarIndex: s.AIPlayers[0].ColonyStars[0], Ships: [5]int{0, 0, 1, 0, 0}})
	if s.AIPlayers[0].FleetStrength >= before {
		t.Fatalf("AI 目標應承受安塔蘭艦隊戰鬥損耗：%d → %d", before, s.AIPlayers[0].FleetStrength)
	}
	if s.LastAntaranNotice == nil || s.LastAntaranNotice.Kind != AntaranNoticeAIEngaged {
		t.Fatalf("AI 戰鬥應有型別化報告：%+v", s.LastAntaranNotice)
	}
}

func TestAntaranNoticeFollowsHotseatTarget(t *testing.T) {
	s := NewDemoSession()
	if len(s.AIPlayers) < 1 || s.SetupHotseat(2) != 2 {
		t.Skip("需要第二個熱座帝國")
	}
	want := &AntaranNotice{Kind: AntaranNoticeLaunched, StarName: "測試星", StarNameEN: "Test", ETA: 3}
	s.LastAntaranNotice = want
	s.Seats[s.ActiveSeat] = s.saveSeat()
	s.loadSeat(s.Seats[s.ActiveSeat])
	if s.LastAntaranNotice == nil || *s.LastAntaranNotice != *want {
		t.Fatalf("安塔蘭通知未隨熱座席位往返：%+v", s.LastAntaranNotice)
	}
}

func TestDestroyedAntaranRaidShipsReturnToNeitherPool(t *testing.T) {
	s := NewDemoSession()
	s.initAntaranInvasionState()
	s.AntaranInvasion.OffensiveShips[0] = 1
	s.AntaranInvasion.DeployedShips[0] = 1
	s.AntaranInvasion.Pending = []AntaranRaidFleet{{
		TargetKind: eventEmpirePlayer.String(), StarIndex: s.Fleet().AtStar, ETA: 1,
		Ships: [5]int{1, 0, 0, 0, 0},
	}}
	// 確定性地讓守軍第一輪摧毀最低階安塔蘭艦；這條只驗證戰後資源池回寫，
	// 不把目前的 combatant 戰力近似誤稱為原版精確 blueprint。
	for i := range s.Fleet().Ships {
		s.Fleet().Ships[i].WeaponAttack = 10000
	}
	s.advanceAntaranRaidFleets()
	if got := s.AntaranInvasion.OffensiveShips[0]; got != 0 {
		t.Fatalf("已摧毀的安塔蘭艦不得回到 offensive pool：%d", got)
	}
	if got := s.AntaranInvasion.DeployedShips[0]; got != 0 {
		t.Fatalf("已摧毀的安塔蘭艦不得繼續算 deployed：%d", got)
	}
	if len(s.AntaranInvasion.Pending) != 0 {
		t.Fatalf("已抵達的 pending fleet 應完成消費：%+v", s.AntaranInvasion.Pending)
	}
}

func TestUndefendedAntaranRaidShipsReturnToOffensivePool(t *testing.T) {
	s := NewDemoSession()
	s.initAntaranInvasionState()
	s.AntaranInvasion.OffensiveShips[0] = 1
	s.AntaranInvasion.DeployedShips[0] = 1
	s.AntaranInvasion.Pending = []AntaranRaidFleet{{
		TargetKind: eventEmpirePlayer.String(), StarIndex: -1, ETA: 1,
		Ships: [5]int{1, 0, 0, 0, 0},
	}}
	s.advanceAntaranRaidFleets()
	if s.AntaranInvasion.OffensiveShips[0] != 1 || s.AntaranInvasion.DeployedShips[0] != 0 {
		t.Fatalf("未被摧毀的攻方應回到 offensive pool 並離開 deployed 子集：%+v", s.AntaranInvasion)
	}
}

// 安塔蘭母星防禦艦隊的組成是反組譯真值(`_n_max_antaran_def_ships` = {0,0,3,2,7,0,0,0,0}),
// 不再是「6 艘同級」的保守預設。
//
// 這一支釘住**數量與分層**(那是真值);艦體尺寸→remake 戰力階梯的對照是推論,
// 所以只驗相對關係,不驗絕對數字。
func TestAntaranHomeFleetMatchesTheDisassembledComposition(t *testing.T) {
	fleet := antaranHomeFleetDefense
	// 12 艘戰艦 + 1 座星際要塞。
	if len(fleet) != antaranDefLargeCount+antaranDefHugeCount+antaranDefTitanCount+1 {
		t.Fatalf("應是 3+2+7 艘 + 1 座要塞 = 13 個防禦單位,得到 %d", len(fleet))
	}
	// 分層:三種戰力值,而且數量比是 3:2:7。
	count := map[int]int{}
	for _, unit := range fleet {
		count[unit.Strength]++
	}
	titan := shipStrength("末日之星")
	// 末日之星那一格 = 7 艘 Harbinger + 1 座要塞(要塞用同等戰力當代理)。
	if got, want := count[titan], antaranDefTitanCount+1; got != want {
		t.Errorf("最高階應有 %d 個(7 艘 Harbinger + 1 座要塞),得到 %d", want, got)
	}
	if got := count[shipStrength("泰坦")]; got != antaranDefHugeCount {
		t.Errorf("Huge 級應有 %d 艘,得到 %d", antaranDefHugeCount, got)
	}
	if got := count[shipStrength("戰艦")]; got != antaranDefLargeCount {
		t.Errorf("Large 級應有 %d 艘,得到 %d", antaranDefLargeCount, got)
	}
	// 相對關係:Harbinger 最多(7),而且比另外兩級加起來還多——這是那張表最顯眼的形狀,
	// 抄反了(例如 7/2/3)這裡會抓到。
	if antaranDefTitanCount <= antaranDefLargeCount+antaranDefHugeCount {
		t.Errorf("Titan 級(%d)應多於 Large+Huge(%d)",
			antaranDefTitanCount, antaranDefLargeCount+antaranDefHugeCount)
	}

	// 這些不是由 remake 戰力階梯反推的數字，而是原版
	// Load_*_Antaran_Combat_Ship_ 的五級 loader 對應：Large/Huge/Titan。
	classCount := map[gamedata.CombatShipClass]int{}
	fortressCount := 0
	for _, unit := range fleet {
		classCount[unit.CombatClass]++
		if unit.Fortress {
			fortressCount++
		}
	}
	if got := classCount[gamedata.SHIP_BATTLESHIP]; got != antaranDefLargeCount {
		t.Errorf("Intruder 應保留為 Battleship 級,得到 %d", got)
	}
	if got := classCount[gamedata.SHIP_TITAN]; got != antaranDefHugeCount {
		t.Errorf("Interdictor 應保留為 Titan 級,得到 %d", got)
	}
	if got := classCount[gamedata.SHIP_DOOMSTAR]; got != antaranDefTitanCount+1 {
		t.Errorf("Harbinger+要塞 應保留為 Doom Star 級,得到 %d", got)
	}
	if fortressCount != 1 {
		t.Errorf("應有且只有一座星際要塞,得到 %d", fortressCount)
	}
}

// 原版即時戰鬥 loader `sub_55738`(Intruder) 與 `sub_55F67`(Harbinger)
// 的戰鬥設計各自含有 ID 31 Fighter Bays，數量分別是 3 與 6；
// `sub_55B12`(Interdictor) 沒有這個槽。
func TestAntaranHomeFleetPreservesKnownFighterBayCounts(t *testing.T) {
	type knownDesign struct {
		loaderRaw uint32
		weaponID  int
		bays      int
	}
	want := map[string]knownDesign{
		"Intruder":      {loaderRaw: 0x55738, weaponID: 31, bays: 3},
		"Interdictor":   {loaderRaw: 0x55B12, bays: 0},
		"Harbinger":     {loaderRaw: 0x55F67, weaponID: 31, bays: 6},
		"Star Fortress": {loaderRaw: 0x4D18E, bays: 0},
	}
	for _, unit := range antaranHomeFleetDefense {
		known := want[unit.OriginalName]
		if unit.CombatLoaderRaw != known.loaderRaw {
			t.Errorf("%s 的 combat loader raw 應為 0x%X,得到 0x%X", unit.OriginalName, known.loaderRaw, unit.CombatLoaderRaw)
		}
		if unit.FighterBayWeaponID != known.weaponID {
			t.Errorf("%s 的 Fighter Bay weapon ID 應為 %d,得到 %d", unit.OriginalName, known.weaponID, unit.FighterBayWeaponID)
		}
		if unit.FighterBayCount != known.bays {
			t.Errorf("%s 的已知 Fighter Bays 應為 %d,得到 %d", unit.OriginalName, known.bays, unit.FighterBayCount)
		}
	}
}

func TestAntaranHomeFleetPreservesKnownCombatWeaponSlots(t *testing.T) {
	want := map[string][]antaranWeaponSlot{
		"Intruder": {
			{WeaponID: 4, Quantity: 4, RawFlags: 0}, {WeaponID: 4, Quantity: 2, RawFlags: 2},
			{WeaponID: 24, Quantity: 5, RawFlags: 0}, {WeaponID: 13, Quantity: 2, RawFlags: 0},
			{WeaponID: 31, Quantity: 3, RawFlags: 0},
		},
		"Interdictor": {
			{WeaponID: 4, Quantity: 6, RawFlags: 0}, {WeaponID: 4, Quantity: 2, RawFlags: 2},
			{WeaponID: 24, Quantity: 15, RawFlags: 0}, {WeaponID: 13, Quantity: 2, RawFlags: 0},
			{WeaponID: 4, Quantity: 8, RawFlags: 4}, {WeaponID: 11, Quantity: 2, RawFlags: 0},
		},
		"Harbinger": {
			{WeaponID: 4, Quantity: 10, RawFlags: 0}, {WeaponID: 4, Quantity: 2, RawFlags: 2},
			{WeaponID: 24, Quantity: 20, RawFlags: 0}, {WeaponID: 13, Quantity: 3, RawFlags: 0},
			{WeaponID: 4, Quantity: 15, RawFlags: 4}, {WeaponID: 11, Quantity: 2, RawFlags: 2},
			{WeaponID: 37, Quantity: 1, RawFlags: 0}, {WeaponID: 31, Quantity: 6, RawFlags: 0},
		},
	}
	for _, unit := range antaranHomeFleetDefense {
		if unit.Fortress {
			wantFortress := []antaranWeaponSlot{
				{WeaponID: 11, Seed: 375, CapacityCap: 99, RawFlags: 2},
				{WeaponID: 4, Seed: 187, CapacityCap: 198, RawFlags: 0},
				{WeaponID: 4, Seed: 187, CapacityCap: 198, RawFlags: 4},
				{WeaponID: 4, Seed: 375, CapacityCap: 99, RawFlags: 2},
			}
			if len(unit.WeaponSlots) != len(wantFortress) {
				t.Errorf("星際要塞應保留 %d 個已解出的槽位，得到 %d", len(wantFortress), len(unit.WeaponSlots))
				continue
			}
			for i := range wantFortress {
				if unit.WeaponSlots[i] != wantFortress[i] {
					t.Errorf("星際要塞 slot %d 應為 %#v，得到 %#v", i, wantFortress[i], unit.WeaponSlots[i])
				}
			}
			continue
		}
		got := unit.WeaponSlots
		wantSlots := want[unit.OriginalName]
		if len(got) != len(wantSlots) {
			t.Fatalf("%s 的已知武器槽數應為 %d,得到 %d", unit.OriginalName, len(wantSlots), len(got))
		}
		for i := range wantSlots {
			if got[i] != wantSlots[i] {
				t.Errorf("%s slot %d 應為 ID=%d qty=%d rawFlags=0x%X,得到 ID=%d qty=%d rawFlags=0x%X",
					unit.OriginalName, i, wantSlots[i].WeaponID, wantSlots[i].Quantity, wantSlots[i].RawFlags,
					got[i].WeaponID, got[i].Quantity, got[i].RawFlags)
			}
		}
	}
}

func TestAntaranFortressDivisorAndRuntimeQuantity(t *testing.T) {
	fortress := antaranFortressSlots()
	cases := []struct {
		name  string
		slot  int
		tech  int
		div   int
		count int
	}{
		{name: "death ray T2", slot: 0, tech: 2, div: 42, count: 8},
		{name: "death ray T3", slot: 0, tech: 3, div: 36, count: 10},
		{name: "particle raw0 other", slot: 1, tech: 0, div: 6, count: 31},
		{name: "particle raw4 T2", slot: 2, tech: 2, div: 5, count: 37},
	}
	for _, tc := range cases {
		slot := fortress[tc.slot]
		if got := antaranFortressDivisor(slot.WeaponID, slot.RawFlags, tc.tech); got != tc.div {
			t.Errorf("%s divisor=%d,want %d", tc.name, got, tc.div)
		}
		if got := antaranFortressRuntimeQuantity(slot, tc.tech); got != tc.count {
			t.Errorf("%s runtime quantity=%d,want %d", tc.name, got, tc.count)
		}
	}
	if got := antaranFortressDivisor(4, 0x10, 2); got != 0 {
		t.Errorf("未知 raw flag 應 fail-closed，得到 divisor=%d", got)
	}
}

func TestAntaranStarFortressFeedsAllDirectWeaponFirepower(t *testing.T) {
	var fortress antaranDefenseUnit
	for _, unit := range antaranHomeFleetDefense {
		if unit.Fortress {
			fortress = unit
			break
		}
	}
	minDamage, maxDamage := antaranWeaponFirepower(fortress)
	if minDamage <= 0 || maxDamage <= minDamage {
		t.Fatalf("星際要塞直接武器火力應進入齊射：%d..%d", minDamage, maxDamage)
	}
	combatant := antaranDefenseCombatant(fortress)
	if combatant.wmin != minDamage || combatant.wmax != maxDamage {
		t.Fatalf("要塞齊射應消費完整武器總和：combatant=%d..%d raw=%d..%d",
			combatant.wmin, combatant.wmax, minDamage, maxDamage)
	}
	if combatant.atk <= fortress.Strength {
		t.Fatalf("要塞 BA 應包含直接火力，得到 %d，基礎 %d", combatant.atk, fortress.Strength)
	}
}

func TestAntaranFighterBaysUseTheExistingQuickBattleContribution(t *testing.T) {
	unit := antaranDefenseUnit{
		OriginalName:       "Intruder",
		Strength:           shipStrength("戰艦"),
		CombatClass:        gamedata.SHIP_BATTLESHIP,
		FighterBayWeaponID: 31,
		FighterBayCount:    3,
	}
	got := antaranDefenseCombatant(unit)
	fighterAttack, fighterHP := gamedata.FighterBayCombatContribution()
	if got.atk != unit.Strength+3*fighterAttack {
		t.Fatalf("Intruder 的攻擊應含 3 艙戰機貢獻,得到 %d", got.atk)
	}
	if got.hp != unit.Strength*3+3*fighterHP {
		t.Fatalf("Intruder 的結構應含 3 艙戰機貢獻,得到 %d", got.hp)
	}
}

// 換成真值之後艦隊變強了——先前是 6×64 = 384,現在是 3×16+2×32+8×64 = 624。
// 這一支確認「終局一戰」沒有因為改組成而變簡單。
func TestAntaranHomeFleetIsStrongerThanTheOldPlaceholder(t *testing.T) {
	total := 0
	for _, unit := range antaranHomeFleetDefense {
		total += unit.Strength
	}
	const oldPlaceholder = 6 * 64
	if total <= oldPlaceholder {
		t.Errorf("真值組成的總戰力 %d 不該低於先前的保守預設 %d", total, oldPlaceholder)
	}
}
