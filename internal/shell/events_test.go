package shell

import (
	"strings"
	"testing"
)

// bcCrashFloor300Turns 是 TestRandomEventsFireAndBounded 300 回合內允許的 BC 下限。
// 理由同 antares_test.go 的 bcCrashFloor80Turns:忠實 yield 經濟緩衝薄,300 回合中夾雜安塔蘭
// 入侵反覆把人口打到剩 1 時,單人口收入結構性覆蓋不了建築維護費,會有一段時間 BC 走負——以本
// 測試固定 EventSeed=42 的確定性軌跡實測,300 回合最低點約 -59,且會在人口回穩後回升(實測
// 300 回合終值為正、達數百 BC),不是無止盡崩潰。這裡抓一個有餘裕但仍能抓到「異常擴大化」的
// 下限,詳見 docs/tech/colony-economy-maintenance.md。
//
// 2026-07-11 更新:接上指揮評等(Command Rating)供需結算(GAME_MANUAL.pdf p.169,
// gamedata.IncomeCommandOverflowCost 從先前零呼叫端的死碼變成 RunEmpireTurn 實際會扣款的邏輯)
// 後,下限重新校準。本測試全程被動跑 300 回合(s.Builds 恆為「不建造」,從未新建軌道衛星;
// s.Ships 也維持開局 3 艘不變),母星開局只有 1 座星基(+1 供給)卻有 3 艘艦艇(殖民船+2 偵察艦
// =3 點需求),缺口固定 2 點,每回合被動扣 20 BC(2×10,IncomeCommandOverflowCostPerPoint)。
// 這是忠實但被動放大的手冊機制(玩家若真的玩,會建戰鬥站/星辰要塞或裁減艦隊來補上缺口,本測試
// 刻意不做任何建造/艦隊決策),300 回合線性攤提實測最低點約 -3710(第 273 回合)、終值約
// -3252,重新抓一個有餘裕但仍能抓到「數值算式跑飛」的下限。
const bcCrashFloor300Turns = -4000

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
