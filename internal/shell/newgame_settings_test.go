package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// newgame_settings_test.go:NEW GAME 畫面那五個設定。
//
// 這裡盯的是「設定有沒有真的生效」,不是「欄位存不存在」——星系年齡先前就是有完整的
// gamedata 實作、卻被一個常數寫死成 Average,欄位加了也不會有人發現沒接上。

// TestGalaxyAgeChangesGalaxy:同一個種子、不同年齡,星系必須不一樣。
// 年齡的效果在光譜加權(gamedata.StarClassWeights)與氣候骰表,兩者都會改變星圖內容。
func TestGalaxyAgeChangesGalaxy(t *testing.T) {
	const seed = 12345
	spectra := func(age gamedata.GalaxyAge) []int {
		s := NewDemoSession()
		s.GalaxyAge, s.GalaxyAgeSet = age, true
		s.SetupNewGame(24, seed, DefaultOpponents)
		out := make([]int, len(s.Stars))
		for i, st := range s.Stars {
			out[i] = st.Spectral
		}
		return out
	}
	young, mature := spectra(gamedata.GalaxyYouthful), spectra(gamedata.GalaxyMature)
	if len(young) != len(mature) {
		t.Fatalf("星數不該隨年齡改變:%d vs %d", len(young), len(mature))
	}
	same := true
	for i := range young {
		if young[i] != mature[i] {
			same = false
			break
		}
	}
	if same {
		t.Error("年輕與成熟星系的光譜分布完全相同——星系年齡沒有接上生成端")
	}
}

// TestGalaxyAgeDefaultsToAverage:沒設過年齡(舊存檔、demo session)要退回普通,
// 不能吃到 gamedata.GalaxyAge 的 Go 零值(GalaxyYouthful)。
func TestGalaxyAgeDefaultsToAverage(t *testing.T) {
	s := NewDemoSession()
	if s.GalaxyAgeSet {
		t.Fatal("demo session 不該自稱設過星系年齡")
	}
	if got := s.galaxyAge(); got != gamedata.GalaxyAverage {
		t.Errorf("未設定時應退回普通(%v),實得 %v", gamedata.GalaxyAverage, got)
	}
}

// TestEmpireCountControlsOpponents:帝國總數要真的決定 AI 對手數。
// 這一欄在原版就是 PLAYERS(2–8),remake 先前把那個框拿去當種族選擇,對手數寫死 3。
func TestEmpireCountControlsOpponents(t *testing.T) {
	for _, empires := range []int{MinEmpires, 5, MaxEmpires} {
		s := NewDemoSession()
		s.SetupNewGame(48, 7, empires-1)
		if got := len(s.AIPlayers) + 1; got != empires {
			t.Errorf("帝國總數 %d:實得 %d(AI 對手 %d)", empires, got, len(s.AIPlayers))
		}
	}
}

// TestHotseatSeatsLimitedByEmpireCount:熱座席位是從 AI 帝國接管來的,
// 所以帝國總數同時就是熱座席位上限。
func TestHotseatSeatsLimitedByEmpireCount(t *testing.T) {
	s := NewDemoSession()
	s.SetupNewGame(48, 7, MaxEmpires-1) // 8 個帝國
	if got := s.SetupHotseat(MaxEmpires); got != MaxEmpires {
		t.Errorf("8 帝國局應開得出 %d 席熱座,實得 %d", MaxEmpires, got)
	}

	s2 := NewDemoSession()
	s2.SetupNewGame(24, 7, MinEmpires-1) // 2 個帝國
	if got := s2.SetupHotseat(4); got != MinEmpires {
		t.Errorf("2 帝國局最多只能 %d 席,實得 %d", MinEmpires, got)
	}
}

// TestDifficultyCountMatchesOriginal:難度五級是反組譯確認的(選擇器選項數 5 +
// NEWGAME.LBX 五張手勢圖)。少一級代表又退回四級的舊狀態。
func TestDifficultyCountMatchesOriginal(t *testing.T) {
	if len(Difficulties) != 5 {
		t.Errorf("難度應為 5 級(原版 Tutor..Impossible),實得 %d", len(Difficulties))
	}
	// 倍率要遞增,否則「難度」這個詞就沒有意義。
	for i := 1; i < len(Difficulties); i++ {
		if Difficulties[i].Mult <= Difficulties[i-1].Mult {
			t.Errorf("難度倍率應遞增:%s(%.2f)未大於 %s(%.2f)",
				Difficulties[i].Name, Difficulties[i].Mult,
				Difficulties[i-1].Name, Difficulties[i-1].Mult)
		}
	}
}

// TestNewGameSettingsSurviveSaveLoad:設定要進存檔,否則讀回來查不到這局是什麼設定。
func TestNewGameSettingsSurviveSaveLoad(t *testing.T) {
	s := NewDemoSession()
	s.GalaxyAge, s.GalaxyAgeSet = gamedata.GalaxyMature, true
	s.TechLevel, s.TechLevelSet = 2, true
	path := t.TempDir() + "/settings.json"
	if err := s.Save(path); err != nil {
		t.Fatalf("存檔失敗:%v", err)
	}
	got, err := LoadSession(path)
	if err != nil {
		t.Fatalf("讀檔失敗:%v", err)
	}
	if !got.GalaxyAgeSet || got.GalaxyAge != gamedata.GalaxyMature {
		t.Errorf("星系年齡沒還原:set=%v age=%v", got.GalaxyAgeSet, got.GalaxyAge)
	}
	if !got.TechLevelSet || got.TechLevel != 2 {
		t.Errorf("科技等級沒還原:set=%v level=%d", got.TechLevelSet, got.TechLevel)
	}
}

// --- 起始科技等級的 gameplay 效果(2026-08-07 接上的第一項) ---

// TestPrewarpFleetCannotLeaveSystem:曲速前開局沒有 FTL,艦隊出不了本星系。
// 手冊直引:「Exploring outside that system is impossible until faster than light (FTL)
// technologies are discovered.」
func TestPrewarpFleetCannotLeaveSystem(t *testing.T) {
	s := NewDemoSession()
	s.SetupNewGame(24, 5, DefaultOpponents)
	s.TechLevel, s.TechLevelSet = TechLevelPrewarp, true

	if s.FleetHasFTL() {
		t.Fatal("曲速前且未研究核分裂時不該有 FTL")
	}
	dest := (s.FleetAtStar + 1) % len(s.Stars)
	if s.SendFleet(dest) {
		t.Error("曲速前開局的艦隊不該派得出去")
	}
	if s.FleetETA != 0 || s.FleetDestStar >= 0 {
		t.Errorf("派遣被擋下時不該留下航行狀態:ETA=%d dest=%d", s.FleetETA, s.FleetDestStar)
	}

	// 研究完 FTL 主題就解禁。
	if s.Player.CompletedTopics == nil {
		s.Player.CompletedTopics = map[gamedata.ResearchTopic]bool{}
	}
	s.Player.CompletedTopics[FTLTopic] = true
	if !s.FleetHasFTL() {
		t.Error("研究完 FTL 主題後應可離開本星系")
	}
	if !s.SendFleet(dest) {
		t.Error("有 FTL 之後派遣應成功")
	}
}

// TestNonPrewarpFleetUnrestricted:一般 / 先進開局本來就配了 FTL 引擎,不該被擋。
func TestNonPrewarpFleetUnrestricted(t *testing.T) {
	for _, lvl := range []int{1, 2} {
		s := NewDemoSession()
		s.SetupNewGame(24, 5, DefaultOpponents)
		s.TechLevel, s.TechLevelSet = lvl, true
		if !s.FleetHasFTL() {
			t.Errorf("科技等級 %d 開局就該有 FTL", lvl)
		}
		if !s.SendFleet((s.FleetAtStar + 1) % len(s.Stars)) {
			t.Errorf("科技等級 %d 的艦隊派遣不該被擋", lvl)
		}
	}
}

// TestTechLevelZeroValueDoesNotFreezeFleet:**零值陷阱的回歸測試**。
// TechLevel 的 Go 零值是 0 = 曲速前;沒有 TechLevelSet 標記的話,舊存檔與任何沒設過
// 這欄的建構路徑會被靜默判成曲速前、艦隊整個凍住。
func TestTechLevelZeroValueDoesNotFreezeFleet(t *testing.T) {
	s := NewDemoSession()
	s.SetupNewGame(24, 5, DefaultOpponents)
	if s.TechLevelSet {
		t.Fatal("demo session 不該自稱設過科技等級")
	}
	if !s.FleetHasFTL() {
		t.Error("沒設過科技等級時應退回「一般」,艦隊不該被凍住")
	}
}
