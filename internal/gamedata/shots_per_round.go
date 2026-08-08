package gamedata

// shots_per_round.go:一回合開幾次火——三個手冊系統共用的機制。
//
// ============ 為什麼三個一起做 ============
//
// 第 69 項那一桶剩下的元件裡,有三個卡在**同一個**缺失機制:
//
//	超載電容      allows a ship's **beam weapons** to fire **twice in a single turn**
//	快速飛彈架    allow a ship to fire **two volleys of missiles** in a single turn
//	時間扭曲加速器 gain an **additional round of activity** at the end of each combat round
//
// remake 的戰術回合先前是「每艘船一回合開一次火」寫死在迴圈裡,三個都無處可接。
// 建一次「行動次數」就一次解三個——這是第 136(速度)/137(移動)/138(狀態)之後,
// 同一種「先建機制再放資料」的第四次。
//
// ============ 冷卻的讀法(手冊寫得比想像中嚴) ============
//
//	超載電容:After firing twice, these weapons **cannot be fired twice in a turn again
//	          until they have spent at least 1 full turn UNUSED**. It takes this turn to
//	          recharge the capacitor.
//	快速飛彈架:it cannot fire its missiles twice in 1 turn again until it has allowed them
//	          to **remain unused for 1 turn**.
//
// **「unused」不是「沒有連射」,是「完全沒開火」。** 兩條都用同一個字。
// 所以充能的條件是「整整一回合沒有開火」——連射之後再單射,仍然沒有充到電。
// 這個差別在實際對戰裡很明顯:玩家不能靠「連射→單射→連射」無限循環。
//
// 時間扭曲加速器**沒有冷卻**:手冊說「the ship gets two combat rounds for every one that
// normal ships get」——那是持續的,不是充放電。

// ShotsPerRoundKind 是一艘船「一回合能多打一次」的適用範圍。
type ShotsPerRoundKind int

const (
	// ShotsNormal 沒有任何加速系統。
	ShotsNormal ShotsPerRoundKind = iota
	// ShotsDoubleBeam 超載電容:只有光束武器連射。
	ShotsDoubleBeam
	// ShotsDoubleMissile 快速飛彈架:只有飛彈連射。
	ShotsDoubleMissile
	// ShotsDoubleAny 時間扭曲加速器:整個行動再來一次,不分武器,**且無冷卻**。
	ShotsDoubleAny
)

// ShotsThisRound 回傳這艘船這一回合能開幾次火。
//
// weaponIsBeam / weaponIsMissile 由呼叫端依武器分類判斷(見 shell.weaponKindByName)。
// charged 是「上一回合完全沒開火」——見檔頭對 unused 的讀法。
func ShotsThisRound(kind ShotsPerRoundKind, weaponIsBeam, weaponIsMissile, charged bool) int {
	switch kind {
	case ShotsDoubleAny:
		return 2 // 無冷卻
	case ShotsDoubleBeam:
		if weaponIsBeam && charged {
			return 2
		}
	case ShotsDoubleMissile:
		if weaponIsMissile && charged {
			return 2
		}
	}
	return 1
}

// ShotsNeedsCooldown 回報這種加速系統連射之後需不需要充能。
//
// 只有時間扭曲加速器不需要(手冊沒有那句 unused 的限制)。
func ShotsNeedsCooldown(kind ShotsPerRoundKind) bool {
	return kind == ShotsDoubleBeam || kind == ShotsDoubleMissile
}

// ShotsRecharge 依這一回合「有沒有開過火」算出下一回合的充能狀態。
//
// 手冊的 unused 是**完全沒開火**,所以 firedThisRound 為真時一律回 false
// ——連射之後再單射並不會充到電。
func ShotsRecharge(firedThisRound bool) bool { return !firedThisRound }
