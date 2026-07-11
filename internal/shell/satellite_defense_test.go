package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// satellite_defense_test.go:直接測試 retaliationAttackers/bestUnlockedWeaponValue 的
// 版本相依 arc-cost 與科技效果(#14),不透過 BombardColony 端對端組裝(端對端回歸測試見
// orbital_bombardment_test.go 既有的 TestBombardColony_Retaliation* 系列)。

// TestBestUnlockedWeaponValue_FallbackWhenNothingUnlocked 驗證:defender 未解鎖任何武器科技時
// (CompletedTopics 為 nil,比照 buildDemoAIOpponents 開局 AI 狀態),beam/missile 各自回退到
// 雷射(Value4/space10)/核飛彈(Value6/space10),ok=false。
func TestBestUnlockedWeaponValue_FallbackWhenNothingUnlocked(t *testing.T) {
	ps := engine.PlayerState{}
	profile := gamedata.Profile15()

	if v, sp, ok := bestUnlockedWeaponValue(ps, profile, WeaponKindBeam); v != 4 || sp != 10 || ok {
		t.Errorf("beam fallback = (%d, %d, %v),want (4, 10, false)", v, sp, ok)
	}
	if v, sp, ok := bestUnlockedWeaponValue(ps, profile, WeaponKindMissile); v != 6 || sp != 10 || ok {
		t.Errorf("missile fallback = (%d, %d, %v),want (6, 10, false)", v, sp, ok)
	}
}

// TestBestUnlockedWeaponValue_PicksHighestUnlockedBeam 驗證:defender 解鎖電漿砲科技
// (TOPIC_PLASMA_PHYSICS)後,beam 挑到電漿砲(Value 依 profile 版本相依)而非雷射,ok=true。
func TestBestUnlockedWeaponValue_PicksHighestUnlockedBeam(t *testing.T) {
	ps := engine.PlayerState{CompletedTopics: map[gamedata.ResearchTopic]bool{
		gamedata.TOPIC_PLASMA_PHYSICS: true,
	}}
	profile := gamedata.Profile15()

	v, sp, ok := bestUnlockedWeaponValue(ps, profile, WeaponKindBeam)
	if !ok {
		t.Fatalf("ok = false,want true(已解鎖電漿砲)")
	}
	if v != profile.PlasmaCannonMaxDamage {
		t.Errorf("Value = %d,want %d(= profile.PlasmaCannonMaxDamage)", v, profile.PlasmaCannonMaxDamage)
	}
	if want := gamedata.WeaponSpaceByName["電漿砲"]; sp != want {
		t.Errorf("Space = %d,want %d", sp, want)
	}
}

// TestRetaliationAttackers_VersionDifference 驗證版本效果:同一 AI(未解鎖任何武器,雷射
// fallback)、同一防禦建築(星基),Profile13(arc25%)算出的反擊 atk 應高於 Profile15
// (arc33%)——arc-cost 較高 → perBeam 較大 → fit 較少 → 防禦略弱,這是 #14 版本差異在
// retaliationAttackers 的落地效果。
func TestRetaliationAttackers_VersionDifference(t *testing.T) {
	defender := engine.PlayerState{} // 未解鎖任何武器,雷射 fallback(Value4/space10)
	buildings := map[string]bool{"星基": true}

	a13 := retaliationAttackers(buildings, defender, gamedata.Profile13())
	a15 := retaliationAttackers(buildings, defender, gamedata.Profile15())
	if len(a13) != 1 || len(a15) != 1 {
		t.Fatalf("預期各恰有 1 個軌道基地 attacker,got len(a13)=%d len(a15)=%d", len(a13), len(a15))
	}

	// 手算(見 gamedata.SatelliteStrengthScale 註解推導同款算法):
	//   1.3 arc25%:perBeam=10+10*25/100=12,fit=250/12=20,atk=20*4/20=4。
	//   1.5 arc33%:perBeam=10+10*33/100=13,fit=250/13=19,atk=19*4/20=3。
	if a13[0].atk != 4 {
		t.Errorf("Profile13 星基 atk = %d,want 4", a13[0].atk)
	}
	if a15[0].atk != 3 {
		t.Errorf("Profile15 星基 atk = %d,want 3", a15[0].atk)
	}
	if a15[0].atk >= a13[0].atk {
		t.Errorf("1.5 atk(%d)應 < 1.3 atk(%d):arc-cost 較高、軌道防禦應略弱", a15[0].atk, a13[0].atk)
	}
}

// TestRetaliationAttackers_TechEffect 驗證科技效果:同一版本(Profile15)、同一防禦建築
// (星基),defender 解鎖電漿砲後的反擊 atk 應高於未解鎖(雷射 fallback)——呼應手冊「軌道基地
// 配備目前最好的武器」的定性描述,基地戰力應隨科技進步變強。
func TestRetaliationAttackers_TechEffect(t *testing.T) {
	profile := gamedata.Profile15()
	buildings := map[string]bool{"星基": true}

	baseline := retaliationAttackers(buildings, engine.PlayerState{}, profile)
	withPlasma := retaliationAttackers(buildings, engine.PlayerState{
		CompletedTopics: map[gamedata.ResearchTopic]bool{gamedata.TOPIC_PLASMA_PHYSICS: true},
	}, profile)

	if len(baseline) != 1 || len(withPlasma) != 1 {
		t.Fatalf("預期各恰有 1 個軌道基地 attacker,got len(baseline)=%d len(withPlasma)=%d", len(baseline), len(withPlasma))
	}
	if withPlasma[0].atk <= baseline[0].atk {
		t.Errorf("解鎖電漿砲後 atk(%d)應 > 未解鎖 atk(%d)(科技效果)", withPlasma[0].atk, baseline[0].atk)
	}
	// 手算:電漿砲 Value=20(Profile15)、space=25,perBeam=25+25*33/100=25+8=33,
	// fit=250/33=7,atk=7*20/20=7;雷射 fallback atk=3(見上一測試推導)。
	if withPlasma[0].atk != 7 {
		t.Errorf("解鎖電漿砲後星基 atk = %d,want 7", withPlasma[0].atk)
	}
}

// TestRetaliationAttackers_MissileBaseIgnoresArcCost 驗證飛彈基地不吃 beam 的 arc-cost:
// 同一防禦建築(僅飛彈基地),Profile13/Profile15 的 arc-cost 不同,但飛彈基地 atk 應相同。
func TestRetaliationAttackers_MissileBaseIgnoresArcCost(t *testing.T) {
	defender := engine.PlayerState{} // 核飛彈 fallback(Value6/space10)
	buildings := map[string]bool{"飛彈基地": true}

	a13 := retaliationAttackers(buildings, defender, gamedata.Profile13())
	a15 := retaliationAttackers(buildings, defender, gamedata.Profile15())
	if len(a13) != 1 || len(a15) != 1 {
		t.Fatalf("預期各恰有 1 個飛彈基地 attacker,got len(a13)=%d len(a15)=%d", len(a13), len(a15))
	}
	if a13[0].kind != WeaponKindMissile || a15[0].kind != WeaponKindMissile {
		t.Fatalf("預期 kind=WeaponKindMissile,got a13=%v a15=%v", a13[0].kind, a15[0].kind)
	}
	// fit=300/10=30,atk=30*6/20=9,版本間應相同(missile 不套 SatelliteBeamArcCostPct)。
	if a13[0].atk != 9 || a15[0].atk != 9 {
		t.Errorf("飛彈基地 atk 應與版本無關皆為 9,got a13=%d a15=%d", a13[0].atk, a15[0].atk)
	}
}

// TestRetaliationAttackers_NoDefensiveBuildingsEmpty 回歸測試:無任何防禦建築(含地面砲台)
// 時回傳空 slice。
func TestRetaliationAttackers_NoDefensiveBuildingsEmpty(t *testing.T) {
	out := retaliationAttackers(map[string]bool{}, engine.PlayerState{}, gamedata.Profile15())
	if len(out) != 0 {
		t.Errorf("無防禦建築應回傳空 slice,got len=%d", len(out))
	}
}

// TestBombardColony_BalanceSanity_OpeningFleetVsHomeworldStarBase 是 #14 改版後的平衡守門測試
// (使用者硬性要求:不可讓開局轟炸退化成全軍覆沒)。用 NewDemoSession 開局的原始三艘無武裝
// 起始艦(homeworldShips:拓荒號/先驅一號/先驅二號,見 session.go)轟炸開局 AI 母星(只有
// 「海軍陸戰隊營」+「星基」,即 homeworldBuildings 預設值),Profile13/Profile15 各自掃過
// Turn=0..14(對應 BombardColony 內部 rng 種子隨 Turn 變化,見該函式頂部說明),逐一記錄
// AttackerShipsLost,斷言最大值 <= 2(絕不能 3 艘全滅)。
//
// [結構性保證,非僅靠校準數字幸運過關] homeworldBuildings() 開局只有「星基」一項軌道防禦
// 建築(無飛彈基地/戰鬥站/星辰要塞),retaliationAttackers 對「僅星基存活」只會回傳恰好 1 個
// attacker——battleVolley 對每個 attacker 只射一發,故單次 BombardColony 呼叫的反擊「結構上」
// 最多只能擊沉 1 艘玩家艦(不論 atk 校準值多少都不可能超過 attacker 數量),不會有「基地一輪
// 反擊團滅三艘船」的情境,這是本測試恆定成立的下限保證,與下方逐一實測互相印證。
func TestBombardColony_BalanceSanity_OpeningFleetVsHomeworldStarBase(t *testing.T) {
	profiles := []struct {
		name    string
		profile gamedata.RuleProfile
	}{
		{"Profile13", gamedata.Profile13()},
		{"Profile15", gamedata.Profile15()},
	}

	for _, p := range profiles {
		maxLost := 0
		for turn := 0; turn <= 14; turn++ {
			s, starIdx := newFleetAtAIHomeSession(t)
			s.RuleProfile = p.profile
			s.Turn = turn

			res := s.BombardColony(starIdx)
			if !res.Ok {
				t.Fatalf("[%s] Turn=%d:前置條件應齊備,got Reason=%q", p.name, turn, res.Reason)
			}
			if res.AttackerShipsLost > maxLost {
				maxLost = res.AttackerShipsLost
			}
			if res.AttackerShipsLost >= 3 {
				t.Errorf("[%s] Turn=%d:AttackerShipsLost=%d,開局三艘起始艦全滅(不可接受)",
					p.name, turn, res.AttackerShipsLost)
			}
		}
		t.Logf("[平衡 sanity] %s:開局艦隊轟炸開局 AI 母星(僅星基),Turn 0..14 掃描,"+
			"AttackerShipsLost 最大值 = %d(上限 2)", p.name, maxLost)
		if maxLost > 2 {
			t.Errorf("[%s] 最大 AttackerShipsLost = %d,want <= 2", p.name, maxLost)
		}
	}
}

// TestRetaliationAttackers_GroundBatterySupported 驗證地面砲台(#14 為完整性支援的分支)也能
// 正確算出 atk,且套用 GroundBatteryBeamArcCostPct(非 SatelliteBeamArcCostPct)。
func TestRetaliationAttackers_GroundBatterySupported(t *testing.T) {
	defender := engine.PlayerState{} // 雷射 fallback
	buildings := map[string]bool{"地面砲台": true}

	out13 := retaliationAttackers(buildings, defender, gamedata.Profile13())
	if len(out13) != 1 {
		t.Fatalf("預期恰有 1 個地面砲台 attacker,got len=%d", len(out13))
	}
	// 1.3 地面砲台 arc-cost=0%:perBeam=10,fit=450/10=45,atk=45*4/20=9。
	if out13[0].atk != 9 {
		t.Errorf("Profile13 地面砲台 atk = %d,want 9", out13[0].atk)
	}

	out15 := retaliationAttackers(buildings, defender, gamedata.Profile15())
	// 1.5 地面砲台 arc-cost=50%:perBeam=10+10*50/100=15,fit=450/15=30,atk=30*4/20=6。
	if out15[0].atk != 6 {
		t.Errorf("Profile15 地面砲台 atk = %d,want 6", out15[0].atk)
	}
}
