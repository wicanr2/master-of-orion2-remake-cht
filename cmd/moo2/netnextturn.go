package main

import (
	"fmt"
	"image/color"
	"net"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
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
// 玩家列的顏色 remake 沒有照原版的 `Get_Net_Next_Turn_Player_Colors_` @ 0xF31BB 取,
// 而是照「等待中 / 已完成」給兩色——原版那支是依帝國旗色配色,接上去要先有網路對局的旗色名冊。
//
// 剩下的網路畫面:`Modem_Setup` / `NullModem_Setup` / `Comm Info` 是**數據機與序列線**的設定,
// 那些硬體現在不存在,remake 走 TCP——替不存在的硬體做設定畫面不是還原,是裝飾。
// (`Choose_Net_Plyrs` / `Choose_Multi_Net_Game` / `Generic_Net_Info` / 輸入框已分別在
// choosenetplyrs.go、choosemultinetgame.go、netinfo.go、inputbox.go 做完。)

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

	// 聊天記錄區:面板是資產 40(nntBotY),偏移量在 internal/netplay/chat.go。
	nntChatX     = nntX + netplay.ChatTextDX     // 5 + 24 = 29
	nntChatFirst = nntBotY + netplay.ChatFirstDY // 243 + 14 = 257

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

	// chat 是聊天記錄,typing 是還沒送出的那一行。
	chat   netplay.ChatLog
	typing string

	// sess 是對局的連線幫浦。**可以是 nil**——單機開這張畫面(截圖廊、示範)時就是 nil,
	// 那時候聊天列仍然能打字,只是話不會離開本機。
	sess *netplay.Session
}

// attach 把連線幫浦接上這張畫面。sess 為 nil 時等同不接。
func (s *netNextTurnScreen) attach(sess *netplay.Session) { s.sess = sess }

// pumpChat 把幫浦這一幀收到的聊天訊息倒進記錄。
//
// **在 Update 這一條線上做**,不是在讀取 goroutine 裡——狀態變更的時機要與封包抵達的
// 時間點無關(見 netplay/session.go 檔頭)。
//
// 非聊天的訊息(turn_done / desync)這張畫面不處理,原樣丟掉:回合表的推進在別處,
// 在這裡順手動它會讓同一件事有兩個入口。
func (s *netNextTurnScreen) pumpChat() {
	if s.sess == nil {
		return
	}
	for _, m := range s.sess.Poll() {
		if m.Kind != netplay.KindChat {
			continue
		}
		s.chat.Append(m.Player, m.Text)
	}
}

// speakerName 回傳某位發話者要顯示的名字(GNN 不用名字)。
func (s *netNextTurnScreen) speakerName(speaker int) string {
	if speaker < 0 || speaker >= len(s.names) {
		return ""
	}
	return s.names[speaker]
}

// sendChat 把 typing 那一行送出去。
//
// 聊天**不走鎖步的回合表**:`netplay.Table` 收的是回合指令,一回合只收一則,而聊天是
// 隨時可送的。兩者共用同一條 TCP 連線,但在協定上是各自獨立的訊息種類。
//
// 本機記錄是**立刻**加的,不等封包繞一圈回來——打完 Enter 就該看到自己那一行。
// 幫浦那一端也因此刻意不把自己送出的訊息回填進收訊佇列(見 Session.Send)。
func (s *netNextTurnScreen) sendChat() {
	m := netplay.ChatMessage(s.me, s.typing)
	s.typing = ""
	if m.Text == "" {
		return // 原版按 Enter 但沒打字也是什麼都不做(`cmp byte_1AAC54, 0 / jz`)
	}
	s.chat.Append(m.Player, m.Text)
	if s.sess != nil {
		// 送不出去(線斷了)不打斷輸入:字已經在自己的記錄裡,而斷線本身由
		// Session.Err 那條線回報,不在這裡多開一個錯誤路徑。
		_ = s.sess.Send(m)
	}
}

// typeChatRunes 把字元加進輸入行,超過原版上限就丟掉(同 inputbox 的作法)。
func (s *netNextTurnScreen) typeChatRunes(rs []rune) {
	for _, r := range rs {
		if r < 0x20 || r == 0x7F {
			continue
		}
		if len(s.typing)+len(string(r)) > netplay.ChatTextMax {
			return
		}
		s.typing += string(r)
	}
}

// backspaceChat 刪掉輸入行最後一個字元(**一個 rune**,不是一個 byte)。
func (s *netNextTurnScreen) backspaceChat() {
	rs := []rune(s.typing)
	if len(rs) == 0 {
		return
	}
	s.typing = string(rs[:len(rs)-1])
}

// netNextTurn 建等待畫面。table 為 nil 時畫成「沒有進行中的網路對局」。
func (b *sceneBuilder) netNextTurn(table *netplay.Table, names []string, me int) *netNextTurnScreen {
	s := &netNextTurnScreen{b: b, table: table, names: names, me: me}
	s.attach(b.netSession())
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
	s.pumpChat()
	// 聊天列:原版這張畫面唯一能做的事就是打字給對手看(`Chat_Box_Input_Loop_` @ 0xF55A4)。
	s.typeChatRunes(ebiten.AppendInputChars(nil))
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		s.backspaceChat()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyKPEnter) {
		s.sendChat()
		return nil
	}
	// 這張畫面在原版是**等待**——除了聊天不能做別的。remake 多給一條退路:
	// 點一下回多人設定畫面,否則連線斷了就卡死在這裡。
	if in.ClickReleased {
		sc, err := s.b.multiPlayer()
		if err != nil {
			return nil // 建不起來就留在原畫面,不要把玩家丟到黑畫面
		}
		// 離開等待畫面等於退出這一局:幫浦連同它的讀取 goroutine 一起收掉,
		// 不然每開一次網路對局就多留一條在背景讀已經沒人看的連線。
		s.b.closeNetSession()
		s.sess = nil
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

	// 回合數與狀態指紋放**中段面板下緣**:下段面板(資產 40)整塊是原版的聊天記錄區,
	// 佔用它就得把聊天往別處挪,那就不是原版的版面了。
	// 指紋擺在畫面上不是裝飾——分岔時兩邊念一下這八個字元就知道是不是同一個狀態,
	// 不必先架 log 收集。
	s.b.fnt.Draw(dst, fmt.Sprintf(s.b.tr("第 %d 回合", "Turn %d"), turn),
		nntX+24, float64(nntMidY)+142, 14, gold)
	if s.b.session != nil {
		s.b.fnt.Draw(dst, s.b.tr("狀態指紋:", "State fingerprint: ")+s.b.session.StateFingerprint(),
			nntX+24, float64(nntMidY)+162, 12, color.RGBA{170, 200, 230, 255})
	}

	// 聊天記錄:下段面板由上而下 14 行,行距 12(見 internal/netplay/chat.go)。
	// 原版每行畫字前先拿資產 40 重畫那一列來擦背景;這裡整張面板每幀重畫,效果相同。
	gnn := color.RGBA{235, 215, 150, 255}
	for i, ln := range s.chat.Lines() {
		if i >= netplay.ChatLogMax {
			break
		}
		col := body
		if ln.IsGNN() {
			col = gnn
		}
		s.b.fnt.Draw(dst, netplay.ChatPrefix(ln.Speaker, s.speakerName(ln.Speaker))+ln.Text,
			nntChatX, float64(nntChatFirst+i*netplay.ChatLineStep), 11, col)
	}

	// 分岔警告:鎖步一旦分岔,繼續玩只會讓兩邊差得更遠,所以蓋在聊天記錄上面也要講。
	if s.table != nil {
		if d := s.table.Desync(); d != "" {
			vector.DrawFilledRect(dst, nntX+16, nntBotY+137, 598, 44, color.RGBA{70, 16, 20, 235}, false)
			vector.StrokeRect(dst, nntX+16, nntBotY+137, 598, 44, 1, color.RGBA{230, 110, 110, 255}, false)
			s.b.fnt.Draw(dst, s.b.tr("⚠ 狀態分岔——對局已不同步,請停止", "⚠ Desync — the game is out of sync, stop"),
				nntX+26, float64(nntBotY)+157, 13, color.RGBA{250, 190, 180, 255})
			s.b.fnt.Draw(dst, d, nntX+26, float64(nntBotY)+175, 11, color.RGBA{240, 200, 195, 255})
		}
	}

	// 輸入列(原版 y=430、高 0x11,前綴同樣是 `"(%s)  "` 配本方玩家名)。
	vector.DrawFilledRect(dst, nntX+16, nntInputY, 598, nntInputH, color.RGBA{18, 24, 38, 220}, false)
	caret := ""
	if (s.tick/30)%2 == 0 {
		caret = "_"
	}
	s.b.fnt.Draw(dst, netplay.ChatPrefix(s.me, s.speakerName(s.me))+s.typing+caret,
		nntChatX, float64(nntInputY)+13, 11, color.RGBA{225, 232, 245, 255})
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
	// ⚠ **累積**到第 f 幀,不是直接把第 f 幀上色。
	//
	// MULTIGM 的動畫面板是 delta 幀:第 0 幀是完整畫面,之後每幀只帶會閃的那幾顆燈。
	// 逐幀獨立上色的話,第 1 幀開始整張面板就消失了,只剩幾個小亮點。
	// 這個 bug 先前沒被發現,是因為截圖廊每張都恰好落在第 0 幀——直到「等待其他玩家加入」
	// 那張刻意多留幾拍才露出來(見 internal/lbx 的 AccumulatedUpToRGBA)。
	return ebiten.NewImageFromImage(im.AccumulatedUpToRGBA(pal, f, keyColor))
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
	s := b.netNextTurn(tb, names, 0)
	// 聊天記錄先放三則,不然截圖廊那張下段面板是空的——看不出版面對不對。
	// 兩種前綴各出現一次(玩家 / GNN),那是原版分兩路的地方。
	s.chat.AppendGNN(b.tr("議會將於下回合開議。", "The council convenes next turn."))
	if len(names) > 1 {
		s.chat.Append(1, b.tr("我這回合下完了,等你。", "Done here, waiting on you."))
	}
	s.chat.Append(0, b.tr("再給我一回合。", "One more turn."))
	return s
}

// netSession 回傳這一局的訊息幫浦,第一次呼叫時才建。
//
// **為什麼是延遲建立**:大廳階段還在收人(`startNetLobby` 的背景 goroutine),
// 那時候建會漏掉後來加入的玩家。等到真的要等回合時,名冊已經定了。
//
// ⚠ 誠實限制:幫浦建好之後**再加入的人不會被納進來**。remake 的對局在開打後不收人,
// 所以現階段碰不到;真要支援中途加入,這裡要改成可增減連線。
//
// 沒有任何連線(單機、截圖廊)時回 nil——呼叫端都容忍 nil。
func (b *sceneBuilder) netSession() *netplay.Session {
	if b.netSess != nil {
		return b.netSess
	}
	switch {
	case b.netLobby != nil:
		// 主機端:名冊上除了自己以外的每個人各一條連線,而且要轉發
		// (星狀拓樸下客戶端之間沒有連線,見 netplay/session.go 檔頭)。
		conns := map[int]net.Conn{}
		for _, p := range b.netLobby.Roster().Players {
			if p.ID == b.netMe {
				continue
			}
			if c := b.netLobby.Conn(p.ID); c != nil {
				conns[p.ID] = c
			}
		}
		if len(conns) == 0 {
			return nil // 還沒有人加入——不要建一個空幫浦擋住之後真的建得起來的那次
		}
		b.netSess = netplay.NewSession(b.netMe, true, conns)
	case b.netConn != nil:
		// 客戶端:只有一條連線(對主機),不轉發。
		b.netSess = netplay.NewSession(b.netMe, false, map[int]net.Conn{0: b.netConn})
	default:
		return nil
	}
	return b.netSess
}

// closeNetSession 收掉幫浦(連同它的讀取 goroutine)。可重複呼叫。
//
// 底層的 net.Conn 也會一起關掉——`Session.Close` 關的就是傳進去的那些連線。
func (b *sceneBuilder) closeNetSession() {
	if b.netSess == nil {
		return
	}
	_ = b.netSess.Close()
	b.netSess = nil
	b.netConn = nil
}
