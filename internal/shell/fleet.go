package shell

// fleet.go:**多艦隊模型**。
//
// remake 先前把玩家的兵力表示成一組欄位:`Ships` + `FleetAtStar` / `FleetDestStar` /
// `FleetETA` / `FleetMarines` / `FleetTanks`。也就是**全帝國只有一支艦隊**,
// 所有的船永遠在同一個地方、只能有一個航行任務。
//
// ============ 這個限制卡住的東西 ============
//
//   - **星圖的遷移連線層**(原版 `Draw_Relocation_Links_` @ 0x85320,4 層裡的最後 1 層):
//     那是「新造的艦從哪個殖民地送到哪」的連線,前提是艦隊有多支、有出發地。
//   - **F1 / F2 循環艦隊**(手冊的快捷鍵):循環集合只有一個元素,按下去等於原地不動。
//   - 分艦隊、同時多線任務——原版最基本的操作。
//
// ============ ⚠ 改寫時最容易犯的錯:`Ships` 有兩種語意 ============
//
// 舊程式碼裡的 `s.Ships` 混著兩件事,而**單一艦隊時兩者剛好相同**,所以分不出來:
//
//	① 「**這支艦隊**的船」——戰鬥、載運陸戰隊、消耗殖民船、航行中損失。
//	② 「**全帝國**的船」——指揮點數(手冊 p.169 明文是全帝國)、國力、艦名編號、外交評估。
//
// 盲目改成 `s.Fleet().Ships` 會讓第 ② 類在**真的有第二支艦隊時**默默算少,
// 而那時候看起來完全正常(數字只是偏小)。所以這裡給**兩個**存取器,
// 改寫時逐處決定用哪一個,並用「兩支艦隊」的測試把分類釘住——
// 不能等到多艦隊 UI 做出來才發現分錯。

// Fleet 是一支艦隊:一群船 + 位置 + 航行任務 + 隨行的地面部隊。
type Fleet struct {
	Ships []Ship
	// AtStar 是目前所在星索引。航行中維持在**出發星**(remake 的航行是整段跳的,
	// 中途沒有位置——原版 `Get_Ship_Icon_Coords_` 那套逐格插值沒有對應物)。
	AtStar int
	// DestStar 是目的星索引(−1 = 無航行任務)。
	DestStar int
	// ETA 是抵達尚需回合數(0 = 已抵達 / 靜止)。
	ETA int
	// Marines / Tanks 是隨這支艦隊出征、已載運的陸戰隊與戰車營
	// (共用同一個運力池,見 ground_invasion.go 的 MarineTransportCapacity)。
	Marines int
	Tanks   int
}

// NewFleet 回傳一支停在 at 星、沒有任務的空艦隊。
func NewFleet(at int) Fleet { return Fleet{AtStar: at, DestStar: -1} }

// ensureFleet 維持「永遠至少有一支艦隊」的不變量,並把 SelectedFleet 夾回合法範圍。
//
// 為什麼要有這條不變量:`Fleet()` 回傳指標讓呼叫端直接讀寫,沒有艦隊時就得回 nil,
// 於是每一個呼叫端都要 nil 檢查——那種檢查漏一個就是 panic。
// 改成「艦隊可以沒有船,但一定存在」,呼叫端的形狀和單艦隊時代完全一樣。
func (s *GameSession) ensureFleet() {
	if len(s.Fleets) == 0 {
		s.Fleets = []Fleet{NewFleet(0)}
	}
	if s.SelectedFleet < 0 || s.SelectedFleet >= len(s.Fleets) {
		s.SelectedFleet = 0
	}
}

// Fleet 回傳**目前操作中**的艦隊(可直接讀寫)。
//
// 用於:移動與派遣、載運/卸下地面部隊、在場戰鬥、消耗殖民船 / 前哨船。
// 要「全帝國的船」請用 AllShips / ShipCount,見檔頭 ⚠。
func (s *GameSession) Fleet() *Fleet {
	s.ensureFleet()
	return &s.Fleets[s.SelectedFleet]
}

// FleetAt 回傳第 i 支艦隊(越界回 nil)。
func (s *GameSession) FleetAt(i int) *Fleet {
	if i < 0 || i >= len(s.Fleets) {
		return nil
	}
	return &s.Fleets[i]
}

// SelectFleet 把操作對象切到第 i 支艦隊(越界則不動)。
func (s *GameSession) SelectFleet(i int) {
	s.recordPlayerCommand(PlayerCommand{Name: CmdSelectFleet, Args: []int{i}})
	if i < 0 || i >= len(s.Fleets) {
		return
	}
	s.SelectedFleet = i
}

// ShipCount 是**全帝國**的艦艇總數。
func (s *GameSession) ShipCount() int {
	n := 0
	for i := range s.Fleets {
		n += len(s.Fleets[i].Ships)
	}
	return n
}

// AllShips 依序展平**全帝國**的艦艇。
//
// 回傳的是新切片(元素是複本),不能拿來改船的狀態——要改請用 EachShip。
func (s *GameSession) AllShips() []Ship {
	out := make([]Ship, 0, s.ShipCount())
	for i := range s.Fleets {
		out = append(out, s.Fleets[i].Ships...)
	}
	return out
}

// EachShip 走訪**全帝國**的每一艘船,給的是可寫指標(修復、上損傷等要用這個)。
func (s *GameSession) EachShip(fn func(sh *Ship)) {
	for i := range s.Fleets {
		for j := range s.Fleets[i].Ships {
			fn(&s.Fleets[i].Ships[j])
		}
	}
}

// removeShipGlobal 移除**全帝國第 k 艘**船(依 AllShips 的順序),回傳被移除的船。
//
// 隨機事件的「損失一艘艦」用這個:原版的事件打的是整個帝國,不是玩家目前選中的那一支。
func (s *GameSession) removeShipGlobal(k int) (Ship, bool) {
	if k < 0 {
		return Ship{}, false
	}
	for i := range s.Fleets {
		n := len(s.Fleets[i].Ships)
		if k < n {
			sh := s.Fleets[i].Ships[k]
			s.Fleets[i].Ships = append(s.Fleets[i].Ships[:k], s.Fleets[i].Ships[k+1:]...)
			return sh, true
		}
		k -= n
	}
	return Ship{}, false
}

// AddShipToHomeFleet 把新造的艦加進「停在該殖民地的艦隊」;沒有這樣的艦隊就加進第一支。
//
// ⚠ 這是**接近**原版而不是等同:原版新艦出現在生產它的殖民地,並依該殖民地的遷移設定
// 自動送往集結點(`Draw_Relocation_Links_` 畫的就是那條線)。remake 還沒有逐殖民地的
// 造艦佇列位置資訊,所以先用「艦隊剛好停在那顆星就併進去」這條規則。
// 逐殖民地造艦做出來之後,這裡要改成真的依生產地放置。
func (s *GameSession) AddShipToHomeFleet(star int, sh Ship) {
	s.ensureFleet()
	for i := range s.Fleets {
		if s.Fleets[i].AtStar == star {
			s.Fleets[i].Ships = append(s.Fleets[i].Ships, sh)
			return
		}
	}
	s.Fleets[0].Ships = append(s.Fleets[0].Ships, sh)
}

// colonyStar 回傳玩家第 i 個殖民地所在的星索引(未知回 −1)。
//
// `PlayerColonyStars` 是平行陣列且可能有 −1 padding(見該欄位註解),所以要走這支而不是直接索引。
func (s *GameSession) colonyStar(i int) int {
	if i < 0 || i >= len(s.PlayerColonyStars) {
		return -1
	}
	return s.PlayerColonyStars[i]
}

// ---- 分艦隊(原版 `Split_Stack_` @ 0x5D689)----
//
// 原版的「艦隊」叫 **ship stack**:`word_192248[stack]` 是頭一艘船的 id,
// `word_1975D6[船 id×5]` 是「下一艘」的指標——**單向串列**。
// `Split_Stack_` 收一組船 id,把它們從原本的串列摘下來、串成新的一個 stack,
// 接在 stack 表的尾端(`word_199A02` 是目前的 stack 數)。
//
// 所以拆分的語意是:**選一組船,抽出來組成一支新艦隊**,位置不變。
// remake 用切片而不是串列,但語意逐項對得上。

// SplitFleet 把第 fleetIdx 支艦隊裡的 ships(艦隊內索引)抽出來組成新的一支。
//
// 回傳新艦隊的索引與是否成功。以下情況不動作:
//   - 艦隊索引越界、沒有選任何船、選了全部(那不是拆分,是原地不動)
//   - **艦隊正在航行**——remake 的航行是整段跳的,中途沒有位置,
//     拆出來的那一半沒有可放的地方。這是 remake 移動模型的後果,不是原版規則。
//
// ⚠ 陸戰隊/戰車營**留在原艦隊**:remake 把它們建模成艦隊層級的數字(不綁定到特定的船),
// 所以拆分時沒有「哪些跟著走」的依據。要改成逐船攜行得先讓運兵成為船的屬性。
func (s *GameSession) SplitFleet(fleetIdx int, ships []int) (int, bool) {
	args := append([]int{fleetIdx}, ships...)
	s.recordPlayerCommand(PlayerCommand{Name: CmdSplitFleet, Args: args})
	s.ensureFleet()
	if fleetIdx < 0 || fleetIdx >= len(s.Fleets) {
		return -1, false
	}
	src := &s.Fleets[fleetIdx]
	if src.DestStar >= 0 {
		return -1, false // 航行中不能拆(見上方 ⚠)
	}
	// 去重 + 濾掉越界,順便算出「要留下的」。
	take := map[int]bool{}
	for _, i := range ships {
		if i >= 0 && i < len(src.Ships) {
			take[i] = true
		}
	}
	if len(take) == 0 || len(take) == len(src.Ships) {
		return -1, false
	}
	moved := make([]Ship, 0, len(take))
	kept := make([]Ship, 0, len(src.Ships)-len(take))
	for i, sh := range src.Ships {
		if take[i] {
			moved = append(moved, sh)
		} else {
			kept = append(kept, sh)
		}
	}
	src.Ships = kept
	f := NewFleet(src.AtStar)
	f.Ships = moved
	s.Fleets = append(s.Fleets, f)
	return len(s.Fleets) - 1, true
}
