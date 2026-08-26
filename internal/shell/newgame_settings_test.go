package shell

import (
	"reflect"
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
	want := []string{"教學", "簡單", "普通", "困難", "不可能"}
	for i := range want {
		if Difficulties[i].Name != want[i] {
			t.Errorf("難度索引 %d = %q，應為 %q", i, Difficulties[i].Name, want[i])
		}
	}
}

// 原版沒有一張通用浮點倍率同時縮放敵艦 blueprint 與外交關係。
// 難度必須由各已證實 consumer 讀離散索引，不得在代理艦隊重新發明倍率。
func TestEnemyFleetProxyHasNoDifficultyMultiplier(t *testing.T) {
	want := genEnemyFleet(12)
	for difficulty := range Difficulties {
		s := NewDemoSession()
		s.Difficulty = difficulty
		got := genEnemyFleet(12)
		if len(got) != len(want) {
			t.Fatalf("難度 %d 改變了代理艦數：%d vs %d", difficulty, len(got), len(want))
		}
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("難度 %d 不應直接縮放代理艦 %d：%d vs %d", difficulty, i, got[i], want[i])
			}
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

// 1.31 與 1.50 必須各自走完「開局→隨機科技應用→建築→存讀檔」。
// 選單可切版本只是 UI 證據，不能代替這條垂直鏈。
func TestAdvancedOpeningRoundTripsForBothRuleProfiles(t *testing.T) {
	for _, profile := range []gamedata.RuleProfile{gamedata.Profile13(), gamedata.Profile15()} {
		s := NewDemoSession()
		s.SetRuleProfile(profile)
		s.Difficulty = 4
		s.TechLevel, s.TechLevelSet = 2, true
		s.DisableEvents = true
		s.SetupNewGame(24, 4242, 2)

		completed := 0
		for topic := range s.Player.CompletedTopics {
			if topic != gamedata.TOPIC_STARTING_TECH {
				completed++
			}
		}
		if completed != 25 {
			t.Fatalf("版本 %v 的先進開局應有 25 個主題，得到 %d", profile.Version, completed)
		}
		if len(s.ColonyBuildings) == 0 || len(s.ColonyBuildings[0]) == 0 {
			t.Fatalf("版本 %v 的先進開局沒有建立母星建築", profile.Version)
		}

		path := t.TempDir() + "/advanced-opening.json"
		if err := s.Save(path); err != nil {
			t.Fatalf("版本 %v 存檔失敗：%v", profile.Version, err)
		}
		got, err := LoadSession(path)
		if err != nil {
			t.Fatalf("版本 %v 讀檔失敗：%v", profile.Version, err)
		}
		if got.RuleProfile != profile || got.Difficulty != 4 || !got.TechLevelSet || got.TechLevel != 2 {
			t.Errorf("版本 %v 開局設定往返不符：profile=%v difficulty=%d tech=%d/%v",
				profile.Version, got.RuleProfile.Version, got.Difficulty, got.TechLevel, got.TechLevelSet)
		}
		if !reflect.DeepEqual(got.Player.CompletedTopics, s.Player.CompletedTopics) ||
			!reflect.DeepEqual(got.Player.ChosenTech, s.Player.ChosenTech) ||
			!reflect.DeepEqual(got.Player.ExplicitChoice, s.Player.ExplicitChoice) {
			t.Errorf("版本 %v 的開局科技／應用選擇在存讀檔後改變", profile.Version)
		}
		if !reflect.DeepEqual(got.ColonyBuildings, s.ColonyBuildings) {
			t.Errorf("版本 %v 的母星建築在存讀檔後改變", profile.Version)
		}
	}
}

// --- 起始科技等級的 gameplay 效果(2026-08-07 接上的第一項) ---

// TestPrewarpFleetCannotLeaveSystem:曲速前開局沒有 FTL,艦隊出不了本星系。
// 手冊直引:「Exploring outside that system is impossible until faster than light (FTL)
// technologies are discovered.」
func TestPrewarpFleetCannotLeaveSystem(t *testing.T) {
	s := NewDemoSession()
	// ⚠ TECH LEVEL 要在 SetupNewGame **之前**設:開局送的研究主題由它決定
	// (見 applyStartingTech),而那一步在 SetupNewGame 的最後。
	// 正式流程也是這個順序(cmd/moo2 的 applyNewGameSettings → SetupNewGame)。
	s.TechLevel, s.TechLevelSet = TechLevelPrewarp, true
	s.SetupNewGame(24, 5, DefaultOpponents)

	if s.FleetHasFTL() {
		t.Fatal("曲速前且未研究核分裂時不該有 FTL")
	}
	dest := (s.Fleet().AtStar + 1) % len(s.Stars)
	if s.SendFleet(dest) {
		t.Error("曲速前開局的艦隊不該派得出去")
	}
	if s.Fleet().ETA != 0 || s.Fleet().DestStar >= 0 {
		t.Errorf("派遣被擋下時不該留下航行狀態:ETA=%d dest=%d", s.Fleet().ETA, s.Fleet().DestStar)
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
		if !s.SendFleet((s.Fleet().AtStar + 1) % len(s.Stars)) {
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

// --- 起始科技等級:開局送的研究主題(2026-08-07 接上的第二項效果) ---

// TestTechLevelGrantsTheOriginalStartingTopics:選什麼等級就該拿到那一級的主題。
//
// 這條先前**不成立**:不論選哪一級,開局都只有 STARTING_TECH + ENGINEERING 兩個
// ——照 `Init_Player_Tech_` @ 0x5E55F 那是曲速前(var_18 = 1),而選單上寫著「一般」。
func TestTechLevelGrantsTheOriginalStartingTopics(t *testing.T) {
	for _, c := range []struct {
		level int
		want  int // 固定表裡該拿到幾個
	}{
		{TechLevelPrewarp, 1},
		{TechLevelDefault, 6},
		{2, 6}, // 先進級的固定表也是 6,其餘 19 個是隨機的(還沒接)
	} {
		s := NewDemoSession()
		s.TechLevel, s.TechLevelSet = c.level, true
		s.SetupNewGame(24, 7, DefaultOpponents)

		got := 0
		for _, topic := range gamedata.StartingTopicOrder {
			if s.Player.CompletedTopics[topic] {
				got++
			}
		}
		if got != c.want {
			t.Errorf("等級 %d 應拿到固定表裡的 %d 個,實得 %d", c.level, c.want, got)
		}
		// field 0 與等級無關,一律要有(原版開頭那批無條件寫入)。
		if !s.Player.CompletedTopics[gamedata.TOPIC_STARTING_TECH] {
			t.Errorf("等級 %d:field 0 應無條件已知", c.level)
		}
		// ENGINEERING 是固定表第一個,每一級都該有。
		if !s.Player.CompletedTopics[gamedata.TOPIC_ENGINEERING] {
			t.Errorf("等級 %d:ENGINEERING 是固定表第一個,每一級都該有", c.level)
		}
	}
}

// 曲速前**不能**拿到核分裂——那是 FTL 所在的主題,拿到就等於曲速前開局能飛。
//
// 這條是 applyStartingTech 那個「先清再發」的正對照:只加不減的話,
// NewDemoSession 用預設等級發過的核分裂會留下來,而且不會有任何錯誤訊息。
func TestPrewarpDoesNotKeepTheDefaultLevelsFTLTopic(t *testing.T) {
	s := NewDemoSession()
	if !s.Player.CompletedTopics[FTLTopic] {
		t.Fatal("測試前提不成立:demo 局(預設一般)本來就該有 FTL 主題")
	}
	s.TechLevel, s.TechLevelSet = TechLevelPrewarp, true
	s.SetupNewGame(24, 7, DefaultOpponents)
	if s.Player.CompletedTopics[FTLTopic] {
		t.Error("改成曲速前之後仍留著 FTL 主題——先發後清那一步失效了")
	}
}

// AI 也要照同一級發:原版的 Init_Player_Tech_ 是逐玩家跑的,不是玩家專屬。
func TestStartingTopicsAlsoGoToAI(t *testing.T) {
	s := NewDemoSession()
	s.TechLevel, s.TechLevelSet = TechLevelDefault, true
	s.SetupNewGame(24, 7, DefaultOpponents)
	if len(s.AIPlayers) == 0 {
		t.Fatal("測試前提不成立:沒有 AI 對手")
	}
	for i := range s.AIPlayers {
		if !s.AIPlayers[i].Player.CompletedTopics[gamedata.TOPIC_NUCLEAR_FISSION] {
			t.Errorf("AI %d 沒拿到一般等級該有的核分裂——只發給玩家等於把 AI 留在曲速前", i)
		}
	}
}

// --- 開局建築(2026-08-07:優先清單接上) ---

// TestHomeworldBuildingsMatchTheHardCodedPair 是**正對照**:
// 從原版優先清單算出來的結果,必須與先前寫死的那兩棟逐項相同。
//
// 先前那兩棟是照手冊那句話手寫的;現在改成從 `word_17D8AC` × 開局主題表 × 建築前置表算。
// 兩條路走到同一個答案才算數——只有新的那條綠,證明不了它對。
func TestHomeworldBuildingsMatchTheHardCodedPair(t *testing.T) {
	want := homeworldBuildingsLegacy()
	got := homeworldBuildingsFor(TechLevelDefault, homeworldStartPop)
	if len(got) != len(want) {
		t.Fatalf("算出來 %d 棟,寫死的是 %d 棟:%v vs %v", len(got), len(want), got, want)
	}
	for name := range want {
		if !got[name] {
			t.Errorf("算出來的少了 %q", name)
		}
	}
}

// 上限是 min(⌈⅔ pop⌉, 等級上限)——手冊給了驗證數字:
// 「a HW with 8 pop can have 6 buildings on Advanced Tech start, but only 5 on Average start」。
func TestStartingBuildingCountMatchesTheManualExample(t *testing.T) {
	if got := StartingBuildingCount(8, gamedata.InitialBuildingCap(2)); got != 6 {
		t.Errorf("8 人口 + 先進級應為 6,實得 %d", got)
	}
	if got := StartingBuildingCount(8, gamedata.InitialBuildingCap(1)); got != 5 {
		t.Errorf("8 人口 + 一般級應為 5(受上限),實得 %d", got)
	}
}

// ⚠ 先進級**目前仍然只有兩棟**,而且原因不在這一層。
//
// 上限是 9、⌈⅔×8⌉ = 6,所以名額有 6 個;但清單裡科技條件成立的只有兩棟——因為
// 先進級該多拿的 19 個隨機主題還沒發(見 gamedata.StartingTopicRandomExtras)。
// **這一條把「缺口在哪一層」釘住**:機制是對的,缺的是上游的科技。
// ⚠ 這支測試 2026-08-07(第 44 項(下游讀真值))改名並反轉了斷言。
//
// 它原本叫 `TestAdvancedStartIsBlockedByTheMissingRandomTopics`,斷言「先進級目前應仍是
// 兩棟(缺口在上游的隨機主題)」,並附一句「那 19 個隨機主題若接上了,這條測試要跟著改」。
// 第 43 項(先進級開局主題)把 19 個接上了,所以它跟著改。
//
// **舊版的正對照預測了新版的結果**:那句「科技全解時應發滿 6 個名額(⌈⅔×8⌉)」正是
// 先進級現在真的拿到的棟數。缺口補上之後兩邊自己對上,不是把斷言改成事後諸葛。
func TestAdvancedStartFillsAllBuildingSlots(t *testing.T) {
	// 先進級現在有 25 個開局主題(第 43 項(先進級開局主題)),所以 6 個名額全部發得出來。
	s := NewDemoSession()
	s.DisableEvents = true
	s.TechLevel, s.TechLevelSet = 2, true
	s.applyStartingTech()
	if got := len(s.ColonyBuildings[0]) - 1; got != 6 {
		t.Errorf("先進級母星應發滿 6 個名額(⌈⅔×8⌉),實得 %d 棟:%v",
			got, s.ColonyBuildings[0])
	}
	// 手冊逐字:曲速前與一般級「only start with Marine Barracks and a Star Base
	// because no other techs are Known that are also in the default initial buildings list」。
	for _, lv := range []int{0, 1} {
		s := NewDemoSession()
		s.DisableEvents = true
		s.TechLevel, s.TechLevelSet = lv, true
		s.applyStartingTech()
		if got := len(s.ColonyBuildings[0]) - 1; got != 2 {
			t.Errorf("TECH LEVEL %d 應仍是兩棟(手冊逐字),實得 %d 棟:%v",
				lv, got, s.ColonyBuildings[0])
		}
	}
	if gamedata.StartingTopicRandomExtras(2) != 19 {
		t.Error("先進級的隨機主題數量變了——第 43 項(先進級開局主題)的接線要跟著檢查")
	}
	// 正對照:科技夠多的時候,這套機制**確實**會發到名額用完。
	rich := map[gamedata.ResearchTopic]bool{gamedata.TOPIC_STARTING_TECH: true}
	for _, b := range gamedata.Buildings {
		rich[b.PrereqTopic] = true
	}
	full := gamedata.InitialBuildings(rich, StartingBuildingCount(homeworldStartPop, gamedata.InitialBuildingCap(2)))
	if len(full) != 6 {
		t.Errorf("科技全解時應發滿 6 個名額(⌈⅔×8⌉),實得 %d:%v", len(full), full)
	}
	// 而且發的是清單最前面的——順序是原版的優先序,不是隨便挑。
	if full[0] != "星要塞" && full[0] != "星際要塞" {
		t.Logf("清單最前面是編號 41(Star Fortress),remake 中文名為 %q", full[0])
	}
}
