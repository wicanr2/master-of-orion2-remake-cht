// Package gamedata:偵測/掃描範圍模型(diff 全量表 #13——掃描/偵測距離)。
//
// ⚠ 全面近似(誠實標註):原版 MOO2 手冊只定性描述「掃描科技越高階看得越遠」「星基/戰鬥站/
// 星辰要塞可增加偵測範圍」,並未公開逐科技的 parsec 數字;openorion2(本專案參考基底)是純
// 渲染殼,零掃描/雷達邏輯可援引(見 rulebook 62 靜態反追溯源——查無來源即誠實標近似,不假造
// 成精確值)。本檔數值全部是「為了讓 remake 有可玩的戰爭迷霧效果」自訂的近似值,調參目標是
// 「開局母星區可見數顆星、遠星入霧」,不是逐科技逐字重現原版數字。
package gamedata

// ParsecToNormalized 是「原版 parsec」→ remake 星圖正規化座標(0..1)的換算常數。
//
// 【近似】無原版來源可查——本專案星圖座標本身是正規化 0..1(見 shell.Star.X/Y 註解),原版
// 銀河尺度沒有對應的「單位換算表」可援引。調參依據改用本專案實際的程序化星系(genGalaxy)
// 量出來的星距,而非憑空假設銀河跨度:NewDemoSession() 預設 24 星(種子 42),鄰近星實測
// 距母星約 0.25(星 1)、0.28(星 5)、0.41(星 6,AI 母星)……(internal/shell/detection_test.go
// TestGameSession_VisibleStars_Homeworld 有印出實測值)。取 1 parsec = 1/10 = 0.1 正規化單位,
// 配合基礎掃描 2 parsec + 母星星基加成 2 parsec,開局偵測半徑 = 4*0.1=0.4,恰好把星 1(0.25)、
// 星 5(0.28)兩顆鄰近星納入可見範圍、其餘十餘顆(0.4 以上)入霧——「母星區可見數顆星、遠星
// 入霧」,不是全圖可見(fog 沒意義)也不是幾乎全霧(看起來像壞掉)。若日後盤面星數/密度調整
// 覺得這圈太大/太小,改這個常數即可,不必動偵測邏輯本身。
const ParsecToNormalized = 1.0 / 10.0

// 掃描科技偵測範圍(parsec,【近似】遞增序):手冊只講「越高階掃描看越遠」的定性順序
// (Space Scanner < Neutron Scanner < Tachyon Scanner),無法查到具體 parsec 值,故本專案採
// 「基礎 2、每升一階 +2」的簡單遞增近似,不是手冊數字。
const (
	scannerRangeBase    = 2 // 無任何掃描科技(開局預設)
	scannerRangeSpace   = 4 // TECH_SPACE_SCANNER
	scannerRangeNeutron = 6 // TECH_NEUTRON_SCANNER
	scannerRangeTachyon = 8 // TECH_TACHYON_SCANNER
)

// ScannerRangeParsec 依已解鎖的掃描科技,回傳偵測範圍(parsec)。取已解鎖科技中最高階者;
// 三項都未解鎖則回傳基礎值。呼叫端(internal/shell)負責用既有「元件/科技解鎖」判定規則
// (componentUnlockedFor 同款模式)算出 hasSpace/hasNeutron/hasTachyon,本函式只管查表,
// 不碰任何 CompletedTopics/ExplicitChoice 細節。
func ScannerRangeParsec(hasSpace, hasNeutron, hasTachyon bool) int {
	switch {
	case hasTachyon:
		return scannerRangeTachyon
	case hasNeutron:
		return scannerRangeNeutron
	case hasSpace:
		return scannerRangeSpace
	default:
		return scannerRangeBase
	}
}

// 軌道基地掃描加成(parsec,【近似】):手冊定性描述軌道基地(星基/戰鬥站/星辰要塞)會增加
// 殖民地的偵測範圍,但同樣無公開 parsec 數字。本專案沿用 orbital_bombardment.go
// retaliationAttackers 已用過的「星辰要塞 > 戰鬥站 > 星基,擇一取代不疊加」慣例(手冊:星辰
// 要塞取代同軌道的戰鬥站/星基,不共存),數值按軌道基地量級遞增類推。
const (
	orbitalScannerBonusStarFortress  = 6 // 星辰要塞
	orbitalScannerBonusBattlestation = 4 // 戰鬥站
	orbitalScannerBonusStarBase      = 2 // 星基
)

// OrbitalScannerBonusParsec 依殖民地已完工建築(buildings,鍵為中文建築名,比照
// shell.GameSession.ColonyBuildings/AIOpponent.ColonyBuildings 的資料形狀),回傳該殖民地
// 額外的偵測範圍加成(parsec)。三種軌道基地擇一取最高階,不疊加;都沒有則回傳 0。
func OrbitalScannerBonusParsec(buildings map[string]bool) int {
	switch {
	case buildings["星辰要塞"]:
		return orbitalScannerBonusStarFortress
	case buildings["戰鬥站"]:
		return orbitalScannerBonusBattlestation
	case buildings["星基"]:
		return orbitalScannerBonusStarBase
	default:
		return 0
	}
}

// DetectionRangeNormalized 把「掃描科技 parsec + 軌道基地加成 parsec + 版本規則加成
// parsec」加總後換算成正規化星圖座標下的偵測半徑,供 internal/shell 拿去跟
// math.Hypot(距離) 比較,判定某顆星是否落在偵測範圍內。
func DetectionRangeNormalized(scannerParsec, orbitalParsec, versionBonusParsec int) float64 {
	total := scannerParsec + orbitalParsec + versionBonusParsec
	return float64(total) * ParsecToNormalized
}
