package shell

// starnav.go:星圖的**鍵盤導覽**(原版 `Cycle_Ship_Icons_` @ 0x82DFF + 手冊的快捷鍵表)。
//
// ============ 兩個獨立來源指同一件事 ============
//
// 反組譯:`sub_82DFF` 由鍵盤跳表叫進來(`sub_825A8` 的 case −1001 等),`bx` 是方向——
// 0 → `inc edx` 往後找、非 0 → `dec edx` 往前找,挑到的丟給 `sub_831B1` 選取。
// 它跑的是 `word_192248[]`(已知艦隊的星索引表),邊走邊檢查擁有者。
//
// 手冊逐字:「You can cycle through the known fleets using the keyboard shortcuts
// [F1] and [F2]. The first moves you through the fleets in one direction, and the second
// takes you back in the other direction.」
//
// 同一段還給了另一組:「[F5] This changes the view to the next colonized star system.
// [F6] This returns the view to the previous colonized system.」
//
// ============ ⚠ 目前只涵蓋玩家自己的艦隊 ============
//
// 多艦隊模型做好之後(見 fleet.go),F1/F2 走的是玩家的每一支艦隊——這一段已經對上原版。
// **仍缺的是可見 AI 艦隊的循環**：AI 主力艦隊已由 ai_fleet.go 保存位置與航程，
// 但 detection.go 尚未建立敵艦可見性，因此本檔仍只循環玩家艦隊。

// KnownFleetStars 回傳「畫面上看得到艦隊」的星索引,依星索引排序、去重
// (同一顆星上有兩支艦隊只算一個落點——循環的是**視角**,停在同一顆星按第二次不會有反應)。
//
// 航行中的艦隊原版也在表裡;remake 的航行是整段跳的,中途沒有位置,所以用出發星。
func (s *GameSession) KnownFleetStars() []int {
	var out []int
	seen := map[int]bool{}
	for i := range s.Fleets {
		at := s.Fleets[i].AtStar
		if at < 0 || at >= len(s.Stars) || seen[at] {
			continue
		}
		seen[at] = true
		out = append(out, at)
	}
	sortInts(out)
	return out
}

// ColonizedStars 回傳玩家已殖民的星索引,依星索引排序。
//
// `PlayerColonyStars` 可能含 −1 padding(見該欄位註解),這裡濾掉。
func (s *GameSession) ColonizedStars() []int {
	var out []int
	seen := map[int]bool{}
	for _, idx := range s.PlayerColonyStars {
		if idx < 0 || idx >= len(s.Stars) || seen[idx] {
			continue
		}
		seen[idx] = true
		out = append(out, idx)
	}
	sortInts(out)
	return out
}

// sortInts 是就地插入排序。清單長度是殖民地數(個位數到數十),不值得引入 sort 相依。
func sortInts(a []int) {
	for i := 1; i < len(a); i++ {
		v := a[i]
		j := i - 1
		for j >= 0 && a[j] > v {
			a[j+1] = a[j]
			j--
		}
		a[j+1] = v
	}
}

// cycleStarList 回傳 list 中「cur 的下一個/上一個」;cur 不在清單裡就回第一個(往後)或
// 最後一個(往前),清單空回 −1。
//
// 環狀:走到尾接回頭。原版那支也是環狀的——它的邊界檢查是「索引 < 艦隊數」與「索引 > −1」,
// 撞到就從另一端重來。
func cycleStarList(list []int, cur int, forward bool) int {
	if len(list) == 0 {
		return -1
	}
	at := -1
	for i, v := range list {
		if v == cur {
			at = i
			break
		}
	}
	if at < 0 {
		if forward {
			return list[0]
		}
		return list[len(list)-1]
	}
	if forward {
		return list[(at+1)%len(list)]
	}
	return list[(at-1+len(list))%len(list)]
}

// CycleFleetStar 是 F1 / F2:循環切換到下一支 / 上一支已知艦隊所在的星。
// 沒有已知艦隊回 −1(呼叫端不要改選取)。
func (s *GameSession) CycleFleetStar(cur int, forward bool) int {
	return cycleStarList(s.KnownFleetStars(), cur, forward)
}

// CycleColonizedStar 是 F5 / F6:循環切換到下一個 / 上一個已殖民星系。
// 沒有殖民地回 −1。
func (s *GameSession) CycleColonizedStar(cur int, forward bool) int {
	return cycleStarList(s.ColonizedStars(), cur, forward)
}
