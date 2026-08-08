package shell

import (
	"fmt"
	"math/rand"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
)

// 安塔蘭勝利(手冊三條勝利路徑之二)的 shell 層整合。權威規則來源見
// internal/engine/victory.go 的 VictoryAntaran 手冊逐字引用,以及
// docs/tech/victory-conditions.md 第 6 節先前記載的 TODO(本檔把它從「完全沒有對應流程」補上
// 「已接:傳送門 + 反攻」)。
//
// 手冊原文(GAME_MANUAL.pdf p.183,「Winning」小節):
//
//	"An alternate method is to seek out and defeat the Antaran home fleet. This involves
//	travelling to the Antaran homeworld, which is not possible until you have the right
//	technology and build a Dimensional Gate. Once you defeat the awe-inspiring Antarans, all
//	the other races in the galaxy recognise your overwhelming superiority and quickly
//	capitulate. (This strategy is not available if you disabled Antaran Attacks when setting up
//	your game.)"
//
// 次元傳送門(Dimensional Portal)建築本身已存在於 gamedata.Buildings(手冊 p.106,前置科技
// TOPIC_MULTIDIMENSIONAL_PHYSICS,見 docs/tech/colony-buildings.md 第三節)——本檔只補「建成後
// 解鎖反攻」這段流程,不重新定義建築資料。

// antaranHomeFleetDefense 是安塔蘭母星防禦艦隊的戰力組成。
//
// ============ 2026-08-08(第 49 項(安塔蘭防禦艦隊)):從「保守預設」換成反組譯真值 ============
//
// 這裡原本是 `{64, 64, 64, 64, 64, 64}`——六艘同級,註解寫著「手冊/openorion2 皆無精確數字,
// remake 保守預設,待考證」。**兩件事都被推翻了:數字有,而且組成根本不是同級的。**
//
// `Load_Antaran_Defense_Fleet_` @ 0x4D141(整支只有 77 bytes):
//
//	if word_199182 < 1: word_199182 = 1     ; 第 5 格(最大艦)強制至少 1
//	for bx = 0..4:                           ; ★ 五種艦體尺寸
//	    while cx < word_19917A[bx*2]:        ; ★ 每種尺寸的數量
//	        Load_Combat_Antaran_Ship_(...)
//	Load_Antaran_Star_Fortress_()            ; ★ 外加一座星際要塞
//
// 數量的上限來自靜態表 `_n_max_antaran_def_ships`(`byte_181746` @ 0x181746),
// 逐位元組解出來是 **{0, 0, 3, 2, 7, 0, 0, 0, 0}**:
//
//	| 索引 | 艦體尺寸 | 上限 |
//	|---|---|---|
//	| 0 | Small(Raider) | **0——永遠不造** |
//	| 1 | Medium(Marauder) | **0——永遠不造** |
//	| 2 | Large(Intruder) | 3 |
//	| 3 | Huge(Interdictor) | 2 |
//	| 4 | Titan(Harbinger) | 7 |
//
// 合計 **12 艘 + 1 座星際要塞**。難度**不改變組成**,只改變累積速度與逐艦裝甲加成。
//
// ⚠ **艦體尺寸的對照是推論,不是查表。** 原版的五級(Small/Medium/Large/Huge/Titan)
// 與 remake 的六級戰力階梯(shipStrength)沒有現成的對照表。這裡取「相對順序保持不變、
// 最大的安塔蘭艦對到 remake 最頂端」:Large→戰艦(16)、Huge→泰坦(32)、Titan→末日之星(64)。
// 合計 3×16 + 2×32 + 7×64 = 560。**這一層是 remake 的映射選擇**,數量與分層是真值。
//
// ⚠ **這是「養滿之後」的上限。** 原版的數量是執行期累積的(`word_19917A` 是 BSS,
// 載入時為 0),`Build_Antaran_Defensive_Ships_` 逐步補到上限;開局就打過去理論上只會遇到
// 保底的 1 艘 Harbinger + 要塞。remake 沒有安塔蘭的資源累積模型,取**上限**
// ——那是「終局一戰」該有的樣子,也與先前那個固定值的用意一致。
//
// 詳細推導見 `docs/re/antaran-defense-fleet.md`。
var antaranHomeFleetDefense = buildAntaranHomeFleetDefense()

// 安塔蘭母星防禦艦隊各尺寸的上限(原版 `_n_max_antaran_def_ships`,逐位元組解出)。
const (
	antaranDefLargeCount = 3 // Intruder
	antaranDefHugeCount  = 2 // Interdictor
	antaranDefTitanCount = 7 // Harbinger
)

// buildAntaranHomeFleetDefense 依上表組出防禦方的逐艦戰力,外加那座星際要塞。
//
// 星際要塞用 remake 既有的軌道防禦換算(`gamedata.StarFortressSpace`,與軌道轟炸那條
// 共用同一把尺,不另立係數)——原版是 `Load_Antaran_Star_Fortress_` @ 0x4D18E 載入的
// 一整套設計,remake 沒有那個粒度。
func buildAntaranHomeFleetDefense() []int {
	out := make([]int, 0, antaranDefLargeCount+antaranDefHugeCount+antaranDefTitanCount+1)
	for i := 0; i < antaranDefLargeCount; i++ {
		out = append(out, shipStrength("戰艦"))
	}
	for i := 0; i < antaranDefHugeCount; i++ {
		out = append(out, shipStrength("泰坦"))
	}
	for i := 0; i < antaranDefTitanCount; i++ {
		out = append(out, shipStrength("末日之星"))
	}
	// 星際要塞:remake 沒有「安塔蘭要塞」的設計資料,取**一艘末日之星的等量戰力**當代理。
	//
	// ⚠ 這是 remake 的代理值,不是原版數字。理由:`gamedata.StarFortressSpace = 1200`
	// 那個常數自己的註解就寫著「【近似】比照 ShipHullSpace(Doom Star)」——
	// 既然 remake 對星際要塞的既有建模就是「比照末日之星」,這裡沿用同一個近似,
	// 不另立第二套換算(兩套近似會漂開)。
	out = append(out, shipStrength("末日之星"))
	return out
}

// dimensionalPortalBuildingName 是次元傳送門在 s.ColonyBuildings 去重 map 裡的 key
// (gamedata.Buildings 裡該項的 NameZH,見 buildings.go)。
const dimensionalPortalBuildingName = "次元傳送門"

// hasDimensionalPortal 判定玩家是否已在任一殖民地建成次元傳送門。
//
// remake 簡化(已誠實記錄於 docs/tech/victory-conditions.md):手冊原文是「select a fleet in the
// same system as the portal」——反攻前置理論上要求「艦隊與傳送門同星系」,但本 remake 的星際
// 航行模型(FleetAtStar/FleetDestStar)不追蹤「殖民地建築在哪個星系」與「艦隊目前在哪個星系」
// 兩者的可達性交叉比對這麼細,故簡化為「玩家帝國內任一殖民地已建成即視為前置滿足」。
func (s *GameSession) hasDimensionalPortal() bool {
	for _, built := range s.ColonyBuildings {
		if built != nil && built[dimensionalPortalBuildingName] {
			return true
		}
	}
	return false
}

// CanAssaultAntares 回傳「玩家現在是否能發起反攻安塔蘭母星」——供 UI 決定是否顯示/啟用按鈕,
// 也是 AssaultAntares 內部檢查的匯出版本(判斷邏輯只寫一份,兩處共用)。
func (s *GameSession) CanAssaultAntares() bool {
	return !s.Victory.Over && !s.DisableEvents && s.hasDimensionalPortal() && len(s.Fleet().Ships) > 0
}

// AssaultAntares 解算「反攻安塔蘭母星」戰鬥(手冊三條勝利路徑之二)。
//
// 前置條件(CanAssaultAntares,不滿足則 ok=false,不消耗艦隊、不觸發戰鬥、不寫 LastBattle):
//   - 遊戲尚未結束(s.Victory.Over)。
//   - 手冊「This strategy is not available if you disabled Antaran Attacks when setting up
//     your game」→ s.DisableEvents 視為關閉安塔蘭攻擊時一併關閉本反攻路徑。
//   - 已建次元傳送門(hasDimensionalPortal)。
//   - 玩家艦隊非空(手冊「select a fleet in the same system as the portal」的最小化對應)。
//
// 戰鬥沿用 ResolveBattle 同款 battleVolley 逐回合解算(每回合雙方齊射,最多 6 回合),防禦方
// 戰力用 antaranHomeFleetDefense(反組譯真值的組成,見該變數註解)。與 ResolveBattle 不同:這是「終局一戰」,PlayerWon
// 要求防禦方**全滅**(len(defenders)==0),不是 ResolveBattle 那種「艦數比較多也算贏」的寬鬆
// 判定——手冊原文「Once you defeat the awe-inspiring Antarans」語意是徹底擊敗,不是打退。
//
//   - 玩家戰勝(防禦方全滅)→ s.AntaranHomeworldConquered=true,下一次 advanceAntaranVictory
//     (EndTurn 呼叫)會偵測並設定 s.Victory=VictoryAntaran。
//   - 玩家戰敗(6 回合內未能全殲防禦方,或己方全滅)→ 不設定勝利旗標,套用艦隊損失
//     (比照 ResolveBattle 呼叫 removeWeakestShip),回傳戰鬥結果供 UI/回合摘要顯示。
//
// 回傳 (BattleResult, ok)——ok=false 代表前置條件不滿足,BattleResult 為零值,呼叫端不應顯示
// 戰鬥結果畫面。
func (s *GameSession) AssaultAntares() (BattleResult, bool) {
	if !s.CanAssaultAntares() {
		return BattleResult{}, false
	}

	mkPlayer := func() []combatant {
		var out []combatant
		for _, sh := range s.Fleet().Ships {
			body := shipStrength(sh.Class)
			atk := body + sh.WeaponAttack
			atk += atk * s.RaceCombatPct / 100 // 種族戰鬥加成,比照 ResolveBattle mkPlayer
			out = append(out, combatant{hp: body * 3, atk: atk, def: body, wmin: atk / 2, wmax: atk,
				shield: shieldReduceByName(sh.Shield), armor: armorHPByName(sh.Armor),
				kind: weaponKindByName(sh.Weapon)})
		}
		return out
	}
	var df []combatant
	for _, st := range antaranHomeFleetDefense {
		df = append(df, combatant{hp: st * 3, atk: st, def: st, wmin: st / 2, wmax: st, armor: st})
	}
	pf := mkPlayer()

	res := BattleResult{Enemy: "安塔蘭母星防禦艦隊", PlayerStart: len(pf), EnemyStart: len(df)}
	// 種子與 ResolveBattle 用不同 offset(987654321 vs 12345),避免同一回合呼叫兩個戰鬥函式時
	// 巧合共用同一亂數序列。
	rng := rand.New(rand.NewSource(int64(s.Turn)*2654435761 + 987654321))
	for round := 1; round <= 6 && len(pf) > 0 && len(df) > 0; round++ {
		eDestroyed := battleVolley(pf, &df, rng)
		pDestroyed := battleVolley(df, &pf, rng)
		res.Log = append(res.Log, fmt.Sprintf("第 %d 回合:擊沉安塔蘭艦 %d ／ 我方損失 %d", round, eDestroyed, pDestroyed))
	}
	res.PlayerLosses = res.PlayerStart - len(pf)
	res.EnemyLosses = res.EnemyStart - len(df)
	res.PlayerWon = len(df) == 0 // 終局一戰:要求防禦方全滅,見函式註解(比 ResolveBattle 嚴格)
	for i := 0; i < res.PlayerLosses; i++ {
		s.removeWeakestShip()
	}
	s.LastBattle = &res
	if res.PlayerWon {
		s.AntaranHomeworldConquered = true
	}
	return res, true
}

// advanceAntaranVictory 是 EndTurn 每回合呼叫的狀態機:偵測 AssaultAntares 是否已戰勝
// (AntaranHomeworldConquered),沿用 engine.CheckAntaranVictory 純函式判定(不重算邏輯),只做
// shell 層「設 Victory 狀態」的整合,比照 advanceConquestVictory/advanceCouncil 同款模式。
func (s *GameSession) advanceAntaranVictory() {
	if s.Victory.Over {
		return
	}
	if !engine.CheckAntaranVictory(s.AntaranHomeworldConquered) {
		return
	}
	s.Victory = VictoryState{Over: true, Reason: engine.VictoryAntaran, Winner: "player", Turn: s.Turn}
}

// AntaranDefenseStrength 回傳安塔蘭母星防禦艦隊的總戰力(供安塔蘭房間畫面顯示戰力對比)。
// 值來自 antaranHomeFleetDefense——**組成(3 Large / 2 Huge / 7 Titan + 1 要塞)是反組譯真值**,
// 艦體尺寸對到 remake 戰力階梯那一層是推論,
// 理由見該變數的註解。
func AntaranDefenseStrength() int {
	n := 0
	for _, st := range antaranHomeFleetDefense {
		n += st
	}
	return n
}

// AntaranDefenseShipCount 回傳安塔蘭母星防禦艦隊的艦數(同上,供畫面顯示)。
func AntaranDefenseShipCount() int { return len(antaranHomeFleetDefense) }

// PlayerFleetStrength 回傳玩家艦隊總戰力(安塔蘭房間用來與上面兩個值對比)。
// 與 AI 態勢判斷共用同一個 playerMilitary,避免畫面上顯示的數字跟實際結算用的不是同一套。
func (s *GameSession) PlayerFleetStrength() int { return s.playerMilitary() }

// AssaultAntaresBlockReason 回傳「為什麼現在不能發動終局反攻」;可以發動時回空字串。
// CanAssaultAntares 只回 bool,玩家看不出卡在哪一條——把四個前置條件逐條講白。
func (s *GameSession) AssaultAntaresBlockReason() string {
	switch {
	case s.Victory.Over:
		return "對局已分出勝負"
	case s.DisableEvents:
		// 手冊 p.183:「This strategy is not available if you disabled Antaran Attacks
		// when setting up your game.」
		return "本局關閉了安塔蘭攻擊,此勝利路徑不可用"
	case !s.hasDimensionalPortal():
		return "尚未建成「" + dimensionalPortalBuildingName + "」——沒有它到不了安塔蘭母星"
	case len(s.Fleet().Ships) == 0:
		return "沒有艦隊可派"
	}
	return ""
}

// GrantDimensionalPortalForGallery 直接把次元傳送門記進第一個殖民地的建築清單。
//
// **僅供 headless 截圖廊使用**:安塔蘭王座廳的「發動終局反攻」按鈕要有傳送門才會亮,
// 而正常玩到多維物理再蓋出傳送門要幾十回合,截圖驗證等不起。正常遊戲流程不會呼叫它
// ——傳送門一律得靠 gamedata.Buildings 的建造佇列蓋出來。
func (s *GameSession) GrantDimensionalPortalForGallery() {
	if len(s.PlayerColonies) == 0 {
		return
	}
	for len(s.ColonyBuildings) < len(s.PlayerColonies) {
		s.ColonyBuildings = append(s.ColonyBuildings, map[string]bool{})
	}
	if s.ColonyBuildings[0] == nil {
		s.ColonyBuildings[0] = map[string]bool{}
	}
	s.ColonyBuildings[0][dimensionalPortalBuildingName] = true
}
