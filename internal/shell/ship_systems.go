package shell

import "github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"

// ship_systems.go:第 133 項盤點出來的「仍缺、可做」那一桶,先接數字最硬的幾個。
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
// ============ 順便修一個第 133 項的缺失 ============
//
// 慣性穩定器的手冊原文是:
//
//	The result is a **+50 addition to the ship's beam defense**, **+25 to the ship's
//	missile evasion**, and a halving of the movement cost for turning the ship in place.
//
// 第 133 項只接了 +25 那一半——當時是從 `gamedata/missile.go` 那一側找元件的,
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
	if sh.Special == "戰鬥掃描器" {
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

// ============ 這一輪**沒有**接的,與各自的理由 ============
//
//	戰鬥艙(Battle Pods)        「add equipment space without increasing the hull size」
//	                            ——remake 沒有逐元件佔格的造艦模型,加了不影響任何東西
//	保安站(Security Stations)   「+20 to the combat rolls of the Marines defending against
//	                            enemy boarding parties」——**登艦戰機制不存在**(第 119 項)
//	增強引擎(Augmented Engines) 「+5 combat speed」——**戰鬥速度模型不存在**(第 128 項)
//	時間扭曲加速器               「an additional round of activity」——同上,需要回合結構
//	測距瞄準器(Rangemaster)     「reducing the absolute range to one-third」——remake 的
//	                            快速結算固定 range=2,只有格子戰術有真距離;接了會讓兩條路
//	                            不一致,而那正是第 131–133 項一直在防的事
//	結構分析儀 / 阿基里斯瞄準器   都要動 `ResolveShotWithMods` 的傷害鏈(過盾後加倍 /
//	                            光束一律無視裝甲)。該函式的參數已經排到第 11 個,
//	                            再加下去該先把攻方/守方系統各收成一個結構——那是重構,
//	                            不該夾在資料項裡做
//	匿蹤力場 / 相位匿蹤          需要「艦艇可見性」與「不可被攻擊」狀態,remake 沒有
//	超載電容 / 快速飛彈架        「一回合開兩次火」需要回合內的射擊次數模型
//	轟炸機庫                    需要戰機中隊模型的轟炸分支(戰機庫已有,轟炸機沒有)
//	能量吸收器                  「吸收 1/4 傷害並可回射」需要儲能狀態
//	傳送器                      需要護盾分面(同第 128 項的電漿網)
//	多相護盾以外的護盾系
