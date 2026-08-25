package shell

import (
	"fmt"
)

// ShipSpecialMount 是原版 special bitset 的 typed runtime 對應。RawID=-1 表示 remake
// 可裝載、但原版分類在武器／機庫／獨立欄位的元件；未知正 raw ID 仍保留名稱與數值。
type ShipSpecialMount struct {
	RawID int    `json:"rawId"`
	Name  string `json:"name"`
}

func cloneSpecialMounts(in []ShipSpecialMount) []ShipSpecialMount {
	return append([]ShipSpecialMount(nil), in...)
}

func specialRawIDForName(name string) int {
	if dev, ok := specialDeviceByName[name]; ok {
		return int(dev)
	}
	return -1
}

func specialNameForRawID(id int) (string, bool) {
	for name, dev := range specialDeviceByName {
		if int(dev) == id {
			return name, true
		}
	}
	return fmt.Sprintf("原版特殊#%d", id), false
}

func specialMountFromOption(index int) ShipSpecialMount {
	name := pick(SpecialOptions, index).Name
	return ShipSpecialMount{RawID: specialRawIDForName(name), Name: name}
}

func specialOptionIndex(name string) (int, bool) {
	for i, c := range SpecialOptions {
		if c.Name == name {
			return i, true
		}
	}
	return 0, false
}

// SpecialOptionIndex 供 UI 將 typed 特殊裝置名稱映回受控選項索引。
func SpecialOptionIndex(name string) (int, bool) { return specialOptionIndex(name) }

// shipHasSpecial 是所有玩法 consumer 的單一入口。typed slots 存在時讀完整集合；舊存檔
// 沒有 slots 才回退單槽相容字串。
func shipHasSpecial(sh Ship, name string) bool {
	if len(sh.Specials) > 0 {
		for _, mount := range sh.Specials {
			if mount.Name == name {
				return true
			}
		}
		return false
	}
	return sh.Special == name
}

func fighterBaysForShip(sh Ship) []FighterKind {
	var bays []FighterKind
	for _, entry := range []struct {
		name string
		kind FighterKind
	}{{"戰機庫", FighterInterceptor}, {"重戰機庫", FighterHeavy}, {"轟炸機庫", FighterBomber},
		{assaultShuttleName, FighterAssaultShuttle}} {
		if shipHasSpecial(sh, entry.name) {
			bays = append(bays, entry.kind)
		}
	}
	return bays
}

func specialNamesUnique(mounts []ShipSpecialMount) bool {
	seen := map[string]bool{}
	for _, mount := range mounts {
		if mount.Name == "" || mount.Name == "無" {
			continue
		}
		if seen[mount.Name] {
			return false
		}
		seen[mount.Name] = true
	}
	return true
}

func specialIDsFromMounts(mounts []ShipSpecialMount) []int {
	ids := make([]int, 0, len(mounts))
	for _, mount := range mounts {
		if mount.RawID > 0 {
			ids = append(ids, mount.RawID)
		}
	}
	return ids
}
