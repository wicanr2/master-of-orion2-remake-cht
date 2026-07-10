package shell

import (
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func TestStarterResearchTopics_CostsFromGamedata(t *testing.T) {
	opts := StarterResearchTopics()
	if len(opts) == 0 {
		t.Fatal("預期至少一個可選研究主題")
	}
	for _, o := range opts {
		want := gamedata.ResearchChoiceFor(o.Topic).Cost
		if o.Cost != want {
			t.Errorf("%s 成本 %d,期望取自 gamedata 的 %d", o.Name, o.Cost, want)
		}
		if o.Name == "" {
			t.Errorf("主題 #%d 缺譯名", int(o.Topic))
		}
	}
}

func TestResearchTopicName_ReturnsEnglishKey(t *testing.T) {
	// ResearchTopicName 現回英文顯示名(= tech.tsv 的 i18n key),中文由顯示層翻。
	if got := ResearchTopicName(gamedata.TOPIC_STARTING_TECH); got != "Starting Tech" {
		t.Errorf("起始科技英文名 = %q,期望 Starting Tech", got)
	}
	// 先前 fallback「研究主題 #N」的主題(如 TOPIC_ADVANCED_GOVERNMENTS)現應有英文名。
	if got := ResearchTopicName(gamedata.TOPIC_ADVANCED_GOVERNMENTS); got != "Advanced Governments" {
		t.Errorf("進階政體英文名 = %q,期望 Advanced Governments", got)
	}
	// 83 個 topic 全收錄,不再有 Topic#N 後備。
	for i := 0; i < 83; i++ {
		got := ResearchTopicName(gamedata.ResearchTopic(i))
		if got == "" || strings.HasPrefix(got, "Topic#") {
			t.Errorf("topic #%d 未收錄英文名,得 %q", i, got)
		}
	}
}

func TestSetResearchTopic_ResetsProgressOnChange(t *testing.T) {
	s := NewDemoSession()
	s.Player.ResearchProgress = 123
	orig := s.Player.ResearchTopic

	// 切到不同主題:進度歸零。
	s.SetResearchTopic(gamedata.TOPIC_MILITARY_TACTICS)
	if s.Player.ResearchTopic != gamedata.TOPIC_MILITARY_TACTICS {
		t.Fatalf("主題未切換,得 %d", int(s.Player.ResearchTopic))
	}
	if s.Player.ResearchProgress != 0 {
		t.Errorf("換題後進度應歸零,得 %d", s.Player.ResearchProgress)
	}

	// 切回相同主題:進度不變(不重置)。
	s.Player.ResearchProgress = 50
	s.SetResearchTopic(gamedata.TOPIC_MILITARY_TACTICS)
	if s.Player.ResearchProgress != 50 {
		t.Errorf("切回同主題不應重置進度,得 %d", s.Player.ResearchProgress)
	}
	_ = orig
}
