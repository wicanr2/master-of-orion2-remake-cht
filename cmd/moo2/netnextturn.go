package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/netplay"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

// netnextturn.go:網路對戰的**等待其他玩家**畫面(原版 `Net_Next_Turn_` @ 0xFC470,
// 版面在 `Load_Net_Next_Turn_Screen_` @ 0xF3E42 / `Draw_Net_Next_Turn_Screen_` @ 0xF1075)。
//
// 這是網路對局裡玩家真的會盯著看的那一張:自己這回合下完令之後,等其他人。
// 規則面(誰到齊了、有沒有分岔)在 internal/netplay,這一檔只負責把它畫出來。
//
// ============ 版面是**算出來的**,不是量的 ============
//
// 這張畫面與 remake 先前移植的那些不一樣:它的座標不是立即數,而是
// `Load_Net_Next_Turn_Screen_` 依**資產尺寸**現算的。把那段組語翻成算式:
//
//	x = (0x280 − 資產42.寬) / 2                     ; 水平置中於 640
//	總高 = 資產42.高 + 資產43.高 + 資產40.高
//	y = max(0, (0x1E0 − 總高) / 2)                  ; 垂直置中於 480
//	[win+0xBF] = 資產42.高 + 資產43.高               ; 第三塊的相對位移
//	[win+0x10E] = 2                                 ; 字型 id
//
// 代進實際資產尺寸(lbxinfo 量的:42 = 630×48 十幀、43 = 630×179、40 = 630×221):
//
//	x = (640 − 630)/2 = 5
//	總高 = 48 + 179 + 221 = 448
//	y = (480 − 448)/2 = 16
//	→ 標題帶 (5,16)、中段 (5,64)、下段 (5,243)
//
// `Add_Net_Next_Turn_Fields_` @ 0xEFCEA 再給兩個數:
//
//	輸入列 y = y + [win+0xBF] + 0xBB = 16 + 227 + 187 = 430,高 0x11 = 17
//	玩家列的間距 0x19 = 25(`add [ebp+var_10], 19h`)
//
// ⚠ **玩家列的起始 y 沒有算出來**:那一段把顏色與字串交給
// `Get_Net_Next_Turn_Player_Colors_` @ 0xF31BB 與繪字常式,起點藏在 window 結構的欄位裡,
// 沒有直接的立即數。這裡取中段面板內縮一段當起點,**間距用真值 25**。
// 追出那個欄位之後換掉即可。
//
// ============ 誠實留白 ============
//
// 標題帶(資產 42)是 10 幀動畫,`Draw_Net_Next_Turn_Screen_` 用 `[win+0x1EA]` 當幀號逐幀播、
// 播完歸零。這裡照做。
//
// 其餘 8 個網路畫面:`Modem_Setup` / `NullModem_Setup` / `Comm Info` 是**數據機與序列線**的
// 設定,那些硬體現在不存在,remake 走 TCP——替不存在的硬體做設定畫面不是還原,是裝飾。
// `Join_Net` / `Choose_Net_Plyrs` / `Choose_Multi_Net_Game` / `Generic_Net_Info` /
// `SendGet_Net_Info` 是連線流程的畫面,要等 UI 端的連線流程做出來才有東西可顯示。

const (
	nntLBX = "multigm.lbx"

	nntBannerAsset = 42 // 630×48,10 幀動畫,自帶調色盤
	nntMidAsset    = 43 // 630×179
	nntBottomAsset = 40 // 630×221
	nntLightAsset  = 45 // 26×29,逐玩家的狀態燈

	// 以下四個由上面那段算式算出來(見檔頭)。
	nntX       = 5
	nntBannerY = 16
	nntMidY    = 64
	nntBotY    = 243

	// 真值:輸入列 y 與高、玩家列間距。
	nntInputY   = 430
	nntInputH   = 17
	nntRowStep  = 25
	nntRowFirst = 104 // ⚠ 估計值(中段面板內縮),見檔頭

	// 標題帶動畫:每幀停幾次重繪(同 starsprite.go 的黑洞,原版每次繪製推進一幀)。
	nntBannerHold = 4
)

// netNextTurnScreen 是等待其他玩家的畫面。
type netNextTurnScreen struct {
	b     *sceneBuilder
	table *netplay.Table
	// names 是各玩家的顯示名(索引 = 玩家編號)。
	names []string
	// me 是本方玩家編號(那一列標成自己)。
	me int

	bg, mid, bottom, light *ebiten.Image
	bannerFrames           []*ebiten.Image
	tick                   int
}

// netNextTurn 建等待畫面。table 為 nil 時畫成「沒有進行中的網路對局」。
func (b *sceneBuilder) netNextTurn(table *netplay.Table, names []string, me int) *netNextTurnScreen {
	s := &netNextTurnScreen{b: b, table: table, names: names, me: me}
	s.bg = b.multigmImage(mpBGAsset, false)
	s.mid = b.multigmImage(nntMidAsset, true)
	s.bottom = b.multigmImage(nntBottomAsset, true)
	s.light = b.multigmImage(nntLightAsset, true)
	// 標題帶自帶調色盤(flags 0x1000),而且有 10 幀——逐幀解出來快取。
	for f := 0; ; f++ {
		im := b.multigmFrame(nntBannerAsset, f)
		if im == nil {
			break
		}
		s.bannerFrames = append(s.bannerFrames, im)
	}
	return s
}

// bannerFrame 回傳這一刻要畫的標題帶幀(沒有資產時回 nil)。
func (s *netNextTurnScreen) bannerFrame() *ebiten.Image {
	if len(s.bannerFrames) == 0 {
		return nil
	}
	return s.bannerFrames[(s.tick/nntBannerHold)%len(s.bannerFrames)]
}

func (s *netNextTurnScreen) update(in shell.InputState) *origTransition {
	s.tick++
	// 這張畫面在原版是**等待**——玩家不能操作,只能等對手。remake 多給一條退路:
	// 點一下回多人設定畫面,否則連線斷了就卡死在這裡。
	if in.ClickReleased {
		sc, err := s.b.multiPlayer()
		if err != nil {
			return nil // 建不起來就留在原畫面,不要把玩家丟到黑畫面
		}
		return &origTransition{next: sc}
	}
	return nil
}

// statusLine 回傳某位玩家的狀態文字。
func (s *netNextTurnScreen) statusLine(player int) (string, color.RGBA) {
	if s.table == nil {
		return s.b.tr("(無對局)", "(no game)"), color.RGBA{140, 150, 170, 255}
	}
	for _, m := range s.table.Missing() {
		if m == player {
			return s.b.tr("等待中…", "waiting…"), color.RGBA{235, 200, 120, 255}
		}
	}
	return s.b.tr("已完成", "ready"), color.RGBA{140, 225, 160, 255}
}

func (s *netNextTurnScreen) draw(dst *ebiten.Image) {
	dst.Fill(color.RGBA{6, 8, 14, 255})
	blit := func(im *ebiten.Image, x, y int) {
		if im == nil {
			return
		}
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(x), float64(y))
		dst.DrawImage(im, op)
	}
	blit(s.bg, 0, 0)
	blit(s.bannerFrame(), nntX, nntBannerY)
	blit(s.mid, nntX, nntMidY)
	blit(s.bottom, nntX, nntBotY)

	if s.b.fnt == nil {
		return
	}
	gold := color.RGBA{240, 220, 120, 255}
	body := color.RGBA{215, 222, 238, 255}

	// 標題:原版的標題帶烘著 "NETWORK PLAYERS",中文模式先擦一塊再疊字
	// (不擦的話兩層字疊在一起,那是 loadgame/gamemenu/confirmbox 都處理過的同一件事)。
	if s.b.lang == i18n.Traditional {
		vector.DrawFilledRect(dst, 170, float64Rect(nntBannerY+12), 300, 26,
			color.RGBA{26, 30, 38, 255}, false)
		s.b.fnt.DrawCentered(dst, "等待其他玩家", 320, float64(nntBannerY)+30, 18, gold)
	}

	turn := 0
	if s.table != nil {
		turn = s.table.Turn()
	}

	for i := 0; i < len(s.names); i++ {
		y := nntRowFirst + i*nntRowStep
		if y > nntBotY-nntRowStep {
			break // 中段面板放不下就不畫,不要溢到下一塊面板上
		}
		// 狀態燈(資產 45)畫在名字左邊。
		blit(s.light, nntX+20, y-16)
		name := s.names[i]
		if i == s.me {
			name += s.b.tr("(你)", " (you)")
		}
		s.b.fnt.Draw(dst, name, nntX+56, float64(y), 13, body)
		text, col := s.statusLine(i)
		s.b.fnt.Draw(dst, text, nntX+300, float64(y), 13, col)
	}

	// 下段面板:回合數與狀態指紋。指紋擺在畫面上不是裝飾——分岔時兩邊念一下這八個字元
	// 就知道是不是同一個狀態,不必先架 log 收集。
	s.b.fnt.Draw(dst, fmt.Sprintf(s.b.tr("第 %d 回合", "Turn %d"), turn),
		nntX+24, float64(nntBotY)+26, 14, gold)
	if s.b.session != nil {
		s.b.fnt.Draw(dst, s.b.tr("狀態指紋:", "State fingerprint: ")+s.b.session.StateFingerprint(),
			nntX+24, float64(nntBotY)+48, 12, color.RGBA{170, 200, 230, 255})
	}

	// 分岔警告:鎖步一旦分岔,繼續玩只會讓兩邊差得更遠,所以講得大聲一點。
	if s.table != nil {
		if d := s.table.Desync(); d != "" {
			vector.DrawFilledRect(dst, nntX+16, nntBotY+70, 598, 44, color.RGBA{70, 16, 20, 235}, false)
			vector.StrokeRect(dst, nntX+16, nntBotY+70, 598, 44, 1, color.RGBA{230, 110, 110, 255}, false)
			s.b.fnt.Draw(dst, s.b.tr("⚠ 狀態分岔——對局已不同步,請停止", "⚠ Desync — the game is out of sync, stop"),
				nntX+26, float64(nntBotY)+90, 13, color.RGBA{250, 190, 180, 255})
			s.b.fnt.Draw(dst, d, nntX+26, float64(nntBotY)+108, 11, color.RGBA{240, 200, 195, 255})
		}
	}

	// 輸入列(原版在 y=430 有一個文字欄位;remake 沒有聊天,畫成提示帶並標明)。
	vector.DrawFilledRect(dst, nntX+16, nntInputY, 598, nntInputH, color.RGBA{18, 24, 38, 220}, false)
	s.b.fnt.Draw(dst, s.b.tr("點一下離開等待畫面(⚠ 原版此列是聊天輸入,remake 未實作)",
		"Click to leave (⚠ the original has a chat field here; not implemented)"),
		nntX+22, float64(nntInputY)+13, 11, color.RGBA{150, 162, 185, 255})
}

// float64Rect 是 vector 那組 API 要 float32、而這一檔的座標常數是 int 的轉接。
func float64Rect(v int) float32 { return float32(v) }

// multigmImage 解 MULTIGM.LBX 的某資產第 0 幀(調色盤一律借背景資產 0 自帶的那份)。
//
// 抽出來是因為這一檔要解四五張圖,而 multiplayer.go 那邊是就地解在建構子裡的。
func (b *sceneBuilder) multigmImage(assetID int, keyColor bool) *ebiten.Image {
	return b.multigmFrameWithKey(assetID, 0, keyColor)
}

// multigmFrame 解某資產的第 f 幀(超出幀數回 nil,呼叫端據此判斷「畫完了」)。
func (b *sceneBuilder) multigmFrame(assetID, f int) *ebiten.Image {
	return b.multigmFrameWithKey(assetID, f, false)
}

func (b *sceneBuilder) multigmFrameWithKey(assetID, f int, keyColor bool) *ebiten.Image {
	prov, err := decodeAsset(b.res, nntLBX, mpBGAsset)
	if err != nil || prov.Embedded == nil {
		return nil
	}
	im, err := decodeAsset(b.res, nntLBX, assetID)
	if err != nil || f >= len(im.Frames) {
		return nil
	}
	// 自帶調色盤的資產(如標題帶)用自己的,其餘借背景那份。
	pal := prov.Embedded
	if im.Embedded != nil {
		pal = im.Embedded
	}
	return ebiten.NewImageFromImage(im.Frames[f].ToRGBA(pal, keyColor))
}

// netNextTurnDemo 建一張「示範用」的等待畫面:表格是即時建的,人名取目前對局的帝國名。
//
// ⚠ 它**不是**連上線的對局——連線流程的 UI 還沒做(見檔頭)。這張畫面存在的意義是
// 版面與狀態顯示先做對:等連線流程接上來時,只要把這裡的 table 換成真的那一張即可。
func (b *sceneBuilder) netNextTurnDemo() *netNextTurnScreen {
	names := []string{b.tr("玩家", "Player")}
	if b.session != nil {
		if b.session.PlayerName != "" {
			names[0] = b.session.PlayerName
		}
		for i := range b.session.AIPlayers {
			names = append(names, b.session.AIPlayers[i].Name)
		}
	}
	if len(names) > 4 {
		names = names[:4]
	}
	turn := 1
	if b.session != nil {
		turn = b.session.Turn
	}
	tb := netplay.NewTable(len(names), turn)
	// 只有自己送了 turn_done —— 那正是「等其他人」的狀態。
	hash := ""
	if b.session != nil {
		hash = b.session.StateHash()
	}
	_ = tb.Add(netplay.Message{Kind: netplay.KindTurnDone, Player: 0, Turn: turn, StateHash: hash})
	return b.netNextTurn(tb, names, 0)
}
