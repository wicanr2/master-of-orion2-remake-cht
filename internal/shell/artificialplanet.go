package shell

// artificialplanet.go:**人造行星**(原版建築表編號 48,一次性 Special 行動)。
//
// ============ 手冊逐字給了完整規格 ============
//
// 「Artificial Planet Construction (Special) — This technology allows a colony in the same
//  system with an asteroid field or gas giant to **assemble this otherwise useless planetary
//  material** into a complete artificial planet that can support a colony. This planet is
//  **Barren, Normal G, and mineral Abundant**. **Gas giants make Huge worlds, and asteroid
//  belts make Large ones.**」
//
// ⚠ **這推翻了 remake 先前的假設。** gap report 第 61 項寫著「人造行星按定義是在既有星系裡
// **再多**一顆世界」,於是把它列為「要有空軌道才蓋得了」。手冊說的不是那樣——
// 它是把**既有的**氣態巨星或小行星帶「assemble」成行星,那顆天體本來就佔著一條軌道。
// 所以前置條件是「同星系有氣態巨星或小行星帶」,不是「有空軌道」。
//
// ============ 反組譯逐項吻合(sub_13FD9 的 loc_143B7 那一段)============
//
// 原版走**兩趟**掃這顆星的 5 條軌道(`word[星 + 0x4A + 軌道×2]`):
//
//	第一趟:找 `planet[+4] == 2`(氣態巨星)→ 結果尺寸 = 4
//	第二趟(第一趟沒找到才跑):找 `planet[+4] == 1`(小行星帶)→ 結果尺寸 = 3
//
// **氣態巨星優先**,而 4 / 3 在尺寸列舉裡正是 **Huge / Large** —— 與手冊那句
// 「Gas giants make Huge worlds, and asteroid belts make Large ones」逐字對上。
// 接著把那顆天體的型別欄寫成 3(一般行星),並改寫整組欄位。
//
// ⚠ 原版另外寫了幾個 remake 沒有對應物的欄位(`byte_17D7FC[尺寸]` / `byte_17D806[尺寸]`
// 那兩張以尺寸索引的表,推測是人口/農場上限)。remake 的這些值本來就由尺寸算出來,
// 不需要另存一份——**這是模型差異,不是漏做**。

import "github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"

// ArtificialPlanetTarget 是人造行星要改造的那顆天體。
type ArtificialPlanetTarget struct {
	Planet int                 // 行星索引
	Size   gamedata.PlanetSize // 改造後的大小(氣態巨星 → Huge、小行星帶 → Large)
}

// FindArtificialPlanetTarget 依原版的順序挑出要改造的天體:**先氣態巨星、再小行星帶**。
//
// 回傳 ok=false 表示這個星系沒有可用的材料。
func (s *GameSession) FindArtificialPlanetTarget(star int) (ArtificialPlanetTarget, bool) {
	// 兩趟,順序就是原版那兩個迴圈的順序——不能合成一趟,那會讓小行星帶在軌道較內時搶先。
	for _, pass := range []struct {
		typ  gamedata.PlanetType
		size gamedata.PlanetSize
	}{
		{gamedata.GAS_GIANT, gamedata.HUGE_PLANET},
		{gamedata.ASTEROIDS, gamedata.LARGE_PLANET},
	} {
		for _, p := range s.PlanetsAt(star) {
			if s.Planets[p].TypeID == pass.typ {
				return ArtificialPlanetTarget{Planet: p, Size: pass.size}, true
			}
		}
	}
	return ArtificialPlanetTarget{}, false
}

// CanBuildArtificialPlanet 回傳「第 i 個殖民地所在的星系能不能蓋人造行星」。
//
// 手冊的前置是「a colony in the same system with an asteroid field or gas giant」——
// 殖民地(呼叫端保證)+ 同星系有材料。
func (s *GameSession) CanBuildArtificialPlanet(colony int) bool {
	star := s.colonyStar(colony)
	if star < 0 {
		return false
	}
	_, ok := s.FindArtificialPlanetTarget(star)
	return ok
}

// BuildArtificialPlanet 把星系裡的氣態巨星 / 小行星帶改造成一顆可殖民的世界。
//
// 回傳被改造的行星索引與是否成功。手冊給的結果是**固定的**:
// Barren 氣候、Normal G 重力、Abundant 礦產;大小依材料而定。
func (s *GameSession) BuildArtificialPlanet(colony int) (int, bool) {
	star := s.colonyStar(colony)
	if star < 0 {
		return -1, false
	}
	t, ok := s.FindArtificialPlanetTarget(star)
	if !ok {
		return -1, false
	}
	p := &s.Planets[t.Planet]
	p.TypeID = gamedata.HABITABLE // 原版把型別欄寫成 3 = 一般行星
	p.SizeID = t.Size
	p.ClimateID = gamedata.BARREN
	p.GravityID = gamedata.NORMAL_G
	p.MineralID = gamedata.ABUNDANT
	p.NoPlanet = false
	// 特殊物產:原版把那一欄清掉(材料被組裝掉了,原本的礦藏不留)。
	p.SpecialID = gamedata.NoSpecial
	p.Climate = climateDisplayName(p.ClimateID)
	p.Gravity = gravityDisplayName(p.GravityID)
	p.Mineral = mineralDisplayName(p.MineralID)
	p.Size = sizeDisplayName(p.SizeID)
	return t.Planet, true
}
