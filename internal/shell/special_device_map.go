package shell

import "github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"

// special_device_map.go:把 remake 的特殊系統中文名對到**原版特殊裝置表**的編號
// (`gamedata.specialDeviceTable`,見 docs/tech/special-device-table.md)。
//
// 有了這張對照,佔格與成本就能改讀執行檔的真值——先前佔格走
// `gamedata.SpecialSpace()` 的 5% 估計、成本走 `Component.Cost` 的 remake 值。
//
// ⚠ **不是每一項都對得上,而且對不上有兩種完全不同的原因**:
//
//	① 原版把它歸在**武器表**(stride 0x1C)而不是特殊裝置表(stride 0x2F):
//	   牽引光束、停滯力場、反飛彈火箭、陀螺去穩器。remake 把它們放進同一個下拉是 UI 取捨,
//	   不是原版分類。它們的真值在武器表裡,那張表這一輪沒抽。
//	② 原版**根本沒有這一項**:戰鬥電腦(原版是獨立的電腦槽,不佔特殊系統位)、
//	   戰機庫/重戰機庫/轟炸機庫(原版是另一組欄位)。
//
// 兩種都保留舊行為(估計佔格 + `Component.Cost`),但 `TestSpecialDeviceMapCoverage`
// 會把未對上的名單釘死——**新增元件卻忘了對照,測試會紅**,不會安靜地退回估計值。
var specialDeviceByName = map[string]gamedata.SpecialDevices{
	"阿基里斯瞄準器": gamedata.SPEC_ACHILLES_TARGETING_UNIT,
	"增強引擎":    gamedata.SPEC_AUGMENTED_ENGINES,
	"自動修復":    gamedata.SPEC_AUTOMATED_REPAIR_UNIT,
	"戰鬥艙":     gamedata.SPEC_BATTLE_PODS,
	"戰鬥掃描器":   gamedata.SPEC_BATTLE_SCANNER,
	"隱形裝置":    gamedata.SPEC_CLOAKING_DEVICE,
	"位移裝置":    gamedata.SPEC_DISPLACEMENT_DEVICE,
	"電子干擾器":   gamedata.SPEC_ECM_JAMMER,
	"快速飛彈架":   gamedata.SPEC_FAST_MISSILE_RACKS,
	"硬化護盾":    gamedata.SPEC_HARD_SHIELDS,
	"重裝甲":     gamedata.SPEC_HEAVY_ARMOR,
	"高能聚焦":    gamedata.SPEC_HIGH_ENERGY_FOCUS,
	"超載電容":    gamedata.SPEC_HYPERX_CAPACITORS,
	"慣性抵消器":   gamedata.SPEC_INERTIAL_NULLIFIER,
	"慣性穩定器":   gamedata.SPEC_INERTIAL_STABILIZER,
	"閃電場":     gamedata.SPEC_LIGHTNING_FIELD,
	"多相護盾":    gamedata.SPEC_MULTIPHASED_SHIELDS,
	"多波電子干擾器": gamedata.SPEC_MULTIWAVE_ECM_JAMMER,
	"能量吸收器":   gamedata.SPEC_ENERGY_ABSORBER,
	"測距瞄準器":   gamedata.SPEC_RANGEMASTER_UNIT,
	"相位匿蹤":    gamedata.SPEC_PHASING_CLOAK,
	"保安站":     gamedata.SPEC_SECURITY_STATIONS,
	"重生程序":    gamedata.SPEC_REGENERATION,
	"強化船體":    gamedata.SPEC_REINFORCED_HULL,
	"偵察實驗室":   gamedata.SPEC_SCOUT_LAB,
	"結構分析儀":   gamedata.SPEC_STRUCTURAL_ANALYZER,
	"時間扭曲加速器": gamedata.SPEC_TIME_WARP_FACILITATOR,
	"部隊艙":     gamedata.SPEC_TROOP_PODS,
	"傳送器":     gamedata.SPEC_TRANSPORTERS,
	"廣域干擾器":   gamedata.SPEC_WIDE_AREA_JAMMER,
}

// specialDeviceSpaceFor 回傳一項特殊系統裝在指定艦級上的佔格。
//
// 對得上原版表的走真值(**可能是負數**,見戰鬥艙);對不上的退回舊的 5% 估計,
// 讓那幾項的行為與這一輪之前逐位元相同。
func specialDeviceSpaceFor(c Component, class gamedata.CombatShipClass) int {
	if c.Name == "" || c.Name == "無" {
		return 0
	}
	if dev, ok := specialDeviceByName[c.Name]; ok {
		return gamedata.SpecialDeviceSpace(dev, class)
	}
	// 退路①:原版把它歸在**武器表**(牽引光束/停滯力場/反飛彈火箭/戰機艙/突擊艇)。
	// 那張表的佔格是**單一數字**(不隨艦級變動),與特殊裝置表不同——武器在原版就是這樣。
	if w, ok := gamedata.OrigWeaponByTech(c.UnlockTech); ok {
		return w.Size
	}
	// 退路②:原版根本沒有這一項(戰鬥電腦)。維持舊的 5% 估計。
	return gamedata.SpecialSpace(gamedata.ShipHullSpace(class), true)
}

// specialDeviceCostFor 回傳一項特殊系統裝在指定艦級上的建造成本(BC)。
//
// 對得上原版表的走真值;對不上的退回元件清單裡的 `Component.Cost`(remake 值)。
func specialDeviceCostFor(c Component, class gamedata.CombatShipClass) int {
	if c.Name == "" || c.Name == "無" {
		return 0
	}
	if dev, ok := specialDeviceByName[c.Name]; ok {
		return gamedata.SpecialDeviceCost(dev, class)
	}
	// ⚠ 武器表的**成本**不走這條退路。理由見 gamedata/weapon_table.go 檔頭:
	// 執行檔的武器成本尺度與 remake 差約四倍,而艦體成本的方向相反,只換一邊會壞平衡。
	// 佔格沒有這個問題(佔格與艦體空間是同一個尺度,而艦體空間已經是原版真值)。
	return c.Cost
}

// playerHasMegafluxers 回傳玩家是否已研究巨型通量器(全船可用空間 ×125/100)。
//
// 判定沿用 groundEquipTechOwned 那套(主題完成 + 未明確抉擇即視為解鎖 / 已抉擇需選中)
// ——巨型通量器屬 TOPIC_HIGH_ENERGY_DISTRIBUTION 的三選一(另兩個是能量吸收器與高能聚焦)。
func (s *GameSession) playerHasMegafluxers() bool {
	return groundEquipTechOwned(s.Player, gamedata.TOPIC_HIGH_ENERGY_DISTRIBUTION,
		gamedata.TECH_MEGAFLUXERS)
}

// HullSpaceFor 回傳這局遊戲下,指定艦級目前的**可用**空間(含巨型通量器加成)。
//
// 套件級的 ShipDesignFits/ShipDesignSpaceUsed 沒有 GameSession 可查,一律以「沒有巨型通量器」
// 計算;有 session 的呼叫端(cmd/moo2 的造艦畫面)應該走這裡與 DesignFitsWithMods。
func (s *GameSession) HullSpaceFor(class string) int {
	classID, _ := shipClassFromName(class)
	return gamedata.DesignHullSpace(classID, s.playerHasMegafluxers())
}

// DesignFitsWithMods 同 ShipDesignFitsWithMods,但把巨型通量器的 +25% 可用空間算進去。
func (s *GameSession) DesignFitsWithMods(class string, weapon, armor, shield, special int, mods []string) bool {
	return ShipDesignSpaceUsedWithMods(class, weapon, armor, shield, special, mods) <=
		s.HullSpaceFor(class)
}

// DesignFitsWithModsAndArc 是艦艇設計畫面的含火線角空間判定；可用艦體空間
// 仍由本局科技狀態決定，與舊入口共用 HullSpaceFor。
func (s *GameSession) DesignFitsWithModsAndArc(class string, weapon, armor, shield, special int, mods []string, arc gamedata.WeaponArc) bool {
	return ShipDesignSpaceUsedWithModsAndArc(class, weapon, armor, shield, special, mods, arc) <=
		s.HullSpaceFor(class)
}
