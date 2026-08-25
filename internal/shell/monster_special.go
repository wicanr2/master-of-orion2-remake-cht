package shell

import "github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"

// 事件怪物專用武器名稱。它們不是玩家可設計元件，但必須與 WeaponKindForName
// 共用同一字串，才能讓原版 blueprint 的逐槽武器進入專用 runtime。
const (
	plasmaFluxName   = "電漿通量"
	causticSlimeName = "腐蝕黏液"

	// PlasmaFluxRadiusPixels 由 sub_ADE18 讀取 CMBTSFX asset 6 半寬；實際資產為 192px。
	PlasmaFluxRadiusPixels = 96
	PlasmaFluxCellPixels   = 20
)

// IsPlasmaFluxName／IsCausticSlimeName 供格子戰術 adapter 辨識兩項怪物專武。
func IsPlasmaFluxName(name string) bool   { return name == plasmaFluxName }
func IsCausticSlimeName(name string) bool { return name == causticSlimeName }

// PlasmaFluxInRange 以 remake 格中心近似原版像素中心，保留歐氏平方判定。
func PlasmaFluxInRange(colDelta, rowDelta int) bool {
	x := colDelta * PlasmaFluxCellPixels
	y := rowDelta * PlasmaFluxCellPixels
	return x*x+y*y <= PlasmaFluxRadiusPixels*PlasmaFluxRadiusPixels
}

// PlasmaFluxAttenuatedDamage 對應 sub_ADE18 的距離平方衰減，圈內最低為 1。
func PlasmaFluxAttenuatedDamage(base, colDelta, rowDelta int) int {
	if base < 1 || !PlasmaFluxInRange(colDelta, rowDelta) {
		return 0
	}
	x := colDelta * PlasmaFluxCellPixels
	y := rowDelta * PlasmaFluxCellPixels
	distanceSquared := x*x + y*y
	radiusSquared := PlasmaFluxRadiusPixels * PlasmaFluxRadiusPixels
	damage := base * (100 - 100*distanceSquared/radiusSquared) / 100
	if damage < 1 {
		damage = 1
	}
	return damage
}

// PlasmaFluxSizeDamage 對應 combat record +0x25+1 次的雙擲值尺寸分段。
// roll 必須回傳 1..limit；nil 時採每段最低 1，供安全 fallback。
func PlasmaFluxSizeDamage(attenuated int, sizeClass gamedata.CombatShipClass, roll func(limit int) int) int {
	if attenuated < 1 {
		return 0
	}
	total := 0
	for segment := 0; segment < int(sizeClass)+1; segment++ {
		first := 1
		if roll != nil {
			first = roll(attenuated)
		}
		if first <= 1 {
			total++
			continue
		}
		value := 1
		if roll != nil {
			value = roll(attenuated)
		}
		if value < 1 {
			value = 1
		}
		total += value
	}
	return total
}

// PlasmaFluxFighterCasualties 對應 sub_ADE18 的 category 4 分支。avoidRoll 是
// 1-based Random(2)；結果 1 時整隊避開。未避開後，每架各擲一次 Random(100)。
func PlasmaFluxFighterCasualties(alive, hpEach, attenuated, avoidRoll int, roll func(limit int) int) int {
	if alive <= 0 || attenuated <= 0 || avoidRoll == 1 {
		return 0
	}
	if hpEach < 1 {
		hpEach = 1
	}
	chance := 25 * attenuated / hpEach
	killed := 0
	for craft := 0; craft < alive; craft++ {
		value := 100
		if roll != nil {
			value = roll(100)
		}
		if value <= chance {
			killed++
		}
	}
	return killed
}

// AddCausticSlimeStrength 模擬 combat record +0x43 的累加寫入。
func AddCausticSlimeStrength(target *CombatShip, strength int) {
	if target == nil || strength <= 0 {
		return
	}
	target.CausticSlimeStrength += strength
}

// TickCausticSlime 模擬 sub_4A5CE：同一強度依序送入四個護盾面，讓每面的
// 護盾不足量繼續穿入裝甲／結構，最後把狀態強度減 5 並夾到零。
func TickCausticSlime(target *CombatShip) (structureDamage int) {
	if target == nil || target.HP <= 0 || target.CausticSlimeStrength <= 0 {
		return 0
	}
	target.EnsureShieldFacings()
	for facing := range target.ShieldFacingHP {
		shot := ResolveSphericalShot(target.CausticSlimeStrength,
			target.ShieldReductionForFacing(facing), target.ArmorHP, target.HardShield, false)
		target.ApplyShieldDamage(facing, shot.ShieldDamage)
		target.ArmorHP = shot.RemainingArmorHP
		target.HP -= shot.DamageToStructure
		structureDamage += shot.DamageToStructure
		if target.HP <= 0 {
			break
		}
	}
	target.CausticSlimeStrength -= 5
	if target.CausticSlimeStrength < 0 {
		target.CausticSlimeStrength = 0
	}
	return structureDamage
}
