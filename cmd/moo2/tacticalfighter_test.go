package main

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

// tacticalfighter_test.go:戰機在格子戰場上的接線護欄(規則本身的測試在 internal/shell)。

// mkFighterScreen 造一個最小戰場:一艘帶戰機庫的我方艦 + 一艘遠處的敵艦。
func mkFighterScreen() *tacticalScreen {
	return &tacticalScreen{
		// b 只被 tr()/log 用到,零值 sceneBuilder(lang 零值 = 繁中)就夠。
		b:      &sceneBuilder{},
		player: []shell.CombatShip{{Name: "母艦", HP: 30, MaxHP: 30, Col: 1, Row: 0, Bay: true}},
		enemy:  []shell.CombatShip{{Name: "敵艦", HP: 100, MaxHP: 100, Col: 6, Row: 0, WeaponMin: 4, WeaponMax: 8}},
		sel:    0,
	}
}

// 一個戰機庫同時只有一隊在場:出擊後不能再出擊,全滅或返航補給後才可以。
func TestOneSquadronPerBayAtATime(t *testing.T) {
	ts := mkFighterScreen()
	if !ts.canLaunchFrom(0) {
		t.Fatal("帶戰機庫的艦應該可以出擊")
	}
	ts.launchFrom(0)
	if ts.canLaunchFrom(0) {
		t.Error("同一個戰機庫的中隊已在場,不該能再出擊一隊")
	}
	// 沒有戰機庫的艦不能出擊。
	ts.player[0].Bay = false
	ts.squads = nil
	if ts.canLaunchFrom(0) {
		t.Error("沒有戰機庫的艦不該能出擊")
	}
}

// 出擊後推進一回合:中隊要飛到敵艦旁邊並造成傷害(手冊:飛到目標身上 point-blank 開火)。
func TestSquadronFliesToTargetAndDealsDamage(t *testing.T) {
	ts := mkFighterScreen()
	ts.launchFrom(0)
	before := ts.enemy[0].HP
	dmg := ts.advanceSquadrons()
	if dmg <= 0 {
		t.Fatalf("中隊應該打得到敵艦,實得傷害 %d", dmg)
	}
	if ts.enemy[0].HP != before-dmg {
		t.Errorf("敵艦血量應扣 %d(%d → %d),實得 %d", dmg, before, before-dmg, ts.enemy[0].HP)
	}
	if d := abs(ts.squads[0].Col-6) + abs(ts.squads[0].Row-0); d > 1 {
		t.Errorf("中隊應停在敵艦旁(距離 ≤1),實得 %d", d)
	}
}

// 彈藥用盡就返航,回到母艦補給後可以再出擊。
func TestSquadronReturnsThenCanLaunchAgain(t *testing.T) {
	ts := mkFighterScreen()
	ts.launchFrom(0)
	for i := 0; i < shell.FighterShots(shell.FighterInterceptor); i++ {
		ts.advanceSquadrons()
	}
	if !ts.squads[0].Returning {
		t.Fatal("打完所有輪次應該返航")
	}
	ts.advanceSquadrons() // 飛回母艦
	if ts.squads[0].Returning {
		t.Error("回到母艦後應該補給完畢、解除返航狀態")
	}
	if ts.squads[0].ShotsLeft != shell.FighterShots(shell.FighterInterceptor) {
		t.Errorf("補給後彈藥應補滿,實得 %d", ts.squads[0].ShotsLeft)
	}
}

// 貼身的敵艦會把戰機打下來(手冊:fighter craft are vulnerable to beam weapons)。
func TestEnemyShootsDownAdjacentFighters(t *testing.T) {
	ts := mkFighterScreen()
	ts.launchFrom(0)
	ts.advanceSquadrons() // 飛到敵艦旁
	killed := ts.enemyFiresAtSquadrons()
	if killed <= 0 {
		t.Fatalf("貼身的敵艦應該打得下戰機,實得 %d", killed)
	}
	if ts.squads[0].Alive != 4-killed {
		t.Errorf("剩餘架數應為 %d,實得 %d", 4-killed, ts.squads[0].Alive)
	}
	// 全滅的中隊要被移出場,不能留一個空 token 在格子上。
	for i := 0; i < 20; i++ {
		ts.enemyFiresAtSquadrons()
	}
	ts.dropDeadSquadrons()
	if len(ts.squads) != 0 {
		t.Errorf("全滅的中隊應被移出場,實得 %d 隊", len(ts.squads))
	}
}

// 出擊鈕要留在畫面裡(640×480)——放到格線右邊會超出螢幕。
func TestLaunchButtonIsOnScreen(t *testing.T) {
	x, y, w, h := launchRect()
	if x < 0 || y < 0 || x+w > moo2ScreenW || y+h > moo2ScreenH {
		t.Errorf("出擊鈕 (%d,%d,%d,%d) 超出 %d×%d 畫面", x, y, w, h, moo2ScreenW, moo2ScreenH)
	}
}

// --- 艦隊列表的 ALL 鈕(手冊 p.32 + p.47:全選/全不選艦艇,不是集結點)---

// mkFleetBuilder 造一個有 3 艘船的最小 session builder。
func mkFleetBuilder() *sceneBuilder {
	s := shell.NewDemoSession()
	s.Fleet().Ships = []shell.Ship{{Name: "甲"}, {Name: "乙"}, {Name: "丙"}}
	return &sceneBuilder{session: s}
}

// 按一次全選,再按一次全不選(手冊括號那句:「If all the ships are already selected,
// this deselects them instead.」)。
func TestSelectAllTogglesInsteadOfAlwaysSelecting(t *testing.T) {
	b := mkFleetBuilder()
	b.toggleSelectAllShips()
	for i := 0; i < 3; i++ {
		if !b.shipPick[i] {
			t.Fatalf("第一次按 ALL 應全選,第 %d 艘沒選到", i)
		}
	}
	b.toggleSelectAllShips()
	for i := 0; i < 3; i++ {
		if b.shipPick[i] {
			t.Errorf("已全選時再按 ALL 應全不選,第 %d 艘仍選著", i)
		}
	}
}

// 只選了一部分時按 ALL 是「補成全選」,不是「取消」——
// 手冊的條件是 **all** the ships are already selected。
func TestSelectAllFromPartialSelectionSelectsEverything(t *testing.T) {
	b := mkFleetBuilder()
	b.shipPick = map[int]bool{1: true}
	b.toggleSelectAllShips()
	for i := 0; i < 3; i++ {
		if !b.shipPick[i] {
			t.Errorf("部分選取時按 ALL 應補成全選,第 %d 艘沒選到", i)
		}
	}
}
