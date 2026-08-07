package shell

import "github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"

// orbit.go:**一星多行星**的資料模型。
//
// remake 先前是 `Stars[i]` ↔ `Planets[i]` **一對一**——一顆星一顆行星。
// MOO2 不是這樣:**每個星系有 5 個軌道**,每個軌道可能有行星、也可能是空的。
//
// ============ 5 個軌道:三個獨立來源 ============
//
//	① 偏移算術:星球結構裡軌道陣列在 +0x4A,而下一個已知欄位(遷移目標,見 relocation.go)
//	   在 +0x54 —— 中間 10 個位元組 = **5 個 word**。
//	② `System_Planet_Scanned_To_Planet_Id_` @ 0x78CDB:
//	     word[星×0x71 + 0x4A + 軌道×2]        ; 軌道 → 行星 id(−1 = 空)
//	③ 走訪那個陣列的迴圈,上界是寫死的:`cmp word ptr [var_4], 5; jge`(0x1CB31)。
//
// 行星本身是**獨立的一張表**(`dword_1930D4`,每筆 0x11 = 17 位元組),
// `Planet_Orbit_` @ 0x783ED 讀 `byte[行星 id×0x11 + 3]` = 它在第幾號軌道。
// 也就是**雙向都有指標**:星 → 軌道 → 行星 id,以及行星 → 軌道號。
//
// ============ 這一層卡著什麼 ============
//
//   - **人造行星**(建築 48):要在同一個星系裡找得到氣態巨星或小行星帶當材料。
//     ⚠ 它**不需要空軌道**——那是 remake 先前推錯的假設,見 artificialplanet.go 檔頭的訂正。
//   - **System 視窗**:原版列的是整個星系的行星,remake 只有一顆。
//   - **同星系多殖民地**。
//
// ============ ⚠ 這一版是「行為不變」的第一階段 ============
//
// 產生器目前仍然每顆星只生一顆行星,放在**軌道 0**,其餘 4 個軌道空著。
// 所以畫面與數值**逐位元不變**——這一階段換的是形狀不是內容。
// 真正生出多顆行星要等骰表接上(原版 `_orbit_to_satellite_type` 那條,見 gap report C-5)。

// StarOrbits 是每個星系的軌道數(原版真值,見檔頭三個來源)。
const StarOrbits = 5

// OrbitEmpty 是「這個軌道沒有行星」(原版的 −1)。
const OrbitEmpty = -1

// PlanetAt 回傳某顆星的**代表行星**索引(沒有回 −1)。
//
// 挑法與 `genPlanets` 原本的規則**逐字相同**:依軌道順序找第一顆一般行星(可殖民);
// 整組都不宜居時才退而取第一個天體。
//
// **這是相容性支點**:一星一行星時它必須等於舊的 `Planets[star]`,
// 否則所有舊呼叫端換過來都會位移。多行星之後它是「主行星」——
// **新程式碼要處理整個星系時請用 `PlanetsAt`**,不要靠這一支。
func (s *GameSession) PlanetAt(star int) int {
	return representativePlanet(s.Stars, s.Planets, star)
}

// representativePlanet 是 `PlanetAt` 的自由函式版本,供**還沒有 GameSession** 的
// 生成階段用(genMonsters 要在星系剛生出來時挑代表行星)。
//
// 兩邊共用同一支,是為了不讓「代表行星怎麼挑」出現第二份實作——
// 那兩份一旦漂開,徵狀是資料錯位而不是崩潰(見第 62 項)。
func representativePlanet(stars []Star, planets []Planet, star int) int {
	if star < 0 || star >= len(stars) {
		return -1
	}
	first := -1
	for _, p := range stars[star].Orbits {
		if p == OrbitEmpty || p < 0 || p >= len(planets) {
			continue
		}
		if first < 0 {
			first = p
		}
		if planets[p].TypeID == gamedata.HABITABLE {
			return p
		}
	}
	return first
}

// PlanetOf 回傳某顆星的代表行星(**可寫指標**);沒有回 nil。
//
// 會改行星資料的呼叫端(隨機事件改礦產/氣候、拓殖消耗特殊物產、抵達時的一次性發現)用這一支。
func (s *GameSession) PlanetOf(star int) *Planet {
	i := s.PlanetAt(star)
	if i < 0 || i >= len(s.Planets) {
		return nil
	}
	return &s.Planets[i]
}

// PlanetDataAt 回傳某顆星代表行星的**複本**(沒有回零值 + false)。
func (s *GameSession) PlanetDataAt(star int) (Planet, bool) {
	if p := s.PlanetOf(star); p != nil {
		return *p, true
	}
	return Planet{}, false
}

// PlanetsAt 回傳某顆星所有軌道上的行星索引(依軌道順序,略過空軌道)。
func (s *GameSession) PlanetsAt(star int) []int {
	var out []int
	for _, p := range s.OrbitsOf(star) {
		if p != OrbitEmpty {
			out = append(out, p)
		}
	}
	return out
}

// OrbitsOf 回傳某顆星的軌道表(越界回全空)。
//
// ⚠ 舊存檔沒有 Orbits,解出來是 5 個零值 **0** —— 那會讓每顆星都宣稱軌道 0 上有行星 0
// (同蟲洞欄位那個坑,見 Star.Wormhole 註解)。讀檔一律走 `normalizeOrbits`。
func (s *GameSession) OrbitsOf(star int) [StarOrbits]int {
	if star < 0 || star >= len(s.Stars) {
		return emptyOrbits()
	}
	return s.Stars[star].Orbits
}

// emptyOrbits 回傳「五個軌道都是空的」。
func emptyOrbits() [StarOrbits]int {
	var o [StarOrbits]int
	for i := range o {
		o[i] = OrbitEmpty
	}
	return o
}

// PlanetStar 回傳某顆行星屬於哪一顆星(找不到回 −1)。
//
// 原版是行星結構自己記著(`Planet_Orbit_` 那一族),remake 目前用反查——
// 星數與行星數都是幾十,反查的成本可以忽略,而少一份要維護的雙向指標就少一種對不上的方式。
func (s *GameSession) PlanetStar(planet int) int {
	for i := range s.Stars {
		for _, p := range s.Stars[i].Orbits {
			if p == planet {
				return i
			}
		}
	}
	return -1
}

// PlanetOrbit 回傳某顆行星在第幾號軌道(找不到回 −1)。對應原版 `Planet_Orbit_` @ 0x783ED。
func (s *GameSession) PlanetOrbit(planet int) int {
	for i := range s.Stars {
		for o, p := range s.Stars[i].Orbits {
			if p == planet {
				return o
			}
		}
	}
	return -1
}

// FreeOrbit 回傳某顆星第一個空軌道的編號(沒有空軌道回 −1)。
//
// ⚠ **人造行星不用這個**——它改造的是既有天體,不新增軌道(見 artificialplanet.go 的訂正)。
// 留著給「真的要新增一顆世界」的場合(目前沒有呼叫端)。
func (s *GameSession) FreeOrbit(star int) int {
	if star < 0 || star >= len(s.Stars) {
		return -1
	}
	for o, p := range s.Stars[star].Orbits {
		if p == OrbitEmpty {
			return o
		}
	}
	return -1
}

// normalizeOrbits 修掉「舊存檔沒有 Orbits」的零值陷阱,並重建缺漏的軌道表。
//
// 判準:整張表都是 0(Go 零值)且行星數與星數相同 → 那是**一星一行星時代的存檔**,
// 重建成「每顆星的軌道 0 放同索引的行星」。這與該版本的實際狀態逐位元一致。
func normalizeOrbits(stars []Star, nPlanets int) {
	for i := range stars {
		allZero := true
		for _, p := range stars[i].Orbits {
			if p != 0 {
				allZero = false
				break
			}
		}
		if !allZero {
			continue // 已經是新格式(至少有一個非零,含 −1)
		}
		stars[i].Orbits = emptyOrbits()
		if i < nPlanets {
			stars[i].Orbits[0] = i // 舊格式:Planets 與 Stars 平行
		}
	}
}
