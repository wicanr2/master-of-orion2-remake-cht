package shell

import "testing"

// hotseat_test.go:熱座席位交換。
//
// 這個系統最容易壞的地方是「換人時漏搬某個欄位」——漏掉的欄位會被下一位玩家繼承,
// 而且不會報錯,只會表現成「我的國庫怎麼變成他的」。所以測試重點放在**隔離**上。

func TestSinglePlayerUnaffectedByHotseat(t *testing.T) {
	s := NewDemoSession()
	beforeAI := len(s.AIPlayers)
	if got := s.SetupHotseat(1); got != 1 {
		t.Errorf("單人局席位數應為 1,實得 %d", got)
	}
	if s.HotseatEnabled() {
		t.Error("單人局不該啟用熱座")
	}
	if len(s.AIPlayers) != beforeAI {
		t.Errorf("單人局不該動到 AI 對手數(%d → %d)", beforeAI, len(s.AIPlayers))
	}
	// 單人局按結束回合就是直接推進世界,不該卡在換人。
	if _, wrapped := s.AdvanceSeat(); !wrapped {
		t.Error("單人局的 AdvanceSeat 應直接回 wrapped=true(沒有下一席要等)")
	}
}

func TestHotseatTakesOverAIEmpires(t *testing.T) {
	s := NewDemoSession()
	beforeAI := len(s.AIPlayers)
	if beforeAI < 1 {
		t.Skip("這局沒有 AI 對手可接管")
	}
	n := s.SetupHotseat(2)
	if n != 2 {
		t.Fatalf("席位數應為 2,實得 %d", n)
	}
	if len(s.AIPlayers) != beforeAI-1 {
		t.Errorf("第 2 席應由一個 AI 對手接管(AI 數 %d → %d,期望 %d)",
			beforeAI, len(s.AIPlayers), beforeAI-1)
	}
	if s.ActiveSeat != 0 {
		t.Errorf("設定完應停在第 0 席,實得 %d", s.ActiveSeat)
	}
	// 接手的席位要有自己的殖民地,否則真人接過去是空的。
	if len(s.Seats[1].PlayerColonies) == 0 {
		t.Error("第 2 席沒有殖民地——接手的玩家會沒有東西可玩")
	}
}

func TestSeatSwapKeepsEmpiresSeparate(t *testing.T) {
	s := NewDemoSession()
	if len(s.AIPlayers) < 1 {
		t.Skip("這局沒有 AI 對手可接管")
	}
	if s.SetupHotseat(2) != 2 {
		t.Skip("接管不成立")
	}
	// 給第 0 席一個好認的狀態。
	s.Player.BC = 12345
	s.PlayerName = "第一帝國"
	seat0Colonies := len(s.PlayerColonies)
	seat0Ships := len(s.Ships)

	next, wrapped := s.AdvanceSeat()
	if next != 1 || wrapped {
		t.Fatalf("兩席局的第一次換人應到第 1 席且未繞回,實得 next=%d wrapped=%v", next, wrapped)
	}
	if s.Player.BC == 12345 {
		t.Error("換人之後國庫還是上一位玩家的——席位沒有隔離")
	}
	if s.PlayerName == "第一帝國" {
		t.Error("換人之後帝國名還是上一位玩家的")
	}

	// 第 1 席做點事,再換回去,第 0 席的狀態要原封不動。
	s.Player.BC = 999
	next, wrapped = s.AdvanceSeat()
	if next != 0 || !wrapped {
		t.Fatalf("兩席局的第二次換人應繞回第 0 席,實得 next=%d wrapped=%v", next, wrapped)
	}
	if s.Player.BC != 12345 {
		t.Errorf("回到第 0 席後國庫應為 12345,實得 %d", s.Player.BC)
	}
	if s.PlayerName != "第一帝國" {
		t.Errorf("回到第 0 席後帝國名應為「第一帝國」,實得 %q", s.PlayerName)
	}
	if len(s.PlayerColonies) != seat0Colonies || len(s.Ships) != seat0Ships {
		t.Errorf("回到第 0 席後殖民地/艦艇數應還原(%d/%d),實得 %d/%d",
			seat0Colonies, seat0Ships, len(s.PlayerColonies), len(s.Ships))
	}
}

// TestIdleSeatEmpiresAlsoAdvance 盯的是熱座最容易被漏掉的一段:非當前席位的帝國
// 也要過回合。漏掉的話遊戲照跑,只有第一位玩家的殖民地在長,其他人永遠停在開局。
func TestIdleSeatEmpiresAlsoAdvance(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true // 隔離隨機事件,只看經濟結算有沒有跑
	if len(s.AIPlayers) < 1 {
		t.Skip("這局沒有 AI 對手可接管")
	}
	if s.SetupHotseat(2) != 2 {
		t.Skip("接管不成立")
	}
	// 給第 1 席(此刻是凍結快照)一個好認的國庫值與研究進度。
	s.Seats[1].Player.BC = 500
	s.Seats[1].Player.ResearchProgress = 0
	beforeBC := s.Seats[1].Player.BC
	beforeRP := s.Seats[1].Player.ResearchProgress

	s.EndTurn()

	if s.Seats[1].Player.BC == beforeBC && s.Seats[1].Player.ResearchProgress == beforeRP {
		t.Errorf("第 1 席的帝國在結算後完全沒變(BC %d、研究 %d)——非當前席位被凍結了",
			beforeBC, beforeRP)
	}
	if s.ActiveSeat != 0 {
		t.Errorf("結算後控制權應仍在第 0 席,實得 %d", s.ActiveSeat)
	}
}

// TestHotseatSurvivesSaveLoad:席位不進存檔的話,熱座局讀回來會變單人局,
// 其他真人的帝國直接消失。
func TestHotseatSurvivesSaveLoad(t *testing.T) {
	s := NewDemoSession()
	if len(s.AIPlayers) < 1 {
		t.Skip("這局沒有 AI 對手可接管")
	}
	if s.SetupHotseat(2) != 2 {
		t.Skip("接管不成立")
	}
	s.Player.BC = 4321
	s.PlayerName = "存檔帝國"
	s.Seats[1].PlayerName = "第二帝國"

	path := t.TempDir() + "/hotseat.json"
	if err := s.Save(path); err != nil {
		t.Fatalf("存檔失敗:%v", err)
	}
	got, err := LoadSession(path)
	if err != nil {
		t.Fatalf("讀檔失敗:%v", err)
	}
	if !got.HotseatEnabled() {
		t.Fatal("讀回來不是熱座局——席位沒有進存檔")
	}
	if got.SeatCount() != 2 {
		t.Errorf("席位數應為 2,實得 %d", got.SeatCount())
	}
	if got.Player.BC != 4321 || got.PlayerName != "存檔帝國" {
		t.Errorf("第 0 席狀態沒還原:BC=%d name=%q", got.Player.BC, got.PlayerName)
	}
	if got.SeatName(1) != "第二帝國" {
		t.Errorf("第 1 席帝國名應為「第二帝國」,實得 %q", got.SeatName(1))
	}
	// 存檔前 Seats[ActiveSeat] 要同步過,否則讀回來會退回上一輪換人時的狀態。
	if got.Seats[0].Player.BC != 4321 {
		t.Errorf("存檔沒同步當前席位:Seats[0].BC=%d", got.Seats[0].Player.BC)
	}
}

func TestHotseatCappedByAvailableEmpires(t *testing.T) {
	s := NewDemoSession()
	// 要的席位比「玩家 + AI 對手」還多時,只能開到有帝國可分的數量。
	got := s.SetupHotseat(MaxHotseatSeats + 5)
	if max := 1 + len(s.AIPlayers) + got - 1; got > max {
		t.Errorf("席位數 %d 超過可分配的帝國數", got)
	}
	if got > MaxHotseatSeats {
		t.Errorf("席位數 %d 超過上限 %d", got, MaxHotseatSeats)
	}
}
