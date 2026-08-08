package gamedata

// weapon_table.go:**原版執行檔的武器表**——46 筆 × (類別 / 彈藥 / 佔格 / 成本 / 傷害)。
//
// 與 special_devices.go 同一趟挖出來的(第 79 項(武器表與艦體表)),位置在
// `Orion2.exe` 資料段 `0x17F807` 起,每筆 28 (0x1C) 位元組:
//
//	+0  dd  名稱字串指標
//	+4  dw  武器編號(0..45 連號)
//	+6  dw  解鎖科技編號(0 = 沒有對應科技,見下)
//	+8  db  **類別**(0 光束 / 1 飛彈 / 2 魚雷 / 3 炸彈 / 4 戰機艙 / 5 特殊武器)
//	+9  db  **彈藥數**(255 = 無限,光束與多數特殊武器)
//	+10 dw  佔格(手冊 p.124-125 的 Size 欄)
//	+12 dw  成本
//	+16 dw  傷害下限
//	+18 dw  傷害上限
//	+20 2×db 第二組傷害(戰機艙這一欄是**艦載機自己的**傷害,見下)
//
// 定位方式:`Special_Devices_Available_`(`0x5F9EA`)在「這一格是武器」那條分支用
// `imul eax, 1Ch` + `word_17F80D`,而 `0x17F80D − 0x17F807 = 6`,與特殊裝置表同一個
// 版面家族(+6 是科技編號)。
//
// ============ 交叉驗證:與 remake 既有的手冊值對撞 ============
//
// remake 的武器傷害與佔格是先前**一項一項從手冊 p.124-125 抄的**,與這張表獨立。
// `weapon_table_test.go` 逐項比對,結果寫在那裡。幾個直接看得出來的:
//
//	雷射     Size 10、傷害 1-4      ← 手冊「1-4」
//	核融合光束 Size 10、傷害 2-6      ← 手冊「2-6」
//	質量投射器 Size 10、傷害 6        ← 手冊「6」
//	中子爆破槍 Size 10、傷害 3-12     ← 手冊「3-12」
//	戰機艙的第二組傷害 1-4 / 4-16 / 2-7 ← 手冊 p.127 攔截機 / 重戰機 / 轟炸機那三列
//
// ============ 兩件這張表講清楚、而手冊沒有的事 ============
//
//  1. **飛彈的 Size 是 0。** 手冊只給了彈架的 10/20/30/35/40,而 remake 先前對四種飛彈
//     一律估 10。這張表說飛彈**本體**不佔格——佔格在彈架上,那是另一張表。
//     所以 remake 那個 10 既沒有被證實也沒有被推翻,它是**建模取捨**,標記已改。
//  2. **編號 40..45 沒有解鎖科技。** 那六筆是安塔蘭/太空怪物專用武器(研究不到),
//     remake 目前沒有它們,也不該有——玩家造不出來。
//
// ⚠ **成本欄沒有接線。** 執行檔的武器成本(雷射 5、死光 75)與 remake 的尺度
// (雷射 20、死光 350)差約四倍,而艦體成本的方向相反(remake 巡防艦 18、執行檔 70)。
// 只換一邊會把相對平衡弄壞,而艦體表的成本欄目前**有一格對不上**(見
// docs/tech/original-weapon-hull-tables.md),所以兩邊都先不動。數字留在這裡備用。

// WeaponCategory 是執行檔武器表 +8 欄的類別。
type WeaponCategory int

const (
	WeaponCatBeam       WeaponCategory = 0 // 光束(彈藥 255 = 無限)
	WeaponCatMissile    WeaponCategory = 1 // 飛彈(彈藥 5)
	WeaponCatTorpedo    WeaponCategory = 2 // 魚雷(彈藥 2)
	WeaponCatBomb       WeaponCategory = 3 // 炸彈(彈藥 10,只對行星有用)
	WeaponCatFighterBay WeaponCategory = 4 // 戰機艙/突擊艇(彈藥 1)
	WeaponCatSpecial    WeaponCategory = 5 // 特殊武器(牽引光束/停滯力場/脈衝星/黑洞產生器…)
)

// OrigWeapon 是原版武器表的一筆。
type OrigWeapon struct {
	ID        int
	Tech      Technology // TECH_NONE = 沒有對應科技(安塔蘭/怪物武器)
	Cat       WeaponCategory
	Ammo      int // 255 = 無限
	Size      int // 佔格;飛彈為 0(佔格在彈架上)
	Cost      int // ⚠ 未接線,見檔頭
	DamageMin int
	DamageMax int
}

// OrigWeaponTable 是原版武器表的逐筆搬運。**不要手改任何數字。**
var OrigWeaponTable = []OrigWeapon{
	{ID: 0, Tech: TECH_NONE, Cat: WeaponCatBeam, Ammo: 255, Size: 0, Cost: 0, DamageMin: 0, DamageMax: 0},
	{ID: 1, Tech: TECH_MASS_DRIVER, Cat: WeaponCatBeam, Ammo: 255, Size: 10, Cost: 7, DamageMin: 6, DamageMax: 6},
	{ID: 2, Tech: TECH_GAUSS_CANNON, Cat: WeaponCatBeam, Ammo: 255, Size: 10, Cost: 10, DamageMin: 18, DamageMax: 18},
	{ID: 3, Tech: TECH_LASER_CANNON, Cat: WeaponCatBeam, Ammo: 255, Size: 10, Cost: 5, DamageMin: 1, DamageMax: 4},
	{ID: 4, Tech: TECH_PARTICLE_BEAM, Cat: WeaponCatBeam, Ammo: 255, Size: 15, Cost: 35, DamageMin: 10, DamageMax: 30},
	{ID: 5, Tech: TECH_FUSION_BEAM, Cat: WeaponCatBeam, Ammo: 255, Size: 10, Cost: 6, DamageMin: 2, DamageMax: 6},
	{ID: 6, Tech: TECH_ION_PULSE_CANNON, Cat: WeaponCatBeam, Ammo: 255, Size: 30, Cost: 15, DamageMin: 2, DamageMax: 10},
	{ID: 7, Tech: TECH_GRAVITON_BEAM, Cat: WeaponCatBeam, Ammo: 255, Size: 15, Cost: 12, DamageMin: 3, DamageMax: 15},
	{ID: 8, Tech: TECH_NEUTRON_BLASTER, Cat: WeaponCatBeam, Ammo: 255, Size: 10, Cost: 8, DamageMin: 3, DamageMax: 12},
	{ID: 9, Tech: TECH_PHASOR, Cat: WeaponCatBeam, Ammo: 255, Size: 10, Cost: 10, DamageMin: 5, DamageMax: 20},
	{ID: 10, Tech: TECH_DISRUPTER_CANNON, Cat: WeaponCatBeam, Ammo: 255, Size: 20, Cost: 25, DamageMin: 40, DamageMax: 40},
	{ID: 11, Tech: TECH_DEATH_RAY, Cat: WeaponCatBeam, Ammo: 255, Size: 30, Cost: 75, DamageMin: 50, DamageMax: 100},
	{ID: 12, Tech: TECH_PLASMA_CANNON, Cat: WeaponCatBeam, Ammo: 255, Size: 25, Cost: 15, DamageMin: 6, DamageMax: 30},
	{ID: 13, Tech: TECH_SPATIAL_COMPRESSOR, Cat: WeaponCatSpecial, Ammo: 255, Size: 50, Cost: 40, DamageMin: 4, DamageMax: 32},
	{ID: 14, Tech: TECH_NUCLEAR_MISSILE, Cat: WeaponCatMissile, Ammo: 5, Size: 0, Cost: 0, DamageMin: 8, DamageMax: 8},
	{ID: 15, Tech: TECH_MERCULITE_MISSILE, Cat: WeaponCatMissile, Ammo: 5, Size: 0, Cost: 0, DamageMin: 14, DamageMax: 14},
	{ID: 16, Tech: TECH_PULSON_MISSILE, Cat: WeaponCatMissile, Ammo: 5, Size: 0, Cost: 0, DamageMin: 20, DamageMax: 20},
	{ID: 17, Tech: TECH_ZEON_MISSILE, Cat: WeaponCatMissile, Ammo: 5, Size: 0, Cost: 0, DamageMin: 30, DamageMax: 30},
	{ID: 18, Tech: TECH_ANTIMATTER_TORPEDOES, Cat: WeaponCatTorpedo, Ammo: 2, Size: 20, Cost: 15, DamageMin: 25, DamageMax: 25},
	{ID: 19, Tech: TECH_PROTON_TORPEDOES, Cat: WeaponCatTorpedo, Ammo: 2, Size: 30, Cost: 20, DamageMin: 40, DamageMax: 40},
	{ID: 20, Tech: TECH_PLASMA_TORPEDOES, Cat: WeaponCatTorpedo, Ammo: 2, Size: 40, Cost: 75, DamageMin: 120, DamageMax: 120},
	{ID: 21, Tech: TECH_NUCLEAR_BOMB, Cat: WeaponCatBomb, Ammo: 10, Size: 5, Cost: 3, DamageMin: 3, DamageMax: 12},
	{ID: 22, Tech: TECH_FUSION_BOMB, Cat: WeaponCatBomb, Ammo: 10, Size: 7, Cost: 5, DamageMin: 4, DamageMax: 24},
	{ID: 23, Tech: TECH_ANTIMATTER_BOMB, Cat: WeaponCatBomb, Ammo: 10, Size: 7, Cost: 6, DamageMin: 5, DamageMax: 40},
	{ID: 24, Tech: TECH_NEUTRONIUM_BOMB, Cat: WeaponCatBomb, Ammo: 10, Size: 10, Cost: 9, DamageMin: 10, DamageMax: 60},
	{ID: 25, Tech: TECH_DEATH_SPORES, Cat: WeaponCatBomb, Ammo: 10, Size: 5, Cost: 5, DamageMin: 10, DamageMax: 10},
	{ID: 26, Tech: TECH_BIOTERMINATOR, Cat: WeaponCatBomb, Ammo: 10, Size: 7, Cost: 8, DamageMin: 20, DamageMax: 20},
	{ID: 27, Tech: TECH_MAULER_DEVICE, Cat: WeaponCatBeam, Ammo: 255, Size: 50, Cost: 75, DamageMin: 100, DamageMax: 100},
	{ID: 28, Tech: TECH_ASSAULT_SHUTTLES, Cat: WeaponCatFighterBay, Ammo: 1, Size: 25, Cost: 10, DamageMin: 0, DamageMax: 0},
	{ID: 29, Tech: TECH_HEAVY_FIGHTER_BAYS, Cat: WeaponCatFighterBay, Ammo: 1, Size: 80, Cost: 50, DamageMin: 0, DamageMax: 0},
	{ID: 30, Tech: TECH_BOMBER_BAYS, Cat: WeaponCatFighterBay, Ammo: 1, Size: 60, Cost: 30, DamageMin: 0, DamageMax: 0},
	{ID: 31, Tech: TECH_FIGHTER_BAYS, Cat: WeaponCatFighterBay, Ammo: 1, Size: 30, Cost: 10, DamageMin: 6, DamageMax: 15},
	{ID: 32, Tech: TECH_STASIS_FIELD, Cat: WeaponCatSpecial, Ammo: 255, Size: 75, Cost: 75, DamageMin: 0, DamageMax: 0},
	{ID: 33, Tech: TECH_ANTIMISSILE_ROCKETS, Cat: WeaponCatSpecial, Ammo: 20, Size: 20, Cost: 5, DamageMin: 0, DamageMax: 0},
	{ID: 34, Tech: TECH_GYRO_DESTABILIZER, Cat: WeaponCatSpecial, Ammo: 255, Size: 75, Cost: 50, DamageMin: 1, DamageMax: 4},
	{ID: 35, Tech: TECH_PLASMA_WEB, Cat: WeaponCatSpecial, Ammo: 255, Size: 40, Cost: 40, DamageMin: 5, DamageMax: 25},
	{ID: 36, Tech: TECH_PULSAR, Cat: WeaponCatSpecial, Ammo: 255, Size: 50, Cost: 30, DamageMin: 2, DamageMax: 24},
	{ID: 37, Tech: TECH_BLACK_HOLE_GENERATOR, Cat: WeaponCatSpecial, Ammo: 255, Size: 150, Cost: 150, DamageMin: 6, DamageMax: 15},
	{ID: 38, Tech: TECH_STELLAR_CONVERTER, Cat: WeaponCatSpecial, Ammo: 255, Size: 500, Cost: 500, DamageMin: 400, DamageMax: 400},
	{ID: 39, Tech: TECH_TRACTOR_BEAM, Cat: WeaponCatSpecial, Ammo: 255, Size: 30, Cost: 20, DamageMin: 0, DamageMax: 0},
	{ID: 40, Tech: TECH_NONE, Cat: WeaponCatTorpedo, Ammo: 1, Size: 25, Cost: 20, DamageMin: 300, DamageMax: 300},
	{ID: 41, Tech: TECH_NONE, Cat: WeaponCatBeam, Ammo: 255, Size: 10, Cost: 10, DamageMin: 5, DamageMax: 10},
	{ID: 42, Tech: TECH_NONE, Cat: WeaponCatBeam, Ammo: 255, Size: 60, Cost: 75, DamageMin: 40, DamageMax: 80},
	{ID: 43, Tech: TECH_NONE, Cat: WeaponCatBeam, Ammo: 255, Size: 50, Cost: 75, DamageMin: 60, DamageMax: 60},
	{ID: 44, Tech: TECH_NONE, Cat: WeaponCatSpecial, Ammo: 255, Size: 250, Cost: 100, DamageMin: 10, DamageMax: 40},
	{ID: 45, Tech: TECH_NONE, Cat: WeaponCatSpecial, Ammo: 255, Size: 300, Cost: 100, DamageMin: 25, DamageMax: 50},
}

// OrigWeaponByTech 依解鎖科技查原版武器紀錄;查無(或 TECH_NONE)回 (零值, false)。
func OrigWeaponByTech(t Technology) (OrigWeapon, bool) {
	if t == TECH_NONE {
		return OrigWeapon{}, false
	}
	for _, w := range OrigWeaponTable {
		if w.Tech == t {
			return w, true
		}
	}
	return OrigWeapon{}, false
}
