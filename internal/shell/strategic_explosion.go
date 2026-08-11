package shell

import "github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"

// strategicExplosionCollateral 是事件爆炸對倖存艦造成的次級結構損傷。
// RawDamage 保留 oracle 的下游消費結果；AppliedDamage 是 remake 艦級 Damage
// 欄位的尺度化寫入值，兩者不能混稱為原版結構欄位。
type strategicExplosionCollateral struct {
	Name          string
	RawDamage     int
	AppliedDamage int
}

// strategicExplosionResult 是「艦船爆炸」事件的戰略層結算結果。
// Potential／RawDamage 直接來自已追回的純 oracle；Collateral 是 remake 近似消費。
type strategicExplosionResult struct {
	Lost        Ship
	Potential   int
	Collateral  []strategicExplosionCollateral
	ChainRemain int
}

const (
	// 原版爆炸值約為 74..274，而 Ship.Damage 是 remake 的艦級 HP 尺度；除以
	// 20 是兩個資料模型間的明示 adapter，不是原版的 /20 公式。
	strategicExplosionDamageScale     = 20
	strategicExplosionCollateralLimit = 3
)

// strategicExplosionTargetType 把 remake 艦級映到 oracle 的 target type。
// 原版 raw type→艦級完整表仍未知；這個單調映射只供 strategic consumer 使用，
// 不回寫 Ship 的 raw picture／艦型欄位。
func strategicExplosionTargetType(sh Ship) int {
	switch sh.Class {
	case "偵察艦", "殖民船", OutpostShipClass:
		return 1
	case "巡防艦", "護衛艦":
		return 2
	case "驅逐艦":
		return 3
	case "巡洋艦":
		return 4
	case "戰艦":
		return 5
	case "泰坦":
		return 6
	case "末日之星":
		return 7
	default:
		return 2
	}
}

// strategicExplosionResistance 是 raw resistance table 不可直接對回 remake 艦級
// 的保守尺度化。保留在單一 helper，日後有完整 raw table 時可替換而不動事件流程。
func strategicExplosionResistance(sh Ship) int {
	strength := shipStrength(sh.Class)
	if strength < 1 {
		strength = 1
	}
	return strength * 5
}

// resolveStrategicShipExplosion 將艦船爆炸事件接到已追回的爆炸 oracle：
// ① Random(201)+74 選主爆炸勢能；② 移除主艦；③ 每個連鎖候選消費 20 點勢能；
// ④ 以 target type／resistance 呼叫下游 consumer，再把結果尺度化寫入倖存艦 Damage。
//
// 主艦移除仍保留原 remake 的安全護欄(至少留一艘)。連鎖不直接擊沉倖存艦，避免
// 未追回 raw fleet／colony record 時把「戰略事件」誤變成一個不可逆的全滅骰；
// Damage 會由既有 repair／戰鬥路徑消費。這是允許的 remake 近似，不宣稱 raw
// sub_39985 的完整 flag／行星表已經重建。
func (s *GameSession) resolveStrategicShipExplosion() (strategicExplosionResult, bool) {
	var out strategicExplosionResult
	if s == nil || s.ShipCount() <= 1 {
		return out, false
	}
	s.eventRandForTest()
	out.Potential = gamedata.OriginalShipExplosionDamageRoll(
		s.eventRand.Intn(gamedata.OriginalShipExplosionRollRange))
	lost, ok := s.removeShipGlobal(s.eventRand.Intn(s.ShipCount()))
	if !ok {
		return strategicExplosionResult{}, false
	}
	out.Lost = lost

	potential := gamedata.OriginalExplosionChainNextPotential(out.Potential)
	for fleetIdx := range s.Fleets {
		for shipIdx := range s.Fleets[fleetIdx].Ships {
			if potential <= 0 || len(out.Collateral) >= strategicExplosionCollateralLimit {
				break
			}
			sh := &s.Fleets[fleetIdx].Ships[shipIdx]
			rawDamage := gamedata.OriginalExplosionDamageConsumer(
				potential, strategicExplosionTargetType(*sh), strategicExplosionResistance(*sh))
			potential = gamedata.OriginalExplosionChainNextPotential(potential)
			if rawDamage <= 0 {
				continue
			}
			applied := rawDamage / strategicExplosionDamageScale
			if applied < 1 {
				applied = 1
			}
			maxDamage := shipMaxHP(*sh) - ShipDamageFloorHP
			if maxDamage < 0 {
				maxDamage = 0
			}
			before := sh.Damage
			if before < 0 {
				before = 0
				sh.Damage = 0
			}
			// 舊存檔或人工編輯可能已把 Damage 留在上限之外；不要為了
			// 套用事件而把既有損傷倒退到 maxDamage，讓爆炸 consumer
			// 只會增加損傷、永不修復艦艇。
			if before >= maxDamage {
				continue
			}
			sh.Damage = before + applied
			if sh.Damage > maxDamage {
				sh.Damage = maxDamage
			}
			if sh.Damage > before {
				out.Collateral = append(out.Collateral, strategicExplosionCollateral{
					Name: sh.Name, RawDamage: rawDamage, AppliedDamage: sh.Damage - before,
				})
			}
		}
		if potential <= 0 || len(out.Collateral) >= strategicExplosionCollateralLimit {
			break
		}
	}
	out.ChainRemain = potential
	return out, true
}
