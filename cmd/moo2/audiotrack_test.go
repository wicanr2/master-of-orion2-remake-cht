package main

import "testing"

// audiotrack_test.go:釘住「場景 → 曲目編號」這張表,以及編號空間的分界。
//
// 這張表的來源是反組譯的立即數(第 73 項(音樂場景表)),不是聆聽推定,所以它應該被當成
// 硬資料保護——改動要先回頭看 docs/tech/audio-track-map.md,不是憑感覺調。

// 實測的 entry id(2026-08-08,玩家自備資料夾,`LoadMusic` 回傳值):
//
//	stream.lbx   count=11 可用=8  entryIDs=[1 2 3 4 5 6 8 10]
//	streamhd.lbx count=21 可用=20 entryIDs=[1..20]
//
// ⚠ **這組數字本身就是一道交叉驗證**:反組譯點名的 STREAM 曲目是 1/2/3(背景)、
// 4/5/6(戰鬥)、8(艦艇設計)、10(殖民地戰鬥)——**正好就是這八個存在的槽**,
// 而它從來沒點過的 7 與 9,正是那兩個非 WAV 的空槽。
// 若編號是「第幾條 WAV」而不是 entry id,這個巧合不會成立。
var (
	streamEntryIDs   = map[int]bool{1: true, 2: true, 3: true, 4: true, 5: true, 6: true, 8: true, 10: true}
	streamHDEntryIDs = func() map[int]bool {
		m := map[int]bool{}
		for i := 1; i <= 20; i++ {
			m[i] = true
		}
		return m
	}()
)

// trackExists 依原版的單一編號空間判斷這個編號在玩家資料夾裡有沒有對應的曲子。
func trackExists(track int) bool {
	if track > 100 {
		return streamHDEntryIDs[track-100]
	}
	return streamEntryIDs[track]
}

func TestSceneTracksExist(t *testing.T) {
	cases := []struct {
		name  string
		track int
	}{
		{"科學室", trackScienceRoom},
		{"事件快報", trackEventScreen},
		{"議會", trackCouncil},
		{"安塔蘭廳", trackAntaranRoom},
		{"艦艇設計", trackShipDesign},
		{"殖民地戰鬥", trackColonyCombat},
	}
	for _, c := range cases {
		if !trackExists(c.track) {
			t.Errorf("%s 的曲目編號 %d 在玩家資料夾裡沒有對應 entry", c.name, c.track)
		}
	}
	for _, tr := range backgroundMusicTracks {
		if !trackExists(tr) {
			t.Errorf("背景音樂編號 %d 沒有對應 entry", tr)
		}
	}
	for _, tr := range combatMusicTracks {
		if !trackExists(tr) {
			t.Errorf("戰鬥音樂編號 %d 沒有對應 entry", tr)
		}
	}
	for _, tr := range bgmDiploBadPool {
		if !trackExists(tr) {
			t.Errorf("外交音樂編號 %d 沒有對應 entry", tr)
		}
	}
}

// TestTrackNumberSpaceSplit 編號空間的分界:≤100 走 STREAM、>100 走 STREAMHD(索引 −100)。
//
// 邊界值 100 屬於 STREAM 那一側(反組譯的判斷是 `> 100` 才切換),而 STREAM 只有 11 個
// entry,所以 100 本身不會對應到任何曲子——這條測的是**分界怎麼算**,不是 100 有沒有曲。
func TestTrackNumberSpaceSplit(t *testing.T) {
	// 107 要解讀成 STREAMHD #7(存在),不是 STREAM #7(空槽)——這一對就是分界本身。
	if !trackExists(107) {
		t.Error("107 應該解讀成 STREAMHD #7 而查得到")
	}
	if trackExists(7) {
		t.Error("STREAM #7 是非 WAV 空槽,不該查得到")
	}
	if !trackExists(8) {
		t.Error("STREAM #8(艦艇設計)存在")
	}
}

// TestBackgroundAndCombatMusicAreDisjoint 手冊與反組譯都把戰鬥音樂寫成**獨立**的一組
// (Play_Combat_Music_ 是自己的函式,不是背景樂的延續)。兩組不該重疊。
func TestBackgroundAndCombatMusicAreDisjoint(t *testing.T) {
	bg := map[int]bool{}
	for _, tr := range backgroundMusicTracks {
		bg[tr] = true
	}
	for _, tr := range combatMusicTracks {
		if bg[tr] {
			t.Errorf("曲目 %d 同時出現在背景與戰鬥兩組", tr)
		}
	}
}
