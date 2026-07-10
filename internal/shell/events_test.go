package shell

import (
	"strings"
	"testing"
)

// bcCrashFloor300Turns 是 TestRandomEventsFireAndBounded 300 回合內允許的 BC 下限。
// 理由同 antares_test.go 的 bcCrashFloor80Turns:忠實 yield 經濟緩衝薄,300 回合中夾雜安塔蘭
// 入侵反覆把人口打到剩 1 時,單人口收入結構性覆蓋不了建築維護費,會有一段時間 BC 走負——這是
// 誠實的經濟後果,不是 bug,詳見 docs/tech/colony-economy-maintenance.md。實際下限數字隨底下
// 指揮評等供需結算的演進變動,見下方更新記錄。
//
// 2026-07-11 更新①:接上指揮評等(Command Rating)供需結算(GAME_MANUAL.pdf p.169,
// gamedata.IncomeCommandOverflowCost 從先前零呼叫端的死碼變成 RunEmpireTurn 實際會扣款的邏輯)
// 後,一度發現開局供給只算了軌道衛星、漏了帝國基礎值,母星開局只有 1 座星基(+1 供給)卻有
// 3 艘開局艦艇(殖民船+2 偵察艦=3 點需求),缺口固定 2 點,每回合被動扣 20 BC,300 回合線性
// 攤提到最低點約 -3710(這是 regression,不是忠實機制——見下一段)。
//
// 2026-07-11 更新②(regression 修復):用真實存檔 SAVE10.GAM oracle 反推確認帝國基礎指揮
// 評等供給應為 5(gamedata.CommandPointsBase,見該常數註解完整推導),shell.
// totalCommandPointsSupply 補上後,開局供給 = 5(基礎)+1(星基)= 6 ≥ 3(需求),不再超支。
// 重新量測 300 回合(EventSeed=42,固定可重現軌跡):最低點約 -51(第 133 回合,安塔蘭入侵/
// 隨機事件把人口打低時單人口收入結構性不足以覆蓋建築維護費的短暫波動,非 bug,同
// bcCrashFloor80Turns 的性質),終值約 +718。下限抓一個有餘裕(約 8 倍)但仍能抓到「數值算式
// 跑飛」的門檻,不再用先前為掩蓋死亡螺旋而放寬的 -4000。
const bcCrashFloor300Turns = -400

// TestRandomEventsFireAndBounded 驗證隨機事件會在多回合中觸發,殖民地人口不低於 1,且 BC
// 不會失控式無下限崩潰(忠實經濟下人口被打到剩 1 的期間會短暫轉負,詳見 bcCrashFloor300Turns
// 註解,這是誠實的經濟後果,不是 bug)。
func TestRandomEventsFireAndBounded(t *testing.T) {
	s := NewDemoSession()
	fired := map[string]int{}
	for i := 0; i < 300; i++ {
		s.EndTurn()
		if s.LastEvent != "" {
			// 以事件類別(冒號前)計數。
			key := s.LastEvent
			if idx := strings.IndexRune(key, ':'); idx > 0 {
				key = key[:idx]
			}
			fired[key]++
		}
		if s.Player.BC < bcCrashFloor300Turns {
			t.Fatalf("第 %d 回合 BC 崩潰超出合理下限:%d(< %d,事件 %q)", i, s.Player.BC, bcCrashFloor300Turns, s.LastEvent)
		}
		for j, c := range s.PlayerColonies {
			if c.Population < 1 {
				t.Fatalf("第 %d 回合殖民地 %d 人口 <1:%d", i, j, c.Population)
			}
		}
	}
	if len(fired) == 0 {
		t.Fatal("300 回合內應至少觸發一次隨機事件")
	}
	t.Logf("觸發事件類別:%v(結束 BC=%d)", fired, s.Player.BC)
}

// TestRandomEventsReproducible 驗證相同 EventSeed 產生相同事件序列(可重現)。
func TestRandomEventsReproducible(t *testing.T) {
	seq := func() []string {
		s := NewDemoSession()
		var out []string
		for i := 0; i < 50; i++ {
			s.EndTurn()
			out = append(out, s.LastEvent)
		}
		return out
	}
	a, b := seq(), seq()
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("第 %d 回合事件不可重現:%q vs %q", i, a[i], b[i])
		}
	}
}
