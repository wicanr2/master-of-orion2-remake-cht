package shell

import (
	"strings"
	"testing"
)

// TestRandomEventsFireAndBounded 驗證隨機事件會在多回合中觸發,且效果有界:
// BC 不為負、殖民地人口不低於 1。
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
		if s.Player.BC < 0 {
			t.Fatalf("第 %d 回合 BC 為負:%d(事件 %q)", i, s.Player.BC, s.LastEvent)
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
	t.Logf("觸發事件類別:%v", fired)
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
