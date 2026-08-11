package shell

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func TestSpyMissionDefaultsAndSupportedToggle(t *testing.T) {
	s := NewDemoSession()
	s.PlayerSpyMissions = nil // 模擬加入任務欄位前的舊存檔
	if got := s.SpyMissionFor(0); got != SpyMissionSteal {
		t.Fatalf("舊存檔缺任務欄位時應退回 STEAL,got %v", got)
	}
	if len(s.PlayerSpyMissions) != len(s.AIPlayers) {
		t.Fatalf("任務陣列應延遲補齊到 AI 數量,got %d want %d",
			len(s.PlayerSpyMissions), len(s.AIPlayers))
	}
	if !s.SetSpyMission(0, SpyMissionHide) {
		t.Fatal("HIDE 應可設定")
	}
	if got := s.SpyMissionFor(0); got != SpyMissionHide {
		t.Fatalf("設定 HIDE 後讀回 %v", got)
	}
	if !s.SetSpyMission(0, SpyMissionSabotage) {
		t.Fatal("SABOTAGE 已有原版建築破壞結算,應可設定")
	}
	if got := s.SpyMissionFor(0); got != SpyMissionSabotage {
		t.Fatalf("設定 SABOTAGE 後讀回 %v", got)
	}
}

func TestCycleSpyMissionCyclesAllSupportedTasks(t *testing.T) {
	s := NewDemoSession()
	if got, ok := s.CycleSpyMission(0); !ok || got != SpyMissionSabotage {
		t.Fatalf("STEAL 循環後應為 SABOTAGE,got (%v,%v)", got, ok)
	}
	if got, ok := s.CycleSpyMission(0); !ok || got != SpyMissionHide {
		t.Fatalf("SABOTAGE 循環後應為 HIDE,got (%v,%v)", got, ok)
	}
	if got, ok := s.CycleSpyMission(0); !ok || got != SpyMissionSteal {
		t.Fatalf("HIDE 循環後應為 STEAL,got (%v,%v)", got, ok)
	}
	if _, ok := s.CycleSpyMission(len(s.AIPlayers)); ok {
		t.Fatal("越界目標不應成功切換任務")
	}
}

func TestSpyMissionHideSkipsTheftAndUsesSpyVsSpy(t *testing.T) {
	attacker := engine.PlayerState{
		CompletedTopics: map[gamedata.ResearchTopic]bool{gamedata.TOPIC_STARTING_TECH: true},
	}
	defender := engine.PlayerState{
		CompletedTopics: map[gamedata.ResearchTopic]bool{
			gamedata.TOPIC_STARTING_TECH:         true,
			gamedata.TOPIC_ADVANCED_CONSTRUCTION: true,
		},
		ChosenTech: map[gamedata.ResearchTopic]gamedata.Technology{
			gamedata.TOPIC_ADVANCED_CONSTRUCTION: gamedata.TECH_AUTOMATED_FACTORIES,
		},
	}
	before := clonePlayerState(attacker)
	msgs, killed := spyMissionAttempt(rand.New(rand.NewSource(1)), SpyMissionHide,
		&attacker, defender, 63, "我方", "AI", 0, 0, 0)
	if killed {
		t.Fatal("基準 HIDE 測試不應擊殺攻方")
	}
	if len(msgs) == 0 || !strings.Contains(msgs[0], "隱匿") {
		t.Fatalf("HIDE 應留下可辨識的結算訊息,got %v", msgs)
	}
	if len(attacker.CompletedTopics) != len(before.CompletedTopics) ||
		attacker.CompletedTopics[gamedata.TOPIC_ADVANCED_CONSTRUCTION] {
		t.Fatalf("HIDE 不應走偷科技支線,got %+v", attacker.CompletedTopics)
	}
}

func TestSpyMissionPersistsAndRestores(t *testing.T) {
	s := NewDemoSession()
	if !s.SetSpyMission(1, SpyMissionHide) {
		t.Fatal("設定第 2 個對手的 HIDE 失敗")
	}
	restored := s.snapshot().restore()
	if got := restored.SpyMissionFor(1); got != SpyMissionHide {
		t.Fatalf("存讀檔後任務 = %v,預期 HIDE", got)
	}
}
