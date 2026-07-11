package shell

import (
	"math/rand"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// TestBuildWeaponOptions_Profile13And15 驗證 BuildWeaponOptions 依 profile 覆寫「電漿砲」
// Value(最大傷害),其餘元件不受影響。
func TestBuildWeaponOptions_Profile13And15(t *testing.T) {
	find := func(opts []Component, name string) Component {
		for _, c := range opts {
			if c.Name == name {
				return c
			}
		}
		t.Fatalf("找不到元件 %q", name)
		return Component{}
	}

	opts13 := BuildWeaponOptions(gamedata.Profile13())
	if got := find(opts13, "電漿砲").Value; got != 30 {
		t.Errorf("Profile13 電漿砲 Value = %d,want 30", got)
	}

	opts15 := BuildWeaponOptions(gamedata.Profile15())
	if got := find(opts15, "電漿砲").Value; got != 20 {
		t.Errorf("Profile15 電漿砲 Value = %d,want 20", got)
	}

	// 其餘元件(如雷射)不受版本影響。
	if find(opts13, "雷射").Value != find(opts15, "雷射").Value {
		t.Errorf("雷射 Value 不應隨 profile 改變")
	}
}

// TestBuildWeaponOptions_Profile15MatchesPackageLevelWeaponOptions 是「預設 Profile15 = 現行
// 硬編值」的 no-op 回歸斷言:BuildWeaponOptions(Profile15()) 必須逐元件等於套件級 WeaponOptions
// (改用 profile 前唯一的資料來源)。
func TestBuildWeaponOptions_Profile15MatchesPackageLevelWeaponOptions(t *testing.T) {
	got := BuildWeaponOptions(gamedata.Profile15())
	if len(got) != len(WeaponOptions) {
		t.Fatalf("len(BuildWeaponOptions(Profile15())) = %d,want %d(=len(WeaponOptions))", len(got), len(WeaponOptions))
	}
	for i := range got {
		if got[i] != WeaponOptions[i] {
			t.Errorf("BuildWeaponOptions(Profile15())[%d] = %+v,want %+v(=WeaponOptions[%d])", i, got[i], WeaponOptions[i], i)
		}
	}
}

// TestNewDemoSession_DefaultsToProfile15 驗證 NewDemoSession 預設注入 Profile15(=現行行為,
// no-op),供未來「主選單選版本」流程對照。
func TestNewDemoSession_DefaultsToProfile15(t *testing.T) {
	s := NewDemoSession()
	want := gamedata.Profile15()
	if s.RuleProfile != want {
		t.Errorf("NewDemoSession().RuleProfile = %+v,want %+v(gamedata.Profile15())", s.RuleProfile, want)
	}
}

// TestSetRuleProfile 驗證 SetRuleProfile 是最小掛勾:可切換到 1.3 且立刻反映在欄位上。
func TestSetRuleProfile(t *testing.T) {
	s := NewDemoSession()
	s.SetRuleProfile(gamedata.Profile13())
	if s.RuleProfile != gamedata.Profile13() {
		t.Errorf("SetRuleProfile(Profile13()) 後 s.RuleProfile = %+v,want %+v", s.RuleProfile, gamedata.Profile13())
	}
}

// TestFleetBombardDamage_VolleyCountFollowsRuleProfile 驗證 fleetBombardDamage 的齊射輪數
// 讀 s.RuleProfile.BombardmentVolleys,而非寫死的 10。用「必中 + 固定滿傷」艦艇
// (deterministicBombardShip,見 orbital_bombardment_test.go)排除 rng 不確定性:每輪傷害固定,
// 故 Profile15(10 輪)的總傷害應為 Profile13(5 輪)的 2 倍。
func TestFleetBombardDamage_VolleyCountFollowsRuleProfile(t *testing.T) {
	build := func(p gamedata.RuleProfile) *GameSession {
		s := NewDemoSession()
		s.RuleProfile = p
		s.Ships = []Ship{deterministicBombardShip()}
		return s
	}

	s15 := build(gamedata.Profile15())
	dmg15 := s15.fleetBombardDamage(rand.New(rand.NewSource(1)))

	s13 := build(gamedata.Profile13())
	dmg13 := s13.fleetBombardDamage(rand.New(rand.NewSource(1)))

	if dmg15 != 10*101 {
		t.Fatalf("Profile15(10 輪) 總傷害 = %d,want %d(10 輪 * 每輪 101 固定滿傷)", dmg15, 10*101)
	}
	if dmg13 != 5*101 {
		t.Fatalf("Profile13(5 輪) 總傷害 = %d,want %d(5 輪 * 每輪 101 固定滿傷)", dmg13, 5*101)
	}
	if dmg15 != 2*dmg13 {
		t.Errorf("Profile15 總傷害應為 Profile13 的 2 倍(10 輪 vs 5 輪),got dmg15=%d dmg13=%d", dmg15, dmg13)
	}
}

// TestBuildShip_PlasmaDamageFollowsRuleProfile 驗證 BuildShip/BuildShipWithMods 造出的電漿砲艦
// WeaponAttack 隨 s.RuleProfile 變動(1.3=30/1.5=20,見 gamedata.RuleProfile.PlasmaCannonMaxDamage),
// 而非永遠讀套件級 WeaponOptions 的硬編 1.5 值。用固定艦級(末日之星)+ 無裝甲/護盾/特殊,排除
// 其他加成幹擾,直接比較兩個 profile 造出艦艇的 WeaponAttack 差 = 30-20 = 10。
func TestBuildShip_PlasmaDamageFollowsRuleProfile(t *testing.T) {
	const plasmaCannonIdx = 9 // WeaponOptions[9] = 電漿砲(見該切片宣告順序,session.go)

	build := func(p gamedata.RuleProfile) *GameSession {
		s := NewDemoSession()
		s.RuleProfile = p
		s.Player.BC = 100000 // 充裕國庫,排除「BC 不足建艦失敗」干擾
		return s
	}

	s13 := build(gamedata.Profile13())
	if ok := s13.BuildShip("末日之星", plasmaCannonIdx, 0, 0, 0); !ok {
		t.Fatalf("Profile13 BuildShip 失敗(預期 BC 充裕應成功)")
	}
	s15 := build(gamedata.Profile15())
	if ok := s15.BuildShip("末日之星", plasmaCannonIdx, 0, 0, 0); !ok {
		t.Fatalf("Profile15 BuildShip 失敗(預期 BC 充裕應成功)")
	}

	// 母星開局艦隊(homeworldShips)已佔用索引 0 起,新造艦附加在尾端。
	atk13 := s13.Ships[len(s13.Ships)-1].WeaponAttack
	atk15 := s15.Ships[len(s15.Ships)-1].WeaponAttack

	if atk13 != 30 {
		t.Errorf("Profile13 電漿砲艦 WeaponAttack = %d,want 30", atk13)
	}
	if atk15 != 20 {
		t.Errorf("Profile15 電漿砲艦 WeaponAttack = %d,want 20", atk15)
	}
	if atk13-atk15 != 10 {
		t.Errorf("Profile13-Profile15 WeaponAttack 差 = %d,want 10(30-20)", atk13-atk15)
	}
}

// TestEndTurn_HyperAdvancedResearchCostFollowsRuleProfile 驗證 EndTurn 把
// gamedata.HyperAdvancedCost(s.RuleProfile) 灌進 s.Player.HyperAdvancedResearchCost(供
// engine.RunResearchPhase 覆寫 Hyper-Advanced Lv1 研究成本,見 engine.PlayerState 該欄位註解)。
// Profile15(NewDemoSession 預設)= 25000 = 套件級硬編值,no-op 回歸;Profile13 = 15000,真的
// 改變。同時驗證 AI 對手(AIPlayers)套用同一份 profile,不會出現玩家 1.3、AI 仍 1.5 的規則不對稱。
func TestEndTurn_HyperAdvancedResearchCostFollowsRuleProfile(t *testing.T) {
	s15 := NewDemoSession() // 預設 Profile15
	s15.EndTurn()
	if s15.Player.HyperAdvancedResearchCost != 25000 {
		t.Errorf("Profile15 EndTurn 後 Player.HyperAdvancedResearchCost = %d,want 25000(套件級硬編值,no-op 回歸)",
			s15.Player.HyperAdvancedResearchCost)
	}
	for i, ai := range s15.AIPlayers {
		if ai.Player.HyperAdvancedResearchCost != 25000 {
			t.Errorf("Profile15 EndTurn 後 AIPlayers[%d].Player.HyperAdvancedResearchCost = %d,want 25000", i, ai.Player.HyperAdvancedResearchCost)
		}
	}

	s13 := NewDemoSession()
	s13.RuleProfile = gamedata.Profile13()
	s13.EndTurn()
	if s13.Player.HyperAdvancedResearchCost != 15000 {
		t.Errorf("Profile13 EndTurn 後 Player.HyperAdvancedResearchCost = %d,want 15000", s13.Player.HyperAdvancedResearchCost)
	}
	for i, ai := range s13.AIPlayers {
		if ai.Player.HyperAdvancedResearchCost != 15000 {
			t.Errorf("Profile13 EndTurn 後 AIPlayers[%d].Player.HyperAdvancedResearchCost = %d,want 15000", i, ai.Player.HyperAdvancedResearchCost)
		}
	}
}

// TestResearchCostForDisplay_FollowsRuleProfile 驗證顯示層 (*GameSession).ResearchCostForDisplay
// 對 Hyper-Advanced 主題套用 s.RuleProfile,對其餘主題與套件級 shell.ResearchCost/
// gamedata.ResearchChoiceFor 相同——確保畫面顯示的成本與 EndTurn 實際結算用的成本一致
// (不會顯示 1.5 的 25000,卻用 1.3 的 15000 結算)。
func TestResearchCostForDisplay_FollowsRuleProfile(t *testing.T) {
	hyperTopic := gamedata.TOPIC_HYPER_BIOLOGY

	s13 := NewDemoSession()
	s13.RuleProfile = gamedata.Profile13()
	if got := s13.ResearchCostForDisplay(hyperTopic); got != 15000 {
		t.Errorf("Profile13 ResearchCostForDisplay(HYPER_BIOLOGY) = %d,want 15000", got)
	}

	s15 := NewDemoSession()
	if got := s15.ResearchCostForDisplay(hyperTopic); got != 25000 {
		t.Errorf("Profile15 ResearchCostForDisplay(HYPER_BIOLOGY) = %d,want 25000", got)
	}

	// 非 Hyper 主題不受 profile 影響,應與套件級 ResearchCost 相同。
	nonHyper := gamedata.TOPIC_ADVANCED_BIOLOGY
	want := ResearchCost(nonHyper)
	if got := s13.ResearchCostForDisplay(nonHyper); got != want {
		t.Errorf("非 Hyper 主題 ResearchCostForDisplay = %d,want %d(= ResearchCost,不受 profile 影響)", got, want)
	}
}
