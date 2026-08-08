package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// 先進級開局應該有 25 個主題(六個固定 + 十九個隨機),不是只有六個。
//
// 這是 gap report 第 30 項(TECH LEVEL 第二效果)留下的缺口:原版主迴圈跑 1 / 6 / 25 次,
// 前 6 次取固定表、第 7 次起隨機挑。
func TestAdvancedTechLevelGrantsTwentyFiveTopics(t *testing.T) {
	for _, tc := range []struct {
		level, want int
		name        string
	}{
		{0, 1, "曲速前"},
		{1, 6, "一般"},
		{2, 25, "先進"},
	} {
		s := NewDemoSession()
		s.DisableEvents = true
		s.TechLevel, s.TechLevelSet = tc.level, true
		s.applyStartingTech()
		// ⚠ 扣掉 TOPIC_STARTING_TECH:那是母星一律有的「開局科技容器」
		// (原版 nxt[0]==0 的自環主題,見第 37 項(研究樹一手驗證)),不算在 1 / 6 / 25 這個計數裡。
		got := 0
		for t0 := range s.Player.CompletedTopics {
			if t0 != gamedata.TOPIC_STARTING_TECH {
				got++
			}
		}
		if got != tc.want {
			t.Errorf("%s 級應有 %d 個開局主題,得到 %d", tc.name, tc.want, got)
		}
	}
}

// 十九個隨機主題要**沿著樹往上走**,不是從同一池子抽 19 次。
// 驗法:六個固定主題全在,而且沒有重複(map 本身保證),總數對得上。
func TestStartingRandomTopicsWalkTheTree(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	s.TechLevel, s.TechLevelSet = 2, true
	s.applyStartingTech()
	for _, t0 := range gamedata.StartingTopicOrder {
		if !s.Player.CompletedTopics[t0] {
			t.Errorf("固定表的主題 %d 應該還在", t0)
		}
	}
	// 隨機挑出來的那 19 個必須都是「當時可研究」的——所以它們的前置主題也一定在。
	// 用領域表反查:每個已完成主題在它領域裡的前一個也必須已完成。
	for area, topics := range gamedata.TechTree() {
		seenGap := false
		for i, tp := range topics {
			done := s.Player.CompletedTopics[tp]
			if !done {
				seenGap = true
				continue
			}
			if seenGap {
				t.Errorf("領域 %d:主題 %d(第 %d 個)已完成,但它前面有沒完成的"+
					"——隨機挑選跳過了前置", area, tp, i)
				break
			}
		}
	}
}

// 決定性:同一顆種子重開得到同樣的開局(網路對戰與存讀檔的要求)。
func TestStartingRandomTopicsAreDeterministic(t *testing.T) {
	build := func() map[gamedata.ResearchTopic]bool {
		s := NewDemoSession()
		s.DisableEvents = true
		s.TechLevel, s.TechLevelSet = 2, true
		s.EventSeed = 4242
		s.applyStartingTech()
		return s.Player.CompletedTopics
	}
	a, b := build(), build()
	if len(a) != len(b) {
		t.Fatalf("兩次數量不同:%d vs %d", len(a), len(b))
	}
	for k := range a {
		if !b[k] {
			t.Errorf("同一顆種子重開應得到同一組主題,%d 只出現在其中一次", k)
		}
	}
}

// 不同種子應該給出不同的開局——否則「隨機」是假的。
func TestStartingRandomTopicsVaryWithSeed(t *testing.T) {
	build := func(seed int64) map[gamedata.ResearchTopic]bool {
		s := NewDemoSession()
		s.DisableEvents = true
		s.TechLevel, s.TechLevelSet = 2, true
		s.EventSeed = seed
		s.applyStartingTech()
		return s.Player.CompletedTopics
	}
	a, b := build(1), build(999)
	same := true
	for k := range a {
		if !b[k] {
			same = false
			break
		}
	}
	if same && len(a) == len(b) {
		t.Error("不同種子給出完全一樣的開局——隨機挑選沒有真的用到種子")
	}
}

// 玩家與 AI 各走一條獨立的流,不會拿到一模一樣的 19 個。
func TestPlayerAndAIGetIndependentRandomTopics(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	s.TechLevel, s.TechLevelSet = 2, true
	s.applyStartingTech()
	if len(s.AIPlayers) == 0 {
		t.Skip("沒有 AI 對手")
	}
	same := true
	for k := range s.Player.CompletedTopics {
		if !s.AIPlayers[0].Player.CompletedTopics[k] {
			same = false
			break
		}
	}
	if same {
		t.Error("玩家與 AI 拿到完全一樣的開局主題——兩條流沒有分開")
	}
}
