package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/ai"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// ai_attack_test.go:AI 突襲的門檻、目標選擇與損失界限。
//
// 這些行為分兩類,測試裡分開標明:
//   - 目標估值用的是原版公式(gamedata.AIColonyValue,移植自 Colony_Worth_To_Player_)
//   - 「何時打、打贏怎樣」是 remake 的模型,測的是**界限**(不會滅殖民地、BC 不為負),
//     不是「數字等於原版」——原版那段還沒反編。

// newRaidTestSession 造一個「AI 已宣戰、軍力壓倒、時間也到了」的 session。
func newRaidTestSession(t *testing.T) *GameSession {
	t.Helper()
	s := NewDemoSession()
	s.Turn = aiRaidGraceTurns + 1
	s.Fleet().Ships = nil // 玩家無艦隊 → playerMilitary()=0,隔離軍力門檻這個變數
	for i := range s.AIPlayers {
		s.AIPlayers[i].StanceName = stanceNames[ai.StanceWar]
		s.AIPlayers[i].FleetStrength = 400
		s.AIPlayers[i].Personality = ai.PersonalityRuthless // 反應強度 100 → 必定動手
		s.AIPlayers[i].LastRaidTurn = 0
	}
	return s
}

// parkAIFleetsAtPlayerColony 把所有 AI 艦隊直接放到玩家殖民地 0 的上空(靜止)。
//
// 2026-08-08 第 48 項之後 AI 有位置了,突襲的前提是「艦隊**抵達**目標」而不是「想打」。
// 只驗結算後果(損失界限/擊退/間隔)的測試用這個把航程跳過去,
// 免得每一支都要先跑十幾回合等艦隊飛到。
func parkAIFleetsAtPlayerColony(s *GameSession) {
	star := s.PlayerColonyStarIndex(0)
	for i := range s.AIPlayers {
		s.AIPlayers[i].FleetStar, s.AIPlayers[i].FleetPosSet = star, true
		s.AIPlayers[i].FleetDestStar, s.AIPlayers[i].FleetETA = -1, 0
	}
}

// aiFleetLaunched 回報有沒有任何 AI 這回合派出了艦隊(出發那一刻的守門是否放行)。
func aiFleetLaunched(s *GameSession) bool {
	for i := range s.AIPlayers {
		if s.AIPlayers[i].FleetETA > 0 {
			return true
		}
	}
	return false
}

// ⚠ 這四個「守門」測試在第 48 項之後驗的是**出發那一端**(aiLaunchRaidFleet),
// 不是結算端。理由:守門條件跟著艦隊模型搬到了出發時刻,若還是只呼叫 advanceAIRaids,
// 這幾支會因為「艦隊本來就不在目標上」而**假綠**——測不到任何東西卻一路通過。
func TestAIRaidGracePeriod(t *testing.T) {
	s := newRaidTestSession(t)
	s.Turn = aiRaidGraceTurns - 1
	s.advanceAIFleets()
	if aiFleetLaunched(s) {
		t.Error("寬限期內不該派出艦隊")
	}
	s.advanceAIRaids()
	if s.LastRaid != "" {
		t.Errorf("寬限期內不該突襲,卻發生了:%s", s.LastRaid)
	}
}

func TestAIRaidNeedsWarStance(t *testing.T) {
	s := newRaidTestSession(t)
	for i := range s.AIPlayers {
		s.AIPlayers[i].StanceName = "中立"
	}
	s.advanceAIFleets()
	if aiFleetLaunched(s) {
		t.Error("非戰爭態勢不該派出艦隊")
	}
	s.advanceAIRaids()
	if s.LastRaid != "" {
		t.Errorf("非戰爭態勢不該突襲,卻發生了:%s", s.LastRaid)
	}
}

// 軍力門檻:玩家艦隊夠強時 AI 不敢動手(aiRaidStrengthMargin)。
func TestAIRaidNeedsStrengthAdvantage(t *testing.T) {
	s := newRaidTestSession(t)
	for i := range s.AIPlayers {
		s.AIPlayers[i].FleetStrength = 10
	}
	// 玩家軍力遠超 AI。
	s.Fleet().Ships = []Ship{{Class: "巡洋艦"}, {Class: "巡洋艦"}, {Class: "巡洋艦"}}
	if pm := s.playerMilitary(); pm == 0 {
		t.Fatal("測試前提不成立:玩家軍力應 > 0")
	}
	s.advanceAIFleets()
	if aiFleetLaunched(s) {
		t.Error("AI 軍力不足不該派出艦隊")
	}
	s.advanceAIRaids()
	if s.LastRaid != "" {
		t.Errorf("AI 軍力不足不該突襲,卻發生了:%s", s.LastRaid)
	}
}

// 和平主義的反應強度是 0(原版 _personality_losing_ground_chance 第 6 欄),從不主動突襲。
func TestAIRaidPacifistNeverRaids(t *testing.T) {
	s := newRaidTestSession(t)
	for i := range s.AIPlayers {
		s.AIPlayers[i].Personality = ai.PersonalityPacifist
	}
	for turn := 0; turn < 30; turn++ {
		s.Turn = aiRaidGraceTurns + turn
		s.advanceAIFleets()
		if aiFleetLaunched(s) {
			t.Fatalf("和平主義 AI(反應強度 %d)不該派出艦隊",
				ai.PersonalityLosingGroundChance(ai.PersonalityPacifist))
		}
		s.advanceAIRaids()
		if s.LastRaid != "" {
			t.Fatalf("和平主義 AI 不該突襲(反應強度 %d),卻發生了:%s",
				ai.PersonalityLosingGroundChance(ai.PersonalityPacifist), s.LastRaid)
		}
	}
}

func TestAIRaidHappensAndBoundsLosses(t *testing.T) {
	s := newRaidTestSession(t)
	popBefore := s.PlayerColonies[0].Population
	bcBefore := s.Player.BC

	parkAIFleetsAtPlayerColony(s)
	s.advanceAIRaids()
	if s.LastRaid == "" || s.LastRaidReport == nil {
		t.Fatal("條件全部滿足時應該發生突襲")
	}
	rep := s.LastRaidReport
	if rep.Repelled {
		t.Fatalf("玩家無艦隊、無駐軍時不該擊退 400 戰力的突襲:%s", rep.Message)
	}
	c := s.PlayerColonies[rep.ColonyIdx]
	// 界限一:突襲不滅殖民地(人口至少留 1)。
	if c.Population < 1 {
		t.Errorf("突襲把殖民地人口打到 %d,應至少留 1", c.Population)
	}
	// 界限二:人口損失不超過 3。
	if rep.PopLost > 3 {
		t.Errorf("單次突襲人口損失 %d,上限應為 3", rep.PopLost)
	}
	if popBefore-c.Population != rep.PopLost {
		t.Errorf("報告的人口損失 %d 與實際 %d 不符", rep.PopLost, popBefore-c.Population)
	}
	// 界限三:國庫不為負。
	if s.Player.BC < 0 {
		t.Errorf("突襲把國庫打成負數 %d", s.Player.BC)
	}
	if bcBefore-s.Player.BC != rep.BCLost {
		t.Errorf("報告的 BC 損失 %d 與實際 %d 不符", rep.BCLost, bcBefore-s.Player.BC)
	}
	// 職業分配總和必須與人口一致,否則之後的經濟結算會錯亂。
	if c.Farmers+c.Workers+c.Scientists != c.Population {
		t.Errorf("職業分配 %d+%d+%d 與人口 %d 對不上",
			c.Farmers, c.Workers, c.Scientists, c.Population)
	}
}

// 玩家艦隊停泊在該星時應能擊退突襲,且 AI 會損失戰力(讓「把艦隊擺對地方」有意義)。
func TestAIRaidRepelledByFleetAtStar(t *testing.T) {
	s := newRaidTestSession(t)
	// 玩家艦隊停在母星(殖民地 0 所在星)。
	s.Fleet().AtStar = s.PlayerColonyStarIndex(0)
	s.Fleet().ETA = 0
	s.Fleet().Ships = []Ship{{Class: "巡洋艦"}, {Class: "巡洋艦"}}

	// 母星升級成戰鬥站。
	//
	// ⚠ 2026-08-07 加的,理由值得寫下來:`colonyDefense` 的軌道防禦先前是
	// `CommandPointsFromBuildings × 10`(自編係數),星基因此值 10 —— 比一艘巡洋艦(8)還強。
	// 那與 `gamedata/satellite.go` 的校準**自相矛盾**:那份校準明講星基 ≈ 驅逐艦 tier(4)。
	// 改用同一套 space 預算推導之後星基只值 3,而 AI 的願打門檻是玩家艦隊的 125%——
	// 兩艘巡洋艦(16)配一座光禿禿的星基,防禦 19 < 門檻 21,**AI 會贏,而那是正確的結果**。
	//
	// 這個測試守的是「把艦隊擺對地方有意義」,不是「星基很強」。所以把母星升級成真的有投資
	// 防禦的樣子(戰鬥站 space 500,是星基的兩倍),而不是把模型改回去遷就測試。
	s.ColonyBuildings[0]["戰鬥站"] = true

	// AI 軍力要同時滿足兩個條件才測得到「擊退」:①過得了 aiRaidStrengthMargin 門檻
	// (否則它根本不動手)②低於殖民地防禦(否則會突破)。艦艇戰力表日後若調整,
	// 這兩個界限會自動跟著走,不用回來改硬編數字。
	pm, def := s.playerMilitary(), s.colonyDefense(0)
	minWilling := pm*aiRaidStrengthMargin/100 + 1
	if minWilling > def {
		t.Fatalf("測試前提不成立:願打門檻 %d 已高於母星防禦 %d,擊退情境不可能出現", minWilling, def)
	}
	for i := range s.AIPlayers {
		s.AIPlayers[i].FleetStrength = minWilling
	}
	strengthBefore := s.AIPlayers[0].FleetStrength

	parkAIFleetsAtPlayerColony(s)
	s.advanceAIRaids()
	if s.LastRaidReport == nil {
		t.Fatal("應該發生一次(被擊退的)突襲")
	}
	if !s.LastRaidReport.Repelled {
		t.Fatalf("防禦 %d >= 攻擊 %d 應擊退,卻被突破:%s",
			s.colonyDefense(0), strengthBefore, s.LastRaidReport.Message)
	}
	if s.AIPlayers[0].FleetStrength >= strengthBefore {
		t.Errorf("擊退後 AI 軍力應下降:%d → %d", strengthBefore, s.AIPlayers[0].FleetStrength)
	}
	if s.PlayerColonies[0].Population != NewDemoSession().PlayerColonies[0].Population {
		t.Error("被擊退的突襲不該造成人口損失")
	}
}

// 突襲間隔:同一個 AI 不能連續回合狂打。
func TestAIRaidRespectsInterval(t *testing.T) {
	s := newRaidTestSession(t)
	parkAIFleetsAtPlayerColony(s)
	s.advanceAIRaids()
	if s.LastRaid == "" {
		t.Fatal("第一次應該要打")
	}
	raider := -1
	for i := range s.AIPlayers {
		if s.AIPlayers[i].LastRaidTurn == s.Turn {
			raider = i
			break
		}
	}
	if raider < 0 {
		t.Fatal("找不到發動突襲的 AI")
	}
	// 只留這一個 AI,下一回合它必須因間隔而不動手。
	s.AIPlayers = s.AIPlayers[raider : raider+1]
	s.Turn++
	s.advanceAIRaids()
	if s.LastRaid != "" {
		t.Errorf("間隔 %d 回合內不該再打,卻發生了:%s", aiRaidInterval, s.LastRaid)
	}
}

// 目標選擇:同樣距離下,產出高的殖民地價值較高(原版 Colony_Worth_To_Player_ 的核心語意)。
func TestAIColonyValueRanksByProduction(t *testing.T) {
	base := gamedata.AIColonyValueInput{
		Population: 8, MaxPop: 12, Food: 8, Industry: 6, Research: 6,
		Climate: gamedata.TERRAN, Gravity: gamedata.NORMAL_G,
	}
	rich := base
	rich.Industry, rich.Research = 20, 20

	obj := gamedata.AIObjectiveBalancedLow
	if gamedata.AIColonyValue(rich, obj) <= gamedata.AIColonyValue(base, obj) {
		t.Error("產出高的殖民地價值應該較高")
	}
	// 前哨站沒有產出,原版直接跳過主公式。
	outpost := rich
	outpost.IsOutpost = true
	if got := gamedata.AIColonyValue(outpost, obj); got != 0 {
		t.Errorf("前哨站價值應為 0,實得 %d", got)
	}
	// 寶石礦加分是金礦的兩倍(原版 2000 vs 1000)。加分在原版是**加完才 >>6**,
	// 整數右移會吃掉零頭(1000>>6=15、2000>>6=31),所以最終差值只能要求「接近兩倍」,
	// 不能要求剛好——這個誤差是原版公式本身帶來的,不是 remake 算錯。
	gold, gem := base, base
	gold.Special, gem.Special = int(gamedata.GoldDeposits), int(gamedata.GemDeposits)
	dGold := gamedata.AIColonyValue(gold, obj) - gamedata.AIColonyValue(base, obj)
	dGem := gamedata.AIColonyValue(gem, obj) - gamedata.AIColonyValue(base, obj)
	if dGold <= 0 || dGem < 2*dGold-2 || dGem > 2*dGold+2 {
		t.Errorf("金礦加分 %d、寶石礦加分 %d,應接近 1:2 且皆 > 0", dGold, dGem)
	}
	// 重力懲罰:高重力星對一般種族價值明顯較低(原版 ×6/12)。
	heavy := base
	heavy.Gravity = gamedata.HEAVY_G
	if gamedata.AIColonyValue(heavy, obj) >= gamedata.AIColonyValue(base, obj) {
		t.Error("高重力星對無天賦種族的價值應較低")
	}
}

// 人口越接近上限,價值越低(原版 `avgMaxPop + 100 - population` 那一項的語意)。
func TestAIColonyValueGrowthRoom(t *testing.T) {
	obj := gamedata.AIObjectiveBalancedLow
	roomy := gamedata.AIColonyValueInput{
		Population: 4, MaxPop: 20, Food: 8, Industry: 6, Research: 6,
		Climate: gamedata.TERRAN, Gravity: gamedata.NORMAL_G,
	}
	full := roomy
	full.Population = 20
	// 人口多會讓 w0*population 那一項變大,但成長空間那一項變小;
	// 這裡只確認公式真的把「成長空間」算進去了(兩者不相等)。
	if gamedata.AIColonyValue(roomy, obj) == gamedata.AIColonyValue(full, obj) {
		t.Error("成長空間應該影響估值")
	}
}

// TestAIRaidsHappenInRealGame 是「這個子系統在真實對局裡真的會發生嗎」的探針。
//
// 加這一條的理由:突襲的門檻疊了四層(寬限回合、戰爭態勢、軍力領先、性格機率),
// 任何一層設太嚴,整個子系統就變成永遠不觸發的死碼,而上面那些單元測試——因為都把
// 前提直接擺好——一條都抓不到。這裡不擺前提,照 NewDemoSession 跑滿 300 回合。
func TestAIRaidsHappenInRealGame(t *testing.T) {
	s := NewDemoSession()
	raids, repelled := 0, 0
	firstTurn := -1
	for i := 0; i < 300; i++ {
		s.EndTurn()
		if s.LastRaidReport == nil {
			continue
		}
		raids++
		if firstTurn < 0 {
			firstTurn = s.Turn
		}
		if s.LastRaidReport.Repelled {
			repelled++
		}
		// 每一次突襲都不能破壞不變式(單元測試只驗了一次,這裡驗全部)。
		for j, c := range s.PlayerColonies {
			if c.Population < 1 {
				t.Fatalf("第 %d 回合突襲後殖民地 %d 人口 <1", s.Turn, j)
			}
			if c.Farmers+c.Workers+c.Scientists != c.Population {
				t.Fatalf("第 %d 回合突襲後殖民地 %d 職業分配與人口不符", s.Turn, j)
			}
		}
	}
	if raids == 0 {
		t.Fatal("300 回合內一次突襲都沒發生——門檻疊太嚴,子系統等於死碼")
	}
	t.Logf("300 回合共 %d 次突襲(其中 %d 次被擊退),最早發生在第 %d 回合,結束 BC=%d",
		raids, repelled, firstTurn, s.Player.BC)
}
