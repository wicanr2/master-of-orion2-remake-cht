package shell

import (
	"os"
	"path/filepath"
	"testing"
)

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
