package shell

import "github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"

// cloak.go:匿蹤家族的**狀態機**。數值在 gamedata/cloak.go,這裡只管「現在是不是隱形」。
//
// 手冊把隱形寫成一個有記憶的狀態,不是一個常駐加成:
//
//	開場          隱形
//	一旦開火      失去隱形(當回合就失去,不是下回合)
//	停火一整回合  才能重新隱形
//
// 「停火一整回合」與行動次數家族的 `Fired` 是**同一個訊號**(第 70 項(陀螺去穩器)已經在記了),
// 所以這裡不另外開一份狀態,只是在回合結束時多讀它一次。

// 匿蹤家族的元件名。
const (
	cloakingDeviceName = "隱形裝置"
	phasingCloakName   = "相位匿蹤"
	rangemasterName    = "測距瞄準器"
	energyAbsorberName = "能量吸收器"
)

// CloakKind 是這艘船帶的匿蹤系統種類。
type CloakKind int

const (
	CloakNone CloakKind = iota
	// CloakDevice 是隱形裝置:未開火時 +80 光束防禦、飛彈 50% 未命中。
	CloakDevice
	// CloakPhasing 是相位匿蹤:未開火時**完全不可被選為目標**,滿 10 回合後降級成 CloakDevice。
	CloakPhasing
)

// shipCloakKind 依元件名判斷這艘船的匿蹤系統。
func shipCloakKind(sh Ship) CloakKind {
	if shipHasSpecial(sh, phasingCloakName) {
		return CloakPhasing
	}
	if shipHasSpecial(sh, cloakingDeviceName) {
		return CloakDevice
	}
	return CloakNone
}

// cloakEffectiveKind 回傳這艘船在第 round 回合**實際生效**的匿蹤種類。
//
// 相位匿蹤過了手冊的 10 回合就降級成隱形裝置(「functions just like a Cloaking Device
// until the end of that combat」)——降級是**永久**的(這場戰鬥內),所以判斷用的是
// 「戰鬥打到第幾回合」而不是「這艘船隱形了幾回合」。
func cloakEffectiveKind(k CloakKind, round int) CloakKind {
	if k == CloakPhasing && round > gamedata.ShipPhasingCloakCombatRounds {
		return CloakDevice
	}
	return k
}

// CloakUntargetable 回傳這艘船此刻是否**完全不能被選為目標**(相位匿蹤,且仍在 10 回合內)。
//
// 與停滯力場同一個形狀,但原因相反:停滯力場是被敵人定住,相位匿蹤是自己躲起來。
func CloakUntargetable(sh CombatShip, round int) bool {
	return sh.Cloaked && cloakEffectiveKind(sh.CloakKind, round) == CloakPhasing
}

// CloakBeamDefenseBonus 回傳這艘船此刻的匿蹤光束防禦加成(未隱形回 0)。
//
// 相位匿蹤在 10 回合內根本打不到(見 CloakUntargetable),所以它走到這裡時已經降級,
// 兩種匿蹤共用同一個 +80。
func CloakBeamDefenseBonus(sh CombatShip, round int) int {
	if !sh.Cloaked || cloakEffectiveKind(sh.CloakKind, round) == CloakNone {
		return 0
	}
	return gamedata.ShipCloakingDeviceBeamDefense
}

// CloakMissileMissChance 回傳來襲飛彈因這艘船隱形而未命中的機率(%),未隱形回 0。
func CloakMissileMissChance(sh CombatShip, round int) int {
	if !sh.Cloaked || cloakEffectiveKind(sh.CloakKind, round) == CloakNone {
		return 0
	}
	return gamedata.ShipCloakingDeviceMissileMissChance
}

// CloakOnFire 是「這艘船開火了」的通知:隱形立刻失效。
//
// 手冊逐字:「When a cloaked ship **does** attack, it loses these bonuses.」——是開火當下,
// 不是下一回合。
//
// ⚠ **能量吸收器例外**:手冊在能量吸收器那一條寫著「A cloaked ship will **not** decloak
// from firing its stored energy」。所以發射儲能不要呼叫這個函式。
func CloakOnFire(sh *CombatShip) {
	sh.Cloaked = false
}

// CloakAdvanceRound 在回合結束時決定下一回合能不能重新隱形。
//
// 手冊逐字:「it must remain uncloaked until it spends **one full turn without firing**;
// then it can recloak.」——與行動次數家族的充能讀同一個 `Fired`,而且語意一致
// (兩邊的「沒開火」都是「完全沒開火」)。
//
// 呼叫端:TacticalAdvanceCharge。放在那裡而不是另外開一個迴圈,是因為兩者要讀的是
// **同一份還沒被清掉的 Fired**——清除順序一錯就會差一回合。
func CloakAdvanceRound(sh *CombatShip) {
	if sh.CloakKind == CloakNone {
		return
	}
	if !sh.Fired {
		sh.Cloaked = true
	}
}
