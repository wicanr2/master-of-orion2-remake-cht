package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func originalPowerTestShip() Ship {
	return Ship{
		Class: playerShipDesignClasses[0], ComputerRaw: 1, ComputerRawKnown: true,
		DesignSizeRaw: 0, DesignSizeRawKnown: true, ArmorRaw: 1, ArmorRawKnown: true,
		CrewLevelRaw: 0, CrewLevelRawKnown: true, OriginalDamageKnown: true,
		WeaponMounts: []ShipWeaponMount{{RawType: 3, WorkingCount: 100, Ammo: 255}},
	}
}

func TestOriginalNPCDirectionalFleetPowerUsesObserverTechnology(t *testing.T) {
	s := &GameSession{AIPlayers: []AIOpponent{
		{Ships: []Ship{originalPowerTestShip()}},
		{Player: engine.PlayerState{}},
	}}
	plain, ok := s.originalNPCDirectionalFleetPower(0, 1)
	if !ok || plain <= 0 {
		t.Fatalf("基準方向國力=%d,%v", plain, ok)
	}
	s.AIPlayers[1].Player.GrantedTechs = map[gamedata.Technology]bool{gamedata.TECH_MOLECULARTRONIC_COMPUTER: true}
	reduced, ok := s.originalNPCDirectionalFleetPower(0, 1)
	if !ok || reduced >= plain {
		t.Fatalf("觀察者最高電腦應降低對手方向國力：plain=%d reduced=%d ok=%v", plain, reduced, ok)
	}
}

func TestOriginalAIPowerMatrixMarksLegacyScalarFallback(t *testing.T) {
	s := &GameSession{AIPlayers: []AIOpponent{{FleetStrength: 77}, {}}}
	power, exact := s.originalAIPowerMatrix()
	if power[0][1] != 77 || exact[0][1] {
		t.Fatalf("舊存檔純量回退=%d exact=%v，預期 77,false", power[0][1], exact[0][1])
	}
	if power[1][0] != 0 || !exact[1][0] {
		t.Fatalf("真正空艦隊應是精確零：%d exact=%v", power[1][0], exact[1][0])
	}
}

func TestOriginalAIHumanDirectionalFleetPowerUsesBothObservers(t *testing.T) {
	s := &GameSession{
		Player:    engine.PlayerState{},
		Fleets:    []Fleet{{Ships: []Ship{originalPowerTestShip()}}},
		AIPlayers: []AIOpponent{{Ships: []Ship{originalPowerTestShip()}}},
	}
	aiToHuman, humanToAI, ok := s.originalAIHumanDirectionalFleetPower(0)
	if !ok || aiToHuman <= 0 || humanToAI <= 0 {
		t.Fatalf("雙向真人國力=%d/%d ok=%v", aiToHuman, humanToAI, ok)
	}
	s.Player.GrantedTechs = map[gamedata.Technology]bool{gamedata.TECH_MOLECULARTRONIC_COMPUTER: true}
	reducedAI, improvedHuman, ok := s.originalAIHumanDirectionalFleetPower(0)
	if !ok || reducedAI >= aiToHuman || improvedHuman <= 0 {
		t.Fatalf("真人觀察科技應降低 AI→真人方向：before=%d after=%d human=%d ok=%v",
			aiToHuman, reducedAI, improvedHuman, ok)
	}
}
