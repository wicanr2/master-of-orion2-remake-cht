package shell

import (
	"math/rand"
	"sort"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// orbital_bombardment.go:軌道轟炸(Orbital Bombardment)引擎層最小接線,對應
// moo2_patch1.5/MANUAL_150.html p.129「Notes on Orbital Assault > Orbital Bombardment」。
// 與 ground_invasion.go 的地面入侵是兩個獨立動作(手冊:轟炸削弱/殺人口,不代表佔領;佔領
// 仍要靠 InvadeColony 的陸戰隊/戰車營入侵)。本檔只碰資料/流程,不碰 UI——BombardColony 已由
// cmd/moo2/interactive.go 的 galaxy() 星系主畫面接上「軌道轟炸」按鈕(2026-07-11,敵殖民地星
// 恆可用,與「發動地面入侵」雙鈕共存,分居 y=402/424 兩列)。

// fleetBombardDamage 依手冊「Estimated Bomb Hits」段落模擬玩家艦隊對殖民地做「s.RuleProfile.
// BombardmentVolleys 輪齊射」(1.5 預設 10 輪、1.3 為 5 輪,見 gamedata.RuleProfile 與
// docs/tech/version-1.3-1.5-diff.md §2),回傳累積造成的總傷害(尚未除以 100、尚未套用 320
// 上限,由呼叫端接 gamedata.GroundBombHitsFromDamage)。
//
// 手冊原文:「All remaining ships fire all weapons 10 times, or as many times as there is
// ammo in 10 turns... and total damage is calculated from it.」逐發解算沿用既有戰術戰鬥公式
// (ResolveShot/ResolveMissileShot,同 battleVolley 的分流邏輯),只是目標從「敵艦」換成
// 「殖民地」,且不模擬殖民地反擊(手冊本段只描述攻方輸出,未提及行星火力回擊,回擊屬於另一套
// 「行星飛彈基地/星基防禦」機制,不在本函式範圍)。
//
// 已知簡化(誠實標註,非杜撰真值,是既有 remake 資料模型限制,非本函式引入):
//   - 沒有「行星護盾」資料(damage.go DamageAfterShield 明講「本函式只處理艦對艦,行星護盾
//     情境不適用」),故護盾/裝甲一律視為 0(無防禦)。
//   - 手冊「Damage of beams and torpedoes is halved just like in tactical combat」與
//     「A better computer helps for beams here too」:目前戰術戰鬥層本身都還沒有獨立的「減半」
//     或「電腦命中加成」函式接線(見 ground.go 檔尾 TODO),故本模擬未套用,直接沿用一般
//     ResolveShot 命中/傷害公式——TODO,待戰術戰鬥層先補上這兩項才能真正對齊手冊轟炸公式。
func (s *GameSession) fleetBombardDamage(rng *rand.Rand) int {
	total := 0
	for round := 0; round < s.RuleProfile.BombardmentVolleys; round++ {
		for _, sh := range s.Ships {
			body := shipStrength(sh.Class)
			atk := body + sh.WeaponAttack
			atk += atk * s.RaceCombatPct / 100
			wmin, wmax := atk/2, atk
			var shot ShotResult
			switch weaponKindByName(sh.Weapon) {
			case WeaponKindMissile:
				amrRoll := rng.Intn(100) + 1
				jamRoll := rng.Intn(100) + 1
				shot = ResolveMissileShot(false, 0, amrRoll, 0, 0, false, jamRoll, wmax, 0, 0, false)
			case WeaponKindSpherical:
				span := wmax - wmin
				r := 0
				if span > 0 {
					r = rng.Intn(span + 1)
				}
				aggD := gamedata.DamageSphericalRoll(wmin, r, 100)
				shot = ResolveSphericalShot(aggD, 0, 0, false, false)
			default:
				roll := rng.Intn(100) + 1
				shot = ResolveShot(atk, wmin, wmax, 2, 0, 0, roll, false, false)
			}
			if shot.Hit {
				total += shot.DamageToStructure
			}
		}
	}
	return total
}

// GroundBombardResult 一次軌道轟炸嘗試的結果(供 UI/測試檢視)。
type GroundBombardResult struct {
	Ok          bool   // 是否成功發動了一場轟炸解算(false = 前置條件不足,未開打)
	Reason      string // Ok=false 時的原因
	TotalDamage int    // RuleProfile.BombardmentVolleys 輪齊射總傷害(套用前見 fleetBombardDamage 註解的已知簡化)
	Hits        int    // gamedata.GroundBombHitsFromDamage(TotalDamage),Estimated Bomb Hits

	// PopulationLost 實際扣減的殖民地人口單位數(1 hit = 1 人口單位,Planet Hits 表)。
	// 2026-07-11 起:這裡的 hits 是「建築吸收後的餘數」,不是 Hits 原始值——見下方函式註解
	// 「建築吸收」段落與 BuildingsDestroyed 欄位。
	PopulationLost int
	RemainingHits  int // 扣完建築+人口後剩餘、未消耗掉的 hits(通常應為 0;殖民地人口歸零時會 > 0)

	// BuildingsDestroyed 是本次轟炸摧毀的建築數(2026-07-11 新增,#7/#8 接線:見下方函式註解
	// 「建築吸收」段落)。0 表示沒有建築被摧毀(可能是 hits 不夠、也可能是該殖民地本來就沒有
	// 建築——兩種情況本欄位都合法回 0,不額外區分)。
	BuildingsDestroyed int
	// BuildingsRemaining 是轟炸結束後該殖民地剩餘的建築數(len(ColonyBuildings[colonyIdx])),
	// 供 UI/測試檢視「還剩多少建築沒被炸掉」。
	BuildingsRemaining int

	// PlanetHitsRequired 是手冊「Planet Hits」表算出的「摧毀這個殖民地全部防禦所需 hits」
	// 估計值(gamedata.GroundPlanetTotalHits),對應手冊 UI 上「Estimated Bomb Hits」旁邊
	// 同時顯示的「Planet Hits」欄——純供顯示參考(讓玩家判斷這波轟炸夠不夠),不影響
	// PopulationLost 的實際扣減(扣減邏輯見下方函式註解)。
	//
	// TODO 誠實限制:AI 側仍無「儲存生產」追蹤(storedProduction 恆 false)。建築數本身
	// (2026-07-11 起 AIOpponent.ColonyBuildings 已備妥)已改用轟炸結束後的實際剩餘建築數,
	// 不再恆 0——見下方函式賦值處。戰車營同理:AI 開局 homeworldBuildings() 沒有裝甲營房,也
	// 無法追蹤是否後續建成,tanks 恆為 0。
	PlanetHitsRequired int
}

// BombardColony 嘗試對 starIdx 這顆星發動一次軌道轟炸(手冊 p.129 Orbital Bombardment)。
// 前置條件與 InvadeColony 對稱,唯獨不需要已載運陸戰隊/戰車營(轟炸是艦隊武器對殖民地開火,
// 不需要地面部隊登陸):
//  1. 玩家艦隊已抵達該星(FleetAtStar==starIdx 且 FleetETA==0)。
//  2. 該星是敵方(Owner==2)且有「已建模」的殖民地(findAIColonyByStar 找得到)。
//  3. 玩家艦隊至少有 1 艘艦(len(s.Ships)>0,否則無武器可轟炸)。
//
// 任一條件不足回傳 Ok=false + Reason,不消耗任何狀態、不呼叫 rng。
//
// 解算:fleetBombardDamage 模擬 RuleProfile.BombardmentVolleys 輪齊射 →
// gamedata.GroundBombHitsFromDamage 換算 hits → hits 先花在摧毀殖民地建築(見下方「建築吸收」
// 段落)→ 建築吸收後的餘數 hits 才交給 gamedata.GroundBombardPopulationLoss 依行星尺寸係數
// (#11,2026-07-11,非差異項)換算實際扣減的人口單位數(大行星較耐轟,近似公式見該函式註解)→
// 扣減 colony.Population(夾在 0 以上)。
//
// 建築吸收(#7/#8 接線,2026-07-11):對應手冊 p.78「飛彈基地只能被軌道轟炸摧毀」+ #8
// civilian building 轟炸裝甲 100hp——忠實模型是「bomb hits 先摧毀殖民地建築,餘數才扣人口」,
// 軌道防禦讓殖民地人口在防禦被轟掉前受保護:
//   - 每棟建築消耗的 hits = gamedata.GroundPlanetHitsPerBuilding + s.RuleProfile.
//     BombardmentBuildingBonusHits(1.3 每棟建築多 +1 hit 才摧毀,CHANGELOG「Undocumented +1
//     hit bonus for civilian buildings during bombardment removed」;1.5 已移除此 bug,見
//     ruleprofile.go 欄位註解)。
//     ⚠ 誠實標註:CHANGELOG 這句話本身語意模糊(是建築多吸一擊、還是建築多受一擊才被摧毀),
//     本 remake 採「每棟建築在 1.3 需多 +1 hit 才摧毀」的保守解讀,非手冊逐字驗證值。
//   - 分配順序:依建築名稱字母序(sort.Strings)固定摧毀,不用 rng——同一批建築、同一 hits
//     輸入,摧毀結果永遠一樣,可重現。hits 不夠摧毀下一棟就停止(不會摧毀「一半」的建築)。
//   - 被摧毀的建築從 aiPlayer.ColonyBuildings[colonyIdx] 刪除(map 是參考型別,直接
//     mutate,不需要寫回)。
//   - colony 無建築(nil/空 map)時,這段等同 no-op,全部 hits 直接進人口損傷——與加這個機制
//     之前的行為逐位元一致(見 orbital_bombardment_test.go 既有測試)。
//
// ⚠ 範圍限制(誠實標註,非本函式應臆測補齊的部分):
//   - 不扣「儲存生產」/駐軍——AI 沒有這些的持久資料可扣,扣了會是憑空生資料,故不做(建築已
//     於本輪補上,見上方「建築吸收」)。
//   - 本輪不做「防禦方反擊摧毀玩家艦艇」(軌道基地對轟炸艦隊開火)——那是下一輪工作,本函式
//     不改動 s.Ships。
//   - 手冊未講「殖民地人口被轟炸到 0」時的後續(是否直接摧毀殖民地/移除星系 Owner):不在本
//     函式臆測補上,留給未來確認手冊或 openorion2 行為後再接(TODO)。目前行為是 Population
//     可以停在 0,殖民地本身仍存在於 aiPlayer.Colonies(不會被移除)。
//   - 轟炸不會使殖民地被佔領(手冊:入侵才佔領),故本函式不改動 Star.Owner。
//
// rng 依「回合數 + 星索引」種子化(與 InvadeColony 同款慣例,但另加不同的乘數避免與入侵用的
// rng 種子巧合撞在一起),同一回合對同一顆星重複呼叫必得到相同結果(建築摧毀順序本身不吃 rng,
// 見上方「分配順序」)。
func (s *GameSession) BombardColony(starIdx int) GroundBombardResult {
	if starIdx < 0 || starIdx >= len(s.Stars) {
		return GroundBombardResult{Reason: "無效的星索引"}
	}
	if s.FleetAtStar != starIdx || s.FleetETA != 0 {
		return GroundBombardResult{Reason: "艦隊尚未抵達該星"}
	}
	star := &s.Stars[starIdx]
	if star.Owner != 2 {
		return GroundBombardResult{Reason: "該星不是敵方殖民地"}
	}
	if len(s.Ships) == 0 {
		return GroundBombardResult{Reason: "艦隊沒有可轟炸的艦艇"}
	}
	aiIdx, colonyIdx, ok := s.findAIColonyByStar(starIdx)
	if !ok {
		return GroundBombardResult{Reason: "該星無可轟炸的殖民地模型(簡化限制,見 AIOpponent.ColonyStars)"}
	}
	aiPlayer := &s.AIPlayers[aiIdx]
	colony := &aiPlayer.Colonies[colonyIdx]

	rng := rand.New(rand.NewSource(int64(s.Turn)*2654435761 + int64(starIdx)*131 + 777))
	totalDamage := s.fleetBombardDamage(rng)
	hits := gamedata.GroundBombHitsFromDamage(totalDamage)

	res := GroundBombardResult{Ok: true, TotalDamage: totalDamage, Hits: hits}

	// --- 建築吸收(#7/#8 接線,見函式檔頭「建築吸收」段落):hits 先花在摧毀建築 ---
	remainingHits := hits
	var buildings map[string]bool
	if colonyIdx < len(aiPlayer.ColonyBuildings) {
		buildings = aiPlayer.ColonyBuildings[colonyIdx]
	}
	if len(buildings) > 0 {
		hitsPerBuilding := gamedata.GroundPlanetHitsPerBuilding + s.RuleProfile.BombardmentBuildingBonusHits
		if hitsPerBuilding < 1 {
			hitsPerBuilding = 1 // 防禦性下限,避免 RuleProfile 誤設非正值造成除零/無限摧毀
		}
		// 固定優先序(依建築名稱字母序),不吃 rng,同一批建築+同一 hits 輸入摧毀結果必重現。
		names := make([]string, 0, len(buildings))
		for name := range buildings {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			if remainingHits < hitsPerBuilding {
				break
			}
			delete(buildings, name)
			remainingHits -= hitsPerBuilding
			res.BuildingsDestroyed++
		}
	}
	res.BuildingsRemaining = len(buildings)

	// 行星尺寸幾何(#11,2026-07-11,docs/tech/version-1.3-1.5-diff.md,非差異項——1.5 系列
	// 中途曾改過又於 1.50.11 修回 classic 3-4-6-7-8,對 1.3 vs 最終 1.5 不構成差異):大行星
	// 較耐轟,近似公式與已知限制見 gamedata.GroundBombardPopulationLoss 註解。colony 無建築時
	// remainingHits==hits,母星預設 LARGE_PLANET(係數 7)時 hits=10 算出 popLoss=8,與加建築
	// 吸收機制前的舊行為(popLoss=hits 直接相等)剛好一致(見 orbital_bombardment_test.go
	// TestBombardColony_ nil 建築回歸測試)。
	popLoss := gamedata.GroundBombardPopulationLoss(remainingHits, colony.PlanetSize)
	if popLoss > colony.Population {
		popLoss = colony.Population
	}
	if popLoss < 0 {
		popLoss = 0
	}
	colony.Population -= popLoss
	res.PopulationLost = popLoss
	res.RemainingHits = remainingHits - popLoss

	// PlanetHitsRequired 純供顯示參考(見 GroundBombardResult 欄位註解):buildings 參數改用
	// len(buildings)(本次轟炸「結束後」剩餘的建築數),與下面 defMarines 用「轟炸後」的
	// colony.Population 同一種語意(顯示「打完這波,對方還剩多少防禦要打」)——AI 建築資料
	// 2026-07-11 起已備妥,不再恆為 0;storedProduction 仍恆 false,見欄位註解的剩餘 TODO。
	defMarines := gamedata.GroundMarineBarracksUnits(s.Turn, colony.Population, colony.PopMax, false)
	defMarineHits := gamedata.GroundMarineHitsToKill(false, hasPoweredArmorFor(aiPlayer.Player))
	res.PlanetHitsRequired = gamedata.GroundPlanetTotalHits(len(buildings), false, colony.Population, 0, defMarines, defMarineHits, 0, 0)

	return res
}
