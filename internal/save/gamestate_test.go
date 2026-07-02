package save

import (
	"encoding/binary"
	"os"
	"testing"
)

// buildSyntheticSave 組出一份最小可解析的存檔:config@0、galaxy@0x31be4、colonyCount@0x25b。
func buildSyntheticSave() []byte {
	size := galaxyOffset + 64
	d := make([]byte, size)
	le := binary.LittleEndian

	// GameConfig @0
	le.PutUint32(d[0:], saveVersion)
	copy(d[4:], []byte("TEST SAVE"))         // saveGameName(37,NUL 結尾)
	le.PutUint32(d[4+saveGameNameSize:], 30) // stardate
	// 14 個旗標接在 stardate 之後;設第一個(multiplayer)= 1 驗證解析。
	d[4+saveGameNameSize+4] = 1

	// colonyCount @0x25b
	le.PutUint16(d[colonyCountOffset:], 7)

	// Galaxy @0x31be4:sizeFactor=2, skip4, width=60, height=45, skip2, nebulas..., nebulaCount=3
	g := galaxyOffset
	d[g] = 2
	le.PutUint16(d[g+5:], 60)
	le.PutUint16(d[g+7:], 45)
	// nebulaCount 位於 galaxy 起點 +31
	d[g+31] = 3
	return d
}

func TestLoadSynthetic(t *testing.T) {
	gs, err := Load(buildSyntheticSave())
	if err != nil {
		t.Fatalf("Load 失敗: %v", err)
	}
	if gs.Config.Version != saveVersion {
		t.Errorf("version = %#x", gs.Config.Version)
	}
	if gs.Config.SaveGameName != "TEST SAVE" {
		t.Errorf("saveGameName = %q", gs.Config.SaveGameName)
	}
	if gs.Config.Stardate != 30 {
		t.Errorf("stardate = %d", gs.Config.Stardate)
	}
	if !gs.Config.Multiplayer {
		t.Errorf("multiplayer 應為 true")
	}
	if gs.Galaxy.SizeFactor != 2 || gs.Galaxy.Width != 60 || gs.Galaxy.Height != 45 {
		t.Errorf("galaxy = %+v", gs.Galaxy)
	}
	if gs.Galaxy.NebulaCount != 3 {
		t.Errorf("nebulaCount = %d", gs.Galaxy.NebulaCount)
	}
	if gs.ColonyCount != 7 {
		t.Errorf("colonyCount = %d", gs.ColonyCount)
	}
}

func TestLoadRejectsBadVersion(t *testing.T) {
	d := buildSyntheticSave()
	d[0] = 0x00 // 破壞版本
	if _, err := Load(d); err == nil {
		t.Fatal("版本錯誤應回傳 error")
	}
}

// TestLoadRealSave 在 MOO2_SAVE_TEST 指向真實 save?.gam 時驗證解析。
func TestLoadRealSave(t *testing.T) {
	path := os.Getenv("MOO2_SAVE_TEST")
	if path == "" {
		t.Skip("未設 MOO2_SAVE_TEST")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	gs, err := Load(data)
	if err != nil {
		t.Fatalf("解析真實存檔失敗: %v", err)
	}
	t.Logf("存檔=%q stardate=%d galaxy=%dx%d sizeFactor=%d nebulaCount=%d colonyCount=%d 檔案大小=%d",
		gs.Config.SaveGameName, gs.Config.Stardate, gs.Galaxy.Width, gs.Galaxy.Height,
		gs.Galaxy.SizeFactor, gs.Galaxy.NebulaCount, gs.ColonyCount, len(data))
	if gs.Galaxy.NebulaCount > maxNebulas {
		t.Errorf("nebulaCount %d 超過上限", gs.Galaxy.NebulaCount)
	}
}
