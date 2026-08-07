package shell

// nebula.go:星雲(原版 `Generate_Nebulas_` @ 0x8C099、`Generate_Number_Of_Nebulas_` @ 0x8C4D3、
// `Point_Is_In_Nebula_N_` @ 0xEB9C8、`Draw_Nebulae_` @ 0x84F8F)。
//
// 星雲不只是星圖上的裝飾,是**有規則的地形**。手冊(GAME_MANUAL.pdf,星圖那一節)寫得很明白:
//
//	"Ships traveling through a nebula are reduced in speed to 1 parsec per turn.
//	 More importantly, the fierce ionization prevents deflector shields from
//	 functioning without Hard Shields technology."
//
// 戰鬥那一節再講一次(p.158):
//
//	"if combat takes place in a nebula, all shields become inoperative,
//	 except for those on ships equipped with Hard Shields."
//
// ============ 一、判定:遮罩像素 > 5 ============
//
// 星雲不是圓形也不是矩形,是**一張圖的形狀**。`Point_Is_In_Nebula_N_`:
//
//	cx = (點x − 星雲x) / 3 ; 若 < 0 → 不在
//	cy = (點y − 星雲y) / 3 ; 若 < 0 → 不在
//	若 cx ≥ 圖寬 或 cy ≥ 圖高 → 不在
//	若 圖像素[cy×寬 + cx] ≤ 5 → 不在
//	否則 → 在
//
// patch 1.5 手冊獨立佐證了這個門檻:「a star is considered "in nebula" if the respective
// pixel value of the nebula picture is greater than 5」,還補了一句「deep in a Nebula a few
// dark pixels can be present causing a star at such a location to be considered "not in nebula"」
// —— 也就是這個判定本身就會有小破洞,那是原版行為。
//
// 遮罩要讀 LBX,而本套件是純規則層不碰資產,所以判定式由外層(cmd/moo2)用
// `SetNebulaProbe` 裝進來(見 route.go)。沒裝就是全部 false(headless 模擬即如此),已標明。
//
// ============ 二、數量:看銀河大小 ============
//
// `Generate_Number_Of_Nebulas_` 是一張 4 路跳表(銀河大小 0..3):
//
//	小   → Random(2) − 1   = 0..1
//	中   → Random(2)       = 1..2
//	大   → Random(3)       = 1..3
//	巨大 → Random(3) + 1   = 2..4
//
// 上限 4 與 `internal/save` 既有的 `maxNebulas = 4`(從存檔格式反推)一致 —— 兩邊獨立對上。
//
// ============ 三、⚠ 沒照抄的:位置 ============
//
// 原版的星雲座標是銀河座標系(遮罩是它的 1/3 解析度),而 remake 的星圖是「格點 + 抖動」的
// 自有模型、座標正規化到 0..1,兩邊接不起來(與蟲洞同一個情況,見 wormhole.go)。
// 這裡改成在正規化空間裡隨機擺,並避開母星。**這是 remake 的選擇,不是原版真值。**
//
// ============ 四、移動懲罰 ============
//
// 手冊那句「reduced in speed to 1 parsec per turn」已經實作:秒差距模型在 `starlane.go`、
// 「有沒有穿過星雲」的逐段判定在 `route.go`。Navigator 的豁免也在同兩支。

import (
	"math/rand"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// nebulaTypes 是星雲外觀的種類數。
//
// 來源是 STARBG.LBX 的資產配置:0..5 是 640×480 的星空層,6 之後每 4 張一組
// (`Load_Nebula_Pictures_` @ 0x8DA07 的內圈跑 4 次,對應 4 個縮放等級),
// 全檔 54 張 → (54 − 6) / 4 = **12 種**。
const nebulaTypes = 12

// maxNebulae 是同時存在的星雲數上限(原版跳表最大值 `Random(3) + 1`)。
const maxNebulae = 4

// galaxySizeClass 把星數換成銀河大小檔位(0..3),供星雲數的跳表索引。
//
// 就是原版的 `Galaxy_Size_From_N_Stars_`(門檻 20/36/54/72),放在
// `gamedata.GalaxySizeFromStars`——**銀河跨距(秒差距)也用同一支**,兩處不能各有一套。
//
// ⚠ 這一點踩過兩次:第一版自己拿原版四檔的星數取中點當界線;第二版改查 `GalaxySizes`,
// 但當時那張表是 remake 自訂的 12/24/36/48,和原版四檔對不上。徵狀都一樣——
// **開局星圖上常常一團星雲都沒有**,而那看起來完全合理。
// 現在 `GalaxySizes` 已改成原版真值(20/36/54/72),三者對齊。
func galaxySizeClass(nStars int) int { return gamedata.GalaxySizeFromStars(nStars) }

// Nebula 是一團星雲。X/Y 是**左上角**(不是中心)的正規化位置,與原版一致
// —— `Point_Is_In_Nebula_N_` 是用 `點 − 星雲` 去索引遮罩的,所以那個座標是原點。
type Nebula struct {
	X, Y float64
	Type int
}

// nebulaCount 依銀河大小回傳星雲數(原版 `Generate_Number_Of_Nebulas_` 的 4 路跳表)。
//
// size 0..3 = 小/中/大/巨大。超出範圍回 0(原版 `cmp ax, 3; ja default` 也是直接 return)。
// rnd(n) 必須回 1..n(與原版 `Random` 同語意)。
func nebulaCount(size int, rnd func(n int) int) int {
	switch size {
	case 0:
		return rnd(2) - 1
	case 1:
		return rnd(2)
	case 2:
		return rnd(3)
	case 3:
		return rnd(3) + 1
	}
	return 0
}

// genNebulae 產生星雲。homeStars 是母星集合(避開,免得開局就被扣護盾)。
//
// ⚠ 位置是 remake 自己排的,見檔頭三。
func genNebulae(size int, homeStars map[int]bool, stars []Star, rnd *rand.Rand) []Nebula {
	n := nebulaCount(size, func(k int) int { return rnd.Intn(k) + 1 })
	if n > maxNebulae {
		n = maxNebulae
	}
	if n <= 0 {
		return nil
	}
	out := make([]Nebula, 0, n)
	for i := 0; i < n; i++ {
		// 挑一個離所有母星夠遠的落點,試不到就放棄這一團(不硬湊)。
		placed := false
		for try := 0; try < 40 && !placed; try++ {
			x, y := rnd.Float64()*0.8, rnd.Float64()*0.8
			ok := true
			for idx := range homeStars {
				if idx < 0 || idx >= len(stars) {
					continue
				}
				dx, dy := stars[idx].X-(x+0.1), stars[idx].Y-(y+0.1)
				if dx*dx+dy*dy < 0.04 { // 母星離星雲中心 0.2 以內就重來
					ok = false
					break
				}
			}
			if ok {
				out = append(out, Nebula{X: x, Y: y, Type: rnd.Intn(nebulaTypes)})
				placed = true
			}
		}
	}
	return out
}

// StarInNebula 回傳某顆星是否在星雲內(索引越界回 false)。
func (s *GameSession) StarInNebula(idx int) bool {
	if idx < 0 || idx >= len(s.Stars) {
		return false
	}
	return s.Stars[idx].InNebula
}

// CombatShieldsDisabled 回傳「在這顆星打起來,沒有 Hard Shields 的艦艇護盾是否失效」。
//
// 手冊 p.158:「if combat takes place in a nebula, all shields become inoperative,
// except for those on ships equipped with Hard Shields.」
func (s *GameSession) CombatShieldsDisabled(starIdx int) bool {
	return s.StarInNebula(starIdx)
}

// shipHasHardShield 回傳這艘船是否裝了硬化護盾(元件名比對,與 shipHasAutoRepair 同作法)。
func shipHasHardShield(sh Ship) bool { return sh.Special == "硬化護盾" }

// nebulaShield 把護盾減傷套上星雲規則:戰鬥發生在星雲內的星系時,沒有硬化護盾的艦艇護盾歸零。
//
// 戰鬥地點取 `FleetAtStar` —— remake 的戰鬥一律發生在艦隊所在星系。
func (s *GameSession) nebulaShield(base int, hardShield bool) int {
	if hardShield || !s.CombatShieldsDisabled(s.Fleet().AtStar) {
		return base
	}
	return 0
}
