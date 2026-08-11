package main

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

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
