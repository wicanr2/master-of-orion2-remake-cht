package gamedata

// boarding.go:登艦戰的手冊數值。
//
// 這是第 80 項(登艦戰)接上的三個元件(保安站 / 傳送器 / 突擊艇)共用規則。
// 手冊把規則寫得很完整,而且**把解算方式直接指回地面戰**:
//
//	The Marines boarding the ship and those defending the ship fight it out **in the same
//	way as ground troops do when a colony is invaded**.
//
// 所以這裡不需要新的解算公式——`ResolveGroundCombatOrig`(原版 `Resolve_Ground_Combat_`
// @ `0xEC601` 抄出來的那一份)就是登艦戰的解算器。這個檔案只放「登艦特有」的數字。

// ShipSecurityStationsDefenseBonus 是保安站給**守方**陸戰隊的加成。
//
// 手冊逐字:「Stations add **20** to the combat rolls of the Marines defending against
// enemy boarding parties.」
const ShipSecurityStationsDefenseBonus = 20

// 突擊艇(Assault Shuttles)的手冊數值。
//
// 手冊逐字:「Assault Shuttles are fighters (like the Interceptors) that carry **1 Marine
// unit**. … Shuttles are installed and launched in **squadrons of 4**. Each shuttle moves at
// **speed 6** modified by your best drive and can take **3 damage** modified by your best
// armor. Once launched, Assault Shuttles fly to the target ship and drop off their Marines,
// which board and attempt capture.」
//
// 中隊人數(4)與其他戰機共用 FighterSquadronSize,不另立常數。
const (
	AssaultShuttleMarinesEach = 1 // 每架載 1 個陸戰隊單位
	AssaultShuttleBaseSpeed   = 6 // 基礎速度(再吃引擎階,同其他戰機)
	AssaultShuttleBaseHits    = 3 // 每架基礎血量(再吃裝甲級數,同其他戰機)
)

// TransporterRangeSquares 是傳送器送陸戰隊的射程(格)。
//
// 手冊逐字:「Transporters allow a ship to send Marines onto an enemy ship from a range of
// **12 squares** — **if the shield facing the attacking ship is disabled**.」
//
// 傳送器的前置不是「射程」而是**面向攻擊方的那面護盾已失效**。格子戰術目前由
// shell.CombatShip 的四面容量承接；艦身旋轉與原版方向命名仍是未解的近似層。
const TransporterRangeSquares = 12

// TransporterBombRangeSquares 是傳送器對行星投彈的延伸射程。
//
// 手冊同一條的最後一句:「transporters extend the range at which a ship can drop bombs on
// a planet to **12 squares** from the normal **3**」。
const (
	TransporterBombRangeSquares = 12
	NormalBombRangeSquares      = 3
)

// 突襲(raid)一次命中摧毀的內部系統數。
//
// 手冊逐字:「Any damage one attacker does, destroys **one or two internal systems**.
// Smaller systems are more likely to be destroyed.」
//
// ⚠ 「smaller systems are more likely」在 remake 沒有落點——一艘船只有一個特殊系統槽,
// 沒有「大小不同的一堆系統」可以挑。所以 remake 只實作「一次 1~2 個」這一半,
// 而 1 或 2 由呼叫端擲骰決定。標在這裡,不假裝整句都做了。
const (
	BoardingRaidSystemsMin = 1
	BoardingRaidSystemsMax = 2
)

// BoardingRaidSystemsDestroyed 依擲骰回傳這一下摧毀幾個內部系統。
// roll 是 [0,2) 的整數(呼叫端提供,理由同 GroundRoll:確定性測試要能重現)。
func BoardingRaidSystemsDestroyed(roll int) int {
	if roll <= 0 {
		return BoardingRaidSystemsMin
	}
	return BoardingRaidSystemsMax
}
