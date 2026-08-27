package shell

import (
	"encoding/json"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// PlayerShipDesignCount 是原版每位玩家的 0..5 六筆 99-byte 艦體設計。
const PlayerShipDesignCount = 6

var playerShipDesignClasses = [...]string{"巡防艦", "驅逐艦", "巡洋艦", "戰艦", "泰坦", "末日之星"}

// ShipBlueprint 是玩家可反覆編輯、建造的持久設計。Weapon/Armor/Shield/Special 是
// 舊存檔與相容 API 使用的元件索引；WeaponMounts/SpecialIDs 保留原版多槽資料形狀。
type ShipBlueprint struct {
	Class                          string
	RawRole                        AutoDesignRole
	Weapon, Armor, Shield, Special int
	Mods                           []string
	Arc                            gamedata.WeaponArc
	Ammo                           int
	WeaponMounts                   []ShipWeaponMount  `json:"weaponMounts,omitempty"`
	SpecialIDs                     []int              `json:"specialIDs,omitempty"`
	Specials                       []ShipSpecialMount `json:"specials,omitempty"`
}

func blueprintFromLoadout(class string, loadout AutoDesignLoadout) ShipBlueprint {
	w := pick(WeaponOptions, loadout.Weapon)
	mods := append([]string(nil), loadout.Mods...)
	return ShipBlueprint{
		Class: class, RawRole: loadout.RawRole,
		Weapon: loadout.Weapon, Armor: loadout.Armor, Shield: loadout.Shield, Special: loadout.Special,
		Mods: mods, Arc: loadout.Arc, Ammo: loadout.Ammo,
		WeaponMounts: []ShipWeaponMount{originalWeaponMount(w, loadout.Mods, loadout.Arc, loadout.Ammo, 1)},
		Specials:     []ShipSpecialMount{specialMountFromOption(loadout.Special)},
	}
}

// EnsureShipDesigns 維持六艦體設計庫不變量。舊存檔缺少的尾端設計依讀回後的當前科技
// 產生合法 mixed 模板；既有設計不會被重建或覆蓋。
func (s *GameSession) EnsureShipDesigns() {
	if len(s.ShipDesigns) > PlayerShipDesignCount {
		s.ShipDesigns = s.ShipDesigns[:PlayerShipDesignCount]
	}
	for len(s.ShipDesigns) < PlayerShipDesignCount {
		i := len(s.ShipDesigns)
		loadout, ok := s.AutoDesignShip(playerShipDesignClasses[i], AutoDesignMixed)
		if !ok {
			loadout = AutoDesignLoadout{RawRole: AutoDesignMixed, Arc: gamedata.ARC_FWD}
		}
		s.ShipDesigns = append(s.ShipDesigns, blueprintFromLoadout(playerShipDesignClasses[i], loadout))
	}
	for i := range s.ShipDesigns {
		if s.ShipDesigns[i].Class == "" {
			s.ShipDesigns[i].Class = playerShipDesignClasses[i]
		}
		if len(s.ShipDesigns[i].Specials) == 0 {
			s.ShipDesigns[i].Specials = []ShipSpecialMount{specialMountFromOption(s.ShipDesigns[i].Special)}
		}
	}
}

// ShipDesign 回傳第 hull 筆設計的值複本；slice 亦深複製，呼叫端必須經 SetShipDesign 更新。
func (s *GameSession) ShipDesign(hull int) (ShipBlueprint, bool) {
	s.EnsureShipDesigns()
	if hull < 0 || hull >= len(s.ShipDesigns) {
		return ShipBlueprint{}, false
	}
	out := s.ShipDesigns[hull]
	out.Mods = append([]string(nil), out.Mods...)
	out.WeaponMounts = cloneWeaponMounts(out.WeaponMounts)
	out.SpecialIDs = append([]int(nil), out.SpecialIDs...)
	out.Specials = cloneSpecialMounts(out.Specials)
	return out, true
}

// SetShipDesign 寫回一筆設計。Class 由 hull 索引決定，避免 UI 或舊存檔把六筆順序漂移。
func (s *GameSession) SetShipDesign(hull int, design ShipBlueprint) bool {
	s.EnsureShipDesigns()
	if hull < 0 || hull >= PlayerShipDesignCount {
		return false
	}
	design.Class = playerShipDesignClasses[hull]
	design.Mods = append([]string(nil), design.Mods...)
	design.WeaponMounts = cloneWeaponMounts(design.WeaponMounts)
	design.SpecialIDs = append([]int(nil), design.SpecialIDs...)
	design.Specials = cloneSpecialMounts(design.Specials)
	s.ShipDesigns[hull] = design
	return true
}

// SetShipDesignLoadout 只更新舊版相容 API 可表達的欄位與第一個 mount；既有其餘 raw 槽保留。
func (s *GameSession) SetShipDesignLoadout(hull int, loadout AutoDesignLoadout) bool {
	return s.SetShipDesignMountLoadout(hull, 0, loadout)
}

// SetShipDesignMountLoadout 更新指定武器槽；裝甲／護盾／特殊仍是整筆 blueprint 共用欄位。
func (s *GameSession) SetShipDesignMountLoadout(hull, mountIndex int, loadout AutoDesignLoadout) bool {
	design, ok := s.ShipDesign(hull)
	if !ok {
		return false
	}
	design.RawRole = loadout.RawRole
	design.Armor = loadout.Armor
	design.Shield = loadout.Shield
	w := pick(BuildWeaponOptions(s.RuleProfile), loadout.Weapon)
	mount := originalWeaponMount(w, loadout.Mods, loadout.Arc, loadout.Ammo, 1)
	if len(design.WeaponMounts) == 0 {
		design.WeaponMounts = []ShipWeaponMount{mount}
	}
	if mountIndex < 0 || mountIndex >= len(design.WeaponMounts) {
		return false
	}
	old := design.WeaponMounts[mountIndex]
	mount.MaxCount, mount.WorkingCount = old.MaxCount, old.WorkingCount
	if mount.MaxCount < 1 {
		mount.MaxCount = 1
	}
	if mount.WorkingCount < 1 || mount.WorkingCount > mount.MaxCount {
		mount.WorkingCount = mount.MaxCount
	}
	design.WeaponMounts[mountIndex] = mount
	if mountIndex == 0 {
		design.Weapon = loadout.Weapon
		design.Mods = append([]string(nil), loadout.Mods...)
		design.Arc, design.Ammo = loadout.Arc, loadout.Ammo
	}
	return s.SetShipDesign(hull, design)
}

// SetShipDesignSpecialMount 更新指定特殊裝置槽並同步第一槽相容欄位。
func (s *GameSession) SetShipDesignSpecialMount(hull, mountIndex, special int) bool {
	design, ok := s.ShipDesign(hull)
	if !ok {
		return false
	}
	if len(design.Specials) == 0 {
		design.Specials = []ShipSpecialMount{specialMountFromOption(design.Special)}
	}
	if mountIndex < 0 || mountIndex >= len(design.Specials) || special < 0 || special >= len(SpecialOptions) {
		return false
	}
	candidate := specialMountFromOption(special)
	for i, mount := range design.Specials {
		if i != mountIndex && candidate.Name != "無" && mount.Name == candidate.Name {
			return false
		}
	}
	design.Specials[mountIndex] = candidate
	if mountIndex == 0 {
		design.Special = special
	}
	design.SpecialIDs = specialIDsFromMounts(design.Specials)
	return s.SetShipDesign(hull, design)
}

func (s *GameSession) AddShipDesignSpecialMount(hull int) (int, bool) {
	design, ok := s.ShipDesign(hull)
	if !ok || len(design.Specials) >= 8 {
		return -1, false
	}
	design.Specials = append(design.Specials, specialMountFromOption(0))
	idx := len(design.Specials) - 1
	return idx, s.SetShipDesign(hull, design)
}

func (s *GameSession) RemoveShipDesignSpecialMount(hull, mountIndex int) bool {
	design, ok := s.ShipDesign(hull)
	if !ok || len(design.Specials) <= 1 || mountIndex < 0 || mountIndex >= len(design.Specials) {
		return false
	}
	design.Specials = append(design.Specials[:mountIndex], design.Specials[mountIndex+1:]...)
	first, known := specialOptionIndex(design.Specials[0].Name)
	if !known {
		return false
	}
	design.Special = first
	design.SpecialIDs = specialIDsFromMounts(design.Specials)
	return s.SetShipDesign(hull, design)
}

// NextUnlockedSpecialForDesign 循環到下一個已解鎖且未被其他槽占用的特殊裝置。
func (s *GameSession) NextUnlockedSpecialForDesign(hull, mountIndex, current int) int {
	design, ok := s.ShipDesign(hull)
	if !ok {
		return current
	}
	for step := 1; step <= len(SpecialOptions); step++ {
		idx := (current + step) % len(SpecialOptions)
		if !s.ComponentUnlocked(SpecialOptions[idx]) {
			continue
		}
		name := SpecialOptions[idx].Name
		duplicate := false
		for i, mount := range design.Specials {
			if i != mountIndex && name != "無" && mount.Name == name {
				duplicate = true
				break
			}
		}
		if !duplicate {
			return idx
		}
	}
	return current
}

// AddShipDesignMount 在目前設計尾端加入一個可編輯槽，最多八槽。
func (s *GameSession) AddShipDesignMount(hull, copyFrom int) (int, bool) {
	design, ok := s.ShipDesign(hull)
	if !ok || len(design.WeaponMounts) >= 8 {
		return -1, false
	}
	if len(design.WeaponMounts) == 0 {
		design.WeaponMounts = []ShipWeaponMount{{RawType: -1, Name: pick(WeaponOptions, design.Weapon).Name,
			MaxCount: 1, WorkingCount: 1, Arc: design.Arc, Ammo: design.Ammo,
			Attack: pick(BuildWeaponOptions(s.RuleProfile), design.Weapon).Value}}
	}
	if copyFrom < 0 || copyFrom >= len(design.WeaponMounts) {
		copyFrom = 0
	}
	mount := design.WeaponMounts[copyFrom]
	mount.RawType, mount.RawMods = -1, 0
	mount.MaxCount, mount.WorkingCount = 1, 1
	design.WeaponMounts = append(design.WeaponMounts, mount)
	idx := len(design.WeaponMounts) - 1
	return idx, s.SetShipDesign(hull, design)
}

// RemoveShipDesignMount 刪除指定槽；設計至少保留一槽。
func (s *GameSession) RemoveShipDesignMount(hull, mountIndex int) bool {
	design, ok := s.ShipDesign(hull)
	if !ok || len(design.WeaponMounts) <= 1 || mountIndex < 0 || mountIndex >= len(design.WeaponMounts) {
		return false
	}
	design.WeaponMounts = append(design.WeaponMounts[:mountIndex], design.WeaponMounts[mountIndex+1:]...)
	return s.syncBlueprintCompatibility(hull, design)
}

// AdjustShipDesignMountCount 在原版單槽 byte 可表達的 1..99 內調整武器數量。
func (s *GameSession) AdjustShipDesignMountCount(hull, mountIndex, delta int) bool {
	design, ok := s.ShipDesign(hull)
	if !ok || mountIndex < 0 || mountIndex >= len(design.WeaponMounts) {
		return false
	}
	n := design.WeaponMounts[mountIndex].MaxCount + delta
	if n < 1 {
		n = 1
	}
	if n > 99 {
		n = 99
	}
	design.WeaponMounts[mountIndex].MaxCount = n
	design.WeaponMounts[mountIndex].WorkingCount = n
	return s.SetShipDesign(hull, design)
}

func (s *GameSession) syncBlueprintCompatibility(hull int, design ShipBlueprint) bool {
	if len(design.WeaponMounts) == 0 {
		return false
	}
	m := design.WeaponMounts[0]
	idx, ok := weaponOptionIndex(m.Name)
	if !ok {
		return false
	}
	design.Weapon, design.Arc, design.Ammo = idx, m.Arc, m.Ammo
	design.Mods = append([]string(nil), m.Mods...)
	return s.SetShipDesign(hull, design)
}

func weaponOptionIndex(name string) (int, bool) {
	for i, c := range WeaponOptions {
		if c.Name == name {
			return i, true
		}
	}
	return 0, false
}

func blueprintMountCount(m ShipWeaponMount) int {
	if m.MaxCount > 0 {
		return m.MaxCount
	}
	if m.WorkingCount > 0 {
		return m.WorkingCount
	}
	return 0
}

// BlueprintDesignCost 回傳完整多槽設計成本；ok=false 表示含未知 raw 武器／mods。
func (s *GameSession) BlueprintDesignCost(design ShipBlueprint) (cost int, ok bool) {
	base := s.DesignCostWithLoadout(design.Class, 0, design.Armor, design.Shield, 0,
		nil, gamedata.ARC_FWD, NormalizeWeaponAmmo(pick(WeaponOptions, 0).Name, 0))
	cost = base
	mounts := design.WeaponMounts
	if len(mounts) == 0 {
		mounts = []ShipWeaponMount{{Name: pick(WeaponOptions, design.Weapon).Name, MaxCount: 1,
			WorkingCount: 1, Arc: design.Arc, Ammo: design.Ammo}}
	}
	for i, mount := range mounts {
		count := blueprintMountCount(mount)
		if count == 0 || mount.Name == "" {
			continue
		}
		weapon, known := weaponOptionIndex(mount.Name)
		if !known || (mount.RawMods != 0 && len(mount.Mods) == 0) {
			return 0, false
		}
		mods := mount.Mods
		if i == 0 && len(mods) == 0 {
			mods = design.Mods
		}
		one := s.DesignCostWithLoadout(design.Class, weapon, 0, 0, 0, mods,
			mount.Arc, mount.Ammo) - ShipCost(design.Class)
		cost += one * count
	}
	specials := design.Specials
	if len(specials) == 0 {
		specials = []ShipSpecialMount{specialMountFromOption(design.Special)}
	}
	if !specialNamesUnique(specials) {
		return 0, false
	}
	for _, mount := range specials {
		if mount.Name == "" || mount.Name == "無" {
			continue
		}
		special, known := specialOptionIndex(mount.Name)
		if !known || (mount.RawID >= 0 && specialRawIDForName(mount.Name) != mount.RawID) {
			return 0, false
		}
		one := s.DesignCostWithLoadout(design.Class, 0, 0, 0, special, nil,
			gamedata.ARC_FWD, 255) - ShipCost(design.Class)
		cost += one
	}
	return cost, true
}

// BlueprintDesignSpaceUsed 同成本入口，但回傳完整多槽空間。
func (s *GameSession) BlueprintDesignSpaceUsed(design ShipBlueprint) (used int, ok bool) {
	used = s.DesignSpaceUsedWithLoadout(design.Class, 0, design.Armor, design.Shield, 0,
		nil, gamedata.ARC_FWD, NormalizeWeaponAmmo(pick(WeaponOptions, 0).Name, 0))
	mounts := design.WeaponMounts
	if len(mounts) == 0 {
		mounts = []ShipWeaponMount{{Name: pick(WeaponOptions, design.Weapon).Name, MaxCount: 1,
			WorkingCount: 1, Arc: design.Arc, Ammo: design.Ammo}}
	}
	for i, mount := range mounts {
		count := blueprintMountCount(mount)
		if count == 0 || mount.Name == "" {
			continue
		}
		weapon, known := weaponOptionIndex(mount.Name)
		if !known || (mount.RawMods != 0 && len(mount.Mods) == 0) {
			return 0, false
		}
		mods := mount.Mods
		if i == 0 && len(mods) == 0 {
			mods = design.Mods
		}
		one := s.DesignSpaceUsedWithLoadout(design.Class, weapon, 0, 0, 0, mods,
			mount.Arc, mount.Ammo)
		used += one * count
	}
	specials := design.Specials
	if len(specials) == 0 {
		specials = []ShipSpecialMount{specialMountFromOption(design.Special)}
	}
	if !specialNamesUnique(specials) {
		return 0, false
	}
	for _, mount := range specials {
		if mount.Name == "" || mount.Name == "無" {
			continue
		}
		special, known := specialOptionIndex(mount.Name)
		if !known || (mount.RawID >= 0 && specialRawIDForName(mount.Name) != mount.RawID) {
			return 0, false
		}
		one := s.DesignSpaceUsedWithLoadout(design.Class, 0, 0, 0, special, nil,
			gamedata.ARC_FWD, 255)
		used += one
	}
	return used, true
}

func (s *GameSession) BlueprintDesignFits(design ShipBlueprint) bool {
	used, ok := s.BlueprintDesignSpaceUsed(design)
	return ok && used <= s.HullSpaceFor(design.Class)
}

// ResetShipDesign 依目前科技重建指定艦體的合法模板。
func (s *GameSession) ResetShipDesign(hull int) bool {
	if hull < 0 || hull >= PlayerShipDesignCount {
		return false
	}
	loadout, ok := s.AutoDesignShip(playerShipDesignClasses[hull], AutoDesignMixed)
	if !ok {
		return false
	}
	s.EnsureShipDesigns()
	s.ShipDesigns[hull] = blueprintFromLoadout(playerShipDesignClasses[hull], loadout)
	return true
}

// UpdatePlayerShipDesignsAfterTech 對映真人戰術分支
// Update_Automatic_Design_Equipment_ @ 0x572AB：六筆設計只升級自動裝備，不重建玩家的
// 武器／特殊裝置。remake 的曲速引擎與成本由當前科技即時計算，因此持久欄位只需更新裝甲。
func (s *GameSession) UpdatePlayerShipDesignsAfterTech() {
	s.EnsureShipDesigns()
	bestArmor := s.bestUnlockedComponent(ArmorOptions)
	for i := range s.ShipDesigns {
		s.ShipDesigns[i].Armor = bestArmor
	}
}

// BuildShipDesign 只在玩家明確按 BUILD 時依指定 blueprint 建造。
func (s *GameSession) BuildShipDesign(hull int) bool {
	design, ok := s.ShipDesign(hull)
	if !ok {
		return false
	}
	payload, err := json.Marshal(design)
	if err != nil {
		return false
	}
	s.recordPlayerCommand(PlayerCommand{Name: CmdBuildShipDesign, Args: []int{hull}, Text: string(payload)})
	return s.buildShipBlueprint(design)
}

func (s *GameSession) buildShipBlueprint(design ShipBlueprint) bool {
	total, known := s.BlueprintDesignCost(design)
	if !known || !s.BlueprintDesignFits(design) || s.Player.BC < total {
		return false
	}
	legacyCost := s.DesignCostWithLoadout(design.Class, design.Weapon, design.Armor, design.Shield,
		design.Special, design.Mods, design.Arc, design.Ammo)
	// 內層相容建造不得再送出單槽 CmdBuildShip；外層命令已攜帶完整 blueprint。
	s.commandReplayDepth++
	built := s.BuildShipWithLoadout(design.Class, design.Weapon, design.Armor, design.Shield,
		design.Special, design.Mods, design.Arc, design.Ammo)
	s.commandReplayDepth--
	if !built {
		return false
	}
	s.Player.BC -= total - legacyCost
	// BuildShipWithLoadout 仍是單槽相容建造核心；成功後把 blueprint 的完整 typed 資料
	// 深複製到剛交付的艦，讓快速／戰術 consumer 可逐步遷移而不再於建造邊界丟槽。
	fleet := s.Fleet()
	if len(fleet.Ships) == 0 {
		return false
	}
	ship := &fleet.Ships[len(fleet.Ships)-1]
	if len(design.WeaponMounts) > 0 {
		ship.WeaponMounts = cloneWeaponMounts(design.WeaponMounts)
	}
	ship.SpecialIDs = append([]int(nil), design.SpecialIDs...)
	ship.Specials = cloneSpecialMounts(design.Specials)
	return true
}
