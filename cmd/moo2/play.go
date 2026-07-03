package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

// play.go:可玩遊戲殼的 ebiten 互動層——Screen 介面 + App(滑鼠/鍵盤輪詢)+ 互動主選單/遊戲畫面。
// 純邏輯(對局狀態、命中判定)在 internal/shell;此處只負責繪製與輸入輪詢。
//
// 為了能 headless 驗證互動流,App 支援「腳本化輸入」(script):每幀注入一個 InputState,
// 可重現「點新遊戲→點結束回合」再截圖,確認互動真的有效(retro-game-playtest 的玩家路徑驗證)。

const playW, playH = 640, 480

// transition 是畫面切換指令。
type transition struct {
	next screen
	quit bool
}

// screen 是一個可互動畫面。
type screen interface {
	update(in shell.InputState) *transition
	draw(dst *ebiten.Image, font *uifont.Font)
}

// --- 主選單畫面 ---

type menuScreen struct{ buttons []shell.Button }

func newMenuScreen() *menuScreen {
	return &menuScreen{buttons: []shell.Button{
		{X: 245, Y: 200, W: 150, H: 36, ID: "new", Label: "新遊戲"},
		{X: 245, Y: 250, W: 150, H: 36, ID: "quit", Label: "結束遊戲"},
	}}
}

func (m *menuScreen) update(in shell.InputState) *transition {
	switch shell.ClickedButton(m.buttons, in) {
	case "new":
		return &transition{next: newGameScreen(shell.NewDemoSession())}
	case "quit":
		return &transition{quit: true}
	}
	return nil
}

func (m *menuScreen) draw(dst *ebiten.Image, font *uifont.Font) {
	font.DrawCentered(dst, "銀河霸主 II — 繁體中文化", playW/2, 110, 26, color.RGBA{240, 220, 120, 255})
	font.DrawCentered(dst, "go / ebiten remake", playW/2, 145, 14, color.RGBA{150, 170, 210, 255})
	drawButtons(dst, font, m.buttons)
}

// --- 遊戲畫面(最小可玩:顯示帝國狀態 + 結束回合)---

type gameScreen struct {
	session *shell.GameSession
	buttons []shell.Button
	msg     string
}

func newGameScreen(s *shell.GameSession) *gameScreen {
	return &gameScreen{session: s, buttons: []shell.Button{
		{X: 430, Y: 420, W: 130, H: 36, ID: "endturn", Label: "結束回合"},
		{X: 240, Y: 420, W: 150, H: 36, ID: "colony", Label: "管理殖民地"},
		{X: 20, Y: 420, W: 130, H: 36, ID: "menu", Label: "返回主選單"},
	}}
}

func (g *gameScreen) update(in shell.InputState) *transition {
	switch shell.ClickedButton(g.buttons, in) {
	case "endturn":
		g.session.EndTurn()
		g.msg = fmt.Sprintf("第 %d 回合結算完成", g.session.Turn-1)
	case "colony":
		return &transition{next: newColonyManageScreen(g.session, 0)}
	case "menu":
		return &transition{next: newMenuScreen()}
	}
	return nil
}

func (g *gameScreen) draw(dst *ebiten.Image, font *uifont.Font) {
	gold := color.RGBA{240, 220, 120, 255}
	body := color.RGBA{220, 225, 235, 255}
	border := color.RGBA{80, 110, 180, 255}
	s := g.session
	font.DrawCentered(dst, fmt.Sprintf("星曆 %d — 帝國概況", s.Turn), playW/2, 30, 20, gold)
	vector.StrokeLine(dst, 16, 50, playW-16, 50, 1, border, false)

	// 玩家帝國
	out := s.LastPlayerOutput
	font.Draw(dst, "【我方帝國】", 32, 70, 16, gold)
	rows := []string{
		fmt.Sprintf("殖民地:%d 座", len(s.PlayerColonies)),
		fmt.Sprintf("國庫:%d BC", s.Player.BC),
		fmt.Sprintf("研究進度:%d", s.Player.ResearchProgress),
		fmt.Sprintf("上回合淨工業:%d ／ 研究:%d ／ 食物盈餘:%d",
			out.TotalNetIndustry, out.TotalResearch, out.TotalFood),
		fmt.Sprintf("上回合稅收:%d BC", out.TaxRevenue),
	}
	y := 96.0
	for _, r := range rows {
		font.Draw(dst, r, 44, y, 14, body)
		y += 24
	}

	// AI 對手
	y += 12
	font.Draw(dst, "【對手】", 32, y, 16, gold)
	y += 26
	for _, a := range s.AIPlayers {
		font.Draw(dst, fmt.Sprintf("%s ｜ 國庫 %d BC ｜ 研究進度 %d",
			a.Name, a.Player.BC, a.Player.ResearchProgress), 44, y, 14, body)
		y += 24
	}

	if g.msg != "" {
		font.DrawCentered(dst, g.msg, playW/2, 400, 14, color.RGBA{120, 220, 140, 255})
	}
	drawButtons(dst, font, g.buttons)
}

// drawButtons 畫一組按鈕(底板 + 邊框 + 置中中文)。
func drawButtons(dst *ebiten.Image, font *uifont.Font, btns []shell.Button) {
	for _, b := range btns {
		vector.DrawFilledRect(dst, float32(b.X), float32(b.Y), float32(b.W), float32(b.H),
			color.RGBA{28, 36, 64, 255}, false)
		vector.StrokeRect(dst, float32(b.X), float32(b.Y), float32(b.W), float32(b.H), 2,
			color.RGBA{90, 130, 200, 255}, false)
		font.DrawCentered(dst, b.Label, float64(b.X)+float64(b.W)/2, float64(b.Y)+float64(b.H)/2,
			15, color.RGBA{230, 235, 245, 255})
	}
}

// --- App(ebiten.Game)---

type playApp struct {
	cur  screen
	font *uifont.Font

	// headless 截圖 + 腳本化輸入(驗證用)
	shotPath string
	frames   int
	script   []shell.InputState
	tick     int
	saved    bool
}

func (a *playApp) pollInput() shell.InputState {
	if a.script != nil { // headless 腳本:每幀取一個注入輸入
		if idx := a.tick - 1; idx >= 0 && idx < len(a.script) {
			return a.script[idx]
		}
		return shell.InputState{}
	}
	x, y := ebiten.CursorPosition()
	return shell.InputState{
		MouseX: x, MouseY: y,
		ClickReleased: inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft),
	}
}

func (a *playApp) Update() error {
	a.tick++
	if t := a.cur.update(a.pollInput()); t != nil {
		if t.quit {
			return ebiten.Termination
		}
		if t.next != nil {
			a.cur = t.next
		}
	}
	if a.shotPath != "" && a.saved {
		return ebiten.Termination
	}
	return nil
}

func (a *playApp) Draw(dst *ebiten.Image) {
	dst.Fill(color.RGBA{16, 20, 40, 255})
	a.cur.draw(dst, a.font)
	if a.shotPath != "" && !a.saved && a.tick >= a.frames {
		if err := saveScreenshot(dst, a.shotPath); err != nil {
			fmt.Println("截圖失敗:", err)
		}
		a.saved = true
	}
}

func (a *playApp) Layout(int, int) (int, int) { return playW, playH }

// runPlay 啟動可玩遊戲殼。script 非 nil 時為 headless 驗證模式(注入輸入 + 截圖)。
func runPlay(fnt *uifont.Font, shot string, frames int, script []shell.InputState) error {
	if fnt == nil {
		return fmt.Errorf("可玩模式需以 -font 指定字型")
	}
	a := &playApp{cur: newMenuScreen(), font: fnt, shotPath: shot, frames: frames, script: script}
	ebiten.SetWindowSize(playW, playH)
	ebiten.SetWindowTitle("Master of Orion II — 繁體中文化")
	return ebiten.RunGame(a)
}

// --- 殖民地管理畫面(核心 gameplay:調整人口分配)---

type colonyManageScreen struct {
	session *shell.GameSession
	idx     int
	buttons []shell.Button
}

func newColonyManageScreen(s *shell.GameSession, idx int) *colonyManageScreen {
	return &colonyManageScreen{session: s, idx: idx, buttons: []shell.Button{
		{X: 300, Y: 130, W: 40, H: 30, ID: "f-", Label: "▼"},
		{X: 345, Y: 130, W: 40, H: 30, ID: "f+", Label: "▲"},
		{X: 300, Y: 170, W: 40, H: 30, ID: "w-", Label: "▼"},
		{X: 345, Y: 170, W: 40, H: 30, ID: "w+", Label: "▲"},
		{X: 300, Y: 210, W: 40, H: 30, ID: "s-", Label: "▼"},
		{X: 345, Y: 210, W: 40, H: 30, ID: "s+", Label: "▲"},
		{X: 20, Y: 420, W: 130, H: 36, ID: "back", Label: "返回帝國"},
		{X: 260, Y: 420, W: 120, H: 36, ID: "next", Label: "下個殖民地"},
		{X: 430, Y: 420, W: 130, H: 36, ID: "endturn", Label: "結束回合"},
	}}
}

// shiftJob 在保持總人口不變下,把 1 人從 from 職務移到 to 職務(from 需有人)。
func shiftJob(cs *engine.ColonyState, from, to string) {
	get := func(j string) *int {
		switch j {
		case "f":
			return &cs.Farmers
		case "w":
			return &cs.Workers
		default:
			return &cs.Scientists
		}
	}
	fp, tp := get(from), get(to)
	if *fp > 0 {
		*fp--
		*tp++
	}
}

func (c *colonyManageScreen) update(in shell.InputState) *transition {
	cs := &c.session.PlayerColonies[c.idx]
	switch shell.ClickedButton(c.buttons, in) {
	case "f+": // 加農夫:從工人抽,工人為 0 則從科學家
		if cs.Workers > 0 {
			shiftJob(cs, "w", "f")
		} else {
			shiftJob(cs, "s", "f")
		}
	case "f-":
		shiftJob(cs, "f", "w")
	case "w+":
		if cs.Scientists > 0 {
			shiftJob(cs, "s", "w")
		} else {
			shiftJob(cs, "f", "w")
		}
	case "w-":
		shiftJob(cs, "w", "s")
	case "s+":
		if cs.Workers > 0 {
			shiftJob(cs, "w", "s")
		} else {
			shiftJob(cs, "f", "s")
		}
	case "s-":
		shiftJob(cs, "s", "w")
	case "next":
		c.idx = (c.idx + 1) % len(c.session.PlayerColonies)
	case "back":
		return &transition{next: newGameScreen(c.session)}
	case "endturn":
		c.session.EndTurn()
	}
	return nil
}

func (c *colonyManageScreen) draw(dst *ebiten.Image, font *uifont.Font) {
	gold := color.RGBA{240, 220, 120, 255}
	body := color.RGBA{220, 225, 235, 255}
	border := color.RGBA{80, 110, 180, 255}
	cs := c.session.PlayerColonies[c.idx]
	out := engine.RunColonyTurn(cs)

	font.DrawCentered(dst, fmt.Sprintf("殖民地 %d / %d — 人口分配", c.idx+1, len(c.session.PlayerColonies)),
		playW/2, 30, 20, gold)
	vector.StrokeLine(dst, 16, 50, playW-16, 50, 1, border, false)
	font.Draw(dst, fmt.Sprintf("總人口:%d ／ 上限 %d", cs.Population, cs.PopMax), 40, 80, 15, body)

	// 三個職務列(名稱 + 人數 + ▲▼)
	labels := []struct {
		name string
		n    int
	}{{"農夫", cs.Farmers}, {"工人", cs.Workers}, {"科學家", cs.Scientists}}
	y := 145.0
	for _, l := range labels {
		font.Draw(dst, fmt.Sprintf("%s:%d", l.name, l.n), 60, y, 16, body)
		y += 40
	}

	// 即時產出預覽(這就是分配的效果)
	font.Draw(dst, "本回合預估產出:", 420, 130, 15, gold)
	preview := []string{
		fmt.Sprintf("食物 %d(消耗 %d)", out.Food, out.FoodConsumed),
		fmt.Sprintf("淨工業 %d", out.NetIndustry),
		fmt.Sprintf("研究 %d", out.Research),
		fmt.Sprintf("成長 %d", out.PopGrowth),
	}
	py := 160.0
	for _, p := range preview {
		font.Draw(dst, p, 430, py, 14, body)
		py += 26
	}
	if out.Starving {
		font.Draw(dst, "⚠ 食物不足,饑荒!", 430, py, 14, color.RGBA{230, 120, 120, 255})
	}

	drawButtons(dst, font, c.buttons)
}
