package gamedata

import "testing"

// special_devices_test.go:把**執行檔裡的原始數字**釘死在測試裡,而不是釘在被測的表上。
//
// 為什麼要重寫一次:`specialDeviceTable` 的 Tech 欄寫的是 `TECH_*` 常數,
// 所以「表裡的科技編號對不對」不能拿那張表自己來驗——只要列舉改了,表會安靜地跟著改。
// 這裡列的是**從 Orion2.exe 讀出來的整數**(0x17EEE0 起、每筆 47 位元組的 +6 欄),
// 兩邊獨立,列舉一動就會紅。

// exeSpecialDeviceTech 是執行檔特殊裝置表的 +6 欄逐筆抄錄。
var exeSpecialDeviceTech = []struct {
	dev  SpecialDevices
	tech int
}{
	{SPEC_ACHILLES_TARGETING_UNIT, 1},
	{SPEC_AUGMENTED_ENGINES, 20},
	{SPEC_AUTOMATED_REPAIR_UNIT, 23},
	{SPEC_BATTLE_PODS, 25},
	{SPEC_BATTLE_SCANNER, 26},
	{SPEC_CLOAKING_DEVICE, 38},
	{SPEC_DAMPER_FIELD, 45},
	{SPEC_DISPLACEMENT_DEVICE, 53},
	{SPEC_ECM_JAMMER, 57},
	{SPEC_ENERGY_ABSORBER, 60},
	{SPEC_EXTENDED_FUEL_TANKS, 63},
	{SPEC_FAST_MISSILE_RACKS, 64},
	{SPEC_HARD_SHIELDS, 81},
	{SPEC_HEAVY_ARMOR, 82},
	{SPEC_HIGH_ENERGY_FOCUS, 85},
	{SPEC_HYPERX_CAPACITORS, 90},
	{SPEC_INERTIAL_NULLIFIER, 93},
	{SPEC_INERTIAL_STABILIZER, 94},
	{SPEC_LIGHTNING_FIELD, 102},
	{SPEC_MULTIPHASED_SHIELDS, 112},
	{SPEC_MULTIWAVE_ECM_JAMMER, 111},
	{SPEC_PHASE_SHIFTER, 125},
	{SPEC_PHASING_CLOAK, 126},
	{SPEC_QUANTUM_DETONATOR, 150},
	{SPEC_RANGEMASTER_UNIT, 151},
	{SPEC_REFLECTION_FIELD, 153},
	{SPEC_REINFORCED_HULL, 56},
	{SPEC_SCOUT_LAB, 158},
	{SPEC_SECURITY_STATIONS, 159},
	{SPEC_SHIELD_CAPACITOR, 161},
	{SPEC_STEALTH_FIELD, 172},
	{SPEC_STRUCTURAL_ANALYZER, 175},
	{SPEC_SUB_SPACE_TELEPORTER, 177},
	{SPEC_TIME_WARP_FACILITATOR, 185},
	{SPEC_TRANSPORTERS, 190},
	{SPEC_TROOP_PODS, 192},
	{SPEC_WARP_DISSIPATOR, 196},
	{SPEC_WIDE_AREA_JAMMER, 199},
	{SPEC_REGENERATION, 23},
}

func TestSpecialDeviceTechMatchesExe(t *testing.T) {
	if len(exeSpecialDeviceTech) != len(specialDeviceTable) {
		t.Fatalf("表長度不符:執行檔 %d 筆,specialDeviceTable %d 筆",
			len(exeSpecialDeviceTech), len(specialDeviceTable))
	}
	for _, c := range exeSpecialDeviceTech {
		st, ok := SpecialDeviceStatsFor(c.dev)
		if !ok {
			t.Errorf("specialDeviceTable 缺 SPEC=%d", c.dev)
			continue
		}
		if int(st.Tech) != c.tech {
			t.Errorf("SPEC=%d 的科技編號:表=%d,執行檔=%d", c.dev, st.Tech, c.tech)
		}
	}
}

// TestBattlePodsSpaceIsHalfHull 是這張表最強的一道交叉驗證:
// 戰鬥艙的負佔格 = 艦體空間的一半(向下取整),而艦體空間那張表是**從手冊 p.121 抄的**。
// 兩份來源獨立,連 12 = floor(25/2) 的截斷都一致。
//
// 這條紅了代表**表抄錯了**(或欄位偏移認錯),不是規則變了——先回去重跑抽取。
func TestBattlePodsSpaceIsHalfHull(t *testing.T) {
	for c := SHIP_FRIGATE; c <= SHIP_DOOMSTAR; c++ {
		want := -(ShipHullSpace(c) / 2)
		if got := SpecialDeviceSpace(SPEC_BATTLE_PODS, c); got != want {
			t.Errorf("戰鬥艙在 %d 級艦的佔格=%d,期望 %d(= −艦體空間 %d / 2)",
				c, got, want, ShipHullSpace(c))
		}
	}
}

// TestSpecialDeviceSpaceIsPerHullClass 釘住「同一個系統在不同艦體上佔格不同」這件事
// ——remake 舊模型是單一估計值,這是最大的行為差異,不能被日後的重構悄悄退回去。
func TestSpecialDeviceSpaceIsPerHullClass(t *testing.T) {
	f := SpecialDeviceSpace(SPEC_HARD_SHIELDS, SHIP_FRIGATE)
	d := SpecialDeviceSpace(SPEC_HARD_SHIELDS, SHIP_DOOMSTAR)
	if f != 10 || d != 250 {
		t.Errorf("硬化護盾佔格:巡防艦=%d(期望 10)、末日之星=%d(期望 250)", f, d)
	}
	if SpecialDeviceCost(SPEC_BATTLE_PODS, SHIP_DOOMSTAR) != 2500 {
		t.Errorf("戰鬥艙在末日之星的成本=%d,期望 2500",
			SpecialDeviceCost(SPEC_BATTLE_PODS, SHIP_DOOMSTAR))
	}
}

// TestSpecialDeviceOutOfRange 邊界:未知裝置與非法艦級一律回 0,不 panic。
func TestSpecialDeviceOutOfRange(t *testing.T) {
	if SpecialDeviceSpace(SpecialDevices(999), SHIP_FRIGATE) != 0 ||
		SpecialDeviceCost(SpecialDevices(999), SHIP_FRIGATE) != 0 {
		t.Error("未知裝置應回 0")
	}
	if SpecialDeviceSpace(SPEC_BATTLE_PODS, CombatShipClass(6)) != 0 ||
		SpecialDeviceSpace(SPEC_BATTLE_PODS, -1) != 0 {
		t.Error("非法艦級應回 0")
	}
}

// TestDesignHullSpaceMegafluxers 手冊給 +25%,執行檔給「×125 再整除 100」的截斷方式。
// 巡防艦 25 → 31(31.25 截掉),這個截斷是**執行檔的行為**,不是四捨五入。
func TestDesignHullSpaceMegafluxers(t *testing.T) {
	cases := []struct {
		class       CombatShipClass
		plain, mega int
	}{
		{SHIP_FRIGATE, 25, 31},
		{SHIP_DESTROYER, 60, 75},
		{SHIP_CRUISER, 120, 150},
		{SHIP_BATTLESHIP, 250, 312},
		{SHIP_TITAN, 500, 625},
		{SHIP_DOOMSTAR, 1200, 1500},
	}
	for _, c := range cases {
		if got := DesignHullSpace(c.class, false); got != c.plain {
			t.Errorf("無巨型通量器 %d 級=%d,期望 %d", c.class, got, c.plain)
		}
		if got := DesignHullSpace(c.class, true); got != c.mega {
			t.Errorf("有巨型通量器 %d 級=%d,期望 %d", c.class, got, c.mega)
		}
	}
}
