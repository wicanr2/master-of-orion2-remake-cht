package shell

import (
	"path/filepath"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// determinism_test.go:決定性的閘門。
//
// 網路對戰的前提是「同樣的種子 + 同樣的指令序列 → 每一台機器算出同樣的狀態」。
// 傳輸層還沒做,但**決定性是規則層自己的性質**,現在就測得了——而且要先測,
// 否則等傳輸層上線才發現規則層本身不決定性,除錯成本會高一個數量級。

// TestSameSeedSameHashAcrossTurns:兩局同種子、同指令(這裡是「什麼都不做,只結束回合」),
// 每一回合的指紋都必須相同。
//
// 這支測試同時是 map 迭代順序的護欄:`ColonyBuildings` 之類的 map 若被拿來直接影響狀態,
// Go 的隨機迭代順序會讓兩局在某個回合分岔。
func TestSameSeedSameHashAcrossTurns(t *testing.T) {
	a, b := NewDemoSession(), NewDemoSession()
	if a.StateHash() == "" {
		t.Fatal("算不出指紋(snapshot 序列化失敗)")
	}
	if a.StateHash() != b.StateHash() {
		t.Fatal("同種子的兩局開場指紋就不一樣")
	}
	for turn := 1; turn <= 120; turn++ {
		a.EndTurn()
		b.EndTurn()
		if ha, hb := a.StateHash(), b.StateHash(); ha != hb {
			t.Fatalf("第 %d 回合分岔:%s vs %s", turn, ha[:12], hb[:12])
		}
	}
}

// TestSameSeedSameHashWithPlayerCommands:加上真的玩家指令(派艦隊、拓殖、排建造、
// 設集結點)之後仍然要一致——「只結束回合」測不到指令路徑上的不決定性。
func TestSameSeedSameHashWithPlayerCommands(t *testing.T) {
	script := func(s *GameSession, turn int) {
		switch turn {
		case 1:
			// 派艦隊去一顆看得見的星(挑法要決定性:取索引最小的可見星)。
			vis := s.VisibleStars()
			for i := range s.Stars {
				if i != s.Fleet().AtStar && i < len(vis) && vis[i] {
					s.SendFleet(i)
					break
				}
			}
		case 3:
			s.EnqueueBuild(0, "住宅", 60)
		case 5:
			if star := s.Fleet().AtStar; star >= 0 {
				s.ColonizeStar(star)
			}
		case 7:
			s.SetColonyRelocation(0, 1)
		case 9:
			s.SetAllStarRelocations(2)
		}
	}
	a, b := NewDemoSession(), NewDemoSession()
	for turn := 1; turn <= 60; turn++ {
		script(a, turn)
		script(b, turn)
		a.EndTurn()
		b.EndTurn()
		if ha, hb := a.StateHash(), b.StateHash(); ha != hb {
			t.Fatalf("第 %d 回合分岔:%s vs %s", turn, ha[:12], hb[:12])
		}
	}
}

// TestSaveLoadRoundTripKeepsHash:存檔 → 讀檔,**當下**的指紋必須一樣。
//
// 這條是存檔完整性:讀回來的局面若與存檔當下不同,那就是有欄位沒進存檔。
func TestSaveLoadRoundTripKeepsHash(t *testing.T) {
	s := NewDemoSession()
	for i := 0; i < 12; i++ {
		s.EndTurn()
	}
	want := s.StateHash()

	path := filepath.Join(t.TempDir(), "determinism.json")
	if err := s.Save(path); err != nil {
		t.Fatalf("存檔失敗:%v", err)
	}
	got, err := LoadSession(path)
	if err != nil {
		t.Fatalf("讀檔失敗:%v", err)
	}
	if h := got.StateHash(); h != want {
		t.Errorf("讀檔後的指紋與存檔當下不同:\n  存 %s\n  讀 %s", want, h)
	}
}

// TestLoadedGameContinuesTheSameRandomStreams:存檔 → 讀檔 → 再跑 N 回合,
// 必須與「不存檔直接跑 N 回合」得到同一個指紋。
//
// 這條先前**不成立**：長壽命亂數流沒有把「抽到第幾個數」存進去，
// 讀檔之後整條流從頭開始——存檔洗事件毫無成本,而且網路對戰時中途讀檔的那台會與其他人分岔。
// 修法見 randstream.go(每次抽取恰好消耗一個原始值,所以「快轉 n 次」就只是丟掉 n 個值)。
func TestLoadedGameContinuesTheSameRandomStreams(t *testing.T) {
	base := NewDemoSession()
	for i := 0; i < 12; i++ {
		base.EndTurn()
	}
	path := filepath.Join(t.TempDir(), "determinism.json")
	if err := base.Save(path); err != nil {
		t.Fatalf("存檔失敗:%v", err)
	}
	loaded, err := LoadSession(path)
	if err != nil {
		t.Fatalf("讀檔失敗:%v", err)
	}
	if base.StateHash() != loaded.StateHash() {
		t.Fatal("測試前提不成立:存檔當下的指紋就對不上(那是 TestSaveLoadRoundTripKeepsHash 的事)")
	}
	// 測試前提二:那幾條流真的抽過東西——都還沒抽的話這支測試等於什麼都沒驗。
	if base.eventRand.Draws() == 0 {
		t.Fatal("測試前提不成立:12 回合下來事件流一次都沒抽過")
	}

	for i := 0; i < 40; i++ {
		base.EndTurn()
		loaded.EndTurn()
	}
	if hb, hl := base.StateHash(), loaded.StateHash(); hb != hl {
		t.Errorf("讀檔之後的後續發展與不讀檔不同:\n  不讀檔 %s\n  讀檔後 %s", hb, hl)
	}
}

// TestRandStreamFastForwardLandsOnTheSameValue 是快轉本身的單元護欄。
//
// 這條之所以成立,是因為 randStream 的每一次抽取**恰好消耗一個原始值**——
// 直接用 math/rand 的 Intn/Float64 是做不到的(兩者從底層取走的數量不一樣)。
func TestRandStreamFastForwardLandsOnTheSameValue(t *testing.T) {
	const seed = 987654321
	a := newRandStream(seed)
	// 混著抽兩種,正是「種類不同就跳不準」的情境。
	for i := 0; i < 17; i++ {
		if i%3 == 0 {
			a.Float64()
		} else {
			a.Intn(37)
		}
	}
	b := restoreRandStream(seed, a.Draws())
	for i := 0; i < 5; i++ {
		if x, y := a.Intn(1000), b.Intn(1000); x != y {
			t.Fatalf("快轉後第 %d 個值就對不上:%d vs %d", i, x, y)
		}
	}
}

// TestHashChangesWhenStateChanges 是指紋本身的正對照:
// 如果指紋對任何改動都不敏感,上面三支測試就等於什麼都沒驗。
func TestHashChangesWhenStateChanges(t *testing.T) {
	s := NewDemoSession()
	before := s.StateHash()
	s.Player.BC += 1
	if after := s.StateHash(); after == before {
		t.Fatal("改了國庫指紋卻沒變——指紋沒有涵蓋到狀態,上面幾支測試都失效")
	}
	s.Player.BC -= 1
	if s.StateHash() != before {
		t.Error("改回來之後指紋應該回到原值(指紋要只依狀態,不依歷史)")
	}
}

// TestRuleVersionSurvivesSaveLoad:主選單選的規則版本(1.3 / 1.5)要撐得過存讀檔。
//
// 這是 TestLoadedGameContinuesTheSameRandomStreams 抓出來的第二個 bug:
// `RuleProfile` **完全沒進存檔**,讀檔後是零值——那既不是 1.3 也不是 1.5,
// 而是「Version=1.3 但所有數值欄位都是 0」的混種:Hyper-Advanced 研究成本、電漿砲傷害、
// 轟炸輪數、守方 Commando 加成、感測器加成、貨運現金加成全部歸零。
//
// CLAUDE.md 把「允許在主選單選擇版本」列為專案目標,而那個選擇撐不過一次存檔。
func TestRuleVersionSurvivesSaveLoad(t *testing.T) {
	for _, want := range []gamedata.RuleProfile{gamedata.Profile13(), gamedata.Profile15()} {
		s := NewDemoSession()
		s.SetRuleProfile(want)
		path := filepath.Join(t.TempDir(), "rule.json")
		if err := s.Save(path); err != nil {
			t.Fatalf("存檔失敗:%v", err)
		}
		got, err := LoadSession(path)
		if err != nil {
			t.Fatalf("讀檔失敗:%v", err)
		}
		if got.RuleProfile != want {
			t.Errorf("版本 %d 的 profile 沒撐過存讀檔:\n  存 %+v\n  讀 %+v",
				want.Version, want, got.RuleProfile)
		}
	}
}

// 舊存檔沒有 ruleVersion 欄位 → 0 → 必須還原成**完整的** Profile13(),
// 不是「Version=1.3 但欄位全 0」的那個零值混種。
func TestLegacySaveWithoutRuleVersionGetsFullProfile13(t *testing.T) {
	s := NewDemoSession()
	s.SetRuleProfile(gamedata.Profile13())
	snap := s.snapshot()
	snap.RuleVersion = 0 // 模擬舊存檔(欄位不存在 → 零值)
	got := snap.restore()
	if got.RuleProfile != gamedata.Profile13() {
		t.Errorf("舊存檔應還原成完整的 Profile13(),實得 %+v", got.RuleProfile)
	}
	if got.RuleProfile.HyperAdvancedLevel1Cost == 0 {
		t.Error("還原出來的 profile 數值欄位是 0——那正是這個 bug 的樣子")
	}
}
