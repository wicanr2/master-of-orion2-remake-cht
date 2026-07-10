package gamedata

import (
	"math/rand"
	"testing"
)

// winRate 跑 n 場獨立(不同 seed)戰鬥,回傳攻方勝率(0..1)。makeForces 每次呼叫都要回傳
// 全新的 GroundForce(ResolveGroundBattle 雖不會就地修改輸入,但每場戰鬥的 rng 不同,獨立
// 建新 force 較不容易寫出誤用共享狀態的測試)。
func winRate(t *testing.T, n int, baseSeed int64, makeForces func() (GroundForce, GroundForce)) float64 {
	t.Helper()
	wins := 0
	for i := 0; i < n; i++ {
		rng := rand.New(rand.NewSource(baseSeed + int64(i)))
		atk, def := makeForces()
		res := ResolveGroundBattle(atk, def, rng)
		if res.AttackerWon {
			wins++
		}
	}
	return float64(wins) / float64(n)
}

// TestResolveGroundBattle_ForceAdvantageWinsMore 驗證解算式定案的核心性質:force 加成
// 明顯較高的一方,在固定兵力下應有明顯偏高的勝率。攻方 force=30、守方 force=0,兩邊各 5
// 單位、hits-to-kill=1(單擊必亡,排除 hits-to-kill 干擾,單純驗證 force 效果)。
func TestResolveGroundBattle_ForceAdvantageWinsMore(t *testing.T) {
	const n = 200
	rate := winRate(t, n, 1000, func() (GroundForce, GroundForce) {
		atk := NewGroundForce(5, 1, 30, false)
		def := NewGroundForce(5, 1, 0, true)
		return atk, def
	})
	if rate <= 0.75 {
		t.Fatalf("force 明顯較高一方(+30 vs +0)200 場勝率 = %.2f,預期 > 0.75", rate)
	}
	t.Logf("force+30 vs force+0(200 場)攻方勝率 = %.2f", rate)
}

// TestResolveGroundBattle_EqualForceBaseline 驗證雙方 force 相等、兵力相等時,理論上
// 應接近 50%(對稱局面下不應有系統性偏袒攻方或守方——除了「平手歸守方」造成的些微不對稱,
// 這在單回合機率上可忽略)。此測試同時作為下一個「雙倍兵力」測試的基準對照。
func TestResolveGroundBattle_EqualForceBaseline(t *testing.T) {
	const n = 200
	rate := winRate(t, n, 2000, func() (GroundForce, GroundForce) {
		atk := NewGroundForce(5, 1, 0, false)
		def := NewGroundForce(5, 1, 0, true)
		return atk, def
	})
	if rate < 0.30 || rate > 0.70 {
		t.Fatalf("force 相等、兵力相等(5v5)200 場攻方勝率 = %.2f,預期落在 [0.30, 0.70] 附近(對稱局面)", rate)
	}
	t.Logf("force 相等 5v5(200 場)攻方勝率 = %.2f(對稱基準)", rate)
}

// TestResolveGroundBattle_DoubleUnitsWinsMoreAtEqualForce 驗證社群經驗法則:force 相等
// 時,兵力雙倍的一方勝率應明顯高於兵力相等時的基準(~50%)。攻方 10 單位 vs 守方 5 單位,
// 兩邊 force 皆為 0。
func TestResolveGroundBattle_DoubleUnitsWinsMoreAtEqualForce(t *testing.T) {
	const n = 200
	rate := winRate(t, n, 3000, func() (GroundForce, GroundForce) {
		atk := NewGroundForce(10, 1, 0, false)
		def := NewGroundForce(5, 1, 0, true)
		return atk, def
	})
	if rate <= 0.75 {
		t.Fatalf("force 相等、攻方兵力雙倍(10 vs 5)200 場攻方勝率 = %.2f,預期 > 0.75(社群經驗:雙倍兵力應明顯佔優)", rate)
	}
	t.Logf("force 相等、兵力雙倍 10v5(200 場)攻方勝率 = %.2f", rate)
}

// TestResolveGroundBattle_ZeroDefenderUnitsAttackerWinsImmediately 邊界:守方一開始就
//没有存活單位,攻方應直接獲勝,且不應消耗 rng(Rounds == 0)。
func TestResolveGroundBattle_ZeroDefenderUnitsAttackerWinsImmediately(t *testing.T) {
	atk := NewGroundForce(3, 1, 0, false)
	def := GroundForce{Units: nil, Force: 999, Defending: true} // force 再高,0 單位也無法防守
	rng := rand.New(rand.NewSource(42))

	res := ResolveGroundBattle(atk, def, rng)
	if !res.AttackerWon {
		t.Fatalf("守方 0 單位時攻方應直接獲勝,got AttackerWon=false")
	}
	if res.AttackerSurvived != 3 {
		t.Fatalf("攻方存活數應維持 3(無需交戰),got %d", res.AttackerSurvived)
	}
	if res.DefenderSurvived != 0 {
		t.Fatalf("守方存活數應為 0,got %d", res.DefenderSurvived)
	}
	if res.Rounds != 0 {
		t.Fatalf("守方 0 單位應不擲骰,Rounds 應為 0,got %d", res.Rounds)
	}
}

// TestResolveGroundBattle_ZeroAttackerUnitsDefenderWins 邊界:攻方一開始就沒有存活單位,
// 守方應直接獲勝。
func TestResolveGroundBattle_ZeroAttackerUnitsDefenderWins(t *testing.T) {
	atk := GroundForce{Units: nil, Force: 999, Defending: false}
	def := NewGroundForce(3, 1, 0, true)
	rng := rand.New(rand.NewSource(7))

	res := ResolveGroundBattle(atk, def, rng)
	if res.AttackerWon {
		t.Fatalf("攻方 0 單位時守方應獲勝,got AttackerWon=true")
	}
	if res.DefenderSurvived != 3 {
		t.Fatalf("守方存活數應維持 3,got %d", res.DefenderSurvived)
	}
	if res.Rounds != 0 {
		t.Fatalf("攻方 0 單位應不擲骰,Rounds 應為 0,got %d", res.Rounds)
	}
}

// TestResolveGroundBattle_HitsToKillRequiresMultipleHits 驗證 hits-to-kill 疊加規則:
// 1 對 1、force 相等,雙方單位 hits-to-kill=3 時,落敗方那唯一的單位需要「恰好 3 次」命中
// 才會陣亡——因此總回合數必落在 [3, 5](3 = 落敗方須承受的 hit 數下限;5 = 落敗方 3 次 +
// 獲勝方在陣亡前最多可承受的 2 次,hits-to-kill-1)。用大量不同 seed 統計驗證此不變量,
// 同時反向確認「並非首次命中就陣亡」(若實作忘記疊加 hits,Rounds 會恆為 1)。
func TestResolveGroundBattle_HitsToKillRequiresMultipleHits(t *testing.T) {
	const n = 300
	const hitsToKill = 3
	for i := 0; i < n; i++ {
		rng := rand.New(rand.NewSource(int64(50000 + i)))
		atk := NewGroundForce(1, hitsToKill, 0, false)
		def := NewGroundForce(1, hitsToKill, 0, true)

		res := ResolveGroundBattle(atk, def, rng)

		if res.Rounds < hitsToKill {
			t.Fatalf("seed=%d: 1v1 hits-to-kill=%d 理應至少 %d 回合才會有一方陣亡,got Rounds=%d",
				50000+i, hitsToKill, hitsToKill, res.Rounds)
		}
		maxRounds := hitsToKill + (hitsToKill - 1)
		if res.Rounds > maxRounds {
			t.Fatalf("seed=%d: 1v1 hits-to-kill=%d 理應最多 %d 回合分出勝負,got Rounds=%d",
				50000+i, hitsToKill, maxRounds, res.Rounds)
		}
		// 勝方必為存活 1、負方必為存活 0(1v1 情境下不存在雙方皆存活或皆陣亡的中間狀態)。
		if res.AttackerWon {
			if res.AttackerSurvived != 1 || res.DefenderSurvived != 0 {
				t.Fatalf("seed=%d: 攻方勝但存活數異常 atk=%d def=%d", 50000+i, res.AttackerSurvived, res.DefenderSurvived)
			}
		} else {
			if res.DefenderSurvived != 1 || res.AttackerSurvived != 0 {
				t.Fatalf("seed=%d: 守方勝但存活數異常 atk=%d def=%d", 50000+i, res.AttackerSurvived, res.DefenderSurvived)
			}
		}
	}
}

// TestResolveGroundBattle_TerminatesAndRoundsBounded 驗證迴圈必然終止:因為每回合必有一方
// 至少扣 1 hit,總回合數不可能超過雙方「總 hit 容量」之和(Σ每單位 hits-to-kill)。用較大的
// 兵力與 hits-to-kill 組合(10 單位、hits-to-kill=3,兩邊共 60 hit 容量)驗證此上界,間接
// 證明不存在死迴圈(若有死迴圈,函式呼叫本身就不會返回,測試會逾時失敗——本測試把「有限步驟
// 內返回」與「返回值合理」都納入斷言)。
func TestResolveGroundBattle_TerminatesAndRoundsBounded(t *testing.T) {
	const n = 100
	const units = 10
	const hitsToKill = 3
	maxPossibleRounds := units*hitsToKill + units*hitsToKill // 雙方總 hit 容量上限(寬鬆上界)

	for i := 0; i < n; i++ {
		rng := rand.New(rand.NewSource(int64(90000 + i)))
		atk := NewGroundForce(units, hitsToKill, 5, false)
		def := NewGroundForce(units, hitsToKill, 5, true)

		res := ResolveGroundBattle(atk, def, rng)

		if res.Rounds <= 0 {
			t.Fatalf("seed=%d: 兩邊皆有兵力時 Rounds 應 > 0,got %d", 90000+i, res.Rounds)
		}
		if res.Rounds > maxPossibleRounds {
			t.Fatalf("seed=%d: Rounds=%d 超過理論上界 %d,實作可能有 bug(未正確扣減 hits)", 90000+i, res.Rounds, maxPossibleRounds)
		}
		// 恰一方全滅、另一方至少存活 1(force 相同、兵力相同,勝方不可能兩邊都全滅)。
		if res.AttackerWon == (res.DefenderSurvived != 0) {
			t.Fatalf("seed=%d: AttackerWon=%v 與 DefenderSurvived=%d 不一致", 90000+i, res.AttackerWon, res.DefenderSurvived)
		}
		if (!res.AttackerWon) == (res.AttackerSurvived != 0) {
			t.Fatalf("seed=%d: AttackerWon=%v 與 AttackerSurvived=%d 不一致", 90000+i, res.AttackerWon, res.AttackerSurvived)
		}
	}
}
