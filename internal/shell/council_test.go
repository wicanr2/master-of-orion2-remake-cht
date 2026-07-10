package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
)

// settleHalfGalaxy 把 s.Stars 前一半(含母星)標為玩家所有,滿足 councilEligible 的
// 「半數銀河已殖民」條件,不動 AI 母星那顆(維持 owner=2)。供議會相關測試共用。
func settleHalfGalaxy(s *GameSession) {
	need := (len(s.Stars) + 1) / 2
	settled := 0
	for i := range s.Stars {
		if s.Stars[i].Owner != 0 {
			settled++
			continue
		}
		if settled >= need {
			break
		}
		s.Stars[i].Owner = 1
		settled++
	}
}

// TestCouncilNotEligibleEarlyGame 驗證新遊戲開局(僅 2 顆星有主,24 顆星系)議會尚未成立
// ——手冊 GAME_MANUAL.pdf p.183「When half of the galaxy has been settled...」門檻未達。
func TestCouncilNotEligibleEarlyGame(t *testing.T) {
	s := NewDemoSession()
	if s.councilEligible() {
		t.Fatalf("開局星系殖民率過低,議會不應成立")
	}
	s.EndTurn()
	if s.LastCouncil != "" {
		t.Fatalf("議會未成立時不應有 LastCouncil 訊息,got %q", s.LastCouncil)
	}
	if s.CouncilMeetings != 0 {
		t.Fatalf("議會未成立時 CouncilMeetings 應為 0,got %d", s.CouncilMeetings)
	}
}

// TestCouncilEligibleAfterHalfSettled 驗證半數銀河殖民 + 2 個存續帝國(本 remake 資料模型
// 覆寫門檻,見 councilMinExtantRacesOverride)後議會成立,且首次達成當回合立即開會(不用等
// councilInterval)。
func TestCouncilEligibleAfterHalfSettled(t *testing.T) {
	s := NewDemoSession()
	settleHalfGalaxy(s)
	if !s.councilEligible() {
		t.Fatalf("半數銀河已殖民且存續帝國數達標,議會應已成立")
	}
	s.EndTurn()
	if s.CouncilMeetings != 1 {
		t.Fatalf("議會成立後應立即召開第 1 屆,got CouncilMeetings=%d", s.CouncilMeetings)
	}
	if s.LastCouncil == "" {
		t.Fatalf("已開會應留下 LastCouncil 訊息")
	}
}

// TestCouncilPlayerWinsBySupermajority 驗證玩家人口達 2/3 多數時直接當選勝利,不需要
// RespondToCouncilElection(手冊:議會無法強迫的是「別人當選、你不同意」,不適用於你自己當選)。
func TestCouncilPlayerWinsBySupermajority(t *testing.T) {
	s := NewDemoSession()
	settleHalfGalaxy(s)
	s.PlayerColonies[0].Population = 100
	s.AIPlayers[0].Colonies[0].Population = 1

	s.EndTurn()

	if !s.Victory.Over {
		t.Fatalf("玩家人口壓倒性領先,應已達2/3多數當選")
	}
	if s.Victory.Reason != engine.VictoryHighCouncil {
		t.Fatalf("勝利路徑應為 engine.VictoryHighCouncil,got %v", s.Victory.Reason)
	}
	if s.Victory.Winner != "player" {
		t.Fatalf("當選者應為 player,got %q", s.Victory.Winner)
	}
	if s.PendingCouncilElection != nil {
		t.Fatalf("玩家自己當選不應留下待回應選舉")
	}
}

// TestCouncilEnemyWinsRequiresPlayerResponse 驗證 AI 達 2/3 多數時不會直接結束遊戲,而是
// 留下 PendingCouncilElection 讓玩家用 RespondToCouncilElection 回應(手冊:「there's no way
// the council can force you to accept a decision you don't agree with」)。拒絕後遊戲不結束,
// 下一屆(councilInterval 回合後)可以再開會。
func TestCouncilEnemyWinsRequiresPlayerResponse(t *testing.T) {
	s := NewDemoSession()
	settleHalfGalaxy(s)
	s.PlayerColonies[0].Population = 1
	s.AIPlayers[0].Colonies[0].Population = 100

	s.EndTurn()

	if s.Victory.Over {
		t.Fatalf("AI當選不應自動結束遊戲,應等玩家回應")
	}
	if s.PendingCouncilElection == nil {
		t.Fatalf("AI 達2/3多數應留下待回應選舉")
	}
	pending := *s.PendingCouncilElection

	// 拒絕:不結束遊戲,清空待決狀態。
	s.RespondToCouncilElection(false)
	if s.Victory.Over {
		t.Fatalf("拒絕接受後遊戲不應結束")
	}
	if s.PendingCouncilElection != nil {
		t.Fatalf("回應後 PendingCouncilElection 應清空")
	}

	// 再跑到下一屆(councilInterval 回合後),讓 AI 再次當選,這次接受。
	for i := 0; i < councilInterval && !s.Victory.Over && s.PendingCouncilElection == nil; i++ {
		s.EndTurn()
	}
	if s.PendingCouncilElection == nil {
		t.Fatalf("下一屆選舉應再度觸發(AI 人口仍壓倒性領先)")
	}
	s.RespondToCouncilElection(true)
	if !s.Victory.Over || s.Victory.Reason != engine.VictoryHighCouncil || s.Victory.Winner != pending.EnemyName {
		t.Fatalf("接受後應以 engine.VictoryHighCouncil 結束,Winner=%q,got Over=%v Reason=%v Winner=%q",
			pending.EnemyName, s.Victory.Over, s.Victory.Reason, s.Victory.Winner)
	}
}

// TestCouncilNoOneReachesSupermajority 驗證票數接近(未達2/3)時流會,不分勝負,下一屆再開。
func TestCouncilNoOneReachesSupermajority(t *testing.T) {
	s := NewDemoSession()
	settleHalfGalaxy(s)
	s.PlayerColonies[0].Population = 50
	s.AIPlayers[0].Colonies[0].Population = 50 // 50/50,遠低於2/3門檻

	s.EndTurn()

	if s.Victory.Over {
		t.Fatalf("票數五五波未達2/3多數,不應分出勝負")
	}
	if s.PendingCouncilElection != nil {
		t.Fatalf("五五波不應有一方達標,不該留下待回應選舉")
	}
	if s.CouncilMeetings != 1 {
		t.Fatalf("應已召開第1屆(流會),got CouncilMeetings=%d", s.CouncilMeetings)
	}
}

// TestCouncilDisableEventsSkips 驗證 DisableEvents=true(供確定性經濟測試隔離用)時議會不開會
// ——比照 advanceAntares/advanceEvents 既有慣例。
func TestCouncilDisableEventsSkips(t *testing.T) {
	s := NewDemoSession()
	settleHalfGalaxy(s)
	s.DisableEvents = true
	s.EndTurn()
	if s.CouncilMeetings != 0 {
		t.Fatalf("DisableEvents=true 時議會不應召開,got CouncilMeetings=%d", s.CouncilMeetings)
	}
}

// TestConquestVictoryWhenAllOpponentsEliminated 驗證手冊第一條勝利路徑:AI 對手的殖民地清單
// 一旦全空(InvadeColony 攻陷其唯一殖民地後的狀態,見 ground_invasion.go),即判定玩家以
// engine.VictoryExtermination 勝利。直接操作 Colonies 切片以聚焦驗證勝利偵測本身,不重跑整套地面戰。
func TestConquestVictoryWhenAllOpponentsEliminated(t *testing.T) {
	s := NewDemoSession()
	s.AIPlayers[0].Colonies = nil

	s.advanceConquestVictory()

	if !s.Victory.Over {
		t.Fatalf("所有 AI 對手殖民地已清空,應判定玩家勝利")
	}
	if s.Victory.Reason != engine.VictoryExtermination {
		t.Fatalf("勝利路徑應為 engine.VictoryExtermination,got %v", s.Victory.Reason)
	}
	if s.Victory.Winner != "player" {
		t.Fatalf("殲滅所有對手的勝利者應為 player,got %q", s.Victory.Winner)
	}
}

// TestConquestVictoryNotTriggeredWhileOpponentAlive 驗證 AI 仍有殖民地時不會誤判勝利。
func TestConquestVictoryNotTriggeredWhileOpponentAlive(t *testing.T) {
	s := NewDemoSession()
	s.advanceConquestVictory()
	if s.Victory.Over {
		t.Fatalf("AI 對手仍有殖民地,不應判定勝利")
	}
}
