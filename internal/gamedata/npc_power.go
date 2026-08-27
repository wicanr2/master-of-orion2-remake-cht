package gamedata

// OriginalNPCPowerMount 是 sub_5EF4B 消費的原版 8-byte 武器槽最小形狀。
type OriginalNPCPowerMount struct {
	WeaponID     int
	WorkingCount int
	RawMods      uint16
	Ammo         int
}

// OriginalNPCShipPowerInput 是單艦對單一 observer 的方向國力輸入。
// BeamAttack／ObserverDefense 與 RemainingDurability 均由原版相鄰 helper 的 typed producer 提供。
type OriginalNPCShipPowerInput struct {
	Mounts                  []OriginalNPCPowerMount
	BeamAttack              int
	ObserverDefense         int
	ObserverWeaponReduction int
	OwnerBestComputer       int
	DesignComputer          int
	RemainingDurability     int
	BestBeamWeaponID        int
	BestBombWeaponID        int
}

var originalNPCPowerModPercent = [...]struct {
	bit     uint16
	percent int
}{
	{0x0002, 100}, {0x0004, -50}, {0x0020, 25}, {0x0080, 50},
	{0x0100, 100}, {0x0200, 100}, {0x0400, 25}, {0x0800, 25},
	{0x1000, 25}, {0x2000, 300}, {0x4000, 50},
}

var originalNPCComputerWeaponReduction = [...]int{0, 1, 3, 5, 7, 10}
var originalNPCDriveDefenseSpeed = [...]int{0, 6, 8, 10, 12, 14, 16, 2}
var originalNPCHullBase = [...]int{4, 10, 30, 50, 80, 150}
var originalNPCArmorPercent = [...]int{-100, 0, 100, 300, 500, 700, 900, 0}

// OriginalNPCShipDurability 對應 sub_582BF、sub_58425 與 sub_5F2F6 的一般艦艇分支。
// 原版的 structure 與 armor 共用艦級基礎表及裝甲科技百分比；強化船體／重型裝甲
// 各自把對應軌乘三，最後才扣除 129-byte Ship 紀錄中的兩條獨立損傷。
func OriginalNPCShipDurability(size, armor int, reinforcedHull, heavyArmor bool, armorDamage, structureDamage int) (int, bool) {
	if size < 0 || size >= len(originalNPCHullBase) || armor < 0 || armor >= len(originalNPCArmorPercent) ||
		armorDamage < 0 || structureDamage < 0 {
		return 0, false
	}
	base := originalNPCHullBase[size]
	scaled := base + base*originalNPCArmorPercent[armor]/100
	structure, armorHP := scaled, scaled
	if reinforcedHull {
		structure *= 3
	}
	if heavyArmor {
		armorHP *= 3
	}
	remaining := structure + armorHP - structureDamage - armorDamage
	if remaining < 0 {
		return 0, false
	}
	return remaining, true
}

// OriginalNPCShipBeamAttack 對應 sub_54D80／sub_54E5B 寫回給 sub_5EF4B 的
// attack 輸出：電腦、可用戰鬥掃描器、艦艇軍官 Weaponry、艦員與種族艦攻。
func OriginalNPCShipBeamAttack(computer, crewLevel, officerWeaponry, racialShipAttack int, battleScanner bool) (int, bool) {
	if computer < 0 || computer >= len(computerBonusTable) || crewLevel < 0 || crewLevel >= len(shipCrewOffenseBonuses) {
		return 0, false
	}
	result := ComputerBonus(computer) + ShipCrewOffenseBonus(crewLevel) + officerWeaponry + racialShipAttack
	if battleScanner {
		result += 50
	}
	return result, true
}

// OriginalNPCComputerWeaponReduction 對應 sub_5F871 對一般帝國的 word_17F6C1 查表。
func OriginalNPCComputerWeaponReduction(computer int) (int, bool) {
	if computer < 0 || computer >= len(originalNPCComputerWeaponReduction) {
		return 0, false
	}
	return originalNPCComputerWeaponReduction[computer], true
}

// OriginalNPCObserverDefense 對應 sub_5EAE9 的一般帝國分支。
func OriginalNPCObserverDefense(drive int, racialShipDefense int, transDimensional bool) (int, bool) {
	if drive < 0 || drive >= len(originalNPCDriveDefenseSpeed) {
		return 0, false
	}
	result := 5*originalNPCDriveDefenseSpeed[drive] + racialShipDefense
	if transDimensional {
		result += 20
	}
	return result, true
}

// OriginalNPCPowerModifier 對應 sub_5EE27；結果以原版 word 上限 64000 封頂。
func OriginalNPCPowerModifier(value int, raw uint16) (int, bool) {
	if value < 0 {
		return 0, false
	}
	percent := 100
	for _, entry := range originalNPCPowerModPercent {
		if raw&entry.bit != 0 {
			percent += entry.percent
		}
	}
	value = value * percent / 100
	if value > 64000 {
		value = 64000
	}
	return value, true
}

// OriginalNPCDurabilityFactor 對應 sub_5EED4 的分段壓縮。
func OriginalNPCDurabilityFactor(remaining int) int {
	if remaining > 100 {
		return 100 + (remaining-100)/4
	}
	return remaining + (100-remaining)/10
}

// OriginalNPCShipPower 對應 sub_5EF4B 的玩家艦艇分支。最多只讀前八槽；非法 raw ID
// 或艦載機缺少其 owner 的最佳武器映射時失敗即關閉。
func OriginalNPCShipPower(in OriginalNPCShipPowerInput) (int, bool) {
	if in.BeamAttack < 0 || in.ObserverDefense < 0 || in.ObserverWeaponReduction < 0 || in.OwnerBestComputer < 0 ||
		in.DesignComputer < 0 || in.RemainingDurability < 0 {
		return 0, false
	}
	total := 0
	for slot, mount := range in.Mounts {
		if slot >= 8 {
			break
		}
		if mount.WorkingCount < 0 {
			return 0, false
		}
		if mount.WeaponID <= 0 || mount.WorkingCount == 0 {
			continue
		}
		weaponID, count := mount.WeaponID, mount.WorkingCount
		if mount.RawMods&0x0040 == 0 && weaponID != 13 && weaponID != 34 {
			switch weaponID {
			case 31:
				weaponID, count = in.BestBeamWeaponID, count*3
			case 29, 30:
				weaponID = in.BestBombWeaponID
			case 28:
				weaponID, count = 27, count*3
			case 36:
				count *= 3
			}
		}
		weapon, ok := OrigWeaponByID(weaponID)
		if !ok || weaponID <= 0 {
			return 0, false
		}
		base := weapon.DamageMax
		if mount.RawMods&0x0040 == 0 && mount.WeaponID != 13 && mount.WeaponID != 34 {
			if mount.RawMods&0x0002 != 0 {
				base += (50*base + 50) / 100
			} else if mount.RawMods&0x0004 != 0 {
				base = (50*base + 50) / 100
			}
			base -= in.ObserverWeaponReduction
			if base < 0 {
				base = 0
			}
		}
		value := base * count
		switch weapon.Cat {
		case WeaponCatBeam:
			hit := in.BeamAttack + 50 - in.ObserverDefense
			if mount.RawMods&0x0010 != 0 {
				hit += 25
			}
			if hit < 10 {
				hit = 10
			} else if hit > 100 {
				hit = 100
			}
			value = value * hit / 100
		case WeaponCatMissile:
			if mount.Ammo == 2 {
				value = 3 * value / 4
			}
		case WeaponCatTorpedo:
			value = 3 * value / 4
		case WeaponCatBomb:
			value = 0
		}
		value, ok = OriginalNPCPowerModifier(value, mount.RawMods)
		if !ok {
			return 0, false
		}
		total += value
	}
	computerFactor := 100
	if delta := in.OwnerBestComputer - in.DesignComputer; delta < 0 {
		computerFactor = 100 - 10*(-delta)
	}
	if computerFactor < 10 {
		computerFactor = 10
	}
	total = total * computerFactor / 100
	factor := OriginalNPCDurabilityFactor(in.RemainingDurability)
	result := total * factor / 100
	if total > 0 && result < 1 {
		result = 1
	}
	return result, true
}
