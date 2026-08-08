package shell

import "github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"

// ship_systems.go:第 68 項(元件盤點+飛彈防禦)盤點出來的「仍缺、可做」那一桶,先接數字最硬的幾個。
//
// ============ 選擇標準 ============
//
// 那一桶有 20 個。這一輪只接**同時滿足兩個條件**的:
//
//	① 手冊給了確切數字(不是「considerably」「vastly」這種形容詞)
//	② remake 已經有承接它的位置(不必先造一個新機制)
//
// 沒接的與理由寫在檔尾,不是漏掉。
//
// ============ 順便修一個第 68 項(元件盤點+飛彈防禦)的缺失 ============
//
// 慣性穩定器的手冊原文是:
//
//	The result is a **+50 addition to the ship's beam defense**, **+25 to the ship's
//	missile evasion**, and a halving of the movement cost for turning the ship in place.
//
// 第 68 項(元件盤點+飛彈防禦)只接了 +25 那一半——當時是從 `gamedata/missile.go` 那一側找元件的,
// 而那個檔案只收飛彈相關的常數,`MissileInertialStabilizer = 25` 看起來就是它的全部效果。
// **從單一檔案回推元件效果會漏東西**;正解是回去讀手冊那一條的全文。
// (`gamedata.BeamDefense` 其實早就有 `inertialStabilizer` 參數並且加 50,只是 shell
// 這條路徑沒有走它——又一個「規則在、呼叫端不在」。)

// shipBeamDefenseBonus 回傳這艘船的元件提供的**光束閃避**加成。
//
// 對照 `gamedata.BeamDefense`(移植自 openorion2 `ShipDesign::beamDefense`):
// 慣性抵消器 +100、慣性穩定器 +50。手冊的慣性穩定器條目也寫著同一個 +50。
func shipBeamDefenseBonus(sh Ship) int {
	switch sh.Special {
	case "慣性抵消器":
		return gamedata.ShipInertialNullifierBeamDefense
	case "慣性穩定器":
		return gamedata.ShipInertialStabilizerBeamDefense
	}
	return 0
}

// shipBeamOffenseBonus 回傳這艘船的元件提供的**光束命中**加成。
//
// 手冊(Battle Scanner):「The scanner increases the ship's chance to hit with beam
// weapons by **50**.」與 `gamedata.BeamOffense` 的 battleScanner 分支同值。
func shipBeamOffenseBonus(sh Ship) int {
	if sh.Special == battleScannerName {
		return gamedata.ShipBattleScannerBeamOffense
	}
	return 0
}

// shipStructureMultiplier 回傳這艘船的結構 HP 倍率(百分點,100 = 無加成)。
//
// 手冊(Reinforced Hull):「A Reinforced Hull **triples** the amount of structural damage
// a ship can sustain before being destroyed.」
//
// ⚠ 手冊同一句還說「also provides more protection for the ship's drive system, tripling
// the amount of damage required to destroy it」——**引擎損毀不在 remake 的模型裡**,
// 那一半不接(remake 的船不是完好就是被擊沉 + 一個累積損傷值,沒有逐系統損毀)。
func shipStructureMultiplier(sh Ship) int {
	if sh.Special == "強化船體" {
		return gamedata.ShipReinforcedHullStructurePercent
	}
	return 100
}

// shipShieldMultiplier 回傳這艘船的護盾吸收倍率(百分點,100 = 無加成)。
//
// 手冊(Multi-Phased Shields):「increasing the maximum amount of damage that they can
// absorb by **50%**」。
func shipShieldMultiplier(sh Ship) int {
	if sh.Special == "多相護盾" {
		return gamedata.ShipMultiPhasedShieldPercent
	}
	return 100
}

// FleetResearchPoints 回傳艦隊裡偵察實驗室每回合產生的研究點數。
//
// 手冊逐字:「This system generates research points each turn; the number depends on the
// size of the ship: **Frigate = 1, Destroyer = 2, Cruiser = 4, Battleship = 8,
// Titan = 16, and Doom Star = 32.**」——罕見地把整張表都列出來了。
//
// ⚠ 手冊那一句還有下半:「the lab allows a fleet in combat with a space monster or the
// Guardian of Orion to analyze the opponent's biology or structure and seek out weaknesses」
// ——**那個弱點分析沒接**:remake 的怪物戰鬥沒有「弱點」這個概念。查得到 ≠ 用得上。
func (s *GameSession) FleetResearchPoints() int {
	total := 0
	for _, fl := range s.Fleets {
		for _, sh := range fl.Ships {
			if sh.Special != "偵察實驗室" {
				continue
			}
			class, ok := shipClassFromName(sh.Class)
			if !ok {
				continue
			}
			total += gamedata.ShipScoutLabResearch(class)
		}
	}
	return total
}

// armorLevelAboveTitanium 回傳這件裝甲比鈦裝甲高幾級(鈦 = 0)。
//
// 戰機 HP 用它——手冊那張戰機表的註腳:「*** base hit points are modified by **2 times
// armor level above Titanium**」。等級序就是 ArmorOptions 的排列(鈦→三鈦→佐特→中子素→
// 精金→氙素),與 gamedata.ArmorStructurePercent 那條階梯同一個順序。
func armorLevelAboveTitanium(name string) int {
	level := 0
	for _, c := range ArmorOptions {
		if c.UnlockTech == gamedata.TECH_NONE {
			continue // 「無裝甲」那一列不算級
		}
		if c.Name == name {
			return level
		}
		level++
	}
	return 0
}

// 狀態類武器的元件名(第 69 項(戰鬥速度與引擎階))。
const (
	tractorBeamName = "牽引光束"
	stasisFieldName = "停滯力場"
	// battleScannerName 的第二個效果(戰鬥之外 +2 parsec 掃描)在 shell/detection.go
	// ——第 68 項(元件盤點+飛彈防禦)接了手冊那段的第一句就收工,第 71 項(探針③內部函式)才補上第二句。
	battleScannerName = "戰鬥掃描器"
)

// boolToInt 把「有沒有裝」換成束數。一艘船只有一個 Special 槽,所以最多一束
// ——手冊講的「multiple Tractor Beams … cumulative」在 remake 只能靠**多艘船**達成,
// 而那正好是原版最常見的用法(一群小船拖住一艘大船)。
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// 行動次數家族的元件名(第 70 項(陀螺去穩器))。
const (
	fastMissileRacksName    = "快速飛彈架"
	hyperXCapacitorsName    = "超載電容"
	timeWarpFacilitatorName = "時間扭曲加速器"
)

// shipShotsKind 回傳這艘船的「一回合多打一次」屬於哪一種。
func shipShotsKind(sh Ship) gamedata.ShotsPerRoundKind {
	switch sh.Special {
	case hyperXCapacitorsName:
		return gamedata.ShotsDoubleBeam
	case fastMissileRacksName:
		return gamedata.ShotsDoubleMissile
	case timeWarpFacilitatorName:
		return gamedata.ShotsDoubleAny
	}
	return gamedata.ShotsNormal
}

// TacticalShotsThisRound 回傳這艘船在格子戰術裡這一回合能開幾次火(第 70 項(陀螺去穩器))。
func TacticalShotsThisRound(sh CombatShip) int {
	if sh.InStasis {
		return 0 // 被停滯力場定住:不能開火(第 69 項(戰鬥速度與引擎階))
	}
	beam := sh.Kind == WeaponKindBeam
	missile := sh.Kind == WeaponKindMissile
	return gamedata.ShotsThisRound(sh.ShotsKind, beam, missile, sh.Charged)
}

// TacticalAdvanceCharge 依這一回合有沒有開過火,更新下一回合的充能狀態。
//
// 手冊的 unused 是「完全沒開火」——連射之後再單射並不會充到電(見 gamedata 檔頭)。
// 沒有冷卻的系統(時間扭曲加速器)一律保持滿電。
func TacticalAdvanceCharge(ships []CombatShip) {
	for i := range ships {
		sh := &ships[i]
		if !gamedata.ShotsNeedsCooldown(sh.ShotsKind) {
			sh.Charged = true
		} else {
			sh.Charged = gamedata.ShotsRecharge(sh.Fired)
		}
		// 匿蹤與充能讀**同一份還沒被清掉的 Fired**(手冊兩邊都是「一整回合沒開火」),
		// 所以放在這裡而不是另外開一個迴圈——清除順序一錯就會差一回合。
		CloakAdvanceRound(sh)
		sh.Fired = false
	}
}

// shipBeamAttackerSystems 把這艘船的元件翻成光束射擊要用的攻方系統旗標。
//
// 一艘船只有一個 Special 槽,所以最多只會有一項為真——寫成一個函式而不是三個判斷,
// 是為了讓「攻方有哪些系統」只有一個真相來源(與第 68 項(元件盤點+飛彈防禦)把傷害鏈收成結構同一個理由)。
func shipBeamAttackerSystems(sh Ship) BeamAttackerSystems {
	return BeamAttackerSystems{
		HEFBonus:           hefBonusFor(sh.Special == highEnergyFocusName),
		StructuralAnalyzer: sh.Special == "結構分析儀",
		AchillesUnit:       sh.Special == "阿基里斯瞄準器",
		Rangemaster:        sh.Special == rangemasterName,
	}
}

// ============ 手冊有、remake 還沒接的元件,與各自的理由 ============
//
// ⚠ 2026-08-08(第 72 項(元件表有≠效果有接))整份重寫。上一版列了 12 項,其中**七項在第 68–70 項已經接掉了**
// (增強引擎 / 時間扭曲加速器 / 結構分析儀 / 阿基里斯瞄準器 / 超載電容 / 快速飛彈架 / 轟炸機庫),
// 理由卻還留在這裡——**這份清單自己就是它警告的那種東西**。每次接掉一項就要回來刪一行。
//
// --- 還沒接的三個,與各自的理由(2026-08-08 重寫)---
//
// B 那一格清空了:能量吸收器 / 戰鬥艙 / 相位匿蹤 / 測距瞄準器 / 隱形裝置五項這一輪全部接完
// (見 cloak.go、energy_absorber.go、special_device_map.go)。**做完一項就回來刪那一行**
// ——這份清單上一版有 12 項,其中 7 項早就接掉了,而沒有任何機制會提醒你。
//
//	保安站(Security Stations)   「+20 to the combat rolls of the Marines defending against
//	                            enemy boarding parties」——**登艦戰不存在**(第 60 項(打得準也閃得掉))
//	傳送器(Transporters)        同上,而且還需要護盾分面
//	匿蹤力場(Stealth Field)     「completely invisible on the Galaxy Map」——remake 的 AI 艦隊
//	                            是抽象戰力、沒有地圖座標(見 detection.go 檔頭),玩家自己的艦隊
//	                            對玩家永遠可見。**沒有「敵方看得到我的艦隊」這件事可以隱藏。**
//
// 三個共通點:缺的是**前置系統**(登艦戰、護盾分面、敵方艦隊的地圖表示),不是缺數字。
// 這與 B 那一格的五項是完全不同的狀況——那五項缺的一直只是「有沒有人回頭查」。
