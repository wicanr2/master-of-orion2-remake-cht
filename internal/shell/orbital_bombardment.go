package shell

import (
	"math/rand"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// orbital_bombardment.go:軌道轟炸(Orbital Bombardment)引擎層最小接線,對應
// moo2_patch1.5/MANUAL_150.html p.129「Notes on Orbital Assault > Orbital Bombardment」。
// 與 ground_invasion.go 的地面入侵是兩個獨立動作(手冊:轟炸削弱/殺人口,不代表佔領;佔領
// 仍要靠 InvadeColony 的陸戰隊/戰車營入侵)。本檔只碰資料/流程,不碰 UI——BombardColony 是
// 供未來 UI 呼叫的引擎函式,目前 cmd/moo2/interactive.go 尚未接對應按鈕(TODO,見下方
// BombardColony 註解與 docs/tech 更新)。

// fleetBombardDamage 依手冊「Estimated Bomb Hits」段落模擬玩家艦隊對殖民地做「10 輪齊射」,
// 回傳累積造成的總傷害(尚未除以 100、尚未套用 320 上限,由呼叫端接
// gamedata.GroundBombHitsFromDamage)。
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
	for round := 0; round < 10; round++ {
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
	TotalDamage int    // 10 輪齊射總傷害(套用前見 fleetBombardDamage 註解的已知簡化)
	Hits        int    // gamedata.GroundBombHitsFromDamage(TotalDamage),Estimated Bomb Hits

	PopulationLost int // 實際扣減的殖民地人口單位數(1 hit = 1 人口單位,Planet Hits 表)
	RemainingHits  int // 扣完人口後剩餘、未消耗掉的 hits(通常應為 0;殖民地人口歸零時會 > 0)

	// PlanetHitsRequired 是手冊「Planet Hits」表算出的「摧毀這個殖民地全部防禦所需 hits」
	// 估計值(gamedata.GroundPlanetTotalHits),對應手冊 UI 上「Estimated Bomb Hits」旁邊
	// 同時顯示的「Planet Hits」欄——純供顯示參考(讓玩家判斷這波轟炸夠不夠),不影響
	// PopulationLost 的實際扣減(扣減直接用 Hits,見下方函式註解)。
	//
	// TODO 誠實限制:AI 側完全沒有建築/儲存生產追蹤(AIOpponent 無 ColonyBuildings,見
	// ground_invasion.go InvadeColony 註解),故本欄位的 buildings/storedProduction 兩項
	// 恆為 0/false,不代表真的沒有建築,只代表「目前資料模型量不到」——不臆測填數字。
	// 戰車營同理:AI 開局 homeworldBuildings() 沒有裝甲營房,也無法追蹤是否後續建成,tanks
	// 恆為 0。
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
// 解算:fleetBombardDamage 模擬 10 輪齊射 → gamedata.GroundBombHitsFromDamage 換算 hits →
// 依手冊 Planet Hits 表「每整數人口 1 hit」直接扣減 colony.Population(夾在 0 以上)。
//
// ⚠ 範圍限制(誠實標註,非本函式應臆測補齊的部分):
//   - 只扣人口,不扣建築/儲存生產/駐軍——AI 沒有這些的持久資料可扣(見 GroundBombardResult
//     欄位註解),扣了會是憑空生資料,故不做。
//   - 手冊未講「殖民地人口被轟炸到 0」時的後續(是否直接摧毀殖民地/移除星系 Owner):不在本
//     函式臆測補上,留給未來確認手冊或 openorion2 行為後再接(TODO)。目前行為是 Population
//     可以停在 0,殖民地本身仍存在於 aiPlayer.Colonies(不會被移除)。
//   - 轟炸不會使殖民地被佔領(手冊:入侵才佔領),故本函式不改動 Star.Owner。
//
// rng 依「回合數 + 星索引」種子化(與 InvadeColony 同款慣例,但另加不同的乘數避免與入侵用的
// rng 種子巧合撞在一起),同一回合對同一顆星重複呼叫必得到相同結果。
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

	popLoss := hits
	if popLoss > colony.Population {
		popLoss = colony.Population
	}
	if popLoss < 0 {
		popLoss = 0
	}
	colony.Population -= popLoss
	res.PopulationLost = popLoss
	res.RemainingHits = hits - popLoss

	// PlanetHitsRequired 純供顯示參考,見 GroundBombardResult 欄位註解的 TODO 範圍限制。
	defMarines := gamedata.GroundMarineBarracksUnits(s.Turn, colony.Population, colony.PopMax, false)
	defMarineHits := gamedata.GroundMarineHitsToKill(false, hasPoweredArmorFor(aiPlayer.Player))
	res.PlanetHitsRequired = gamedata.GroundPlanetTotalHits(0, false, colony.Population, 0, defMarines, defMarineHits, 0, 0)

	return res
}
