package shell

import (
	"fmt"
	"math/rand"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// 安塔蘭勝利(手冊三條勝利路徑之二)的 shell 層整合。權威規則來源見
// internal/engine/victory.go 的 VictoryAntaran 手冊逐字引用,以及
// docs/tech/victory-conditions.md 第 6 節先前記載的 TODO(本檔把它從「完全沒有對應流程」補上
// 「已接:傳送門 + 反攻」)。
//
// 手冊原文(GAME_MANUAL.pdf p.183,「Winning」小節):
//
//	"An alternate method is to seek out and defeat the Antaran home fleet. This involves
//	travelling to the Antaran homeworld, which is not possible until you have the right
//	technology and build a Dimensional Gate. Once you defeat the awe-inspiring Antarans, all
//	the other races in the galaxy recognise your overwhelming superiority and quickly
//	capitulate. (This strategy is not available if you disabled Antaran Attacks when setting up
//	your game.)"
//
// 次元傳送門(Dimensional Portal)建築本身已存在於 gamedata.Buildings(手冊 p.106,前置科技
// TOPIC_MULTIDIMENSIONAL_PHYSICS,見 docs/tech/colony-buildings.md 第三節)——本檔只補「建成後
// 解鎖反攻」這段流程,不重新定義建築資料。

// antaranWeaponSlot 是原版安塔蘭即時戰鬥設計的一個已解出武器槽切片。
// RawFlags 保留戰鬥記錄 +0x56 的原始 16 位元值；它的每個 bit 尚未命名，不能把它
// 直接當成 ARM/FST/其他改造。
type antaranWeaponSlot struct {
	WeaponID int
	// Quantity 是一般安塔蘭艦設計已讀到的槽位數量。星際要塞不把
	// 99/198 填在這裡，避免把容量上限誤當成固定 runtime 數量。
	Quantity int
	// Seed／CapacityCap 僅用於星際要塞：loader 先寫 seed，之後由
	// sub_6EE8E @ 0x6EE8E 的 divisor 與 live tech 分支算出 runtime quantity；
	// CapacityCap 是 99/198 的上限／拆槽規則。
	Seed        int
	CapacityCap int
	RawFlags    uint16
}

// antaranDefenseUnit 是安塔蘭母星防禦艦隊的一個防禦單位。
//
// Strength 是 remake 既有的戰力代理；OriginalName、CombatClass 與 Fortress 則保留
// 原版載入器已證實的艦級資料。把艦級從單純整數拆出來很重要：球形武器的傷害會依
// 目標艦體級數變化，不能讓終局戰所有安塔蘭艦都退化成零值 Frigate。
type antaranDefenseUnit struct {
	OriginalName string
	Strength     int
	CombatClass  gamedata.CombatShipClass
	Fortress     bool
	// CombatLoaderRaw 是原版即時戰鬥 loader 的 IDA 平面 raw 位址；它是證據索引，
	// 不是檔案偏移，也不是重製引擎用來推導艦級的數值。
	CombatLoaderRaw uint32
	// FighterBayCount 是原版即時防禦設計中 ID 31(Fighter Bays) 的數量。
	// FighterBayWeaponID 是原版武器表 ID；目前已證實的戰機艙是 ID 31。
	// 標準戰機艙接進既有快速戰鬥戰機貢獻模型；要塞的直接武器槽則由
	// antaranWeaponFirepower 接入終局齊射。其餘單艦的槽位仍保存 raw 資料，
	// 不把未追回的敵艦下游消費端誤套進所有艦級。
	FighterBayWeaponID int
	FighterBayCount    int
	WeaponSlots        []antaranWeaponSlot
}

// 下列槽位直接來自 `dword_192864 + 313*ship + 0x52 + 0x0B*slot`：
// weapon ID 在 +0x00、數量在 +0x02、RawFlags 在 +0x04。它們是已知的設計資料，
// 不是敵方即時命中／傷害公式；目前只使用 ID 31 的戰機艙貢獻。
var antaranKnownWeaponSlots = map[string][]antaranWeaponSlot{
	"Intruder": {
		{WeaponID: 4, Quantity: 4, RawFlags: 0},
		{WeaponID: 4, Quantity: 2, RawFlags: 2},
		{WeaponID: 24, Quantity: 5, RawFlags: 0},
		{WeaponID: 13, Quantity: 2, RawFlags: 0},
		{WeaponID: 31, Quantity: 3, RawFlags: 0},
	},
	"Interdictor": {
		{WeaponID: 4, Quantity: 6, RawFlags: 0},
		{WeaponID: 4, Quantity: 2, RawFlags: 2},
		{WeaponID: 24, Quantity: 15, RawFlags: 0},
		{WeaponID: 13, Quantity: 2, RawFlags: 0},
		{WeaponID: 4, Quantity: 8, RawFlags: 4},
		{WeaponID: 11, Quantity: 2, RawFlags: 0},
	},
	"Harbinger": {
		{WeaponID: 4, Quantity: 10, RawFlags: 0},
		{WeaponID: 4, Quantity: 2, RawFlags: 2},
		{WeaponID: 24, Quantity: 20, RawFlags: 0},
		{WeaponID: 13, Quantity: 3, RawFlags: 0},
		{WeaponID: 4, Quantity: 15, RawFlags: 4},
		{WeaponID: 11, Quantity: 2, RawFlags: 2},
		{WeaponID: 37, Quantity: 1, RawFlags: 0},
		{WeaponID: 31, Quantity: 6, RawFlags: 0},
	},
}

// antaranFortressSeedBase 是 `P` 的原始整數算式結果：
// signed_div(signed_div(5*word_180140, 4), 2) = 750。
const antaranFortressSeedBase = 750

// antaranFortressWeaponSlots 是 `Load_Antaran_Star_Fortress_` @ 0x4D18E
// 直接寫出的四個非空槽。ID／seed／raw flags／CapacityCap 為已證實；99/198
// 是 loader 後段的容量上限與拆槽規則，不是固定 runtime 數量。完整的 flags
// bit 名稱仍未知，因此不把它們改名成 ARM/FST。class 6 的 raw table power
// 是 0x17F69C = 900；它是 loader 的設計欄位，不可誤當成舊探針報出的 25，
// 也不在這裡與 remake 的 Doom Star Strength 代理混算。
var antaranFortressWeaponSlots = []antaranWeaponSlot{
	{WeaponID: 11, Seed: antaranFortressSeedBase / 2, CapacityCap: 99, RawFlags: 2},
	{WeaponID: 4, Seed: antaranFortressSeedBase / 4, CapacityCap: 198, RawFlags: 0},
	{WeaponID: 4, Seed: antaranFortressSeedBase / 4, CapacityCap: 198, RawFlags: 4},
	{WeaponID: 4, Seed: antaranFortressSeedBase / 2, CapacityCap: 99, RawFlags: 2},
}

func antaranWeaponSlotsFor(name string) []antaranWeaponSlot {
	// 每艘單位拿自己的 slice，避免未來某個戰鬥階段修改 runtime record 時污染證據範本。
	return append([]antaranWeaponSlot(nil), antaranKnownWeaponSlots[name]...)
}

func antaranFortressSlots() []antaranWeaponSlot {
	return append([]antaranWeaponSlot(nil), antaranFortressWeaponSlots...)
}

// antaranFortressDivisor 轉寫 sub_6EE8E @ 0x6EE8E 對要塞四槽的已追回中間鏈。
// techTier 是 sub_6E70A 的 live tech 結果；目前只把已觀測的 T=2、T=3 與其他
// bucket 分開，沒有把科技 byte 重新命名成未證實的玩法名稱。raw 0/2/4 是這四槽
// 已證實的輸入；其他 raw flags 失敗即回 0，避免虛構百分比效果。
func antaranFortressDivisor(weaponID int, rawFlags uint16, techTier int) int {
	k := 0
	switch weaponID {
	case 4:
		k = 15
	case 11:
		k = 30
	default:
		return 0
	}

	percent := 0
	switch rawFlags {
	case 0:
		percent = 100
	case 2:
		percent = 200
	case 4:
		percent = 50
	default:
		return 0
	}

	// sub_6A636 @ 0x6A636 supplies the first percentage (100/200/50).
	// sub_6A406 @ 0x6A406 is zero for raw 0/2/4, and the fortress CX=1
	// branch contributes no modeExtra, so X reduces to K*percent/100.
	x := k * percent / 100
	base := 400
	switch techTier {
	case 2:
		base = 700
	case 3:
		base = 600
	}
	divisor := (base*x + 500) / 1000
	if divisor < 1 {
		return 1
	}
	return divisor
}

// antaranFortressRuntimeQuantity 以 seed／raw flags／live tech bucket 計算一槽的
// runtime quantity，最後才套用已證實的 99/198 CapacityCap。這個 helper 不直接
// 改變目前終局快速戰鬥的 full-cap policy，讓沒有原版 live empire state 時仍維持
// 可玩的終局規模；它為日後接入實際 tech state 保留正確的除法邊界。
func antaranFortressRuntimeQuantity(slot antaranWeaponSlot, techTier int) int {
	if slot.Seed <= 0 || slot.CapacityCap <= 0 {
		return slot.Quantity
	}
	divisor := antaranFortressDivisor(slot.WeaponID, slot.RawFlags, techTier)
	if divisor <= 0 {
		return 0
	}
	quantity := slot.Seed / divisor
	if quantity > slot.CapacityCap {
		return slot.CapacityCap
	}
	if quantity < 0 {
		return 0
	}
	return quantity
}

// antaranHomeFleetDefense 是安塔蘭母星防禦艦隊的戰力與艦級組成。
//
// ============ 2026-08-08(第 49 項(安塔蘭防禦艦隊)):從「保守預設」換成反組譯真值 ============
//
// 這裡原本是 `{64, 64, 64, 64, 64, 64}`——六艘同級,註解寫著「手冊/openorion2 皆無精確數字,
// remake 保守預設,待考證」。**兩件事都被推翻了:數字有,而且組成根本不是同級的。**
//
// `Load_Antaran_Defense_Fleet_` @ 0x4D141(整支只有 77 bytes):
//
//	if word_199182 < 1: word_199182 = 1     ; 第 5 格(最大艦)強制至少 1
//	for bx = 0..4:                           ; ★ 五種艦體尺寸
//	    while cx < word_19917A[bx*2]:        ; ★ 每種尺寸的數量
//	        Load_Combat_Antaran_Ship_(...)
//	Load_Antaran_Star_Fortress_()            ; ★ 外加一座星際要塞
//
// 數量的上限來自靜態表 `_n_max_antaran_def_ships`(`byte_181746` @ 0x181746),
// 逐位元組解出來是 **{0, 0, 3, 2, 7, 0, 0, 0, 0}**:
//
//	| 索引 | 艦體尺寸 | 上限 |
//	|---|---|---|
//	| 0 | Small(Raider) | **0——永遠不造** |
//	| 1 | Medium(Marauder) | **0——永遠不造** |
//	| 2 | Large(Intruder) | 3 |
//	| 3 | Huge(Interdictor) | 2 |
//	| 4 | Titan(Harbinger) | 7 |
//
// 合計 **12 艘 + 1 座星際要塞**。難度**不改變組成**,只改變累積速度與逐艦裝甲加成。
//
// ⚠ **艦體尺寸的對照是推論,不是查表。** 原版的五級(Small/Medium/Large/Huge/Titan)
// 與 remake 的六級戰力階梯(shipStrength)沒有現成的對照表。這裡取「相對順序保持不變、
// 最大的安塔蘭艦對到 remake 最頂端」:Large→戰艦(16)、Huge→泰坦(32)、Titan→末日之星(64)。
// 合計 3×16 + 2×32 + 7×64 = 560。**這一層是 remake 的映射選擇**,數量與分層是真值。
//
// ⚠ **這是「養滿之後」的上限。** 原版的數量是執行期累積的(`word_19917A` 是 BSS,
// 載入時為 0),`Build_Antaran_Defensive_Ships_` 逐步補到上限;開局就打過去理論上只會遇到
// 保底的 1 艘 Harbinger + 要塞。remake 沒有安塔蘭的資源累積模型,取**上限**
// ——那是「終局一戰」該有的樣子,也與先前那個固定值的用意一致。
//
// 詳細推導見 `docs/re/antaran-defense-fleet.md`。
var antaranHomeFleetDefense = buildAntaranHomeFleetDefense()

// 安塔蘭母星防禦艦隊各尺寸的上限(原版 `_n_max_antaran_def_ships`,逐位元組解出)。
const (
	antaranDefLargeCount = 3 // Intruder
	antaranDefHugeCount  = 2 // Interdictor
	antaranDefTitanCount = 7 // Harbinger
)

// buildAntaranHomeFleetDefense 依上表組出防禦方的逐艦戰力,外加那座星際要塞。
//
// 星際要塞用 remake 既有的軌道防禦換算(`gamedata.StarFortressSpace`,與軌道轟炸那條
// 共用同一把尺,不另立係數)——原版是 `Load_Antaran_Star_Fortress_` @ 0x4D18E 載入的
// 一整套設計,remake 沒有那個粒度。
func buildAntaranHomeFleetDefense() []antaranDefenseUnit {
	out := make([]antaranDefenseUnit, 0, antaranDefLargeCount+antaranDefHugeCount+antaranDefTitanCount+1)
	for i := 0; i < antaranDefLargeCount; i++ {
		out = append(out, antaranDefenseUnit{
			OriginalName: "Intruder", Strength: shipStrength("戰艦"), CombatClass: gamedata.SHIP_BATTLESHIP,
			CombatLoaderRaw: 0x55738, FighterBayWeaponID: 31, FighterBayCount: 3,
			WeaponSlots: antaranWeaponSlotsFor("Intruder"),
		})
	}
	for i := 0; i < antaranDefHugeCount; i++ {
		out = append(out, antaranDefenseUnit{
			OriginalName: "Interdictor", Strength: shipStrength("泰坦"), CombatClass: gamedata.SHIP_TITAN,
			CombatLoaderRaw: 0x55B12,
			WeaponSlots:     antaranWeaponSlotsFor("Interdictor"),
		})
	}
	for i := 0; i < antaranDefTitanCount; i++ {
		out = append(out, antaranDefenseUnit{
			OriginalName: "Harbinger", Strength: shipStrength("末日之星"), CombatClass: gamedata.SHIP_DOOMSTAR,
			CombatLoaderRaw: 0x55F67, FighterBayWeaponID: 31, FighterBayCount: 6,
			WeaponSlots: antaranWeaponSlotsFor("Harbinger"),
		})
	}
	// 星際要塞的艦體強度仍沿用 remake 的 Doom Star 尺度代理；但原版已追回
	// 四個直接武器槽，這些槽會在 antaranDefenseCombatant 進入終局齊射，
	// 不再把要塞當成只有一個 64 戰力的空殼。
	out = append(out, antaranDefenseUnit{
		OriginalName: "Star Fortress", Strength: shipStrength("末日之星"),
		CombatClass: gamedata.SHIP_DOOMSTAR, Fortress: true, CombatLoaderRaw: 0x4D18E,
		WeaponSlots: antaranFortressSlots(),
	})
	return out
}

// antaranWeaponFirepower 彙整一個安塔蘭設計對「艦艇」齊射可消費的直接火力。
// 炸彈只對行星有效，戰機艙另走 Fighter Bay 狀態機，兩者不在這個總和重複計算。
// 一般艦使用 raw Quantity；星際要塞目前採終局 full-cap policy，所以使用
// CapacityCap，而不是假裝某個未知 live tech bucket 的 runtime quantity。raw flags
// 的正式名稱尚未知，先保留原值但不擅自改變 OrigWeaponTable 的傷害範圍。
func antaranWeaponFirepower(unit antaranDefenseUnit) (minDamage, maxDamage int) {
	for _, slot := range unit.WeaponSlots {
		quantity := slot.Quantity
		if unit.Fortress && slot.CapacityCap > 0 {
			quantity = slot.CapacityCap
		}
		if slot.WeaponID < 0 || slot.WeaponID >= len(gamedata.OrigWeaponTable) || quantity <= 0 {
			continue
		}
		weapon := gamedata.OrigWeaponTable[slot.WeaponID]
		switch weapon.Cat {
		case gamedata.WeaponCatBomb, gamedata.WeaponCatFighterBay:
			continue
		}
		minDamage += quantity * weapon.DamageMin
		maxDamage += quantity * weapon.DamageMax
	}
	return minDamage, maxDamage
}

// antaranDefenseCombatant 把一艘已知安塔蘭防禦設計轉成現有快速戰鬥模型。
// 艦體基礎仍沿用 remake 的 Strength 代理；已證實的標準戰機艙則沿用玩家快速
// 戰鬥同一組「一艙戰機的攻擊／結構」消費端，避免另造一套敵方近似數字。
func antaranDefenseCombatant(unit antaranDefenseUnit) combatant {
	atk := unit.Strength
	hp := unit.Strength * 3
	if unit.FighterBayWeaponID == 31 && unit.FighterBayCount > 0 {
		fighterAttack, fighterHP := gamedata.FighterBayCombatContribution()
		atk += unit.FighterBayCount * fighterAttack
		hp += unit.FighterBayCount * fighterHP
	}
	wmin, wmax := atk/2, atk
	if unit.Fortress {
		// 要塞的四槽已追回，讓它們真正進入 battleShot 的 wmin/wmax。
		// atk 仍保留 Doom Star 艦體的命中基礎，再加直接火力的中位量，
		// 避免把「傷害總和」誤當成單一 BA；這是快速戰鬥編排層的量綱映射，
		// 不是對 raw +0x36／+0x52 欄位的重新命名。
		weaponMin, weaponMax := antaranWeaponFirepower(unit)
		if weaponMax > 0 {
			wmin, wmax = weaponMin, weaponMax
			atk += (weaponMin + weaponMax) / 2
		}
	}
	return combatant{
		hp: hp, atk: atk, def: unit.Strength,
		wmin: wmin, wmax: wmax, armor: unit.Strength,
		sizeClass: unit.CombatClass,
	}
}

// dimensionalPortalBuildingName 是次元傳送門在 s.ColonyBuildings 去重 map 裡的 key
// (gamedata.Buildings 裡該項的 NameZH,見 buildings.go)。
const dimensionalPortalBuildingName = "次元傳送門"

// hasDimensionalPortal 判定玩家是否已在任一殖民地建成次元傳送門。
//
// remake 簡化(已誠實記錄於 docs/tech/victory-conditions.md):手冊原文是「select a fleet in the
// same system as the portal」——反攻前置理論上要求「艦隊與傳送門同星系」,但本 remake 的星際
// 航行模型(FleetAtStar/FleetDestStar)不追蹤「殖民地建築在哪個星系」與「艦隊目前在哪個星系」
// 兩者的可達性交叉比對這麼細,故簡化為「玩家帝國內任一殖民地已建成即視為前置滿足」。
func (s *GameSession) hasDimensionalPortal() bool {
	for _, built := range s.ColonyBuildings {
		if built != nil && built[dimensionalPortalBuildingName] {
			return true
		}
	}
	return false
}

// CanAssaultAntares 回傳「玩家現在是否能發起反攻安塔蘭母星」——供 UI 決定是否顯示/啟用按鈕,
// 也是 AssaultAntares 內部檢查的匯出版本(判斷邏輯只寫一份,兩處共用)。
func (s *GameSession) CanAssaultAntares() bool {
	return !s.Victory.Over && !s.DisableEvents && s.hasDimensionalPortal() && len(s.Fleet().Ships) > 0
}

// AssaultAntares 解算「反攻安塔蘭母星」戰鬥(手冊三條勝利路徑之二)。
//
// 前置條件(CanAssaultAntares,不滿足則 ok=false,不消耗艦隊、不觸發戰鬥、不寫 LastBattle):
//   - 遊戲尚未結束(s.Victory.Over)。
//   - 手冊「This strategy is not available if you disabled Antaran Attacks when setting up
//     your game」→ s.DisableEvents 視為關閉安塔蘭攻擊時一併關閉本反攻路徑。
//   - 已建次元傳送門(hasDimensionalPortal)。
//   - 玩家艦隊非空(手冊「select a fleet in the same system as the portal」的最小化對應)。
//
// 戰鬥沿用 ResolveBattle 同款 battleVolley 逐回合解算(每回合雙方齊射,最多 6 回合),防禦方
// 戰力用 antaranHomeFleetDefense(反組譯真值的組成,見該變數註解)。與 ResolveBattle 不同:這是「終局一戰」,PlayerWon
// 要求防禦方**全滅**(len(defenders)==0),不是 ResolveBattle 那種「艦數比較多也算贏」的寬鬆
// 判定——手冊原文「Once you defeat the awe-inspiring Antarans」語意是徹底擊敗,不是打退。
//
//   - 玩家戰勝(防禦方全滅)→ s.AntaranHomeworldConquered=true,下一次 advanceAntaranVictory
//     (EndTurn 呼叫)會偵測並設定 s.Victory=VictoryAntaran。
//   - 玩家戰敗(6 回合內未能全殲防禦方,或己方全滅)→ 不設定勝利旗標,套用艦隊損失
//     (比照 ResolveBattle 呼叫 removeWeakestShip),回傳戰鬥結果供 UI/回合摘要顯示。
//
// 回傳 (BattleResult, ok)——ok=false 代表前置條件不滿足,BattleResult 為零值,呼叫端不應顯示
// 戰鬥結果畫面。
func (s *GameSession) AssaultAntares() (BattleResult, bool) {
	s.recordPlayerCommand(PlayerCommand{Name: CmdAssaultAntares})
	if !s.CanAssaultAntares() {
		return BattleResult{}, false
	}

	mkPlayer := func() []combatant {
		var out []combatant
		galacticLoreBonus := s.galacticLoreCombatBonus()
		for _, sh := range s.Fleet().Ships {
			body := shipStrength(sh.Class)
			atk := body + sh.WeaponAttack + galacticLoreBonus
			atk += atk * s.RaceCombatPct / 100 // 種族戰鬥加成,比照 ResolveBattle mkPlayer
			out = append(out, combatant{hp: body * 3, atk: atk, def: body, wmin: atk / 2, wmax: atk,
				shield: shieldReduceByName(sh.Shield), armor: armorHPByName(sh.Armor),
				kind: weaponKindByName(sh.Weapon)})
		}
		return out
	}
	var df []combatant
	for _, unit := range antaranHomeFleetDefense {
		df = append(df, antaranDefenseCombatant(unit))
	}
	pf := mkPlayer()

	res := BattleResult{Enemy: "安塔蘭母星防禦艦隊", PlayerStart: len(pf), EnemyStart: len(df)}
	// 種子與 ResolveBattle 用不同 offset(987654321 vs 12345),避免同一回合呼叫兩個戰鬥函式時
	// 巧合共用同一亂數序列。
	rng := rand.New(rand.NewSource(int64(s.Turn)*2654435761 + 987654321))
	for round := 1; round <= 6 && len(pf) > 0 && len(df) > 0; round++ {
		eDestroyed := battleVolley(pf, &df, rng)
		pDestroyed := battleVolley(df, &pf, rng)
		res.Log = append(res.Log, fmt.Sprintf("第 %d 回合:擊沉安塔蘭艦 %d ／ 我方損失 %d", round, eDestroyed, pDestroyed))
	}
	res.PlayerLosses = res.PlayerStart - len(pf)
	res.EnemyLosses = res.EnemyStart - len(df)
	res.PlayerWon = len(df) == 0 // 終局一戰:要求防禦方全滅,見函式註解(比 ResolveBattle 嚴格)
	for i := 0; i < res.PlayerLosses; i++ {
		s.removeWeakestShip()
	}
	s.LastBattle = &res
	if res.PlayerWon {
		s.AntaranHomeworldConquered = true
	}
	return res, true
}

// advanceAntaranVictory 是 EndTurn 每回合呼叫的狀態機:偵測 AssaultAntares 是否已戰勝
// (AntaranHomeworldConquered),沿用 engine.CheckAntaranVictory 純函式判定(不重算邏輯),只做
// shell 層「設 Victory 狀態」的整合,比照 advanceConquestVictory/advanceCouncil 同款模式。
func (s *GameSession) advanceAntaranVictory() {
	if s.Victory.Over {
		return
	}
	if !engine.CheckAntaranVictory(s.AntaranHomeworldConquered) {
		return
	}
	s.queueAntaranDefeatBroadcast()
	s.Victory = VictoryState{Over: true, Reason: engine.VictoryAntaran, Winner: "player", Turn: s.Turn}
}

// AntaranDefenseStrength 回傳安塔蘭母星防禦艦隊的總戰力(供安塔蘭房間畫面顯示戰力對比)。
// 值來自 antaranHomeFleetDefense——**組成(3 Large / 2 Huge / 7 Titan + 1 要塞)是反組譯真值**,
// 艦體尺寸對到 remake 戰力階梯那一層是推論,
// 理由見該變數的註解。
func AntaranDefenseStrength() int {
	n := 0
	for _, unit := range antaranHomeFleetDefense {
		n += unit.Strength
	}
	return n
}

// AntaranDefenseShipCount 回傳安塔蘭母星防禦艦隊的艦數(同上,供畫面顯示)。
func AntaranDefenseShipCount() int { return len(antaranHomeFleetDefense) }

// PlayerFleetStrength 回傳玩家艦隊總戰力(安塔蘭房間用來與上面兩個值對比)。
// 與 AI 態勢判斷共用同一個 playerMilitary,避免畫面上顯示的數字跟實際結算用的不是同一套。
func (s *GameSession) PlayerFleetStrength() int { return s.playerMilitary() }

// AssaultAntaresBlockReason 回傳「為什麼現在不能發動終局反攻」;可以發動時回空字串。
// CanAssaultAntares 只回 bool,玩家看不出卡在哪一條——把四個前置條件逐條講白。
func (s *GameSession) AssaultAntaresBlockReason() string {
	switch {
	case s.Victory.Over:
		return "對局已分出勝負"
	case s.DisableEvents:
		// 手冊 p.183:「This strategy is not available if you disabled Antaran Attacks
		// when setting up your game.」
		return "本局關閉了安塔蘭攻擊,此勝利路徑不可用"
	case !s.hasDimensionalPortal():
		return "尚未建成「" + dimensionalPortalBuildingName + "」——沒有它到不了安塔蘭母星"
	case len(s.Fleet().Ships) == 0:
		return "沒有艦隊可派"
	}
	return ""
}

// GrantDimensionalPortalForGallery 直接把次元傳送門記進第一個殖民地的建築清單。
//
// **僅供 headless 截圖廊使用**:安塔蘭王座廳的「發動終局反攻」按鈕要有傳送門才會亮,
// 而正常玩到多維物理再蓋出傳送門要幾十回合,截圖驗證等不起。正常遊戲流程不會呼叫它
// ——傳送門一律得靠 gamedata.Buildings 的建造佇列蓋出來。
func (s *GameSession) GrantDimensionalPortalForGallery() {
	if len(s.PlayerColonies) == 0 {
		return
	}
	for len(s.ColonyBuildings) < len(s.PlayerColonies) {
		s.ColonyBuildings = append(s.ColonyBuildings, map[string]bool{})
	}
	if s.ColonyBuildings[0] == nil {
		s.ColonyBuildings[0] = map[string]bool{}
	}
	s.ColonyBuildings[0][dimensionalPortalBuildingName] = true
}
