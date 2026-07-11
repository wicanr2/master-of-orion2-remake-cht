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

// antaranHomeFleetDefense 是安塔蘭母星防禦艦隊的戰力組成(remake 保守預設)。
//
// ⚠ 誠實聲明(手冊/openorion2 皆無精確數字):`GAME_MANUAL.pdf`「Winning」小節只用「the
// awe-inspiring Antarans」定性描述母星防禦強度,全文搜尋未見任何具體艦隊組成或戰力數字;
// openorion2(`docs/tech/rules-implementation-audit.md` 第 10 項已記載)對 victory/winner 相關
// 邏輯全 repo 零命中,自然也沒有母星防禦艦隊的資料可抄。
//
// 保守預設:6 艘「末日之星」等級戰力(shipStrength("末日之星")==64,MOO2 六級艦體中最高等級,
// 見 session.go shipStrength),合計戰力 384——確保玩家不能用隨手一支小艦隊反攻,呼應手冊
// 「awe-inspiring」的定性描述,同時仍是「打得贏」的固定值(不是無限强的裝飾性數字):玩家投入
// 同等量級的末日之星艦隊即可一戰。**這是 remake 保守預設,非手冊或 openorion2 給出的精確值**,
// 待考證(見 docs/tech/victory-conditions.md TODO)。
var antaranHomeFleetDefense = []int{64, 64, 64, 64, 64, 64}

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
	return !s.Victory.Over && !s.DisableEvents && s.hasDimensionalPortal() && len(s.Ships) > 0
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
// 戰力用 antaranHomeFleetDefense 保守預設。與 ResolveBattle 不同:這是「終局一戰」,PlayerWon
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
		for _, sh := range s.Ships {
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
