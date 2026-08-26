package main

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/assets"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

func TestBuildQueueOKHonorsAutoDeleteSetting(t *testing.T) {
	resolver, err := assets.NewResolver(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	makeScreen := func(auto bool) (*buildQueueScreen, *shell.GameSession) {
		sess := shell.NewDemoSession()
		sess.Builds[0] = shell.ColonyBuild{Name: shell.TradeGoodsBuildName}
		if !sess.EnqueueBuild(0, "研究實驗室", 60) {
			t.Fatal("測試前提：應能把一般產品排在貿易品後方")
		}
		settings := sess.EffectiveGameSettings()
		settings.AutoDeleteTradeGoodHousing = auto
		sess.ApplyGameSettings(settings)
		b := &sceneBuilder{session: sess, colonyIdx: 0, lang: i18n.Traditional, res: resolver}
		return &buildQueueScreen{b: b, idx: 0}, sess
	}

	screen, sess := makeScreen(false)
	tr := screen.update(shell.InputState{ClickReleased: true, MouseX: 590, MouseY: 455})
	if tr == nil {
		t.Fatal("設定關閉且模式阻塞產品時，OK 應開啟確認框")
	}
	if _, ok := tr.next.(*confirmScreen); !ok {
		t.Fatalf("設定關閉應轉到 confirmScreen，得到 %T", tr.next)
	}
	if got := sess.BlockingBuildMode(0); got != shell.TradeGoodsBuildName {
		t.Fatalf("尚未確認前不得刪除，得到 blocking=%q", got)
	}

	screen, sess = makeScreen(true)
	screen.update(shell.InputState{ClickReleased: true, MouseX: 590, MouseY: 455})
	if got := sess.BlockingBuildMode(0); got != "" || sess.Builds[0].Name != "研究實驗室" {
		t.Fatalf("設定開啟應直接清除阻塞模式並遞補產品：blocking=%q build=%+v", got, sess.Builds[0])
	}
}

func TestBuildQueueControlsToggleAutoAndRepeat(t *testing.T) {
	sess := shell.NewDemoSession()
	sess.Builds[0] = shell.ColonyBuild{}
	b := &sceneBuilder{session: sess, colonyIdx: 0, lang: i18n.Traditional}
	screen := &buildQueueScreen{b: b, idx: 0}

	screen.update(shell.InputState{ClickReleased: true, MouseX: 555, MouseY: 353})
	if !sess.ColonyAutoBuild(0) {
		t.Fatal("AUTO BUILD 熱區應切換規則層開關")
	}
	screen.update(shell.InputState{ClickReleased: true, MouseX: 559, MouseY: 420})
	if !screen.repeatArmed {
		t.Fatal("REPEAT BUILD 熱區應進入選取目標狀態")
	}
	if !sess.SetRepeatBuild(0, gamedata.ColonyShipActionName, 1) {
		t.Fatal("測試前提：殖民船應能指定為重複目標")
	}
	screen.repeatArmed = false
	screen.update(shell.InputState{ClickReleased: true, MouseX: 559, MouseY: 420})
	if got := sess.RepeatBuildFor(0); got.Name != "" {
		t.Fatalf("已有目標時再次按 REPEAT BUILD 應取消，得到 %+v", got)
	}
}

func TestRefitScreenPreviewsSelectedCandidate(t *testing.T) {
	sess := shell.NewDemoSession()
	sess.Builds[0] = shell.ColonyBuild{}
	sess.Player.CompletedTopics[gamedata.TOPIC_CHEMISTRY] = true
	sess.Fleet().Ships = append(sess.Fleet().Ships, shell.Ship{
		Name: "介面試作巡防艦", Class: "巡防艦",
		Weapon: "無武裝", Armor: "無裝甲", Shield: "無護盾", Special: "無",
	})
	b := &sceneBuilder{session: sess, colonyIdx: 0, lang: i18n.Traditional}
	screen := &refitScreen{b: b, colony: 0, selected: 0}
	if got := len(screen.candidates()); got != 1 {
		t.Fatalf("只應列出新加的可改裝巡防艦，得到 %d 艘", got)
	}
	job, err := screen.selectedPreview()
	if err != nil {
		t.Fatalf("選取艦艇應能預覽改裝：%v", err)
	}
	if job.Source.Name != "介面試作巡防艦" || job.Target.Armor != "鈦裝甲" {
		t.Fatalf("改裝預覽未接到規則層自動模板：%+v", job)
	}
}
