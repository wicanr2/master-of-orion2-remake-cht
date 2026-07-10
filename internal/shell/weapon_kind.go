package shell

// weapon_kind.go:武器戰鬥解算路徑分類(beam / missile / spherical)。
//
// 分類依據優先順序:①WeaponOptions(session.go)裡的武器名是否對應手冊明確的飛彈/球形
// 武器清單;②查無對應則預設 beam(現行絕大多數武器,含「無武裝」——空武裝仍沿用 beam
// 分支,反正 wmin=wmax=0 不影響結果,只是不要漏判成其他分支)。
//
// 已核對出處(手冊 moo2_patch1.5/MANUAL_150.html):
//   - missile:「核飛彈」「麥克萊特飛彈」對應手冊 missile.go 已移植的
//     MissileWarheadNuclear / MissileWarheadMerculite(Beam Defense of Missiles 表,
//     p117-120),兩者是 WeaponOptions 目前僅有的兩個飛彈武器。
//   - spherical:手冊「Notes on Spherical Damage > Spherical Weapons」(p126)明列的球形
//     武器是 Pulsar、Plasma Flux(Eel 專屬)、Spatial Compressor、Engine Explosion——
//     WeaponOptions 目前完全沒有對應武器(死光/Death Ray 不在此列!死光是一般光束武器,
//     且是 damage.go DamageForHit「Different Min-Max Damage」worked example 的出處,
//     混進 spherical 會與已核對的手冊數字矛盾,故明確排除)。因此 spherical 分支目前
//     沒有任何武器掛載,ResolveSphericalShot 只提供已測試的解算函式待未來新增球形武器
//     元件(如 Pulsar/恆星轉換器等真科技樹尚未映射的武器)時串接。
//
// 新增武器到 WeaponOptions 時,記得同步檢查是否需要在這裡新增分類(預設落到 beam 不會
// 編譯錯誤,但戰鬥行為會不忠實)。
type WeaponKind int

const (
	WeaponKindBeam WeaponKind = iota
	WeaponKindMissile
	WeaponKindSpherical
)

// weaponKindByName 依 Component.Name(WeaponOptions 的武器名)分類戰鬥解算路徑。
func weaponKindByName(name string) WeaponKind {
	switch name {
	case "核飛彈", "麥克萊特飛彈":
		return WeaponKindMissile
	default:
		return WeaponKindBeam
	}
}
