package gamedata

import "math"

// weapon_arc.go:艦艇武器火線角的可證實資料與佔格／成本公式。
//
// 證據：
//   - 原版 GAM 的 ShipWeapon 結構把 arc 存成 uint8；openorion2
//     `src/gamestate.h:676-683` 保留 ARC_FWD/FWD_EXT/BACK_EXT/BACK/
//     ARC_MONSTER_360/ARC_360 的原始值。
//   - `docs/knowledge-base/manual-cht/03-combat.md` p.127-128 的手冊表給出
//     Fwd Ext、Back Ext、360 Degree 的 +25%／+25%／+50% 佔格／成本。
//   - 戰術射界的方向遮罩與 16 向朝向已由 Orion2.exe 的
//     Relative_Bearing @ 0x32AD1、Relative_Bearing_XY @ 0x32A20、
//     Move_Ship @ 0x3F5F1 交叉確認；這裡保留原始 1/2/4/8/15/16 編碼，
//     供 shell 的格子戰術消費。原始輸入與位址基準見 docs/re/weapon-arcs.md。

// WeaponArcCostPercent 回傳火線角相對於基本武器的額外佔格／成本百分比。
// 前向與後向是基準；兩種延伸弧 +25%；全向 +50%。
func WeaponArcCostPercent(arc WeaponArc) int {
	switch arc {
	case ARC_FWD_EXT, ARC_BACK_EXT:
		return 25
	case ARC_360, ARC_MONSTER_360:
		return 50
	case ARC_FWD, ARC_BACK:
		return 0
	default:
		return 0
	}
}

// WeaponArcAdjustedValue 套用火線角的佔格／成本百分比。
// 與 WeaponSpaceWithMods／WeaponCostWithMods 相同，採整數除法捨去小數。
func WeaponArcAdjustedValue(base int, arc WeaponArc) int {
	if base <= 0 {
		return base
	}
	percent := WeaponArcCostPercent(arc)
	return base + base*percent/100
}

// CombatFacingCount 是原版艦戰 heading 的離散方向數。原版 heading 存在
// combat ship record +0x23，Move_Ship @ 0x3F5F1 將角度量化成 0..15。
const CombatFacingCount = 16

// IsDirectionalWeaponArc 回報 arc 是否是四個 120 度方向弧之一。
// ARC_MONSTER_360／ARC_360 都是全向，不走 Relative_Bearing 的位元檢查。
func IsDirectionalWeaponArc(arc WeaponArc) bool {
	switch arc {
	case ARC_FWD, ARC_FWD_EXT, ARC_BACK_EXT, ARC_BACK:
		return true
	default:
		return false
	}
}

// NormalizeCombatFacing 把 heading 夾回原版的 0..15 範圍。
func NormalizeCombatFacing(facing int) int {
	facing %= CombatFacingCount
	if facing < 0 {
		facing += CombatFacingCount
	}
	return facing
}

// combatAngleDegrees 是原版 angle_to_sin @ 0x1384B9 的可讀等價物：+X 為 0 度、
// +Y 為 90 度，結果順時針增加並落在 0..359。原版用整數正切查表；這裡用標準
// 函式庫建立同一個離散角度，再由同一個 22.5 度量化規則消費，避免把 UI 像素座標
// 直接當成未定義的方向編碼。
func combatAngleDegrees(dx, dy int) int {
	if dx == 0 && dy == 0 {
		return 0
	}
	angle := math.Atan2(float64(dy), float64(dx)) * 180 / math.Pi
	if angle < 0 {
		angle += 360
	}
	result := int(math.Round(angle)) % 360
	if result < 0 {
		result += 360
	}
	return result
}

// CombatFacingForVector 對應 Move_Ship @ 0x3F5F1 的 16 向量化。
// 0=右、4=上、8=左、12=下；這個座標慣例與 tactical grid 的列方向一致。
func CombatFacingForVector(dx, dy int) int {
	angle := combatAngleDegrees(dx, dy)
	// 原版先取 -angle，再以 (4*angle+45)/90 四捨五入到 22.5 度格。
	reversed := (360 - angle) % 360
	return ((4*reversed + 45) / 90) % CombatFacingCount
}

// RelativeBearingMaskForVector 對應 Relative_Bearing_XY @ 0x32A20：以攻擊方
// heading 校正目標的絕對角度後，交給 Relative_Bearing @ 0x32AD1 產生四位元遮罩。
// 邊界條件刻意保留原版的 <=／>= 重疊行為，不能改寫成四個互斥象限。
func RelativeBearingMaskForVector(dx, dy, facing int) int {
	angle := combatAngleDegrees(dx, dy)
	facing = NormalizeCombatFacing(facing)
	angle -= 22*facing + facing/2
	angle %= 360
	if angle < 0 {
		angle += 360
	}
	return RelativeBearingMask(angle)
}

// RelativeBearingMask 對應原版 Relative_Bearing @ 0x32AD1。
func RelativeBearingMask(angle int) int {
	angle %= 360
	if angle < 0 {
		angle += 360
	}
	mask := 0
	if angle <= 60 || angle >= 300 {
		mask |= 1
	}
	if angle <= 120 || angle >= 240 {
		mask |= 2
	}
	if angle >= 60 && angle <= 300 {
		mask |= 4
	}
	if angle >= 120 && angle <= 240 {
		mask |= 8
	}
	return mask
}

// WeaponArcAllowsRelativeBearing 套用原版武器 arc raw value 與 bearing mask。
// ARC_MONSTER_360 的 raw 15 是四個方向位元全開；ARC_360 的 raw 16 是另一個
// 全向標記，兩者都必須明確視為全向。
func WeaponArcAllowsRelativeBearing(arc WeaponArc, bearingMask int) bool {
	if arc == ARC_MONSTER_360 || arc == ARC_360 {
		return true
	}
	if !IsDirectionalWeaponArc(arc) {
		return false
	}
	return int(arc)&bearingMask != 0
}
