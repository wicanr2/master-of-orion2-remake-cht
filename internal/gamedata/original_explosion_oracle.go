package gamedata

// original_explosion_oracle.go 是 IDA raw 0x3868F、0x39985、0x40C2A、0x494A8
// 已證實常數的純函式層。未知的 table／sub_4B0D3 條件都以參數保留，不把
// 反編譯器猜出的全域名稱當成已證實語意。

const (
	OriginalShipExplosionRollRange    = 0xC9  // Random(201)
	OriginalShipExplosionRollOffset   = 0x4A  // +74
	OriginalColonyExplosionRollRange  = 0x191 // Random(401)
	OriginalColonyExplosionRollOffset = 0x95  // +149
	OriginalExplosionChainStep        = 0x14  // each spherical step subtracts 20
)

func boundedOriginalRoll(roll, rangeSize int) int {
	if rangeSize <= 0 {
		return 0
	}
	if roll < 0 {
		return 0
	}
	if roll >= rangeSize {
		return rangeSize - 1
	}
	return roll
}

// OriginalShipExplosionDamageRoll 對應 raw 0x3868F 的 Random(0xC9)+0x4A。
func OriginalShipExplosionDamageRoll(roll int) int {
	return boundedOriginalRoll(roll, OriginalShipExplosionRollRange) + OriginalShipExplosionRollOffset
}

// OriginalColonyExplosionPotentialRoll 對應 raw 0x3868F 的殖民地路徑
// Random(0x191)+0x95。
func OriginalColonyExplosionPotentialRoll(roll int) int {
	return boundedOriginalRoll(roll, OriginalColonyExplosionRollRange) + OriginalColonyExplosionRollOffset
}

// OriginalExplosionChainNextPotential 對應 raw 0x3868F／0x39985 的每步減 0x14。
func OriginalExplosionChainNextPotential(potential int) int {
	if potential <= 0 {
		return 0
	}
	if potential <= OriginalExplosionChainStep {
		return 0
	}
	return potential - OriginalExplosionChainStep
}

// OriginalEngineExplosionBasePotential 對應 raw 0x494A8：
// sizeHitPoints[size] * 5 * (maxEngineLevel + 1)，而 size>=5 時最後的 +1
// 變成 +2。sizeHitPoints 是 raw word_17F6C1 的外部表，刻意不在這裡重抄未知值。
func OriginalEngineExplosionBasePotential(size, maxEngineLevel int, sizeHitPoints []int) int {
	if size < 0 || size >= len(sizeHitPoints) || maxEngineLevel < 0 {
		return 0
	}
	engineFactor := 1
	if size >= 5 {
		engineFactor = 2
	}
	return sizeHitPoints[size] * 5 * (maxEngineLevel + engineFactor)
}

// OriginalEngineExplosionRaw14Branch 對應 raw 0x494A8 的 `3*damage/2` 分支。
// enabled 的實際判定由 sub_4B0D3(ship,0x14) 決定；該科技／旗標尚未完成
// 語意命名，所以由呼叫端傳入，不在此猜名稱。
func OriginalEngineExplosionRaw14Branch(base int, enabled bool) int {
	if base <= 0 || !enabled {
		if base < 0 {
			return 0
		}
		return base
	}
	return 3 * base / 2
}

// OriginalExplosionDamageConsumer 對應 raw 0x40C2A：targetType=7 的船型
// 先取 damage/4，其餘船型扣除 raw size resistance；結果不低於零。resistance
// 由 word_17F6C1[targetType] 之類的 raw 表提供。
func OriginalExplosionDamageConsumer(damage, targetType, resistance int) int {
	if damage <= 0 {
		return 0
	}
	if targetType == 7 {
		return damage / 4
	}
	if resistance < 0 {
		resistance = 0
	}
	if damage <= resistance {
		return 0
	}
	return damage - resistance
}
