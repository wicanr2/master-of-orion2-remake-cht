package main

import "testing"

func TestShipWireOffsets(t *testing.T) {
	if shipRecordsOffset+500*shipWireSize != 203596 {
		t.Fatalf("ship 陣列結尾與 SAVE10 SeqEnd 不符")
	}
	if shipStarOffset != 101 {
		t.Fatalf("Ship.Star wire offset 漂移：%d", shipStarOffset)
	}
	if strategicModeOffset != 262 {
		t.Fatalf("SAVE 戰鬥模式 offset 漂移：%d", strategicModeOffset)
	}
}
