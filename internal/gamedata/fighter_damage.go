package gamedata

// fighter_damage.go：原版 Fighter Bays 表中的「艦載機自身傷害」欄。
//
// 證據：Orion2.exe 的武器表位於 0x17F807、每筆 0x1C bytes；ID 31
// (Fighter Bays) 的第二組傷害不是母艦武器的 6..15，而是戰機型別的
// 1..4 / 4..16 / 2..7。手冊 p.127 的 Interceptor / Heavy Fighter /
// Bomber 三列與該組數值交叉吻合。這裡只保存已證實的範圍，不把未知的
// 原版艦載機槽位或敵方藍圖推成玩家科技。

// FighterDamageRange 是一架戰機一次對艦射擊的原版傷害範圍。
// Heavy Fighter 的 4..16 是其「光束 + 炸彈」下游總傷害範圍；它不是
// 母艦粒子束的 10..30，也不應再套用母艦武器的傷害表。
type FighterDamageRange struct {
	Min int
	Max int
}

// FighterDamageRangeForKind 依 shell.FighterKind 的穩定 ordinal 回傳原版表值。
//
// 為避免 gamedata 依賴 shell，參數刻意使用 int：
//
//	0 = Interceptor, 1 = Heavy Fighter, 2 = Bomber, 3 = Assault Shuttle。
//
// 突擊艇不開火，故回傳零值。未知 ordinal 也回傳零值，讓呼叫端 fail-closed。
func FighterDamageRangeForKind(kind int) (FighterDamageRange, bool) {
	switch kind {
	case 0: // Interceptor
		return FighterDamageRange{Min: 1, Max: 4}, true
	case 1: // Heavy Fighter (beam + bomb)
		return FighterDamageRange{Min: 4, Max: 16}, true
	case 2: // Bomber (one bomb)
		return FighterDamageRange{Min: 2, Max: 7}, true
	default:
		return FighterDamageRange{}, false
	}
}
