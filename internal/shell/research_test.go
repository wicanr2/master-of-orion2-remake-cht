package shell

import (
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

func TestResearchTopicName_Fallback(t *testing.T) {
	if got := ResearchTopicName(gamedata.TOPIC_STARTING_TECH); got != "起始科技" {
		t.Errorf("起始科技名 = %q", got)
	}
	// 未收錄主題應回後備字串而非空。
	if got := ResearchTopicName(gamedata.TOPIC_ADVANCED_GOVERNMENTS); got == "" {
		t.Error("未收錄主題不應回空字串")
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
