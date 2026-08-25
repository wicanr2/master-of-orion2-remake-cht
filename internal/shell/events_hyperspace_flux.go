package shell

func originalHyperspaceFluxEnds(age, roll1Based int) bool {
	if age > 20 {
		return true
	}
	return age > 4 && roll1Based == 1
}

// hyperspaceFluxActive 對應 sub_233FA：原版狀態 2／4／6 都算 active；remake 的 typed
// record 不保存 GNN 展示子狀態，因此 record 存在就代表規則仍生效。
func (s *GameSession) hyperspaceFluxActive() bool {
	for i := range s.PersistentEvents {
		if s.PersistentEvents[i].Kind == PersistentHyperspaceFlux {
			return true
		}
	}
	return false
}

func (s *GameSession) startHyperspaceFlux() (string, bool) {
	if s.hyperspaceFluxActive() {
		return "", false
	}
	// advancePersistentEvents 在 step 前先 Turns++；以 -1 讓建立當回合的第一次 consumer
	// 看見 raw age=0，對齊 sub_2230A 寫零後同回合進 sub_206A2。
	s.PersistentEvents = append(s.PersistentEvents, PersistentEvent{Kind: PersistentHyperspaceFlux, StarIndex: -1, Turns: -1})
	return "超空間亂流席捲銀河，非跨維度艦隊的星際航行暫時停滯", true
}

func (s *GameSession) stepHyperspaceFlux(e *PersistentEvent) (bool, string, string) {
	// sub_206A2：raw age >4 才擲 1/20；raw age >20 時共用尾端強制結束。
	ended := e.Turns > 20
	if !ended && e.Turns > 4 {
		ended = originalHyperspaceFluxEnds(e.Turns, s.eventRoll(20))
	}
	if ended {
		return true, "超空間亂流已消散，星際航線恢復通行",
			"The hyperspace flux dissipated; interstellar travel has resumed."
	}
	return false, "", ""
}
