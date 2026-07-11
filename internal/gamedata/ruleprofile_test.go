package gamedata

import "testing"

// TestProfile13Values 驗證 patch 1.3 規則 profile 的三個確證數值(見
// docs/tech/version-1.3-1.5-diff.md §2)。
func TestProfile13Values(t *testing.T) {
	p := Profile13()
	if p.Version != VersionClassic13 {
		t.Errorf("Version = %v,want VersionClassic13", p.Version)
	}
	if p.HyperAdvancedLevel1Cost != 15000 {
		t.Errorf("HyperAdvancedLevel1Cost = %d,want 15000", p.HyperAdvancedLevel1Cost)
	}
	if p.PlasmaCannonMaxDamage != 30 {
		t.Errorf("PlasmaCannonMaxDamage = %d,want 30", p.PlasmaCannonMaxDamage)
	}
	if p.BombardmentVolleys != 5 {
		t.Errorf("BombardmentVolleys = %d,want 5", p.BombardmentVolleys)
	}
}

// TestProfile15Values 驗證 patch 1.5 規則 profile 的三個值,並確認等於本專案改用 profile 前
// 的現行硬編值(techtree.go 25000 / session.go 20 / orbital_bombardment.go 10)——這是
// 「預設 Profile15 = 現行值,本次重構在預設路徑上是 no-op」的核心回歸斷言。
func TestProfile15Values(t *testing.T) {
	p := Profile15()
	if p.Version != VersionCommunity15 {
		t.Errorf("Version = %v,want VersionCommunity15", p.Version)
	}

	const wantHyperCost = 25000       // techtree.go 8 個 TOPIC_HYPER_* 的硬編 Cost
	const wantPlasmaDamage = 20       // session.go 電漿砲 Component.Value 硬編值
	const wantBombardmentVolleys = 10 // orbital_bombardment.go for round<10 的硬編上限

	if p.HyperAdvancedLevel1Cost != wantHyperCost {
		t.Errorf("HyperAdvancedLevel1Cost = %d,want %d(=techtree.go 現行硬編值)", p.HyperAdvancedLevel1Cost, wantHyperCost)
	}
	if p.PlasmaCannonMaxDamage != wantPlasmaDamage {
		t.Errorf("PlasmaCannonMaxDamage = %d,want %d(=session.go 現行硬編值)", p.PlasmaCannonMaxDamage, wantPlasmaDamage)
	}
	if p.BombardmentVolleys != wantBombardmentVolleys {
		t.Errorf("BombardmentVolleys = %d,want %d(=orbital_bombardment.go 現行硬編值)", p.BombardmentVolleys, wantBombardmentVolleys)
	}

	// 交叉核對:techtree.go 8 條 TOPIC_HYPER_* 的表定值本身也必須等於 wantHyperCost,
	// 證明「Profile15 = 現行值」不是巧合湊出來的,是真的讀同一個既有事實。
	hyperTopics := []ResearchTopic{
		TOPIC_HYPER_BIOLOGY, TOPIC_HYPER_POWER, TOPIC_HYPER_PHYSICS, TOPIC_HYPER_CONSTRUCTION,
		TOPIC_HYPER_FIELDS, TOPIC_HYPER_CHEMISTRY, TOPIC_HYPER_COMPUTERS, TOPIC_HYPER_SOCIOLOGY,
	}
	for _, topic := range hyperTopics {
		if got := ResearchChoiceFor(topic).Cost; got != wantHyperCost {
			t.Errorf("ResearchChoiceFor(%v).Cost = %d,want %d", topic, got, wantHyperCost)
		}
		if !IsHyperAdvancedTopic(topic) {
			t.Errorf("IsHyperAdvancedTopic(%v) = false,want true", topic)
		}
	}
}

// TestHyperAdvancedCost 驗證查詢時覆寫函式對兩個 profile 各自回傳正確值。
func TestHyperAdvancedCost(t *testing.T) {
	if got := HyperAdvancedCost(Profile13()); got != 15000 {
		t.Errorf("HyperAdvancedCost(Profile13()) = %d,want 15000", got)
	}
	if got := HyperAdvancedCost(Profile15()); got != 25000 {
		t.Errorf("HyperAdvancedCost(Profile15()) = %d,want 25000", got)
	}
}

// TestIsHyperAdvancedTopic_NonHyperTopicFalse 驗證非 HYPER 主題不會被誤判。
func TestIsHyperAdvancedTopic_NonHyperTopicFalse(t *testing.T) {
	nonHyper := []ResearchTopic{TOPIC_PHYSICS, TOPIC_CHEMISTRY, TOPIC_STARTING_TECH, TOPIC_PLASMA_PHYSICS}
	for _, topic := range nonHyper {
		if IsHyperAdvancedTopic(topic) {
			t.Errorf("IsHyperAdvancedTopic(%v) = true,want false", topic)
		}
	}
}
