package shell

import "testing"

// TestSetupNewGameRebuildsGalaxyAndAI 驗證 SetupNewGame(正式新遊戲流程 customrace.go/
// raceselect.go 的 applyAndStart 改呼叫的入口)重生星系後,AI 對手的 ColonyStars 真的指向
// 「新」星系裡被標成 AI(Owner==2)的星索引——不是舊版 RegenGalaxy 那種「星系換了但 AIPlayers
// 沒跟著換」的 stale 資料,對照 buildDemoAIOpponents/SetupNewGame 的實作說明。
func TestSetupNewGameRebuildsGalaxyAndAI(t *testing.T) {
	s := NewDemoSession()

	s.SetupNewGame(36, 123, 3)

	if len(s.AIPlayers) != 3 {
		t.Fatalf("SetupNewGame(numAI=3) 應建 3 個 AI 對手,got %d", len(s.AIPlayers))
	}
	for i, a := range s.AIPlayers {
		if len(a.ColonyStars) != 1 {
			t.Fatalf("AI[%d] 應恰有 1 顆已知母星索引,got %d", i, len(a.ColonyStars))
		}
		star := a.ColonyStars[0]
		if star < 1 || star >= 36 {
			t.Fatalf("AI[%d] 母星索引 %d 應落在新星系範圍 [1,36),不是玩家母星 0", i, star)
		}
		if s.Stars[star].Owner != 2 {
			t.Fatalf("AI[%d] 母星(新星系星 %d)Owner 應為 2(AI),got %d——代表星系與 AIPlayers 對不上號",
				i, star, s.Stars[star].Owner)
		}
	}

	if len(s.PlayerSpies) != 3 {
		t.Fatalf("PlayerSpies 應與新 AIPlayers 平行重置為長度 3,got %d", len(s.PlayerSpies))
	}
	for i, n := range s.PlayerSpies {
		if n != 0 {
			t.Fatalf("PlayerSpies[%d] 重置後應為 0,got %d", i, n)
		}
	}

	if s.Stars[0].Owner != 1 {
		t.Fatalf("新星系星 0 應仍是玩家母星(Owner==1),got %d", s.Stars[0].Owner)
	}

	if len(s.Planets) != len(s.Stars) {
		t.Fatalf("genPlanets 後 Planets 長度應與 Stars 一致,got %d vs %d", len(s.Planets), len(s.Stars))
	}

	// 連續推進回合不應 panic(含 AI 造艦/擴張、議會資格判斷等既有流程)。
	for i := 0; i < 5; i++ {
		s.EndTurn()
	}
	_ = s.councilEligible() // 至少不 panic;3 AI + 玩家 = 4 存續帝國,資格判斷本身不應出錯
}

// TestRegenGalaxyStillWorksAsOneAICompat 驗證 RegenGalaxy 轉呼叫 SetupNewGame(n, seed, 1) 後,
// 舊「只需 1 AI」語意的呼叫端仍能拿到 1 個 AI 對手、且星系/AI 對得上號(不是退化成不建 AI)。
func TestRegenGalaxyStillWorksAsOneAICompat(t *testing.T) {
	s := NewDemoSession()

	s.RegenGalaxy(24, 99)

	if len(s.AIPlayers) != 1 {
		t.Fatalf("RegenGalaxy 應轉呼叫 SetupNewGame(numAI=1),got %d 個 AI", len(s.AIPlayers))
	}
	star := s.AIPlayers[0].ColonyStars[0]
	if s.Stars[star].Owner != 2 {
		t.Fatalf("RegenGalaxy 後唯一 AI 母星(星 %d)Owner 應為 2,got %d", star, s.Stars[star].Owner)
	}
}
