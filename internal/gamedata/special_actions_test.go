package gamedata

import "testing"

func TestSpecialActionByNameZH(t *testing.T) {
	if a, ok := SpecialActionByNameZH(TerraformActionName); !ok || a.PrereqTopic != TOPIC_GENETIC_MUTATIONS {
		t.Errorf("地形改造前置科技應為 TOPIC_GENETIC_MUTATIONS,got %+v ok=%v", a, ok)
	}
	if a, ok := SpecialActionByNameZH(GaiaTransformationActionName); !ok || a.PrereqTopic != TOPIC_TRANS_GENETICS {
		t.Errorf("蓋亞轉化前置科技應為 TOPIC_TRANS_GENETICS,got %+v ok=%v", a, ok)
	}
	if a, ok := SpecialActionByNameZH(SoilEnrichmentActionName); !ok || a.PrereqTopic != TOPIC_ADVANCED_BIOLOGY {
		t.Errorf("土壤改良前置科技應為 TOPIC_ADVANCED_BIOLOGY,got %+v ok=%v", a, ok)
	}
	if a, ok := SpecialActionByNameZH(FreighterFleetActionName); !ok || a.PrereqTopic != TOPIC_NUCLEAR_FISSION {
		t.Errorf("運輸艦隊前置科技應為 TOPIC_NUCLEAR_FISSION,got %+v ok=%v", a, ok)
	}
	if _, ok := SpecialActionByNameZH("不存在的行動"); ok {
		t.Error("不存在的名稱應回傳 ok=false")
	}
}

func TestAvailableSpecialActions(t *testing.T) {
	if got := AvailableSpecialActions(nil); len(got) != 0 {
		t.Errorf("nil completedTopics 應回傳空清單,got %d 項", len(got))
	}
	completed := map[ResearchTopic]bool{TOPIC_ADVANCED_BIOLOGY: true}
	got := AvailableSpecialActions(completed)
	if len(got) != 1 || got[0].NameZH != SoilEnrichmentActionName {
		t.Errorf("只解鎖 TOPIC_ADVANCED_BIOLOGY 時應只看到土壤改良,got %+v", got)
	}

	completed[TOPIC_NUCLEAR_FISSION] = true
	got = AvailableSpecialActions(completed)
	if len(got) != 2 {
		t.Errorf("再解鎖 TOPIC_NUCLEAR_FISSION 後應看到土壤改良+運輸艦隊共 2 項,got %+v", got)
	}
	found := false
	for _, a := range got {
		if a.NameZH == FreighterFleetActionName {
			found = true
		}
	}
	if !found {
		t.Errorf("解鎖 TOPIC_NUCLEAR_FISSION 後應看到運輸艦隊,got %+v", got)
	}
}
