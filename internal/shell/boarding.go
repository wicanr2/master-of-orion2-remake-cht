package shell

import "github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"

// boarding.go:登艦戰。第 80 項(登艦戰)把三個元件接到同一套規則。
//
// ============ 為什麼現在做得了 ============
//
// 手冊把解算方式**直接指回地面戰**:「The Marines boarding the ship and those defending
// the ship fight it out in the same way as ground troops do when a colony is invaded.」
// 而地面戰的原版解算器(`gamedata.ResolveGroundCombatOrig`,抄自 `Resolve_Ground_Combat_`
// @ `0xEC601`)早就在了。缺的從來不是公式,是**沒有人把那兩句話接起來**。
//
// 陸戰隊人數也早就在:`gamedata.ShipHullMarines`(手冊 p.121 的 Marines 欄,5/8/12/20/30/50),
// 部隊艙翻倍也已經接在 `MarineTransportCapacity` 上。
//
// ============ 三個元件的現況 ============
//
//	突擊艇  ✅ 本檔 + fighter.go 的 FighterAssaultShuttle
//	保安站  ✅ 守方 +20(手冊逐字)
//	傳送器  ✅ 護盾分面與 12 格前置已接；硬化護盾仍會阻擋傳送器。

// BoardingIntent 是登艦的目的。手冊把這兩種寫成玩家在 Board 按鈕之後的選擇。
type BoardingIntent int

const (
	// BoardingCapture 奪船:目標是殺光守軍。手冊「their object is to kill off the entire
	// defending crew. If they are successful, ownership of that ship changes.」
	BoardingCapture BoardingIntent = iota
	// BoardingRaid 突襲:目標是破壞內部系統。手冊「Marines cannot do damage directly to the
	// armor or structure of a target ship. Neither can they do damage to the shields」
	// ——所以突襲**不會**讓目標掉血,只會拆系統。
	BoardingRaid
)

// BoardingParty 是一次登艦行動的輸入。
type BoardingParty struct {
	Intent BoardingIntent
	// Marines / Strength 是攻方登艦的單位數與每單位戰力(同地面戰的 Strength 欄)。
	Marines  int
	Strength int
	// HitsToKill 是攻方單位的耐受值(gamedata.GroundMarineHitsToKill)。
	HitsToKill int
}

// BoardingDefense 是被登艦那一方的狀態。
type BoardingDefense struct {
	Marines    int
	Strength   int
	HitsToKill int
	// StrengthBonus 是領袖等額外規則給守方陸戰隊的固定戰力。保安站的 +20
	// 仍由 SecurityStations 另算，避免把元件與軍官效果混成同一個旗標。
	StrengthBonus int
	// SecurityStations 是保安站:手冊「add 20 to the combat rolls of the Marines defending」。
	SecurityStations bool
}

// BoardingResult 是一次登艦的結果。
type BoardingResult struct {
	// Captured 只有在 Intent==BoardingCapture 且守軍全滅時為真。
	Captured bool
	// AttackerSurvived / DefenderSurvived 是雙方剩餘的陸戰隊單位數。
	AttackerSurvived int
	DefenderSurvived int
	// SystemsDestroyed 只有 Intent==BoardingRaid 時有意義:這次突襲拆掉幾個內部系統。
	SystemsDestroyed int
	Rounds           int
}

// ResolveBoarding 解一次登艦戰。
//
// roll 與地面戰共用同一種注入式擲骰(確定性測試要能重現,見 gamedata.GroundRoll)。
//
// 建模說明:
//   - 雙方都只有**一種**部隊(陸戰隊),所以 GroundSide 的四個類型只用第 0 格。
//     地面戰那邊會用到坦克/機械戰士,登艦戰不會——手冊講的是 Marines。
//   - 保安站的 +20 加在**守方的 Strength** 上。手冊寫的是「combat rolls」,而
//     `GroundCombatRound` 的比較式是 `Strength + roll(100)`,加在 Strength 上與
//     「加在擲骰上」在這個式子裡是同一件事。
//   - 突襲的系統破壞:手冊「Any damage one attacker does, destroys one or two internal
//     systems」——**以「守方挨了幾下」計**,每一下擲一次 1/2。
func ResolveBoarding(p BoardingParty, d BoardingDefense, roll gamedata.GroundRoll) BoardingResult {
	if p.Marines <= 0 {
		return BoardingResult{DefenderSurvived: d.Marines}
	}
	mk := func(count, strength, hits int) *gamedata.GroundSide {
		if hits < 1 {
			hits = 1
		}
		var s, c, h [gamedata.GroundUnitTypes]int
		s[0], c[0], h[0] = strength, count, hits
		return gamedata.NewGroundSide(s, c, h)
	}
	defStrength := d.Strength + d.StrengthBonus
	if d.SecurityStations {
		defStrength += gamedata.ShipSecurityStationsDefenseBonus
	}
	atk := mk(p.Marines, p.Strength, p.HitsToKill)
	def := mk(d.Marines, defStrength, d.HitsToKill)

	res := BoardingResult{}
	// 突襲要逐回合看「守方挨了幾下」,所以不能直接用 ResolveGroundCombatOrig 一路跑到底。
	// 奪船則沒有這個需求,但兩者共用同一個迴圈可以少一條分支。
	const maxRounds = 10000
	for !atk.Exhausted() && !def.Exhausted() && res.Rounds < maxRounds {
		res.Rounds++
		gamedata.GroundCombatRound(atk, def, roll)
		if p.Intent == BoardingRaid && def.HitType != gamedata.GroundNone {
			res.SystemsDestroyed += gamedata.BoardingRaidSystemsDestroyed(roll(2))
		}
	}
	res.AttackerSurvived = atk.AliveUnits()
	res.DefenderSurvived = def.AliveUnits()
	res.Captured = p.Intent == BoardingCapture && res.DefenderSurvived == 0 && res.AttackerSurvived > 0
	return res
}

// 登艦相關元件名。
const (
	assaultShuttleName   = "突擊艇"
	securityStationsName = "保安站"
	transportersName     = "傳送器"
)

// shipHasSecurityStations / shipHasAssaultShuttles / shipHasTransporters 是元件比對。
func shipHasSecurityStations(sh Ship) bool { return shipHasSpecial(sh, securityStationsName) }
func shipHasAssaultShuttles(sh Ship) bool  { return shipHasSpecial(sh, assaultShuttleName) }
func shipHasTransporters(sh Ship) bool     { return shipHasSpecial(sh, transportersName) }

// ShipMarineComplement 回傳這艘船上的陸戰隊單位數。
//
// 手冊 p.121 的 Marines 欄給艦體上限(5/8/12/20/30/50),部隊艙翻倍
// (手冊 p.79「doubling the number of Marines on board a ship」)。
//
// ⚠ 這與 `MarineTransportCapacity` 讀同一組規則,但**語意不同**:那個算的是
// 「這支艦隊能載多少地面部隊去打殖民地」,這個算的是「這艘船上有多少人可以打登艦戰」。
// 手冊把它們寫成同一批人(「The additional marines both defend the ship and can board
// enemy ships」),所以兩邊用同一個數字是對的,但不要把兩個函式合併——它們會分開演化。
func ShipMarineComplement(sh Ship) int {
	class, _ := shipClassFromName(sh.Class)
	n := gamedata.ShipHullMarines(class)
	if shipHasTroopPods(sh) {
		n *= gamedata.GroundTroopPodsMultiplier
	}
	return n
}

// ============ 戰術畫面上的登艦 ============

// ShipBoardingReach 是相容舊呼叫端的無目標版本；它只保留突擊艇與貼身規則。
// 有目標時請用 ShipBoardingReachAgainst，才能判斷傳送器的護盾分面。
//
// 手冊給兩條投送路徑,各自的前置不同:
//   - **突擊艇**:飛過去放人,不要求貼身(中隊速度 6,實務上一回合到得了)。
//   - **傳送器**:12 格內直接傳送,但要求「面向攻擊方的護盾**已被打穿**」；硬化護盾
//     仍然阻擋傳送器(見 gamedata/boarding.go)。
//
// 兩者都沒有的船只能貼身(相鄰格)登艦。⚠ 這一條**是 remake 的補充,手冊沒有**:
// 手冊只講突擊艇與傳送器。標在這裡,不假裝是原版規則。
func ShipBoardingReach(att CombatShip, distance int) bool {
	if att.AssaultShuttles {
		return true
	}
	return distance <= 1
}

// ShipBoardingReachAgainst 是手冊的登艦射程判定。
// 傳送器要求 12 格內、面向攻擊方的護盾已失效；硬化護盾即使分面歸零仍會阻擋。
// 相鄰格則沿用 remake 既有的貼身登艦補充規則。
func ShipBoardingReachAgainst(att, def CombatShip, distance int) bool {
	if att.AssaultShuttles {
		return true
	}
	if distance <= 1 {
		return true
	}
	if !att.Transporters || distance > gamedata.TransporterRangeSquares || def.HardShield {
		return false
	}
	facing := ShieldFacingForShot(att, def)
	return def.ShieldFacingDown(facing)
}

// ShipBoardingPartySize 是這艘船這一次能派出去的陸戰隊單位數。
//
// 突擊艇是一個中隊 4 架、每架載 1 個單位,所以上限是 4;貼身登艦則傾巢而出。
// 兩者都不會超過艦上實際人數。
func ShipBoardingPartySize(att CombatShip) int {
	n := att.Marines
	if att.AssaultShuttles {
		if cap := gamedata.FighterSquadronSize * gamedata.AssaultShuttleMarinesEach; cap < n {
			n = cap
		}
	}
	return n
}

// ShipBoardingAttack 解一次戰術畫面上的登艦,並就地更新雙方的陸戰隊人數與奪船旗標。
//
// 戰力用的是**地面戰那一套**(手冊:「fight it out in the same way as ground troops do」),
// 攻守方所屬帝國的陸戰隊、艦員 Bo 與艦隊領袖加成都已在 CombatShip 進場時固定，
// 解算器不再把玩家資料錯套給 AI。
func (s *GameSession) ShipBoardingAttack(att, def *CombatShip, intent BoardingIntent, roll gamedata.GroundRoll) BoardingResult {
	sent := ShipBoardingPartySize(*att)
	if sent <= 0 {
		return BoardingResult{DefenderSurvived: def.Marines}
	}
	res := ResolveBoarding(
		BoardingParty{Intent: intent, Marines: sent,
			Strength:   att.MarineStrength + att.BoardingBonus + att.CommandoBonus,
			HitsToKill: att.MarineHitsToKill},
		BoardingDefense{Marines: def.Marines,
			Strength:         def.MarineStrength + def.BoardingBonus + def.CommandoBonus,
			HitsToKill:       def.MarineHitsToKill,
			StrengthBonus:    def.SecurityBonus,
			SecurityStations: def.SecurityStations},
		roll)

	def.Marines = res.DefenderSurvived
	def.SystemsDisabled += res.SystemsDestroyed
	// 突擊艇是單程的:陣亡的回不來,倖存的也留在對方船上。貼身登艦則倖存者算回自己船上。
	if att.AssaultShuttles {
		att.Marines -= sent
	} else {
		att.Marines += res.AttackerSurvived - sent
	}
	if att.Marines < 0 {
		att.Marines = 0
	}
	if res.Captured {
		def.Captured = true
	}
	return res
}
