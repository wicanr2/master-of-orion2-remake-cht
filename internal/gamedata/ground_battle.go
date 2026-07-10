package gamedata

import "math/rand"

// 地面戰鬥「解算式」(Resolve),對應 docs/tech/ground-combat-algorithm.md 的
// 「★ 解算式定案(2026-07-10)」:
//
//	解算結構取自一代(1oom)game_ground.c 的 game_ground_kill——每回合雙方各擲一次 d100
//	加上自己的 force 加成,較低者敗(平手歸守方,即 v_atk <= v_def 時攻方受創);敗方
//	「最前一個單位」累積 1 hit,hits 達其 hits-to-kill 才真正陣亡(pop -1)。反覆至一方
//	單位歸零。
//
//	force 本身用二代(MOO2)手冊加成表算(GroundArmorTechBonus 等,已在 ground.go),
//	由呼叫端(或下方 builder)算好整數 force 後傳入,本檔不重算加成表。
//
// 本檔不引入 wall-clock 亂數:所有隨機性一律透過呼叫端傳入的 *rand.Rand,以便寫確定性
// (seed 化)測試。

// GroundUnit 地面戰的最小單位:目前剩餘的「命中承受量」。
// 初始值 = 該單位的 hits-to-kill(見 GroundMarineHitsToKill / GroundTankHitsToKill /
// GroundBattleoidHitsToKill);每受 1 hit 遞減 1,降到 0 視為陣亡(從隊伍移除)。
type GroundUnit struct {
	RemainingHits int
}

// GroundForce 地面戰一方的兵力:一組 GroundUnit + 已經算好的 force 加成值。
//
// Force 由呼叫端依 docs/tech/ground-combat-algorithm.md 的公式算好後填入,本檔不重算:
//
//	force = Σ(GroundArmorTechBonus + GroundEquipmentTechBonus + GroundRaceCombatBonus
//	          + GroundBattleoidCombatBonus) 視情況套用 GroundApplyLowGPenalty;
//	守方另加 GroundSubterraneanBonus(true)。
//
// Defending 只作為文件用途保留(呼叫端算 Force 時已經把守方加成算進去),本檔的解算邏輯
// 本身不需要再讀 Defending 欄位——保留是因為呼叫端組 GroundForce 時常見需要記錄「這是哪一
// 方」,避免呼叫端另外包一層 struct。
type GroundForce struct {
	Units     []GroundUnit
	Force     int
	Defending bool
}

// AliveUnits 回傳目前仍存活(RemainingHits > 0)的單位數。
func (f GroundForce) AliveUnits() int {
	n := 0
	for _, u := range f.Units {
		if u.RemainingHits > 0 {
			n++
		}
	}
	return n
}

// NewGroundForce 建 GroundForce 的 builder helper:count 個單位,每個單位初始
// hitsToKill(呼叫端先用 GroundMarineHitsToKill / GroundTankHitsToKill /
// GroundBattleoidHitsToKill 算好),搭配已算好的 force 加成值。
//
// hitsToKill <= 0 視為 1(至少要能被命中一次陣亡,避免呼叫端傳錯值造成除以 0 或無限
// 不死的單位)。
func NewGroundForce(count, hitsToKill, force int, defending bool) GroundForce {
	if hitsToKill <= 0 {
		hitsToKill = 1
	}
	units := make([]GroundUnit, count)
	for i := range units {
		units[i] = GroundUnit{RemainingHits: hitsToKill}
	}
	return GroundForce{Units: units, Force: force, Defending: defending}
}

// GroundResult 一場地面戰的解算結果。
type GroundResult struct {
	AttackerWon      bool // true = 攻方勝(守方單位全滅)
	AttackerSurvived int  // 戰後攻方存活單位數
	DefenderSurvived int  // 戰後守方存活單位數
	Rounds           int  // 總共擲了幾回合(供測試/紀錄用,非規則需要)
}

// firstAliveIndex 回傳隊伍中「最前一個存活單位」的 index;找不到回 -1。
// 「最前一個」對應規則描述「敗方最前一個單位受創」——本實作以 slice 順序做為隊形順序
// (呼叫端決定單位順序即決定戰鬥中的「前排/後排」)。
func firstAliveIndex(units []GroundUnit) int {
	for i, u := range units {
		if u.RemainingHits > 0 {
			return i
		}
	}
	return -1
}

// ResolveGroundBattle 解算一場地面戰,回傳勝方與雙方存活單位數。
//
// 規則(docs/tech/ground-combat-algorithm.md 定案):
//
//	每回合:v_atk = d100 + atk.Force、v_def = d100 + def.Force(d100 = rng.Intn(100)+1)。
//	v_atk <= v_def → 攻方落敗一次(平手歸守方);否則守方落敗一次。
//	落敗方「最前一個存活單位」受 1 hit,RemainingHits 歸 0 即該單位陣亡。
//	反覆至一方無存活單位為止,該方落敗、另一方獲勝。
//
// atk/def 兩個 GroundForce 不會被就地修改——內部先深拷貝一份 Units 再操作,呼叫端傳入的
// 原始資料維持不變。
//
// 邊界:任一方一開始就沒有存活單位,直接判給另一方獲勝,不擲骰(避免 rng 呼叫次數因邊界
// 案例而不同,也避免 0 單位方仍可能「贏」的錯誤)。
func ResolveGroundBattle(atk, def GroundForce, rng *rand.Rand) GroundResult {
	atkUnits := append([]GroundUnit(nil), atk.Units...)
	defUnits := append([]GroundUnit(nil), def.Units...)

	atkAlive := countAlive(atkUnits)
	defAlive := countAlive(defUnits)

	if atkAlive == 0 && defAlive == 0 {
		// 雙方都沒兵力:規則對此未定義輸贏,以「守方勝」(攻方連存在的兵力都沒有,無法
		// 佔領)做為明確且不會誤導呼叫端的預設值。
		return GroundResult{AttackerWon: false, AttackerSurvived: 0, DefenderSurvived: 0}
	}
	if atkAlive == 0 {
		return GroundResult{AttackerWon: false, AttackerSurvived: 0, DefenderSurvived: defAlive}
	}
	if defAlive == 0 {
		return GroundResult{AttackerWon: true, AttackerSurvived: atkAlive, DefenderSurvived: 0}
	}

	rounds := 0
	for atkAlive > 0 && defAlive > 0 {
		rounds++
		vAtk := rng.Intn(100) + 1 + atk.Force
		vDef := rng.Intn(100) + 1 + def.Force

		if vAtk <= vDef {
			// 攻方落敗(含平手):攻方最前存活單位受 1 hit。
			idx := firstAliveIndex(atkUnits)
			atkUnits[idx].RemainingHits--
			if atkUnits[idx].RemainingHits <= 0 {
				atkAlive--
			}
		} else {
			// 守方落敗:守方最前存活單位受 1 hit。
			idx := firstAliveIndex(defUnits)
			defUnits[idx].RemainingHits--
			if defUnits[idx].RemainingHits <= 0 {
				defAlive--
			}
		}
	}

	return GroundResult{
		AttackerWon:      defAlive == 0,
		AttackerSurvived: atkAlive,
		DefenderSurvived: defAlive,
		Rounds:           rounds,
	}
}

func countAlive(units []GroundUnit) int {
	n := 0
	for _, u := range units {
		if u.RemainingHits > 0 {
			n++
		}
	}
	return n
}
