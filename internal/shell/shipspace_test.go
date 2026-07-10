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
