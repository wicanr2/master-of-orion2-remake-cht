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

// ============ 原版的設定流程:兩段點選,而且有明確的合法性規則 ============
//
// `Star_Relocation_` @ 0x75180 收「起點指標、終點指標、剛點到的星」:
//
//	if *起點 == −1:  驗證這顆星能不能當**起點** → 通過就記起來,結束
//	else:            驗證能不能當**終點** → 通過就記起來;**選到同一顆 = 取消**
//
// 合法性在 `Okay_To_Set_Relocate_Star_` @ 0x75035(`dl` 區分是不是終點):
//
//	① 光譜 6(黑洞)→ 不行(起點終點都不行,只是訊息不同:0x83 / 0x84)
//	② `Player_Has_Visited_` 為假 → 不行(訊息 0x85 / 0x86)
//	③ `Star_Guarded_By_Monster_` 為真:
//	     終點 → sprintf(訊息 0x87, 星, `Race_Name_(怪獸)`) 之後
//	            `User_Box_(kind=1)` = **是/否確認框**,答案就是這條規則的結果
//	     起點 → 直接不行(而且**不出訊息**:`loc_7511B` 只把結果清 0)
//	④ 起點另外要求 `Player_Has_Colony_In_System_`(訊息 0x88)
//
// ⚠ **2026-08-07 訂正**:這段原本寫著「③ 目的星上有艦隊 → 跳確認框」,
// 而且註明「remake 沒有 modal 對話框的基礎設施,目前直接允許」。
// 逐指令讀過 `Okay_To_Set_Relocate_Star_` 之後:**那個條件是怪獸不是艦隊**
// (`sub_7A47A` = `Star_Guarded_By_Monster_`,符號表裡就有名字)。
// 確認框的基礎設施也做了(見 cmd/moo2/confirmbox.go,版面取自 `Confirmation_Box_` @ 0x77658)。

// ============ AI 不設集結點——那是**原版行為**,不是 remake 的缺口 ============
//
// gap report 一度把「AI 的遷移設定」列為缺口(「AI 沒有逐星的艦隊位置,所以沒有遷移可設」)。
// 2026-08-07 逐一追過那個欄位的**所有寫入者**之後:原版的 AI 也不設集結點。
//
//	Universe_Generation_        把 [星 + 玩家×2 + 0x54] 對 8 個玩家全部初始化成 −1
//	Set_Relocation_             唯一呼叫端是 Star_Relocation_(玩家的兩段點選)
//	Clear_Star_Relocation_      唯一呼叫端也是 Star_Relocation_(點回同一顆 = 取消)
//	Set_All_Star_Relocations_   呼叫端是星圖輸入處理器 sub_73980 與 Main_Screen_
//	Clear_All_Star_Relocations_ 呼叫端是 sub_73980
//
// 五個寫入者**沒有一個在 AI 的程式碼裡**。讀取端 `Redirect_Newly_Built_Ships_` 確實是
// 逐玩家跑的(它收 player 參數、查 `Has_Relocation_(星, 玩家)`),所以 AI 的欄位有人讀——
// 只是永遠是 −1。欄位之所以逐玩家,是因為星球結構本來就替 8 個玩家各留一格
// (多人對戰時每個**人類**玩家用自己那格)。
//
// ⚠ 方法上的一個坑,記下來:第一次用 `grep '\*2+54h]'` 找寫入者時**漏掉了兩個**
// (`Set_All` / `Clear_All` 把 `星基 + 玩家×2` 先加好再 `mov [eax+54h], bx`,
// 定址式裡沒有 `*2`)。正確做法是先切出每一支函式,再找「同時碰 `dword_19306C`、`71h`
// 與 `+54h]`」的那些——那組條件把兩個漏網的都撈了回來,也就是這個結論的正對照。
//
// 所以 remake 這邊**什麼都不用做**。要替 AI 加集結點會是加一條原版沒有的規則。

// RelocateRefusal 是「這顆星不能當起點/終點」的原因(空字串 = 可以)。
type RelocateRefusal string

// CanRelocateFrom 檢查某顆星能不能當遷移起點,回傳不行的原因(空 = 可以)。
func (s *GameSession) CanRelocateFrom(star int) RelocateRefusal {
	if star < 0 || star >= len(s.Stars) {
		return "沒有這顆星"
	}
	if s.Stars[star].Spectral == blackHoleSpectral {
		return "黑洞不能當遷移起點"
	}
	if !s.Stars[star].Explored {
		return "還沒探索過這顆星"
	}
	// 怪獸盤據的星當**起點**直接不行——原版連訊息都不出(`loc_7511B` 只把結果清 0)。
	// remake 出一句話:靜默失敗在有滑鼠提示的介面裡只會讓玩家以為按鈕壞了。
	if s.StarGuardedByMonster(star) {
		return RelocateRefusal("那裡被" + s.MonsterNameAtStar(star) + "盤據,不能當遷移起點")
	}
	if colonyIndexAt(s, star) < 0 {
		return "那裡沒有你的殖民地——遷移是從自己的殖民地送出去的"
	}
	return ""
}

// CanRelocateTo 檢查某顆星能不能當遷移終點,回傳不行的原因(空 = 可以)。
//
// ⚠ 怪獸不在這裡擋——原版對**終點**的怪獸是問一句(是/否確認框),不是拒絕。
// 見 RelocateToNeedsConfirm。
func (s *GameSession) CanRelocateTo(star int) RelocateRefusal {
	if star < 0 || star >= len(s.Stars) {
		return "沒有這顆星"
	}
	if s.Stars[star].Spectral == blackHoleSpectral {
		return "黑洞不能當遷移終點"
	}
	if !s.Stars[star].Explored {
		return "還沒探索過這顆星"
	}
	return ""
}

// RelocateToNeedsConfirm 回傳「設成這顆終點之前要先問玩家的那句話」;空字串 = 不用問。
//
// 原版 `Okay_To_Set_Relocate_Star_` 的第 ③ 條:終點被怪獸盤據時
// sprintf(訊息 0x87, 星名, 怪獸名) 之後跳 `User_Box_(kind=1)`(是/否)。
// **它不是拒絕**——玩家說是就照設,新造的艦會一艘艘送進怪獸的嘴裡,那是玩家的選擇。
//
// ⚠ 訊息 0x87 的原文在 LBX 的字串表裡,remake 沒有逐字抄(那是遊戲文字不是規則);
// 這裡用等義的中文,並保留「星名 + 怪獸名」這兩個原版會填進去的參數。
func (s *GameSession) RelocateToNeedsConfirm(star int) string {
	if star < 0 || star >= len(s.Stars) {
		return ""
	}
	if !s.StarGuardedByMonster(star) {
		return ""
	}
	return s.Stars[star].Name + "被" + s.MonsterNameAtStar(star) +
		"盤據,送過去的艦艇會遭到攻擊。仍要把集結點設在那裡嗎?"
}

// colonyIndexAt 回傳玩家在某顆星的殖民地索引(沒有回 −1)。
func colonyIndexAt(s *GameSession, star int) int {
	for i, st := range s.PlayerColonyStars {
		if st == star && i < len(s.PlayerColonies) {
			return i
		}
	}
	return -1
}

// SetStarRelocation 是原版那條路徑:用**起點星**而不是殖民地索引來設定。
//
// 起訖同一顆 = 取消(原版 `Cancel_Star_Relocation_`)。回傳不行的原因(空 = 成功)。
func (s *GameSession) SetStarRelocation(from, to int) RelocateRefusal {
	if r := s.CanRelocateFrom(from); r != "" {
		return r
	}
	if r := s.CanRelocateTo(to); r != "" {
		return r
	}
	if !s.SetColonyRelocation(colonyIndexAt(s, from), to) {
		return "設定失敗"
	}
	return ""
}

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

// ---- 一次改全部(原版 `Set_All_Star_Relocations_` @ 0x785EC / `Clear_All_Star_Relocations_` @ 0x77BB1)----
//
// 艦隊列表上的 **ALL**(remake 譯「全部」)鈕多半就是這一對(第 59 項記下的推測)。

// SetAllStarRelocations 把**已經有集結點**的殖民地全部改送到同一顆星。
//
// ⚠ **這不是「把每個殖民地都設成這顆星」** —— 原版那支的迴圈裡有一道檢查:
//
//	if word[星 + 0x54 + 玩家×2] != −1:      ; 只改已經有設定的
//	    word[...] = 目標星
//
// 沒設過的殖民地**不會被順便設上**。這是一個猜不到的細節:直覺會做成「全部設成這顆」,
// 而那會讓玩家按一下就把所有新殖民地的產出全部抽走。
//
// 回傳被改到幾個。
func (s *GameSession) SetAllStarRelocations(to int) int {
	if to < 0 || to >= len(s.Stars) {
		return 0
	}
	s.growRelocation()
	n := 0
	for i := range s.ColonyRelocateTo {
		if s.ColonyRelocateTo[i] == ColonyRelocationNone {
			continue // 沒設過的不動(見上方 ⚠)
		}
		if s.colonyStar(i) == to {
			s.ColonyRelocateTo[i] = ColonyRelocationNone // 送往自己 = 取消(同單筆的規則)
		} else {
			s.ColonyRelocateTo[i] = to
		}
		n++
	}
	return n
}

// ClearAllStarRelocations 清掉所有集結點,回傳清掉幾個。
func (s *GameSession) ClearAllStarRelocations() int {
	n := 0
	for i := range s.ColonyRelocateTo {
		if s.ColonyRelocateTo[i] != ColonyRelocationNone {
			s.ColonyRelocateTo[i] = ColonyRelocationNone
			n++
		}
	}
	return n
}
