package netplay

import "testing"

func TestTurnReadyRequiresAllPlayersAndMatchingReplayHash(t *testing.T) {
	tab := NewTable(2, 4)
	if err := tab.Add(Message{Kind: KindTurnDone, Player: 0, Turn: 4}); err != nil {
		t.Fatal(err)
	}
	if err := tab.Add(Message{Kind: KindTurnDone, Player: 1, Turn: 4}); err != nil {
		t.Fatal(err)
	}
	if tab.ReadyComplete() {
		t.Fatal("尚未收到 turn_ready 不應視為完成")
	}
	if err := tab.AddReady(Message{Kind: KindTurnReady, Player: 0, Turn: 4, StateHash: "same"}); err != nil {
		t.Fatal(err)
	}
	if tab.ReadyComplete() {
		t.Fatal("只收到一位玩家 ready 不應視為完成")
	}
	if err := tab.AddReady(Message{Kind: KindTurnReady, Player: 1, Turn: 4, StateHash: "other"}); err != nil {
		t.Fatal(err)
	}
	if !tab.ReadyComplete() || tab.ReadyDesync() == "" {
		t.Fatal("兩個不同的重播指紋應該被拒絕")
	}
}

func TestTurnTableRejectsUnknownPlayerWhenSized(t *testing.T) {
	tab := NewTable(2, 1)
	if err := tab.Add(Message{Kind: KindTurnDone, Player: 2, Turn: 1}); err == nil {
		t.Fatal("不在名冊內的玩家不能加入回合表")
	}
	if err := tab.AddReady(Message{Kind: KindTurnReady, Player: -1, Turn: 1, StateHash: "x"}); err == nil {
		t.Fatal("不在名冊內的玩家不能送 turn_ready")
	}
}
