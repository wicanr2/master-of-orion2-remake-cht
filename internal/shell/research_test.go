package shell

import (
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func TestAvailableResearchTopics_CostsFromGamedata(t *testing.T) {
	opts := AvailableResearchTopics(nil)
	if len(opts) != 8 {
		t.Fatalf("8 個研究領域各一個主題,得到 %d 個", len(opts))
	}
	for _, o := range opts {
		want := gamedata.ResearchChoiceFor(o.Topic).Cost
		if o.Cost != want {
			t.Errorf("%s 成本 %d,期望取自 gamedata 的 %d", o.Name, o.Cost, want)
		}
		if o.Name == "" {
			t.Errorf("主題 #%d 缺名稱", int(o.Topic))
		}
	}
}

// 選單只給隊首:某領域的第二個主題,在第一個還沒完成前不該出現。
func TestAvailableResearchTopicsOffersOnlyTheHeadOfEachArea(t *testing.T) {
	head := AvailableResearchTopics(nil)
	got := map[gamedata.ResearchTopic]bool{}
	for _, o := range head {
		got[o.Topic] = true
	}
	// Construction 領域的順序是 ENGINEERING → ADVANCED_ENGINEERING → ADVANCED_CONSTRUCTION。
	if !got[gamedata.TOPIC_ENGINEERING] {
		t.Error("什麼都沒完成時,Construction 的隊首應是 TOPIC_ENGINEERING")
	}
	if got[gamedata.TOPIC_ADVANCED_CONSTRUCTION] {
		t.Error("TOPIC_ADVANCED_CONSTRUCTION 排在第三,前兩個沒完成不該出現")
	}
	// 完成前兩個之後,第三個才輪到。
	done := map[gamedata.ResearchTopic]bool{
		gamedata.TOPIC_ENGINEERING:          true,
		gamedata.TOPIC_ADVANCED_ENGINEERING: true,
	}
	s := &GameSession{}
	s.Player.CompletedTopics = done
	got2 := map[gamedata.ResearchTopic]bool{}
	for _, o := range AvailableResearchTopics(s) {
		got2[o.Topic] = true
	}
	if !got2[gamedata.TOPIC_ADVANCED_CONSTRUCTION] {
		t.Error("前兩個完成後,TOPIC_ADVANCED_CONSTRUCTION 應該可選了")
	}
	if got2[gamedata.TOPIC_ENGINEERING] {
		t.Error("已完成的主題不該再出現在選單上")
	}
}

func TestResearchQueueExcludesXenonTechnology(t *testing.T) {
	for _, topic := range researchQueue() {
		if topic == gamedata.TOPIC_XENON_TECHNOLOGY {
			t.Fatal("researchQueue 不應把 Xenon Technologies 排入正常研究")
		}
	}
}

func TestSetResearchTopicRejectsUnresearchableTopic(t *testing.T) {
	s := NewDemoSession()
	beforeTopic, beforeProgress := s.Player.ResearchTopic, s.Player.ResearchProgress
	s.SetResearchTopic(gamedata.TOPIC_XENON_TECHNOLOGY)
	if s.Player.ResearchTopic != beforeTopic || s.Player.ResearchProgress != beforeProgress {
		t.Fatalf("不可研究主題不應改變研究狀態: topic %v→%v progress %d→%d",
			beforeTopic, s.Player.ResearchTopic, beforeProgress, s.Player.ResearchProgress)
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
