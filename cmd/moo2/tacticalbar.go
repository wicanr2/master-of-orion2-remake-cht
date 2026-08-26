package main

import (
	"fmt"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

// tacticalbar.go:格子戰術畫面底部控制列七顆按鈕的行為。
//
// ============ 這一輪之前的狀態 ============
//
// 七顆鈕**中文化做完了、熱區一個都沒有**。畫面上看起來可以點,點下去什麼都不會發生
// ——比沒翻譯更糟,因為它會讓人以為功能壞了。
//
// 第 86 項(hi-res 畫布)跑 2× 畫廊時才發現:先前每一輪都在看這張截圖,而
// 「按鈕長得對」與「按鈕能按」是兩個不同的問題,盤點時只問了前者
// (同 CLAUDE.md 的「元件表裡有 ≠ 效果有接」)。
//
// ============ 七顆鈕各自對應到什麼 ============
//
//	自動  ✅ 用同一套格子規則自動打完(不是另一個解算器,所以結果與手動一致)
//	掃描  ✅ 手冊「Scan gives you information about an enemy ship」→ 進掃描模式,點敵艦看資料
//	登船  ✅ 手冊的登艦戰(第 80 項),接 shell.ResolveBoarding
//	撤退  ✅ 保留倖存艦離場,判定為未勝
//	等待  ✅ 將目前艦移到本回合未行動艦之後
//	完成  ✅ 結束目前艦的行動；全部完成後才結算敵方回擊
//	選項  ⚠ 設定畫面已完成，但戰鬥中開啟後返回同一場戰局的轉場尚未接線
//
// 尚未完成的選項按鈕會說明為什麼沒有反應；其餘按鈕都接到實際戰鬥狀態。

// tacticalMode 是控制列切換出來的點擊模式。
type tacticalMode int

const (
	tacticalModeNormal tacticalMode = iota // 點敵艦 = 開火
	tacticalModeScan                       // 點敵艦 = 顯示資料
	tacticalModeBoard                      // 點敵艦 = 派登艦隊
)

// barButtonHit 回傳落在哪一顆控制列按鈕上(-1 = 沒中)。
//
// 矩形取 barButtonsCHT 的中心 + 原版按鈕尺寸 54×18(座標來源見那張表的註解)。
// **熱區與繪製共用同一份中心座標**——先前戰術控制列與畫面別處都吃過「兩份寫死座標
// 各自漂移」的虧(見 designModChipRect)。
func barButtonHit(mx, my int) int {
	for i, b := range barButtonsCHT {
		if hitBox(mx, my, b.cx-27, b.cy-9, 54, 18) {
			return i
		}
	}
	return -1
}

// handleBarButton 處理控制列點擊;回傳 true 表示這一下已被消化(不要再當成棋盤點擊)。
func (t *tacticalScreen) handleBarButton(idx int) bool {
	if idx < 0 || idx >= len(barButtonsCHT) {
		return false
	}
	if clickSound != nil {
		clickSound()
	}
	switch barButtonsCHT[idx].action {
	case "auto":
		t.autoResolve()
	case "scan":
		t.toggleMode(tacticalModeScan, uiText(t.b.lang, "tactical.mode.scan_on"))
	case "board":
		t.toggleMode(tacticalModeBoard, uiText(t.b.lang, "tactical.mode.board_on"))
	case "retreat":
		t.retreat()
	case "wait":
		t.waitSelectedAction()
	case "done":
		t.finishSelectedAction()
	case "options":
		t.log = uiText(t.b.lang, "tactical.options.combat_return_unavailable")
	default:
		return false
	}
	return true
}

// toggleMode 切到某個點擊模式;已在該模式則切回一般模式。
func (t *tacticalScreen) toggleMode(m tacticalMode, onMsg string) {
	if t.mode == m {
		t.mode = tacticalModeNormal
		t.log = uiText(t.b.lang, "tactical.mode.normal")
		return
	}
	t.mode = m
	t.log = onMsg
}

// retreat 撤出戰鬥:倖存艦全部保留,但這場不算贏。
//
// 手冊把撤退寫成「你可以隨時離開戰場」,代價是戰術目標沒達成。remake 這裡把它接到既有的
// 戰後流程(over/won),所以 ApplyCombatOutcome 會照常保留倖存艦、記錄一場敗戰。
func (t *tacticalScreen) retreat() {
	if t.over {
		return
	}
	t.over, t.won = true, false
	t.log = fmt.Sprintf(uiText(t.b.lang, "tactical.retreat.summary"), len(t.player))
}

// autoResolve 用**同一套格子規則**把剩下的戰鬥打完。
//
// ⚠ 刻意不呼叫快速結算(battleVolley 那一條路):兩條路的公式細節不同,自動打完的結果
// 若與手動不一致,玩家會學到「按自動比較划算(或比較虧)」——那是規則漏洞不是便利功能。
// 這裡就是重複執行「挑一個打得到的敵艦 → fireRound」,直到分出勝負。
//
// 迴圈上限是硬性的:fireRound 在沒有任何艦在射程內時**不會推進回合**,雙方又都不會移動,
// 那就是一個永遠不結束的迴圈(headless 畫廊踩過同款空轉,見 CLAUDE.md 的界限紀律)。
func (t *tacticalScreen) autoResolve() {
	const maxRounds = 200
	if t.shipInitiativeEnabled() {
		// 合併主動權模式不能再呼叫「全體玩家艦齊射」的 fireRound；逐一消費目前
		// 玩家行動，排在中間的敵艦由 advanceInitiativeQueue 自動執行。
		for n := 0; n < maxRounds*max(1, len(t.player)+len(t.enemy)) && !t.over; n++ {
			actor := t.currentInitiativePlayerIndex()
			if actor < 0 {
				t.advanceInitiativeQueue()
				continue
			}
			target := t.nearestReachableEnemyForShip(actor)
			if target < 0 {
				t.finishSelectedAction()
				continue
			}
			t.fireSelectedShip(target)
		}
		if !t.over {
			t.log = uiText(t.b.lang, "tactical.auto.round_cap")
		}
		return
	}
	for n := 0; n < maxRounds && !t.over; n++ {
		target := t.nearestReachableEnemy()
		if target < 0 {
			t.log = uiText(t.b.lang, "tactical.auto.no_target")
			return
		}
		before := t.round
		t.fireRound(target)
		if t.round == before { // 這一發沒推進回合(相位匿蹤/停滯),換一個目標會無限繞
			t.log += uiText(t.b.lang, "tactical.auto.stopped_suffix")
			return
		}
	}
	if !t.over {
		t.log = uiText(t.b.lang, "tactical.auto.round_cap")
	}
}

func (t *tacticalScreen) nearestReachableEnemyForShip(playerIndex int) int {
	if playerIndex < 0 || playerIndex >= len(t.player) || t.player[playerIndex].InStasis {
		return -1
	}
	best := -1
	for i := range t.enemy {
		if t.enemy[i].InStasis || shell.CloakUntargetable(t.enemy[i], t.round+1) {
			continue
		}
		if abs(t.player[playerIndex].Col-t.enemy[i].Col)+abs(t.player[playerIndex].Row-t.enemy[i].Row) > fireRange {
			continue
		}
		if best < 0 || t.enemy[i].HP < t.enemy[best].HP {
			best = i
		}
	}
	return best
}

// nearestReachableEnemy 回傳「任一我方艦射程內、且血最少」的敵艦索引(-1 = 都打不到)。
func (t *tacticalScreen) nearestReachableEnemy() int {
	best := -1
	for i := range t.enemy {
		if t.enemy[i].InStasis || shell.CloakUntargetable(t.enemy[i], t.round+1) {
			continue
		}
		reachable := false
		for j := range t.player {
			if t.player[j].InStasis {
				continue
			}
			if abs(t.player[j].Col-t.enemy[i].Col)+abs(t.player[j].Row-t.enemy[i].Row) <= fireRange {
				reachable = true
				break
			}
		}
		if !reachable {
			continue
		}
		if best < 0 || t.enemy[i].HP < t.enemy[best].HP {
			best = i
		}
	}
	return best
}

// scanEnemy 產生一艘敵艦的資料摘要(手冊的 SCAN)。
func (t *tacticalScreen) scanEnemy(idx int) string {
	e := t.enemy[idx]
	return fmt.Sprintf(uiText(t.b.lang, "tactical.scan.summary"),
		combatShipLabel(t.b.lang, t.b.session, e.Name), e.HP, e.MaxHP, e.ArmorHP, e.Attack, e.Defense, e.WeaponMin, e.WeaponMax,
		e.ShieldReduction, e.Marines)
}

// boardEnemy 讓選中的我方艦對 idx 敵艦發動登艦戰。
//
// 這裡只做 UI 該做的事:選取檢查、距離檢查、戰報文字。**規則本身在 shell**
// (`GameSession.ShipBoardingAttack`)——戰力怎麼算、誰倖存、奪不奪得到,都不在畫面層。
func (t *tacticalScreen) boardEnemy(idx int) {
	if t.sel < 0 || t.sel >= len(t.player) {
		t.log = uiText(t.b.lang, "tactical.board.select_ship")
		return
	}
	att := &t.player[t.sel]
	def := &t.enemy[idx]
	dist := abs(att.Col-def.Col) + abs(att.Row-def.Row)
	if !shell.ShipBoardingReachAgainst(*att, *def, dist) {
		t.log = uiText(t.b.lang, "tactical.board.unavailable")
		return
	}
	if shell.ShipBoardingPartySize(*att) <= 0 {
		t.log = fmt.Sprintf(uiText(t.b.lang, "tactical.board.no_marines"), att.Name)
		return
	}
	attName, defName := att.Name, combatShipLabel(t.b.lang, t.b.session, def.Name)
	res := t.b.session.ShipBoardingAttack(att, def, shell.BoardingCapture, func(n int) int {
		if n <= 0 {
			return 0
		}
		return t.rng.Intn(n)
	})
	if res.Captured {
		t.log = fmt.Sprintf(uiText(t.b.lang, "tactical.board.captured"),
			attName, defName)
		// 手冊:被奪的船換旗。remake 直接把它移出敵方序列——留在場上會繼續對玩家開火。
		t.enemy = append(t.enemy[:idx], t.enemy[idx+1:]...)
		if len(t.enemy) == 0 {
			t.over, t.won = true, true
			t.log += uiText(t.b.lang, "tactical.board.victory_suffix")
		}
		return
	}
	t.log = fmt.Sprintf(uiText(t.b.lang, "tactical.board.repelled"),
		attName, res.DefenderSurvived, res.AttackerSurvived, res.Rounds)
}

// modeHint 是畫在控制列上方的模式提示(一般模式回空字串)。
func (t *tacticalScreen) modeHint() string {
	switch t.mode {
	case tacticalModeScan:
		return uiText(t.b.lang, "tactical.mode.scan_hint")
	case tacticalModeBoard:
		return uiText(t.b.lang, "tactical.mode.board_hint")
	}
	return ""
}
