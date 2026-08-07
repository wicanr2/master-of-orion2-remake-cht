package shell

import (
	"reflect"
	"testing"
)

// hotseat_seat_test.go:席位快照的反射護欄。
//
// ============ 這支測試補的是一個「指名了不存在的護欄」的註解 ============
//
// `hotseat.go` 的 seat 型別上原本寫著「`TestSeatFieldsCoverPlayerSide` 用反射盯著它」,
// 而那支測試**根本不存在**(2026-08-07 加集結點欄位時發現)。指名了不存在的護欄比沒有註解
// 更危險——它讓人以為這裡有人在看,於是加欄位時不會特別小心。
//
// 這支測試是真的:漏抄任何一個 seat 欄位都會紅。

// TestSeatRoundTripKeepsEveryField:每個 seat 欄位都要被 saveSeat 存進去、被 loadSeat 裝回來。
//
// 作法:用反射把 seat 的每個欄位塞成一個**可辨識的非零值**,裝進 session,
// 再 saveSeat 抓回來比對。漏抄的欄位會停在零值,一比就露出來。
//
// 這正是熱座最容易出錯的地方——漏抄一個欄位,換人之後下一位玩家會繼承上一位的狀態,
// 而那看起來完全像是遊戲規則的一部分。
func TestSeatRoundTripKeepsEveryField(t *testing.T) {
	var want seat
	v := reflect.ValueOf(&want).Elem()
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		if !f.CanSet() {
			continue
		}
		fillDistinct(f, i+1)
	}

	// ⚠ SelectedFleet 不能用通用填值:它有**不變量**(必須落在 Fleets 範圍內,見 fleet.go
	// ensureFleet),隨便填一個大數會被夾回 0,測試就會誤報「沒往返」。
	// 給兩支艦隊 + 選第 1 支:既是合法值,又與零值不同,兩個條件都滿足。
	want.Fleets = []Fleet{{AtStar: 0, DestStar: -1}, {AtStar: 1, DestStar: -1}}
	want.SelectedFleet = 1

	s := &GameSession{}
	s.loadSeat(want)
	got := s.saveSeat()

	gv, wv := reflect.ValueOf(got), reflect.ValueOf(want)
	for i := 0; i < wv.NumField(); i++ {
		name := wv.Type().Field(i).Name
		if !reflect.DeepEqual(gv.Field(i).Interface(), wv.Field(i).Interface()) {
			t.Errorf("欄位 %s 沒有完整往返:存進去 %v,抓回來 %v\n"+
				"→ saveSeat / loadSeat 少抄了它。熱座換人後,下一位會繼承上一位的這個狀態。",
				name, wv.Field(i).Interface(), gv.Field(i).Interface())
		}
	}
}

// fillDistinct 依型別塞一個與零值不同、且各欄位互不相同的值。
func fillDistinct(f reflect.Value, n int) {
	switch f.Kind() {
	case reflect.Int:
		f.SetInt(int64(100 + n))
	case reflect.String:
		f.SetString("v" + string(rune('A'+n%26)))
	case reflect.Bool:
		f.SetBool(true)
	case reflect.Slice:
		f.Set(reflect.MakeSlice(f.Type(), 1, 1))
		fillDistinct(f.Index(0), n)
	case reflect.Ptr:
		f.Set(reflect.New(f.Type().Elem()))
	case reflect.Map:
		f.Set(reflect.MakeMap(f.Type()))
	case reflect.Struct:
		for i := 0; i < f.NumField(); i++ {
			if f.Field(i).CanSet() {
				fillDistinct(f.Field(i), n*31+i)
			}
		}
	default:
		// 其他型別(具名整數等)用整數路徑試,設不了就跳過——
		// 跳過的欄位這支測試蓋不到,那是誠實的覆蓋率缺口不是假綠。
		if f.CanInt() {
			f.SetInt(int64(100 + n))
		}
	}
}

// TestSeatRelocationIsPlayerSide:集結點是玩家側狀態,不能被下一位席位繼承。
//
// 這條單獨釘一次,因為它是 2026-08-07 新加的欄位,也是上面那支反射測試的第一個實例。
func TestSeatRelocationIsPlayerSide(t *testing.T) {
	s := &GameSession{Stars: make([]Star, 4)}
	s.ColonyRelocateTo = []int{2}
	a := s.saveSeat()

	s.ColonyRelocateTo = []int{3} // 換人期間被別的席位改掉
	s.loadSeat(a)
	if len(s.ColonyRelocateTo) != 1 || s.ColonyRelocateTo[0] != 2 {
		t.Errorf("換回原席位應拿回自己的集結點 [2],實得 %v", s.ColonyRelocateTo)
	}
}
