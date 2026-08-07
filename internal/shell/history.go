package shell

// 帝國歷史記錄(原版 module 122 的 Record_History_ 對應物)。
//
// 為什麼要有:2026-08-06 反組譯原版 Orion2.exe 的除錯符號表後確認,原版有一整個
// 歷史記錄子系統(module 122,73 個函式,含 Record_History_/Bill_Init_/Is_Ignoring_),
// 而 INFO 畫面的「History Graph」子畫面(Draw_History_Subscreen_)就是靠它畫國力折線圖。
// remake 先前完全沒有這層,所以 History Graph 做不出來(使用者實測 issue #5-1)。
// 見 docs/re/01-gap-report.md Part B。
//
// 本檔只做「每回合抓一次快照」這個最小忠實核心:原版逐項記錄的欄位清單尚未逐一反編
// (函式邊界問題,見 docs/re/00-orion2-symbols.md),故這裡記的是 remake 已有、
// 且對「國力走勢」有意義的四個維度,不臆造原版欄位。

// HistorySample 是某一回合、某個帝國的國力快照。
type HistorySample struct {
	Population int `json:"pop"`      // 帝國總人口
	BC         int `json:"bc"`       // 國庫
	Research   int `json:"research"` // 該回合研究產出(RP)
	Fleet      int `json:"fleet"`    // 艦隊戰力
}

// HistoryTurn 是一個回合的全帝國快照。Empires[0] 恆為玩家,其後依 AIPlayers 順序。
type HistoryTurn struct {
	Turn    int             `json:"turn"`
	Empires []HistorySample `json:"empires"`
}

// historyMaxTurns 是保留的回合數上限。折線圖只需要走勢,不需要無限長的紀錄;
// 超過就丟最舊的(環形語意),避免長局存檔無限膨脹。
const historyMaxTurns = 400

// recordHistory 在每回合結算後抓一次全帝國快照,存進 s.History。
// 由 EndTurn 呼叫(結算完成、數值已更新之後)。
func (s *GameSession) recordHistory() {
	emp := make([]HistorySample, 0, 1+len(s.AIPlayers))

	pop := 0
	for _, c := range s.PlayerColonies {
		pop += c.Population
	}
	emp = append(emp, HistorySample{
		Population: pop,
		BC:         s.Player.BC,
		Research:   s.LastPlayerOutput.TotalResearch,
		Fleet:      s.playerFleetStrength(),
	})

	for i := range s.AIPlayers {
		a := &s.AIPlayers[i]
		apop := 0
		for _, c := range a.Colonies {
			apop += c.Population
		}
		emp = append(emp, HistorySample{
			Population: apop,
			BC:         a.Player.BC,
			Research:   0, // AI 的研究產出目前不逐回合外露,留 0 不臆造
			Fleet:      a.FleetStrength,
		})
	}

	s.History = append(s.History, HistoryTurn{Turn: s.Turn, Empires: emp})
	if len(s.History) > historyMaxTurns {
		s.History = s.History[len(s.History)-historyMaxTurns:]
	}
}

// playerFleetStrength 把玩家艦隊換算成與 AIOpponent.FleetStrength 同尺度的戰力值,
// 讓折線圖上玩家與 AI 的軍力可以直接比較(AI 端是抽象累積值,見 advanceAI)。
func (s *GameSession) playerFleetStrength() int {
	n := 0
	for _, sh := range s.AllShips() { // 國力是**全帝國**的,不是目前選中那一支艦隊
		n += sh.WeaponAttack + sh.BonusHP/2
	}
	return n
}

// HistoryEmpireNames 回傳折線圖各條線對應的帝國名(順序同 HistoryTurn.Empires)。
func (s *GameSession) HistoryEmpireNames() []string {
	names := make([]string, 0, 1+len(s.AIPlayers))
	names = append(names, "你")
	for i := range s.AIPlayers {
		names = append(names, s.AIPlayers[i].Name)
	}
	return names
}

// HistoryMetric 是折線圖可顯示的指標。
type HistoryMetric int

const (
	HistoryPopulation HistoryMetric = iota
	HistoryBC
	HistoryFleet
)

// HistoryMetricName 回傳指標中文名(UI 用)。
func HistoryMetricName(m HistoryMetric) string {
	switch m {
	case HistoryBC:
		return "國庫"
	case HistoryFleet:
		return "艦隊戰力"
	default:
		return "人口"
	}
}

// HistorySeries 取出某指標的所有帝國時間序列,回傳 series[帝國][回合] 與各回合的 turn 值。
// 供 UI 畫折線圖;無資料時回傳 nil。
func (s *GameSession) HistorySeries(m HistoryMetric) (series [][]int, turns []int) {
	if len(s.History) == 0 {
		return nil, nil
	}
	n := 0
	for _, h := range s.History {
		if len(h.Empires) > n {
			n = len(h.Empires)
		}
	}
	series = make([][]int, n)
	for i := range series {
		series[i] = make([]int, 0, len(s.History))
	}
	turns = make([]int, 0, len(s.History))
	for _, h := range s.History {
		turns = append(turns, h.Turn)
		for i := 0; i < n; i++ {
			v := 0
			if i < len(h.Empires) {
				switch m {
				case HistoryBC:
					v = h.Empires[i].BC
				case HistoryFleet:
					v = h.Empires[i].Fleet
				default:
					v = h.Empires[i].Population
				}
			}
			series[i] = append(series[i], v)
		}
	}
	return series, turns
}

// PlayerFleetStrengthForUI 是 playerFleetStrength 的匯出版本(cmd/moo2 種族統計畫面用)。
func (s *GameSession) PlayerFleetStrengthForUI() int { return s.playerFleetStrength() }
