package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// fighter_test.go:戰機中隊的規則護欄。手冊引文見 fighter.go 檔頭。

// TestSquadronIsFourForEveryKind 釘住一個**被讀錯過**的真值。
//
// 舊值是「攔截機 4、重戰機 2」,來源寫著「手冊 GM p.127 出擊數欄」——
// 那一欄是 **Shots**(每架返航前開幾次火),不是中隊人數。中隊規模正文寫了兩次,一律 4:
// p.157「launched to a target in squadrons of four」、p.83「launched in squadrons of 4」。
func TestSquadronIsFourForEveryKind(t *testing.T) {
	if gamedata.FighterSquadronSize != 4 {
		t.Fatalf("中隊規模應為 4(手冊 p.157/p.83),實得 %d", gamedata.FighterSquadronSize)
	}
	for _, k := range []FighterKind{FighterInterceptor, FighterHeavy} {
		f := NewFighterSquadron(k, false, 0, 0, 0, 1, 0)
		if f.Alive != 4 {
			t.Errorf("%s 出擊應是 4 架,實得 %d", FighterKindName(k), f.Alive)
		}
	}
}

// 射擊次數才是那一欄:攔截機 4(「fire 4 times at point-blank range」)、
// 重戰機 2(「drop one bomb and fire a beam … then … the other bomb and fire a beam again」)。
func TestShotsMatchTheManualProse(t *testing.T) {
	if got := FighterShots(FighterInterceptor); got != 4 {
		t.Errorf("攔截機射擊次數應為 4,實得 %d", got)
	}
	if got := FighterShots(FighterHeavy); got != 2 {
		t.Errorf("重戰機射擊次數應為 2,實得 %d", got)
	}
}

// 血量:手冊給的是**每架**的(攔截機 2、重戰機 5),而且隨裝甲級每級 +2。
func TestFighterHitsArePerCraftAndScaleWithArmor(t *testing.T) {
	f := NewFighterSquadron(FighterInterceptor, false, 0, 0, 0, 1, 0)
	if f.HPEach != 2 {
		t.Errorf("攔截機每架血量應為 2,實得 %d", f.HPEach)
	}
	armored := NewFighterSquadron(FighterInterceptor, false, 0, 0, 0, 1, 3)
	if armored.HPEach != 2+2*3 {
		t.Errorf("裝甲高 3 級應為 2+6=8,實得 %d", armored.HPEach)
	}
	h := NewFighterSquadron(FighterHeavy, false, 0, 0, 0, 1, 0)
	if h.HPEach != 5 {
		t.Errorf("重戰機每架血量應為 5,實得 %d", h.HPEach)
	}
}

func TestEnemyFighterProfileScalesByGeneratedHullStrength(t *testing.T) {
	cases := []struct {
		strength int
		bay      bool
		kind     FighterKind
	}{
		{2, false, FighterInterceptor},
		{8, true, FighterInterceptor},
		{16, true, FighterHeavy},
		{32, true, FighterBomber},
	}
	for _, tc := range cases {
		kind, bay := EnemyFighterProfileForStrength(tc.strength)
		if kind != tc.kind || bay != tc.bay {
			t.Errorf("戰力 %d 的敵方戰機藍圖=%v/%v,期望=%v/%v", tc.strength, kind, bay, tc.kind, tc.bay)
		}
	}
}

// 傷害是一架一架吃的:打光一架才溢到下一架(手冊的血量是每架的,不是整隊一條血條)。
func TestDamageKillsCraftOneAtATime(t *testing.T) {
	f := NewFighterSquadron(FighterInterceptor, false, 0, 0, 0, 1, 0) // 4 架 × 2 血
	if killed := f.TakeHit(3); killed != 1 {
		t.Fatalf("3 點傷害應打掉 1 架,實得 %d", killed)
	}
	if f.Alive != 3 {
		t.Errorf("應剩 3 架,實得 %d", f.Alive)
	}
	if f.HPEach != 1 {
		t.Errorf("下一架應剩 1 血,實得 %d", f.HPEach)
	}
	if killed := f.TakeHit(100); killed != 3 || !f.Dead() {
		t.Errorf("過量傷害應把剩下 3 架打光,實得 killed=%d alive=%d", killed, f.Alive)
	}
	// 打光之後再打不該打出負數或幽靈架數。
	if killed := f.TakeHit(10); killed != 0 || f.Alive != 0 {
		t.Errorf("已全滅的中隊不該再被打掉東西,實得 killed=%d alive=%d", killed, f.Alive)
	}
}

// 只有攔截機能打別的戰機(手冊 p.157 的 dogfight 條款)。
func TestOnlyInterceptorsDogfight(t *testing.T) {
	i := NewFighterSquadron(FighterInterceptor, false, 0, 0, 0, 1, 0)
	h := NewFighterSquadron(FighterHeavy, false, 0, 0, 0, 1, 0)
	if !i.CanTargetFighter() {
		t.Error("攔截機應該可以打敵方戰機")
	}
	if h.CanTargetFighter() {
		t.Error("重戰機不該能打敵方戰機(手冊:除了攔截機都不能纏鬥)")
	}
}

// 打完最後一輪就轉返航;回到母艦後補血補彈,但**不補人**。
func TestOutOfShotsReturnsAndRecoveryDoesNotRefillCrew(t *testing.T) {
	f := NewFighterSquadron(FighterHeavy, false, 2, 0, 0, 1, 0) // 2 輪
	f.TakeHit(5)                                                // 掉一架
	f.SpendShot()
	if f.Returning {
		t.Error("還有一輪就不該返航")
	}
	f.SpendShot()
	if !f.Returning {
		t.Error("彈藥用盡應該返航(手冊:out of shots → return to carrier)")
	}
	f.Recover()
	if f.ShotsLeft != FighterShots(FighterHeavy) || f.HPEach != f.MaxHPEach || f.Returning {
		t.Errorf("返航應補滿彈與血並解除返航狀態,實得 shots=%d hp=%d returning=%v",
			f.ShotsLeft, f.HPEach, f.Returning)
	}
	if f.Alive != 3 {
		t.Errorf("返航不補人——手冊寫的是 any **surviving** fighters,應仍是 3 架,實得 %d", f.Alive)
	}
}

// 速度走 CombatFighterSpeed:基礎 + 2×(FTL−1) + 跨維 4。
func TestSquadronSpeedUsesTheManualFormula(t *testing.T) {
	f := NewFighterSquadron(FighterInterceptor, false, 0, 0, 0, 3, 0)
	want := gamedata.CombatFighterSpeed(gamedata.CombatFighterBaseSpeedInterceptor, 3)
	if f.Speed != want {
		t.Errorf("速度應為 %d,實得 %d", want, f.Speed)
	}
	if f.Speed <= gamedata.CombatFighterBaseSpeedInterceptor {
		t.Error("測試前提不成立:FTL 3 的速度應該比基礎速度快,否則這支測試等於沒驗公式")
	}
}

// 移動:每回合最多走 Speed 格,貼到目標旁邊(曼哈頓 ≤1)就算到位——
// 手冊說戰機是飛到目標身上 at point-blank range 開火的,不像艦艇有射程。
func TestStepTowardStopsAtPointBlank(t *testing.T) {
	f := NewFighterSquadron(FighterInterceptor, false, 0, 0, 0, 1, 0)
	f.Speed = 2
	if f.StepToward(7, 0) {
		t.Error("兩格的移動力不該一次飛到 7 格外")
	}
	if f.Col != 2 || f.Row != 0 {
		t.Errorf("應走到 (2,0),實得 (%d,%d)", f.Col, f.Row)
	}
	f.Speed = 10
	if !f.StepToward(7, 0) {
		t.Error("移動力足夠時應該貼到目標旁邊")
	}
	if d := fighterDist(f.Col, f.Row, 7, 0); d > 1 {
		t.Errorf("應停在目標旁(距離 ≤1),實得 %d", d)
	}
}

// 球狀武器/自爆:戰機有 50% 機率閃掉(手冊 p.157)。
func TestFighterAvoidsSphericalIsFiftyPercent(t *testing.T) {
	avoided := 0
	for roll := 1; roll <= 100; roll++ {
		if FighterAvoidsSpherical(roll) {
			avoided++
		}
	}
	if avoided != 50 {
		t.Errorf("1..100 應有 50 個骰值閃得掉,實得 %d", avoided)
	}
}

// 母艦沒了要記成 −1 不是 0——0 是第一艘船(同 Star.Wormhole/ColonyRelocateTo 那個零值陷阱)。
func TestCarrierNoneIsMinusOne(t *testing.T) {
	f := NewFighterSquadron(FighterInterceptor, false, -1, 3, 2, 1, 0)
	if f.Carrier != -1 {
		t.Errorf("沒有母艦應記 −1,實得 %d", f.Carrier)
	}
}
