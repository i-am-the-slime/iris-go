package refcell

import "testing"

func TestNewReadWrite(t *testing.T) {
	cell := New(1)()
	Write(5, cell)()
	if got := Read(cell)(); got != 5 {
		t.Fatalf("read after write = %d, want 5", got)
	}
	if got := WriteST(7, cell)(); got != 7 || cell.Value != 7 {
		t.Fatalf("WriteST = %d (cell %d), want 7", got, cell.Value)
	}
}

func TestModifyImplReturnsValueAndStoresState(t *testing.T) {
	cell := New(10)()
	got := ModifyImpl(func(s int) struct {
		State int
		Value string
	} {
		return struct {
			State int
			Value string
		}{State: s + 1, Value: "old"}
	}, cell)()
	if got != "old" || cell.Value != 11 {
		t.Fatalf("modify = %q, cell %d", got, cell.Value)
	}
}

func TestNewWithSelfSeesItsOwnCell(t *testing.T) {
	var seen *struct{ Value int }
	cell := NewWithSelf(func(self *struct{ Value int }) int {
		seen = self
		return 3
	})()
	if seen != cell || cell.Value != 3 {
		t.Fatalf("newWithSelf did not pass its own cell")
	}
}
