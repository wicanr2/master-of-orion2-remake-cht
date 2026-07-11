package shell

import (
	"math"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// detection.go:diff 全量表 #13(掃描/偵測距離)——輕量戰爭迷霧。啟用 Star.Explored 這個
// 「已維護但先前無人讀取」的死旗標(母星開局即 true、advanceFleet 抵達星時設 true,見
// session.go),讓「已探索」「玩家自己的星」「落在玩家偵測範圍內」三者之一即視為可見,其餘
// 星對玩家而言是未知(fog)。
//
// ⚠ 本檔只算「可見與否」,純視覺用途——不 gate 任何操作(選星/派艦/殖民/轟炸皆不受影響,
// 玩家仍可對著霧裡的星派艦探索)。真正的繪製改動在 cmd/moo2/interactive.go drawStarmap。
//
// ⚠ 不做敵艦 map blip:AI 艦隊在本 remake 是抽象戰力(AIOpponent.FleetStrength),沒有地圖
// 座標,本輪偵測範圍只用來決定「玩家看不看得到某顆星」,不處理「看不看得到某支艦隊」。

// bestPlayerScannerParsec 回傳玩家目前已解鎖的最佳掃描科技對應偵測範圍(parsec,見
// gamedata.ScannerRangeParsec)。掃描科技本身無對應 Component(componentUnlockedFor 那套走
// 元件解鎖),沿用 groundEquipTechOwned(ground_invasion.go)的「主題完成 + 未明確抉擇即視為
// 解鎖 / 已明確抉擇需選中該科技」判定規則,直接查 CompletedTopics/ExplicitChoice/ChosenTech:
//   - TECH_SPACE_SCANNER   屬 TOPIC_PHYSICS(ResearchAll:true,全解,無需明確抉擇)
//   - TECH_NEUTRON_SCANNER 屬 TOPIC_NEUTRINO_PHYSICS(與中子爆能砲同主題,二選一)
//   - TECH_TACHYON_SCANNER 屬 TOPIC_TACHYON_PHYSICS(與電漿掃描儀/超光速通訊同主題,三選一)
func bestPlayerScannerParsec(ps engine.PlayerState) int {
	hasSpace := groundEquipTechOwned(ps, gamedata.TOPIC_PHYSICS, gamedata.TECH_SPACE_SCANNER)
	hasNeutron := groundEquipTechOwned(ps, gamedata.TOPIC_NEUTRINO_PHYSICS, gamedata.TECH_NEUTRON_SCANNER)
	hasTachyon := groundEquipTechOwned(ps, gamedata.TOPIC_TACHYON_PHYSICS, gamedata.TECH_TACHYON_SCANNER)
	return gamedata.ScannerRangeParsec(hasSpace, hasNeutron, hasTachyon)
}

// detectionSource 是一個偵測源(殖民地或艦隊所在星)在星圖上的位置與其軌道基地加成。
type detectionSource struct {
	starIdx       int
	orbitalParsec int // 軌道基地(星基/戰鬥站/星辰要塞)偵測加成,艦隊沒有軌道基地故為 0
}

// playerDetectionVisible 是 GameSession.VisibleStars 的純函式核心,抽出來方便單元測試(不依賴
// *GameSession)。對 stars 每一顆星判定是否對玩家可見,回傳等長的 []bool。
//
// 可見條件(任一成立即可見):
//  1. 該星 Explored(艦隊曾抵達)。
//  2. 該星 Owner==1(玩家自己的殖民地——自己的星當然看得到)。
//  3. 落在任一玩家偵測源(每個殖民地所在星、艦隊目前所在星)的偵測範圍內。
//
// scannerParsec 是玩家目前最佳掃描科技的偵測範圍(呼叫端先用 bestPlayerScannerParsec 算好),
// versionBonusParsec 是這局遊戲版本規則的偵測加成(RuleProfile.SensorRangeVersionBonusParsec,
// #13:1.5 比 1.3 多 +1 parsec)。
func playerDetectionVisible(stars []Star, playerColonyStars []int, fleetAtStar int, colonyBuildings []map[string]bool, scannerParsec, versionBonusParsec int) []bool {
	visible := make([]bool, len(stars))

	// 蒐集本局所有玩家偵測源:各殖民地所在星(含軌道基地加成)+ 艦隊目前所在星(無軌道加成)。
	var sources []detectionSource
	for i, idx := range playerColonyStars {
		if idx < 0 || idx >= len(stars) {
			continue // PlayerColonyStars 可能有 -1 padding(見該欄位註解),略過未知星索引
		}
		orbitalParsec := 0
		if i < len(colonyBuildings) {
			orbitalParsec = gamedata.OrbitalScannerBonusParsec(colonyBuildings[i])
		}
		sources = append(sources, detectionSource{starIdx: idx, orbitalParsec: orbitalParsec})
	}
	if fleetAtStar >= 0 && fleetAtStar < len(stars) {
		sources = append(sources, detectionSource{starIdx: fleetAtStar, orbitalParsec: 0})
	}

	for i, st := range stars {
		if st.Explored || st.Owner == 1 {
			visible[i] = true
			continue
		}
		for _, src := range sources {
			rangeNorm := gamedata.DetectionRangeNormalized(scannerParsec, src.orbitalParsec, versionBonusParsec)
			origin := stars[src.starIdx]
			if math.Hypot(origin.X-st.X, origin.Y-st.Y) <= rangeNorm {
				visible[i] = true
				break
			}
		}
	}
	return visible
}

// VisibleStars 回傳這局遊戲目前每顆星是否對玩家可見(等長 []bool,索引對應 s.Stars)。供
// cmd/moo2 的 drawStarmap 決定 fog 繪製,一次算好整個星圖再逐星查表,避免逐星重算。
func (s *GameSession) VisibleStars() []bool {
	scannerParsec := bestPlayerScannerParsec(s.Player)
	return playerDetectionVisible(s.Stars, s.PlayerColonyStars, s.FleetAtStar, s.ColonyBuildings, scannerParsec, s.RuleProfile.SensorRangeVersionBonusParsec)
}

// starVisible 是 VisibleStars 對單一星索引的便利包裝(主要供測試/未來零星呼叫端使用;
// cmd/moo2 逐幀繪製整張星圖時應改呼叫 VisibleStars 一次取整個陣列,避免每顆星都重算一次)。
func (s *GameSession) starVisible(starIdx int) bool {
	if starIdx < 0 || starIdx >= len(s.Stars) {
		return false
	}
	return s.VisibleStars()[starIdx]
}
