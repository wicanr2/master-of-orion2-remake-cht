package main

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
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

func TestMultipleDistinctBaysLaunchOneSquadronEach(t *testing.T) {
	ts := mkFighterScreen()
	ts.player[0].Bays = []shell.FighterKind{shell.FighterInterceptor, shell.FighterHeavy}
	ts.launchFrom(0)
	if !ts.canLaunchFrom(0) {
		t.Fatal("第一座機庫出擊後，第二座不同機庫仍應可出擊")
	}
	ts.launchFrom(0)
	if ts.canLaunchFrom(0) || len(ts.squads) != 2 || ts.squads[0].Kind == ts.squads[1].Kind {
		t.Fatalf("兩座機庫應各派一隊：%+v", ts.squads)
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

// 戰機接戰前,敵艦尚未使用過的 PD 應先開火(手冊 p.117),並能消耗戰機中隊。
func TestPointDefenseFiresBeforeFighterEngagement(t *testing.T) {
	ts := mkFighterScreen()
	ts.enemy[0].WeaponName = "雷射"
	ts.enemy[0].Mods = []string{string(gamedata.ModPointDefense)}
	ts.enemy[0].Attack = 100
	ts.enemy[0].WeaponMax = 8
	ts.launchFrom(0)
	before := ts.squads[0].Alive

	ts.advanceSquadrons()
	if !ts.enemy[0].PointDefenseSpent {
		t.Fatal("戰機接戰前敵艦的 PD 應標記為本回合已使用")
	}
	if ts.squads[0].Alive >= before {
		t.Fatalf("PD 命中後應先打掉至少一架戰機，接戰前後 %d→%d", before, ts.squads[0].Alive)
	}
}

// 出擊時要把參戰艦隊算好的戰機 Beam Defense 加成帶進中隊，不能只留在 CombatShip。
func TestLaunchCarriesFighterBeamDefenseBonuses(t *testing.T) {
	ts := mkFighterScreen()
	ts.player[0].FighterRacialDefenseBonus = 25
	ts.player[0].FighterPilotBonus = 10
	ts.player[0].FighterHelmsmanBonus = 5
	ts.launchFrom(0)
	if len(ts.squads) != 1 {
		t.Fatalf("應建立一隊戰機，實得 %d 隊", len(ts.squads))
	}
	f := ts.squads[0]
	if f.FighterRacialDefenseBonus != 25 || f.FighterPilotBonus != 10 || f.FighterHelmsmanBonus != 5 {
		t.Fatalf("戰機 Beam Defense 加成未隨出擊帶入: %+v", f)
	}
	if got := gamedata.CombatFighterBeamDefense(f.Speed,
		f.FighterRacialDefenseBonus, f.FighterPilotBonus, f.FighterHelmsmanBonus); got <= 5*f.Speed {
		t.Fatalf("帶加成的戰機防禦應高於基礎 5*Speed，實得 %d", got)
	}
}

// 主要目標仍有效時，中隊不應因另一艘艦變得更近就自動改追；
// 手冊 p.157 只允許在主要目標失效後重選。
func TestSquadronKeepsValidPrimaryTarget(t *testing.T) {
	ts := mkFighterScreen()
	ts.enemy = append(ts.enemy, shell.CombatShip{
		Name: "近敵艦", HP: 100, MaxHP: 100, Col: 4, Row: 0,
		WeaponMin: 1, WeaponMax: 1,
	})
	ts.launchFrom(0)
	ts.advanceSquadrons()
	if got := ts.squads[0].TargetName; got != "近敵艦" {
		t.Fatalf("第一次出擊應鎖定最近的主要目標，得到 %q", got)
	}

	// 讓另一艘艦變成幾何上更近的目標；原本的「近敵艦」仍然存活。
	ts.enemy[0].Col = 2
	if got := ts.advanceSquadrons(); got <= 0 {
		t.Fatal("主要目標仍有效時，第二輪仍應對它開火")
	}
	if got := ts.squads[0].TargetName; got != "近敵艦" {
		t.Errorf("有效主要目標不應被最近距離改寫，得到 %q", got)
	}
}

// 主要目標被摧毀後，中隊才會自動改鎖另一艘仍存活的敵艦。
func TestSquadronRetargetsOnlyAfterPrimaryTargetDies(t *testing.T) {
	ts := mkFighterScreen()
	ts.enemy = append(ts.enemy, shell.CombatShip{
		Name: "次要艦", HP: 100, MaxHP: 100, Col: 4, Row: 0,
		WeaponMin: 1, WeaponMax: 1,
	})
	ts.launchFrom(0)
	ts.advanceSquadrons()
	if ts.squads[0].TargetName != "次要艦" {
		t.Fatalf("前置條件:應先鎖定次要艦，得到 %q", ts.squads[0].TargetName)
	}
	ts.enemy[1].HP = 0
	ts.enemy[0].Col = 4
	if got := ts.advanceSquadrons(); got <= 0 {
		t.Fatal("主要目標失效後應重選並開火")
	}
	if got := ts.squads[0].TargetName; got != "敵艦" {
		t.Errorf("主要目標失效後應改鎖敵艦，得到 %q", got)
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

func TestEnemySquadronFliesBackAndAttacksPlayerShip(t *testing.T) {
	ts := &tacticalScreen{
		b:      &sceneBuilder{},
		player: []shell.CombatShip{{Name: "玩家艦", HP: 80, MaxHP: 80, Col: 1, Row: 0}},
		enemy:  []shell.CombatShip{{Name: "敵母艦", HP: 100, MaxHP: 100, Col: 6, Row: 0, Bay: true, BayKind: shell.FighterInterceptor}},
	}
	ts.launchEnemySquadrons()
	if len(ts.squads) != 1 || !ts.squads[0].Enemy {
		t.Fatalf("敵方戰機庫應在開場建立一隊敵方中隊，實得 %+v", ts.squads)
	}
	before := ts.player[0].HP
	ts.advanceSquadrons()
	if ts.player[0].HP >= before {
		t.Fatalf("敵方中隊接戰後應傷害玩家艦，血量 %d→%d", before, ts.player[0].HP)
	}
	if ts.squads[0].TargetName != "玩家艦" {
		t.Errorf("敵方中隊主要目標應鎖定玩家艦，得到 %q", ts.squads[0].TargetName)
	}
}

func TestSecondSlotPointDefenseFiresAtEnemyFighterWhileOff(t *testing.T) {
	ts := &tacticalScreen{
		b: &sceneBuilder{},
		player: []shell.CombatShip{{
			Name: "玩家艦", HP: 80, MaxHP: 80, Col: 1, Row: 0,
			Attack: 100, WeaponName: "雷射", WeaponMax: 8,
			WeaponMounts: []shell.ShipWeaponMount{
				{Name: "雷射", WorkingCount: 1, Attack: 8},
				{Name: "雷射", WorkingCount: 1, Attack: 20,
					Mods: []string{string(gamedata.ModPointDefense)}},
			},
			WeaponModes: []shell.TacticalWeaponMode{
				shell.TacticalWeaponReady, shell.TacticalWeaponOff,
			},
		}},
		enemy: []shell.CombatShip{{
			Name: "敵母艦", HP: 100, MaxHP: 100, Col: 6, Row: 0,
			Bay: true, BayKind: shell.FighterInterceptor,
		}},
	}
	ts.launchEnemySquadrons()
	before := ts.squads[0].Alive
	ts.advanceSquadrons()
	if len(ts.player[0].PointDefenseSpentSlots) < 2 || !ts.player[0].PointDefenseSpentSlots[1] {
		t.Fatalf("第二槽 PD 應在敵方戰機接戰前自動開火：%v", ts.player[0].PointDefenseSpentSlots)
	}
	if ts.player[0].WeaponModes[1] != shell.TacticalWeaponOff {
		t.Fatalf("自動 PD 不得把紅色槽切回可用：%v", ts.player[0].WeaponModes)
	}
	if ts.squads[0].Alive >= before {
		t.Fatalf("紅色第二槽 PD 應先傷害敵方中隊：%d→%d", before, ts.squads[0].Alive)
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

// TestTacticalBarButtonsAreClickable 控制列七顆鈕**每一顆都要點得到**。
//
// 這條測的是第 87 項(控制列熱區)那個缺口:七顆鈕中文化做完了、熱區一個都沒有,
// 畫面上看起來能點、點下去什麼都不會發生。**「按鈕長得對」與「按鈕能按」是兩件事**,
// 而截圖只能證明前者——所以這件事只能用測試守。
func TestTacticalBarButtonsAreClickable(t *testing.T) {
	for i, b := range barButtonsCHT {
		if got := barButtonHit(b.cx, b.cy); got != i {
			t.Errorf("點 %s(%s)的正中央命中第 %d 顆,期望第 %d 顆", b.label, b.orig, got, i)
		}
	}
	// 控制列以外不該中(棋盤區在 y<365)。
	if got := barButtonHit(320, 200); got != -1 {
		t.Errorf("棋盤中央(320,200)不該命中控制列,實得 %d", got)
	}
}

// TestTacticalBarButtonsDoNotOverlap 相鄰按鈕的熱區不可重疊(重疊會讓其中一顆永遠點不到)。
func TestTacticalBarButtonsDoNotOverlap(t *testing.T) {
	for i := range barButtonsCHT {
		for j := i + 1; j < len(barButtonsCHT); j++ {
			a, b := barButtonsCHT[i], barButtonsCHT[j]
			if abs(a.cx-b.cx) < 54 && abs(a.cy-b.cy) < 18 {
				t.Errorf("%s 與 %s 的熱區重疊(中心相距 %d,%d)",
					a.label, b.label, abs(a.cx-b.cx), abs(a.cy-b.cy))
			}
		}
	}
}
