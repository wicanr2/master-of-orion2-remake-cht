package shell

import "github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"

// missile_defense.go:手冊 p.123 的飛彈防禦系統——把已經移植好的數字接上元件。
//
// ============ 缺的一直是元件,不是規則 ============
//
// `gamedata/missile.go` 早就把整段搬完了,每一個都附手冊原文與精確數字:
//
//	MissileJammerECM           70   電子干擾器
//	MissileJammerMultiWave    100   多波電子干擾器
//	MissileJammerWideAreaSelf 130   廣域干擾器(對自己)
//	MissileJammerWideAreaFleet 70   廣域干擾器(對艦隊其餘船艦,不與其他 jammer 疊加)
//	MissileInertialStabilizer  25   慣性穩定器
//	MissileInertialNullifier   50   慣性抵消器
//	MissileLightningFieldDestroyChance 50  閃電場:每一枚來襲飛彈/魚雷/戰機各 50% 直接摧毀
//	MissileDisplacementDeviceMissChance 30 位移裝置:一律 30% 完全未命中
//
// 七個常數,**生產端全部是 0**。`battleVolley` 的飛彈分支註解把理由寫得很清楚:
//
//	防守方的飛彈閃避先前恆傳 0(「艦艇設計/軍官系統尚未提供這些元件」)。
//	那句話對 ECM 干擾器/慣性穩定器**仍成立**
//
// 那句「仍成立」在寫下的當天是對的。這一項把那些元件補進 `SpecialOptions`,它就不成立了。
//
// ============ 誠實留白 ============
//
//   - **廣域干擾器只給自己那一格(130)。** 手冊另有「對艦隊其餘船艦 +70,且不與其他
//     jammer 疊加」——remake 的戰列是逐艦獨立的,沒有「艦隊層加成」這個概念,
//     硬塞會變成每艘船各自加一次,那不是手冊講的東西。艦隊層加成留白。
//   - **一艘船只有一個 Special 槽。** 原版可以同時裝干擾器與慣性穩定器,remake 不行
//     ——這是既有的設計限制(見 Ship.Special),不是這一項引入的。
//   - **匿蹤裝置的 50% 未命中沒接。** `MissileCloakingDeviceMissChance` 手冊寫明
//     「僅在裝置**啟動**時」,而 remake 沒有「啟動/未啟動」狀態。查得到 ≠ 用得上。

// shipMissileEvasionBonus 回傳這艘船的元件提供的飛彈閃避加成。
//
// 取**最佳**不是加總:一艘船只有一個 Special 槽,本來就裝不了兩個;
// 寫成 switch 是為了讓「哪個元件給多少」一眼看得出來,而不是把它藏進一張 map。
func shipMissileEvasionBonus(sh Ship) int {
	switch sh.Special {
	case "電子干擾器":
		return gamedata.MissileJammerECM
	case "多波電子干擾器":
		return gamedata.MissileJammerMultiWave
	case "廣域干擾器":
		return gamedata.MissileJammerWideAreaSelf
	case "慣性穩定器":
		return gamedata.MissileInertialStabilizer
	case "慣性抵消器":
		return gamedata.MissileInertialNullifier
	}
	return 0
}

// shipHasLightningField 回報這艘船裝了閃電場(每一枚來襲飛彈各 50% 直接摧毀)。
func shipHasLightningField(sh Ship) bool { return sh.Special == "閃電場" }

// shipHasDisplacementDevice 回報這艘船裝了位移裝置(飛彈一律 30% 完全未命中)。
func shipHasDisplacementDevice(sh Ship) bool { return sh.Special == "位移裝置" }

// shipHasTroopPods 回報這艘船裝了部隊艙(陸戰隊運力加倍)。
func shipHasTroopPods(sh Ship) bool { return sh.Special == "部隊艙" }
