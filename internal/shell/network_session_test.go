package shell

import "testing"

func TestNetworkSnapshotReplayAndUIHashNormalization(t *testing.T) {
	left := NewDemoSession()
	right := NewDemoSession()
	names := []string{"主機", "客戶端"}
	if got := left.SetupNetworkWithNames(names); got != len(names) {
		t.Fatalf("主機席位數錯誤：%d", got)
	}
	if got := right.SetupNetworkWithNames(names); got != len(names) {
		t.Fatalf("客戶端席位數錯誤：%d", got)
	}
	base, err := left.MarshalSnapshot()
	if err != nil {
		t.Fatal(err)
	}
	left, err = RestoreSnapshot(base)
	if err != nil {
		t.Fatal(err)
	}
	right, err = RestoreSnapshot(base)
	if err != nil {
		t.Fatal(err)
	}
	left.SelectedStar, left.SelectedFleet, left.ShowRelocationLines = 3, 1, true
	right.SelectedStar, right.SelectedFleet, right.ShowRelocationLines = 7, 0, false
	if left.NetworkStateHash() != right.NetworkStateHash() {
		t.Fatal("純 UI 選取差異不應造成網路狀態分歧")
	}
	if err := left.ApplyPlayerCommandsForSeat(0, []PlayerCommand{{Name: CmdCycleTaxRate}}); err != nil {
		t.Fatal(err)
	}
	if err := left.ApplyPlayerCommandsForSeat(1, []PlayerCommand{{Name: CmdCycleTaxRate}}); err != nil {
		t.Fatal(err)
	}
	if err := right.ApplyPlayerCommandsForSeat(0, []PlayerCommand{{Name: CmdCycleTaxRate}}); err != nil {
		t.Fatal(err)
	}
	if err := right.ApplyPlayerCommandsForSeat(1, []PlayerCommand{{Name: CmdCycleTaxRate}}); err != nil {
		t.Fatal(err)
	}
	if left.NetworkStateHash() != right.NetworkStateHash() {
		t.Fatal("相同席位順序重播後狀態應一致")
	}
}
