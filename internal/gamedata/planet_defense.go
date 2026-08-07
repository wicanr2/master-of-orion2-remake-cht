package gamedata

// planet_defense.go:**戰機基地**與**行星版恆星轉換器**——兩棟建築,同一個接點。
//
// remake 早就有「殖民地在被軌道轟炸時反擊」那條路徑(`shell.retaliationAttackers`),
// 而且已經接了星基 / 戰鬥站 / 星辰要塞 / 飛彈基地 / 地面砲台。這兩棟缺的不是子系統,
// 是沒人把手冊那兩段翻成程式碼。
//
// ============ 戰機基地(手冊 p.79)============
//
//	These house **10 Interceptor squadrons, 6 Bomber squadrons, or 4 Heavy Fighter
//	squadrons**, depending on the most advanced fighter technology you've discovered.
//	(Note that Interceptors are available immediately.) All ground-based squadrons of
//	fighter craft are **totally renewed every 10 turns**.
//	Fighter Garrisons can only be destroyed by orbital bombardment.
//
// 三個中隊數不是隨科技變強而是**隨科技變少**:10 → 6 → 4。因為每一階的戰機更強,
// 所以總戰力仍然是往上走的——照抄這三個數,不要「順手」讓它遞增。
//
// 「Interceptors are available immediately」那句括號說明了為什麼沒有「還沒有戰機科技」
// 這一格:攔截機是開局就有的。
//
// ============ 行星版恆星轉換器(手冊 p.111)============
//
//	It fires a plasma blast that inflicts **400 points of damage to each side** of a
//	target — **1,600 total damage** — regardless of range and defense.
//
// 「regardless of range and defense」是這門砲的重點:**不受距離與防禦影響**。
// remake 的反擊解算是抽象的,沒有「面」的概念,所以取的是**單面 400**
// 而不是 1600——反擊只打一次,不是同時打四面。
//
// ⚠ 手冊那段講的是「Building/System」共用的武器規格;**艦載版**那句
// 「destroys an entire planet — turns it into an asteroid belt — when fired from orbit」
// 是另一回事(遊戲外的行星摧毀),不在這一檔。
//
// ============ 維護費 ============
//
// | 建築 | 手冊 | remake 建築表(原版執行檔 `off_17EB3D + 12`) |
// |---|---|---|
// | Fighter Garrison | 2 BC | **2** |
// | Stellar Converter(行星版) | 6 BC | **6** |
//
// ============ 誠實留白 ============
//
//   - 「every 10 turns 全數整補」沒接:remake 的反擊解算不追蹤基地損耗,
//     每次轟炸都用滿編的中隊數。**這對防禦方有利**,方向要記住。
//   - 「只能被軌道轟炸摧毀」已經是既有行為(建築吸收 hits 那條路),不必另做。

// 戰機基地的中隊數(手冊 p.79 逐字)。數字隨科技**遞減**,見檔頭。
const (
	FighterGarrisonInterceptorSquadrons  = 10
	FighterGarrisonBomberSquadrons       = 6
	FighterGarrisonHeavyFighterSquadrons = 4
)

// FighterGarrisonTier 是戰機基地依「已解鎖的最高階戰機科技」分的三檔。
type FighterGarrisonTier int

const (
	FighterGarrisonInterceptor FighterGarrisonTier = iota
	FighterGarrisonBomber
	FighterGarrisonHeavyFighter
)

// FighterGarrisonSquadrons 回傳這一檔的中隊數。
//
// 超出範圍回攔截機那一檔:手冊說攔截機「available immediately」,那是保底的那一格。
func FighterGarrisonSquadrons(tier FighterGarrisonTier) int {
	switch tier {
	case FighterGarrisonBomber:
		return FighterGarrisonBomberSquadrons
	case FighterGarrisonHeavyFighter:
		return FighterGarrisonHeavyFighterSquadrons
	}
	return FighterGarrisonInterceptorSquadrons
}

// FighterBomberCombatContribution 回傳一個**轟炸機**中隊對抽象戰力的貢獻(攻擊, HP)。
//
// 補齊既有的攔截機/重戰機兩支:轟炸機每次出手只投彈(手冊 p.83,`FighterShotsBomber` = 1),
// 炸彈必中。HP 沿用攔截機那一格——手冊 p.127 的表只給了攔截機 2 與重戰機 5 兩個數字,
// **轟炸機那一格沒給**,所以取最接近的下界而不是自己插一個值,並在這裡標明。
func FighterBomberCombatContribution() (atk, hp int) {
	return FighterSquadronSize * FighterShotsBomber * fighterBombDamageApprox,
		FighterSquadronSize * FighterHitsInterceptor
}

// FighterGarrisonCombatContribution 回傳一座戰機基地對殖民地反擊的貢獻(攻擊, HP)。
//
// = 該檔次的中隊數 × 單一中隊的貢獻。
func FighterGarrisonCombatContribution(tier FighterGarrisonTier) (atk, hp int) {
	var a, h int
	switch tier {
	case FighterGarrisonBomber:
		a, h = FighterBomberCombatContribution()
	case FighterGarrisonHeavyFighter:
		a, h = FighterHeavyBayCombatContribution()
	default:
		a, h = FighterBayCombatContribution()
	}
	n := FighterGarrisonSquadrons(tier)
	return a * n, h * n
}

// StellarConverterDamagePerSide 是恆星轉換器一發打在**一面**上的傷害(手冊 p.111:400)。
const StellarConverterDamagePerSide = 400

// StellarConverterSides 是手冊那句「1,600 total」隱含的面數(1600 / 400 = 4)。
//
// 記下來是因為 1600 這個數字在手冊裡是直接寫的,而 4 是推出來的——
// 兩個都留著,任何人要用哪一個都看得到它的來歷。
const StellarConverterSides = 4

// StellarConverterTotalDamage 是手冊寫的四面總傷(1600)。
const StellarConverterTotalDamage = StellarConverterDamagePerSide * StellarConverterSides

// StellarConverterRetaliationAttack 回傳行星版恆星轉換器對殖民地反擊的攻擊值。
//
// 取**單面 400** 而不是 1600:remake 的反擊解算是抽象的、沒有「面」的概念,
// 一次反擊是一發打在一個目標上,不是同時打四面(見檔頭)。
func StellarConverterRetaliationAttack() int { return StellarConverterDamagePerSide }
