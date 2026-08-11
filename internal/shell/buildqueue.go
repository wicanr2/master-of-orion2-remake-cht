package shell

import "github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"

// 建造佇列(原版殖民地畫面右下的 7 格 BUILD QUEUE)。
//
// 為什麼要有:反組譯原版 `Add_Build_Queue_Fields_`(module 105)確認佇列是
//
//	queue_fields[7];  for (i = 0; i < 7; i++) queue_fields[i] = Add_Hidden_Field(E_Strings(12), 41);
//
// ——**7 格**,不是一格。remake 先前每個殖民地只有一個 `Builds[i]`,玩家每回合都得回來
// 手動指定下一項,原版則是一次排好一串、完工自動接下一項。這是流程層的差異,不只是 UI。
//
// 相容設計:`Builds[i]` 保留為「當前建造中的項目」(佇列第一格),本檔的 `BuildQueue[i]`
// 只存**後續**排隊項。既有讀 `s.Builds[i]` 的程式碼與存檔完全不受影響;完工時
// advanceBuilds 呼叫 popNextBuild 把下一項遞補上來。

// BuildQueueTotalSlots 是原版殖民地畫面的建造佇列格數(含當前建造中那一格)。
const BuildQueueTotalSlots = 7

// buildQueueBacklogMax 是「後續排隊項」的上限 = 總格數扣掉當前建造那一格。
const buildQueueBacklogMax = BuildQueueTotalSlots - 1

// ensureBuildQueue 確保 BuildQueue 的長度與殖民地數對齊(新殖民地建立後呼叫)。
func (s *GameSession) ensureBuildQueue() {
	for len(s.BuildQueue) < len(s.PlayerColonies) {
		s.BuildQueue = append(s.BuildQueue, nil)
	}
	if len(s.BuildQueue) > len(s.PlayerColonies) {
		s.BuildQueue = s.BuildQueue[:len(s.PlayerColonies)]
	}
	for len(s.AutoBuild) < len(s.PlayerColonies) {
		s.AutoBuild = append(s.AutoBuild, false)
	}
	if len(s.AutoBuild) > len(s.PlayerColonies) {
		s.AutoBuild = s.AutoBuild[:len(s.PlayerColonies)]
	}
	for len(s.RepeatBuild) < len(s.PlayerColonies) {
		s.RepeatBuild = append(s.RepeatBuild, ColonyBuild{})
	}
	if len(s.RepeatBuild) > len(s.PlayerColonies) {
		s.RepeatBuild = s.RepeatBuild[:len(s.PlayerColonies)]
	}
}

// BuildQueueFor 回傳殖民地 i 的完整佇列顯示內容:第 0 項是當前建造中的項目,其後是排隊項。
// 當前無建造時第 0 項是零值 ColonyBuild(對應原版佇列首格的空白)。供 UI 用。
func (s *GameSession) BuildQueueFor(i int) []ColonyBuild {
	if i < 0 || i >= len(s.Builds) {
		return nil
	}
	s.ensureBuildQueue()
	out := make([]ColonyBuild, 0, BuildQueueTotalSlots)
	out = append(out, s.Builds[i])
	if i < len(s.BuildQueue) {
		out = append(out, s.BuildQueue[i]...)
	}
	return out
}

// EnqueueBuild 把一個建造項排進殖民地 i 的佇列。
// 當前沒有建造中的項目時直接成為當前項,否則排到隊尾。
// 佇列已滿(含當前項共 BuildQueueTotalSlots 格)回 false。
func (s *GameSession) EnqueueBuild(i int, name string, cost int) bool {
	s.recordPlayerCommand(PlayerCommand{Name: CmdEnqueueBuild, Args: []int{i, cost}, Text: name})
	return s.enqueueBuildValue(i, ColonyBuild{Name: name, Cost: cost})
}

// SetCurrentBuild 直接指定殖民地 i 當前建造項,不動後續佇列。
// 這是既有 UI「選一個建造項」的行為,保留給不使用佇列的呼叫端。
//
// ⚠ 進度歸零:換建造項會丟掉已累積的 Progress。這與原版一致(原版換建造同樣不保留進度),
// 呼叫端若要保留應自行判斷同名再跳過。
func (s *GameSession) SetCurrentBuild(i int, name string, cost int) {
	if i < 0 || i >= len(s.Builds) {
		return
	}
	s.Builds[i] = ColonyBuild{Name: name, Cost: cost}
}

// DequeueBuild 移除殖民地 i 佇列中第 pos 格的項目(pos 以 BuildQueueFor 的索引為準:
// 0 = 當前建造中)。移除當前項時由後續項遞補;pos 越界時無動作。回傳是否真的移除了東西。
func (s *GameSession) DequeueBuild(i, pos int) bool {
	s.recordPlayerCommand(PlayerCommand{Name: CmdDequeueBuild, Args: []int{i, pos}})
	if i < 0 || i >= len(s.Builds) || pos < 0 {
		return false
	}
	s.ensureBuildQueue()
	if pos == 0 {
		if s.Builds[i].Name == "" {
			return false
		}
		removed := s.Builds[i]
		s.discardQueuedBuild(i, removed)
		s.Builds[i] = ColonyBuild{}
		s.popNextBuild(i)
		return true
	}
	q := s.BuildQueue[i]
	if pos-1 >= len(q) {
		return false
	}
	s.discardQueuedBuild(i, q[pos-1])
	s.BuildQueue[i] = append(q[:pos-1], q[pos:]...)
	return true
}

// popNextBuild 把殖民地 i 佇列的下一項遞補成當前建造項。當前項非空或佇列為空時無動作。
// 由 advanceBuilds 在一項完工後呼叫。
func (s *GameSession) popNextBuild(i int) {
	s.ensureBuildQueue()
	if i < 0 || i >= len(s.Builds) || s.Builds[i].Name != "" {
		return
	}
	if i >= len(s.BuildQueue) || len(s.BuildQueue[i]) == 0 {
		s.refreshAutoBuild(i)
		return
	}
	next := s.BuildQueue[i][0]
	s.BuildQueue[i] = s.BuildQueue[i][1:]

	// 跳過已經蓋好的建築:排隊期間同一棟可能已由別的途徑完成(例如事件給予),
	// 原版不會重複蓋。一路跳到找到還沒蓋的、或佇列空掉為止。
	for next.Name != "" && next.Refit == nil && s.buildAlreadyDone(i, next.Name) {
		if len(s.BuildQueue[i]) == 0 {
			s.refreshAutoBuild(i)
			return
		}
		next = s.BuildQueue[i][0]
		s.BuildQueue[i] = s.BuildQueue[i][1:]
	}
	if next.Name != "" && (next.Refit != nil || !s.buildAlreadyDone(i, next.Name)) {
		s.Builds[i] = next
		return
	}
	s.refreshAutoBuild(i)
}

// buildAlreadyDone 回傳殖民地 i 是否已經有這棟建築。
// Special 一次性行動(地形改造等)可重複套用,一律回 false(見 advanceBuilds 註解)。
func (s *GameSession) buildAlreadyDone(i int, name string) bool {
	if _, isSpecial := gamedata.SpecialActionByNameZH(name); isSpecial {
		return false
	}
	if name == TradeGoodsBuildName || name == HousingBuildName {
		return false // 貿易品/住宅是持續性選項,不是一次性建築
	}
	if i < 0 || i >= len(s.ColonyBuildings) || s.ColonyBuildings[i] == nil {
		return false
	}
	return s.ColonyBuildings[i][name]
}

// AvailableBuildOptions 回傳玩家目前科技可建的所有選項(含「不建造」「貿易品」兩個特殊項)。
// 供 UI 列出建造清單;內部沿用既有的 availableBuildOptions,不另建一套 gate 規則。
func (s *GameSession) AvailableBuildOptions() []ColonyBuild {
	return availableBuildOptions(s.Player.CompletedTopics)
}

// BuildQueueBacklogLen 回傳殖民地 i 後續排隊項的數量(不含當前建造)。供 UI 顯示「+N」。
func (s *GameSession) BuildQueueBacklogLen(i int) int {
	if i < 0 || i >= len(s.BuildQueue) {
		return 0
	}
	return len(s.BuildQueue[i])
}

// PlayerColonyStarIndex 回傳玩家第 idx 個殖民地所在的星索引;未知或越界回 -1。
// (PlayerColonyStars 的 -1 語意見該欄位註解。)
func (s *GameSession) PlayerColonyStarIndex(idx int) int {
	if idx < 0 || idx >= len(s.PlayerColonyStars) {
		return -1
	}
	return s.PlayerColonyStars[idx]
}

// ColonyHasBuilding 回傳殖民地 idx 是否已建有該建築。
func (s *GameSession) ColonyHasBuilding(idx int, name string) bool {
	return s.buildAlreadyDone(idx, name)
}

// ColonyBuildingNames 回傳殖民地 idx 已建建築的名稱清單(未排序)。
func (s *GameSession) ColonyBuildingNames(idx int) []string {
	if idx < 0 || idx >= len(s.ColonyBuildings) || s.ColonyBuildings[idx] == nil {
		return nil
	}
	out := make([]string, 0, len(s.ColonyBuildings[idx]))
	for n := range s.ColonyBuildings[idx] {
		out = append(out, n)
	}
	return out
}

// BuildETATurns 估算殖民地 idx 當前建造項還需幾回合完工。
//
// 用上一回合實際投入建造的產能推算(與 advanceBuilds 同一條算式:淨工業扣掉稅率抽走的那份),
// 不另立公式。無建造項、無結算資料或投入為 0 時回 0(UI 據此不顯示估計值)。
func (s *GameSession) BuildETATurns(idx int) int {
	if idx < 0 || idx >= len(s.Builds) {
		return 0
	}
	b := s.Builds[idx]
	if b.Name == "" || b.Cost <= 0 || b.Progress*2+b.ProgressHalf >= b.Cost*2 {
		return 0
	}
	if idx >= len(s.LastPlayerOutput.Colonies) {
		return 0
	}
	co := s.LastPlayerOutput.Colonies[idx]
	if co.Cybernetic {
		perTurnHalf := co.NetIndustryHalf * (100 - s.Player.TaxRate) / 100
		if perTurnHalf <= 0 {
			return 0
		}
		remainHalf := b.Cost*2 - (b.Progress*2 + b.ProgressHalf)
		return (remainHalf + perTurnHalf - 1) / perTurnHalf
	}
	perTurn := co.NetIndustry * (100 - s.Player.TaxRate) / 100
	if perTurn <= 0 {
		return 0
	}
	remain := b.Cost - b.Progress
	return (remain + perTurn - 1) / perTurn
}
