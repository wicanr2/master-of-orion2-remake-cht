package gamedata

// space_monster.go:太空怪獸(守衛星系的巨獸)。
//
// remake 先前完全沒有這個系統,而它其實一直被程式碼「引用」著——colonization.go 檔頭抄的手冊
// 原文就寫著殖民船要「as long as all space monsters and enemy ships have been cleared from
// that planet's system」,但那個 gate 從來沒有東西可擋。
//
// --- 種類:反組譯 + 手冊互證 ---
//
// 執行檔裡有五個連續的怪獸名字串(檔案位移 0x1F742C 起):
//
//	Guardian, Amoeba, Dragon, Hydra, Crystal
//
// 對應五個 `Load_*_Ship_Design_` 函式(Guardian/Amoeba/Dragon/Hydra/Crystal),以及
// `_monster_names` @ 0x199266、`_n_space_monsters` @ 0x19C319、
// `_user_wants_n_space_monsters` @ 0x19A006(**新遊戲設定值**,原版可調怪獸數量)。
// 名字順序即列舉順序。
//
// (另有 Eel,但它只出現在 `Load_Eel_Ship_Design_`,是隨機事件的入侵怪獸,不是星系守衛,
// 故不在這張表裡——手冊 p.180 也把 Eel 歸在「invades the galaxy」那一類。)
//
// --- 戰鬥數值:手冊 p.114「Monster Traits」逐字 ---
//
//	Crystal Ray(水晶):"inflicting 40-80 points of damage"
//	Plasma Breath(海德拉):"always strikes its target … maximum of 60 points of damage"
//	Phasor Eye(巨龍):"inflicting 5–10 points of damage"
//	Dragon Breath(巨龍):"always hits … maximum of 300 points … losing 15 points … speed 24"
//	Caustic Slime(變形蟲):"25–50 points of enveloping damage … loses 5 points of strength" 每回合
//
// 守護者(Guardian of Orion)的數值手冊沒有逐條列出(它是獵戶座的守衛,獨立於一般怪獸),
// remake 依「三條勝利路徑之一的最終關卡」給它全表最高的戰力,並明確標為 remake 估值。
//
// --- 生成規則:手冊 p.60 逐字 ---
//
//	"Except for Space Monsters, no system will have more than one special; a system with a
//	 monster will always have another special — that's usually what drew the monster there
//	 in the first place."
//
// 也就是:①一般星系最多一個 special ②**有怪獸的星系一定另外還有一個 special**。
// 這是一條可直接落地的生成規則,不是推測。

// SpaceMonster 是星系守衛怪獸的種類(順序 = 執行檔字串表順序)。
type SpaceMonster int

const (
	MonsterNone     SpaceMonster = iota
	MonsterGuardian              // 獵戶座守護者
	MonsterAmoeba                // 太空變形蟲
	MonsterDragon                // 太空巨龍
	MonsterHydra                 // 太空海德拉
	MonsterCrystal               // 太空水晶
)

// MonsterStats 是一種怪獸的戰鬥數值。
type MonsterStats struct {
	NameZH string
	NameEN string
	// DamageMin/DamageMax 是單發傷害範圍(手冊 p.114 的數字;只給上限的用 0 當下限標記,
	// 由 AlwaysHits 一起判讀)。
	DamageMin, DamageMax int
	// AlwaysHits 對應手冊的 "always strikes its target" / "always hits"
	//(電漿吐息、龍焰);remake 的戰鬥解算據此跳過命中判定。
	AlwaysHits bool
	// Structure 是怪獸的結構值(戰力/血量基準)。
	// ⚠ **remake 估值**:手冊只給武器傷害,沒有給怪獸的結構/裝甲數字。
	// 這裡依「手冊描述的威脅等級」排序給值,標明非手冊實據。
	Structure int
	// Estimated 標記 Structure 是估值(非手冊/反編實據),供文件與 UI 誠實呈現。
	Estimated bool
}

// monsterStats 五種怪獸的資料。傷害值來自手冊 p.114 逐字;Structure 是 remake 估值。
var monsterStats = map[SpaceMonster]MonsterStats{
	MonsterAmoeba: {
		NameZH: "太空變形蟲", NameEN: "Space Amoeba",
		DamageMin: 25, DamageMax: 50, // Caustic Slime:25-50 點包覆傷害
		Structure: 60, Estimated: true,
	},
	MonsterCrystal: {
		NameZH: "太空水晶", NameEN: "Space Crystal",
		DamageMin: 40, DamageMax: 80, // Crystal Ray:40-80 點
		Structure: 90, Estimated: true,
	},
	MonsterHydra: {
		NameZH: "太空海德拉", NameEN: "Space Hydra",
		DamageMin: 30, DamageMax: 60, AlwaysHits: true, // Plasma Breath:必中,上限 60
		Structure: 80, Estimated: true,
	},
	MonsterDragon: {
		NameZH: "太空巨龍", NameEN: "Space Dragon",
		// 巨龍有兩種攻擊:相位眼(5-10)與龍焰(必中,上限 300、每格 -15)。
		// remake 的快速結算沒有格子距離,取龍焰在中距離的有效值當單發傷害上限,
		// 下限取相位眼——這個折衷是 remake 的,手冊兩種武器都逐字列在上面。
		DamageMin: 10, DamageMax: 150, AlwaysHits: true,
		Structure: 120, Estimated: true,
	},
	MonsterGuardian: {
		NameZH: "獵戶座守護者", NameEN: "Guardian",
		// ⚠ 全部是 remake 估值:手冊沒有逐條列出守護者的武器數值(它是獵戶座的最終關卡)。
		DamageMin: 50, DamageMax: 200,
		Structure: 300, Estimated: true,
	},
}

// MonsterStatsFor 回傳該怪獸的數值;MonsterNone 或未知回 ok=false。
func MonsterStatsFor(m SpaceMonster) (MonsterStats, bool) {
	st, ok := monsterStats[m]
	return st, ok
}

// MonsterNameZH 回傳怪獸中文名;無怪獸回空字串。
func MonsterNameZH(m SpaceMonster) string {
	if st, ok := monsterStats[m]; ok {
		return st.NameZH
	}
	return ""
}

// GuardStarMonsters 是「會在星圖生成時守衛星系」的怪獸種類(不含守護者——
// 守護者只守獵戶座,由勝利路徑另行處理,見 antaran_victory.go 的同款分工)。
var GuardStarMonsters = []SpaceMonster{MonsterAmoeba, MonsterCrystal, MonsterHydra, MonsterDragon}

// RollGuardMonster 依 roll(1..len(GuardStarMonsters))挑一種守衛怪獸。
// 原版的挑法尚未反編(`Make_System_Monsters_` @ 0x7CDC5 只讀到「數量」那一段),
// 這裡是等機率——**選法是 remake 的,種類清單有硬證**。
func RollGuardMonster(roll int) SpaceMonster {
	if len(GuardStarMonsters) == 0 {
		return MonsterNone
	}
	i := (roll - 1) % len(GuardStarMonsters)
	if i < 0 {
		i += len(GuardStarMonsters)
	}
	return GuardStarMonsters[i]
}

// DefaultGuardMonsterCount 是每 N 顆星放一隻守衛怪獸的密度。
//
// 原版把數量做成**新遊戲設定**(`_user_wants_n_space_monsters` @ 0x19A006),
// remake 的新遊戲畫面目前沒有這個選項,先用一個固定密度。⚠ remake 值。
const DefaultGuardMonsterCount = 8

// GuardMonsterCountFor 回傳 n 顆星的星圖該放幾隻守衛怪獸(至少 1,除非星圖太小)。
func GuardMonsterCountFor(stars int) int {
	if stars < DefaultGuardMonsterCount {
		return 0
	}
	return stars / DefaultGuardMonsterCount
}
