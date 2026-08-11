package main

import (
	"math/rand"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

func mkTurnScreen() *tacticalScreen {
	return &tacticalScreen{
		b: &sceneBuilder{},
		player: []shell.CombatShip{
			{Name: "甲艦", HP: 40, MaxHP: 40, Col: 1, Row: 1, Attack: 100, WeaponMin: 10, WeaponMax: 10},
			{Name: "乙艦", HP: 40, MaxHP: 40, Col: 1, Row: 2, Attack: 100, WeaponMin: 10, WeaponMax: 10},
		},
		enemy: []shell.CombatShip{{Name: "敵艦", HP: 200, MaxHP: 200, Col: 2, Row: 1, WeaponMin: 0, WeaponMax: 0}},
		sel:   0,
		rng:   rand.New(rand.NewSource(1)),
	}
}

func TestTacticalWaitMovesShipBehindReadyShips(t *testing.T) {
	ts := mkTurnScreen()
	ts.waitSelectedAction()
	if ts.sel != 1 {
		t.Fatalf("甲艦等待後應輪到乙艦,實得索引 %d", ts.sel)
	}
	if !ts.waited[0] || ts.acted[0] {
		t.Fatalf("WAIT 應標記等待但不可直接標成完成:waited=%v acted=%v", ts.waited, ts.acted)
	}

	ts.finishSelectedAction()
	if ts.sel != 0 || !ts.acted[1] {
		t.Fatalf("乙艦完成後應回到等待中的甲艦:sel=%d acted=%v", ts.sel, ts.acted)
	}
	if ts.round != 0 {
		t.Fatalf("尚有甲艦未行動時不應結算回合,實得回合 %d", ts.round)
	}
}

func TestTacticalDoneEndsRoundAfterLastShip(t *testing.T) {
	ts := mkTurnScreen()
	ts.finishSelectedAction()
	if ts.round != 0 || ts.sel != 1 {
		t.Fatalf("第一艦 DONE 後應輪到第二艦:round=%d sel=%d", ts.round, ts.sel)
	}
	ts.finishSelectedAction()
	if ts.round != 1 {
		t.Fatalf("最後一艦 DONE 後應進入下一回合,實得 %d", ts.round)
	}
	if len(ts.acted) != 2 || ts.acted[0] || ts.acted[1] {
		t.Fatalf("新回合應重置逐艦狀態:acted=%v", ts.acted)
	}
}

// 完整包可能沒有攜帶 COMBAT.LBX；此時仍會畫 fallback 控制列，且按鈕必須真的接到
// 戰術行為，不能退化成「看得到、點不到」。
func TestTacticalFallbackControlBarAcceptsClick(t *testing.T) {
	ts := mkTurnScreen()
	if ts.bar != nil {
		t.Fatal("測試前提應是不帶原版 COMBAT.LBX 控制列")
	}
	scan := barButtonsCHT[1]
	ts.update(shell.InputState{MouseX: scan.cx, MouseY: scan.cy, ClickReleased: true})
	if ts.mode != tacticalModeScan {
		t.Fatalf("fallback 控制列的 SCAN 應切進掃描模式，得到 %d", ts.mode)
	}
}

func TestTacticalManualFireOnlyUsesSelectedShip(t *testing.T) {
	ts := mkTurnScreen()
	before := ts.enemy[0].HP
	ts.fireSelectedShip(0)
	if !ts.player[0].Fired {
		t.Fatal("手動開火應標記選中的甲艦")
	}
	if ts.player[1].Fired {
		t.Fatal("逐艦行動不應讓未輪到的乙艦一起開火")
	}
	if ts.enemy[0].HP >= before {
		t.Fatalf("甲艦在射程內應造成傷害:%d → %d", before, ts.enemy[0].HP)
	}
	if ts.round != 0 {
		t.Fatalf("尚有乙艦未行動時不應立即結算敵方回合,實得 %d", ts.round)
	}
}

func TestTacticalManualFireRespectsWeaponArc(t *testing.T) {
	ts := &tacticalScreen{
		b: &sceneBuilder{},
		player: []shell.CombatShip{
			{Name: "前向艦", HP: 40, MaxHP: 40, Col: 1, Row: 1,
				Facing: 0, Attack: 100, WeaponMin: 10, WeaponMax: 10,
				WeaponName: "雷射", WeaponArc: gamedata.ARC_FWD},
			{Name: "待命艦", HP: 40, MaxHP: 40, Col: 1, Row: 2},
		},
		enemy: []shell.CombatShip{{Name: "後方目標", HP: 200, MaxHP: 200, Col: 0, Row: 1, WeaponMin: 0, WeaponMax: 0}},
		sel:   0,
		rng:   rand.New(rand.NewSource(1)),
	}
	before := ts.enemy[0].HP
	ts.fireSelectedShip(0)
	if ts.player[0].Fired {
		t.Fatal("目標在前向射界外時不應消費開火行動")
	}
	if ts.enemy[0].HP != before {
		t.Fatalf("射界外不應造成傷害: %d -> %d", before, ts.enemy[0].HP)
	}
	if ts.log == "" {
		t.Fatal("射界外應顯示可操作的提示")
	}

	ts.player[0].Facing = 8
	ts.fireSelectedShip(0)
	if !ts.player[0].Fired {
		t.Fatal("轉向目標後應能消費開火行動")
	}
}
