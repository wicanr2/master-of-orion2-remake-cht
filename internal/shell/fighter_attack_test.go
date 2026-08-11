package shell

import "testing"

func TestResolveFighterAttackUsesOriginalHitGateAndRange(t *testing.T) {
	miss := ResolveFighterAttack(FighterAttackInput{
		DamageMin: 1, DamageMax: 4, Roll: 39,
	})
	if miss.Hit || miss.Damage != 0 || miss.ModifiedRoll != 39 {
		t.Fatalf("roll 39 should miss: %#v", miss)
	}

	min := ResolveFighterAttack(FighterAttackInput{
		DamageMin: 1, DamageMax: 4, Roll: 40,
	})
	if !min.Hit || min.Damage != 1 || min.ModifiedRoll != 40 {
		t.Fatalf("roll 40 should deal range minimum: %#v", min)
	}

	max := ResolveFighterAttack(FighterAttackInput{
		DamageMin: 1, DamageMax: 4, Roll: 100,
	})
	if !max.Hit || max.Damage != 5 || max.ModifiedRoll != 100 {
		t.Fatalf("roll 100 should preserve original max+1 endpoint: %#v", max)
	}
}

func TestResolveFighterAttackDoesNotInventRawFlagFourBranch(t *testing.T) {
	plain := ResolveFighterAttack(FighterAttackInput{
		Attack: 100, Defense: 100, DamageMin: 4, DamageMax: 16,
		Roll: 40,
	})
	withRaw := ResolveFighterAttack(FighterAttackInput{
		Attack: 100, Defense: 100, DamageMin: 4, DamageMax: 16,
		Roll: 40, RawFlags: 4,
	})
	if withRaw.AttackBonus != plain.AttackBonus || withRaw.Damage != plain.Damage ||
		withRaw.Hit != plain.Hit {
		t.Fatalf("sub_3AD57 的不可達 raw flag 4 分支不應被虛構: plain=%#v raw=%#v", plain, withRaw)
	}
}

func TestResolveFighterBombUsesItsIndependentInterpolation(t *testing.T) {
	first := ResolveFighterBomb(FighterBombInput{
		DamageMin: 2, DamageMax: 7, Roll: 1,
	})
	if first.Damage != 2 {
		t.Fatalf("sub_3AC20 roll 1 應落在 min，得到 %#v", first)
	}

	last := ResolveFighterBomb(FighterBombInput{
		DamageMin: 2, DamageMax: 7, Roll: 100,
	})
	if last.Damage != 7 {
		t.Fatalf("sub_3AC20 roll 100 應落在 max，得到 %#v", last)
	}

	// sub_3AC20 不走 sub_3AD57 的 40 命中門檻；低骰仍會產生直接傷害。
	low := ResolveFighterBomb(FighterBombInput{
		DamageMin: 2, DamageMax: 7, Roll: 39,
	})
	if low.Damage != 4 {
		t.Fatalf("sub_3AC20 roll 39 不應被當成 miss，得到 %#v", low)
	}

	flat := ResolveFighterBomb(FighterBombInput{
		DamageMin: 7, DamageMax: 7, Roll: 50,
	})
	if flat.Damage != 7 {
		t.Fatalf("sub_3AC20 min=max 應直接回 min，得到 %#v", flat)
	}
}
