package audio

// dump_test.go:用 stub archive 驗證 RawMusic/RawSounds 回傳的原始 WAV bytes、
// 筆數與名稱對應正確(不需真實遊戲資料)。

import "testing"

func TestRawMusic(t *testing.T) {
	wavFor := func(seed byte) []byte {
		return buildWAV(22050, 1, 8, []byte{seed, seed + 1, seed + 2})
	}

	arc := &stubArchive{assets: [][]byte{
		[]byte("not a wav, egg-slot string"), // entry 0:非 WAV,應略過
		wavFor(10),                           // entry 1
		wavFor(20),                           // entry 2
	}}

	got, err := RawMusic(arc)
	if err != nil {
		t.Fatalf("RawMusic 失敗: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("RawMusic 回傳 %d 筆, want 2", len(got))
	}
	if got[0].Index != 1 || got[1].Index != 2 {
		t.Errorf("RawMusic index = [%d,%d], want [1,2]", got[0].Index, got[1].Index)
	}
	for _, rw := range got {
		if rw.Name != "" {
			t.Errorf("RawMusic entry %d Name = %q, want 空字串", rw.Index, rw.Name)
		}
		if len(rw.WAV) == 0 {
			t.Errorf("RawMusic entry %d WAV 為空", rw.Index)
		}
		// 原封 bytes:必須是完整 RIFF/WAVE,且與 ExtractWAV 直接切出的結果一致。
		raw, _ := arc.Asset(rw.Index)
		want, ok := ExtractWAV(raw)
		if !ok {
			t.Fatalf("entry %d 預期可 ExtractWAV", rw.Index)
		}
		if string(rw.WAV) != string(want) {
			t.Errorf("RawMusic entry %d WAV bytes 與 ExtractWAV 不一致", rw.Index)
		}
	}
}

func TestRawMusic_AllEmpty(t *testing.T) {
	arc := &stubArchive{assets: [][]byte{[]byte("no wav here")}}
	got, err := RawMusic(arc)
	if err != nil {
		t.Fatalf("RawMusic 失敗: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("RawMusic = %d 筆, want 0", len(got))
	}
}

func TestRawSounds(t *testing.T) {
	wavFor := func(seed byte) []byte {
		return buildWAV(22050, 1, 8, []byte{seed, seed + 1, seed + 2})
	}

	var nameTable []byte
	nameTable = append(nameTable, nameRec("BUTTON1")...)
	nameTable = append(nameTable, nameRec("BUTTON2")...)
	nameTable = append(nameTable, nameRec("NEW.....")...) // 終止標記之後不應再對到名稱

	arc := &stubArchive{assets: [][]byte{
		nameTable,  // entry 0
		wavFor(10), // entry 1 -> BUTTON1
		wavFor(20), // entry 2 -> BUTTON2
		wavFor(30), // entry 3:NEW 標記之後,無名稱可對應,但仍應抽出(Name="")
	}}

	got, err := RawSounds(arc)
	if err != nil {
		t.Fatalf("RawSounds 失敗: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("RawSounds 回傳 %d 筆, want 3", len(got))
	}

	want := map[int]string{1: "BUTTON1", 2: "BUTTON2", 3: ""}
	for _, rw := range got {
		if rw.Name != want[rw.Index] {
			t.Errorf("RawSounds entry %d Name = %q, want %q", rw.Index, rw.Name, want[rw.Index])
		}
		if len(rw.WAV) == 0 {
			t.Errorf("RawSounds entry %d WAV 為空", rw.Index)
		}
	}
}

func TestRawSounds_TooFewEntries(t *testing.T) {
	arc := &stubArchive{assets: [][]byte{}}
	if _, err := RawSounds(arc); err == nil {
		t.Error("RawSounds(空 archive) 應回傳錯誤")
	}
}
