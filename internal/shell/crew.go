package shell

import "github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"

// crew.go:**艦員經驗**接進遊戲迴圈。
//
// 規格(四條加成軌、EP 門檻、統帥種族的平移)全在 `gamedata/crew.go`,這一檔負責三件事:
// 每回合累積、戰後加經驗、太空學院的兩個效果。
//
// ============ 等級不存,只存經驗 ============
//
// `Ship` 只加了一個 `CrewXP` 欄位,等級一律由 `shipCrewLevel` 現算。
// 存兩個欄位遲早會不同步(升級時忘了更新其中一個),存一個不會。
//
// 太空學院的「造出來的船起始等級 +1」因此也是用經驗表達:起始 `CrewXP` 直接設成
// 那一級的門檻(`gamedata.CrewXPForLevel`),而不是另開一個「起始等級」欄位。
//
// ============ 誠實留白 ============
//
//   - AI 的單一主力艦隊也走同一個每回合／太空學院／Instructor 累積；AI 尚未支援多艦隊，
//     所有實艦共用主力艦隊所在星。
//   - `ShipCrewBoardingBonus` 已接入快速與格子戰術登艦；攻守方各自使用所屬帝國的
//     種族、科技、Commando／Security 艦隊最大值。
//   - 手冊說經驗來自「turn **in space**」。remake 沒有「船在港內 vs 在太空」的區別
//     ——所有船都在艦隊裡、艦隊永遠在某顆星上。所以這裡的實作是「每回合每艘船 +1」,
//     與手冊在 remake 的模型下等價,但如果之後加了船塢/駐港狀態,這裡要跟著改。

// shipCrewLevel 回傳一艘船目前的艦員等級(由 CrewXP 推導)。
func (s *GameSession) shipCrewLevel(sh Ship) int {
	return gamedata.CrewLevelForXP(sh.CrewXP, s.RaceWarlord())
}

// ShipCrewLevel 是 shipCrewLevel 的公開版本,供 UI 顯示。
func (s *GameSession) ShipCrewLevel(sh Ship) int { return s.shipCrewLevel(sh) }

// ShipCrewLevelName 回傳艦員等級的中文名。
func ShipCrewLevelName(level int) string {
	switch level {
	case gamedata.CrewGreen:
		return "新兵"
	case gamedata.CrewRegular:
		return "正規兵"
	case gamedata.CrewVeteran:
		return "老兵"
	case gamedata.CrewElite:
		return "精銳"
	case gamedata.CrewUltraElite:
		return "超級精銳"
	}
	return "新兵"
}

// spaceAcademiesAt 回傳玩家在 starIdx 那顆星上有幾座太空學院。
//
// 手冊 p.97:「for every Space Academy present in that system」——是**每座**都算,
// 所以回的是數量而不是 bool。remake 目前一顆星最多一個玩家殖民地,所以實際上是 0 或 1;
// 回數量是為了不把「一顆星只能有一個殖民地」這個 remake 限制烘進規則裡。
func (s *GameSession) spaceAcademiesAt(starIdx int) int {
	n := 0
	for i := range s.PlayerColonies {
		if s.colonyStar(i) != starIdx {
			continue
		}
		if i < len(s.ColonyBuildings) && s.ColonyBuildings[i][spaceAcademyName] {
			n++
		}
	}
	return n
}

// spaceAcademyName 是建築表裡的中文名(gamedata.Buildings 的 NameZH)。
const spaceAcademyName = "太空學院"

// newShipCrewXP 回傳一艘剛在 colonyIdx 造好的船的起始經驗。
//
// 太空學院把起始等級推高一級(手冊 p.97),用「那一級的門檻經驗」表達。
// colonyIdx < 0 代表不是從某個殖民地造出來的(事件給的船之類),走一般起始等級。
func (s *GameSession) newShipCrewXP(colonyIdx int) int {
	level := gamedata.CrewStartingLevel(s.RaceWarlord())
	if colonyIdx >= 0 && colonyIdx < len(s.ColonyBuildings) && s.ColonyBuildings[colonyIdx][spaceAcademyName] {
		level += gamedata.SpaceAcademyStartingLevelBonus
	}
	if max := gamedata.CrewMaxLevel(s.RaceWarlord()); level > max {
		level = max
	}
	xp := gamedata.CrewXPForLevel(level, s.RaceWarlord())
	if xp < 0 {
		return 0
	}
	return xp
}

// advanceCrewExperience 每回合替所有船加經驗，最後依原版 sub_149D5 夾到 500。
//
// 手冊 p.121:每回合在太空 +1;p.97:艦隊所在星系每有一座太空學院再 +1。
func (s *GameSession) advanceCrewExperience() {
	// 教官(SKILL_INSTRUCTOR)是**帝國層**技能(手冊「all ship crews in your empire」),
	// 所以在迴圈外算一次,不分艦隊。
	instructor := leaderInstructorXPBonus(s.Leaders)
	for i := range s.Fleets {
		f := &s.Fleets[i]
		gain := gamedata.CrewXPPerTurnInSpace + instructor
		// 航行中的艦隊不在任何星系上,拿不到學院加成——AtStar 只在停泊時有意義。
		if f.ETA <= 0 && f.AtStar >= 0 {
			gain += s.spaceAcademiesAt(f.AtStar) * gamedata.SpaceAcademyXPPerTurn
		}
		for j := range f.Ships {
			f.Ships[j].CrewXP = gamedata.CrewXPAfterTurnGain(f.Ships[j].CrewXP, gain)
		}
	}
	for i := range s.AIPlayers {
		a := &s.AIPlayers[i]
		gain := gamedata.CrewXPPerTurnInSpace + leaderInstructorXPBonus(a.Leaders)
		if a.FleetETA <= 0 {
			star := aiFleetStar(*a)
			for ci, colonyStar := range a.ColonyStars {
				if colonyStar == star && ci < len(a.ColonyBuildings) && a.ColonyBuildings[ci][spaceAcademyName] {
					gain += gamedata.SpaceAcademyXPPerTurn
				}
			}
		}
		for j := range a.Ships {
			a.Ships[j].CrewXP = gamedata.CrewXPAfterTurnGain(a.Ships[j].CrewXP, gain)
		}
	}
}

// awardBattleCrewXP 把一場勝仗的經驗發給指定的倖存參戰艦。
//
// destroyedHullClassSum 是被摧毀敵艦的 1-based 艦體級總和；eligible 使用 Fleet.Ships
// 當下索引。sub_4B184 只寫 winner-side 且仍連到持久 Ship record 的 battle record，
// 因此不能把沒有參戰的殖民船／前哨船也一起升級。
func (s *GameSession) awardBattleCrewXP(destroyedHullClassSum int, eligible map[int]bool) int {
	if len(eligible) == 0 {
		return 0
	}
	xp := gamedata.CrewBattleXPFromDestroyedHullClassSum(destroyedHullClassSum)
	f := s.Fleet()
	for j := range f.Ships {
		if eligible[j] {
			f.Ships[j].CrewXP += xp
		}
	}
	return xp
}

func hullClassSum(classes []int) int {
	sum := 0
	for _, class := range classes {
		if class > 0 {
			sum += class
		}
	}
	return sum
}

// shipSizeClassFromStrength 把 remake 的戰力值換回手冊的艦體等級(1–6)。
//
// remake 的 `shipStrength` 是 2 的冪:巡防 2、驅逐 4、巡洋 8、戰艦 16、泰坦 32、末日 64
// ——正好對上手冊的 size class 1..6。偵察艦/殖民船的 1 不是艦體等級(見 artemis.go 的
// 同款說明),回 1(Frigate)。
func shipSizeClassFromStrength(st int) int {
	switch {
	case st >= 64:
		return 6
	case st >= 32:
		return 5
	case st >= 16:
		return 4
	case st >= 8:
		return 3
	case st >= 4:
		return 2
	}
	return 1
}

// destroyedEnemySizeClasses 從「開打前的敵艦戰力清單」與「結束時倖存的 combatant」
// 還原出被擊沉的那些的艦體等級(1–6)。
//
// 為什麼要還原而不是邊打邊記:`battleVolley` 就地把陣亡者從切片移除,呼叫端拿不到
// 「誰死了」。而敵艦的 `atk` 就是它的戰力值、戰鬥中不變,所以用多重集合相減就還原得出來
// ——不必為了這件事去改戰鬥迴圈的介面。
func destroyedEnemySizeClasses(startStrengths []int, survivors []combatant) []int {
	alive := map[int]int{}
	for _, c := range survivors {
		alive[c.atk]++
	}
	var out []int
	for _, st := range startStrengths {
		if alive[st] > 0 {
			alive[st]--
			continue
		}
		out = append(out, shipSizeClassFromStrength(st))
	}
	return out
}

// helmsmanEvasionBonus 回傳艦上軍官的舵手技能貢獻的飛彈閃避加成。
//
// 手冊:「Half bonus of the Helmsman value」——所以是技能值的一半
// (`gamedata.MissileHelmsmanEvasionBonus`)。
//
// ⚠ 只算**艦艇軍官**(`l.Ship`):舵手是開船的,殖民地領袖不會坐在艦橋上。
// 這與 `starlane.go` 挑領航員(Navigator)是同一條規則,那邊也只看 l.Ship。
//
// 多位軍官取**最佳的那一位**,不加總——手冊只在 Megawealth/Researcher 明說累加,
// 其餘一律取最佳(見 gamedata/leader_skill_apply.go)。
func (s *GameSession) helmsmanEvasionBonus() int {
	best := 0
	for _, l := range s.Leaders {
		if !l.Ship {
			continue
		}
		tier := leaderSkillTier(l, int(gamedata.SKILL_HELMSMAN))
		if tier <= 0 {
			continue
		}
		v := gamedata.LeaderSkillBonus(int(gamedata.SKILL_HELMSMAN), tier,
			leaderDisplayLevelToExpLevel(l.Level))
		if ev := gamedata.MissileHelmsmanEvasionBonus(v); ev > best {
			best = ev
		}
	}
	return best
}

// FleetCrewSummary 回傳艦隊的艦員狀態摘要,供 UI 顯示。
//
// ⚠ 這是**艦隊層級**的摘要,不是逐艦資料:remake 目前沒有逐艦資訊面板,
// 而「一個只會默默上升的等級對玩家等於不存在」(`gamedata.CrewXPToNextLevel` 的檔頭)。
// 取**最低**那一艘當代表——艦隊的戰力由最弱的那條線決定,報最高的會讓玩家高估自己。
//
// ok=false 表示艦隊沒有任何參戰艦(支援艦不算,理由同 mkPlayerCombatantsIndexed)。
func (s *GameSession) FleetCrewSummary() (level int, toNext int, ok bool) {
	level = -1
	for _, sh := range s.Fleet().Ships {
		if isSupportShipClass(sh.Class) {
			continue
		}
		lv := s.shipCrewLevel(sh)
		if level < 0 || lv < level {
			level, toNext, ok = lv, gamedata.CrewXPToNextLevel(sh.CrewXP, s.RaceWarlord()), true
		}
	}
	if !ok {
		return 0, 0, false
	}
	return level, toNext, true
}
