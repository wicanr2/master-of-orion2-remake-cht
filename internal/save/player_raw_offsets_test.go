package save

import "testing"

func TestPlayerLoadPreservesRaw60EWithoutChangingStride(t *testing.T) {
	data := make([]byte, 3753)
	data[0x60E] = 1
	r := newReader(data)
	var p Player
	p.load(r)
	if p.Raw60E != 1 {
		t.Fatalf("Player+0x60E=%d，want 1", p.Raw60E)
	}
	if r.at() != 3753 {
		t.Fatalf("Player stride=%d，want 3753", r.at())
	}
}

func TestPlanetLoadUsesSeventeenByteStride(t *testing.T) {
	data := make([]byte, 17)
	data[16] = 0xA5
	r := newReader(data)
	var p Planet
	p.load(r)
	if p.Flags != 0xA5 || r.at() != 17 {
		t.Fatalf("Planet flags/stride=%#x/%d，want 0xa5/17", p.Flags, r.at())
	}
}
