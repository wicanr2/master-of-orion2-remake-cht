package shell

// route.go:星圖航線(艦隊從 A 到 B 沿路會碰到什麼)。
//
// 第 16 項(秒差距模型)把距離與航速換成秒差距之後,還有三條手冊規則卡在同一個前置:
// **「艦隊沿路經過哪些東西」**。先前的星圖是「兩點直接算 ETA」,根本沒有「沿路」這個概念。
//
//	黑洞    「No ship can safely pass within 2 parsecs of a black hole
//	         (unless the ship contains an officer with the Navigator skill).」
//	干擾場  Warp Field Interdictor:「radius of 3 full parsecs … slows all enemy ships
//	         approaching the system to a speed of 1 parsec per turn.」
//	星雲    「Ships traveling **through** a nebula are reduced in speed to 1 parsec per turn.」
//	         —— 重點在 through:兩端都在雲外但直線穿過去,一樣要降速。
//
// 三條都是「這條線段離某個東西多近」或「這條線段有沒有穿過某塊區域」,所以一個線段模型全解。
//
// ============ ⚠ 直線,不是原版的逐格路徑 ============
//
// 原版的艦隊在銀河座標上每回合前進 N 秒差距,實際軌跡由目的地決定;這裡用**起訖點的直線**
// 當航線。對「離黑洞多近」「有沒有穿過星雲」這種判定,直線與逐格前進的結果幾乎相同
// (原版也是朝目的地直走)。真正的差別是原版會**中途停在某一格**,而 remake 的艦隊在
// 抵達前沒有中間位置——那是 ETA 模型的限制,不是這一檔的。

import (
	"math"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// SetNebulaProbe 裝上「某個正規化座標點是否落在星雲內」的判定式,並立刻重算每顆星的旗標。
//
// 判定要讀星雲圖的遮罩,而本套件是純規則層不碰資產,所以由 cmd/moo2 提供
// (見 cmd/moo2/nebula.go)。傳 nil 等於「沒有星雲判定」:旗標全清、沿路判定一律 false。
//
// ⚠ 這個欄位**不進存檔**(未匯出),讀檔後要重新裝一次。
func (s *GameSession) SetNebulaProbe(fn func(x, y float64) bool) {
	s.nebulaProbe = fn
	s.refreshStarNebulaFlags()
}

// refreshStarNebulaFlags 重算每顆星的「在星雲內」。
//
// 手冊(patch 1.5):「Mapgen prevents Black Holes from appearing in Nebulas」——
// remake 採較保守的做法:不改星的光譜,只是不把黑洞標成在星雲內。
func (s *GameSession) refreshStarNebulaFlags() {
	for i := range s.Stars {
		s.Stars[i].InNebula = false
	}
	if s.nebulaProbe == nil {
		return
	}
	for i := range s.Stars {
		if s.Stars[i].Spectral == blackHoleSpectral {
			continue
		}
		s.Stars[i].InNebula = s.nebulaProbe(s.Stars[i].X, s.Stars[i].Y)
	}
}

// starParsecXY 回傳某顆星的秒差距座標(正規化 0..1 依銀河檔位的真實跨距換算)。
func (s *GameSession) starParsecXY(i int) (float64, float64) {
	w, h := gamedata.GalaxyParsecSpan(s.GalaxySizeClass())
	return s.Stars[i].X * w, s.Stars[i].Y * h
}

// pointToSegmentDist 回傳點 (px,py) 到線段 (ax,ay)-(bx,by) 的最短距離。
//
// 用線段而不是直線:目的星之外的延長線上有黑洞不該擋路。
func pointToSegmentDist(ax, ay, bx, by, px, py float64) float64 {
	dx, dy := bx-ax, by-ay
	den := dx*dx + dy*dy
	if den == 0 {
		return math.Hypot(px-ax, py-ay)
	}
	t := ((px-ax)*dx + (py-ay)*dy) / den
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	return math.Hypot(px-(ax+t*dx), py-(ay+t*dy))
}

// routeEndpointsValid 檢查兩個星索引合法且不同。
func (s *GameSession) routeEndpointsValid(from, to int) bool {
	return from >= 0 && from < len(s.Stars) && to >= 0 && to < len(s.Stars) && from != to
}

// RouteBlockedByBlackHole 回傳這條航線是否被黑洞擋住。
//
// 手冊:「No ship can safely pass within 2 parsecs of a black hole (unless the ship contains
// an officer with the Navigator skill).」
//
// ⚠ 起訖點本身是黑洞不算「擋住」——那是玩家自己選的目的地,擋的是**路過**。
func (s *GameSession) RouteBlockedByBlackHole(from, to int) bool {
	if !s.routeEndpointsValid(from, to) || s.FleetHasNavigator() {
		return false
	}
	ax, ay := s.starParsecXY(from)
	bx, by := s.starParsecXY(to)
	for i := range s.Stars {
		if i == from || i == to || s.Stars[i].Spectral != blackHoleSpectral {
			continue
		}
		px, py := s.starParsecXY(i)
		if pointToSegmentDist(ax, ay, bx, by, px, py) < gamedata.BlackHoleAvoidParsecs {
			return true
		}
	}
	return false
}

// RouteInterdicted 回傳這條航線是否經過敵方曲速場干擾器的作用範圍(半徑 3 秒差距)。
//
// 手冊:干擾場「slows all enemy ships approaching the system to a speed of 1 parsec per turn」。
// 這裡的「敵方」= 星的擁有者是 AI(`Owner == 2`)且該殖民地蓋了曲速場干擾器。
func (s *GameSession) RouteInterdicted(from, to int) bool {
	if !s.routeEndpointsValid(from, to) {
		return false
	}
	ax, ay := s.starParsecXY(from)
	bx, by := s.starParsecXY(to)
	for i := range s.Stars {
		if s.Stars[i].Owner != 2 || !s.starHasInterdictor(i) {
			continue
		}
		px, py := s.starParsecXY(i)
		if pointToSegmentDist(ax, ay, bx, by, px, py) <= gamedata.InterdictorRadiusParsecs {
			return true
		}
	}
	return false
}

// starHasInterdictor 回傳某顆星上的 AI 殖民地是否蓋了曲速場干擾器。
func (s *GameSession) starHasInterdictor(starIdx int) bool {
	for _, a := range s.AIPlayers {
		for i, st := range a.ColonyStars {
			if st != starIdx || i >= len(a.ColonyBuildings) {
				continue
			}
			if a.ColonyBuildings[i] != nil && a.ColonyBuildings[i][interdictorBuildingName] {
				return true
			}
		}
	}
	return false
}

// interdictorBuildingName 是曲速力場干擾器的中文建築名。
//
// ⚠ `ColonyBuildings` 是以**中文名**當 key 的 map,所以這裡只能寫死字串。
// `TestInterdictorBuildingNameMatchesTable` 用英文名回查 `gamedata.Buildings` 核對,
// 避免哪天譯名改了這裡靜靜失效(找不到 key 就是「沒有干擾場」,不會報錯)。
const interdictorBuildingName = "曲速力場干擾器"

// routeNebulaSamples 決定沿線取樣點數:每秒差距 4 點,至少 8 點。
//
// 取樣而不是解析求交,是因為星雲的形狀是一張**遮罩圖**不是幾何圖形(見 nebula.go),
// 只能逐點問。4 點/秒差距在最大的銀河(50 秒差距跨距)也就 200 點,而這只在派遣時算一次。
func routeNebulaSamples(parsecs float64) int {
	n := int(parsecs * 4)
	if n < 8 {
		n = 8
	}
	return n
}

// RouteCrossesNebula 回傳這條航線是否**穿過**星雲(不只是兩端在不在雲裡)。
//
// 手冊的字是「Ships traveling **through** a nebula」,所以兩端都在雲外、直線穿過去的情況
// 一樣算。這是第 16 項(秒差距模型)那個近似(只看起訖點)的正解。
func (s *GameSession) RouteCrossesNebula(from, to int) bool {
	if !s.routeEndpointsValid(from, to) {
		return false
	}
	if s.Stars[from].InNebula || s.Stars[to].InNebula {
		return true // 端點在雲裡,不必取樣
	}
	if s.nebulaProbe == nil {
		return false // 沒有遮罩就不憑空判定(headless 模擬即此路徑)
	}
	ax, ay := s.starParsecXY(from)
	bx, by := s.starParsecXY(to)
	n := routeNebulaSamples(math.Hypot(bx-ax, by-ay))
	// 取樣走正規化座標(探針的介面),避免在這裡重複做一次換算。
	nax, nay := s.Stars[from].X, s.Stars[from].Y
	nbx, nby := s.Stars[to].X, s.Stars[to].Y
	for k := 1; k < n; k++ { // 跳過兩端(上面已判過)
		t := float64(k) / float64(n)
		if s.nebulaProbe(nax+(nbx-nax)*t, nay+(nby-nay)*t) {
			return true
		}
	}
	return false
}
