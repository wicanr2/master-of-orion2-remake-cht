package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/netplay"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

// netinfo.go:連線流程的**狀態面板**(原版 `Reload_Generic_Net_Info_` @ 0xF53D7
// + `Draw_Generic_Net_Info_Screen_` @ 0xF19C7 + `Draw_SendGet_Net_Info_Screen_` @ 0xF2C8B)。
//
// ============ 反組譯先推翻了「還缺幾張畫面」這個數字 ============
//
// gap report 第 29 項(決定性化)結尾說「剩 5 張」,第 29 項(決定性化)做掉一張說「剩 4 張」。抽版面時發現
// 那 4 張其實是 **2 張**:
//
//	0xF19C7  Draw_Generic_Net_Info_Screen_
//	0xF19C7  Draw_Join_Net_Screen_          ← 同一個位址
//
// 兩個名字指向同一段程式碼。再往上追,`Reload_Generic_Net_Info_` 收一個**資產編號**當參數,
// 而下面這一整排都只是「帶不同資產編號呼叫它」:
//
//	| 呼叫者                             | 資產 | 尺寸(lbxinfo) | 意思 |
//	|---|---|---|---|
//	| `Reload_Waiting_For_Joiners_Screen_` @ 0xF552A | 0x0F=15 | 479×150,10 幀 | 等其他玩家加入 |
//	| `Reload_Join_Net_Screen_`           @ 0xF54CF | 0x17=23 | 478×70,4 幀   | 加入對局中 |
//	| `Reload_Wait_For_Race_Info_`        @ 0xF551B | 0x18=24 | 480×116,4 幀  | 等種族資料 |
//	| `Reload_Initializing_Net_Info_`     @ 0xF54BE | 0x19=25 | 478×70,4 幀   | 初始化連線 |
//	| `Reload_Sending_Data_Info_`         @ 0xF54D9 | 0x1A=26 | 411×105,4 幀  | 傳送資料 |
//	| `Reload_Generating_Map_Info_`       @ 0xF53CB | 0x1E=30 | 478×70,5 幀   | 產生星圖 |
//	| `Reload_Getting_Data_Info_`         @ 0xF54A0 | 0x1F=31 | 443×105,4 幀  | 接收資料 |
//
// 也就是說「`Join_Net`」「`Generic_Net_Info`」「`SendGet_Net_Info`」不是三張畫面,
// 是**同一張畫面的三個狀態**——原版連線流程走到哪一步就換哪一張圖。
//
// 這是靜態反追溯源的典型收穫:照著畫面名清單一張一張做,會做出七份幾乎一樣的程式碼;
// 追到共用的那個 loader 才看得出真正的形狀是「一個面板 + 一個狀態列舉」。
//
// ============ 版面 ============
//
// `Reload_Generic_Net_Info_` 的定位段(同第 29 項(決定性化),又是**算**出來的):
//
//	w = 資產.寬  h = 資產.高
//	[win+0xF3] = (0x280 − w) / 2      ; 水平置中於 640
//	[win+0xF5] = (0x1E0 − h) / 2      ; 垂直置中於 480
//	[win+0x10E] = 4                   ; 字型 id
//	[win+0x10C] = h
//
// `Draw_Generic_Net_Info_Screen_` 每次呼叫把 `[win+0x1EA]`(幀號)+1,畫完一輪歸零
// ——與第 29 項(決定性化)標題帶同一種逐幀播法。
//
// 兩個位置是立即數,不是算的:
//
//	`Add_Waiting_For_Joiners_Field_` @ 0xF0801:(winX+0x9E, winY+0x6A)
//	`Draw_SendGet_Net_Info_Screen_` @ 0xF2C8B:進度數字
//	    [win+0x10F]==0(傳送) → (winX+0x72, winY+0x42)
//	    [win+0x10F]==1(接收) → (winX+0x79, winY+0x41)
//
// `[win+0x10F]` 由 `Reload_Sending_Data_Info_` 設 0、`Reload_Getting_Data_Info_` 設 1
// ——傳送與接收共用同一段繪製,只差那兩個像素的位移。照抄。
//
// ⚠ **第一版把 `Add_Waiting_For_Joiners_Field_` 讀成「已加入人數」的欄位,那是錯的。**
// 它呼叫的是 `sub_1151B0` = **`Add_Button_Field_`**(符號表 0x1151B0),也就是那個座標是
// 一顆**按鈕**——把資產 15 的十幀攤開來看,(+158,+106) 正好是 `START NET GAME` 的左上角。
// 名字裡的 "Waiting_For_Joiners" 指的是「這顆鈕加在哪張畫面上」,不是它顯示什麼。
//
// 抓到的方式值得記:先是截圖上那串數字壓在 START NET GAME 上,才回去查 0x1151B0 是什麼。
// **符號名是二手推論,被呼叫的函式是一手事實**——名字聽起來像什麼不算證據。
//
// ============ 誠實留白 ============
//
// 原版這幾張的**字**印在圖上(美術烘進 LBX),所以中文得擦底疊字。擦底的範圍取面板中央,
// 不是整張——邊框是美術的一部分。
//
// remake 目前只有 `netInfoJoining` / `netInfoWaitingForJoiners` 兩個狀態會真的被連線流程用到
// (大廳只做到「誰連進來了」,見 choosenetplyrs.go)。其餘五個狀態的資產與版面都對好了,
// 等連線流程走到那幾步時直接切過去即可——**不是死碼,是還沒有觸發點**。

// netInfoState 是狀態面板的七個狀態。值就是原版的資產編號——不另編一套號,
// 免得多一層要對照的表。
type netInfoState int

const (
	netInfoWaitingForJoiners netInfoState = 15
	netInfoJoining           netInfoState = 23
	netInfoWaitRaceInfo      netInfoState = 24
	netInfoInitializing      netInfoState = 25
	netInfoSendingData       netInfoState = 26
	netInfoGeneratingMap     netInfoState = 30
	netInfoGettingData       netInfoState = 31
)

// netInfoStates 是七個狀態的列舉(供測試逐個檢查版面)。
func netInfoStates() []netInfoState {
	return []netInfoState{
		netInfoWaitingForJoiners, netInfoJoining, netInfoWaitRaceInfo,
		netInfoInitializing, netInfoSendingData, netInfoGeneratingMap, netInfoGettingData,
	}
}

// netInfoCaption 是每個狀態的中文說明(原版是烘在圖上的英文)。
func (b *sceneBuilder) netInfoCaption(st netInfoState) string {
	switch st {
	case netInfoWaitingForJoiners:
		return b.tr("等待其他玩家加入", "Waiting for players to join")
	case netInfoJoining:
		return b.tr("加入對局中", "Joining game")
	case netInfoWaitRaceInfo:
		return b.tr("等待種族資料", "Waiting for race info")
	case netInfoInitializing:
		return b.tr("初始化連線", "Initializing network")
	case netInfoSendingData:
		return b.tr("傳送資料", "Sending data")
	case netInfoGeneratingMap:
		return b.tr("產生星圖", "Generating map")
	case netInfoGettingData:
		return b.tr("接收資料", "Getting data")
	}
	return ""
}

// netInfoIsReceiving 對應原版的 `[win+0x10F]`:接收 = 1、傳送 = 0。
// 這個旗標只影響進度數字的位置(見檔頭)。
func netInfoIsReceiving(st netInfoState) bool { return st == netInfoGettingData }

// netInfoHasProgress 說這個狀態會不會顯示進度數字
// (原版只有 `Draw_SendGet_Net_Info_Screen_` 這條路徑會印)。
func netInfoHasProgress(st netInfoState) bool {
	return st == netInfoSendingData || st == netInfoGettingData
}

// netInfoWindow 依資產尺寸算面板左上角(見檔頭的算式)。
//
// 抽成函式與 cnpWindow 同理:座標由資產尺寸決定,寫死常數就測不到算式本身。
func netInfoWindow(w, h int) (x, y int) {
	return (moo2ScreenW - w) / 2, (moo2ScreenH - h) / 2
}

// 進度數字的位移(立即數,見檔頭)。
const (
	netInfoSendProgX, netInfoSendProgY = 0x72, 0x42
	netInfoRecvProgX, netInfoRecvProgY = 0x79, 0x41
	// START NET GAME 鈕(Add_Waiting_For_Joiners_Field_ → Add_Button_Field_)。
	netInfoStartBtnX, netInfoStartBtnY = 0x9E, 0x6A
	// 按鈕尺寸量自資產 15 攤開的幀(反組譯沒有給,標成量的)。
	netInfoStartBtnW, netInfoStartBtnH = 160, 25
	// 面板動畫:每幀停幾次重繪(同 nntBannerHold)。
	netInfoFrameHold = 5
)

// netInfoProgressPos 回傳進度數字的螢幕座標。
func netInfoProgressPos(winX, winY int, receiving bool) (x, y int) {
	if receiving {
		return winX + netInfoRecvProgX, winY + netInfoRecvProgY
	}
	return winX + netInfoSendProgX, winY + netInfoSendProgY
}

// netInfoScreen 是狀態面板畫面。
type netInfoScreen struct {
	b     *sceneBuilder
	state netInfoState
	// joined / total 是「等待加入」時的人數;progress 是傳送/接收的百分比。
	joined, total, progress int
	// onCancel 非 nil 時,點一下就走這條路(大廳可以放棄等待)。
	onCancel func() *origTransition
	// hosting 決定要不要畫 START NET GAME 鈕——只有主機開得了局,
	// 客戶端看到一顆按不動的鈕比沒有更糟。
	hosting bool
	// lobby 非 nil 時每幀重讀名冊,把「已加入幾人」畫成活的
	//(原版 `Add_Waiting_For_Joiners_Field_` 就是這個欄位)。
	lobby *netplay.Lobby

	bg     *ebiten.Image
	frames []*ebiten.Image
	tick   int
}

func (b *sceneBuilder) netInfo(st netInfoState) *netInfoScreen {
	s := &netInfoScreen{b: b, state: st}
	s.bg = b.multigmImage(mpBGAsset, false)
	for f := 0; ; f++ {
		im := b.multigmFrame(int(st), f)
		if im == nil {
			break
		}
		s.frames = append(s.frames, im)
	}
	return s
}

// netInfoDemo 是截圖廊用的狀態面板(「等待其他玩家加入」,原版流程裡最常看到的那一張)。
func (b *sceneBuilder) netInfoDemo() *netInfoScreen {
	s := b.netInfo(netInfoWaitingForJoiners)
	s.joined, s.total, s.hosting = 2, 4, true
	return s
}

func (s *netInfoScreen) frame() *ebiten.Image {
	if len(s.frames) == 0 {
		return nil
	}
	return s.frames[(s.tick/netInfoFrameHold)%len(s.frames)]
}

func (s *netInfoScreen) update(in shell.InputState) *origTransition {
	s.tick++
	if s.lobby != nil {
		s.joined = len(s.lobby.Roster().Players)
	}
	if !in.ClickReleased {
		return nil
	}
	if s.onCancel != nil {
		return s.onCancel()
	}
	sc, err := s.b.multiPlayer()
	if err != nil {
		return nil
	}
	return &origTransition{next: sc}
}

func (s *netInfoScreen) draw(dst *ebiten.Image) {
	dst.Fill(color.RGBA{6, 8, 14, 255})
	if s.bg != nil {
		op := &ebiten.DrawImageOptions{}
		dst.DrawImage(s.bg, op)
	}
	im := s.frame()
	if im == nil {
		return
	}
	w, h := im.Bounds().Dx(), im.Bounds().Dy()
	winX, winY := netInfoWindow(w, h)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(winX), float64(winY))
	dst.DrawImage(im, op)

	if s.b.fnt == nil {
		return
	}
	// 原版把字烘在圖上,中文得擦底疊字。擦底的左緣取 STATUS 欄的內緣(量的,+98)
	// ——再往左就把 "STATUS" 那個標籤也擦掉了,而它是美術的一部分。
	vector.DrawFilledRect(dst, float32(winX+98), float32(winY+h/2-16), float32(w-98-24), 30,
		color.RGBA{14, 16, 22, 240}, false)
	s.b.fnt.DrawCentered(dst, s.b.netInfoCaption(s.state),
		float64(winX+w/2), float64(winY+h/2)+6, 16, color.RGBA{240, 220, 120, 255})

	if s.state == netInfoWaitingForJoiners {
		// 已加入人數畫在 STATUS 欄裡(說明文字右側)——原版沒有給這個欄位的座標,
		// 這是**量出來的**估計值,標明以免日後被當成真值。
		if s.total > 0 {
			s.b.fnt.Draw(dst, fmt.Sprintf("%d / %d", s.joined, s.total),
				float64(winX+w-96), float64(winY+h/2)+6, 13, color.RGBA{215, 222, 238, 255})
		}
		// START NET GAME 鈕:位置是反組譯的真值,中文字擦底疊上去。
		if !s.hosting {
			return
		}
		bx, by := winX+netInfoStartBtnX, winY+netInfoStartBtnY
		vector.DrawFilledRect(dst, float32(bx+4), float32(by+4),
			float32(netInfoStartBtnW-8), float32(netInfoStartBtnH-8),
			color.RGBA{58, 58, 52, 255}, false)
		s.b.fnt.DrawCentered(dst, s.b.tr("開始連線對局", "START NET GAME"),
			float64(bx+netInfoStartBtnW/2), float64(by)+18, 13, color.RGBA{236, 232, 210, 255})
	}
	if netInfoHasProgress(s.state) {
		x, y := netInfoProgressPos(winX, winY, netInfoIsReceiving(s.state))
		s.b.fnt.Draw(dst, fmt.Sprintf("%d%%", s.progress), float64(x), float64(y)+14, 13,
			color.RGBA{215, 222, 238, 255})
	}
}
