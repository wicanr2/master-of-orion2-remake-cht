package shell

import "github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"

// fighter.go:**格子戰場上的獨立戰機中隊**。
//
// remake 先前只有「戰機庫 → 母艦戰力 +N」這個抽象加成(見 session.go StartCombat)。
// 那在快速結算裡說得過去,但格子戰術戰鬥少了一整個兵種:戰機是**自己會動的單位**,
// 而且手冊給了它一整套與艦艇不同的規則。
//
// ============ 手冊逐字(GAME_MANUAL.pdf,pdfminer 直接萃取,非 OCR)============
//
//	p.157「All fighter craft are installed in ships and launched to a target in squadrons
//	       of four. You cannot separate the fighters in a squadron to direct them at
//	       multiple targets.」
//	p.157「Fighter craft fly to their target and use whatever weapons they have at
//	       point-blank range. If their primary target is no longer valid while they still
//	       have shots available, fighters select a new target automatically. Except for
//	       Assault shuttles, fighters will attempt to return to their carrier once they are
//	       out of shots. Once safely back, any surviving fighters get repairs, rearm,
//	       refuel, and can be launched again.」
//	p.157「Fighter craft always attack from the direction of their target's weakest shield
//	       facing.」
//	p.157「Like missiles, fighter craft are vulnerable to beam weapons (though not to
//	       jamming). Fighter crafts' speed and armor (the best available) both contribute
//	       to their defense.」
//	p.157「Fighters have a 50% chance to avoid the effects of any ship self-destruct, warp
//	       core breach explosion, or spherical weapon (Pulsar, Spatial Compressor, Plasma
//	       Flux).」
//	p.157「With the exception of Interceptors, fighter craft cannot engage one another in
//	       dogfights; they only target enemy ships, bases, and planets. Interceptors can
//	       attack enemy fighter craft of all types.」
//	p.82 「(Interceptors) move at speed 10 modified by your best drive and can take 2
//	       damage modified by your best armor. Interceptors fly directly to their target and
//	       fire 4 times at point-blank range. They then return to the carrier…」
//	p.83 「(Heavy Fighters) can take 5 damage (plus what the armor absorbs), and move at
//	       speed 8 modified by your best drive.」
//	p.87 「Fighters and missiles take damage as if they were size class one-half.」
//
// 數值常數(中隊 4、射擊次數、基礎血量、基礎速度)都在 gamedata/combat.go,
// 這一檔只放「一隊戰機在戰場上怎麼動、怎麼打、怎麼回家」的狀態機。
//
// ⚠ 沒有建模的部分,誠實列出:
//   - 「always attack from the weakest shield facing」——remake 的護盾是單一數值
//     (`CombatShip.ShieldReduction`),沒有四面分別的護盾,所以這條無處可套。
//     等護盾分面做出來時,這裡是它的掛勾點。
//   - 轟炸機/突擊梭:前者要炸彈對行星的規則、後者要把陸戰隊送上敵艦,
//     兩者各自依賴另一套系統。這一輪只做攔截機與重戰機(兩種純對艦戰機)。

// FighterKind 是戰機型別。
//
// 手冊 p.127 列了**四種**(攔截機 / 轟炸機 / 重戰機 / 突擊艇),這裡有**三種**
// ——突擊艇沒有,理由見下方 FighterBomber 的註解(登艦戰機制不存在,而手冊也沒給它血量)。
// 檔頭寫清楚「為什麼是三個」,因為先前這裡寫著「四種」而底下只有兩個,
// **那種不一致沒有任何測試會抓到**(第 71 項)。
type FighterKind int

const (
	FighterInterceptor FighterKind = iota
	FighterHeavy
	// FighterBomber 手冊(Bomber Bays):「Bombers are short-range fighters similar to
	// Interceptors, except that these carry **one bomb**. Each bomber can attack either a
	// planet **or a ship**. Bombers are installed and launched in squadrons of 4… They move
	// at **speed 8** … and can take **4 damage**…」
	//
	// ⚠ 2026-08-08(第 71 項)補上。本型別的檔頭一直寫著「手冊 p.127 的**四種**」,
	// 而底下只有兩個——`gamedata` 那邊的速度與射擊次數**兩組都是四型都齊的**,
	// 缺的只有 shell 這一側的型別與血量常數。**註解說四種、程式碼只有兩種**,
	// 這種不一致沒有任何測試會抓到。
	//
	// 突擊艇(Assault Shuttle)仍然沒有:它的用途是把陸戰隊送上敵艦,而**登艦戰機制
	// 不存在**(第 61 項)。加一個不會做任何事的型別只是把洞藏起來。
	FighterBomber
)

// FighterKindName 回傳中文型別名。
func FighterKindName(k FighterKind) string {
	switch k {
	case FighterHeavy:
		return "重戰機"
	case FighterBomber:
		return "轟炸機"
	default:
		return "攔截機"
	}
}

// FighterBaseSpeed / FighterBaseHits / FighterShots 是各型的手冊真值(轉手 gamedata)。
func FighterBaseSpeed(k FighterKind) int {
	switch k {
	case FighterHeavy:
		return gamedata.CombatFighterBaseSpeedHeavyFighter
	case FighterBomber:
		return gamedata.CombatFighterBaseSpeedBomber
	}
	return gamedata.CombatFighterBaseSpeedInterceptor
}

func FighterBaseHits(k FighterKind) int {
	switch k {
	case FighterHeavy:
		return gamedata.FighterHitsHeavyFighter
	case FighterBomber:
		return gamedata.FighterHitsBomber
	}
	return gamedata.FighterHitsInterceptor
}

func FighterShots(k FighterKind) int {
	switch k {
	case FighterHeavy:
		return gamedata.FighterShotsHeavyFighter
	case FighterBomber:
		return gamedata.FighterShotsBomber
	}
	return gamedata.FighterShotsInterceptor
}

// FighterSquadron 是戰場上的一隊戰機。
//
// **一隊是一個單位**——手冊明說不能把隊裡的戰機分開指向不同目標,所以這裡沒有逐架的位置,
// 只有整隊的格位與「還剩幾架」。
type FighterSquadron struct {
	Kind  FighterKind
	Enemy bool // false = 玩家的,true = 敵方的
	// Carrier 是母艦在該方 CombatShip 陣列中的索引;−1 = 母艦已被擊毀(回不去了)。
	//
	// ⚠ 索引型欄位的「沒有」必須是 −1 不是 0——0 是第一艘船。
	Carrier int
	Col     int
	Row     int
	// Alive 是還剩幾架(0 = 整隊被打光)。HPEach 是每架目前的血量,
	// 打起來是「一架一架掉」而不是整隊共用一條血條——手冊的血量是**每架**的。
	Alive     int
	HPEach    int
	MaxHPEach int
	// ShotsLeft 是**整隊**還能執行幾輪攻擊(每輪隊裡每架各開一次火)。
	// 歸零就返航(手冊:「fighters will attempt to return to their carrier once they are
	// out of shots」)。
	ShotsLeft int
	Speed     int  // 每回合可移動的格數(CombatFighterSpeed)
	Returning bool // 正在返航
}

// NewFighterSquadron 出擊一隊戰機。
//
// ftlLevel 是目前的引擎階(手冊:速度 = 基礎 + 2×(FTL−1) + 跨維加成);
// armorLevelAboveTitanium 是裝甲比鈦高幾級(手冊:血量 = 基礎 + 2×級數)。
func NewFighterSquadron(kind FighterKind, enemy bool, carrier, col, row, ftlLevel, armorLevelAboveTitanium int) FighterSquadron {
	hp := gamedata.FighterHitsWithArmor(FighterBaseHits(kind), armorLevelAboveTitanium)
	return FighterSquadron{
		Kind: kind, Enemy: enemy, Carrier: carrier, Col: col, Row: row,
		Alive: gamedata.FighterSquadronSize, HPEach: hp, MaxHPEach: hp,
		ShotsLeft: FighterShots(kind),
		Speed:     gamedata.CombatFighterSpeed(FighterBaseSpeed(kind), ftlLevel),
	}
}

// Dead 回傳整隊是否已被打光。
func (f *FighterSquadron) Dead() bool { return f.Alive <= 0 }

// CanTargetFighter 回傳這一型戰機能不能打**別的戰機**。
//
// 手冊 p.157:「With the exception of Interceptors, fighter craft cannot engage one another
// in dogfights… Interceptors can attack enemy fighter craft of all types.」
func (f *FighterSquadron) CanTargetFighter() bool { return f.Kind == FighterInterceptor }

// FighterSquadronDamage 回傳這一隊這一輪打出去的總傷害(每架各開一次火)。
//
// perCraft 是一架一次射擊的傷害(由呼叫端依當局武器科技算出來——手冊給的是武裝**類型**
// 不是固定數,不在這裡臆造)。呼叫端要自己扣 ShotsLeft(見 SpendShot)。
func FighterSquadronDamage(alive, perCraft int) int {
	if alive < 0 {
		alive = 0
	}
	return alive * perCraft
}

// SpendShot 消耗一輪攻擊;打完最後一輪就轉成返航(突擊梭除外,但這一輪沒做突擊梭)。
func (f *FighterSquadron) SpendShot() {
	if f.ShotsLeft > 0 {
		f.ShotsLeft--
	}
	if f.ShotsLeft <= 0 {
		f.Returning = true
	}
}

// TakeHit 對這一隊造成 dmg 點傷害,回傳被打掉幾架。
//
// 手冊的血量是**每架**的,所以傷害是一架一架吃:打光一架才溢到下一架。
// (先前的抽象結算把整隊當成母艦的一條血,那是快速結算的簡化,不是這裡的規則。)
func (f *FighterSquadron) TakeHit(dmg int) int {
	killed := 0
	for dmg > 0 && f.Alive > 0 {
		if dmg < f.HPEach {
			f.HPEach -= dmg
			break
		}
		dmg -= f.HPEach
		f.Alive--
		killed++
		f.HPEach = f.MaxHPEach
	}
	if f.Alive <= 0 {
		f.Alive, f.HPEach = 0, 0
	}
	return killed
}

// FighterAvoidsSpherical 回傳這一擊是否被戰機閃掉。
//
// 手冊 p.157:「Fighters have a 50% chance to avoid the effects of any ship self-destruct,
// warp core breach explosion, or spherical weapon (Pulsar, Spatial Compressor, Plasma Flux).」
// roll 是 1..100 的擲骰(由呼叫端的戰鬥亂數流提供,不在這裡開新的亂數源)。
func FighterAvoidsSpherical(roll int) bool { return roll <= 50 }

// StepToward 把這一隊往 (col,row) 移動,最多走 Speed 格(曼哈頓步進,先橫後縱)。
//
// 回傳是否已經**貼到目標旁邊**(曼哈頓距離 ≤ 1)——手冊:戰機是飛到目標身上
// 「at point-blank range」開火的,不像艦艇有射程。
func (f *FighterSquadron) StepToward(col, row int) bool {
	budget := f.Speed
	for budget > 0 && fighterDist(f.Col, f.Row, col, row) > 1 {
		if f.Col != col {
			f.Col += sign(col - f.Col)
		} else {
			f.Row += sign(row - f.Row)
		}
		budget--
	}
	return fighterDist(f.Col, f.Row, col, row) <= 1
}

// Recover 是「安全返航」:手冊「Once safely back, any surviving fighters get repairs,
// rearm, refuel, and can be launched again」——補血、補彈,可以再出擊。
//
// ⚠ **不會補人**:被打掉的戰機是真的沒了,手冊寫的是 "any surviving fighters"。
func (f *FighterSquadron) Recover() {
	f.HPEach = f.MaxHPEach
	f.ShotsLeft = FighterShots(f.Kind)
	f.Returning = false
}

func fighterDist(c1, r1, c2, r2 int) int { return absInt(c1-c2) + absInt(r1-r2) }

func sign(n int) int {
	switch {
	case n > 0:
		return 1
	case n < 0:
		return -1
	}
	return 0
}

func absInt(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
