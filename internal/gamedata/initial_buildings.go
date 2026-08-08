package gamedata

// initial_buildings.go:**開局母星已建成的建築**——優先清單 + 上限。
//
// ============ 這份清單先前只存在於手冊的一句話裡 ============
//
// `shell.StartingBuildingCount` 的註解寫著:
//
//	此函式只回傳「上限」,實際會生成哪些建築仍取決於 **initial_buildings 優先清單**與已知科技
//
// 那份清單 2026-08-07 從 `Init_Homeworld_Colony2_` @ 0x13A3D 挖到了:
//
//	loc_13CCB:
//	    movsx ebx, word_17D8AC[ecx]      ; ← 優先清單(word 陣列)
//	    test  ebx, ebx / jz 結束          ; 0 收尾
//	    movzx eax, byte_199CB5            ; TECH LEVEL
//	    movzx ax, ds:byte_13A3A[eax]      ; ← 上限表
//	    cmp   ax, [已放幾棟] / jle 結束
//
// **上限表 `byte_13A3A` = db 3, 5, 9** —— 與手冊逐字相同
//(「capped to 3 for Pre-warp, 5 for Average/Postwarp and 9 for Advanced」)。
// remake 的 `shell.BuildingCapPreWarp/Average/Advanced` 早就是這三個數,現在有了第二個來源。
//
// ============ 四條獨立的線指到同一個答案 ============
//
// 手冊說「Pre-warp and Average Tech games only start with **Marine Barracks and a Star Base**
// because no other techs are Known that are also in the default initial buildings list」。
// 拿這份清單 × 第 31 項挖到的六個開局主題 × remake 自己的建築前置表跑一遍:
// 清單裡科技條件成立的**正好只有 `Star Base`(40)與 `Marine Barracks`(22)兩棟**。
//
// 清單本身、開局主題表、建築前置表、手冊那句話——四個獨立來源互相對上,不是巧合。
//
// ============ 順序有意義 ============
//
// 清單開頭是 41 → 8 → 40,也就是 **Star Fortress → Battlestation → Star Base**:
// 同一條防禦升級鏈的**最強的排最前面**。原版是照這個順序往下發到上限為止,
// 所以科技越先進,拿到的就是越上層的那一棟。照抄順序,不重新排。

// InitialBuildingOrder 是原版 `word_17D8AC` 的優先清單(原版建築編號,順序照原版)。
//
// ⚠ 37 與 18 兩個編號 remake 的 `OrigBuildingID` 對照表裡沒有——**保留在清單裡**而不是刪掉:
// 刪掉會讓順序看起來是連續的,而實際上中間缺了兩個。解析不到就跳過(見 InitialBuildings)。
var InitialBuildingOrder = []int{
	41, // Star Fortress
	8,  // Battlestation
	40, // Star Base
	21, // Hydroponic Farm
	22, // Marine Barracks
	15, // Biospheres
	7,  // Automated Factories
	20, // Holo Simulator
	37, // ⚠ remake 對照表沒有
	4,  // Astro University
	31, // Pleasure Dome
	33, // Recyclotron
	34, // Robotic Factory
	39, // Spaceport
	2,  // Armor Barracks
	12, // Deep Core Mine
	13, // Core Waste Dumps
	32, // Pollution Processor
	35, // Research Laboratory
	10, // Cloning Center
	43, // Subterranean Farms
	16, // Food Replicators
	28, // Planetary Radiation Shield
	25, // Planetary Gravity Generator
	18, // ⚠ remake 對照表沒有
	19, // Galactic Cybernet
	24, // Planetary Flux Shield
	26, // Missile Base
	47, // Fighter Garrison
	27, // Ground Batteries
	6,  // Autolab
	5,  // Atmospheric Renewer
}

// InitialBuildingCap 是各 TECH LEVEL 的開局建築數上限(原版 `byte_13A3A` = 3, 5, 9)。
//
// 超出 0..2 回「一般」那一級的 5:NEW GAME 只給三級(見 StartingTopicCount 的同款說明)。
func InitialBuildingCap(techLevel int) int {
	switch techLevel {
	case 0:
		return 3
	case 1:
		return 5
	case 2:
		return 9
	}
	return 5
}

// InitialBuildings 依優先清單挑出開局該建成的建築名(remake 的中文名),最多 limit 棟。
//
// 規則照原版:**照清單順序**逐個看,科技條件成立的就收,收滿 limit 為止。
// 不是「挑最好的」也不是隨機——順序本身就是原版的優先序(見檔頭)。
//
// known 為 nil 時視為什麼都不知道(回空),不是「全部都知道」——
// 開局狀態誤判成全解會讓母星憑空多出十幾棟建築。
func InitialBuildings(known map[ResearchTopic]bool, limit int) []string {
	if limit <= 0 || len(known) == 0 {
		return nil
	}
	byID := map[int]Building{}
	for _, b := range Buildings {
		if id, ok := OrigBuildingID[b.NameEN]; ok {
			byID[id] = b
		}
	}
	out := make([]string, 0, limit)
	for _, id := range InitialBuildingOrder {
		b, ok := byID[id]
		if !ok {
			continue // 對照表沒有這個編號(見 InitialBuildingOrder 的說明)
		}
		if !known[b.PrereqTopic] {
			continue
		}
		out = append(out, b.NameZH)
		if len(out) >= limit {
			break
		}
	}
	return out
}
