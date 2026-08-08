package main

// commandpoints.go:**指揮點數視窗**(原版 `Show_Command_Points_Screen_` @ 0x8BAB9)。
//
// 原版有這個獨立畫面,remake 先前只在星圖右欄第 2 格顯示一個淨值數字——玩家看得到
// 「還剩幾點」,但看不到那個數字是怎麼來的(起始值多少、建築給了多少、艦隊吃掉多少)。
// 指揮點數不足會直接扣國庫(手冊 p.37「每超額 1 點 → 每回合 −1 BC」),所以「為什麼是這個
// 數字」是玩家真的需要的資訊。
//
// ============ 畫面結構(反組譯)============
//
// `Show_Command_Points_Screen_` 很短,結構一目了然:
//
//	sub_1191CA(eax=&Draw_Mini_Main_Screen_, edx=1)   ; 背景重繪掛成「迷你星圖」
//	sub_11438B(0, 0, 0x27F, 0x1DF, key=0x1B)         ; 整螢幕隱形欄位,ESC 關閉
//	sub_128C32(0, 0, 0x27F, 0x1DF, 0)                ; Fill 清畫面
//	Draw_Mini_Main_Screen_()                          ; 畫迷你星圖當底
//	Show_Command_Points_(玩家索引)                    ; → sub_E2000 組文字 → sub_DDF24 顯示
//
// 也就是**「迷你星圖當背景 + 一塊文字視窗,ESC / 點擊關閉」**。
//
// ============ 內容:符號名就是權威 ============
//
// 文字本身在執行期才載入的字串區塊裡(`sub_DD4FD` 用 `repne scasb` 逐條走過去),沒有解出
// 英文原句。但**執行檔裡帶著符號表**,那幾個訊息的名字直接說明了視窗顯示什麼:
//
//	_starting_command_points_msg     起始指揮點數
//	_total_command_points_msg        指揮點數總計
//	_total_command_points_used_msg   已使用(複數)
//	_total_command_point_used_msg    已使用(單數)—— 原版連單複數都分了兩條
//	_command_summary_msg             總結
//	_command_points_window_field     這個視窗的欄位
//
// ⚠ 所以這個畫面的**結構與欄位組成是原版真值,中文用字是 remake 自己的**。
// 真的要對到原版的句子,得先把 `sub_DD4FD` 載入的那個字串區塊解出來。

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

// 文字視窗的框(remake 版面:置中於星圖框內)。原版的視窗座標由 `sub_DDF24` 這支
// 泛用訊息視窗決定,那支同時被十幾個畫面共用,座標是傳進去的——沒有「指揮點數專用」的
// 一組立即數可抄,所以這裡是 remake 自己排的。
const (
	cpPanelX, cpPanelY = 150, 130
	cpPanelW, cpPanelH = 340, 236
)

type commandPointsScreen struct {
	b *sceneBuilder
}

// commandPoints 建指揮點數視窗。
func (b *sceneBuilder) commandPoints() (origScreen, error) {
	if b.session == nil {
		return nil, fmt.Errorf("無對局")
	}
	return &commandPointsScreen{b: b}, nil
}

func (s *commandPointsScreen) update(in shell.InputState) *origTransition {
	// 原版是整螢幕隱形欄位 + ESC(0x1B):點哪裡都關。
	// ⚠ remake 的 `shell.InputState` 目前只帶滑鼠,沒有鍵盤欄位——ESC 那一半還沒接,
	// 接的話要動到共用的輸入結構,不在這一項的範圍。
	if in.ClickReleased {
		if clickSound != nil {
			clickSound()
		}
		return s.b.goTo(s.b.galaxy, "星圖")
	}
	return nil
}

func (s *commandPointsScreen) draw(dst *ebiten.Image) {
	b := s.b
	if b.fnt == nil || b.session == nil {
		return
	}
	sess := b.session

	// 背景:原版掛的是 `Draw_Mini_Main_Screen_`,remake 沒有迷你星圖這個變體,
	// 用同一張星圖(星空底 + 星球)當背景——語意一致:視窗浮在星圖上。
	b.drawStarmapBackground(dst)
	b.drawNebulae(dst, sess.Nebulae)
	drawStarmap(b, dst, b.fnt, sess.Stars, sess.SelectedStar, sess.VisibleStars())

	// 文字視窗。
	fillPanel(dst, cpPanelX, cpPanelY, cpPanelW, cpPanelH,
		color.RGBA{10, 14, 30, 238}, false)
	vector.StrokeRect(dst, cpPanelX, cpPanelY, cpPanelW, cpPanelH, 1,
		color.RGBA{90, 130, 200, 255}, false)

	gold := color.RGBA{240, 220, 120, 255}
	body := color.RGBA{214, 222, 238, 255}
	dim := color.RGBA{140, 152, 178, 255}
	warn := color.RGBA{235, 150, 140, 255}

	b.fnt.DrawCentered(dst, b.tr("指揮點數", "COMMAND POINTS"),
		cpPanelX+cpPanelW/2, cpPanelY+22, 15, gold)

	// 四個欄位對應原版的四條訊息(見檔頭)。
	base := gamedata.CommandPointsBase
	fromBldg := 0
	for i := range sess.PlayerColonies {
		if i < len(sess.ColonyBuildings) {
			fromBldg += gamedata.CommandPointsFromBuildings(sess.ColonyBuildings[i])
		}
	}
	// ⚠ 用現算值,不是 `Player.CommandPointsSupply` 那個快取欄位——那個只在 EndTurn 更新,
	// 開局或剛蓋好星基時是舊值,會顯示「起始 5 + 建築 1,總計卻是 1」這種自相矛盾的組合。
	total := sess.CommandPointsSupplyNow()
	used := sess.CommandPointsUsedNow()

	y := cpPanelY + 56
	row := func(label string, v int, c color.Color) {
		b.fnt.Draw(dst, label, cpPanelX+22, float64(y), 12, dim)
		b.fnt.Draw(dst, fmt.Sprintf("%d", v), float64(cpPanelX+cpPanelW-60), float64(y), 13, c)
		y += 26
	}
	row(b.tr("起始指揮點數", "Starting Command Points"), base, body)
	row(b.tr("軌道基地提供", "From orbital bases"), fromBldg, body)
	row(b.tr("指揮點數總計", "Total Command Points"), total, body)
	row(b.tr("已使用", "Total Command Points Used"), used, body)

	// 淨值 + 超額懲罰。手冊 p.37:每超額 1 點,每回合 −1 BC。
	net := total - used
	c := body
	if net < 0 {
		c = warn
	}
	y += 18
	vector.StrokeLine(dst, float32(cpPanelX+20), float32(y-12),
		float32(cpPanelX+cpPanelW-20), float32(y-14), 1, color.RGBA{60, 78, 118, 255}, false)
	b.fnt.Draw(dst, b.tr("淨餘", "Net"), cpPanelX+22, float64(y), 13, gold)
	b.fnt.Draw(dst, fmt.Sprintf("%d", net), float64(cpPanelX+cpPanelW-60), float64(y), 14, c)
	if net < 0 {
		y += 24
		b.fnt.Draw(dst, fmt.Sprintf(b.tr("超額 %d 點 → 每回合 −%d BC", "Over by %d → −%d BC/turn"), -net, -net),
			cpPanelX+22, float64(y), 11, warn)
	}

	b.fnt.DrawCentered(dst, b.tr("點一下關閉", "click to close"),
		cpPanelX+cpPanelW/2, cpPanelY+cpPanelH-16, 11, dim)
}
