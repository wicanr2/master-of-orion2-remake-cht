package shell

import (
	"math/rand"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// orbital_bombardment.go:軌道轟炸(Orbital Bombardment)引擎層最小接線,對應
// moo2_patch1.5/MANUAL_150.html p.129「Notes on Orbital Assault > Orbital Bombardment」。
// 與 ground_invasion.go 的地面入侵是兩個獨立動作(手冊:轟炸削弱/殺人口,不代表佔領;佔領
// 仍要靠 InvadeColony 的陸戰隊/戰車營入侵)。本檔只碰資料/流程,不碰 UI——BombardColony 已由
// cmd/moo2/interactive.go 的 galaxy() 星系主畫面接上「軌道轟炸」按鈕(2026-07-11,敵殖民地星
// 恆可用,與「發動地面入侵」雙鈕共存,分居 y=402/424 兩列)。

// fleetBombardDamage 依 Strategic_Bombardment_ @ 0x4257E 固定執行三個外層攻擊回合。
// patch 1.31／1.50 的 5／10 只套用炸彈攻擊當量，不再錯誤放大所有武器。
//
// 手冊原文:「All remaining ships fire all weapons 10 times, or as many times as there is
// ammo in 10 turns... and total damage is calculated from it.」逐發解算沿用既有戰術戰鬥公式
// (ResolveShot/ResolveMissileShot,同 battleVolley 的分流邏輯),只是目標從「敵艦」換成
// 「殖民地」,且不模擬殖民地反擊(手冊本段只描述攻方輸出,未提及行星火力回擊,回擊屬於另一套
// 「行星飛彈基地/星基防禦」機制,不在本函式範圍)。
//
// 2026-08-07:**行星護盾接上了**。shieldReduction 是這個殖民地三面護盾裡最強那一面的
// 每次攻擊減傷(gamedata.PlanetaryShieldReduction,手冊 −5 / −10 / −20)。接在**逐發**
// 傷害那一行而不是總傷害——手冊寫的是「per attack」，接在總和後只扣一次會失真。
//
// 已知簡化(誠實標註,非杜撰真值,是既有 remake 資料模型限制,非本函式引入):
//   - 艦對艦意義下的「行星裝甲」仍是 0(damage.go DamageAfterShield 明講不處理行星情境);
//     護盾只有建築給的那個定值,沒有原版可能另有的護盾等級模型。
//   - 手冊「Damage of beams and torpedoes is halved just like in tactical combat」與
//     「A better computer helps for beams here too」:目前戰術戰鬥層本身都還沒有獨立的「減半」
//     或「電腦命中加成」函式接線(見 ground.go 檔尾 TODO),故本模擬未套用,直接沿用一般
//     ResolveShot 命中/傷害公式——TODO,待戰術戰鬥層先補上這兩項才能真正對齊手冊轟炸公式。
func (s *GameSession) fleetBombardDamage(rng *rand.Rand, shieldReduction int) int {
	const (
		strategicBombardmentRounds     = 3
		strategicBombardmentDamageStop = 30000
	)
	total := 0
	for round := 0; round < strategicBombardmentRounds; round++ {
		for _, sh := range s.Fleet().Ships {
			body := shipStrength(sh.Class)
			atk := body + sh.WeaponAttack
			atk += atk * s.RaceCombatPct / 100
			wmin, wmax := atk/2, atk
			kind := weaponKindByName(sh.Weapon)
			shots := 1
			if kind == WeaponKindBomb {
				if round > 0 {
					continue
				}
				shots = s.RuleProfile.BombardmentBombAttacks
				if shots < 0 {
					shots = 0
				}
			}
			for shotNo := 0; shotNo < shots; shotNo++ {
				var shot ShotResult
				switch kind {
				case WeaponKindMissile:
					amrRoll := rng.Intn(100) + 1
					jamRoll := rng.Intn(100) + 1
					shot = ResolveMissileShot(false, 0, amrRoll, 0, 0, false, jamRoll, wmax, 0, 0, false, MissileDefenses{})
				case WeaponKindSpherical:
					span := wmax - wmin
					r := 0
					if span > 0 {
						r = rng.Intn(span + 1)
					}
					aggD := gamedata.DamageSphericalRoll(wmin, r, 100)
					shot = ResolveSphericalShot(aggD, 0, 0, false, false)
				case WeaponKindBomb:
					shot = ShotResult{Hit: true, DamageToStructure: wmax}
				default:
					roll := rng.Intn(100) + 1
					shot = ResolveShot(atk, wmin, wmax, 2, 0, 0, roll, false, false)
				}
				if shot.Hit {
					total += gamedata.PlanetaryShieldedDamage(shot.DamageToStructure, shieldReduction)
					if total > strategicBombardmentDamageStop {
						return total
					}
				}
			}
		}
	}
	return total
}

// GroundBombardResult 一次軌道轟炸嘗試的結果(供 UI/測試檢視)。
type GroundBombardResult struct {
	Ok          bool   // 是否成功發動了一場轟炸解算(false = 前置條件不足,未開打)
	Reason      string // Ok=false 時的原因
	TotalDamage int    // 原版固定三外圈（炸彈另套 5／10 當量）的累積傷害
	Hits        int    // gamedata.StrategicBombardmentHitsFromDamage(TotalDamage)，原版 runtime /40

	// PopulationLost 是 sub_DCEBD 隨機候選池與後續生物武器實際扣掉的人口。
	PopulationLost    int
	RemainingHits     int // 扣完建築+人口後剩餘、未消耗掉的 hits(通常應為 0;殖民地人口歸零時會 > 0)
	MarinesLost       int // sub_DCEBD 寫回的駐防陸戰隊損失
	TanksLost         int // sub_DCEBD 寫回的駐防戰車損失
	BuildProgressLost int // sub_DCEBD 結果 +0x43 對應的建造進度損失

	// BioWeaponKills 是**生物武器**額外殺掉的人口(手冊 p.99,第 52 項(生物武器分類)接的)。
	// 已含在 PopulationLost 裡,獨立記一份是為了讓「這幾個人是被孢子殺的」看得出來
	// ——屏障護盾把這一項擋成 0 而其他傷害照常,兩者混在一起就看不出護盾有沒有生效。
	BioWeaponKills int

	// BuildingsDestroyed 是本次轟炸摧毀的建築數(2026-07-11 新增,#7/#8 接線:見下方函式註解
	// 「建築吸收」段落)。0 表示沒有建築被摧毀(可能是 hits 不夠、也可能是該殖民地本來就沒有
	// 建築——兩種情況本欄位都合法回 0,不額外區分)。
	BuildingsDestroyed int
	// BuildingsRemaining 是轟炸結束後該殖民地剩餘的建築數(len(ColonyBuildings[colonyIdx])),
	// 供 UI/測試檢視「還剩多少建築沒被炸掉」。
	BuildingsRemaining int

	// DefenderRetaliated 這次轟炸是否觸發了防禦方反擊(2026-07-11 新增,見下方函式「防禦方
	// 反擊」段落)。true 表示 BuildingsRemaining 裡至少有一座軌道基地(星基/戰鬥站/星辰要塞)
	// 或飛彈基地存活,並已對玩家艦隊打了一輪反擊齊射。false 表示本次轟炸把防禦建築全炸掉了
	// (或該殖民地本來就沒有這些建築)——沒有存活防禦建築時完全不呼叫反擊解算,逐位元回歸
	// 加這個機制之前的行為。
	DefenderRetaliated bool
	// AttackerShipsLost 是防禦方反擊這輪齊射擊沉的玩家艦艇數(DefenderRetaliated=false 時
	// 恆為 0)。已從 s.Fleet().Ships 移除(移除規則見下方函式註解)。
	AttackerShipsLost int

	// ColonyName 是被轟炸的星名(供轟炸畫面標題;engine.ColonyState 本身沒有名稱欄位,
	// 與 GroundInvasionResult.ColonyName 同款處理)。
	ColonyName string
	// PopulationBefore 是開炸前的殖民地人口,供畫面顯示「炸掉多少 / 原本多少」。
	PopulationBefore int

	// PlanetHitsRequired 是 Get_Colony_Hits_ @ 0x42371 已證實的轟炸後殖民地本體耐久：
	// 人口 + 士兵 + 戰車 + 每棟非軌道建築 40。Battlestation／Star Base／Star Fortress
	// 另有戰鬥者，不重複計入。它供 UI／測試檢視，不反向定義尚未追回的傷亡分配順序。
	PlanetHitsRequired int
}

func originalColonyBuildingIDs(buildings map[string]bool) []int {
	ids := make([]int, 0, len(buildings))
	for name, active := range buildings {
		if !active {
			continue
		}
		if id, ok := gamedata.OriginalBuildingIDForName(name); ok {
			ids = append(ids, id)
		}
	}
	return ids
}

// retaliationAttackers 依「轟炸建築吸收階段之後仍存活」的防禦建築,組出防禦方反擊用的
// []combatant(供 battleVolley 當 attackers)。
//
// 2026-07-11 改版(diff 全量表 #14「衛星/軌道防禦基地當 space 預算武器平台」):不再用固定的
// shipStrength tier 表(舊版 4/8/16),改把每座防禦建築建模成「固定 space 預算,塞入 defender
// (AI)依科技解鎖的最佳武器,beam 佔格另外套用版本相依 arc-cost」,由此推導反擊戰力——見
// gamedata/satellite.go(space/scale 常數 + fit 公式)與 bestUnlockedWeaponValue(挑最佳已解鎖
// 武器,ground_invasion.go)。
//
// ⚠ 誠實標註(近似,非手冊逐字數字,設計已由使用者確認採用並校準):本 remake 沒有「殖民地
// 綁定的正規太空戰」(openorion2 是渲染殼、無獨立殖民地空戰引擎),故用「軌道轟炸階段冒充太空
// 戰、由防禦建築反擊」近似手冊 p.129 一節「軌道基地/飛彈基地會對轟炸艦隊開火」的描述。
// 星基/戰鬥站/星辰要塞的 space 預算(250/500/1200)是【近似】值(比照 ShipHullSpace 同量級
// 艦體借用,見 satellite.go 檔頭說明);飛彈基地 300、地面砲台 450 space 是手冊 p.78/p.81
// 【確認】值。SatelliteStrengthScale=20 是把「塞入把數 × 武器 Value」換算回艦級 atk 的校準
// 除數,選定依據見 gamedata.SatelliteStrengthScale 註解的完整推導(雷射參考點下星基/戰鬥站
// 重現舊 tier 4/8,星辰要塞算出 20,非近似 19——如實記錄,不假造成 19)。
//
//   - 軌道基地(擇一,取代不疊加——手冊:星辰要塞取代同軌道的戰鬥站/星基,不共存):
//     hull space 分別取 gamedata.StarFortressSpace/BattlestationSpace/StarBaseSpace,配
//     defender 已解鎖的最佳 beam 武器(bestUnlockedWeaponValue),beam 佔格套用
//     profile.SatelliteBeamArcCostPct 的 arc-cost 後算 fit,fit*Value/SatelliteStrengthScale
//     即 atk(最少 1,避免有基地卻算出 0 戰力的不合理狀態)。
//   - 飛彈基地(若存活,與上面軌道基地並列,不互斥——手冊描述是兩種獨立的行星防禦設施):
//     hull space 固定 gamedata.MissileBaseSpace(300,確認值),配 defender 已解鎖的最佳
//     missile 武器——missile 不吃 beam 的 arc-cost(見 SatelliteBeamSpaceWithArc 註解),
//     直接用 WeaponSpaceByName 原始佔格算 fit。
//   - 地面砲台(若存活；玩家與 AI 的正常建造候選都可完成此建築):
//     hull space 固定 gamedata.GroundBatterySpace(450,確認值),套用
//     profile.GroundBatteryBeamArcCostPct 的 arc-cost,算法同軌道基地。
//   - wmin/wmax 換算比照 fleetBombardDamage/mkPlayerCombatants 同款慣例(wmin=atk/2,
//     wmax=atk)。shield/armor 刻意留 0:這些 attacker 不建模 HP,本輪基地不受玩家反殺
//     損傷(基地存續與否走「建築吸收」那條路徑,不走這裡的戰鬥解算)。
//
// 沒有任何防禦建築存活時回傳空 slice(呼叫端據此判斷 DefenderRetaliated=false,不呼叫
// battleVolley)。
func retaliationAttackers(buildings map[string]bool, defender engine.PlayerState, profile gamedata.RuleProfile) []combatant {
	var out []combatant

	// satelliteAtk 是「軌道基地/地面砲台」共用的推導公式:hull space 塞最佳 beam 武器
	// (套 arc-cost),換算成艦級 atk。
	satelliteAtk := func(hullSpace, arcCostPct int) int {
		bv, bs, _ := bestUnlockedWeaponValue(defender, profile, WeaponKindBeam)
		perBeam := gamedata.SatelliteBeamSpaceWithArc(bs, arcCostPct)
		fit := gamedata.SatelliteWeaponFitCount(hullSpace, perBeam)
		atk := fit * bv / gamedata.SatelliteStrengthScale
		if atk < 1 {
			atk = 1
		}
		return atk
	}

	orbitalHull := 0
	switch {
	case buildings["星辰要塞"]:
		orbitalHull = gamedata.StarFortressSpace
	case buildings["戰鬥站"]:
		orbitalHull = gamedata.BattlestationSpace
	case buildings["星基"]:
		orbitalHull = gamedata.StarBaseSpace
	}
	if orbitalHull > 0 {
		atk := satelliteAtk(orbitalHull, profile.SatelliteBeamArcCostPct)
		out = append(out, combatant{atk: atk, wmin: atk / 2, wmax: atk, kind: WeaponKindBeam})
	}

	if buildings["飛彈基地"] {
		mv, ms, _ := bestUnlockedWeaponValue(defender, profile, WeaponKindMissile)
		fit := gamedata.SatelliteWeaponFitCount(gamedata.MissileBaseSpace, ms) // missile 不吃 beam arc-cost
		atk := fit * mv / gamedata.SatelliteStrengthScale
		if atk < 1 {
			atk = 1
		}
		out = append(out, combatant{atk: atk, wmin: atk / 2, wmax: atk, kind: WeaponKindMissile})
	}

	if buildings["地面砲台"] {
		atk := satelliteAtk(gamedata.GroundBatterySpace, profile.GroundBatteryBeamArcCostPct)
		out = append(out, combatant{atk: atk, wmin: atk / 2, wmax: atk, kind: WeaponKindBeam})
	}

	// 戰機基地(手冊 p.79):地面的戰機中隊,依已解鎖的最高階戰機科技分三檔
	// ——攔截機 10 隊 / 轟炸機 6 隊 / 重戰機 4 隊。中隊數隨科技**遞減**(每階更強),
	// 照抄不要改成遞增。見 gamedata/planet_defense.go。
	//
	// ⚠ **已知問題,如實記錄**:用 remake 現行的單隊近似值算出來是
	// 攔截機 480 / 轟炸機 120 / 重戰機 256 ——**研究出轟炸機艙反而讓基地變弱**。
	// 那不是手冊的意思(手冊只給中隊數,沒說哪一檔比較強),是 `combat.go` 那兩個
	// 標明過的近似值(`fighterBeamDamageApprox=3` / `fighterBombDamageApprox=5`)造成的假象:
	// 攔截機每架 4 次射擊、轟炸機每架只 1 次投彈,4×3=12 對 1×5=5。
	//
	// **不在這裡硬調數字去湊一條好看的曲線**——要修的是那兩個近似值,而那需要手冊或
	// 反組譯給出戰機的真實傷害,目前兩邊都沒有。這段註解是給下一個要校準的人看的。
	if buildings["戰機基地"] {
		atk, _ := gamedata.FighterGarrisonCombatContribution(fighterGarrisonTierFor(defender))
		// 這裡的 atk 已是整座基地全部中隊的貢獻,不再過 SatelliteStrengthScale
		// ——那個除數是「hull space 塞武器」那條推導用的,戰機的數字是手冊直接給的中隊數。
		out = append(out, combatant{atk: atk, wmin: atk / 2, wmax: atk, kind: WeaponKindBeam})
	}

	// 行星版恆星轉換器:400 傷/面、不受距離與防禦影響。
	//
	// 2026-08-07 移到這裡:先前它只出現在 `colonyDefense`(ai_attack.go)裡,
	// 所以**它擋得住 AI 來襲、卻對軌道轟炸完全不反擊**——同一棟建築在兩條路徑上
	// 行為不一致。搬進 retaliationAttackers 之後兩邊共用同一個來源,
	// `colonyDefense` 那邊的獨立加總同步移除(否則會變成雙重計算)。
	if buildings[gamedata.StellarConverterName] {
		atk := gamedata.StellarConverterRetaliationAttack()
		out = append(out, combatant{atk: atk, wmin: atk, wmax: atk, kind: WeaponKindBeam})
	}

	return out
}

// fighterGarrisonTierFor 依 defender 已解鎖的最高階戰機科技決定戰機基地的檔次。
//
// 手冊那句「Note that Interceptors are available immediately」說明了為什麼沒有
// 「還沒有戰機科技」這一格——攔截機是開局就有的保底檔。
func fighterGarrisonTierFor(defender engine.PlayerState) gamedata.FighterGarrisonTier {
	//
	// ⚠ 兩個科技**不在同一個主題**:轟炸機艙在 TOPIC_ADVANCED_ROBOTICS(11)、
	// 重戰機艙在 TOPIC_SUPERSCALAR_CONSTRUCTION(42)。這是寫這段時猜錯、被
	// 第 37 項(研究樹一手驗證)挖出來的一手科技表(`gamedata.OrigTechTopic`)當場抓到的
	// ——原本兩個都寫 ADVANCED_ROBOTICS,重戰機那一檔會永遠進不去。
	switch {
	case groundEquipTechOwned(defender, gamedata.TOPIC_SUPERSCALAR_CONSTRUCTION, gamedata.TECH_HEAVY_FIGHTER_BAYS):
		return gamedata.FighterGarrisonHeavyFighter
	case groundEquipTechOwned(defender, gamedata.TOPIC_ADVANCED_ROBOTICS, gamedata.TECH_BOMBER_BAYS):
		return gamedata.FighterGarrisonBomber
	}
	return gamedata.FighterGarrisonInterceptor
}

// BombardColony 嘗試對 starIdx 這顆星發動一次軌道轟炸(手冊 p.129 Orbital Bombardment)。
// 前置條件與 InvadeColony 對稱,唯獨不需要已載運陸戰隊/戰車營(轟炸是艦隊武器對殖民地開火,
// 不需要地面部隊登陸):
//  1. 玩家艦隊已抵達該星(FleetAtStar==starIdx 且 FleetETA==0)。
//  2. 該星是敵方(Owner==2)且有「已建模」的殖民地(findAIColonyByStar 找得到)。
//  3. 玩家艦隊至少有 1 艘艦(len(s.Fleet().Ships)>0,否則無武器可轟炸)。
//
// 任一條件不足回傳 Ok=false + Reason,不消耗任何狀態、不呼叫 rng。
//
// 解算:fleetBombardDamage 模擬原版固定三外圈（炸彈另套版本 5／10 次當量）→
// gamedata.StrategicBombardmentHitsFromDamage 依原版 /40 換算 hits →
// gamedata.ResolveStrategicColonyDamage 依 sub_DCEBD @ 0xDCEBD 的候選池，隨機回寫
// 可摧毀建築、陸戰隊、戰車、建造進度與人口。
//
// 傷亡分配由 IDA 已證實的 sub_DCEBD 候選池決定；一般建築、駐軍、建造進度與人口
// 在同一個隨機池，不再使用舊版字母序建築吸收／行星尺寸人口近似。raw
// 8/9/26/27/40/41/42/47 由其他戰鬥者或結果分支處理，因此不進本 helper。1.3 的
// BombardmentBuildingBonusHits 仍是 CHANGELOG 語意近似，非本份 1.50 executable 證實值。
//
// ⚠ 範圍限制(誠實標註,非本函式應臆測補齊的部分):
//   - 不扣「儲存生產」/駐軍——AI 沒有這些的持久資料可扣,扣了會是憑空生資料,故不做(建築已
//     於本輪補上,見上方「建築吸收」)。
//   - 本輪不做「防禦方反擊摧毀玩家艦艇」(軌道基地對轟炸艦隊開火)——那是下一輪工作,本函式
//     不改動 s.Fleet().Ships。
//   - 手冊未講「殖民地人口被轟炸到 0」時的後續(是否直接摧毀殖民地/移除星系 Owner):不在本
//     函式臆測補上,留給未來確認手冊或 openorion2 行為後再接(TODO)。目前行為是 Population
//     可以停在 0,殖民地本身仍存在於 aiPlayer.Colonies(不會被移除)。
//   - 轟炸不會使殖民地被佔領(手冊:入侵才佔領),故本函式不改動 Star.Owner。
//
// rng 依「回合數 + 星索引」種子化(與 InvadeColony 同款慣例,但另加不同的乘數避免與入侵用的
// rng 種子巧合撞在一起),同一回合對同一顆星重複呼叫必得到相同結果(建築摧毀順序本身不吃 rng,
// 見上方「分配順序」)。
func (s *GameSession) BombardColony(starIdx int) GroundBombardResult {
	s.recordPlayerCommand(PlayerCommand{Name: CmdBombardColony, Args: []int{starIdx}})
	if starIdx < 0 || starIdx >= len(s.Stars) {
		return GroundBombardResult{Reason: "無效的星索引"}
	}
	if s.Fleet().AtStar != starIdx || s.Fleet().ETA != 0 {
		return GroundBombardResult{Reason: "艦隊尚未抵達該星"}
	}
	star := &s.Stars[starIdx]
	if star.Owner != 2 {
		return GroundBombardResult{Reason: "該星不是敵方殖民地"}
	}
	if len(s.Fleet().Ships) == 0 {
		return GroundBombardResult{Reason: "艦隊沒有可轟炸的艦艇"}
	}
	aiIdx, colonyIdx, ok := s.findAIColonyByStar(starIdx)
	if !ok {
		return GroundBombardResult{Reason: "該星無可轟炸的殖民地模型(簡化限制,見 AIOpponent.ColonyStars)"}
	}
	aiPlayer := &s.AIPlayers[aiIdx]
	colony := &aiPlayer.Colonies[colonyIdx]

	// 建築清單要**先**取出來:行星護盾是建築給的,而它影響的是逐發傷害,
	// 所以必須在解算傷害之前就知道有沒有護盾(先前這段在傷害之後,護盾接不上)。
	var buildings map[string]bool
	if colonyIdx < len(aiPlayer.ColonyBuildings) {
		buildings = aiPlayer.ColonyBuildings[colonyIdx]
	}
	shield := gamedata.PlanetaryShieldReduction(buildings)
	// 生物武器擋不擋必須在一般傷亡分配前取樣；屏障護盾即使同輪被抽中摧毀，
	// 仍會擋住這一輪已開始投放的生物武器。
	bioBlocked := gamedata.BiologicalWeaponBlocked(buildings)

	rng := rand.New(rand.NewSource(int64(s.Turn)*2654435761 + int64(starIdx)*131 + 777))
	totalDamage := s.fleetBombardDamage(rng, shield)
	hits := gamedata.StrategicBombardmentHitsFromDamage(totalDamage)

	res := GroundBombardResult{Ok: true, TotalDamage: totalDamage, Hits: hits,
		ColonyName: s.starName(starIdx), PopulationBefore: colony.Population}

	marines, tanks := 0, 0
	if colonyIdx < len(aiPlayer.ColonyMarines) {
		marines = aiPlayer.ColonyMarines[colonyIdx]
	}
	if colonyIdx < len(aiPlayer.ColonyTanks) {
		tanks = aiPlayer.ColonyTanks[colonyIdx]
	}
	buildingCost := gamedata.GroundPlanetHitsPerBuilding + s.RuleProfile.BombardmentBuildingBonusHits
	casualties := gamedata.ResolveStrategicColonyDamage(gamedata.StrategicColonyDamageState{
		Population: colony.Population, LastPopulationPoints: colony.BombardmentLastPopulationPoints,
		Marines: marines, Tanks: tanks, BuildProgress: colony.BombardmentBuildProgress,
		RawBuildingIDs:  originalColonyBuildingIDs(buildings),
		MarineHitCost:   gamedata.GroundMarineHitsToKill(aiRaceHasTrait(*aiPlayer, gamedata.TRAIT_HIGH_G), hasPoweredArmorFor(aiPlayer.Player)),
		TankHitCost:     tankHitsToKillFor(aiPlayer.Player, aiRaceHasTrait(*aiPlayer, gamedata.TRAIT_HIGH_G)),
		BuildingHitCost: buildingCost,
	}, hits, rng.Intn)
	colony.Population = casualties.State.Population
	colony.BombardmentLastPopulationPoints = casualties.State.LastPopulationPoints
	colony.BombardmentBuildProgress = casualties.State.BuildProgress
	if colonyIdx < len(aiPlayer.ColonyMarines) {
		aiPlayer.ColonyMarines[colonyIdx] = casualties.State.Marines
	}
	if colonyIdx < len(aiPlayer.ColonyTanks) {
		aiPlayer.ColonyTanks[colonyIdx] = casualties.State.Tanks
	}
	for _, destroyedID := range casualties.DestroyedBuildingIDs {
		for name, active := range buildings {
			if id, ok := gamedata.OriginalBuildingIDForName(name); active && ok && id == destroyedID {
				delete(buildings, name)
				break
			}
		}
	}
	normalizeColonyJobsAfterPopulationLoss(colony)
	res.PopulationLost = casualties.PopulationLost
	res.MarinesLost = casualties.MarinesLost
	res.TanksLost = casualties.TanksLost
	res.BuildProgressLost = casualties.BuildProgressLost
	res.BuildingsDestroyed = len(casualties.DestroyedBuildingIDs)
	res.BuildingsRemaining = len(buildings)
	res.RemainingHits = casualties.DamageRemaining

	// --- 生物武器(手冊 p.99;第 52 項(生物武器分類)接上)---
	//
	// 「invading ships must introduce them into the target planet's atmosphere **by orbital
	// bombardment**. Each spore pod launched has a 10% chance to kill one unit of colonist
	// population.」——所以投放點就是這裡,而且是**在一般轟炸傷害之外**再殺人口。
	//
	// 屏障護盾那句「biological weapons cannot enter the planet's atmosphere」是**完全擋掉**,
	// 不是減傷,所以擋住時整段跳過(見 gamedata.BiologicalWeaponBlocked)。
	//
	// ⚠ **莢數取「艦隊艦艇數」是 remake 的建模選擇。** 手冊說「每一個發射出去的孢子莢」,
	// 但沒說一次轟炸投幾莢,而 remake 沒有「哪幾艘船掛了生物武器、各帶幾莢」的模型。
	// 一艘船一莢是最直白的近似,不是手冊數字。
	if !bioBlocked {
		pct := gamedata.BestBiologicalWeaponKillPercent(func(tech gamedata.Technology) bool {
			topic, ok := gamedata.OrigTechTopic(tech)
			return ok && groundEquipTechOwned(s.Player, topic, tech)
		})
		if kills := gamedata.BiologicalWeaponPopKills(len(s.Fleet().Ships), pct, rng.Intn); kills > 0 {
			if kills > colony.Population {
				kills = colony.Population
			}
			colony.Population -= kills
			res.PopulationLost += kills
			res.BioWeaponKills = kills
			normalizeColonyJobsAfterPopulationLoss(colony)
		}
	}

	// --- 防禦方反擊(2026-07-11 新增,見 retaliationAttackers 函式頂部「誠實標註」段落):
	// 建築吸收 + 人口損失之後,只有「這次轟炸打完仍存活」的防禦建築(此刻的 buildings,已經過
	// 上方建築吸收迴圈刪除摧毀項)才有資格反擊——本次轟炸把防禦建築全炸掉時無反擊(壓制了
	// 防禦,合理),完全不呼叫下面的戰鬥解算(逐位元回歸加這個機制之前的行為)。
	if attackers := retaliationAttackers(buildings, aiPlayer.Player, s.RuleProfile); len(attackers) > 0 {
		defenders := s.mkPlayerCombatants()
		// 只打一輪齊射(battleVolley 本身就是「每個存活 attacker 對第一個存活 defender 射一發」
		// 的單輪函式,非到一方全滅的迴圈)——手冊語意是「一次轟炸換一次反擊」,不是讓基地把
		// 整支艦隊掃光,同一個種子化 rng(見上方 BombardColony 開頭)保持同回合同星可重現。
		shipsLost := battleVolley(attackers, &defenders, rng)
		res.DefenderRetaliated = true
		res.AttackerShipsLost = shipsLost
		for i := 0; i < shipsLost; i++ {
			s.removeWeakestShip()
		}
		// 艦隊被打薄後,運力池 MarineTransportCapacity()(= len(s.Fleet().Ships) * 每艘 4 名,見
		// ground_invasion.go)跟著縮小,已載運的 FleetMarines/FleetTanks 若超出新容量要夾下來
		// (LoadMarines/LoadTanks 平時只在「載運當下」檢查上限,不會在艦隊事後被打薄時自動夾,
		// 故轟炸反擊這裡要補做,否則會出現「容量 0 卻還載著陸戰隊」的不合理狀態)。
		if len(s.Fleet().Ships) == 0 {
			s.Fleet().Marines = 0
			s.Fleet().Tanks = 0
		} else if room := s.MarineTransportCapacity(); s.Fleet().Marines+s.Fleet().Tanks > room {
			// 陸戰隊優先保留、戰車營吃剩下的額度——比照 LoadTanks 的 room 扣除順序
			// (room = capacity - FleetMarines - FleetTanks,即「陸戰隊先佔額度」)。
			if s.Fleet().Marines > room {
				s.Fleet().Marines = room
				s.Fleet().Tanks = 0
			} else {
				s.Fleet().Tanks = room - s.Fleet().Marines
			}
		}
	}

	// 原版 Get_Colony_Hits_ 讀的是轟炸後仍在 Colony record 裡的實際人口、士兵、戰車與
	// 建築旗標，不以人口重新推算兵營應有部隊。平行陣列由 normalizeAIColonyGroundForces
	// 在各玩家路徑維持；舊存檔缺欄位時安全回退為 0。
	marines, tanks = 0, 0
	if colonyIdx < len(aiPlayer.ColonyMarines) {
		marines = aiPlayer.ColonyMarines[colonyIdx]
	}
	if colonyIdx < len(aiPlayer.ColonyTanks) {
		tanks = aiPlayer.ColonyTanks[colonyIdx]
	}
	res.PlanetHitsRequired = gamedata.OriginalColonyCombatHits(
		colony.Population, marines, tanks, originalColonyBuildingIDs(buildings),
	)

	return res
}

func normalizeColonyJobsAfterPopulationLoss(colony *engine.ColonyState) {
	if colony == nil {
		return
	}
	for colony.Farmers+colony.Workers+colony.Scientists > colony.Population {
		switch {
		case colony.Workers >= colony.Farmers && colony.Workers >= colony.Scientists && colony.Workers > 0:
			engine.RemovePopulationGroupUnit(colony, gamedata.WORKER)
			colony.Workers--
		case colony.Farmers >= colony.Scientists && colony.Farmers > 0:
			engine.RemovePopulationGroupUnit(colony, gamedata.FARMER)
			colony.Farmers--
		case colony.Scientists > 0:
			engine.RemovePopulationGroupUnit(colony, gamedata.SCIENTIST)
			colony.Scientists--
		default:
			return
		}
	}
}
