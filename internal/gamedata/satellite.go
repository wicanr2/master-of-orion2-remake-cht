// satellite.go:軌道防禦基地/衛星「space 預算 → 塞武器 → 推導反擊戰力」模型
// (diff 全量表 #14「衛星/地面砲台佔格」,docs/tech/version-1.3-1.5-diff.md)。
//
// 設計背景:1.3/1.5 對「衛星光束武器 arc-cost」有真差異——1.3 衛星 arc-cost +25%、地面砲台
// +0%;1.5 統一改過(CHANGELOG_150.TXT 1.50.7),衛星最終 +33.3%(1.50.10 再修一次)、地面
// 砲台 +50%。本檔把這差異接進 internal/shell 的軌道防禦反擊戰力推導(retaliationAttackers),
// 取代先前純手動指派的 shipStrength tier 對照表(4/8/16),讓反擊戰力真的「隨科技變強」+
// 「隨版本 arc-cost 不同而不同」,而不是固定寫死的三個數字。
//
// [手冊確認值] moo2_patch1.5/GAME_MANUAL.pdf(pdftotext -layout 逐字抽取,非 OCR):
//   - p.78「Missile Base」原文:「The planetary Missile Base is a defensive emplacement
//     equipped with as many launchers full of the best missiles you have as will fit in
//     300 space.」→ MissileBaseSpace = 300,確認值。
//   - p.81「Ground Batteries」原文:「This building contains Heavy Mount and Point Defense
//     versions of your best available beam weapons – as many as fit in 450 space.」→
//     GroundBatterySpace = 450,確認值(手冊同時說明地面砲台預裝的是 HV/PD mod 版本武器,
//     本 remake 簡化模型不逐一套用 mod 佔格,只用基礎武器 + arc-cost 近似,見下方
//     GroundBatteryBeamArcCostPct 使用處的誠實標註)。
//
// [近似值,誠實標註] Star Base(p.75)/Battlestation(p.79)/Star Fortress(p.83)三個「軌道
// 基地(satellite)」手冊只有質化描述(「much better armed than a Battlestation」「adds ...
// +20 to the Beam Attack」等),全文找不到任何一個具體 space 數字。本檔借用
// shipspace.go ShipHullSpace 同一組艦體空間值(Battleship=250 / Titan=500 / Doom Star=1200)
// 當估計:三座軌道基地彼此「取代不疊加、火力遞增」的定性順序(星辰要塞 > 戰鬥站 > 星基,見
// buildings.go colonyBestOrbitalBase)與三個標準艦體等級的空間/火力遞增序一致,借用同一組
// 已知忠實數字只是「找不到手冊數字時,退而求其次借用同量級的另一組已知數字」,不是憑空捏造,
// 但仍是不折不扣的估計值,待更精確來源(如逆向 EXE 常數表)出現前應視為佔位。
package gamedata

const (
	StarBaseSpace      = 250  // 【近似】比照 ShipHullSpace(Battleship),見上方檔頭說明
	BattlestationSpace = 500  // 【近似】比照 ShipHullSpace(Titan)
	StarFortressSpace  = 1200 // 【近似】比照 ShipHullSpace(Doom Star)
	MissileBaseSpace   = 300  // 【確認】手冊 p.78
	GroundBatterySpace = 450  // 【確認】手冊 p.81
)

// SatelliteStrengthScale 是「rawStrength(= 塞入的武器把數 × 單把 Value)換算成艦級 atk」的
// 校準除數。選 20 的推導(全程整數運算,雷射 Value=4、WeaponSpaceByName["雷射"]=10 為參考點,
// 套用下方 SatelliteBeamSpaceWithArc/SatelliteWeaponFitCount 兩個函式):
//
//	1.3(SatelliteBeamArcCostPct=25%):perBeam = 10 + 10*25/100 = 10+2 = 12。
//	  星基   (250 space):fit=250/12=20,raw=20*4=80,  atk=80/20  = 4。
//	  戰鬥站 (500 space):fit=500/12=41,raw=41*4=164, atk=164/20 = 8。
//	  星辰要塞(1200 space):fit=1200/12=100,raw=100*4=400,atk=400/20 = 20。
//
// 星基/戰鬥站兩個錨點(4、8)與本 remake 改用本模型之前的既有手動 tier(shipStrength 驅逐艦
// 4/巡洋艦 8)完全重現。星辰要塞算出 20,而非設計討論階段用「不做整數截斷、逐步以浮點 12.5」
// 粗估的 ≈19(1200/12.5=96,96*4/20=19.2→19)——差異來自本檔實作嚴格遵照
// WeaponSpaceWithMods 既有慣例(`base + base*pct/100` 全程整數運算,見 weapon_mods.go),
// 25% 對 base=10 算出整數 12(非 12.5),使星辰要塞的 fit 剛好整除(1200/12=100,無條件捨去
// 不生效),atk 因此是 20 而非約 19。20 相對於改動前的既有 tier 16 是同一數量級的合理成長
// (+25%,呼應「随 space 預算精算後變強」的設計目的),不是計算錯誤——本檔如實記錄
// 實際值 20,不假造成 19 只為了湊近似;平衡是否成立以下方(shell 層)「平衡 sanity」實測
// (轟炸損艦數)為準,不是以「是否重現到小數點」為準。
const SatelliteStrengthScale = 20

// SatelliteBeamSpaceWithArc 依基礎佔格(如 WeaponSpaceByName 查到的值)套用版本相依的 beam
// arc-cost 百分比,回傳佔格後的實際空間。公式與運算慣例比照 WeaponSpaceWithMods:
// base + base*pct/100,全程整數運算、無條件捨去,最少 1。
//
// 只對 beam 武器有意義——missile 不吃 arc-cost(手冊 Missile Base 完全沒提到 arc 概念,飛彈
// 是彈架容量制,見 shipspace.go 頂部說明),呼叫端對 missile 應直接用 WeaponSpaceByName 的
// 原始佔格,不要呼叫本函式。
func SatelliteBeamSpaceWithArc(baseSpace, arcCostPct int) int {
	space := baseSpace + baseSpace*arcCostPct/100
	if space < 1 {
		space = 1
	}
	return space
}

// SatelliteWeaponFitCount 回傳指定 space 預算下,能塞入多少把「每把佔 perWeaponSpace 格」的
// 武器(無條件捨去)。spaceBudget<=0 或 perWeaponSpace<=0 視為塞不下任何一把,回 0。
func SatelliteWeaponFitCount(spaceBudget, perWeaponSpace int) int {
	if spaceBudget <= 0 || perWeaponSpace <= 0 {
		return 0
	}
	return spaceBudget / perWeaponSpace
}
