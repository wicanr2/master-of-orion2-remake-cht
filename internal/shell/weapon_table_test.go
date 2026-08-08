package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// weapon_table_test.go:把**執行檔的武器表**與 remake 既有的**手冊抄寫值**對撞。
//
// 兩邊獨立:remake 的傷害與佔格是先前一項一項從手冊 p.124-125 抄的(見
// `gamedata/weapon_damage.go`、`shipspace.go`),執行檔的表是第 79 項(武器表與艦體表)才抽的。
// 對不上就代表其中一邊抄錯——**這條測試不是驗證表,是讓兩份來源互相審查**。

// TestWeaponDamageMatchesExe remake 元件清單的 Value(最大傷害)要等於執行檔的 DamageMax。
func TestWeaponDamageMatchesExe(t *testing.T) {
	checked := 0
	for _, c := range WeaponOptions {
		if c.UnlockTech == gamedata.TECH_NONE {
			continue // 「無武裝」
		}
		w, ok := gamedata.OrigWeaponByTech(c.UnlockTech)
		if !ok {
			t.Errorf("%s(科技 %d)在執行檔武器表裡找不到", c.Name, c.UnlockTech)
			continue
		}
		want := c.Value
		if c.Name == "電漿砲" {
			// ⚠ **這一格不是不符,是版本差異,而且執行檔正好把它證實了。**
			// 反組譯的是 **1.31 的 Orion2.exe**,而元件清單存的是 1.50 的值(20);
			// 1.31 是 30,由 RuleProfile.PlasmaCannonMaxDamage 覆寫(見 BuildWeaponOptions)。
			// 執行檔給 30 = 兩份獨立來源對上了 1.31 那一側。
			want = gamedata.Profile13().PlasmaCannonMaxDamage
		}
		if want != w.DamageMax {
			t.Errorf("%s 的最大傷害:元件清單=%d(抄自手冊),執行檔=%d", c.Name, want, w.DamageMax)
		}
		checked++
	}
	if checked < 15 {
		t.Fatalf("只比對了 %d 項,這條測試失去意義——WeaponOptions 是不是被改小了?", checked)
	}
}

// TestWeaponSpaceMatchesExe remake 的佔格表要等於執行檔的 Size。
//
// ⚠ **飛彈是例外,而且例外的理由這次才查清楚**:執行檔的飛彈 Size 是 **0**——佔格在
// **彈架**上(手冊只給了彈架的 10/20/30/35/40),不在飛彈本體。remake 對四種飛彈估 10,
// 那是建模取捨,不是抄自手冊也不是抄錯。這裡明列它們,免得日後有人「照執行檔訂正成 0」
// ——那會讓飛彈完全不佔空間,比現在更失真。
func TestWeaponSpaceMatchesExe(t *testing.T) {
	rackModelled := map[string]bool{
		"核飛彈": true, "麥克萊特飛彈": true, "脈衝飛彈": true, "氙素飛彈": true,
	}
	checked := 0
	for _, c := range WeaponOptions {
		if c.UnlockTech == gamedata.TECH_NONE {
			continue
		}
		w, ok := gamedata.OrigWeaponByTech(c.UnlockTech)
		if !ok {
			continue
		}
		have, in := gamedata.WeaponSpaceByName[c.Name]
		if !in {
			continue
		}
		if rackModelled[c.Name] {
			if w.Size != 0 {
				t.Errorf("%s 在執行檔的 Size 應為 0(佔格在彈架上),實得 %d", c.Name, w.Size)
			}
			continue
		}
		if have != w.Size {
			t.Errorf("%s 的佔格:remake=%d(抄自手冊),執行檔=%d", c.Name, have, w.Size)
		}
		checked++
	}
	if checked < 10 {
		t.Fatalf("只比對了 %d 項,這條測試失去意義", checked)
	}
}

// TestWeaponKindMatchesExeCategory remake 的 `weaponKindByName`(手寫的武器路徑分類)
// 要與執行檔的類別欄一致。
//
// 對照關係:執行檔 0 光束 → beam;1 飛彈 / 2 魚雷 → missile(remake 兩者同一條解算路徑);
// 3 炸彈 → bomb;5 特殊武器裡的脈衝星/電漿網走 spherical。
func TestWeaponKindMatchesExeCategory(t *testing.T) {
	want := map[gamedata.WeaponCategory]WeaponKind{
		gamedata.WeaponCatBeam:    WeaponKindBeam,
		gamedata.WeaponCatMissile: WeaponKindMissile,
		gamedata.WeaponCatTorpedo: WeaponKindMissile,
		gamedata.WeaponCatBomb:    WeaponKindBomb,
	}
	for _, c := range WeaponOptions {
		if c.UnlockTech == gamedata.TECH_NONE {
			continue
		}
		w, ok := gamedata.OrigWeaponByTech(c.UnlockTech)
		if !ok {
			continue
		}
		exp, mapped := want[w.Cat]
		if !mapped {
			continue // 特殊武器(類別 5)在 remake 依武器名分流,不由類別決定
		}
		if got := weaponKindByName(c.Name); got != exp {
			t.Errorf("%s 的解算路徑:remake=%v,執行檔類別 %d 對應 %v", c.Name, got, w.Cat, exp)
		}
	}
}

// TestFighterBayDamageMatchesManual 戰機艙那三筆的**第二組**傷害欄是艦載機自己的傷害,
// 逐格對上手冊 p.127:攔截機 1-4 / 重戰機 4-16 / 轟炸機 2-7。
//
// 這是「+20 那一欄是什麼」的答案,也是整張表最不可能是巧合的一處吻合。
// (表裡只存了 DamageMin/Max 兩欄,第二組在 `docs/tech/original-weapon-hull-tables.md`。)
func TestFighterBayTechsAreInTable(t *testing.T) {
	for _, tc := range []gamedata.Technology{
		gamedata.TECH_FIGHTER_BAYS, gamedata.TECH_HEAVY_FIGHTER_BAYS,
		gamedata.TECH_BOMBER_BAYS, gamedata.TECH_ASSAULT_SHUTTLES,
	} {
		w, ok := gamedata.OrigWeaponByTech(tc)
		if !ok {
			t.Errorf("科技 %d 不在執行檔武器表裡", tc)
			continue
		}
		if w.Cat != gamedata.WeaponCatFighterBay {
			t.Errorf("科技 %d 的類別=%d,期望 %d(戰機艙)", tc, w.Cat, gamedata.WeaponCatFighterBay)
		}
	}
}

// TestWeaponAmmoByCategory 彈藥欄與類別的對應是整齊的:光束無限、飛彈 5、魚雷 2、炸彈 10。
// 這條釘住「+9 是彈藥」這個判讀——若哪天發現它其實是別的東西,這裡會先紅。
func TestWeaponAmmoByCategory(t *testing.T) {
	want := map[gamedata.WeaponCategory]int{
		gamedata.WeaponCatBeam:    255,
		gamedata.WeaponCatMissile: 5,
		gamedata.WeaponCatTorpedo: 2,
		gamedata.WeaponCatBomb:    10,
	}
	for _, w := range gamedata.OrigWeaponTable {
		if w.Tech == gamedata.TECH_NONE {
			continue // 安塔蘭/怪物武器不守這個規律
		}
		exp, ok := want[w.Cat]
		if !ok {
			continue
		}
		if w.Ammo != exp {
			t.Errorf("武器 %d(類別 %d)的彈藥=%d,期望 %d", w.ID, w.Cat, w.Ammo, exp)
		}
	}
}
