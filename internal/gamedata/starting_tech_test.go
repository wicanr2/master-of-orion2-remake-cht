package gamedata

import "testing"

// starting_tech_test.go:開局研究主題表的護欄。
//
// 這張表先前是「拿不到所以不臆造」的狀態,現在是反組譯真值——測試要釘住的是
// **那六個編號與那三個數量**,不是「有幾個主題」這種模糊的東西。

// word_18111C 的六個 word:1Dh, 37h, 16h, 39h, 1Ch, 17h。
func TestStartingTopicOrderMatchesTheOriginalTable(t *testing.T) {
	want := []struct {
		raw   int
		topic ResearchTopic
	}{
		{0x1D, TOPIC_ENGINEERING},
		{0x37, TOPIC_NUCLEAR_FISSION},
		{0x16, TOPIC_CHEMISTRY},
		{0x39, TOPIC_PHYSICS},
		{0x1C, TOPIC_ELECTRONICS},
		{0x17, TOPIC_COLD_FUSION},
	}
	if len(StartingTopicOrder) != len(want) {
		t.Fatalf("表長應為 %d,實得 %d", len(want), len(StartingTopicOrder))
	}
	for i, w := range want {
		if int(StartingTopicOrder[i]) != w.raw {
			t.Errorf("第 %d 個應為 0x%X(%d),實得 %d", i, w.raw, w.raw, StartingTopicOrder[i])
		}
		if StartingTopicOrder[i] != w.topic {
			t.Errorf("第 %d 個主題對不上", i)
		}
	}
	// 手冊獨立說「預設的第一個是 field #29」——與反組譯的 word_18111C[0] 互證。
	if StartingTopicOrder[0] != 29 {
		t.Errorf("手冊說第一個是 #29,實得 %d", StartingTopicOrder[0])
	}
	// 第二個是 FTL 所在的主題:手冊說 Average「已具備星際航行所需的全部科技」。
	if StartingTopicOrder[1] != TOPIC_NUCLEAR_FISSION {
		t.Error("第二個應是核分裂(FTL 所在的主題),否則一般開局會沒有 FTL")
	}
}

// var_18 = 1 / 6 / 25。
func TestStartingTopicCountMatchesTheOriginalImmediates(t *testing.T) {
	for _, c := range []struct{ level, want int }{
		{0, 1}, {1, 6}, {2, 25},
	} {
		if got := StartingTopicCount(c.level); got != c.want {
			t.Errorf("等級 %d 應送 %d 個,實得 %d", c.level, c.want, got)
		}
	}
	// 超出範圍退回「一般」——原版對第四級是讀到未初始化的堆疊值,那個不照抄。
	for _, lv := range []int{-1, 3, 99} {
		if got := StartingTopicCount(lv); got != 6 {
			t.Errorf("超出範圍的等級 %d 應退回 6,實得 %d", lv, got)
		}
	}
}

// 送出的清單是固定表的**前 N 個**,不是隨機挑。
func TestStartingTopicsTakesThePrefix(t *testing.T) {
	if got := StartingTopics(0); len(got) != 1 || got[0] != TOPIC_ENGINEERING {
		t.Errorf("曲速前應只送 ENGINEERING,實得 %v", got)
	}
	got := StartingTopics(1)
	if len(got) != 6 {
		t.Fatalf("一般應送 6 個,實得 %d", len(got))
	}
	for i := range got {
		if got[i] != StartingTopicOrder[i] {
			t.Errorf("第 %d 個與固定表不同", i)
		}
	}
	// 25 級不會回 25 個:超過固定表的部分原版是隨機挑,remake 還沒接(見下一條)。
	if n := len(StartingTopics(2)); n != 6 {
		t.Errorf("先進級目前只回固定表的 6 個,實得 %d", n)
	}
}

// 缺口要是一個看得見的數字,不是註解裡的一句話。
func TestStartingTopicRandomExtrasReportsTheGap(t *testing.T) {
	if got := StartingTopicRandomExtras(0); got != 0 {
		t.Errorf("曲速前不該有隨機額外主題,實得 %d", got)
	}
	if got := StartingTopicRandomExtras(1); got != 0 {
		t.Errorf("一般不該有隨機額外主題,實得 %d", got)
	}
	if got := StartingTopicRandomExtras(2); got != 19 {
		t.Errorf("先進級照原版還要再隨機送 25−6 = 19 個,實得 %d", got)
	}
}

// 回傳的是拷貝——呼叫端改它不該汙染那張表。
func TestStartingTopicsReturnsACopy(t *testing.T) {
	got := StartingTopics(1)
	got[0] = TOPIC_STARTING_TECH
	if StartingTopicOrder[0] != TOPIC_ENGINEERING {
		t.Error("改了回傳值就把原表改掉了")
	}
}
