package shell

import (
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// outpost.go:軍事前哨站(Outpost)。
//
// remake 先前完全沒有這個系統——氣態巨星與小行星帶只是「不能殖民的星」,玩家對它們無事可做。
// 原版裡它們是前哨站的地盤,而前哨站是「把版圖往外推」的主要手段。
//
// --- 手冊逐字(GAME_MANUAL.pdf,patch1.5 隨附完整手冊,pdftotext 直接萃取,非 OCR)---
//
//	p.85 Outpost Ship:"An Outpost Ship is similar to a Colony Ship, except that it is used to
//	     establish a military outpost in a system, rather than a new colony. A military outpost
//	     extends the reach of your scanners and the range of your ships, but produces nothing.
//	     Since there are no full-time residents, an outpost does not need to be established on a
//	     habitable world; you can put them on gas giants and in asteroid belts. If a colony is
//	     created at an outpost, the building remains and is repurposed as Marine Barracks.
//	     An Outpost Ship requires 1 Command Rating point or 10 BC in maintenance each turn until
//	     it is dismantled to build an outpost."
//	p.56  "You can build a military outpost in a close orbit around a gas giant (and in an
//	      asteroid belt), but colonies can only survive on a solid planet."
//	p.119 "Outpost Ships each build a military outpost on a single planet. These outposts act as
//	      scanning stations and as refueling stops for fleets."
//	p.133 "No fleet can travel any farther from one of your colonies or outposts than the
//	      shortest range of any ship in that fleet."
//
// 由這幾段可以逐條落地的效果:
//   1. 只能由前哨船建立,建完消耗該船(與殖民船同款)。
//   2. **不限宜居行星**——氣態巨星/小行星帶正是它的用途;一般行星也可以(手冊沒排除)。
//   3. **沒有人口、沒有任何產出**("produces nothing"),所以不建 ColonyState,
//      不進 PlayerColonies——這點很重要,否則帝國經濟會憑空多出一個殖民地。
//   4. 掃描站:併進偵測源(detection.go),讓星圖的可見範圍真的往外推。
//   5. 之後在同一顆星建殖民地時,前哨站「改建為海軍陸戰隊營」(手冊逐字),
//      即新殖民地起始就有 Marine Barracks 這棟建築。
//
// ⚠ 尚未建模、誠實留白的部分:
//   - 「加油站 / 延伸艦艇航程」(p.119/p.133):remake 的 SendFleet 沒有航程上限這個概念
//     (只有距離換算 ETA),沒有可套用的機制,不臆造一個。前哨站因此目前只兌現「掃描站」
//     這一半。等航程系統補上時,這裡是它的掛勾點。
//   - 手冊 p.50 提到有科技能把氣態巨星/小行星帶的前哨站升級成可住人殖民地
//     (行星固定 Barren/Normal-G/Abundant,氣態巨星化為 Huge、小行星帶化為 Large),
//     那需要對應的科技旗標,待科技樹接上後再做。

// OutpostShipClass 是前哨船的艦體等級字串(命名慣例同 ColonyShipClass)。
const OutpostShipClass = "前哨船"

// OutpostBuildName 是建造選單裡的前哨船項目名。
const OutpostBuildName = OutpostShipClass

// OutpostMarineBarracks 是前哨站改建後留下的建築名(手冊:"repurposed as Marine Barracks")。
// 值必須與 gamedata 的建築名一致,否則 applyBuildingEffect 查不到、等於白給一棟空建築。
const OutpostMarineBarracks = "海軍陸戰隊營"

// Outpost 是一座軍事前哨站。
type Outpost struct {
	StarIndex int // 所在星索引
	Turn      int // 建立回合(供顯示/未來的年資規則)
}

// findOutpostShipIndex 回傳玩家艦隊中第一艘前哨船在 s.Ships 的索引;找不到回 -1。
func (s *GameSession) findOutpostShipIndex() int {
	for i, sh := range s.Ships {
		if sh.Class == OutpostShipClass {
			return i
		}
	}
	return -1
}

// FleetHasOutpostShip 回傳玩家艦隊是否載有至少一艘前哨船(供 UI 決定是否顯示「建前哨站」鈕)。
func (s *GameSession) FleetHasOutpostShip() bool { return s.findOutpostShipIndex() >= 0 }

// HasOutpostAt 回傳玩家在該星是否已有前哨站。
func (s *GameSession) HasOutpostAt(starIdx int) bool {
	for _, o := range s.Outposts {
		if o.StarIndex == starIdx {
			return true
		}
	}
	return false
}

// OutpostResult 是一次建前哨站嘗試的結果(欄位風格對稱 ColonizationResult)。
type OutpostResult struct {
	Ok     bool
	Reason string // Ok=false 時的原因
	Star   int    // Ok=true 時的星索引
}

// BuildOutpost 在 starIdx 建立軍事前哨站。前置條件:
//  1. 玩家艦隊已抵達該星(FleetAtStar==starIdx 且 FleetETA==0),與 ColonizeStar 同款。
//  2. 該星無主(Owner==0)——別人的地盤不能插旗。
//  3. 艦隊載有前哨船。
//  4. 該星有天體可建(NoPlanet 的黑洞不行)。
//  5. 該星尚無自己的前哨站。
//
// 成功:消耗一艘前哨船、記一筆 Outpost、把該星標成玩家所有(Owner=1)並標記已探索。
// **不建立殖民地**——手冊「produces nothing」,前哨站沒有人口也沒有產出。
func (s *GameSession) BuildOutpost(starIdx int) OutpostResult {
	if starIdx < 0 || starIdx >= len(s.Stars) {
		return OutpostResult{Reason: "星索引無效"}
	}
	if s.FleetAtStar != starIdx || s.FleetETA != 0 {
		return OutpostResult{Reason: "艦隊尚未抵達該星"}
	}
	if s.Stars[starIdx].Owner != 0 {
		return OutpostResult{Reason: "該星已有歸屬,不可建立前哨站"}
	}
	if s.HasOutpostAt(starIdx) {
		return OutpostResult{Reason: "該星已有前哨站"}
	}
	// 怪獸擋路(見 monster.go)。手冊只對殖民船寫明這條,但怪獸就盤據在那個星系裡,
	// 沒有理由前哨船能無視它——比照處理,並在此標明這是延伸而非手冊逐字。
	if reason := s.monsterBlockReason(starIdx); reason != "" {
		return OutpostResult{Reason: reason}
	}
	shipIdx := s.findOutpostShipIndex()
	if shipIdx < 0 {
		return OutpostResult{Reason: "艦隊未載運前哨船"}
	}
	if starIdx < len(s.Planets) && s.Planets[starIdx].NoPlanet {
		return OutpostResult{Reason: "該星系沒有可供建立前哨站的天體"}
	}

	s.Ships = append(s.Ships[:shipIdx], s.Ships[shipIdx+1:]...) // 消耗這艘前哨船
	s.Outposts = append(s.Outposts, Outpost{StarIndex: starIdx, Turn: s.Turn})
	s.Stars[starIdx].Owner = 1
	s.Stars[starIdx].Explored = true
	return OutpostResult{Ok: true, Star: starIdx}
}

// outpostStarIndices 回傳玩家所有前哨站所在的星索引(供偵測來源使用)。
func (s *GameSession) outpostStarIndices() []int {
	if len(s.Outposts) == 0 {
		return nil
	}
	out := make([]int, 0, len(s.Outposts))
	for _, o := range s.Outposts {
		out = append(out, o.StarIndex)
	}
	return out
}

// consumeOutpostForColony 在 starIdx 建立殖民地時處理既有前哨站:移除該筆前哨站,
// 回傳是否有前哨站被改建(手冊:"If a colony is created at an outpost, the building remains
// and is repurposed as Marine Barracks")。呼叫端據此替新殖民地補上那棟建築。
func (s *GameSession) consumeOutpostForColony(starIdx int) bool {
	for i, o := range s.Outposts {
		if o.StarIndex != starIdx {
			continue
		}
		s.Outposts = append(s.Outposts[:i], s.Outposts[i+1:]...)
		return true
	}
	return false
}

// OutpostBuildAvailable 回傳前哨船是否已可建造。
//
// 手冊沒把前哨船列在任何科技之下(它與殖民船一樣是開局就有的支援艦),故恆可建造——
// 這與 AvailableBuildOptions 裡「住宅」恆可用是同一個處理,不是漏了 gate。
func OutpostBuildAvailable() bool { return true }

// planetSupportsOutpost 回傳該行星資料是否適合建前哨站(供 UI 顯示提示用)。
// 手冊 p.85:前哨站不需要宜居世界,氣態巨星與小行星帶正是它的用途;一般行星也可以。
// 只有「沒有任何天體」的黑洞星系不行。
func planetSupportsOutpost(p Planet) bool { return !p.NoPlanet }

// OutpostTargetHint 回傳在該星建前哨站的用途提示(供 UI 顯示)。
func (s *GameSession) OutpostTargetHint(starIdx int) string {
	if starIdx < 0 || starIdx >= len(s.Planets) {
		return ""
	}
	p := s.Planets[starIdx]
	if !planetSupportsOutpost(p) {
		return "無天體可用"
	}
	if p.TypeID != gamedata.HABITABLE && p.TypeID != 0 {
		return planetTypeDisplayName(p.TypeID) + "・僅能建前哨站"
	}
	return "可建前哨站(掃描站)"
}
