package shell

import (
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// AutoDesignRole 是原版 Auto_Design_Ship_ @ 0x616A5 內部 switch 的 raw 0..7 類型。
// 名稱只描述目前由呼叫鏈證實的武器家族；不是覆蓋原始 raw 值的反組譯改名。
type AutoDesignRole uint8

const (
	AutoDesignMixed AutoDesignRole = iota
	AutoDesignFighterA
	AutoDesignMissile
	AutoDesignSpecialMissile
	AutoDesignFighterB
	AutoDesignSpecialBeam
	AutoDesignBeam
	AutoDesignMixedTheme
)

// AutoDesignLoadout 是 remake 單武器／單特殊槽模型能保存的自動設計結果。
type AutoDesignLoadout struct {
	Weapon, Armor, Shield, Special int
	Mods                           []string
	Arc                            gamedata.WeaponArc
	Ammo                           int
	RawRole                        AutoDesignRole
}

func autoWeaponAllowed(role AutoDesignRole, name string) bool {
	kind := weaponKindByName(name)
	switch role {
	case AutoDesignMissile, AutoDesignSpecialMissile:
		return kind == WeaponKindMissile
	case AutoDesignSpecialBeam:
		return kind == WeaponKindSpherical
	default:
		return kind == WeaponKindBeam || kind == WeaponKindSpherical
	}
}

func autoSpecialAllowed(role AutoDesignRole, name string) bool {
	if name == "無" {
		return true
	}
	switch role {
	case AutoDesignFighterA, AutoDesignFighterB:
		return name == "戰機庫" || name == "重戰機庫" || name == "轟炸機庫"
	case AutoDesignSpecialMissile, AutoDesignSpecialBeam, AutoDesignMixedTheme:
		return true
	default:
		return false
	}
}

func (s *GameSession) unlockedDescending(opts []Component, allowed func(string) bool) []int {
	out := make([]int, 0, len(opts))
	for i := len(opts) - 1; i >= 0; i-- {
		if s.ComponentUnlocked(opts[i]) && allowed(opts[i].Name) {
			out = append(out, i)
		}
	}
	return out
}

// AutoDesignShip 依原版已證實的八類呼叫骨架建立目前資料模型可表達的最佳合法設計。
// 原版有八個武器槽與八個特殊槽；remake 目前各一槽，因此保留角色的武器家族、最佳
// 已解鎖電腦／裝甲／護盾語意與空間守門，不能宣稱逐槽配置完全一致。
func (s *GameSession) AutoDesignShip(class string, role AutoDesignRole) (AutoDesignLoadout, bool) {
	return s.autoDesignShipFor(s.Player, class, role)
}

// autoDesignShipFor 與 AutoDesignShip 使用同一套規則，但解鎖判定取指定帝國科技。
// AI 藍圖不得借用真人的 ComponentUnlocked 狀態。
func (s *GameSession) autoDesignShipFor(ps engine.PlayerState, class string, role AutoDesignRole) (AutoDesignLoadout, bool) {
	if role > AutoDesignMixedTheme {
		role = AutoDesignMixed
	}
	unlockedDescending := func(opts []Component, allowed func(string) bool) []int {
		out := make([]int, 0, len(opts))
		for i := len(opts) - 1; i >= 0; i-- {
			if componentUnlockedFor(ps, opts[i]) && allowed(opts[i].Name) {
				out = append(out, i)
			}
		}
		return out
	}
	bestUnlocked := func(opts []Component) int {
		for i := len(opts) - 1; i >= 0; i-- {
			if componentUnlockedFor(ps, opts[i]) {
				return i
			}
		}
		return 0
	}
	weapons := unlockedDescending(WeaponOptions, func(name string) bool {
		return name != "無武裝" && autoWeaponAllowed(role, name)
	})
	if len(weapons) == 0 {
		weapons = []int{0}
	}
	specials := unlockedDescending(SpecialOptions, func(name string) bool {
		return autoSpecialAllowed(role, name)
	})
	if len(specials) == 0 {
		specials = []int{0}
	}
	armor := bestUnlocked(ArmorOptions)
	shield := bestUnlocked(ShieldOptions)
	view := *s
	view.Player = ps
	for _, weapon := range weapons {
		for _, special := range specials {
			name := WeaponOptions[weapon].Name
			arc := DefaultWeaponArc(name)
			ammo := NormalizeWeaponAmmo(name, 0)
			if view.DesignFitsWithLoadout(class, weapon, armor, shield, special, nil, arc, ammo) {
				return AutoDesignLoadout{
					Weapon: weapon, Armor: armor, Shield: shield, Special: special,
					Arc: arc, Ammo: ammo, RawRole: role,
				}, true
			}
		}
	}
	return AutoDesignLoadout{}, false
}
