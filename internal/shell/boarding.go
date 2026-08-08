package shell

import "github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"

// boarding.go:登艦戰。第 60 項(打得準也閃得掉)以來擋著三個元件的那個缺席機制。
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
//	傳送器  ❌ **仍然擋著**,而且擋門理由這次講清楚了:手冊的前置是「面向攻擊方的護盾
//	          **已經被打穿**」,而 remake 的護盾是每發固定減傷,**既沒有分面也不會崩**。
//	          缺的不是射程(那是 12 格,常數已經在 gamedata),是護盾要有「崩潰」這個狀態。

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
	defStrength := d.Strength
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
func shipHasSecurityStations(sh Ship) bool { return sh.Special == securityStationsName }
func shipHasAssaultShuttles(sh Ship) bool  { return sh.Special == assaultShuttleName }

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
