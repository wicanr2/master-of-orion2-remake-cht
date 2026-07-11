package shell

import "github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"

// weapon_mods.go:艦艇設計畫面用的武器改造(mod)選項與存檔字串 <-> gamedata.WeaponModCode
// 轉換小工具。武器改造本體的手冊出處/佔格/傷害公式全在 gamedata/weapon_mods.go,本檔只是
// 給 shell/UI 層用的薄封裝(選項清單、beam 判斷、切換邏輯),不重複定義任何數字。

// WeaponModOptions 是艦艇設計畫面提供的可勾選武器改造清單。只收手冊裡對 beam 武器有精確
// 數字、且本 remake 戰鬥解算(ResolveShotWithMods)已接線的 8 個 mod;飛彈專屬 mod
// (ARM/ECCM/EMG/FST/MV,以及 NR 的魚雷版)手冊雖也有精確數字,但 remake 的飛彈解算
// (ResolveMissileShot)尚無 mod 掛鉤機制,故不放進本清單,避免 UI 讓玩家選了卻沒效果的
// mod(見 docs/tech/weapon-mods.md 的 TODO)。
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
}

// WeaponModLabelZH 回傳武器改造代碼的中文顯示名;查無回代碼本身(不應發生,防禦性寫法)。
func WeaponModLabelZH(mod gamedata.WeaponModCode) string {
	if s, ok := weaponModLabelZH[mod]; ok {
		return s
	}
	return string(mod)
}

// WeaponIsBeam 回傳武器元件名是否走 beam 戰鬥解算路徑(weapon_kind.go)。武器改造系統
// 目前只對 beam 生效(手冊 HV/PD/AF/CO 明文只講 beam 武器;AP/ENV/NR/SP 手冊雖也適用
// 魚雷,但本 remake 的飛彈路徑未接 mod 掛鉤,見 WeaponModOptions 註解),UI 與計算函式都靠
// 這個判斷決定要不要套用 mods。
func WeaponIsBeam(name string) bool {
	return weaponKindByName(name) == WeaponKindBeam
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
