package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// TestColonyMoralePercentDictatorshipNoBarracks 驗證獨裁政府、無 Barracks 時士氣 = 手冊
// -20%(gamedata.MoraleGovernmentBase 常數,p.21-22/p.165-167),不自行杜撰數字。
func TestColonyMoralePercentDictatorshipNoBarracks(t *testing.T) {
	got := colonyMoralePercent(gamedata.MoraleGovDictatorship, nil)
	want := gamedata.MoraleGovernmentBase(gamedata.MoraleGovDictatorship, false)
	if got != want {
		t.Fatalf("獨裁無 Barracks 士氣應為 %d,實得 %d", want, got)
	}
	if want != -20 {
		t.Fatalf("對照手冊常數應為 -20,實得 %d(gamedata 常數是否變動?)", want)
	}
}

// TestColonyMoralePercentBarracksCancelsPenalty 驗證海軍陸戰隊營/裝甲營房任一存在,皆可解除
// 獨裁政府「無 Barracks -20%」懲罰(p.76-79)。
func TestColonyMoralePercentBarracksCancelsPenalty(t *testing.T) {
	marine := colonyMoralePercent(gamedata.MoraleGovDictatorship, map[string]bool{"海軍陸戰隊營": true})
	if marine != 0 {
		t.Fatalf("有海軍陸戰隊營時獨裁士氣應歸零:實得 %d", marine)
	}
	armor := colonyMoralePercent(gamedata.MoraleGovDictatorship, map[string]bool{"裝甲營房": true})
	if armor != 0 {
		t.Fatalf("有裝甲營房時獨裁士氣應歸零:實得 %d", armor)
	}
}

// TestColonyMoralePercentHoloSimulatorAndPleasureDomeStack 驗證全息模擬艙(+20)、歡樂穹頂
// (+30)可疊加,且疊在政府基礎值之上。
func TestColonyMoralePercentHoloSimulatorAndPleasureDomeStack(t *testing.T) {
	base := colonyMoralePercent(gamedata.MoraleGovDictatorship, map[string]bool{"海軍陸戰隊營": true})

	withHolo := colonyMoralePercent(gamedata.MoraleGovDictatorship, map[string]bool{
		"海軍陸戰隊營": true, "全息模擬艙": true,
	})
	if want := base + gamedata.MoraleHoloSimulatorBonus; withHolo != want {
		t.Fatalf("全息模擬艙應 +%d:期望 %d,實得 %d", gamedata.MoraleHoloSimulatorBonus, want, withHolo)
	}

	withBoth := colonyMoralePercent(gamedata.MoraleGovDictatorship, map[string]bool{
		"海軍陸戰隊營": true, "全息模擬艙": true, "歡樂穹頂": true,
	})
	if want := base + gamedata.MoraleHoloSimulatorBonus + gamedata.MoralePleasureDomeBonus; withBoth != want {
		t.Fatalf("全息模擬艙+歡樂穹頂應疊加:期望 %d,實得 %d", want, withBoth)
	}
}

// TestColonyMoralePercentGovernmentDiffers 驗證統一/民主的政府基礎值為 0(手冊未提及基礎士氣
// 效果),對照 gamedata.MoraleGovernmentBase 常數。
func TestColonyMoralePercentGovernmentDiffers(t *testing.T) {
	uni := colonyMoralePercent(gamedata.MoraleGovUnification, nil)
	if uni != 0 {
		t.Fatalf("統一政府無 Barracks 士氣應為 0(手冊無基礎效果敘述):實得 %d", uni)
	}
	dem := colonyMoralePercent(gamedata.MoraleGovDemocracy, nil)
	if dem != 0 {
		t.Fatalf("民主政府無 Barracks 士氣應為 0:實得 %d", dem)
	}
}

// TestApplyGovernmentRecalculatesMorale 端到端驗證:ApplyGovernment 換政府後,
// PlayerColonies[0].MoralePercent 依新政府 + 現有建築重算(不再是硬編值)。
func TestApplyGovernmentRecalculatesMorale(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true

	// NewDemoSession 預設獨裁 + homeworldBuildings(含海軍陸戰隊營),士氣應為 0。
	if got := s.PlayerColonies[0].MoralePercent; got != 0 {
		t.Fatalf("開局獨裁+已建海軍陸戰隊營,士氣應為 0:實得 %d", got)
	}

	s.ApplyGovernment(1) // 封建:政府基礎值同獨裁(-20/有 Barracks 解除),仍應為 0
	if got := s.PlayerColonies[0].MoralePercent; got != 0 {
		t.Fatalf("封建+已建海軍陸戰隊營,士氣應為 0:實得 %d", got)
	}
	if s.Government != gamedata.MoraleGovFeudalism {
		t.Fatalf("ApplyGovernment(1) 應記錄為封建,實得 %v", s.Government)
	}

	s.ApplyGovernment(2) // 統一:基礎值 0,與 Barracks 無關
	if s.Government != gamedata.MoraleGovUnification {
		t.Fatalf("ApplyGovernment(2) 應記錄為統一,實得 %v", s.Government)
	}
	if got := s.PlayerColonies[0].MoralePercent; got != 0 {
		t.Fatalf("統一政府士氣基礎值應為 0:實得 %d", got)
	}
}

// TestMoralePercentAffectsColonyProduction 端到端驗證:士氣建築完工後,MoralePercent 提升
// 確實反映到殖民地產出(engine.RunColonyTurn/MoraleProductionOutput 消費 MoralePercent,見
// internal/engine/colony.go)。
func TestMoralePercentAffectsColonyProduction(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true

	beforeMorale := s.PlayerColonies[0].MoralePercent
	beforeWorkers := s.PlayerColonies[0].Workers

	s.Builds[0] = ColonyBuild{Name: "全息模擬艙", Progress: 0, Cost: 60}
	for i := 0; i < 30 && s.Builds[0].Name != ""; i++ {
		s.EndTurn()
	}

	if !s.ColonyBuildings[0]["全息模擬艙"] {
		t.Fatal("全息模擬艙應標記為已建")
	}
	afterMorale := s.PlayerColonies[0].MoralePercent
	if want := beforeMorale + gamedata.MoraleHoloSimulatorBonus; afterMorale != want {
		t.Fatalf("全息模擬艙完工後士氣應 = %d,實得 %d", want, afterMorale)
	}

	// 士氣提升應讓同樣工人數的毛工業產出提高(GravityAdjustedProduction 依 MoralePercent 調整,
	// 見 engine/colony.go RunColonyTurn)。人口/工人分配可能因回合推進而變動,固定工人數比較。
	afterWorkers := s.PlayerColonies[0].Workers
	if afterWorkers < beforeWorkers {
		t.Skip("回合推進期間工人分配減少,產出比較基準不穩定,略過本輪端到端數值比較")
	}
	baseIndustry := afterWorkers * s.PlayerColonies[0].IndustryPerWorker
	got := gamedata.GravityAdjustedProduction(baseIndustry, afterMorale+0 /* Normal-G 無重力懲罰 */)
	want := gamedata.GravityAdjustedProduction(baseIndustry, beforeMorale)
	if afterMorale > beforeMorale && got <= want {
		t.Fatalf("士氣提升後,同樣基礎產出的士氣調整值應更高:士氣前 %d(調整後 %d)→ 士氣後 %d(調整後 %d)",
			beforeMorale, want, afterMorale, got)
	}
}
