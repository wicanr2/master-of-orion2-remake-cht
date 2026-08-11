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
//   - 「艦身旋轉」與原版方向名稱仍未解出；但戰機已依手冊接到
//     `FighterDamageAtWeakestShield`，會選四面容量最低者並扣該面。
//   - 一般敵方艦隊的原版逐艦藍圖仍沒有全部取回；因此 shell 以「抽象艦體戰力 →
//     敵方戰機藍圖」的固定表建立出擊槽位。但一旦出擊，戰機已使用原版第二組
//     傷害範圍與兩條 raw 傷害下游（見 fighter_attack.go）；未追回的是哪一艘敵艦帶哪個槽位，
//     不是命中／傷害鏈本身。

// FighterKind 是戰機型別。
//
// 手冊 p.127 列了**四種**(攔截機 / 轟炸機 / 重戰機 / 突擊艇),這裡有**三種**
// ——突擊艇沒有,理由見下方 FighterBomber 的註解(登艦戰機制不存在,而手冊也沒給它血量)。
// 檔頭寫清楚「為什麼是三個」,因為先前這裡寫著「四種」而底下只有兩個,
// **那種不一致沒有任何測試會抓到**(第 70 項(陀螺去穩器))。
type FighterKind int

const (
	FighterInterceptor FighterKind = iota
	FighterHeavy
	// FighterBomber 手冊(Bomber Bays):「Bombers are short-range fighters similar to
	// Interceptors, except that these carry **one bomb**. Each bomber can attack either a
	// planet **or a ship**. Bombers are installed and launched in squadrons of 4… They move
	// at **speed 8** … and can take **4 damage**…」
	//
	// ⚠ 2026-08-08(第 70 項(陀螺去穩器))補上。本型別的檔頭一直寫著「手冊 p.127 的**四種**」,
	// 而底下只有兩個——`gamedata` 那邊的速度與射擊次數**兩組都是四型都齊的**,
	// 缺的只有 shell 這一側的型別與血量常數。**註解說四種、程式碼只有兩種**,
	// 這種不一致沒有任何測試會抓到。
	//
	FighterBomber
	// FighterAssaultShuttle 手冊(Assault Shuttles):「Assault Shuttles are fighters (like
	// the Interceptors) that carry **1 Marine unit**. … installed and launched in
	// **squadrons of 4**. Each shuttle moves at **speed 6** modified by your best drive and
	// can take **3 damage** modified by your best armor. Once launched, Assault Shuttles fly
	// to the target ship and drop off their Marines, which board and attempt capture.」
	//
	// ⚠ 2026-08-08(第 80 項(登艦戰))補上。這裡先前寫著「突擊艇仍然沒有:登艦戰機制不存在,
	// 加一個不會做任何事的型別只是把洞藏起來」——**那句話是對的,而且它現在過期了**:
	// 登艦戰在 boarding.go 建好了。手冊那句「like the Interceptors」把它歸進戰機家族,
	// 所以它走同一套 FighterSquadron,只是抵達目標時做的是登艦而不是開火。
	FighterAssaultShuttle
)

// FighterKindName 回傳中文型別名。
func FighterKindName(k FighterKind) string {
	switch k {
	case FighterHeavy:
		return "重戰機"
	case FighterBomber:
		return "轟炸機"
	case FighterAssaultShuttle:
		return "突擊艇"
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
	case FighterAssaultShuttle:
		// 手冊「moves at speed 6 modified by your best drive」。
		return gamedata.AssaultShuttleBaseSpeed
	}
	return gamedata.CombatFighterBaseSpeedInterceptor
}

func FighterBaseHits(k FighterKind) int {
	switch k {
	case FighterHeavy:
		return gamedata.FighterHitsHeavyFighter
	case FighterBomber:
		return gamedata.FighterHitsBomber
	case FighterAssaultShuttle:
		// 手冊「can take 3 damage modified by your best armor」。
		return gamedata.AssaultShuttleBaseHits
	}
	return gamedata.FighterHitsInterceptor
}

func FighterShots(k FighterKind) int {
	switch k {
	case FighterHeavy:
		return gamedata.FighterShotsHeavyFighter
	case FighterBomber:
		return gamedata.FighterShotsBomber
	case FighterAssaultShuttle:
		// 突擊艇不開火——它飛到目標旁邊放下陸戰隊就沒事了(手冊:「drop off their
		// Marines … unpiloted shuttles are set adrift」)。0 = 抵達即用完。
		return 0
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
	// CarrierName 是跨艦隊戰損壓縮仍穩定的母艦識別；Carrier 只作目前陣列的快取索引。
	CarrierName string
	Col         int
	Row         int
	// Alive 是還剩幾架(0 = 整隊被打光)。HPEach 是每架目前的血量,
	// 打起來是「一架一架掉」而不是整隊共用一條血條——手冊的血量是**每架**的。
	Alive     int
	HPEach    int
	MaxHPEach int
	// ShotsLeft 是**整隊**還能執行幾輪攻擊(每輪隊裡每架各開一次火)。
	// 歸零就返航(手冊:「fighters will attempt to return to their carrier once they are
	// out of shots」)。
	ShotsLeft int
	Speed     int // 每回合可移動的格數(CombatFighterSpeed)
	// 以下是這隊戰機被艦艇光束／PD 瞄準時使用的 Beam Defense 加成；
	// 由出擊時的母艦 CombatShip 帶入，不把敵方未建模的艦隊資料臆造進來。
	FighterRacialDefenseBonus int
	FighterPilotBonus         int
	FighterHelmsmanBonus      int
	Returning                 bool // 正在返航
	// TargetName 是目前的主要目標識別。手冊 p.157 說戰機會飛向主要目標，
	// 只有主要目標失效、而且仍有可用射擊時才自動重選；因此呼叫端不可每回合
	// 直接改追最近艦。戰術畫面中的敵艦名稱在一場戰鬥內唯一，足以跨過敵艦
	// 陣列的戰損壓縮；空名稱則由呼叫端採保守的每回合選擇。
	TargetName string
}

// EnemyFighterProfileForStrength 是 remake 對抽象敵艦戰力的固定戰機藍圖。
// 原版 genEnemyFleet 的逐艦設計槽位尚未完整取回，不能把這張表當成反組譯真值；
// 但表內每型戰機的規則仍完全走手冊數值，且輸入只依敵艦戰力，保證同一回合可重播。
//
// 戰力 8 起代表艦體有標準戰機庫；戰力 16 起使用重戰機庫；戰力 32 起使用轟炸機庫。
// 小於 8 的偵察／輕型艦不帶戰機庫，避免低階敵艦憑空取得戰機火力。
func EnemyFighterProfileForStrength(strength int) (kind FighterKind, hasBay bool) {
	switch {
	case strength >= 32:
		return FighterBomber, true
	case strength >= 16:
		return FighterHeavy, true
	case strength >= 8:
		return FighterInterceptor, true
	default:
		return FighterInterceptor, false
	}
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
