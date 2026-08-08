package gamedata

// weapon_damage.go:**光束武器的傷害表**(GAME_MANUAL.pdf p.124)。
//
// ============ 「需要 OCR / 找完整手冊」那句話是錯的 ============
//
// `docs/tech/component-values.md` 記著武器數值的來源等級:
//
//	其餘武器 Value 為依科技階遞增的**單調估計**(雷射 4→死光 25),保持排序合理與遊戲可玩,
//	但未經手冊逐條核對。
//
// 而它的待辦第一條寫著:
//
//	- [ ] OCR 掃描版手冊武器/裝甲/護盾附錄(若附錄存在於該 PDF;**9 頁本可能不含,需找完整手冊**)
//
// **那份完整手冊一直都在。** `moo2_patch1.5/GAME_MANUAL.pdf` 是可直接抽文字的(44 萬字元),
// 而且 `shipspace.go` 的 `WeaponSpaceByName` 就是從**同一頁 p.124** 抽出來的 Size 欄
// ——只抽了 Size,旁邊的 Damage 欄沒抽。
//
// ============ 估計值錯得有多遠 ============
//
//	武器            remake 估計   手冊
//	雷射             4           1-4     ✓
//	質量投射器        8           6(固定) ✗ 高估
//	中子爆破槍        12          3-12    ✓
//	核融合光束        16          2-6     ✗✗ 高估近 3 倍
//	高斯砲           18          18(固定) ✓
//	相位砲           19          5-20    ✗ 低估
//	死光             25          50-100  ✗✗✗ 低估 4 倍
//
// 最嚴重的不是數字偏差,是**排序被弄反了**:核融合光束在 remake 比中子爆破槍強(16 > 12),
// 手冊上它比中子爆破槍**弱**(最大 6 vs 12)。單調遞增的估計法必然出這種錯——
// 它假設「科技越後面越強」,而手冊的武器線本來就不是單調的。
//
// ============ 誠實留白 ============
//
//   - **電漿砲版本相依**:1.31 是 6-30、1.50 是 4-20(MANUAL_150.html 明載)。
//     這裡放 1.50 的值,1.31 由 `RuleProfile.PlasmaCannonMaxDamage` 覆寫(既有機制)。
//   - **只收 remake 元件表裡有的那幾項。** 手冊還列了離子脈衝砲、引力波束、干擾者、
//     重錘裝置、粒子束——remake 的 `WeaponOptions` 沒有它們,不在這一輪擴充元件表。
//   - **飛彈/炸彈/特殊武器另有表**(p.125-127),不在本檔:remake 目前把飛彈當一般
//     元件塞在同一份 `WeaponOptions` 裡,要分開得先改元件模型。

// WeaponDamageRange 是一項武器的傷害範圍(固定傷害時 Min == Max)。
type WeaponDamageRange struct{ Min, Max int }

// beamWeaponDamage 是手冊 p.124 的 BEAM 表,鍵用 shell 的元件中文名。
var beamWeaponDamage = map[string]WeaponDamageRange{
	"雷射":    {1, 4},    // Laser Cannon 1-4
	"質量投射器": {6, 6},    // Mass Driver 6(固定值,手冊沒有範圍)
	"中子爆破槍": {3, 12},   // Neutron Blaster 3-12
	"核融合光束": {2, 6},    // Fusion Beam 2-6
	"高斯砲":   {18, 18},  // Gauss Cannon 18(固定值)
	"相位砲":   {5, 20},   // Phasor 5-20
	"電漿砲":   {4, 20},   // Plasma Cannon 4-20(1.50;1.31 為 6-30,由 RuleProfile 覆寫)
	"死光":    {50, 100}, // Death Ray 50-100

	// 第 64 項(武器傷害真表)補上的光束項(手冊 p.124 同一張表)。
	"離子脈衝砲": {2, 10},    // Ion Pulse Cannon 2-10
	"引力波束":  {3, 15},    // Graviton Beam 3-15
	"干擾者":   {40, 40},   // Disrupter 40(固定值)
	"粒子束":   {10, 30},   // Particle Beam 10-30
	"重錘裝置":  {100, 100}, // Mauler Device 100(固定值;手冊 specials 欄「always hits」)
}

// bombWeaponDamage 是手冊 p.126 的 BOMB 表。**炸彈只能打行星**(見 shell 的 WeaponKindBomb)。
//
// ⚠ 手冊同表還有死亡孢子(10%)與生物滅絕者(20%),它們給的是**殺人口的百分比**不是傷害,
// 而且第 52 項(生物武器分類)已經用另一條路徑(科技擁有 → 轟炸時擲骰殺人口)接好了。**不要**把它們
// 也加進元件表——那會讓同一條規則生效兩次。
var bombWeaponDamage = map[string]WeaponDamageRange{
	"核彈":   {3, 12},  // Nuclear Bomb 3-12
	"融合彈":  {4, 24},  // Fusion Bomb 4-24
	"反物質彈": {5, 40},  // Anti-Matter Bomb 5-40
	"中子彈":  {10, 60}, // Neutronium Bomb 10-60
}

// sphericalWeaponDamage 是手冊 p.126 明列的球形武器裡、remake 掛得上的那兩項
// (p.127 的數值表)。
//
// ⚠ 這兩項的傷害手冊寫的是「**per size class of target**」——表裡放的是**單一級數**的
// 基準值,乘上目標艦體級數是呼叫端的事(見 shell 的 battleVolley 球形分支)。
// 手冊球形清單裡的另外兩項不在這裡:電漿通量是海鰻怪獸專屬、引擎爆炸不是可裝載武器。
var sphericalWeaponDamage = map[string]WeaponDamageRange{
	"脈衝星":   {1, 24}, // Pulsar 1-24 per size class
	"空間壓縮器": {4, 32}, // Spatial Compressor 4-32 structural hits
	// 陀螺去穩器(第 70 項(陀螺去穩器)):手冊「causes **1–4 points of structural damage multiplied by
	// the size class of the ship**. **Shields and armor offer no protection** and are not
	// damaged.」——依級數乘 + 豁免盾甲,兩個特徵都與球形家族相同,所以走同一條路。
	"陀螺去穩器": {1, 4},
}

// missileWeaponDamage 是手冊 p.125 的 MISSILE 表(固定傷害,飛彈「never miss」但會被攔截/干擾)。
var missileWeaponDamage = map[string]WeaponDamageRange{
	"核飛彈":    {8, 8},   // Nuclear Missile 8
	"麥克萊特飛彈": {14, 14}, // Merculite Missile 14
	"脈衝飛彈":   {20, 20}, // Pulson Missile 20
	"氙素飛彈":   {30, 30}, // Zeon Missile 30
	"質子魚雷":   {25, 25}, // Proton/A-M Torpedo 25
}

// WeaponDamageByName 回傳某武器元件的手冊傷害範圍;不在表上回 ok=false
// (呼叫端據此保留既有值,不要靜默當成 0 傷害)。
func WeaponDamageByName(name string) (WeaponDamageRange, bool) {
	if d, ok := beamWeaponDamage[name]; ok {
		return d, true
	}
	if d, ok := missileWeaponDamage[name]; ok {
		return d, true
	}
	if d, ok := bombWeaponDamage[name]; ok {
		return d, true
	}
	d, ok := sphericalWeaponDamage[name]
	return d, ok
}
