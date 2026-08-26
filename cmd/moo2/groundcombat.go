package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/assets"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/lbx"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

// groundcombat.go:地面戰畫面(原版 `Colony_Landing` / `Colony_Combat`)。
//
// remake 先前把一整場地面入侵壓成星系主畫面上的**一行字**(「入侵勝利!佔領此星(存活 3
// ／敵剩 0)」)。解算本身是完整的(internal/shell/ground_invasion.go + gamedata 的
// ResolveGroundBattle),缺的只是把它畫出來。
//
// ============ 版面:全部取自反組譯,沒有一個座標是估的 ============
//
// 兩側兵力面板(`sub_B8BC7` 攻方 / `sub_B8C8B` 守方,兩者逐指令對稱):
//
//	                       攻方                守方
//	  面板外框貼圖 (x,y)    (1, 40)             (378, 40)
//	  暗化填色 x0,y0,x1,y1  2,41,259,184        379,41,638,184
//	  標題/統計文字 x       130                 508
//	  文字列 y              50 + 列序×11        50 + 列序×11
//	  兵種欄位基準 X        1                   378
//
// 兩個函式語意由符號表確認,不是猜的:`sub_1210FD` = **`Print_Centered_`**(所以 130/508 是
// **置中**錨點而非左緣)、`sub_1298DE` = **`Darken_Fill_`**(所以面板是壓在地表上的半透明暗塊,
// 不是不透明底圖——原版底下透出的是殖民地地表)。
//
// 面板寬 261 這個數字有兩個獨立來源互相印證:①`Print_Troop_Totals_`(0xB896D)算兵種
// 欄位 x 時是 `261 / (兵種數+1) × (序號+1) + 基準X`(`mov eax, 105h / idiv ebx`);
// ②COLGCBT.LBX 資產 21 的實際尺寸正是 **261×149**。所以資產 21 是**兵力統計面板的外框**,
// 不是戰場視窗——戰場是整個 640×480 畫面本身。
//
// 部隊落點(`sub_B88B2`,逐指令):
//
//	X = 基準X[side] + Random_(50) − 20      → 攻方 31..80、守方 571..620
//	Y = min(360 + Random_(85), 430)         → 361..430
//	基準X 來自常數 dword_B6CDE = 0x024E0032 → word[0]=50(攻方)、word[1]=590(守方)
//
// (`Random_` @ 0x1247A0 回 1..n,不是 0..n−1,見 docs/re/01-gap-report.md 的語意訂正。)
//
// 其他已釘死但本畫面用不到的事實,一併記錄免得重查:
//   - 每側最多 **40** 個單位、每單位 **25** 位元組的記錄(`byte_19EB94 + side*0x3E8 + i*0x19`),
//     欄位:+0 X、+2 Y、+5 狀態、+6 兵種、+7 側、+0x0A 動畫偏移(`Random_(3) − 2`)。
//   - **4 種兵種**(`Print_Troop_Totals_` 的 `cmp si, 4`、`Player_Troop_Anim_` 的 4-case switch)。
//   - 兵種小圖示來自 **RACEICON.LBX**(`Player_Troop_Anim_` @ 0xBB723:每族 13 個,
//     索引 = 種族×13 + 7/8/9/10),不是 COLGCBT。
//   - `Replace_Colgcbt_Color_With_Player_Colors_` @ 0xB8EFB:COLGCBT sprite 上有一段
//     **保留色會被換成該帝國旗色**——所以 dump 出來士兵腳下那塊洋紅色不是影子,是旗色佔位。
//
// ============ 美術來源 ============
//
//	`sub_B6D51` 載 COLONY.LBX 資產 7 + COLGCBT.LBX 資產 0/1/2/3/4 到五個全域,再對其中四個
//	呼叫 `Replace_Colgcbt_Color_With_Player_Colors_`。**五個**正好對上一組單位的五段動畫。
//
//	COLGCBT.LBX 共 22 個資產,幀數排列自己把分組講清楚了:
//	  資產 0-4   23×27,幀數 11/17/12/16/1
//	  資產 5-9   23×27,幀數 11/17/12/18/1   ← 與上一組同構,兩組步兵
//	  資產 10-14 37×33,幀數 4/16/14/15/1    ← 履帶戰車(側視)
//	  資產 15-19 37×33,幀數 12/18/7/13/1    ← 機動裝甲兵(直立機甲)
//	  資產 20    9×8(彈著/爆點),資產 21 261×149(兵力面板外框)
//	「五個一組、最後一格恆為 1 幀」是四組共有的形狀,與 sub_B6D51 載五個資產完全吻合。
//
//	⚠ 調色盤未定案:COLGCBT **所有資產都沒有內嵌調色盤**,COLONY.LBX 資產 0-8 也都沒有。
//	原版的地面戰畫在**已經載入的殖民地畫面上**,直接沿用當時的調色盤,檔案裡自然不帶一份。
//	remake 借 COLBLDG.LBX#0(殖民地建築美術,同一個畫面家族)上色,渲染結果合理
//	(綠色步兵/戰車/機甲)。**這是尚未證實的選擇**,不是反組譯釘死的結論。
//
// ============ 呈現範圍(誠實聲明)============
//
//	① 原版是逐幀動畫的即時戰鬥;remake 的 ResolveGroundBattle 是一次算完的回合制解算,
//	   沒有逐單位位置/時間軸可驅動動畫。這個畫面呈現**戰後定格**,單位按原版落點公式排開。
//	② 原版的底圖是該殖民地的地表(畫在殖民地畫面上);remake 沒有那一層,用深色底代替。
//	③ 原版落點的 Random_ 是即時亂數;remake 用單位序號推出的固定散布(見 spreadX/spreadY),
//	   同一場戰報重畫結果一致——這是刻意的,截圖驗證需要可重現。
//	④ 原版有 4 種兵種;remake 的地面戰只模型化陸戰隊與戰車營兩種。

// 原版兵力面板座標(`sub_B8BC7` / `sub_B8C8B`)。
const (
	gcPanelW, gcPanelH           = 261, 149 // COLGCBT 資產 21 的實際尺寸,同時是 Print_Troop_Totals_ 的除數
	gcAtkPanelX                  = 1        // 攻方面板貼圖左上
	gcDefPanelX                  = 378      // 守方面板貼圖左上
	gcPanelY                     = 40
	gcAtkTextX                   = 130      // 攻方面板文字 x
	gcDefTextX                   = 508      // 守方面板文字 x
	gcTextY0                     = 50       // 首列 y
	gcTextRowH                   = 16       // 原版列距 11；runtime CJK 10px 字墨高為 16，故採不重疊安全列。
	gcDarkenAtkX0, gcDarkenAtkX1 = 2, 259   // Darken_Fill_(sub_B8BC7)
	gcDarkenDefX0, gcDarkenDefX1 = 379, 638 // Darken_Fill_(sub_B8C8B)
	gcDarkenY0, gcDarkenY1       = 41, 184
)

// 原版部隊落點(`sub_B88B2` + 常數 dword_B6CDE)。
const (
	gcAtkBaseX  = 50  // dword_B6CDE word[0]
	gcDefBaseX  = 590 // dword_B6CDE word[1]
	gcSpreadN   = 50  // Random_(50)
	gcSpreadSub = 20  // − 20
	gcBaseY     = 360 // 360 + Random_(85)
	gcSpreadYN  = 85
	gcMaxY      = 430 // 上限(原版 cmp eax, 1AEh)
)

// COLGCBT 資產分組(依據見檔頭)。各組取首段動畫的首幀當靜態圖示。
const (
	gcbtFrameAsset     = 21 // 兵力面板外框
	gcbtMarineAsset    = 0  // 步兵(第一組)
	gcbtDefMarineAsset = 5  // 步兵(第二組,拿來當守方,兩組同構)
	gcbtTankAsset      = 10 // 履帶戰車
	gcbtBattleoidAsset = 15 // 機動裝甲兵
)

// gcbtPaletteProvider 是 COLGCBT sprite 借用的調色盤來源(見檔頭「調色盤未定案」)。
const gcbtPaletteProvider = "colbldg.lbx"

// 兩張 raw index 表由 sub_B8EFB 的 byte_1828D3／byte_182913 匯出。
var groundColorMapC0 = [8][8]uint8{
	{0x92, 0xEC, 0xED, 0xEE, 0xEF, 0x96, 0x97, 0x98},
	{0xDA, 0xDB, 0xDC, 0xD4, 0xDE, 0xDF, 0xE0, 0xE1},
	{0xC0, 0xC1, 0xC2, 0xC3, 0xC4, 0xC5, 0xC6, 0xC7},
	{0x59, 0x5D, 0x61, 0x63, 0x65, 0x67, 0x69, 0x6C},
	{0x70, 0x72, 0x73, 0x74, 0x75, 0x76, 0x77, 0x87},
	{0xD2, 0xD3, 0xD4, 0xD5, 0xD6, 0xD7, 0xD8, 0xD9},
	{0x79, 0x7A, 0x7B, 0x7C, 0x7C, 0x7D, 0x7E, 0x7F},
	{0xB7, 0xB8, 0xB9, 0xBA, 0xBB, 0xBC, 0xBD, 0xBD},
}

var groundColorMapE8 = [8][4]uint8{
	{0xEC, 0xED, 0xEE, 0xEF}, {0xDB, 0xDD, 0xDF, 0xE1},
	{0xE8, 0xE8, 0xEA, 0xEB}, {0x56, 0x61, 0x67, 0x6B},
	{0x72, 0x74, 0x76, 0x77}, {0xD3, 0xD5, 0xD6, 0xD8},
	{0x79, 0x7B, 0x7D, 0x7F}, {0xB8, 0xB9, 0xBA, 0xBB},
}

// remapGroundPlayerIndexes 對應 sub_B8EFB 的兩段 palette-index 替換。raw color 2
// 是 COLGCBT 原生色，原版會直接返回；其他顏色分別替換 C0..C7 與 E8..EB。
func remapGroundPlayerIndexes(src *lbx.Frame, flag int) *lbx.Frame {
	if src == nil || flag == 2 {
		return src
	}
	if flag < 0 || flag >= len(groundColorMapC0) {
		flag = 0
	}
	out := &lbx.Frame{W: src.W, H: src.H, Index: append([]uint8(nil), src.Index...), Written: append([]bool(nil), src.Written...)}
	for i, idx := range out.Index {
		switch {
		case idx >= 0xC0 && idx <= 0xC7:
			out.Index[i] = groundColorMapC0[flag][idx-0xC0]
		case idx >= 0xE8 && idx <= 0xEB:
			out.Index[i] = groundColorMapE8[flag][idx-0xE8]
		}
	}
	return out
}

// loadGroundSprite 取 COLGCBT 某資產的首幀。任何一步失敗都回 nil,畫面退成純色方塊。
func loadGroundSprite(res *assets.Resolver, assetID, flag int, replacePlayerColor bool) *ebiten.Image {
	prov, err := decodeAsset(res, gcbtPaletteProvider, 0)
	if err != nil || prov.Embedded == nil {
		return nil
	}
	im, err := decodeAsset(res, "colgcbt.lbx", assetID)
	if err != nil || len(im.Frames) == 0 {
		return nil
	}
	frame := im.Frames[0]
	if replacePlayerColor {
		frame = remapGroundPlayerIndexes(frame, flag)
	}
	rgba := frame.ToRGBA(prov.Embedded, true)
	return ebiten.NewImageFromImage(rgba)
}

// spreadX / spreadY 是原版落點公式的**可重現版**:把 Random_(n) 換成由單位序號推出的固定值。
// 分布範圍與原版完全一致(X:基準−19..基準+30;Y:361..430),只是不擲骰。
// 兩個乘數取互質的小質數,讓連號單位不會排成整齊的斜線。
func spreadX(base, i int) int { return base + (i*17)%gcSpreadN + 1 - gcSpreadSub }
func spreadY(i int) int {
	y := gcBaseY + (i*29)%gcSpreadYN + 1
	if y > gcMaxY {
		y = gcMaxY
	}
	return y
}

// groundCombatScreen 是一場地面入侵的戰報畫面。
type groundCombatScreen struct {
	b   *sceneBuilder
	fnt *uifont.Font
	res shell.GroundInvasionResult

	panel, marine, defMarine, tank, battleoid *ebiten.Image
}

func newGroundCombatScreen(b *sceneBuilder, res shell.GroundInvasionResult) *groundCombatScreen {
	playSceneBGM(trackColonyCombat) // Colony_Combat_Screen_ → STREAM #10(第 73 項(音樂場景表))
	return &groundCombatScreen{
		b: b, fnt: b.fnt, res: res,
		panel:     loadGroundSprite(b.res, gcbtFrameAsset, 0, false),
		marine:    loadGroundSprite(b.res, gcbtMarineAsset, res.AttackerColor, true),
		defMarine: loadGroundSprite(b.res, gcbtDefMarineAsset, res.DefenderColor, true),
		tank:      loadGroundSprite(b.res, gcbtTankAsset, res.AttackerColor, true),
		battleoid: loadGroundSprite(b.res, gcbtBattleoidAsset, res.AttackerColor, true),
	}
}

// 「繼續」鈕:原版這個畫面是戰鬥中即時推進、沒有這顆鈕(戰完自動回殖民地畫面),
// remake 是戰後定格,需要一個離開的出口。擺在兩個面板下方(面板下緣 y=189)的空白帶,
// 不覆蓋任何原版元素。
func (g *groundCombatScreen) contRect() (int, int, int, int) { return 265, 232, 110, 30 }

func groundCombatTitleTextRect() textSafeRect {
	return textSafeRect{x: 10, y: 5, w: 620, h: 26, insetX: 4, insetY: 1, lineH: 24}
}

func groundCombatSideTextRect(textX, row int) textSafeRect {
	return textSafeRect{x: textX - 125, y: gcTextY0 - gcTextRowH/2 + row*gcTextRowH,
		w: gcPanelW - 10, h: gcTextRowH, insetX: 2, lineH: gcTextRowH}
}

func groundCombatOutcomeTextRect() textSafeRect {
	return textSafeRect{x: 40, y: 192, w: 560, h: 32, insetX: 5, insetY: 1, lineH: 30}
}

func (g *groundCombatScreen) continueTextRect() textSafeRect {
	x, y, w, h := g.contRect()
	return textSafeRect{x: x, y: y, w: w, h: h, insetX: 5, insetY: 1, lineH: h - 2}
}

func (g *groundCombatScreen) update(in shell.InputState) *origTransition {
	if !in.ClickReleased {
		return nil
	}
	if x, y, w, h := g.contRect(); hitRect(in, x, y, w, h) {
		return g.b.goTo(g.b.galaxy, uiText(g.b.lang, "groundcombat.transition.galaxy"))
	}
	return nil
}

// drawTroops 依原版落點公式擺 n 個單位。footY 是腳底 y(單位站在地上而不是浮著),
// flip=true 時左右鏡射(守方面向攻方)。startIdx 讓步兵與載具用不同的序號區段,散布才不重疊。
func drawTroops(dst *ebiten.Image, sp *ebiten.Image, n, baseX, startIdx int, flip bool) {
	if sp == nil || n <= 0 {
		return
	}
	w, h := sp.Bounds().Dx(), sp.Bounds().Dy()
	for i := 0; i < n; i++ {
		k := startIdx + i
		x, footY := spreadX(baseX, k), spreadY(k)
		op := &ebiten.DrawImageOptions{}
		if flip {
			op.GeoM.Scale(-1, 1)
			op.GeoM.Translate(float64(w), 0)
		}
		op.GeoM.Translate(float64(x-w/2), float64(footY-h))
		drawPanelImage(dst, sp, op)
	}
}

// drawSidePanel 畫一側的兵力面板(外框 + 逐列文字),座標全部來自 sub_B8BC7 / sub_B8C8B。
func (g *groundCombatScreen) drawSidePanel(dst *ebiten.Image, panelX, textX, dx0, dx1 int, lines []string, col color.RGBA) {
	// Darken_Fill_:原版是把地表壓暗,不是塗一塊不透明底。remake 沒有地表層,
	// 用半透明黑達成同樣的「壓暗」語意(底下是純色時效果等同深色塊)。
	fillPanel(dst, float32(dx0), gcDarkenY0, float32(dx1-dx0), gcDarkenY1-gcDarkenY0,
		color.RGBA{0, 0, 0, 150}, false)
	if g.panel != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(panelX), gcPanelY)
		drawPanelImage(dst, g.panel, op)
	} else {
		vector.StrokeRect(dst, float32(panelX), gcPanelY, gcPanelW, gcPanelH, 1.5,
			color.RGBA{140, 150, 130, 255}, false)
	}
	for i, ln := range lines {
		groundCombatSideTextRect(textX, i).drawCentered(dst, g.fnt, ln, 10, col)
	}
}

func (g *groundCombatScreen) draw(dst *ebiten.Image) {
	// 原版底圖是該殖民地的地表(這個畫面疊在殖民地畫面上);remake 沒有那一層,用深色底。
	dst.Fill(color.RGBA{18, 22, 16, 255})
	if g.fnt == nil {
		return
	}
	gold := color.RGBA{240, 220, 120, 255}
	body := color.RGBA{228, 234, 220, 255}
	atkCol := color.RGBA{150, 205, 245, 255}
	defCol := color.RGBA{240, 160, 140, 255}

	// 標題:原版 Draw_Colony_Landing_Screen_ 在 (319, 10) 畫一行字(mov eax,13Fh / mov edx,0Ah)。
	// 那個 y 是文字**上緣**不是中心(當中心的話會被畫面上緣切掉),remake 的 DrawCentered
	// 以中心對齊,故加半個字高。
	const titleSize = 16
	title := uiText(g.b.lang, "groundcombat.title.default")
	if g.res.ColonyName != "" {
		title = fmt.Sprintf(uiText(g.b.lang, "groundcombat.title.colony"), g.res.ColonyName)
	}
	groundCombatTitleTextRect().drawCentered(dst, g.fnt, title, titleSize, gold)

	r := g.res
	g.drawSidePanel(dst, gcAtkPanelX, gcAtkTextX, gcDarkenAtkX0, gcDarkenAtkX1, []string{
		uiText(g.b.lang, "groundcombat.side.attacker"),
		fmt.Sprintf(uiText(g.b.lang, "groundcombat.attacker.marines"), r.AttackerMarinesStart, r.AttackerMarinesSurvived),
		fmt.Sprintf(uiText(g.b.lang, "groundcombat.attacker.armor"), r.AttackerTanksStart, r.AttackerTanksSurvived),
		fmt.Sprintf(uiText(g.b.lang, "groundcombat.attacker.survivors"), r.AttackerSurvived),
	}, atkCol)
	g.drawSidePanel(dst, gcDefPanelX, gcDefTextX, gcDarkenDefX0, gcDarkenDefX1, []string{
		uiText(g.b.lang, "groundcombat.side.defender"),
		fmt.Sprintf(uiText(g.b.lang, "groundcombat.defender.garrison"), r.DefenderStart, r.DefenderSurvived),
		fmt.Sprintf(uiText(g.b.lang, "groundcombat.defender.rounds"), r.Rounds),
	}, defCol)

	// 戰場:整個畫面。部隊落點用原版公式(見檔頭)。
	drawTroops(dst, g.marine, r.AttackerMarinesStart, gcAtkBaseX, 0, false)
	if r.AttackerTanksStart > 0 {
		sp := g.tank
		if g.b.session != nil && g.b.session.HasBattleoids() {
			sp = g.battleoid // 已研究機動裝甲兵 → 換成機甲(手冊 p.101)
		}
		drawTroops(dst, sp, r.AttackerTanksStart, gcAtkBaseX, 20, false)
	}
	drawTroops(dst, g.defMarine, r.DefenderStart, gcDefBaseX, 7, true)

	outcome, outCol := uiText(g.b.lang, "groundcombat.outcome.repelled"), defCol
	if r.AttackerWon {
		outcome, outCol = uiText(g.b.lang, "groundcombat.outcome.success"), color.RGBA{160, 230, 160, 255}
		if r.StarCaptured {
			outcome = uiText(g.b.lang, "groundcombat.outcome.captured")
		}
	}
	groundCombatOutcomeTextRect().drawCentered(dst, g.fnt, outcome, 15, outCol)

	cx, cy, cw, ch := g.contRect()
	fillPanel(dst, float32(cx), float32(cy), float32(cw), float32(ch), color.RGBA{38, 44, 34, 255}, false)
	vector.StrokeRect(dst, float32(cx), float32(cy), float32(cw), float32(ch), 1.5, color.RGBA{150, 170, 130, 255}, false)
	g.continueTextRect().drawCentered(dst, g.fnt, uiText(g.b.lang, "groundcombat.button.continue"), 14, body)
}

// groundCombat 進入地面戰畫面。
func (b *sceneBuilder) groundCombat(res shell.GroundInvasionResult) (origScreen, error) {
	if b.session == nil {
		return nil, fmt.Errorf("%s", uiText(b.lang, "groundcombat.error.no_session"))
	}
	return newGroundCombatScreen(b, res), nil
}
