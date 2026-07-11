package shell

import "testing"

// TestShipDesignSpaceUsedWeaponOnly 只算武器佔格(裝甲/護盾依手冊行為不佔空間,見 session.go
// ShipDesignSpaceUsed 註解),特殊系統未選裝時不額外加估計值。
func TestShipDesignSpaceUsedWeaponOnly(t *testing.T) {
	// 雷射(index 1)= 10 格;裝甲/護盾隨便選一個非 0 index,驗證確實不影響空間。
	used := ShipDesignSpaceUsed("護衛艦", 1 /*雷射*/, 3 /*佐特裝甲*/, 2 /*第三級護盾*/, 0 /*無特殊*/)
	if used != 10 {
		t.Errorf("ShipDesignSpaceUsed(雷射,佐特裝甲,第三級護盾,無)=%d, want 10(裝甲/護盾不佔空間)", used)
	}
}

// TestShipDesignSpaceUsedWithSpecial 選裝特殊系統時,額外加上 SpecialSpace 估計值。
func TestShipDesignSpaceUsedWithSpecial(t *testing.T) {
	used := ShipDesignSpaceUsed("末日之星", 0 /*無武裝*/, 0, 0, 1 /*戰鬥電腦*/)
	// 末日之星 hull=1200,估計 5% = 60。
	if used != 60 {
		t.Errorf("ShipDesignSpaceUsed(末日之星,無武裝,特殊)=%d, want 60", used)
	}
}

// TestShipDesignFitsSmallHullOverflow 小艦體塞大元件 → 超格(ShipDesignFits=false)。
func TestShipDesignFitsSmallHullOverflow(t *testing.T) {
	// 護衛艦空間 25;死光(index 9)佔 30 格,必超格。
	deathRayIdx := -1
	for i, c := range WeaponOptions {
		if c.Name == "死光" {
			deathRayIdx = i
		}
	}
	if deathRayIdx < 0 {
		t.Fatal("找不到「死光」元件,WeaponOptions 定義可能已變動")
	}
	if ShipDesignFits("護衛艦", deathRayIdx, 0, 0, 0) {
		t.Error("護衛艦裝死光(Size 30 > hull 25)應該回報 ShipDesignFits=false")
	}
}

// TestShipDesignFitsLargeHullAccommodates 大艦體容納同一元件組合。
func TestShipDesignFitsLargeHullAccommodates(t *testing.T) {
	deathRayIdx := -1
	for i, c := range WeaponOptions {
		if c.Name == "死光" {
			deathRayIdx = i
		}
	}
	if deathRayIdx < 0 {
		t.Fatal("找不到「死光」元件")
	}
	if !ShipDesignFits("末日之星", deathRayIdx, 0, 0, 0) {
		t.Error("末日之星裝死光(Size 30 < hull 1200)應該回報 ShipDesignFits=true")
	}
}

// TestShipDesignFitsBoundary 邊界:恰好用滿空間應視為 fits(<=,非 <)。
func TestShipDesignFitsBoundary(t *testing.T) {
	// 驅逐艦 hull=60;雷射 10 格 * 6 把不可行(無多把武器模型),改用單元件邊界驗證:
	// 電漿砲 Size=25,護衛艦 hull=25,恰好用滿。
	plasmaIdx := -1
	for i, c := range WeaponOptions {
		if c.Name == "電漿砲" {
			plasmaIdx = i
		}
	}
	if plasmaIdx < 0 {
		t.Fatal("找不到「電漿砲」元件")
	}
	if !ShipDesignFits("護衛艦", plasmaIdx, 0, 0, 0) {
		t.Error("護衛艦裝電漿砲(Size 25 == hull 25)恰好用滿,應該回報 ShipDesignFits=true(<=)")
	}
}

// TestShipDesignFitsUnknownClassApproximatesFrigate 未知艦體等級(如偵察艦)近似 Frigate 空間判定。
func TestShipDesignFitsUnknownClassApproximatesFrigate(t *testing.T) {
	deathRayIdx := -1
	for i, c := range WeaponOptions {
		if c.Name == "死光" {
			deathRayIdx = i
		}
	}
	if ShipDesignFits("偵察艦", deathRayIdx, 0, 0, 0) {
		t.Error("偵察艦(近似 Frigate hull 25)裝死光(30)應該超格")
	}
}

// ---- 武器改造(mod)佔格接線:ShipDesignSpaceUsedWithMods / ShipDesignFitsWithMods ----

func laserIndex(t *testing.T) int {
	t.Helper()
	for i, c := range WeaponOptions {
		if c.Name == "雷射" {
			return i
		}
	}
	t.Fatal("找不到「雷射」元件")
	return -1
}

// TestShipDesignSpaceUsedWithMods_NoModsMatchesLegacy 無 mods 時應與舊版 ShipDesignSpaceUsed
// 完全一致(回歸)。
func TestShipDesignSpaceUsedWithMods_NoModsMatchesLegacy(t *testing.T) {
	li := laserIndex(t)
	legacy := ShipDesignSpaceUsed("護衛艦", li, 0, 0, 0)
	withNil := ShipDesignSpaceUsedWithMods("護衛艦", li, 0, 0, 0, nil)
	if legacy != withNil {
		t.Errorf("無 mods 應與舊版一致,legacy=%d withNil=%d", legacy, withNil)
	}
}

// TestShipDesignSpaceUsedWithMods_HeavyMountIncreasesSpace 掛 HV 武器佔格應加倍(+100%)。
func TestShipDesignSpaceUsedWithMods_HeavyMountIncreasesSpace(t *testing.T) {
	li := laserIndex(t)
	base := ShipDesignSpaceUsed("護衛艦", li, 0, 0, 0) // 雷射 10 格
	withHV := ShipDesignSpaceUsedWithMods("護衛艦", li, 0, 0, 0, []string{"HV"})
	if withHV != base*2 {
		t.Errorf("掛 HV 佔格應加倍,base=%d withHV=%d", base, withHV)
	}
}

// TestShipDesignFitsWithMods_HeavyMountCanOverflow 原本剛好塞得下的設計,掛 HV 後可能超格,
// ShipDesignFitsWithMods 應正確擋下。
func TestShipDesignFitsWithMods_HeavyMountCanOverflow(t *testing.T) {
	plasmaIdx := -1
	for i, c := range WeaponOptions {
		if c.Name == "電漿砲" {
			plasmaIdx = i
		}
	}
	if plasmaIdx < 0 {
		t.Fatal("找不到「電漿砲」元件")
	}
	// 護衛艦(hull 25)裝電漿砲(Size 25)恰好塞滿,見 TestShipDesignFitsBoundary。
	if !ShipDesignFits("護衛艦", plasmaIdx, 0, 0, 0) {
		t.Fatal("前提失敗:無 mod 應恰好塞滿")
	}
	if ShipDesignFitsWithMods("護衛艦", plasmaIdx, 0, 0, 0, []string{"HV"}) {
		t.Error("掛 HV 後電漿砲佔格變 50,應超出護衛艦 25 空間,ShipDesignFitsWithMods 應回 false")
	}
}

// TestShipDesignSpaceUsedWithMods_IgnoredForMissile mods 對飛彈武器無效(手冊 HV/PD/AF/CO
// 明文只講 beam,且飛彈路徑未接 mod 掛鉤)。
func TestShipDesignSpaceUsedWithMods_IgnoredForMissile(t *testing.T) {
	missileIdx := -1
	for i, c := range WeaponOptions {
		if c.Name == "核飛彈" {
			missileIdx = i
		}
	}
	if missileIdx < 0 {
		t.Fatal("找不到「核飛彈」元件")
	}
	base := ShipDesignSpaceUsed("護衛艦", missileIdx, 0, 0, 0)
	withMods := ShipDesignSpaceUsedWithMods("護衛艦", missileIdx, 0, 0, 0, []string{"HV"})
	if base != withMods {
		t.Errorf("飛彈武器掛 mods 不應改變佔格,base=%d withMods=%d", base, withMods)
	}
}

// TestDesignCostWithMods_HeavyMountIncreasesCost 手冊「adds to the size AND cost」,成本應與
// 佔格用同一套百分比公式。
func TestDesignCostWithMods_HeavyMountIncreasesCost(t *testing.T) {
	li := laserIndex(t)
	base := DesignCost("護衛艦", li, 0, 0, 0)
	withHV := DesignCostWithMods("護衛艦", li, 0, 0, 0, []string{"HV"})
	laserCost := WeaponOptions[li].Cost
	want := base + laserCost // +100% of laser cost
	if withHV != want {
		t.Errorf("掛 HV 成本應多加一份雷射成本,base=%d withHV=%d want=%d", base, withHV, want)
	}
}
