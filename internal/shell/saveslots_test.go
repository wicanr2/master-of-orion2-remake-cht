package shell

import (
	"os"
	"path/filepath"
	"testing"
)

// saveslots_test.go:存檔槽。原版規格(10 格、末格自動存檔)逐條驗;
// 「帝國名進不進存檔」是這輪抓到的真 bug,單獨一條護欄。

func TestAutoSaveSlotKeepsLegacyFilename(t *testing.T) {
	dir := t.TempDir()
	// 自動存檔槽必須沿用舊檔名 save.json,否則升級後玩家的既有存檔會憑空消失。
	if got, want := SaveSlotPath(dir, AutoSaveSlot), filepath.Join(dir, "save.json"); got != want {
		t.Errorf("自動存檔槽路徑應為 %s,實得 %s", want, got)
	}
	if got, want := SaveSlotPath(dir, 0), filepath.Join(dir, "save1.json"); got != want {
		t.Errorf("第一格路徑應為 %s(槽號從 1 起算,同原版 SAVE%%u.GAM),實得 %s", want, got)
	}
	if AutoSaveSlot != SaveSlotCount-1 {
		t.Errorf("原版 drawSlot:最後一格才是自動存檔,實得 AutoSaveSlot=%d / 共 %d 格",
			AutoSaveSlot, SaveSlotCount)
	}
}

func TestEmptyDirHasNoSaves(t *testing.T) {
	dir := t.TempDir()
	if AnySaveExists(dir) {
		t.Error("空目錄不該回報有存檔——主選單靠這個決定 Continue/Load 要不要停用")
	}
	if s := LatestSaveSlot(dir); s != -1 {
		t.Errorf("空目錄的最近存檔槽應為 -1,實得 %d", s)
	}
	for _, sl := range ReadSaveSlots(dir) {
		if sl.Exists {
			t.Errorf("空目錄卻回報第 %d 格有存檔", sl.Slot)
		}
	}
}

func TestSaveSlotRoundTripKeepsEmpireName(t *testing.T) {
	dir := t.TempDir()
	s := NewDemoSession()
	s.Turn = 42
	s.PlayerName = "銀河聯邦"
	s.FlagColor = 3
	if err := s.Save(SaveSlotPath(dir, 2)); err != nil {
		t.Fatalf("寫存檔: %v", err)
	}
	info := ReadSaveSlot(dir, 2)
	if !info.Exists {
		t.Fatal("剛寫的存檔讀不到")
	}
	// 這兩個欄位先前根本沒進存檔,讀檔後帝國名會變回預設——存檔槽列表要顯示它才抓到。
	if info.Empire != "銀河聯邦" {
		t.Errorf("帝國名應為「銀河聯邦」,實得 %q", info.Empire)
	}
	if info.Turn != 42 {
		t.Errorf("回合應為 42,實得 %d", info.Turn)
	}
	if info.Stardate != "3541" { // 3500 + 42 − 1
		t.Errorf("星曆應為 3541(3500 起算,見 StartStardate),實得 %s", info.Stardate)
	}
	gs, err := LoadSession(SaveSlotPath(dir, 2))
	if err != nil {
		t.Fatalf("讀檔: %v", err)
	}
	if gs.PlayerName != "銀河聯邦" || gs.FlagColor != 3 {
		t.Errorf("讀檔後帝國名/旗色應保留,實得 %q / %d", gs.PlayerName, gs.FlagColor)
	}
	if !AnySaveExists(dir) {
		t.Error("已有存檔卻回報沒有")
	}
	if got := LatestSaveSlot(dir); got != 2 {
		t.Errorf("最近存檔槽應為 2,實得 %d", got)
	}
}

func TestCorruptSaveReportedAsEmpty(t *testing.T) {
	dir := t.TempDir()
	// 列表寧可顯示空格,也不要秀出一個點下去會爆的格子。
	if err := os.WriteFile(SaveSlotPath(dir, 0), []byte("{ 這不是 JSON"), 0o644); err != nil {
		t.Fatal(err)
	}
	if ReadSaveSlot(dir, 0).Exists {
		t.Error("壞掉的存檔不該顯示為可讀")
	}
}
