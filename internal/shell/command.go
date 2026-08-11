package shell

import (
	"fmt"
	"sort"
	"strings"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// command.go:**玩家指令**——把「玩家按了哪顆鈕」變成一筆可序列化、可重播的資料。
//
// 三個用途,而且三個都不是「為了網路才做的」:
//
//	網路對戰   兩台機器要套用**同樣的指令序列**才會算出同樣的狀態(見 internal/netplay)
//	回放/除錯  一局的完整指令序列 + 起始種子 = 可完整重現的 bug 報告
//	熱座       其實已經在做同一件事,只是指令直接就地套用、沒有序列化
//
// ⚠ **這一層不做前置檢查。** `ColonizePlanet` 之類的方法自己會回絕不合法的操作
// (艦隊沒到、行星已被佔…),指令層再檢查一次只會變成兩份會漂開的規則。
// 指令層唯一的責任是「把名字與參數對到正確的方法」,並且對**不認得的指令名回報錯誤**——
// 靜默忽略在鎖步裡是最糟的處理:一邊套用了、另一邊沒有,而且沒有人會知道。
//
// ⚠ 這裡刻意不相依 `internal/netplay`:規則層不該知道傳輸層。兩邊的欄位形狀一樣,
// 由組裝端(cmd/moo2)轉換——那是唯一同時認識兩層的地方。

// PlayerCommand 是一條玩家指令。
//
// 參數用位置式的 `Args []int` 而不是具名 map:具名參數會讓「新增一個參數」變成相容性問題,
// 而位置參數配上「指令名決定參數意義」的約定簡單得多(每一條的意義見 ApplyPlayerCommand)。
type PlayerCommand struct {
	Name string `json:"name"`
	Args []int  `json:"args,omitempty"`
	Text string `json:"text,omitempty"` // 少數需要字串參數的指令(建造項目名、艦體等級…)
}

// 指令名。用常數而不是散落的字串字面值——打錯字會在編譯期被抓到,
// 而不是在對局跑了二十回合之後以「對面沒套用到」的形式出現。
const (
	CmdSendFleet            = "send_fleet"             // Args[0]=目的星
	CmdSelectFleet          = "select_fleet"           // Args[0]=艦隊索引
	CmdSplitFleet           = "split_fleet"            // Args[0]=艦隊索引, Args[1:]=要拆出去的艦艇索引
	CmdColonizeStar         = "colonize_star"          // Args[0]=星
	CmdColonizePlanet       = "colonize_planet"        // Args[0]=行星
	CmdBuildOutpost         = "build_outpost"          // Args[0]=星
	CmdOutpostOnPlanet      = "outpost_on_planet"      // Args[0]=行星
	CmdLoadMarines          = "load_marines"           // Args[0]=殖民地
	CmdLoadTanks            = "load_tanks"             // Args[0]=殖民地
	CmdInvadeColony         = "invade_colony"          // Args[0]=星
	CmdBombardColony        = "bombard_colony"         // Args[0]=星
	CmdAttackMonster        = "attack_monster"         // Args[0]=星
	CmdEnqueueBuild         = "enqueue_build"          // Args[0]=殖民地, Args[1]=成本, Text=項目名
	CmdDequeueBuild         = "dequeue_build"          // Args[0]=殖民地, Args[1]=佇列位置
	CmdShiftJob             = "shift_job"              // Args[0]=殖民地, Text="來源>目標"(見 shiftJobArgs)
	CmdCycleTaxRate         = "cycle_tax_rate"         // 無參數
	CmdSetRelocation        = "set_relocation"         // Args[0]=殖民地, Args[1]=目標星(−1=取消)
	CmdSetStarReloc         = "set_star_reloc"         // Args[0]=起點星, Args[1]=終點星
	CmdSetAllReloc          = "set_all_reloc"          // Args[0]=目標星
	CmdClearAllReloc        = "clear_all_reloc"        // 無參數
	CmdHireMerc             = "hire_merc"              // 無參數
	CmdAssaultAntares       = "assault_antares"        // 無參數
	CmdMindControl          = "mind_control"           // Args[0]=敵方星
	CmdCycleColonyBuild     = "cycle_colony_build"     // Args[0]=殖民地
	CmdTrainSpy             = "train_spy"              // Args[0]=AI 索引
	CmdSetSpyMission        = "set_spy_mission"        // Args[0]=AI 索引, Args[1]=任務
	CmdCycleSpyMission      = "cycle_spy_mission"      // Args[0]=AI 索引
	CmdTrainAgent           = "train_agent"            // 無參數
	CmdDismissAgent         = "dismiss_agent"          // 無參數
	CmdSetResearch          = "set_research"           // Args[0]=研究主題
	CmdChooseResearch       = "choose_research"        // Args[0]=科技
	CmdDiplomacy            = "diplomacy"              // Text=action NUL enemy
	CmdOfferCashGift        = "offer_cash_gift"        // Args[0]=金額, Text=enemy
	CmdOfferTechGift        = "offer_tech_gift"        // Args[0]=主題, Args[1]=科技, Text=enemy
	CmdOfferStarGift        = "offer_star_gift"        // Args[0]=星, Text=enemy
	CmdClearAudience        = "clear_audience"         // Args[0]=AI 索引
	CmdRespondCouncil       = "respond_council"        // Args[0]=1 接受 / 0 拒絕
	CmdBuildShip            = "build_ship"             // Args[0:5]=元件, Args[5]=火線角, Text=艦體 NUL mods
	CmdHireMercAt           = "hire_merc_at"           // Args[0]=傭兵索引
	CmdAssignColonyLeader   = "assign_colony_leader"   // Args[0]=殖民地, Args[1]=領袖
	CmdUnassignColonyLeader = "unassign_colony_leader" // Args[0]=殖民地
	CmdAssignShipOfficer    = "assign_ship_officer"    // Args[0]=艦隊, Args[1]=艦, Args[2]=領袖
	CmdUnassignShipOfficer  = "unassign_ship_officer"  // Args[0]=艦隊, Args[1]=艦
	CmdReturnShipOfficer    = "return_ship_officer"    // Text=領袖名
	CmdDismissColonyLeader  = "dismiss_colony_leader"  // Text=領袖名
	CmdDismissShipOfficer   = "dismiss_ship_officer"   // Text=領袖名
	CmdCombatOutcome        = "combat_outcome"         // Args[0]=我方起始, Args[1]=敵方起始, Args[2]=勝敗; Text=存活艦名
	CmdEndTurn              = "end_turn"               // 無參數(鎖步裡由協定層決定何時推進,這裡是單機/回放用)
)

// PlayerCommandNames 回傳所有已知的指令名(排序過,供 UI/文件列表與測試比對)。
//
// ⚠ 這份清單就是「網路對戰目前支援到哪」的答案。UI 上做得到、但不在這張表裡的操作,
// 在網路對戰時**不會同步過去**——所以新增 UI 動作時要一起補這裡,不是可選的收尾。
func PlayerCommandNames() []string {
	names := []string{
		CmdAssaultAntares, CmdAttackMonster, CmdAssignColonyLeader, CmdAssignShipOfficer,
		CmdBombardColony, CmdBuildOutpost, CmdBuildShip, CmdChooseResearch,
		CmdClearAllReloc, CmdClearAudience, CmdColonizePlanet, CmdColonizeStar,
		CmdCombatOutcome, CmdCycleColonyBuild, CmdCycleSpyMission, CmdCycleTaxRate,
		CmdDequeueBuild, CmdDismissAgent, CmdDismissColonyLeader, CmdDismissShipOfficer,
		CmdDiplomacy, CmdEndTurn, CmdEnqueueBuild, CmdHireMerc, CmdHireMercAt,
		CmdInvadeColony, CmdLoadMarines, CmdLoadTanks, CmdMindControl,
		CmdOfferCashGift, CmdOfferStarGift, CmdOfferTechGift, CmdOutpostOnPlanet,
		CmdRespondCouncil, CmdReturnShipOfficer, CmdSelectFleet, CmdSendFleet,
		CmdSetAllReloc, CmdSetRelocation, CmdSetResearch, CmdSetSpyMission,
		CmdSetStarReloc, CmdShiftJob, CmdSplitFleet, CmdTrainAgent, CmdTrainSpy,
		CmdUnassignColonyLeader, CmdUnassignShipOfficer,
	}
	sort.Strings(names)
	return names
}

// arg 取第 i 個參數;沒有就回 def。
//
// 缺參數回預設值而不是報錯,是因為**參數數量不足代表送出端有 bug**,而鎖步中途中止一局
// 比套用一次無效操作更糟——無效操作會被規則層自己擋掉(見檔頭)。
// 真正該報錯的是「不認得的指令名」,那才是兩邊版本不一致的訊號。
func arg(c PlayerCommand, i, def int) int {
	if i < len(c.Args) {
		return c.Args[i]
	}
	return def
}

// ApplyPlayerCommand 套用一條指令。
//
// 回傳錯誤只有一種情況:**指令名不認得**。那代表對面的版本與這邊不同,
// 繼續玩下去一定會分岔,呼叫端應該停下來。
func (s *GameSession) ApplyPlayerCommand(c PlayerCommand) error {
	switch c.Name {
	case CmdSendFleet:
		s.SendFleet(arg(c, 0, -1))
	case CmdSelectFleet:
		s.SelectFleet(arg(c, 0, 0))
	case CmdSplitFleet:
		if len(c.Args) > 1 {
			s.SplitFleet(c.Args[0], c.Args[1:])
		}
	case CmdColonizeStar:
		s.ColonizeStar(arg(c, 0, -1))
	case CmdColonizePlanet:
		s.ColonizePlanet(arg(c, 0, -1))
	case CmdBuildOutpost:
		s.BuildOutpost(arg(c, 0, -1))
	case CmdOutpostOnPlanet:
		s.BuildOutpostOnPlanet(arg(c, 0, -1))
	case CmdLoadMarines:
		s.LoadMarines(arg(c, 0, -1))
	case CmdLoadTanks:
		s.LoadTanks(arg(c, 0, -1))
	case CmdInvadeColony:
		s.InvadeColony(arg(c, 0, -1))
	case CmdBombardColony:
		s.BombardColony(arg(c, 0, -1))
	case CmdAttackMonster:
		s.AttackMonster(arg(c, 0, -1))
	case CmdEnqueueBuild:
		s.EnqueueBuild(arg(c, 0, -1), c.Text, arg(c, 1, 0))
	case CmdDequeueBuild:
		s.DequeueBuild(arg(c, 0, -1), arg(c, 1, -1))
	case CmdShiftJob:
		from, to := shiftJobArgs(c.Text)
		s.ShiftColonyJob(arg(c, 0, -1), from, to)
	case CmdCycleTaxRate:
		s.CycleTaxRate()
	case CmdSetRelocation:
		s.SetColonyRelocation(arg(c, 0, -1), arg(c, 1, ColonyRelocationNone))
	case CmdSetStarReloc:
		s.SetStarRelocation(arg(c, 0, -1), arg(c, 1, -1))
	case CmdSetAllReloc:
		s.SetAllStarRelocations(arg(c, 0, -1))
	case CmdClearAllReloc:
		s.ClearAllStarRelocations()
	case CmdHireMerc:
		s.HireMerc()
	case CmdAssaultAntares:
		s.AssaultAntares()
	case CmdMindControl:
		s.MindControlColony(arg(c, 0, -1))
	case CmdCycleColonyBuild:
		s.CycleColonyBuild(arg(c, 0, -1))
	case CmdTrainSpy:
		s.TrainSpy(arg(c, 0, -1))
	case CmdSetSpyMission:
		s.SetSpyMission(arg(c, 0, -1), SpyMission(arg(c, 1, int(SpyMissionSteal))))
	case CmdCycleSpyMission:
		s.CycleSpyMission(arg(c, 0, -1))
	case CmdTrainAgent:
		s.TrainDefensiveAgent()
	case CmdDismissAgent:
		s.DismissDefensiveAgent()
	case CmdSetResearch:
		s.SetResearchTopic(gamedata.ResearchTopic(arg(c, 0, 0)))
	case CmdChooseResearch:
		s.ChooseResearchTech(gamedata.Technology(arg(c, 0, 0)))
	case CmdDiplomacy:
		action, enemy := splitCommandText(c.Text)
		s.DiplomacyResponse(action, enemy)
	case CmdOfferCashGift:
		s.OfferCashGift(c.Text, arg(c, 0, 0))
	case CmdOfferTechGift:
		s.OfferTechnologyGift(c.Text, gamedata.ResearchTopic(arg(c, 0, 0)), gamedata.Technology(arg(c, 1, 0)))
	case CmdOfferStarGift:
		s.OfferStarGift(c.Text, arg(c, 0, -1))
	case CmdClearAudience:
		s.ClearAudienceRequest(arg(c, 0, -1))
	case CmdRespondCouncil:
		s.RespondToCouncilElection(arg(c, 0, 0) != 0)
	case CmdBuildShip:
		class, modsText := splitCommandText(c.Text)
		s.BuildShipWithModsAndArc(class, arg(c, 0, 0), arg(c, 1, 0), arg(c, 2, 0), arg(c, 3, 0), splitNUL(modsText), gamedata.WeaponArc(arg(c, 5, int(gamedata.ARC_FWD))))
	case CmdHireMercAt:
		s.HireMercAt(arg(c, 0, -1))
	case CmdAssignColonyLeader:
		s.AssignLeaderToColony(arg(c, 0, -1), arg(c, 1, -1))
	case CmdUnassignColonyLeader:
		s.UnassignLeaderFromColony(arg(c, 0, -1))
	case CmdAssignShipOfficer:
		s.AssignOfficerToShip(arg(c, 0, -1), arg(c, 1, -1), arg(c, 2, -1))
	case CmdUnassignShipOfficer:
		s.UnassignOfficerFromShip(arg(c, 0, -1), arg(c, 1, -1))
	case CmdReturnShipOfficer:
		s.ReturnShipOfficerToPool(c.Text)
	case CmdDismissColonyLeader:
		s.DismissColonyLeader(c.Text)
	case CmdDismissShipOfficer:
		s.DismissShipOfficer(c.Text)
	case CmdCombatOutcome:
		parts := splitNUL(c.Text)
		if len(parts) == 0 || parts[0] == "" {
			// 清單／回放相容性測試會用零值指令探測每個名稱；真正的
			// 戰鬥結果一定由 ApplyCombatOutcome 帶上敵方名稱。
			return nil
		}
		survivors := make(map[string]bool, len(parts)-1)
		for _, name := range parts[1:] {
			if name != "" {
				survivors[name] = true
			}
		}
		s.ApplyCombatOutcome(parts[0], arg(c, 0, 0), arg(c, 1, 0), survivors, arg(c, 2, 0) != 0)
	case CmdEndTurn:
		s.EndTurn()
	default:
		return fmt.Errorf("shell: 不認得的指令 %q(對面的版本可能與這邊不同)", c.Name)
	}
	return nil
}

// ApplyPlayerCommands 依序套用一批指令,遇到不認得的就停下來並回報。
//
// 停下來而不是跳過:鎖步裡「跳過一條」等於兩邊從此不同步,而且不會有人發現。
func (s *GameSession) ApplyPlayerCommands(cmds []PlayerCommand) error {
	s.commandReplayDepth++
	defer func() { s.commandReplayDepth-- }()
	for i, c := range cmds {
		if err := s.ApplyPlayerCommand(c); err != nil {
			return fmt.Errorf("第 %d 條指令:%w", i, err)
		}
	}
	return nil
}

// splitCommandText 解析以 NUL 分隔的字串欄位。玩家名稱、艦名與敵方名稱
// 不會包含 NUL；其餘部分保留原樣，讓格式錯誤由規則或回合指紋揭露。
func splitCommandText(text string) (string, string) {
	parts := splitNUL(text)
	if len(parts) == 0 {
		return "", ""
	}
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], strings.Join(parts[1:], "\x00")
}

func splitNUL(text string) []string { return strings.Split(text, "\x00") }

// shiftJobArgs 把 "農夫>工人" 拆成兩個職務名(`ShiftColonyJob` 收的是中文職務名)。
//
// 用一個字串欄位而不是兩個:`PlayerCommand` 只有一個 Text 欄位,而職務轉移天生是一對。
// 分隔符用 '>' —— 職務名裡不會有它。
func shiftJobArgs(text string) (from, to string) {
	for i := 0; i < len(text); i++ {
		if text[i] == '>' {
			return text[:i], text[i+1:]
		}
	}
	return text, ""
}
