package main

import (
	"fmt"
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/netplay"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

// choosenetplyrs.go:連線大廳的**玩家名冊**畫面
//(原版 `Choose_Network_Plyrs_Screen_` @ 0xF0E17,版面在 `Load_Choose_Net_Plyrs_Screen_`
// @ 0xF3FC6 + `Add_Choose_Net_Plyrs_Fields_` @ 0xEFB50)。
//
// 大廳本身在 `internal/netplay/lobby.go`(主機聽、客戶端連、主機廣播名冊);
// 這一檔只負責把名冊畫出來。
//
// ============ 版面同樣是**算**出來的,而且面板會隨人數長高 ============
//
// `Choose_Network_Plyrs_Screen_` 的定位段:
//
//	x   = (0x280 − 資產27.寬) / 2
//	總高 = 資產28.高 × [win+0x1E1] − 1 + 資產27.高 + 資產29.高
//	y   = (0x1E0 − 總高) / 2
//
// `[win+0x1E1]` 是列數——**中段面板(資產 28)每位玩家重複一次**,所以整個視窗會隨人數長高。
// 這是 remake 至今移植的畫面裡第一個「尺寸隨資料變」的版面。
//
// 資產(lbxinfo 量的):27 = 479×81(10 幀動畫)、28 = 479×36、29 = 479×38。
//
//	x = (640 − 479)/2 = 80
//	4 人:總高 = 36×4 − 1 + 81 + 38 = 262 → y = (480 − 262)/2 = 109
//	8 人:總高 = 36×8 − 1 + 81 + 38 = 406 → y = (480 − 406)/2 = 37
//
// `Add_Choose_Net_Plyrs_Fields_` 給每一列的**點擊區**(逐項都是立即數):
//
//	x1 = [win+0xF3] + 0x6A   (+106)
//	y1 = [win+0xF5] + (i × 0x24 + 0x40)   (+ i×36 + 64)
//	x2 = [win+0xF3] + 0x1B3  (+435)
//	y2 = y1 + 0x1D           (+29)
//
// 也就是每列 **329×29**,列距 36(= 資產 28 的高,對得上)。
// `[win+0xBB + i×2]` 是逐玩家的 widget id 陣列,初值 0xFC18(無效)——**最多 8 列**。
//
// ============ 誠實留白 ============
//
// 原版這張畫面可以點列來指派種族/顏色(`sub_EFABA` 在每列旁再建一組欄位)。
// remake 的大廳只做到「誰連進來了」,種族選擇仍走既有的單機流程——
// 那一段要等連線流程把種族選擇也納進來才有意義,不先做半套。

const (
	cnpBannerAsset = 27 // 479×81,10 幀
	cnpRowAsset    = 28 // 479×36,每位玩家一片
	cnpFootAsset   = 29 // 479×38

	cnpBannerW, cnpBannerH = 479, 81
	cnpRowH                = 36
	cnpFootH               = 38

	cnpMaxRows = 8 // [win+0xBB] 那個 widget id 陣列的長度

	// 每一列點擊區(相對視窗左上,取自 Add_Choose_Net_Plyrs_Fields_ 的立即數)。
	cnpRowX1    = 0x6A  // +106
	cnpRowX2    = 0x1B3 // +435
	cnpRowFirst = 0x40  // +64
	cnpRowStep  = 0x24  // 36
	cnpRowH2    = 0x1D  // 29
)

// chooseNetPlayersScreen 是名冊畫面。
type chooseNetPlayersScreen struct {
	b      *sceneBuilder
	roster netplay.Roster
	me     int
	// hosting 決定底下那行提示要講「等其他人加入」還是「等主機開始」。
	hosting bool
	addr    string

	bg, row, foot *ebiten.Image
	banner        []*ebiten.Image
	tick          int
	// lobby 非 nil 時每幀重讀名冊(主機端;客戶端的名冊是連上時一次拿到的)。
	lobby *netplay.Lobby
}

// cnpWindow 依人數算出視窗左上角(見檔頭的算式)。
//
// 抽成函式是為了讓測試能重算一次——這張畫面的座標會隨人數變,寫死常數就測不到了。
func cnpWindow(rows int) (x, y int) {
	if rows < 1 {
		rows = 1
	}
	total := cnpRowH*rows - 1 + cnpBannerH + cnpFootH
	return (moo2ScreenW - cnpBannerW) / 2, (moo2ScreenH - total) / 2
}

// cnpInfoBaselines 回傳大廳狀態兩行字的基線 y(位址 / 種子)。
//
// 這兩行**畫在視窗外面、底框下方**。第一版畫在底框(資產 29)裡面,結果第一行壓在
// 那圈金屬圓角上、第二行掉到視窗外——底框那 38 px 的可見內容只有頂端那條邊框,
// 底下是透明的,不是可以擺字的空間。截圖看出來的,不是讀程式看出來的。
func cnpInfoBaselines(winY, rows int) (y1, y2 int) {
	bottom := winY + cnpBannerH + rows*cnpRowH + cnpFootH
	return bottom + 12, bottom + 27
}

// cnpRowRect 回傳第 i 列的點擊區(螢幕座標)。
func cnpRowRect(winX, winY, i int) (x, y, w, h int) {
	x1 := winX + cnpRowX1
	y1 := winY + i*cnpRowStep + cnpRowFirst
	return x1, y1, (winX + cnpRowX2) - x1, cnpRowH2
}

func (b *sceneBuilder) chooseNetPlayers(r netplay.Roster, me int, hosting bool, addr string) *chooseNetPlayersScreen {
	s := &chooseNetPlayersScreen{b: b, roster: r, me: me, hosting: hosting, addr: addr}
	s.bg = b.multigmImage(mpBGAsset, false)
	s.row = b.multigmImage(cnpRowAsset, true)
	s.foot = b.multigmImage(cnpFootAsset, true)
	for f := 0; ; f++ {
		im := b.multigmFrame(cnpBannerAsset, f)
		if im == nil {
			break
		}
		s.banner = append(s.banner, im)
	}
	return s
}

// chooseNetPlayersDemo 是截圖廊用的名冊(不開 socket)。
//
// 截圖廊不能真的開一個聽 port 的大廳:它跑在 docker 裡、而且一次跑完就結束,
// 留一個半開的 listener 給後續的截圖沒有意義。名冊的**畫面**與資料從哪來無關。
func (b *sceneBuilder) chooseNetPlayersDemo() *chooseNetPlayersScreen {
	r := netplay.Roster{Seed: 20260807, Players: []netplay.Player{
		{ID: 0, Name: b.tr("指揮官", "Commander")},
		{ID: 1, Name: "Sakkra"},
		{ID: 2, Name: "Psilon"},
	}}
	return b.chooseNetPlayers(r, 0, true, "192.168.1.20:24501")
}

func (s *chooseNetPlayersScreen) update(in shell.InputState) *origTransition {
	s.tick++
	if s.lobby != nil {
		s.roster = s.lobby.Roster() // 背景 goroutine 一直在收人,畫面每幀重讀
	}
	if !in.ClickReleased {
		return nil
	}
	// remake 的大廳沒有「點列指派種族」(見檔頭),所以點任何地方都是離開。
	sc, err := s.b.multiPlayer()
	if err != nil {
		return nil
	}
	return &origTransition{next: sc}
}

func (s *chooseNetPlayersScreen) draw(dst *ebiten.Image) {
	dst.Fill(color.RGBA{6, 8, 14, 255})
	blit := func(im *ebiten.Image, x, y int) {
		if im == nil {
			return
		}
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(x), float64(y))
		drawPanelImage(dst, im, op)
	}
	blit(s.bg, 0, 0)

	rows := len(s.roster.Players)
	if rows < 1 {
		rows = 1
	}
	if rows > cnpMaxRows {
		rows = cnpMaxRows
	}
	winX, winY := cnpWindow(rows)

	if len(s.banner) > 0 {
		blit(s.banner[(s.tick/4)%len(s.banner)], winX, winY)
	}
	for i := 0; i < rows; i++ {
		blit(s.row, winX, winY+cnpBannerH+i*cnpRowH)
	}
	blit(s.foot, winX, winY+cnpBannerH+rows*cnpRowH)

	if s.b.fnt == nil {
		return
	}
	gold := color.RGBA{240, 220, 120, 255}
	body := color.RGBA{215, 222, 238, 255}
	dim := color.RGBA{150, 162, 185, 255}

	if s.b.lang == i18n.Traditional {
		fillPanel(dst, float32(winX+120), float32(winY+22), 240, 26,
			color.RGBA{26, 30, 38, 255}, false)
		s.b.fnt.DrawCentered(dst, "連線玩家", float64(winX+cnpBannerW/2), float64(winY)+40, 18, gold)
	}

	for i := 0; i < rows; i++ {
		x, y, w, _ := cnpRowRect(winX, winY, i)
		name := s.b.tr("(空位)", "(open)")
		col := dim
		if i < len(s.roster.Players) {
			name = s.roster.Players[i].Name
			col = body
			if i == s.me {
				name += s.b.tr("(你)", " (you)")
			}
			if i == 0 {
				name += s.b.tr("・主機", " · host")
			}
		}
		s.b.fnt.Draw(dst, fmt.Sprintf("%d.", i+1), float64(x-24), float64(y)+20, 13, dim)
		s.b.fnt.Draw(dst, name, float64(x+8), float64(y)+20, 13, col)
		_ = w
	}

	hint := s.b.tr("等待主機開始對局", "Waiting for the host to start")
	if s.hosting {
		hint = s.b.tr("等其他玩家加入——位址:", "Waiting for players — address: ") + s.addr
	}
	seed := fmt.Sprintf(s.b.tr("種子 %d(由主機決定並廣播)", "Seed %d (chosen and broadcast by the host)"),
		s.roster.Seed)
	y1, y2 := cnpInfoBaselines(winY, rows)
	fillPanel(dst, float32(winX+8), float32(y1-13), float32(cnpBannerW-16), float32(y2-y1+18),
		color.RGBA{10, 12, 18, 235}, false)
	s.b.fnt.Draw(dst, hint, float64(winX+16), float64(y1), 12, body)
	s.b.fnt.Draw(dst, seed, float64(winX+16), float64(y2), 11, dim)
}

// --- 大廳的開/加入(把 internal/netplay 的大廳接到畫面上)---

// netLobbyAddr 是預設的大廳位址。
//
// 固定 port 而不是隨機:區網對戰時對方要打得出這個位址,隨機 port 等於每次都要口頭傳一次。
// 24501 沒有被 IANA 指派給任何服務,撞到的機會低。
const netLobbyAddr = "0.0.0.0:24501"

// netLobbyDialAddr 是「加入」時的**後備**目標。
//
// 正常路徑是區網探索(見 choosemultinetgame.go):主機廣播、客戶端列出來點一場。
// 探索開不起來(埠被佔、沒有網路)時才退回這個位址,至少同一台機器上測得動。
// 要打任意位址仍然需要文字輸入框——那是還沒做的一項,不是這裡偷懶。
const netLobbyDialAddr = "127.0.0.1:24501"

// hostNetLobby 先問對局名稱,再開大廳。
//
// 先問名稱是原版的順序:`Multi_Player_Screen_` @ 0xF4D99 在開局前呼叫
// `Change_MP_Game_Name_` @ 0xF5777(長度上限 8、且要與既有對局不同名)。
// 名稱是別人在清單上看到的東西,開完局才改就晚了。
func (b *sceneBuilder) hostNetLobby() (origScreen, error) {
	def := b.tr("主機玩家", "Host")
	if b.session != nil && b.session.PlayerName != "" {
		def = b.session.PlayerName
	}
	if r := []rune(def); len(r) > netplay.GameNameMax {
		def = string(r[:netplay.GameNameMax])
	}
	under, err := b.multiPlayer()
	if err != nil {
		return nil, err
	}
	return b.inputBox(under, b.tr("對局名稱", "Game name"), def, netplay.GameNameMax,
		func(name string) *origTransition {
			sc, err := b.startNetLobby(name)
			if err != nil {
				return nil // 開不起來就留在輸入框上,不要無聲跳走
			}
			return &origTransition{next: sc}
		}), nil
}

// startNetLobby 用指定的對局名稱開大廳並進「等待其他玩家加入」。
//
// ⚠ 這裡**不阻塞等人加入**:UI 是單執行緒的,`AcceptOne` 會把整個畫面凍住。
// 收人放在背景 goroutine,畫面每幀重讀 `lobby.Roster()`。
func (b *sceneBuilder) startNetLobby(name string) (origScreen, error) {
	seed := int64(1)
	if b.session != nil {
		seed = b.session.EventSeed
	}
	if name == "" {
		name = b.tr("主機玩家", "Host")
	}
	lb, err := netplay.Host(netLobbyAddr, name, seed)
	if err != nil {
		return nil, err
	}
	b.netLobby = lb
	b.netMe = 0 // 主機恆為名冊上的第 0 位
	// 一併廣播,否則區網上的人看不到這場對局(原版靠 IPX 的服務公告,
	// TCP 沒有那個能力——見 internal/netplay/discovery.go)。
	gameName := name
	if r := []rune(gameName); len(r) > netplay.GameNameMax {
		gameName = string(r[:netplay.GameNameMax])
	}
	if an, err := netplay.Announce(netplay.Game{
		Name: gameName, Addr: lb.Addr(), Players: 1, Max: cnpMaxRows,
	}, "", time.Second); err == nil {
		b.netAnnouncer = an
	}
	go func() {
		// 收到上限或大廳關掉為止。錯誤不往上拋——UI 端看得到的是「名冊有沒有多一個人」,
		// 而 AcceptOne 的錯誤多半是「還沒有人來」的逾時。
		for i := 1; i < cnpMaxRows; i++ {
			if _, err := lb.AcceptOne(0); err != nil {
				return
			}
			// 人數變了要更新廣播內容,不然清單上永遠寫 1 人。
			if b.netAnnouncer != nil {
				b.netAnnouncer.Update(netplay.Game{
					Name: gameName, Addr: lb.Addr(),
					Players: len(lb.Roster().Players), Max: cnpMaxRows,
				})
			}
		}
	}()
	// 原版的順序是**先等人、再指派**:`Reload_Waiting_For_Joiners_Screen_` 是主機開局後
	// 看到的那一張(人數欄位由 `Add_Waiting_For_Joiners_Field_` 逐幀更新),
	// 點過去才是 `Choose_Net_Plyrs` 的名冊。照這個順序接。
	wait := b.netInfo(netInfoWaitingForJoiners)
	wait.lobby, wait.total, wait.hosting = lb, cnpMaxRows, true
	wait.onCancel = func() *origTransition {
		sc := b.chooseNetPlayers(lb.Roster(), 0, true, lb.Addr())
		sc.lobby = lb
		return &origTransition{next: sc}
	}
	return wait, nil
}

// joinNetLobby 進「選擇要加入的對局」清單(原版 Choose_Multi_Net_Game)。
//
// 探索開不起來時退回直接連本機——沒有清單也要能測得動,總比一個什麼都不做的按鈕好。
func (b *sceneBuilder) joinNetLobby() (origScreen, error) {
	if br, err := netplay.Browse(""); err == nil {
		b.netBrowser = br
		return b.chooseMultiNetGame(br), nil
	}
	return b.joinNetGame(netplay.Game{Name: "local", Addr: netLobbyDialAddr})
}

// joinNetGame 連上指定的一場對局並進名冊畫面。
func (b *sceneBuilder) joinNetGame(g netplay.Game) (origScreen, error) {
	name := b.tr("玩家", "Player")
	if b.session != nil && b.session.PlayerName != "" {
		name = b.session.PlayerName
	}
	conn, me, roster, err := netplay.Join(g.Addr, name, 3*time.Second)
	if err != nil {
		return nil, err
	}
	b.netConn = conn
	b.netMe = me // 主機指派的編號——聊天要靠它標「這句是誰說的」
	return b.chooseNetPlayers(roster, me, false, g.Addr), nil
}

// netInfoJoining 這個狀態在 remake 裡**沒有停留的時機**:`netplay.Join` 是同步的,
// 連上或逾時都在同一個呼叫裡結束,沒有「連線中」那一段可以顯示。
// 原版有,是因為它的連線走 IPX / 數據機,協商要好幾秒。
// 版面與資產已經對好(見 netinfo.go),等哪天連線改成非同步就有觸發點——
// 現在硬插一張會是假的載入畫面。
