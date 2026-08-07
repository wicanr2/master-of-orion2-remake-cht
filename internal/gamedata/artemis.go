package gamedata

// artemis.go:**阿提米絲系統網**——環繞整個星系的水雷網。
//
// 手冊 p.86 給了完整規格,一個字都不用猜:
//
//	Any enemy ship entering that system has a chance to set off mines based on its
//	**size class**: Frigate = 20%, Destroyer = 30%, Cruiser = 40%, Battleship = 50%,
//	Titan = 80%, and Doom Star = 100%. Any affected ship sustains damage from
//	**8-28 mines**. Each mine inflicts **20 damage minus the ship's shield class**.
//
// ============ 三件事各自獨立 ============
//
//	① 觸發機率  逐艦、依艦體等級,大船躲不掉(末日之星 100%)
//	② 水雷數量  8–28,每艘中招的船各擲一次
//	③ 每枚傷害  20 − 護盾等級
//
// 這三件事是**相乘**的關係,所以護盾的價值在這裡被放大了:第十級護盾把每枚從 20 壓到 10,
// 對一艘挨了 28 枚的船就是少了 280 點。remake 的護盾名字本身就帶等級
// (第一/三/五/七/十級),不需要另建對照表——那個「級」就是手冊說的 shield class。
//
// ============ 為什麼大船反而危險 ============
//
// 機率隨體積上升(20% → 100%),而傷害不隨體積下降。所以水雷網對艦隊的效果是
// **專打主力艦**:一群巡防艦大多能開過去,一艘末日之星必中。
// 這與原版把它放在 Planetoid Construction(很後期)是一致的——它是拿來擋大艦隊的。
//
// ============ 維護費 5 BC ============
//
// | 來源 | 維護費 |
// |---|---|
// | 手冊 p.86 | 5 BC |
// | remake 建築表(原版執行檔 `off_17EB3D + 12`) | **5** |
//
// ============ 誠實留白 ============
//
//   - 手冊列的六個 size class 就是 MOO2 的六種艦體。remake 另有「偵察艦」「殖民船」兩個
//     類別——**那不是原版的艦體等級**,原版的偵察艦/殖民船/前哨艦/運輸艦都是蓋在
//     **Frigate 艦體**上的設計。所以它們套 Frigate 的 20%,這是照原版的艦體分類,
//     不是替沒有數字的東西編一個。
//   - 手冊沒說水雷網會不會被消耗、會不會對同一支艦隊重複觸發。這裡照字面做:
//     「any enemy ship **entering** that system」= 每次進入該星系時逐艦擲一次,不消耗。

// 水雷網的三個常數(手冊 p.86)。
const (
	ArtemisMinesMin      = 8
	ArtemisMinesMax      = 28
	ArtemisDamagePerMine = 20
)

// ArtemisHullClass 是手冊列出的六個 size class。
type ArtemisHullClass int

const (
	ArtemisFrigate ArtemisHullClass = iota
	ArtemisDestroyer
	ArtemisCruiser
	ArtemisBattleship
	ArtemisTitan
	ArtemisDoomStar
)

// artemisTriggerPercent 是逐艦體等級的觸發機率(手冊逐字)。
var artemisTriggerPercent = [6]int{20, 30, 40, 50, 80, 100}

// ArtemisTriggerPercent 回傳某個艦體等級踩到水雷的機率(百分比)。
//
// 超出範圍回 Frigate 的 20%:那是最小的艦體,拿它當退路不會讓未知的東西白白免疫
// ——回 0 才是,那會讓任何分類失誤變成「這艘船穿得過雷區」。
func ArtemisTriggerPercent(hull ArtemisHullClass) int {
	if hull < 0 || int(hull) >= len(artemisTriggerPercent) {
		return artemisTriggerPercent[ArtemisFrigate]
	}
	return artemisTriggerPercent[hull]
}

// ArtemisMineDamage 回傳一枚水雷對某護盾等級的船造成的傷害(手冊:20 − 護盾等級)。
//
// 不會低於 0——第二十級護盾不存在,但真的傳進來時負傷害會變成幫對方補血。
func ArtemisMineDamage(shieldClass int) int {
	d := ArtemisDamagePerMine - shieldClass
	if d < 0 {
		return 0
	}
	return d
}

// ArtemisMineCount 把一個 0..(ArtemisMinesMax-ArtemisMinesMin) 的骰值換成命中的水雷數。
//
// 拆出來是為了讓呼叫端把亂數源留在自己手上(決定性回合結算要能重播,見 determinism.go)。
func ArtemisMineCount(roll int) int {
	span := ArtemisMinesMax - ArtemisMinesMin
	if roll < 0 {
		roll = 0
	}
	if roll > span {
		roll = span
	}
	return ArtemisMinesMin + roll
}

// ArtemisMineRollSpan 是 ArtemisMineCount 接受的骰值上界(含),供呼叫端算 rng.Intn 的參數。
const ArtemisMineRollSpan = ArtemisMinesMax - ArtemisMinesMin

// ArtemisShipDamage 回傳一艘中招的船這次受到的總傷害。
func ArtemisShipDamage(mines, shieldClass int) int {
	if mines <= 0 {
		return 0
	}
	return mines * ArtemisMineDamage(shieldClass)
}
