package shell

// starlane.go:星圖移動的秒差距模型。
//
// 先前的 `SendFleet` 用的是 `ETA = ceil(正規化距離 × 8)` —— 一個沒有速度概念的固定換算。
// 那讓手冊裡**四條以「秒差距/回合」表述的規則全部無處可掛**(星雲、黑洞、Navigator、
// Warp Field Interdictor)。這一檔把星圖換成真的秒差距 + 航速。
//
// 數值全部在 `gamedata/starlane.go`,那裡標了每一個的出處(1 秒差距 = 30 單位、四檔銀河尺寸、
// 星數門檻都是反組譯真值;引擎速度是手冊逐條)。這裡只做接線。
//
// ============ ⚠ 一個結構性差異 ============
//
// 原版的星圖是「銀河座標 + 逐格路徑」,艦隊每回合前進 N 秒差距、**沿路**會經過別的星域,
// 所以「穿越星雲時降速」是逐段判定的。remake 的星圖是「兩點之間直接算 ETA」,沒有路徑。
//
// 於是這裡把「穿越」近似成**起點或終點在星雲內**。差別會出現在「兩端都在星雲外、
// 但直線穿過一團星雲」這種情況——原版會降速,remake 不會。要做到逐段判定,
// 得先有路徑模型,那是另一件事。**這個近似是明講的,不是沒想到。**

import (
	"math"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// GalaxySizeClass 回傳目前銀河的大小檔位(0..3)。
func (s *GameSession) GalaxySizeClass() int {
	return gamedata.GalaxySizeFromStars(len(s.Stars))
}

// ParsecsBetweenStars 回傳兩顆星之間的距離(整數秒差距,與原版
// `Parsecs_Between_Stars_` 同語意:無條件進位)。
//
// remake 的星座標是正規化 0..1,這裡先依銀河檔位的真實跨距換成秒差距再算。
func (s *GameSession) ParsecsBetweenStars(a, b int) int {
	if a < 0 || a >= len(s.Stars) || b < 0 || b >= len(s.Stars) {
		return 0
	}
	w, h := gamedata.GalaxyParsecSpan(s.GalaxySizeClass())
	dx := (s.Stars[a].X - s.Stars[b].X) * w
	dy := (s.Stars[a].Y - s.Stars[b].Y) * h
	return gamedata.ParsecsBetween(dx, dy)
}

// driveTechOwned 判定某引擎科技是否已擁有,規則與 `groundEquipTechOwned` 一致
// (主題完成、未明確抉擇 → 視為解鎖;已明確抉擇 → 需選中該科技)。
func driveTechOwned(ps engine.PlayerState, topic gamedata.ResearchTopic, tech gamedata.Technology) bool {
	return groundEquipTechOwned(ps, topic, tech)
}

// FleetDriveTier 回傳玩家已研究到的最高引擎階(0 = 沒有 FTL)。
func (s *GameSession) FleetDriveTier() int {
	return gamedata.DriveTierFromTechs(func(topic gamedata.ResearchTopic, tech gamedata.Technology) bool {
		return driveTechOwned(s.Player, topic, tech)
	})
}

// navigatorSkillLabel 是領航技能在 `Leader.Skill` 裡的中文標籤。
//
// 直接掃字串而不走 `leaderSkillIDByName`,是照 `commandoLeaderTier` 的既有作法:
// 那張表只服務「殖民地經濟被動加成」,把單位/語意都不同的加成混進同一張映射表會出事
// (見 session.go 該表上方的說明)。
const navigatorSkillLabel = "領航員"

// FleetHasNavigator 回傳艦隊裡是否有領航技能的軍官。
//
// 手冊:「Navigator: Allows a fleet to ignore the movement restrictions caused by nebulae
// and black holes. In addition, a navigator increases the speed of the fleet by
// 1 or 2 parsecs per turn.」
//
// ⚠ 只認**艦艇軍官**(`Ship == true`)——手冊那句是「a fleet … the ship contains an officer」,
// 殖民地領袖不隨艦隊走。
func (s *GameSession) FleetHasNavigator() bool {
	for _, l := range s.Leaders {
		if l.Ship && l.Skill == navigatorSkillLabel {
			return true
		}
	}
	return false
}

// FleetSpeedParsecs 回傳艦隊目前的每回合秒差距(還沒套星雲/干擾場的懲罰)。
//
// ⚠ Navigator 的「+1 或 +2」手冊沒寫判準(推測看技能等級)。remake 取 **+1**,
// 是保守下界不是真值——多給的那 1 點會直接讓遠征快一回合,寧可少不要多。
func (s *GameSession) FleetSpeedParsecs() int {
	tier := s.FleetDriveTier()
	if tier < 1 && s.FleetHasFTL() {
		// 有 FTL 但科技表上查不到任何引擎:一律當成**核融引擎**。
		// 這不是隨手補的下界——手冊講得很清楚「Nuclear Drive … is the slowest of the
		// faster than light (FTL) propulsion systems」,能超光速就至少有它。
		//
		// 會走到這條的是「一般/進階開局」:`FleetHasFTL` 對非曲速前開局直接回 true,
		// 不看科技表(見該函式)。少了這個下界,那些開局的航速會是 0 → ETA 全被夾成 1,
		// 整個秒差距模型形同虛設,而且**畫面上完全看不出來**。
		tier = 1
	}
	sp := gamedata.FleetSpeedForDrive(tier)
	if sp <= 0 {
		return 0
	}
	if s.FleetHasNavigator() {
		sp++
	}
	return sp
}

// fleetSpeedForTrip 回傳這一趟的實際航速,套上星雲懲罰。
//
// 星雲:手冊「reduced in speed to 1 parsec per turn」;Navigator 可無視(手冊同一段)。
// ⚠「穿越」近似成起點或終點在星雲內,見檔頭。
func (s *GameSession) fleetSpeedForTrip(from, to int) int {
	sp := s.FleetSpeedParsecs()
	if sp <= 0 {
		return 0
	}
	if s.FleetHasNavigator() {
		return sp
	}
	if s.StarInNebula(from) || s.StarInNebula(to) {
		return gamedata.NebulaSpeed
	}
	return sp
}

// FleetETATo 回傳從 from 飛到 to 需要幾回合(至少 1)。
//
// 沒有 FTL 回 0(呼叫端當成「去不了」)。
func (s *GameSession) FleetETATo(from, to int) int {
	sp := s.fleetSpeedForTrip(from, to)
	if sp <= 0 {
		return 0
	}
	pc := s.ParsecsBetweenStars(from, to)
	eta := int(math.Ceil(float64(pc) / float64(sp)))
	if eta < 1 {
		eta = 1
	}
	return eta
}
