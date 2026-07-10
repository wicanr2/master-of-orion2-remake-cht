package gamedata

import (
	"strings"
	"testing"
)

// TestTopicEnglishNameCoversAll83 確保 83 個 ResearchTopic(0..82)都有英文顯示名,
// 不再退回 Topic#N 後備(對應 galaxy 畫面消除「研究主題 #N」fallback 的需求)。
func TestTopicEnglishNameCoversAll83(t *testing.T) {
	for i := 0; i < 83; i++ {
		got := TopicEnglishName(ResearchTopic(i))
		if got == "" || strings.HasPrefix(got, "Topic#") {
			t.Errorf("topic #%d 缺英文名,得 %q", i, got)
		}
	}
}

// TestTopicEnglishNameMatchTSV 確保每個 topic 英文名都能在 tech.tsv 查到中文,
// 否則顯示層翻譯會查無、退回英文(功能沒壞但翻譯缺漏)。
func TestTopicEnglishNameMatchTSV(t *testing.T) {
	keys := loadTechTSVKeys(t)
	for i := 0; i < 83; i++ {
		name := TopicEnglishName(ResearchTopic(i))
		if !keys[name] {
			t.Errorf("topic #%d 英文名 %q 不在 tech.tsv key 集合內", i, name)
		}
	}
}

// TestTopicEnglishNameKnownSamples 抽樣核對,防止 enum↔名配對跑掉。
func TestTopicEnglishNameKnownSamples(t *testing.T) {
	cases := []struct {
		topic ResearchTopic
		want  string
	}{
		{TOPIC_STARTING_TECH, "Starting Tech"},
		{TOPIC_ADVANCED_BIOLOGY, "Advanced Biology"},
		{TOPIC_CHEMISTRY, "Chemistry"}, // index 22:先前 galaxy 顯示「研究主題 #22」的案例
		{TOPIC_ANTIMATTER_FISSION, "Anti-Matter Fission"},
		{TOPIC_MULTIDIMENSIONAL_PHYSICS, "Multi-Dimensional Physics"},
		{TOPIC_HYPER_SOCIOLOGY, "Hyper Sociology"},
	}
	for _, c := range cases {
		if got := TopicEnglishName(c.topic); got != c.want {
			t.Errorf("TopicEnglishName(%d) = %q,預期 %q", int(c.topic), got, c.want)
		}
	}
}

// TestTopicEnglishNameFallback 查無回 Topic#N,不 panic。
func TestTopicEnglishNameFallback(t *testing.T) {
	if got := TopicEnglishName(ResearchTopic(9999)); got != "Topic#9999" {
		t.Errorf("TopicEnglishName(未知) = %q,預期 Topic#9999", got)
	}
}
