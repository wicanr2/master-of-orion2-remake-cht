package shell

import "testing"

func TestDiplomacyResponseTargetsNamedOpponent(t *testing.T) {
	s := NewDemoSession()
	if len(s.AIPlayers) < 2 {
		t.Fatal("測試需要至少兩個 AI 對手")
	}
	before := make([]int, len(s.AIPlayers))
	for i, a := range s.AIPlayers {
		before[i] = a.Relation
	}
	target := s.AIPlayers[1].Name
	if got := s.DiplomacyResponse("trade", target); got.Code == "" {
		t.Fatal("貿易提議應回傳外交回應")
	}
	want := clampRelation(before[1] + 10*(100+s.raceDiploBonusPct())/100)
	if s.AIPlayers[1].Relation != want {
		t.Fatalf("指定對手關係 = %d,預期 %d", s.AIPlayers[1].Relation, want)
	}
	if s.AIPlayers[0].Relation != before[0] {
		t.Fatalf("指定第 2 個對手時不應改動第 1 個對手,got %d want %d",
			s.AIPlayers[0].Relation, before[0])
	}
}

func TestDiplomacyResponseActionsHaveBoundedEffects(t *testing.T) {
	s := NewDemoSession()
	target := s.AIPlayers[0].Name
	s.AIPlayers[0].Relation = 40
	s.DiplomacyResponse("peace", target)
	if s.AIPlayers[0].Relation != 40 {
		t.Fatalf("和平提議後關係不應超過上限,got %d", s.AIPlayers[0].Relation)
	}
	s.AIPlayers[0].Relation = -40
	s.DiplomacyResponse("threat", target)
	if s.AIPlayers[0].Relation != -40 {
		t.Fatalf("威脅後關係不應低於下限,got %d", s.AIPlayers[0].Relation)
	}
	if got := s.DiplomacyResponse("unknown", target); got.Code != "" {
		t.Fatalf("未知外交動作應回空字串,got %q", got.Code)
	}
}

func TestDiplomacyResponseReturnsTypedCodes(t *testing.T) {
	s := NewDemoSession()
	target := s.AIPlayers[0].Name
	if got := s.DiplomacyResponse("trade", target); got.Code != DiploResultTradeStarted || got.Enemy != target {
		t.Fatalf("首次貿易結果=%+v，預期 typed trade_started", got)
	}
	if got := s.DiplomacyResponse("trade", target); got.Code != DiploResultTradeExists {
		t.Fatalf("重複貿易結果=%+v，預期 typed trade_exists", got)
	}
	s.Player.BC = 4
	if got := s.OfferCashGift(target, 10); got.Code != DiploResultCashInsufficient || got.Available != 4 || got.Amount != 10 {
		t.Fatalf("現金不足結果未保存格式參數：%+v", got)
	}
}

func TestOfferCashGiftTransfersTreasuryAndImprovesRelation(t *testing.T) {
	s := NewDemoSession()
	target := &s.AIPlayers[0]
	s.Player.BC = 40
	target.Player.BC = 7
	target.Relation = 0

	if got := s.OfferCashGift(target.Name, 10); got.Code == "" {
		t.Fatal("現金餽贈應回傳外交回應")
	}
	if s.Player.BC != 30 || target.Player.BC != 17 {
		t.Fatalf("現金餽贈國庫方向錯誤: player=%d ai=%d", s.Player.BC, target.Player.BC)
	}
	if target.Relation <= 0 {
		t.Fatalf("現金餽贈應改善關係,got %d", target.Relation)
	}
}

func TestOfferCashGiftRejectsInsufficientTreasuryWithoutMutation(t *testing.T) {
	s := NewDemoSession()
	target := &s.AIPlayers[0]
	s.Player.BC = 4
	target.Player.BC = 7
	target.Relation = -3

	if got := s.OfferCashGift(target.Name, 10); got.Code == "" {
		t.Fatal("國庫不足仍應回傳外交錯誤訊息")
	}
	if s.Player.BC != 4 || target.Player.BC != 7 || target.Relation != -3 {
		t.Fatalf("國庫不足時不應改變外交狀態: player=%d ai=%d relation=%d",
			s.Player.BC, target.Player.BC, target.Relation)
	}
}
