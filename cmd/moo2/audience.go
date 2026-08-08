package main

// audience.go:星圖上方的會談請求燈(原版 `Draw_Diplomacy_Request_Lights_` @ 0x83D06)。
//
// 狀態面在 `internal/shell/audience.go`(誰在請求、來意是什麼);這一檔只管畫出來與點進去。
//
// ============ 版面是真值 ============
//
// 原版那支的排法很明確:
//
//	x = 0x1FA − 已畫的個數 × 圖寬     ; 0x1FA = 506
//	y = 5
//
// 也就是**從 x=506 往左排、貼在星圖上緣**。先請求的在最右邊。
// 每個燈是該種族的逐格動畫(`byte_19C148[種族]` 是目前幀,播完循環)。
//
// ⚠ **圖不是原版的。** 原版用的是 per-race 動畫,指標存在 `dword_19C128[種族]`,
// 那個陣列由別處填,資產來源沒追。這裡用「旗色方塊 + 來意首字」把資訊呈現出來——
// 玩家要看得出「誰在敲門、為什麼」,那才是這一層的作用。追到資產再換。

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

const (
	// audienceLightRightX 是最右邊那盞燈的右緣(原版 `mov edx, 1FAh`)。
	audienceLightRightX = 506
	// audienceLightY 是燈的上緣(原版 `mov edx, 5` 當 y 傳進繪製)。
	audienceLightY = 5
	// audienceLightW / audienceLightH 是 remake 的燈大小。原版的尺寸來自種族動畫本身,
	// 沒有解出來——這裡取一個能放下一個字的方塊。
	audienceLightW = 22
	audienceLightH = 16
)

// audienceLightRect 回傳第 n 盞燈(n 從 0 起,0 = 最右)的矩形。
//
// 往左排,與原版 `x = 506 − n×寬` 同向。
func audienceLightRect(n int) (x, y, w, h int) {
	return audienceLightRightX - (n+1)*audienceLightW, audienceLightY, audienceLightW, audienceLightH
}

// audienceReasonGlyph 把來意壓成一個字元,塞得進燈裡。
func (b *sceneBuilder) audienceReasonGlyph(reason string) string {
	switch reason {
	case shell.AudienceReasonWar:
		return b.tr("戰", "W")
	case shell.AudienceReasonTrade:
		return b.tr("貿", "T")
	case shell.AudienceReasonAlliance:
		return b.tr("盟", "A")
	}
	return b.tr("談", "?")
}

// audienceReasonColor 依來意給燈的底色:宣戰紅、提議偏綠/藍。
//
// 顏色是 remake 的選擇(原版用的是種族動畫,沒有「來意色」這個概念),
// 但它承載的是真的資訊——玩家不必點進去就知道是壞消息還是好消息。
func audienceReasonColor(reason string) color.RGBA {
	switch reason {
	case shell.AudienceReasonWar:
		return color.RGBA{170, 45, 40, 255}
	case shell.AudienceReasonTrade:
		return color.RGBA{40, 120, 70, 255}
	case shell.AudienceReasonAlliance:
		return color.RGBA{45, 80, 150, 255}
	}
	return color.RGBA{70, 70, 90, 255}
}

// drawAudienceLights 畫出所有請求會談的對手。
func (b *sceneBuilder) drawAudienceLights(dst *ebiten.Image) {
	sess := b.session
	if sess == nil || b.fnt == nil {
		return
	}
	for n, idx := range sess.AudienceRequests() {
		reason := sess.AudienceReasonFor(idx)
		x, y, w, h := audienceLightRect(n)
		if x < 0 {
			break // 排到左邊界外就不畫了(對手數超過版面能放的)
		}
		fillPanel(dst, float32(x), float32(y), float32(w), float32(h),
			audienceReasonColor(reason), false)
		vector.StrokeRect(dst, float32(x), float32(y), float32(w), float32(h), 1,
			color.RGBA{230, 220, 180, 255}, false)
		b.fnt.DrawCentered(dst, b.audienceReasonGlyph(reason),
			float64(x+w/2), float64(y+h/2), 12, color.RGBA{245, 240, 230, 255})
	}
}

// audienceLightAt 回傳點 (mx,my) 落在哪一盞燈上對應的對手索引,沒有回 -1。
func (b *sceneBuilder) audienceLightAt(mx, my int) int {
	sess := b.session
	if sess == nil {
		return -1
	}
	for n, idx := range sess.AudienceRequests() {
		x, y, w, h := audienceLightRect(n)
		if mx >= x && mx < x+w && my >= y && my < y+h {
			return idx
		}
	}
	return -1
}

// audienceEnemyName 回傳某個對手索引的名稱(越界回空字串)。
func audienceEnemyName(sess *shell.GameSession, idx int) string {
	if sess == nil || idx < 0 || idx >= len(sess.AIPlayers) {
		return ""
	}
	return sess.AIPlayers[idx].Name
}
