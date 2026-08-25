package shell

import "github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"

// weapon_mods.go:艦艇設計畫面用的武器改造(mod)選項與存檔字串 <-> gamedata.WeaponModCode
// 轉換小工具。武器改造本體的手冊出處/佔格/傷害公式全在 gamedata/weapon_mods.go,本檔只是
// 給 shell/UI 層用的薄封裝(選項清單、beam 判斷、切換邏輯),不重複定義任何數字。

// WeaponModOptions 是光束武器在艦艇設計畫面提供的可勾選改造清單。飛彈/魚雷另由
// WeaponModOptionsForWeapon 回傳適用清單，避免 UI 顯示「勾了卻不會生效」的改造。
var WeaponModOptions = []gamedata.WeaponModCode{
	gamedata.ModHeavyMount,
	gamedata.ModPointDefense,
	gamedata.ModAutoFire,
	gamedata.ModContinuousFire,
	gamedata.ModArmorPiercing,
	gamedata.ModEnveloping,
	gamedata.ModNoRangeDissipation,
	gamedata.ModShieldPiercing,
}

// WeaponModOptionsForWeapon 回傳指定武器在目前 remake 解算器中可使用的改造。
// ARM/FST 的原版完整戰術資料流仍有未知欄位，但目前已接上 raw flag、標準飛彈
// Beam Defense／攔截耐久與快速戰鬥的 PD 垂直切片，因此不再把它們藏在造艦畫面之外。
func WeaponModOptionsForWeapon(weaponName string) []gamedata.WeaponModCode {
	if WeaponIsBeam(weaponName) {
		return WeaponModOptions
	}
	if weaponKindByName(weaponName) != WeaponKindMissile {
		return nil
	}
	options := []gamedata.WeaponModCode{
		gamedata.ModArmoredMissile,
		gamedata.ModFastMissile,
		gamedata.ModMissileECCM,
		gamedata.ModEmissionsGuidance,
		gamedata.ModMIRV,
	}
	if WeaponIsTorpedo(weaponName) {
		options = append(options,
			gamedata.ModNoRangeDissipation,
			gamedata.ModEnveloping,
			gamedata.ModOverloadedTorpedo,
		)
	}
	return options
}

// WeaponModOptionsForWeaponAtLevel 只回傳已達原版微型化門檻的改造。
func WeaponModOptionsForWeaponAtLevel(weaponName string, level int) []gamedata.WeaponModCode {
	all := WeaponModOptionsForWeapon(weaponName)
	out := make([]gamedata.WeaponModCode, 0, len(all))
	for _, mod := range all {
		if gamedata.WeaponModUnlockedAtLevel(mod, level) {
			out = append(out, mod)
		}
	}
	return out
}

func weaponMiniaturizationLevel(c Component, completed map[gamedata.ResearchTopic]bool, hyper map[gamedata.ResearchTopic]int) int {
	if c.UnlockTech == gamedata.TECH_NONE {
		return 0
	}
	return gamedata.WeaponMiniaturizationLevelWithHyper(c.UnlockTech, completed, hyper)
}

// WeaponModOptionsForPlayer 是造艦 UI 的科技狀態入口。
func (s *GameSession) WeaponModOptionsForPlayer(weapon int) []gamedata.WeaponModCode {
	c := pick(WeaponOptions, weapon)
	return WeaponModOptionsForWeaponAtLevel(c.Name, weaponMiniaturizationLevel(c, s.Player.CompletedTopics, s.Player.HyperAdvancedLevels))
}

// WeaponModLabelZH 是武器改造代碼的中文顯示名(艦艇設計畫面用)。
var weaponModLabelZH = map[gamedata.WeaponModCode]string{
	gamedata.ModHeavyMount:         "重型平台(HV)",
	gamedata.ModPointDefense:       "點防禦(PD)",
	gamedata.ModAutoFire:           "連續開火(AF)",
	gamedata.ModContinuousFire:     "持續火力(CO)",
	gamedata.ModArmorPiercing:      "穿甲(AP)",
	gamedata.ModEnveloping:         "包覆式(ENV)",
	gamedata.ModNoRangeDissipation: "無射程衰減(NR)",
	gamedata.ModShieldPiercing:     "穿盾(SP)",
	gamedata.ModMissileECCM:        "反反制電子(ECCM)",
	gamedata.ModEmissionsGuidance:  "排放導引(EMG)",
	gamedata.ModMIRV:               "多彈頭(MV)",
	gamedata.ModArmoredMissile:     "重裝飛彈(ARM)",
	gamedata.ModFastMissile:        "快速飛彈(FST)",
	gamedata.ModOverloadedTorpedo:  "魚雷過載(OVR)",
}

// weaponModLabelEN 是同一組改造的**英文顯示名**(手冊 p.115 的 Weapon Mods 附錄用語)。
//
// ⚠ 2026-08-08(第 85 項(元件名英文))補。先前只有中文那一份,而艦艇設計畫面在英文模式下
// 照樣畫中文——那不是漏 `tr()`,是**這一側只有中文資料**,補 `tr()` 補不到。
// 括號裡的縮寫兩種語言相同(手冊本來就用 HV/PD/AF/CO/AP/ENV/NR/SP)。
var weaponModLabelEN = map[gamedata.WeaponModCode]string{
	gamedata.ModHeavyMount:         "Heavy Mount (HV)",
	gamedata.ModPointDefense:       "Point Defense (PD)",
	gamedata.ModAutoFire:           "Auto-Fire (AF)",
	gamedata.ModContinuousFire:     "Continuous Fire (CO)",
	gamedata.ModArmorPiercing:      "Armor Piercing (AP)",
	gamedata.ModEnveloping:         "Enveloping (ENV)",
	gamedata.ModNoRangeDissipation: "No Range Dissipation (NR)",
	gamedata.ModShieldPiercing:     "Shield Piercing (SP)",
	gamedata.ModMissileECCM:        "Electronic Counter-Countermeasures (ECCM)",
	gamedata.ModEmissionsGuidance:  "Emissions Guidance (EMG)",
	gamedata.ModMIRV:               "Multiple Independently Targetable Reentry Vehicle (MV)",
	gamedata.ModArmoredMissile:     "Armored Missile (ARM)",
	gamedata.ModFastMissile:        "Fast Missile (FST)",
	gamedata.ModOverloadedTorpedo:  "Overloaded Torpedo (OVR)",
}

// WeaponModLabelZH 回傳武器改造代碼的中文顯示名;查無回代碼本身(不應發生,防禦性寫法)。
func WeaponModLabelZH(mod gamedata.WeaponModCode) string {
	if s, ok := weaponModLabelZH[mod]; ok {
		return s
	}
	return string(mod)
}

// WeaponModLabelEN 回傳武器改造代碼的英文顯示名;查無回代碼本身。
func WeaponModLabelEN(mod gamedata.WeaponModCode) string {
	if s, ok := weaponModLabelEN[mod]; ok {
		return s
	}
	return string(mod)
}

// WeaponIsBeam 回傳武器元件名是否走 beam 戰鬥解算路徑(weapon_kind.go)。它只負責分類；
// 武器改造的完整適用性由 WeaponModOptionsForWeapon 與 WeaponModCodesForWeapon 判斷。
func WeaponIsBeam(name string) bool {
	return weaponKindByName(name) == WeaponKindBeam
}

// WeaponIsTorpedo 回報武器是否是武器表中的玩家魚雷或怪物專用 raw type 40 魚雷。
func WeaponIsTorpedo(name string) bool {
	switch name {
	case "反物質魚雷", "質子魚雷", "電漿魚雷", "巨龍吐息":
		return true
	default:
		return false
	}
}

// MissileRawWeaponKind 將玩家元件名對到原版 category 21 的 raw kind。
// 四種標準飛彈對應 `Missile_Dcv` / `Fighter_Ocv` 的 0x0E..0x11 分支；質子魚雷的
// 0x12 對應是強推論（它落在原版 0x12/0x13/0x14 的 Dcv=5 分支），魚雷不進目前
// 的 PD 飛彈攔截路徑。
func MissileRawWeaponKind(name string) (int, bool) {
	switch name {
	case "核飛彈":
		return gamedata.MissileKindNuclear, true
	case "麥克萊特飛彈":
		return gamedata.MissileKindMerculite, true
	case "脈衝飛彈":
		return gamedata.MissileKindPulson, true
	case "氙素飛彈":
		return gamedata.MissileKindZeon, true
	case "質子魚雷":
		return gamedata.MissileKindProtonTorpedo, true
	case "反物質魚雷":
		return gamedata.MissileKindAntiMatterTorpedo, true
	case "電漿魚雷":
		return gamedata.MissileKindPlasmaTorpedo, true
	default:
		return 0, false
	}
}

// WeaponModAppliesToWeapon 回報一個已序列化改造是否真的適用於指定武器。
func WeaponModAppliesToWeapon(weaponName string, mod gamedata.WeaponModCode) bool {
	for _, allowed := range WeaponModOptionsForWeapon(weaponName) {
		if allowed == mod {
			return true
		}
	}
	return false
}

// WeaponModCodesForWeapon 將存檔字串轉成計算層代碼，並丟掉切換武器後殘留的舊改造。
// 這個過濾既用於設計成本／佔格，也用於戰鬥，確保 JSON 裡的歷史資料不會繞過 UI
// 直接改變一支不支援該改造的武器。
func WeaponModCodesForWeapon(weaponName string, mods []string) []gamedata.WeaponModCode {
	if len(mods) == 0 {
		return nil
	}
	out := make([]gamedata.WeaponModCode, 0, len(mods))
	for _, mod := range WeaponModCodesFromStrings(mods) {
		if WeaponModAppliesToWeapon(weaponName, mod) {
			out = append(out, mod)
		}
	}
	return out
}

func filterWeaponModsAtLevel(weaponName string, mods []string, level int) []string {
	if len(mods) == 0 {
		return nil
	}
	out := make([]string, 0, len(mods))
	for _, mod := range WeaponModCodesForWeapon(weaponName, mods) {
		if gamedata.WeaponModUnlockedAtLevel(mod, level) {
			out = append(out, string(mod))
		}
	}
	return out
}

// FilterWeaponModsForWeapon 回傳仍適用於指定武器的存檔字串版本，供設計畫面切換武器時
// 清理舊選擇；與 WeaponModCodesForWeapon 共用同一份適用性規則。
func FilterWeaponModsForWeapon(weaponName string, mods []string) []string {
	if len(mods) == 0 {
		return nil
	}
	allowed := WeaponModCodesForWeapon(weaponName, mods)
	if len(allowed) == 0 {
		return nil
	}
	out := make([]string, 0, len(allowed))
	for _, mod := range allowed {
		out = append(out, string(mod))
	}
	return out
}

// ToggleWeaponMod 切換 mods 中是否含 mod,回傳新的切片(不修改原切片)。HV/PD 手冊明訂
// 「mutually exclusive」,勾選其一時會自動移除另一個(不需要玩家自己先取消)。
func ToggleWeaponMod(mods []string, mod gamedata.WeaponModCode) []string {
	code := string(mod)
	out := make([]string, 0, len(mods)+1)
	found := false
	for _, m := range mods {
		if m == code {
			found = true
			continue // 取消勾選:直接跳過,不放進 out
		}
		if mod == gamedata.ModHeavyMount && m == string(gamedata.ModPointDefense) {
			continue // 勾 HV 時移除既有 PD(互斥)
		}
		if mod == gamedata.ModPointDefense && m == string(gamedata.ModHeavyMount) {
			continue // 勾 PD 時移除既有 HV(互斥)
		}
		out = append(out, m)
	}
	if !found {
		out = append(out, code)
	}
	return out
}

// HasWeaponMod 回傳 mods(存檔字串切片)中是否含指定 mod。
func HasWeaponMod(mods []string, mod gamedata.WeaponModCode) bool {
	for _, m := range mods {
		if m == string(mod) {
			return true
		}
	}
	return false
}

// weaponModCodes 把存檔用的 []string 轉成 gamedata 計算函式要的 []gamedata.WeaponModCode。
func weaponModCodes(mods []string) []gamedata.WeaponModCode {
	return WeaponModCodesFromStrings(mods)
}

// WeaponModCodesFromStrings 是 weaponModCodes 的匯出版本,供 cmd/moo2 等外部呼叫端
// (如 fireRound 的 CombatShip.Mods)轉換用,避免各自重複寫一份轉換迴圈。
func WeaponModCodesFromStrings(mods []string) []gamedata.WeaponModCode {
	if len(mods) == 0 {
		return nil
	}
	out := make([]gamedata.WeaponModCode, len(mods))
	for i, m := range mods {
		out[i] = gamedata.WeaponModCode(m)
	}
	return out
}
