package shell

import "github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"

// shield_facing.go:格子戰術的四面護盾狀態。
//
// 一手證據:
//   - 手冊說護盾以「per facing」吸收傷害，且傳送器要求面向攻擊方的那一面護盾失效。
//   - 原版艦艇記錄保存四個連續分面值；命中鏈會先選面再扣該面。反組譯索引:
//     Orion2.exe `sub_36810` 讀 `[ship+0x104+facing*2]`。
//
// remake 原先沒有艦身朝向；這裡以固定世界座標的四向分面承接規則。這是強推論／
// 近似，不宣稱已完全抄出原版旋轉與艦身朝向；容量邊界與傳送器前置則是有落點的規則。

// ShieldFacingForShot 回傳從 target 指向 attacker 的四向分面索引。
// 索引 0..3 只作穩定鍵，不替原版未解出的美術方向命名。
func ShieldFacingForShot(attacker, target CombatShip) int {
	dx := attacker.Col - target.Col
	dy := attacker.Row - target.Row
	if dx == 0 && dy == 0 {
		return 0
	}
	if shieldFacingAbs(dx) >= shieldFacingAbs(dy) {
		if dx >= 0 {
			return 0
		}
		return 2
	}
	if dy >= 0 {
		return 1
	}
	return 3
}

func shieldFacingAbs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

func (s *CombatShip) ensureShieldFacings() {
	if s.ShieldFacingsInitialized {
		return
	}
	for i := range s.ShieldFacingHP {
		s.ShieldFacingHP[i] = gamedata.DamageShieldCapacityForShipClass(s.ShieldReduction, s.SizeClass)
	}
	s.ShieldFacingsInitialized = true
}

// EnsureShieldFacings 建立尚未初始化的四面護盾容量；已存在的戰鬥狀態不會被重設。
func (s *CombatShip) EnsureShieldFacings() {
	s.ensureShieldFacings()
}

// ShieldFacingDown 回傳指定面是否已失效。無護盾視為該面已失效；硬化護盾不在
// 這裡判定，因為它是獨立系統，會在傳送器規則中阻擋投送。
func (s CombatShip) ShieldFacingDown(facing int) bool {
	if s.ShieldReduction <= 0 {
		return true
	}
	if facing < 0 || facing >= len(s.ShieldFacingHP) {
		return false
	}
	if !s.ShieldFacingsInitialized {
		// 舊測試與未經 StartCombat 建立的值沒有分面快照；有護盾時採保守阻擋。
		return false
	}
	return s.ShieldFacingHP[facing] <= 0
}

// ShieldReductionForFacing 將現有的每發固定減傷套到指定分面；分面失效後只
// 剩下獨立的 HardShield 減傷(由 DamageAfterShield 另外處理)。
func (s CombatShip) ShieldReductionForFacing(facing int) int {
	if s.ShieldFacingDown(facing) {
		return 0
	}
	return s.ShieldReduction
}

// WeakestShieldFacing 回傳目前護盾容量最低的分面。
// 手冊 p.157 明定戰機從目標最弱護盾面攻擊；平手時取最低索引，保留固定 RNG
// 以外的確定性。這裡只決定分面，不替未證實的艦身旋轉／方位命名。
func (s CombatShip) WeakestShieldFacing() int {
	if s.ShieldReduction <= 0 || !s.ShieldFacingsInitialized {
		return 0
	}
	weakest := 0
	for i := 1; i < len(s.ShieldFacingHP); i++ {
		if s.ShieldFacingHP[i] < s.ShieldFacingHP[weakest] {
			weakest = i
		}
	}
	return weakest
}

// FighterDamageAtWeakestShield 解算一架／一隊戰機在最弱護盾面上的護盾／結構分流。
// perCraft 現在由 fighter_attack.go 的原版第二組傷害範圍與命中結果提供；
// 本函式只負責可重用的分面容量消費，讓逐架射擊仍能在同一面狀態上累積。
func FighterDamageAtWeakestShield(target *CombatShip, alive, perCraft int) (structureDamage, shieldDamage int) {
	if target == nil || alive <= 0 || perCraft <= 0 {
		return 0, 0
	}
	target.ensureShieldFacings()
	facing := target.WeakestShieldFacing()
	reduction := target.ShieldReductionForFacing(facing)
	perCraftAfterShield := perCraft - reduction
	if perCraftAfterShield < 1 {
		perCraftAfterShield = 1
	}
	intendedShieldDamage := alive * (perCraft - perCraftAfterShield)
	if intendedShieldDamage < 0 {
		intendedShieldDamage = 0
	}
	capacity := target.ShieldFacingHP[facing]
	if intendedShieldDamage > capacity {
		intendedShieldDamage = capacity
	}
	if intendedShieldDamage > 0 {
		target.ApplyShieldDamage(facing, intendedShieldDamage)
	}
	shieldDamage = intendedShieldDamage
	structureDamage = alive*perCraft - shieldDamage
	return structureDamage, shieldDamage
}

// ApplyShieldDamage 扣除命中分面的護盾吸收容量。absorbedDamage 是一般護盾實際
// 吸收的數值，不含硬化護盾的獨立減傷；Shield Piercing 不應呼叫此方法。
func (s *CombatShip) ApplyShieldDamage(facing, absorbedDamage int) int {
	s.ensureShieldFacings()
	if facing < 0 || facing >= len(s.ShieldFacingHP) || absorbedDamage <= 0 {
		return 0
	}
	s.ShieldFacingHP[facing] -= absorbedDamage
	if s.ShieldFacingHP[facing] < 0 {
		s.ShieldFacingHP[facing] = 0
	}
	return s.ShieldFacingHP[facing]
}
