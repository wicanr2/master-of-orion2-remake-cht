// Package gamedata:偵測/掃描範圍模型(diff 全量表 #13——掃描/偵測距離)。
//
// ⚠ 2026-08-08(第 71 項(探針③內部函式))訂正:這裡原本寫著「手冊只定性描述『掃描科技越高階看得越遠』,
// 並未公開逐科技的 parsec 數字」——**那句話是錯的,不是過期的**。手冊三個掃描科技的條目
// 各自明寫了自己的 parsec 值(空間 2 / 迅子 4 / 中子 6),就在同一份 PDF 裡。
// 而且順序也寫反了:原本的近似值把迅子當成最高階(8),手冊說**中子(6)才比迅子(4)遠**。
// 「查不到所以標近似」是誠實的做法,前提是**真的查過**。
//
// 仍屬近似的部分(手冊確實沒給):軌道基地(星基/戰鬥站/星辰要塞)的加成,見下方常數註解。
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

// 掃描科技偵測範圍(parsec):**手冊逐字**,三個條目各自寫在自己那一段裡。
//
//   - Space Scanner:「The standard scanner can detect a Frigate class ship at a range of
//     **2 parsecs**.」
//   - Tachyon Scanner:「These scanners have a base detection range (for Frigates) of
//     **4 parsecs**.」
//   - Neutron Scanner:「These scanners have a base detection range (for Frigates) of
//     **6 parsecs**.」
//
// ⚠ 空間掃描儀 = 2 = 基礎值,所以它對範圍其實**沒有加成**——手冊叫它「the standard
// scanner」,它就是基準本身。這看起來像抄錯,不是:TOPIC_PHYSICS 是 ResearchAll 主題
// (開局幾乎立刻拿到),把它當基準與手冊的敘述一致。
//
// ⚠ 手冊同一段還給了「被偵測方的艦體越大越早被看到」的修正(驅逐艦 +1、巡洋艦 +2、
// 戰艦 +3、泰坦 +4、末日之星 +5)。**本 remake 不實作那一條**,因為它修正的是「看不看得到
// 某支艦隊」,而本專案的偵測只決定「看不看得到某顆星」(AI 艦隊是抽象戰力,沒有地圖座標,
// 見 shell/detection.go 檔頭)。缺的是前置系統,不是這個數字。
const (
	scannerRangeBase    = 2 // 無任何掃描科技(開局預設,與 standard scanner 同值)
	scannerRangeSpace   = 2 // TECH_SPACE_SCANNER —— 手冊「2 parsecs」
	scannerRangeTachyon = 4 // TECH_TACHYON_SCANNER —— 手冊「4 parsecs」
	scannerRangeNeutron = 6 // TECH_NEUTRON_SCANNER —— 手冊「6 parsecs」
)

// 掃描科技對敵方飛彈閃避的抵銷(點):**手冊逐字**,寫在迅子/中子兩個條目的最後一句。
//
//   - Tachyon Scanner:「…also reduce the effectiveness of enemy missile jamming systems,
//     lowering the target's Missile Evasion by **20 points**.」
//   - Neutron Scanner:同句,「by **40 points**」。
//   - Space Scanner:**沒有這一句**,故為 0(不是漏抄)。
//
// 手冊 p.123 的干擾機率範例也用了這組數字:「攻擊方 Tachyon Scanner 已知加成 20」
// → P = [(87)−20]/2 = 33%(見 MissileJamChance 的註解)。兩處互相印證。
const (
	scannerJamReductionSpace   = 0
	scannerJamReductionTachyon = 20
	scannerJamReductionNeutron = 40
)

// ShipBattleScannerScanParsecBonus 戰鬥掃描器**在戰鬥之外**的第二個效果。
//
// 手冊(Battle Scanner)整段兩句話:「The scanner increases the ship's chance to hit with
// beam weapons by 50. Furthermore, ships equipped with Battle Scanners have a scanning range
// **2 parsecs greater** when in normal or hyperspace (**outside of combat**).」
//
// ⚠ 第 68 項(元件盤點+飛彈防禦)接了第一句,漏了第二句——而且還寫了一條測試(TestBattleScannerRaisesOnlyAccuracy)
// 把「只加命中」釘住。**測試釘住的是我當時的理解,不是手冊。** 這與第 68 項(元件盤點+飛彈防禦)慣性穩定器
// (同樣是一個元件兩個效果、只接了一半)是同一個坑,連續兩項都踩。
const ShipBattleScannerScanParsecBonus = 2

// ScannerRangeParsec 依已解鎖的掃描科技,回傳偵測範圍(parsec)。
//
// 取已解鎖科技中**範圍最大**者(不是「最高階」——手冊的中子比迅子遠,先前那版按科技樹階序
// 挑,結果研究出迅子反而覆蓋掉更遠的中子)。三項都未解鎖則回傳基礎值。呼叫端
// (internal/shell)負責用既有「元件/科技解鎖」判定規則算出 hasSpace/hasNeutron/hasTachyon。
func ScannerRangeParsec(hasSpace, hasNeutron, hasTachyon bool) int {
	best := scannerRangeBase
	for _, c := range []struct {
		owned bool
		val   int
	}{{hasSpace, scannerRangeSpace}, {hasTachyon, scannerRangeTachyon}, {hasNeutron, scannerRangeNeutron}} {
		if c.owned && c.val > best {
			best = c.val
		}
	}
	return best
}

// ScannerMissileEvasionReduction 依已解鎖的掃描科技,回傳「攻方掃描器抵銷目標飛彈閃避」的點數
// (MissileJamChance 的 attackerScannerBonus)。同樣取**抵銷最多**者,理由同上。
func ScannerMissileEvasionReduction(hasSpace, hasNeutron, hasTachyon bool) int {
	best := 0
	for _, c := range []struct {
		owned bool
		val   int
	}{{hasSpace, scannerJamReductionSpace}, {hasTachyon, scannerJamReductionTachyon},
		{hasNeutron, scannerJamReductionNeutron}} {
		if c.owned && c.val > best {
			best = c.val
		}
	}
	return best
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
