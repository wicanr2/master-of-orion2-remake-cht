package gamedata

// 艦艇設計「空間格」模型:各艦體(hull)的總可用空間,以及元件佔用的空間。
//
// 來源(權威,可直接查證,非掃描圖):moo2_patch1.5/GAME_MANUAL.pdf(patch 1.5 隨附完整手冊,
// 188 頁,pdftotext 抽字確認可正常抽取文字,非影像 PDF)。逐項出處見 docs/tech/ship-design-space.md。
//   - 艦體總空間表:p.121「Class / Cost / Space / Marines / Armor / Struct. / Comp. / Drive / Shield」表。
//   - 武器佔格(Weapons Area 的 Size 欄):p.124(束射 Beam)、p.125-126(飛彈/魚雷/炸彈)、
//     p.126-127(戰機/特殊武器)。
//
// [HARD 誠實原則] 手冊 p.128-129(Design Dock 說明)明白指出:「Note that, contrary to weapons,
// special systems cost more to install in larger ships and take up more space in a larger hull.」
// ——確認「特殊系統佔格依艦體大小縮放」這個機制存在,但手冊全文(GAME_MANUAL.pdf 188 頁 +
// MANUAL_150.html)都沒有給出任何一個特殊系統的精確空間數字或百分比公式。本檔的
// SpecialSpaceEstimatePercent 是誠實標註的**估計值**,不是手冊數字,見該常數註解與
// docs/tech/ship-design-space.md 的待辦清單。
//
// [HARD 誠實原則 2] 手冊 p.121-122 同時指出:裝甲(Armor)、護盾(Shield)、電腦(Computer)、
// 引擎(Drive)是「Automatics」——每艘船自動裝上目前科技最好的一套(可退而求其次裝較差的
// 護盾/電腦以省成本),原文:「Don't worry, the engines, armor, and fuel cells do not take up
// space that might be used for optional systems.」也就是說,**真正佔用「Weapons
// Area」/「Specials Area」共用空間預算的只有武器與特殊系統兩類,不含裝甲/護盾**。本專案既有的
// 簡化模型(internal/shell/session.go 的四下拉:武器/裝甲/護盾/特殊)在裝甲/護盾這兩項已經是
// 對手冊的簡化(手冊沒有「選裝甲/護盾佔格」這件事);`shell.ShipDesignSpaceUsed` 為保留既有呼叫
// 介面仍接受裝甲/護盾參數,但依手冊行為兩者一律不計入空間,不是遺漏或臆造。

// ShipHullSpace 回傳各艦體等級的總空間(可裝載武器/特殊系統的空間上限)。
// 資料來源:GAME_MANUAL.pdf p.121 表格,索引沿用既有 CombatShipClass(0=Frigate..5=DoomStar,
// 與 formulas.go 的 computerHPTable/driveHPTable 同一組 size 索引,兩表已交叉核對一致,
// 見 docs/tech/ship-design-space.md)。
func ShipHullSpace(class CombatShipClass) int {
	if class < 0 || int(class) >= len(shipHullSpaceTable) {
		return 0
	}
	return shipHullSpaceTable[class]
}

// shipHullSpaceTable 各艦體總空間(GAME_MANUAL.pdf p.121:Frigate 25 / Destroyer 60 /
// Cruiser 120 / Battleship 250 / Titan 500 / Doom Star 1200)。
var shipHullSpaceTable = [6]int{25, 60, 120, 250, 500, 1200}

// WeaponSpaceByName 各武器元件的佔用空間(手冊「Size」欄,p.124-127),對照
// internal/shell/session.go 的 WeaponOptions 元件名(Component.Name)。
//
// 武器佔格是**固定值**,不隨艦體大小縮放(手冊全文未提及武器佔格會依艦體改變;會依艦體大小
// 縮放的只有「特殊系統」,見下方 SpecialSpaceEstimatePercent)。
//
// 飛彈類(核飛彈/麥克萊特飛彈)手冊給的是「依彈架容量(x2/x5/x10/x15/x20)遞增的一組值」
// (10/20/30/35/40,p.125「size and cost per rack size」),不是單一值;本專案簡化模型(session.go)
// 未實作彈架容量選擇,取最小彈架(x2 = 10)當估計值,已於下表逐項標註「確認值」vs「估計」。
var WeaponSpaceByName = map[string]int{
	"雷射":     10, // Laser Cannon,p.124 確認值(手冊表列 Size=10)
	"核飛彈":    10, // Nuclear Missile,估計(手冊實際為彈架 10/20/30/35/40,取最小彈架 x2)
	"質量投射器":  10, // Mass Driver,p.124 確認值
	"中子爆破槍":  10, // Neutron Blaster,p.124 確認值
	"核融合光束":  10, // Fusion Beam,p.124 確認值
	"麥克萊特飛彈": 10, // Merculite Missile,估計(同核飛彈,取最小彈架 x2)
	"高斯砲":    10, // Gauss Cannon,p.124 確認值
	"相位砲":    10, // Phasor,p.124 確認值
	"電漿砲":    25, // Plasma Cannon,p.124 確認值(1.31/1.50 傷害係數不同但 Size 不受影響,
	// 見 docs/tech/component-values.md 的版本相依記錄)
	"死光": 30, // Death Ray,p.124 確認值
}

// SpecialSpaceEstimatePercent 是「特殊系統佔空間 = 艦體總空間的 X%」的估計係數。
//
// 手冊確認「依艦體大小縮放」這個機制存在(見檔案開頭 [HARD 誠實原則]),但沒有給出任何精確數字。
// 5% 是刻意保守、非手冊來源的估計值,只用來讓空間驗證函式在目前簡化模型下有個非零、合理量級的
// 佔格數,避免「特殊系統完全不佔空間」這個更失真的預設。**這不是手冊數字**,精確值待查
// (見 docs/tech/ship-design-space.md 待辦)。
const SpecialSpaceEstimatePercent = 5

// SpecialSpace 回傳指定艦體空間下,裝一套特殊系統的估計佔格(hullSpace * SpecialSpaceEstimatePercent%,
// 無條件捨去,最少 1 格)。hasSpecial=false(未選裝特殊系統)回 0。
func SpecialSpace(hullSpace int, hasSpecial bool) int {
	if !hasSpecial || hullSpace <= 0 {
		return 0
	}
	sp := hullSpace * SpecialSpaceEstimatePercent / 100
	if sp < 1 {
		sp = 1
	}
	return sp
}
