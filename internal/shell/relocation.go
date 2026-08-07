package shell

// relocation.go:**遷移(集結點)**——星圖 4 層裡最後一層的規則面。
//
// 原版 `Draw_Relocation_Links_` @ 0x85320 是主畫面第 2 層(見 cmd/moo2/shipicon.go 檔頭的
// 圖層順序)。remake 先前把它列為缺口,卡在「艦隊是單一集合」;多艦隊做出來之後
// (見 fleet.go),缺的只剩這一層自己的資料。
//
// ============ 資料模型是真值 ============
//
// 兩支函式讀的是**同一個欄位**,互相印證:
//
//	sub_784F0(星, 玩家) → word[星×0x71 + 0x54 + 玩家×2]        ; 遷移目標星
//	sub_78C94(星, 玩家) → 上面那個欄位 != -1                     ; 「有沒有設定」
//
// 也就是**每個(星, 玩家)一個目標星索引**,−1 = 沒設定。
// `Draw_Relocation_Links_` 的迴圈就是:走每一顆星,目前玩家在那裡有設定就畫一條線過去。
//
// remake 的殖民地是 `PlayerColonies[i]` + 平行的 `PlayerColonyStars[i]`,對玩家而言
// 「(星, 玩家)」與「第 i 個殖民地」是一對一,所以這裡用平行陣列 `ColonyRelocateTo`。
//
// ============ 手冊給了它的作用 ============
//
// 「Relocation Lines controls the appearance of travel lines for those ships being
//  automatically relocated between star systems. (You set up your Relocation orders on the
//  Fleet Operations console.)」
//
// 兩件事:① 新造的艦會**自動**被送往集結點;② 那條線有**顯示開關**
// (原版 `byte_199BE4`,手冊那組 ALT+Fn 設定裡的一項——⚠ 是不是 F6 沒有確認,
//  見 gap report 第 54 項對 PDF 邊欄標籤的保留)。

// ColonyRelocationNone 是「沒有設定集結點」(原版的 −1)。
const ColonyRelocationNone = -1

// SetColonyRelocation 設定第 i 個殖民地的集結點。
//
// star 給 ColonyRelocationNone 或該殖民地自己所在的星 = 取消(新艦留在原地)。
// 回傳是否有生效。
func (s *GameSession) SetColonyRelocation(colony, star int) bool {
	if colony < 0 || colony >= len(s.PlayerColonies) {
		return false
	}
	if star != ColonyRelocationNone && (star < 0 || star >= len(s.Stars)) {
		return false
	}
	s.growRelocation()
	if star == s.colonyStar(colony) {
		star = ColonyRelocationNone // 送往自己 = 不用送
	}
	s.ColonyRelocateTo[colony] = star
	return true
}

// ColonyRelocation 回傳第 i 個殖民地的集結點(沒設定回 ColonyRelocationNone)。
func (s *GameSession) ColonyRelocation(colony int) int {
	if colony < 0 || colony >= len(s.ColonyRelocateTo) {
		return ColonyRelocationNone
	}
	return s.ColonyRelocateTo[colony]
}

// growRelocation 把平行陣列補齊到殖民地數,新格補 ColonyRelocationNone。
//
// ⚠ 不能用 Go 零值當預設:0 是**星 0(母星)**的索引,那會讓每個新殖民地一建好就把新艦
// 全部往母星送。這是本檔最容易埋的錯——補齊時一定要填 −1。
func (s *GameSession) growRelocation() {
	for len(s.ColonyRelocateTo) < len(s.PlayerColonies) {
		s.ColonyRelocateTo = append(s.ColonyRelocateTo, ColonyRelocationNone)
	}
	if len(s.ColonyRelocateTo) > len(s.PlayerColonies) {
		s.ColonyRelocateTo = s.ColonyRelocateTo[:len(s.PlayerColonies)]
	}
}

// RelocationLink 是星圖上要畫的一條遷移連線。
type RelocationLink struct{ From, To int }

// RelocationLinks 回傳目前要畫的所有遷移連線(原版那支迴圈的結果)。
//
// 過濾掉起訖相同與越界的;顯示開關關掉時回 nil(對齊原版:那支函式開頭就檢查 `byte_199BE4`,
// 關掉是**整層不畫**,不是畫得淡一點)。
func (s *GameSession) RelocationLinks() []RelocationLink {
	if !s.ShowRelocationLines {
		return nil
	}
	var out []RelocationLink
	for i := range s.PlayerColonies {
		to := s.ColonyRelocation(i)
		from := s.colonyStar(i)
		if to == ColonyRelocationNone || from < 0 || to < 0 ||
			from >= len(s.Stars) || to >= len(s.Stars) || from == to {
			continue
		}
		out = append(out, RelocationLink{From: from, To: to})
	}
	return out
}

// deliverNewShip 把某個殖民地剛造好的艦交付出去:有集結點就自動送過去,沒有就留在原地。
//
// ⚠ 「自動送過去」在 remake 是**建一支往目的地航行的艦隊**(或併進已經要去那裡的那一支),
// 而不是瞬間移動——手冊說的是 "ships being automatically relocated",那是一段航程,
// 星圖上畫得出來的那條線就是它。
func (s *GameSession) deliverNewShip(colony int, sh Ship) {
	from := s.colonyStar(colony)
	to := s.ColonyRelocation(colony)
	if to == ColonyRelocationNone || from < 0 || to < 0 || from == to {
		s.AddShipToHomeFleet(from, sh)
		return
	}
	// 已經有一支從同一顆星出發、往同一個目的地的艦隊 → 併進去(不然每造一艘就多一支艦隊)。
	for i := range s.Fleets {
		if s.Fleets[i].AtStar == from && s.Fleets[i].DestStar == to {
			s.Fleets[i].Ships = append(s.Fleets[i].Ships, sh)
			return
		}
	}
	f := NewFleet(from)
	f.Ships = []Ship{sh}
	f.DestStar = to
	f.ETA = s.FleetETATo(from, to)
	if f.ETA < 1 {
		f.ETA = 1 // 至少一回合;0 會被 advanceFleet 當成「已抵達」而永遠不推進
	}
	s.Fleets = append(s.Fleets, f)
}
