package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// bcCrashFloor80Turns 是 TestAntaresRaidsScheduleAndEscalate 80 回合內允許的 BC 下限。
//
// 忠實 yield 經濟(母星 Terran/Abundant,見 docs/tech/colony-economy-maintenance.md)下,
// 建築維護費固定 3 BC/回合,但人口只剩 1 時,不論把僅存的 1 人配置成農夫或工人,收入都不到
// 3 BC(食物盈餘出售 0.5 BC/單位、稅收 40%,單人口撐死賺 1~2 BC)——這不是本輪任何算式錯誤,
// 而是「建築維護費不隨人口規模縮小」這個手冊本身就有的機制,在忠實(零緩衝)經濟下被誠實呈現
// 出來:此測試刻意無艦隊防禦、吃滿安塔蘭入侵傷害,人口會被反覆打到剩 1(母星人口下限本身仍
// 受下方斷言保護),因此本測試不再要求「BC 絕不為負」(那個假設建立在舊 placeholder 經濟
// NetBC 穩定 +3/回合累積出的巨額緩衝上,忠實經濟沒有這個緩衝)。改驗證「BC 不會失控式無下限
// 崩潰」——以本測試固定 EventSeed=42 的確定性軌跡實測,80 回合最低點在回合 43 觸底後回升
// (2026-07-12 校正母星分配 農4/工1/科3 後為 -24,先前 農4/工3/科1 較高工業時為 -3;較忠實的
// 低工業母星緩衝更薄故觸底更深,但仍有界且會恢復,非螺旋崩潰),這裡抓一個有餘裕但仍能抓到
// 「異常擴大化」的下限。2026-07-12 再校正開局 BC 100→50(SAVE10 oracle)後,同軌跡最低點
// 由 -24 降到 -31(回合43 觸底後仍回升至 -3、人口守 1;因 BC 低時買不起的支出會跳過而自限,
// 非線性下移),故下限放寬到 -40 留餘裕,仍能抓真正的無下限螺旋。
const bcCrashFloor80Turns = -40

// TestAntaresRaidsScheduleAndEscalate 驗證安塔蘭入侵:前期寬限不觸發,達排程回合週期性觸發,
// 次數遞增(升級),母星人口不低於 1,且 BC 不會失控式無下限崩潰(見 bcCrashFloor80Turns 註解:
// 忠實經濟下人口被打到剩 1 時,單人口收入結構性不足以覆蓋建築維護費,短暫轉負是誠實的經濟後果,
// 不是 bug)。
func TestAntaresRaidsScheduleAndEscalate(t *testing.T) {
	s := NewDemoSession()
	s.Fleet().Ships = nil // 無艦隊防禦,吃滿傷害(方便觀察)

	raidTurns := []int{}
	for i := 0; i < 80; i++ {
		s.EndTurn()
		if s.LastAntares != "" {
			raidTurns = append(raidTurns, s.Turn)
			if s.LastAntaresEN == "" {
				t.Error("安塔蘭警報已有中文報告時也應有英文報告")
			}
		}
		if s.Player.BC < bcCrashFloor80Turns {
			t.Fatalf("BC 崩潰超出合理下限:%d(< %d)", s.Player.BC, bcCrashFloor80Turns)
		}
		if s.PlayerColonies[0].Population < 1 {
			t.Fatalf("母星人口 <1:%d", s.PlayerColonies[0].Population)
		}
	}
	if len(raidTurns) < 2 {
		t.Fatalf("80 回合內應有多次安塔蘭入侵,實得 %d 次:%v", len(raidTurns), raidTurns)
	}
	// 首次不早於寬限回合。
	if raidTurns[0] < antaresStartTurn {
		t.Fatalf("首次入侵 %d 早於寬限 %d", raidTurns[0], antaresStartTurn)
	}
	// 週期一致。
	if raidTurns[1]-raidTurns[0] != antaresInterval {
		t.Fatalf("入侵週期應為 %d,實得 %d", antaresInterval, raidTurns[1]-raidTurns[0])
	}
	if s.AntaresRaids != len(raidTurns) {
		t.Fatalf("AntaresRaids 計數 %d != 觸發次數 %d", s.AntaresRaids, len(raidTurns))
	}
	t.Logf("安塔蘭入侵回合:%v(共 %d 次)", raidTurns, s.AntaresRaids)
}

// TestAntaresDefenseReducesDamage 驗證母星有艦隊時損失較低。
func TestAntaresDefenseReducesDamage(t *testing.T) {
	run := func(withFleet bool) int {
		s := NewDemoSession()
		// 隔離變數:拿掉 AI 對手。這條測試量的是「安塔蘭入侵造成的 BC 損失」,但它其實是用
		// 「該回合國庫的總變化」當代理值,任何其他也會扣 BC 的系統都會混進來。AI 突襲
		// (2026-08-06 新增,見 ai_attack.go)正是這樣一個系統,而且它的發動條件**看玩家軍力**
		// ——正好是這條測試在對照的那個變數,不隔離就會反向污染結果。
		s.AIPlayers = nil
		if !withFleet {
			s.Fleet().Ships = nil
		} else {
			s.Fleet().Ships = []Ship{{Name: "衛戍艦", Class: "戰艦"}}
			s.Fleet().AtStar = 0
		}
		startBC := 0
		bcLoss := 0
		for i := 0; i < 40; i++ {
			before := s.Player.BC
			s.EndTurn()
			if s.LastAntares != "" {
				bcLoss += before - s.Player.BC
			}
			_ = startBC
		}
		return bcLoss
	}
	undefended := run(false)
	defended := run(true)
	if defended >= undefended {
		t.Fatalf("有母星艦隊防禦應損失較少 BC:防禦 %d vs 無防禦 %d", defended, undefended)
	}
	t.Logf("BC 損失:無防禦 %d、有防禦 %d", undefended, defended)
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
