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
// AI 主力艦隊已有位置與實艦；本檔同時提供星圖使用的可見敵方航線 typed 查詢。
// 查詢不改 AI 航行或偵測規則，只把已存在的可見狀態交給 UI。

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

// bestPlayerScannerJamReduction 回傳玩家最佳掃描科技對敵方飛彈閃避的抵銷點數
// (MissileJamChance 的 attackerScannerBonus)。判定規則與 bestPlayerScannerParsec 完全相同,
// 只是查另一張表——**同一個科技的第二個效果**(手冊迅子/中子條目的最後一句)。
//
// ⚠ 建模取捨,不假裝是逐字還原:手冊寫的是「Tachyon Scanners **installed in ships**」,
// 也就是逐艦裝載;本 remake 的掃描科技沒有對應的可造艦元件(整套走「研究出來即全域生效」,
// 見 bestPlayerScannerParsec 的註解),所以這裡也照同一個模型算成帝國級。
// 要改成逐艦,得先有掃描器元件槽,那是另一件事。
func bestPlayerScannerJamReduction(ps engine.PlayerState) int {
	hasSpace := groundEquipTechOwned(ps, gamedata.TOPIC_PHYSICS, gamedata.TECH_SPACE_SCANNER)
	hasNeutron := groundEquipTechOwned(ps, gamedata.TOPIC_NEUTRINO_PHYSICS, gamedata.TECH_NEUTRON_SCANNER)
	hasTachyon := groundEquipTechOwned(ps, gamedata.TOPIC_TACHYON_PHYSICS, gamedata.TECH_TACHYON_SCANNER)
	return gamedata.ScannerMissileEvasionReduction(hasSpace, hasNeutron, hasTachyon)
}

// detectionSource 是一個偵測源(殖民地、前哨站或艦隊所在星)在星圖上的位置與其**額外加成**。
type detectionSource struct {
	starIdx int
	// bonusParsec 是這個偵測源在帝國基礎掃描範圍之上的額外 parsec。兩種來源:
	//   - 殖民地:軌道基地(星基/戰鬥站/星辰要塞),見 gamedata.OrbitalScannerBonusParsec
	//   - 艦隊:艦上的戰鬥掃描器(手冊「scanning range 2 parsecs greater … outside of combat」)
	// 前哨站兩者皆無,故為 0。
	bonusParsec int
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
func playerDetectionVisible(stars []Star, playerColonyStars []int, fleetSources []detectionSource, colonyBuildings []map[string]bool, scannerParsec, versionBonusParsec int, outpostStars []int) []bool {
	visible := make([]bool, len(stars))

	// 蒐集本局所有玩家偵測源:各殖民地所在星(含軌道基地加成)+ 每支艦隊所在星 + 前哨站。
	var sources []detectionSource
	for i, idx := range playerColonyStars {
		if idx < 0 || idx >= len(stars) {
			continue // PlayerColonyStars 可能有 -1 padding(見該欄位註解),略過未知星索引
		}
		bonus := 0
		if i < len(colonyBuildings) {
			bonus = gamedata.OrbitalScannerBonusParsec(colonyBuildings[i])
		}
		sources = append(sources, detectionSource{starIdx: idx, bonusParsec: bonus})
	}
	// ⚠ 2026-08-08(第 71 項(探針③內部函式))修正:這裡原本只收「**目前選中的那一支**」艦隊
	// (`s.Fleet().AtStar`)。多艦隊模型上線後那就變成 bug——切換選中的艦隊會改變戰爭迷霧,
	// 而「選中哪一支」是選單狀態,不該影響遊戲規則。現在每支艦隊都是偵測源。
	for _, src := range fleetSources {
		if src.starIdx >= 0 && src.starIdx < len(stars) {
			sources = append(sources, src)
		}
	}
	// 前哨站是掃描站(手冊 p.119「These outposts act as scanning stations」),與殖民地一樣
	// 是常駐偵測源;它沒有軌道基地也沒有艦上元件,故無加成。
	for _, idx := range outpostStars {
		if idx >= 0 && idx < len(stars) {
			sources = append(sources, detectionSource{starIdx: idx, bonusParsec: 0})
		}
	}

	for i, st := range stars {
		if st.Explored || st.Owner == 1 {
			visible[i] = true
			continue
		}
		for _, src := range sources {
			rangeNorm := gamedata.DetectionRangeNormalized(scannerParsec, src.bonusParsec, versionBonusParsec)
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
	if s.RaceOmniscience() {
		visible := make([]bool, len(s.Stars))
		for i := range visible {
			visible[i] = true
		}
		return visible
	}
	scannerParsec := bestPlayerScannerParsec(s.Player)
	return playerDetectionVisible(s.Stars, s.PlayerColonyStars, s.fleetDetectionSources(), s.ColonyBuildings,
		scannerParsec, s.RuleProfile.SensorRangeVersionBonusParsec, s.outpostStarIndices())
}

// EnemyFleetVisibleToPlayer 回報玩家是否能在目前 remake 的抽象模型中看見
// AI 的主力艦隊。全知直接看見；其餘情況沿用該 AI 艦隊所在星的星圖可見性。
// AI 艦隊目前沒有獨立的長距離 blip 座標，故不假裝提供逐艦傳感器精度。
func (s *GameSession) EnemyFleetVisibleToPlayer(aiIdx int) bool {
	if s == nil || aiIdx < 0 || aiIdx >= len(s.AIPlayers) {
		return false
	}
	if s.RaceOmniscience() {
		return true
	}
	starIdx := aiFleetStar(s.AIPlayers[aiIdx])
	if starIdx < 0 || starIdx >= len(s.Stars) {
		return false
	}
	visible := s.VisibleStars()
	return starIdx < len(visible) && visible[starIdx]
}

// EnemyFleetMove 是目前可向玩家顯示的一支在途 AI 主力艦隊。
// 只保存穩定索引與規則資料；顏色、線型及動畫 tick 屬 UI，不進存檔。
type EnemyFleetMove struct {
	AIIndex  int
	FromStar int
	ToStar   int
	ETA      int
}

// VisibleEnemyFleetMoves 實作 SETTINGS 的 Enemy Moves 玩家可見契約。
//
// 原版 byte_199BDF 沒有已閉合的直接玩法 xref；help 只證實「允許看見敵方移動」。
// remake 因而採保守顯示閘門：起點與目的地都必須在目前可見星圖內，避免由航線洩漏霧區。
func (s *GameSession) VisibleEnemyFleetMoves() []EnemyFleetMove {
	if s == nil || !s.EffectiveGameSettings().EnemyMoves {
		return nil
	}
	visible := s.VisibleStars()
	seen := func(star int) bool {
		return star >= 0 && star < len(visible) && visible[star]
	}
	moves := make([]EnemyFleetMove, 0, len(s.AIPlayers))
	for i, ai := range s.AIPlayers {
		from, to := aiFleetStar(ai), ai.FleetDestStar
		if ai.FleetETA <= 0 || from == to || !seen(from) || !seen(to) {
			continue
		}
		moves = append(moves, EnemyFleetMove{
			AIIndex: i, FromStar: from, ToStar: to, ETA: ai.FleetETA,
		})
	}
	return moves
}

// PlayerFleetVisibleToAI 回報 AI 是否能偵測玩家艦隊。匿蹤艦在目前 AI
// 沒有獨立傳感器掃描器的抽象模型中採保守的完全遮蔽；一般艦隊只要與某個
// AI 主力艦隊同星就被視為接觸。這個 API 供外交／突襲與未來 blip UI 共用。
func (s *GameSession) PlayerFleetVisibleToAI(aiIdx int) bool {
	if s == nil || aiIdx < 0 || aiIdx >= len(s.AIPlayers) {
		return false
	}
	if s.RaceStealthyShips() {
		return false
	}
	aiStar := aiFleetStar(s.AIPlayers[aiIdx])
	if aiStar < 0 {
		return false
	}
	for _, fleet := range s.Fleets {
		if fleet.AtStar == aiStar && fleet.ETA == 0 {
			return true
		}
	}
	return false
}

// fleetDetectionSources 把**每一支**艦隊變成一個偵測源,額外加成取該艦隊裡最好的艦上掃描器。
//
// 「取最好的一個而不是加總」與軌道基地同慣例(星辰要塞取代戰鬥站,不疊加):裝了三台戰鬥掃描器
// 的艦隊不會看得比裝一台遠——手冊那句的主詞是「ships equipped with Battle Scanners」,
// 講的是**這艘船**的掃描範圍,不是艦隊的疊加。
func (s *GameSession) fleetDetectionSources() []detectionSource {
	out := make([]detectionSource, 0, len(s.Fleets))
	for i := range s.Fleets {
		f := &s.Fleets[i]
		if f.AtStar < 0 {
			continue
		}
		bonus := 0
		for _, sh := range f.Ships {
			if shipHasSpecial(sh, battleScannerName) {
				bonus = gamedata.ShipBattleScannerScanParsecBonus
				break
			}
		}
		out = append(out, detectionSource{starIdx: f.AtStar, bonusParsec: bonus})
	}
	return out
}

// starVisible 是 VisibleStars 對單一星索引的便利包裝(主要供測試/未來零星呼叫端使用;
// cmd/moo2 逐幀繪製整張星圖時應改呼叫 VisibleStars 一次取整個陣列,避免每顆星都重算一次)。
func (s *GameSession) starVisible(starIdx int) bool {
	if starIdx < 0 || starIdx >= len(s.Stars) {
		return false
	}
	return s.VisibleStars()[starIdx]
}
