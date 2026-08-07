package shell

// wormhole.go:蟲洞(原版 `Generate_Wormhole_Links_` @ 0x8CC15 / `Set_Wormhole_Id_` @ 0x8D6D6)。
//
// 蟲洞是 MOO2 星圖上把兩顆遠星連起來的捷徑,是**遊戲機制**不只是裝飾:艦隊走蟲洞
// 一回合就到,不必橫越整個銀河。remake 先前完全沒有這個概念,連 `Star` 都沒有欄位——
// 星圖第 1 層 `Draw_Wormhole_Links_` 因此畫不出來(卡的是資料模型不是繪圖)。
//
// ⚠ **別和隨機事件的「蟲洞」搞混**:MOO2 兩者都有,remake 也是。
//
//	隨機事件(`applyWormhole`,events*.go):一次性好事,把**正在航行**的艦隊直接送到,
//	                                      手冊 p.181「moves that fleet to their
//	                                      destination in a single turn」。
//	本檔(星圖蟲洞):**永久**存在於星圖上的兩星連線,任何時候都能走。
//
// 兩者共用「一回合就到」這個語意,而那句手冊原文正好是本檔 ETA = 1 的錨點。
//
// ============ 原版的產生規則(照抄得到的部分)============
//
// `Generate_Wormhole_Links_`:
//
//	taken[72] = {0}
//	for 每個玩家:  taken[該玩家母星] = 1        ; 母星不可當端點
//	for k = 0 .. _n_wormholes-1:
//	    最多試 200 次(0xC8)挑一顆 A:
//	        s = Random(星數 − 1)                 ; ⚠ 原版 Random 回 1..n,所以永遠不會挑到星 0
//	        排除:taken[s] / star[s].+0x28 != 0 / sub_79001(s) / 黑洞 / 已有蟲洞
//	    Set_Wormhole_Id_(A)
//
// `Set_Wormhole_Id_` 再挑另一端 B,條件是:
//
//	距離(A,B) > galaxySizeParam + 3      ; **最短距離門檻**,蟲洞不會連兩顆鄰星
//	!taken[B] / B 沒有蟲洞 / B 不是黑洞 / B.+0x28 == 0 / B != A
//	候選收滿 19 個(0x13)就停,再從候選裡挑
//
// ============ ⚠ 沒照抄的:數量 ============
//
// `_n_wormholes`(`byte_182245`)不是常數——它在**銀河產生的過程中逐星累加**
// (`sub_8C840`),上限是 `galaxySizeParam × 4 + 4`。要忠實重現得連整個銀河產生器一起搬,
// 而 remake 的星圖是「格點 + 抖動」的自有模型,兩邊接不起來。
//
// 這裡改用**與上限同構的規則**:`星數/8`,夾在 1..4。理由是原版上限也是隨銀河大小線性成長、
// 且個位數量級(24 星的小銀河給 3 個)。**這是 remake 的選擇,不是原版真值。**
// 真的要對齊,得先把原版的銀河產生器搬過來。
//
// 最短距離門檻同理:原版是 `galaxySizeParam + 3` 的**銀河座標**單位,remake 的座標是
// 正規化 0..1,沒有換算依據。用 0.45(對角線約 1.41 的三分之一)保住「不連鄰星」這個語意。

// wormholeMinDist 是兩端的最短距離(正規化座標)。見檔頭 ⚠。
const wormholeMinDist = 0.45

// wormholeCount 依星數決定蟲洞數。見檔頭 ⚠(與原版上限同構,非原版真值)。
func wormholeCount(nStars int) int {
	n := nStars / 8
	if n < 1 {
		n = 1
	}
	if n > 4 {
		n = 4
	}
	return n
}

// genWormholes 在 stars 上就地配對蟲洞。homeStars 是不可當端點的母星集合。
//
// rnd 必須是決定性的(同一顆種子產生同一組蟲洞),與星系其餘部分一致。
// 傳入的 stars 的 `Wormhole` 必須都是 -1(genGalaxy 已如此建構)。
func genWormholes(stars []Star, homeStars map[int]bool, rnd func(n int) int) {
	if len(stars) < 2 {
		return
	}
	// 可當端點:不是母星、不是黑洞、還沒有蟲洞。
	eligible := func(i int) bool {
		if i < 0 || i >= len(stars) {
			return false
		}
		if homeStars[i] || stars[i].Wormhole >= 0 {
			return false
		}
		return stars[i].Spectral != blackHoleSpectral
	}
	dist2 := func(a, b int) float64 {
		dx := stars[a].X - stars[b].X
		dy := stars[a].Y - stars[b].Y
		return dx*dx + dy*dy
	}

	want := wormholeCount(len(stars))
	minD2 := wormholeMinDist * wormholeMinDist
	for made := 0; made < want; made++ {
		// 挑 A:原版是「最多試 200 次」的拒絕取樣,這裡照抄那個上限。
		a := -1
		for try := 0; try < 200 && a < 0; try++ {
			if c := rnd(len(stars)); eligible(c) {
				a = c
			}
		}
		if a < 0 {
			return // 沒得挑了,不硬湊
		}
		// 挑 B:原版收滿 19 個候選就停,這裡同樣掃一遍取候選再抽一個。
		var cand []int
		for b := 0; b < len(stars) && len(cand) < 19; b++ {
			if b == a || !eligible(b) || dist2(a, b) <= minD2 {
				continue
			}
			cand = append(cand, b)
		}
		if len(cand) == 0 {
			continue // 這顆 A 找不到夠遠的對象,換下一輪
		}
		b := cand[rnd(len(cand))]
		stars[a].Wormhole = b
		stars[b].Wormhole = a // ⚠ 必須雙向,見 Star.Wormhole 註解
	}
}

// blackHoleSpectral 是 Star.Spectral 的黑洞值(原版星球結構 +0x16 == 6)。
const blackHoleSpectral = 6

// normalizeWormholes 把不合法的蟲洞欄位清成 -1。
//
// 三種來源都需要它:
//   - **舊存檔**沒有這個欄位,JSON 解出來是零值 0 → 每顆星都宣稱與星 0 有蟲洞。
//   - 索引越界(星數變了)。
//   - 單向蟲洞——openorion2 對這種狀態直接丟例外,remake 選擇清掉而不是崩潰。
func normalizeWormholes(stars []Star) {
	for i := range stars {
		w := stars[i].Wormhole
		if w < 0 || w >= len(stars) || w == i {
			stars[i].Wormhole = -1
		}
	}
	for i := range stars {
		if w := stars[i].Wormhole; w >= 0 && stars[w].Wormhole != i {
			stars[i].Wormhole = -1
		}
	}
}

// WormholeBetween 回傳 a、b 兩星之間是否有蟲洞。
func (s *GameSession) WormholeBetween(a, b int) bool {
	if a < 0 || a >= len(s.Stars) || b < 0 || b >= len(s.Stars) || a == b {
		return false
	}
	return s.Stars[a].Wormhole == b && s.Stars[b].Wormhole == a
}
