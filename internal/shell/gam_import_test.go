package shell

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/save"
)

func TestApplyGAMPopulationProfilesUsesPlayerSlots(t *testing.T) {
	raw := &save.GameState{Players: make([]save.Player, 2)}
	raw.Players[0].Traits[gamedata.TRAIT_FARMING] = 2
	raw.Players[0].Traits[gamedata.TRAIT_LOW_G] = 1
	raw.Players[0].Traits[gamedata.TRAIT_LITHOVORE] = 1
	raw.Players[1].Traits[gamedata.TRAIT_SCIENCE] = 2
	raw.Players[1].Traits[gamedata.TRAIT_HIGH_G] = 1
	raw.Players[1].Traits[gamedata.TRAIT_CYBERNETIC] = 1
	c := engine.ColonyState{
		Population: 2, Farmers: 1, Scientists: 1,
		PopulationGroups: []engine.PopulationGroup{
			{RaceSlot: 0, RaceSlotKnown: true, Farmers: 1},
			{RaceSlot: 1, RaceSlotKnown: true, Scientists: 1},
		},
	}
	applyGAMPopulationProfiles(&c, raw, 0)
	if c.OwnerFoodBonus != 2 || !c.OwnerRaceProfileKnown || c.OwnerRaceSlot != 0 {
		t.Fatalf("owner slot profile 錯誤：%+v", c)
	}
	if c.PopulationGroups[0].Gravity != gamedata.LOW_G || c.PopulationGroups[0].FoodBonus != 2 ||
		!c.PopulationGroups[0].Lithovore || c.PopulationGroups[1].Gravity != gamedata.HEAVY_G ||
		c.PopulationGroups[1].ResearchBonus != 2 || !c.PopulationGroups[1].Cybernetic {
		t.Fatalf("逐 player slot profile 錯誤：%+v", c.PopulationGroups)
	}
}

func TestApplyGAMPopulationProfilesSpecialSlots(t *testing.T) {
	raw := &save.GameState{Players: make([]save.Player, 1)}
	c := engine.ColonyState{
		Population: 2, Farmers: 1, Workers: 1,
		PopulationGroups: []engine.PopulationGroup{
			{RaceSlot: gamedata.AndroidColonistSlot, RaceSlotKnown: true, Workers: 1},
			{RaceSlot: gamedata.NativeColonistSlot, RaceSlotKnown: true, Farmers: 1},
		},
	}
	applyGAMPopulationProfiles(&c, raw, 0)
	android, natives := c.PopulationGroups[0], c.PopulationGroups[1]
	if !android.ProfileKnown || !android.GravityImmune || android.FoodBonus != 6 ||
		android.IndustryBonus != 3 || android.ResearchBonus != 3 {
		t.Fatalf("Android profile 錯誤：%+v", android)
	}
	if !natives.ProfileKnown || !natives.GravityImmune || natives.FoodBonus != 4 ||
		natives.IndustryBonus != 0 || natives.ResearchBonus != 0 {
		t.Fatalf("Natives profile 錯誤：%+v", natives)
	}
}

func TestImportedShipPreservesBankruptcyFields(t *testing.T) {
	raw := save.Ship{Mission: 7}
	raw.Design.Name = "OUTPOST"
	raw.Design.Type = uint8(gamedata.OUTPOST_SHIP)
	raw.Design.Cost = 240
	ship := importedShip(raw, 0, &GAMImportReport{})
	if !ship.RawTypeKnown || ship.RawType != gamedata.OUTPOST_SHIP ||
		!ship.RawMissionKnown || ship.RawMission != 7 || ship.ProductionCost != 240 {
		t.Fatalf("GAM 艦艇破產欄位未保存：%+v", ship)
	}
}

func TestImportedShipPreservesAllWeaponMountsAndSpecialIDs(t *testing.T) {
	design := save.ShipDesign{Name: "MULTI"}
	design.Weapons[0] = save.ShipWeapon{Type: 3, MaxCount: 12, WorkingCount: 10, Arc: 4, Mods: 5, Ammo: 255}
	design.Weapons[1] = save.ShipWeapon{Type: 16, MaxCount: 3, WorkingCount: 2, Arc: 16, Mods: 0, Ammo: 5}
	design.Specials[0] = 1 << 2
	design.Specials[2] = 1 << 1 // raw ID 17
	report := &GAMImportReport{}
	mounts := importedWeaponMounts(design, report)
	if len(mounts) != 2 {
		t.Fatalf("八槽 importer 應保存兩個有效武器槽，得到 %+v", mounts)
	}
	if mounts[0].RawType != 3 || mounts[0].MaxCount != 12 || mounts[0].WorkingCount != 10 || mounts[0].RawMods != 5 {
		t.Fatalf("第一武器槽欄位未完整保存：%+v", mounts[0])
	}
	if mounts[1].RawType != 16 || mounts[1].Arc != 16 || mounts[1].Ammo != 5 {
		t.Fatalf("第二武器槽欄位未完整保存：%+v", mounts[1])
	}
	ids := importedSpecialIDs(design)
	if len(ids) != 2 || ids[0] != 2 || ids[1] != 17 {
		t.Fatalf("special bitset 應完整保存 raw 2/17，得到 %v", ids)
	}
}

func TestWeaponMountsRoundTripSnapshot(t *testing.T) {
	s := NewDemoSession()
	s.Fleet().Ships[0].WeaponMounts = []ShipWeaponMount{{
		RawType: 16, Name: "脈衝飛彈", MaxCount: 3, WorkingCount: 2, Arc: 16, Ammo: 5, Attack: 20,
	}}
	s.Fleet().Ships[0].SpecialIDs = []int{2, 17}
	b, err := s.MarshalSnapshot()
	if err != nil {
		t.Fatal(err)
	}
	got, err := RestoreSnapshot(b)
	if err != nil {
		t.Fatal(err)
	}
	ship := got.Fleet().Ships[0]
	if len(ship.WeaponMounts) != 1 || ship.WeaponMounts[0].RawType != 16 || ship.WeaponMounts[0].WorkingCount != 2 {
		t.Fatalf("weapon mounts 快照往返失真：%+v", ship.WeaponMounts)
	}
	if len(ship.SpecialIDs) != 2 || ship.SpecialIDs[1] != 17 {
		t.Fatalf("special IDs 快照往返失真：%v", ship.SpecialIDs)
	}
}

// TestImportGAMFixture 是抽樣真檔測試；未提供私有原版檔時保持 skip，避免
// 公開 CI 需要攜帶受版權保護的 `.GAM`。
func TestImportGAMFixture(t *testing.T) {
	path := os.Getenv("MOO2_SAVE_TEST")
	if path == "" {
		t.Skip("未設 MOO2_SAVE_TEST")
	}
	session, report, err := LoadGAMSession(path)
	if err != nil {
		t.Fatalf("匯入真實 GAM 失敗: %v", err)
	}
	if session == nil || report.SourceVersion != 0xE0 {
		t.Fatalf("GAM 匯入結果無效：session=%v report=%+v", session != nil, report)
	}
	if report.StarCount == 0 || len(session.Stars) != report.StarCount {
		t.Fatalf("星系未完整轉入：report=%d session=%d", report.StarCount, len(session.Stars))
	}
	if len(session.Fleets) == 0 {
		t.Fatal("匯入後至少應有一支空／非空艦隊")
	}
	loaded, err := LoadSession(path)
	if err != nil || loaded == nil || len(loaded.Stars) != len(session.Stars) {
		t.Fatalf("LoadSession 的 .GAM magic 分流失敗：err=%v loaded=%v", err, loaded != nil)
	}
	if len(session.PlayerColonies) != len(session.PlayerColonyPlanets) ||
		len(session.PlayerColonies) != len(session.ColonyBuildings) {
		t.Fatalf("殖民地平行陣列未對齊：colonies=%d planets=%d buildings=%d",
			len(session.PlayerColonies), len(session.PlayerColonyPlanets), len(session.ColonyBuildings))
	}
	t.Logf("GAM=%q stardate=%d turn=%d stars=%d planets=%d colonies=%d outposts=%d ai=%d ships=%d leaders=%d skippedBuildings=%d notes=%v",
		report.SaveGameName, report.Stardate, report.Turn, report.StarCount, report.PlanetCount,
		report.ImportedColonies, report.ImportedOutposts, report.ImportedAI, report.ImportedShips,
		report.ImportedLeaders, report.SkippedBuildings, report.Notes)
}

// TestReadSaveSlotRecognizesGAMFixture 驗證原版 SAVE10.GAM 可以從載入畫面所用的
// 存檔槽摘要入口被找到，而且不會把原始路徑當成之後的 remake 寫入路徑。
func TestReadSaveSlotRecognizesGAMFixture(t *testing.T) {
	source := os.Getenv("MOO2_SAVE_TEST")
	if source == "" {
		t.Skip("未設 MOO2_SAVE_TEST")
	}
	data, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	native := filepath.Join(dir, "SAVE10.GAM")
	if err := os.WriteFile(native, data, 0o600); err != nil {
		t.Fatal(err)
	}
	got := ReadSaveSlot(dir, AutoSaveSlot)
	if !got.Exists || !got.NativeGAM || got.Path != native {
		t.Fatalf("載入槽沒有辨識原版 GAM：%+v", got)
	}
	if got.Turn < 1 || got.Stardate == "" {
		t.Fatalf("原版 GAM 摘要沒有星曆／回合：%+v", got)
	}
	if jsonPath := SaveSlotPath(dir, AutoSaveSlot); jsonPath == got.Path {
		t.Fatalf("GAM 匯入後不應把 JSON 寫入原版路徑：%s", jsonPath)
	}
}
