package shell

import (
	"math/rand"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// artemis.go:**阿提米絲系統網**接進艦隊移動。
//
// 規格(觸發機率、水雷數、每枚傷害)全在 `gamedata/artemis.go`,這一檔只負責三件事:
// 把 remake 的中文艦體名對到手冊的六個 size class、把護盾名對到護盾等級、
// 在艦隊抵達敵方星系時擲骰並把傷害寫進 `Ship.Damage`。
//
// ============ 觸發時機 ============
//
// 手冊寫的是「any enemy ship **entering** that system」——是**進入**的那一刻,
// 不是停留每回合。所以掛在 `advanceFleet` 的抵達那一段,和一次性發現
// (`discoverSystemSpecials`)同一個位置。
//
// ============ 護盾等級從名字讀 ============
//
// remake 的護盾元件叫「第一級護盾」「第三級護盾」…「第十級護盾」——那個「級」
// 就是手冊說的 shield class,不需要另建對照表。用一張小表而不是解析字串:
// 解析中文數字比查表脆弱,而且護盾只有六種。
//
// ============ 誠實留白 ============
//
//   - 玩家與 AI 艦隊抵達都已接同一組水雷規則；AI 的艦艇資料仍是匯總戰力，故 AI 路徑
//     以無護盾承傷並換算戰力損失，不能冒稱逐艦 parity。
//   - 玩家艦隊被水雷炸沉的船直接從艦隊移除,不進戰鬥記錄——remake 沒有「戰場外損失」這個
//     事件類別,硬塞進戰鬥記錄會讓歷史紀錄出現一場沒有敵人的仗。

// artemisHullClass 把 remake 的中文艦體名對到手冊的六個 size class。
//
// 偵察艦 / 殖民船 走 Frigate:原版沒有這兩個艦體等級,那些都是蓋在 Frigate 艦體上的
// 設計(見 gamedata/artemis.go 的誠實留白)。
func artemisHullClass(class string) gamedata.ArtemisHullClass {
	switch class {
	case "驅逐艦":
		return gamedata.ArtemisDestroyer
	case "巡洋艦":
		return gamedata.ArtemisCruiser
	case "戰艦":
		return gamedata.ArtemisBattleship
	case "泰坦":
		return gamedata.ArtemisTitan
	case "末日之星":
		return gamedata.ArtemisDoomStar
	}
	return gamedata.ArtemisFrigate
}

// artemisShieldClass 把護盾元件名對到手冊的 shield class(名字裡的「級」)。
func artemisShieldClass(shield string) int {
	switch shield {
	case "第一級護盾":
		return 1
	case "第三級護盾":
		return 3
	case "第五級護盾":
		return 5
	case "第七級護盾":
		return 7
	case "第十級護盾":
		return 10
	}
	return 0 // 無護盾,或未知元件——不給折扣,不假裝有防護
}

// ArtemisStrike 是一次水雷網結算的結果(供 UI / 測試檢視)。
type ArtemisStrike struct {
	StarName string
	// ShipsHit 是踩到雷的船數,ShipsLost 是因此沉掉的船數。
	ShipsHit  int
	ShipsLost int
	// TotalDamage 是這次所有中招船受到的傷害總和。
	TotalDamage int
	// LostNames 是沉掉的船名(照原本在艦隊裡的順序)。
	LostNames []string
}

// starHasArtemisNet 回報某顆星上的敵方殖民地是否建了阿提米絲系統網。
func (s *GameSession) starHasArtemisNet(starIdx int) bool {
	aiIdx, colonyIdx, ok := s.findAIColonyByStar(starIdx)
	if !ok {
		return false
	}
	ai := &s.AIPlayers[aiIdx]
	if colonyIdx >= len(ai.ColonyBuildings) {
		return false
	}
	return builtMapHasOriginalBuildingID(ai.ColonyBuildings[colonyIdx], 3)
}

// applyArtemisMines 對剛抵達 starIdx 的艦隊 f 結算一次水雷網。
//
// 回傳 nil 代表這顆星沒有水雷網、或一艘船都沒踩到——呼叫端據此決定要不要提示玩家。
//
// 亂數用星系與回合當種子,與軌道轟炸同款:同一局重播會得到同樣的結果
// (網路對戰的決定性要求,見 determinism.go)。
func (s *GameSession) applyArtemisMines(f *Fleet, starIdx int) *ArtemisStrike {
	if f == nil || len(f.Ships) == 0 || !s.starHasArtemisNet(starIdx) {
		return nil
	}
	rng := rand.New(rand.NewSource(int64(s.Turn)*2246822519 + int64(starIdx)*7919 + 31337))
	out := &ArtemisStrike{StarName: s.starName(starIdx)}
	surv := f.Ships[:0:0] // 新底層陣列,不就地改寫還要讀的那一份
	for _, sh := range f.Ships {
		// ① 逐艦擲觸發:大船躲不掉。
		if rng.Intn(100) >= gamedata.ArtemisTriggerPercent(artemisHullClass(sh.Class)) {
			surv = append(surv, sh)
			continue
		}
		// ② 中招的船各擲一次水雷數(8–28)。
		mines := gamedata.ArtemisMineCount(rng.Intn(gamedata.ArtemisMineRollSpan + 1))
		// ③ 每枚 20 − 護盾等級。
		dmg := gamedata.ArtemisShipDamage(mines, artemisShieldClass(sh.Shield))
		out.ShipsHit++
		out.TotalDamage += dmg

		sh.Damage += dmg
		if sh.Damage >= shipMaxHP(sh) {
			out.ShipsLost++
			out.LostNames = append(out.LostNames, sh.Name)
			continue // 沉了,不進倖存清單
		}
		surv = append(surv, sh)
	}
	f.Ships = surv
	if out.ShipsHit == 0 {
		return nil
	}
	return out
}
