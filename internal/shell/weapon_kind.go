package shell

import "github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"

// weapon_kind.go:武器戰鬥解算路徑分類(beam / missile / spherical)。
//
// 分類依據優先順序:①WeaponOptions(session.go)裡的武器名是否對應手冊明確的飛彈/球形/炸彈
// 武器清單;②查無對應則預設 beam(現行絕大多數武器,含「無武裝」——空武裝仍沿用 beam
// 分支,反正 wmin=wmax=0 不影響結果,只是不要漏判成其他分支)。
//
// 已核對出處(手冊 moo2_patch1.5/MANUAL_150.html):
//   - missile:「核飛彈」「麥克萊特飛彈」「脈衝飛彈」「氙素飛彈」「質子魚雷」對應手冊
//     missile.go 已移植的 category 21／Missile 表；標準四種各有 Beam Defense，魚雷走
//     魚雷傷害模型。
//   - spherical:手冊「Notes on Spherical Damage > Spherical Weapons」(p126)明列的球形
//     武器是 Pulsar、Plasma Flux(Eel 專屬)、Spatial Compressor、Engine Explosion——
//     WeaponOptions 目前已掛載 Pulsar／Spatial Compressor／Gyro Destabilizer；死光
//     (Death Ray) 仍是一般光束武器。Plasma Flux／Engine Explosion 等特殊來源尚未進入
//     玩家設計表。
//
// 新增武器到 WeaponOptions 時,記得同步檢查是否需要在這裡新增分類(預設落到 beam 不會
// 編譯錯誤,但戰鬥行為會不忠實)。
type WeaponKind int

const (
	WeaponKindBeam WeaponKind = iota
	WeaponKindMissile
	WeaponKindSpherical
	// WeaponKindBomb 是**只能打行星**的炸彈(第 64 項(武器傷害真表))。
	//
	// 手冊 p.126 逐字:「Bombs installed in a ship are only useful against **planetary
	// targets**」——所以它不是「傷害比較低的光束」,是**在艦隊戰裡完全沒有作用**。
	// 這正是它必須自成一類的理由:落到預設的 beam 分支會讓一艘掛核彈的船當光束艦用。
	WeaponKindBomb
)

// weaponKindByName 依 Component.Name(WeaponOptions 的武器名)分類戰鬥解算路徑。
func weaponKindByName(name string) WeaponKind {
	switch name {
	case "脈衝星", spatialCompressorName, gyroDestabilizerName:
		// 手冊 p.126「Notes on Spherical Damage」明列的球形武器是 Pulsar、Plasma Flux、
		// Spatial Compressor、Engine Explosion。remake 掛得上的是前後這兩項——
		// 電漿通量是海鰻怪獸專屬、引擎爆炸不是可裝載武器。
		return WeaponKindSpherical
	case "核彈", "融合彈", "反物質彈", "中子彈":
		// 執行檔的 category 表把這四項全歸在 **category 19(炸彈)**(第 52 項(生物武器分類)解出的
		// enum 語意),與手冊 p.126 的 BOMB 表列的正好是同一批——兩個獨立來源同意。
		return WeaponKindBomb
	case "核飛彈", "麥克萊特飛彈", "脈衝飛彈", "氙素飛彈", "反物質魚雷", "質子魚雷", "電漿魚雷":
		// 第 64 項(武器傷害真表)補的三項與既有兩項同類:執行檔的 category 表把它們全歸在
		// **category 21(飛彈/魚雷)**,與手冊 p.125 的 MISSILE 表一致——
		// 兩個獨立來源同意,不是照名字裡有「飛彈」兩個字分的。
		return WeaponKindMissile
	default:
		return WeaponKindBeam
	}
}

// antiMissileRocketName 是反飛彈火箭元件在 SpecialOptions 裡的名稱。
//
// 抽成常數而不是散在各處寫字串字面值:`shipHasAutoRepair` 那一族先前就是這樣做的,
// 而字串比對打錯字不會編譯錯誤——只會安靜地永遠不成立。
const antiMissileRocketName = "反飛彈火箭"

// spatialCompressorName 是空間壓縮器元件名(手冊唯一明講豁免護盾與裝甲的球形武器)。
const spatialCompressorName = "空間壓縮器"

// gyroDestabilizerName 是陀螺去穩器元件名。
//
// 它不在手冊 p.126 的球形武器清單上,但**兩個定義性特徵都有**:傷害依目標級數相乘、
// 完全豁免護盾與裝甲。第 64 項(武器傷害真表)當時把它擋在外面,理由是「光束路徑沒有 per size class
// 這個乘數」——那句話對,而正確的解法不是替光束加一個乘數,是**認出它其實是球形家族**。
const gyroDestabilizerName = "陀螺去穩器"

// weaponBypassesShieldAndArmor 回報這項武器是否直接打結構、無視護盾與裝甲。
//
// 手冊只有空間壓縮器與陀螺去穩器兩項寫著豁免(前者「does all damage to **structure only**,
// ignoring shields and armor」,後者「**Shields and armor offer no protection** and are not
// damaged」);脈衝星沒有那一句,所以**不要**推廣到整個球形類別。
func weaponBypassesShieldAndArmor(name string) bool {
	return name == spatialCompressorName || name == gyroDestabilizerName
}

// shipSizeClass 回傳艦體等級(球形武器的「per size class of target」要用)。
//
// 直接複用 shipClassFromName 的既有對照;它對不在手冊六級表裡的艦體(偵察艦)
// 回護衛艦近似,那是既有決定。
func shipSizeClass(class string) gamedata.CombatShipClass {
	c, _ := shipClassFromName(class)
	return c
}

// highEnergyFocusName 是高能聚焦系統的元件名(手冊 p.87「High Energy Focus (System)」)。
//
// 與 antiMissileRocketName 同款:名字在兩個地方出現(SpecialOptions 與戰列建構),
// 拉成常數免得其中一邊被改掉之後安靜失效。
const highEnergyFocusName = "高能聚焦"

// hefBonusFor 把「有沒有裝高能聚焦」換成傷害百分點加成。
//
// 手冊只給傷害那一項:「It does not improve the chances of hitting a target at a greater
// distance, nor does it prevent the normal drop-off of damage over range.」
// ——所以它只餵 DamageMountAdjustedValue 的 hefBonus,不碰命中門檻也不碰 dissipation。
func hefBonusFor(hasHEF bool) int {
	if hasHEF {
		return gamedata.DamageMountBonusHEF
	}
	return 0
}

// HEFDamageBonus 是 hefBonusFor 的匯出版本,供 cmd/moo2 的格子戰術戰鬥使用。
func HEFDamageBonus(hasHEF bool) int { return hefBonusFor(hasHEF) }

// heavyArmorName 是重裝甲系統的元件名(手冊「Heavy Armor (System)」)。
const heavyArmorName = "重裝甲"

// effectiveArmorHP 回傳這艘船的實際裝甲 HP:裝甲科技的值,裝了重裝甲再乘三。
//
// 手冊逐字:「Installing Heavy Armor triples the amount of damage the ship's armor can
// sustain before damage gets through to the structure and internal systems.」
// ——乘的是**裝甲**那一池,不是結構,所以只動 armorHPByName 的結果。
func effectiveArmorHP(sh Ship) int {
	hp := armorHPByName(sh.Armor)
	if sh.Special == heavyArmorName {
		hp *= gamedata.ArmorHeavyArmorMultiplier
	}
	return hp
}

// shipNegatesArmorPiercing 回報這艘船是否讓敵方的穿甲(AP)改造失效。
//
// 手冊給了兩條路,兩條都算:
//
//	Xentronium Armor:「Negates armor piercing effects of enemy weapons.」
//	Heavy Armor (System):「also negates the Armor Piercing abilities of enemy weapons
//	                       that hit the ship.」
//
// 這正是 `gamedata.DamageApplyArmor` 那個從寫出來就恆傳 false 的 apNegated 參數
// ——規則寫好了、註解連兩句原文都抄了,缺的只是「哪艘船算數」。
func shipNegatesArmorPiercing(sh Ship) bool {
	if sh.Special == heavyArmorName {
		return true
	}
	return gamedata.ArmorNegatesArmorPiercing(armorUnlockTechByName(sh.Armor))
}

// armorUnlockTechByName 依裝甲元件名查其解鎖科技;查無回 TECH_NONE。
func armorUnlockTechByName(name string) gamedata.Technology {
	for _, c := range ArmorOptions {
		if c.Name == name {
			return c.UnlockTech
		}
	}
	return gamedata.TECH_NONE
}
