package shell

// fighter_attack.go：原版戰機對艦命中／傷害下游的純解算。
//
// IDA 證據（Orion2.exe image address）：sub_3AD57 @ 0x3AD57（外部符號對
// Fire_Fighter_Bomb／Fire_Fighter_Beam 有衝突，故保留 raw 名稱）對艦艇路徑
// 先取 raw OCV-like 修正與目標防禦欄，擲 1..100；roll <= 95 時才加入修正，
// 只把上界夾到 100，再以 40 為命中門檻。命中後才把武器表的 min/max
// 送入下游。相鄰的 sub_3AC20 @ 0x3AC20 是另一條直接傷害插值公式，沒有
// 這個 40 命中門檻；兩者不可因外部符號名稱衝突而互換。sub_3A0B9 @ 0x3A0B9
// 是戰機／飛彈 runtime record 的受傷消費端；艦艇則進 sub_39985 @ 0x39985
// 的盾／甲／結構鏈。

// FighterAttackInput 是一次戰機對艦射擊的原始可觀測輸入。
type FighterAttackInput struct {
	Attack    int
	Defense   int
	DamageMin int
	DamageMax int
	Roll      int // 原版 random(100) 的 1..100 外部化擲骰
	// RawFlags 保留呼叫端未來傳入的原始欄位；目前 sub_3AD57 的
	// var_18 只由 word_17F815 的 bit 8 轉成 0 或 0x40，隨後的
	// test var_18,4 永遠不成立，因此本解算器不把 RawFlags 當成
	// 已證實的傷害改造旗標。
	RawFlags uint16
}

// FighterAttackResult 是命中／傷害下游的結果；盾與裝甲仍由呼叫端的
// CombatShip 狀態消費，避免把格線戰鬥的分面狀態偷偷複製一份。
type FighterAttackResult struct {
	Hit          bool
	Damage       int
	ModifiedRoll int
	AttackBonus  int
	HitThreshold int
}

// FighterBombInput 是 raw sub_3AC20 @ 0x3AC20 的純公式輸入。
//
// 原版的 R 由 sub_1247A0(100) 產生 1..100；呼叫端在這裡外化擲骰，讓
// 公式可以用固定輸入抽樣驗證。這條函式不接受 Attack／Defense，因為 IDA
// 的 sub_3AC20 路徑沒有 sub_3AD57 的 OCV／40 命中門檻。
type FighterBombInput struct {
	DamageMin int
	DamageMax int
	Roll      int // 原版 sub_1247A0(100) 的 1..100 外部化擲骰
}

// FighterBombResult 是 sub_3AC20 的直接傷害輸出。盾、裝甲與結構的消費仍由
// 呼叫端處理，避免在純公式層複製 CombatShip 狀態。
type FighterBombResult struct {
	Damage int
}

// ResolveFighterAttack 轉寫原版戰機對艦的命中門檻與傷害插值。
//
// roll <= 95 才套用攻防差；95 以上的原版特殊尾端擲骰保留原值。命中門檻
// 是 40。原始指令使用 `(max-min+1)`，因此在 modified=100 時可能得到
// max+1；這個看似反直覺的端點行為要保留，不能自行封頂。
func ResolveFighterAttack(in FighterAttackInput) FighterAttackResult {
	result := FighterAttackResult{HitThreshold: 40}
	if in.DamageMin < 0 {
		in.DamageMin = 0
	}
	if in.DamageMax < in.DamageMin {
		in.DamageMax = in.DamageMin
	}
	if in.Roll <= 0 {
		return result
	}

	minDamage, maxDamage := in.DamageMin, in.DamageMax
	attackBonus := in.Attack - in.Defense
	result.AttackBonus = attackBonus

	modified := in.Roll
	if modified <= 95 {
		modified += attackBonus
	}
	// sub_3AD57 沒有下界夾限；負值只會在命中門檻判斷失敗。
	if modified > 100 {
		modified = 100
	}
	result.ModifiedRoll = modified
	if modified < result.HitThreshold {
		return result
	}

	damage := minDamage
	if maxDamage > minDamage {
		damage += (modified - result.HitThreshold) * (maxDamage - minDamage + 1) /
			(100 - result.HitThreshold)
	}
	result.Hit = true
	result.Damage = damage
	return result
}

// ResolveFighterBomb 轉寫 sub_3AC20 @ 0x3AC20 的直接傷害插值。
//
// 對合法原版輸入，令 S=max-min：
//
//	S > 0: D = min + floor((floor(100/(2*S)) + R) * S / 100)
//	S <= 0: D = min
//
// 原始指令沒有命中門檻，也沒有在這條公式末端自行封頂。Roll <= 0 只
// 是重製 API 的失敗即關閉防護；原版 sub_1247A0(100) 不會產生這種值。
func ResolveFighterBomb(in FighterBombInput) FighterBombResult {
	result := FighterBombResult{}
	if in.Roll <= 0 {
		return result
	}

	span := in.DamageMax - in.DamageMin
	if span <= 0 {
		result.Damage = in.DamageMin
		return result
	}

	result.Damage = in.DamageMin +
		((100/(2*span)+in.Roll)*span)/100
	return result
}
