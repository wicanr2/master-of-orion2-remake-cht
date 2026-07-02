package save

import (
	"encoding/binary"
	"os"
	"testing"
)

// expectedSeqEnd 是順序區(colonies…ships)解析後應收斂到的 offset。
// 此值經真實 SAVE10.GAM 驗證:當時各計數皆合理(colonyCount 5/planetCount 82/
// starCount 36/playerCount 5/shipCount 21)、有名星系數==starCount==36、首星 "Orion",
// 證明各實體結構欄位精確對位;203596 即該自洽 schema 的順序區結尾(< galaxy@0x31be4,無碰撞)。
// 作為回歸護欄:任一結構大小被改動 → 此值偏移,兩個測試都會抓到。
const expectedSeqEnd = 203596

// buildSyntheticSave 組出一份完整大小的全零存檔(所有計數 = 0),用來在無版權測資下
// 驗證整條解析鏈(尤其 SeqEnd 收斂 → 各結構大小正確)。
func buildSyntheticSave() []byte {
	d := make([]byte, 208000)
	le := binary.LittleEndian

	// GameConfig @0
	le.PutUint32(d[0:], saveVersion)
	copy(d[4:], []byte("TEST SAVE"))
	le.PutUint32(d[4+saveGameNameSize:], 30) // stardate
	d[4+saveGameNameSize+4] = 1              // multiplayer flag

	// Galaxy @0x31be4
	g := galaxyOffset
	d[g] = 2
	le.PutUint16(d[g+5:], 60)
	le.PutUint16(d[g+7:], 45)
	d[g+31] = 3 // nebulaCount

	// 所有計數(colonyCount@0x25b、planetCount、…)留 0,實體全零。
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
	if gs.ColonyCount != 0 {
		t.Errorf("colonyCount = %d(全零檔應為 0)", gs.ColonyCount)
	}
	// 關鍵:順序區收斂到預期 offset → 各實體結構大小正確。
	if gs.SeqEnd != expectedSeqEnd {
		t.Errorf("SeqEnd = %d,預期 %d(某結構大小對不齊)", gs.SeqEnd, expectedSeqEnd)
	}
	if len(gs.Colonies) != maxColonies || len(gs.Ships) != maxShips || len(gs.Players) != maxPlayers {
		t.Errorf("實體陣列長度錯:colonies=%d ships=%d players=%d", len(gs.Colonies), len(gs.Ships), len(gs.Players))
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
	t.Logf("存檔=%q stardate=%d galaxy=%dx%d colonyCount=%d planetCount=%d starCount=%d playerCount=%d shipCount=%d SeqEnd=%d 檔案=%d",
		gs.Config.SaveGameName, gs.Config.Stardate, gs.Galaxy.Width, gs.Galaxy.Height,
		gs.ColonyCount, gs.PlanetCount, gs.StarCount, gs.PlayerCount, gs.ShipCount, gs.SeqEnd, len(data))
	// 計數合理(readCount 已擋上限)+ 順序區收斂 → schema 對齊真實存檔。
	if gs.SeqEnd != expectedSeqEnd {
		t.Errorf("SeqEnd = %d,預期 %d", gs.SeqEnd, expectedSeqEnd)
	}
	if gs.PlayerCount == 0 || gs.PlayerCount > maxPlayers {
		t.Errorf("playerCount = %d 不合理", gs.PlayerCount)
	}
	// 抽查:找出人類玩家(personality==100)與非空星名,證明欄位真的對位。
	humans := 0
	for i := range gs.Players {
		if gs.Players[i].Personality == 100 {
			humans++
		}
	}
	named := 0
	for i := range gs.Stars {
		if gs.Stars[i].Name != "" {
			named++
		}
	}
	t.Logf("人類玩家數=%d 有名星系數=%d 首星=%q", humans, named, gs.Stars[0].Name)
	for i := 0; i < gs.PlayerCount; i++ {
		t.Logf("  玩家%d 名稱=%q 種族=%q personality=%d picture=%d", i,
			gs.Players[i].Name, gs.Players[i].Race, gs.Players[i].Personality, gs.Players[i].Picture)
	}
}
