package gamedata

// cloak.go:匿蹤家族 + 測距瞄準器 + 能量吸收器的手冊數值。
//
// 這四項先前都在「擋門理由已經過期或本來就錯」那一格(ship_systems.go 檔尾),
// 共通點是**手冊給了確切數字,而 remake 缺的機制其實都已經建好了**:
//
//	隱形裝置   要「這一回合有沒有開火」→ 第 70 項(陀螺去穩器)的 Fired 就是這個語意
//	相位匿蹤   要「不可被選為目標」→ 停滯力場(第 69 項(戰鬥速度與引擎階))已經有這個形狀
//	測距瞄準器 要「真實格距離」→ 格子戰術本來就傳 dist,不是固定 2
//	能量吸收器 要跨回合的儲能 → 第 70 項(陀螺去穩器)建了 Charged/Fired 的跨回合狀態

// ShipCloakingDeviceBeamDefense 是隱形裝置在**未開火**時對光束的防禦加成。
//
// 手冊逐字:「as long as the ship does not attack, it has an **+80 bonus to its defense
// against beam weapons**」。
const ShipCloakingDeviceBeamDefense = 80

// ShipCloakingDeviceMissileMissChance 是隱形裝置在未開火時,來襲飛彈/魚雷的未命中機率(%)。
//
// 手冊逐字:「missiles and torpedoes have a **50% chance to miss**」。
//
// ⚠ 這是**獨立於閃避判定的一道**,不是把閃避加 50:手冊把它與 +80 光束防禦並列成
// 兩種武器各自的規則,而 remake 的飛彈解算本來就把閃電場/位移裝置做成各自一道獨立擲骰
// (見 MissileDefenses)。放在同一個位置最不容易與既有規則互相汙染。
const ShipCloakingDeviceMissileMissChance = 50

// ShipPhasingCloakCombatRounds 是相位匿蹤能維持「完全不可被攻擊」的戰鬥回合數。
//
// 手冊逐字:「While cloaked, the ship cannot be attacked. **After 10 turns in combat**,
// side effects … Its effect must be tuned down, and thus it **functions just like a
// Cloaking Device** until the end of that combat.」
//
// 也就是說相位匿蹤不是「更強的隱形裝置」,是**前 10 回合無敵、之後退化成隱形裝置**。
const ShipPhasingCloakCombatRounds = 10

// EnergyAbsorberStoredFraction 是能量吸收器轉存的比例分母:手冊「**One-quarter** of all
// the potential damage that reaches a ship is diverted to and stored」。
const EnergyAbsorberStoredFraction = 4

// EnergyAbsorberStored 回傳一次攻擊會被轉存的能量。
//
// ⚠ **「potential damage that reaches a ship」的讀法**:取「已命中、尚未扣護盾與裝甲」
// 那一刻的傷害。手冊沒有明說是扣盾前還是扣盾後,兩種讀法都通;取扣盾前是因為手冊用的詞是
// 「reaches a ship」(抵達這艘船)而不是「penetrates」(穿透)——同一段手冊講結構分析儀時
// 用的正是 penetrate,兩個詞在同一份文件裡分得很開。標在這裡,不假裝手冊講了。
func EnergyAbsorberStored(potentialDamage int) int {
	if potentialDamage <= 0 {
		return 0
	}
	return potentialDamage / EnergyAbsorberStoredFraction
}

// RangemasterRangeSquares 回傳測距瞄準器修正後、**只用於命中判定**的距離。
//
// 手冊逐字:「reducing the absolute range (which is used to compute accuracy and to hit
// penalties) to **one-third** of the actual range. Note that the **dissipation of damage
// potential is not affected** by this system.」
//
// 兩句話對應程式裡兩個不同的位置:射程等級算兩次,一次給 to-hit(用這個縮短後的距離)、
// 一次給傷害衰減(用真實距離)。先前 remake 只算一次,所以這個系統無論怎麼接都會連傷害
// 一起改——那正是手冊特地寫了第二句要排除的事。
func RangemasterRangeSquares(squares int) int {
	if squares <= 0 {
		return 0
	}
	return squares / 3
}
