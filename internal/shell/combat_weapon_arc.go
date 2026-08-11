package shell

import "github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"

// combat_weapon_arc.go:把武器設計的 raw arc 接到格子戰術的實際射擊。
//
// 原版的戰術消費不是用「艦位相減後自行猜象限」，而是：
//   1. Relative_Bearing_XY @ 0x32A20 以攻擊艦的 heading 校正目標角度；
//   2. Relative_Bearing @ 0x32AD1 產生重疊的四位元方向遮罩；
//   3. 射擊端以 `(bearing & weaponArc) != 0` 判斷合法性。
//
// CombatShip 的 Facing 是 remake 對原版 combat record +0x23 的 0..15
// heading 映射。快速結算沒有格位／朝向，所以不在這裡套用這條規則。

// WeaponArcAllowsCombatShot 回報 attacker 是否能以目前艦首朝向攻擊 target。
// 舊測試與外部建構的 CombatShip 若沒有任何武器名稱與 arc，保留舊行為並視為
// 未知資料；真正由 StartCombat 建立的船一定會填入合法 arc，因此不會繞過規則。
func WeaponArcAllowsCombatShot(attacker, target CombatShip) bool {
	arc := attacker.WeaponArc
	if !gamedata.IsDirectionalWeaponArc(arc) && arc != gamedata.ARC_360 && arc != gamedata.ARC_MONSTER_360 {
		if attacker.WeaponName == "" {
			return true
		}
		arc = NormalizeWeaponArc(attacker.WeaponName, arc)
	}
	if arc == gamedata.ARC_360 || arc == gamedata.ARC_MONSTER_360 {
		return true
	}
	bearing := gamedata.RelativeBearingMaskForVector(
		target.Col-attacker.Col,
		target.Row-attacker.Row,
		attacker.Facing,
	)
	return gamedata.WeaponArcAllowsRelativeBearing(arc, bearing)
}
