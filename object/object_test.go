package object

import (
	"reflect"
	"slices"
	"testing"
)

func build(entries ...string) *struct {
	Index map[string]int
	Ints  int
	Keys  []string
	Vals  []string
} {
	m := New[string]()()
	for _, k := range entries {
		Poke(k, k+"!", m)()
	}
	return m
}

func TestKeysEnumerateInInsertionOrderEvenWhenIntegerLike(t *testing.T) {
	m := build("b", "10", "a", "2", "01", "4294967295", "4294967294")
	got := Keys(m)
	want := []string{"b", "10", "a", "2", "01", "4294967295", "4294967294"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Keys = %v, want %v", got, want)
	}
}

func TestPokeKeepsPositionAndDeleteReindexes(t *testing.T) {
	m := build("x", "y", "z")
	Poke("x", "again", m)()
	Delete("y", m)()
	Poke("y", "back", m)()
	if got := Keys(m); !reflect.DeepEqual(got, []string{"x", "z", "y"}) {
		t.Fatalf("Keys = %v", got)
	}
	if got := Lookup("none", func(v string) string { return v }, "x", m); got != "again" {
		t.Fatalf("Lookup x = %q", got)
	}
	if got := Lookup("none", func(v string) string { return v }, "z", m); got != "z!" {
		t.Fatalf("Lookup z = %q", got)
	}
}

func TestCopyIsIndependentOfItsSource(t *testing.T) {
	m := build("a")
	c := CopyST(m)()
	Poke("b", "b!", c)()
	if Size(m) != 1 || Size(c) != 2 {
		t.Fatalf("sizes %d %d", Size(m), Size(c))
	}
}

func TestEmptyAcceptsReadsAndWrites(t *testing.T) {
	e := Empty[int]()
	if Size(e) != 0 || len(Keys(e)) != 0 {
		t.Fatal("empty is not empty")
	}
	Poke("k", 1, CopyST(e)())()
}

func TestFromEntriesKeepsAssignmentOrder(t *testing.T) {
	m := FromEntries([]string{"zeta", "1", "alpha"}, []int{1, 2, 3})
	if got := Keys(m); !reflect.DeepEqual(got, []string{"zeta", "1", "alpha"}) {
		t.Fatalf("Keys = %v", got)
	}
	if got := Lookup(0, func(v int) int { return v }, "alpha", m); got != 3 {
		t.Fatalf("Lookup alpha = %d", got)
	}
}

func TestTablesFindKeysBeforeAndAfterTheyOutgrowScanning(t *testing.T) {
	m := New[int]()()
	keys := []string{}
	for i := range 3 * scanLimit {
		key := string(rune('a' + i))
		keys = append(keys, key)
		Poke(key, i, m)()
		if (m.Index != nil) != (len(m.Keys) > scanLimit) {
			t.Fatalf("with %d keys Index is %v", len(m.Keys), m.Index)
		}
		for j, earlier := range keys {
			if got := Lookup(-1, func(v int) int { return v }, earlier, m); got != j {
				t.Fatalf("after %d pokes %s = %d", i+1, earlier, got)
			}
		}
	}
	if got := Keys(m); !reflect.DeepEqual(got, keys) {
		t.Fatalf("Keys = %v", got)
	}
}

func TestScannedTablesDeleteAndCopy(t *testing.T) {
	m := FromEntries([]string{"x", "y", "z", "y"}, []int{1, 2, 3, 4})
	if m.Index != nil || !reflect.DeepEqual(m.Keys, []string{"x", "y", "z"}) || UnsafeIndex(m, "y") != 4 {
		t.Fatalf("FromEntries = %+v", *m)
	}
	c := CopyST(m)()
	Delete("x", c)()
	if _, found := Slot(c, "x"); found || UnsafeIndex(c, "z") != 3 || UnsafeIndex(m, "x") != 1 {
		t.Fatalf("copy %+v, source %+v", *c, *m)
	}
}

func entries(m *struct {
	Index map[string]int
	Ints  int
	Keys  []string
	Vals  []string
}) []string {
	return ToArrayWithKey(func(k, v string) string { return k + "=" + v }, m)
}

func TestInsertAndRemoveLeaveTheirSourceUnchanged(t *testing.T) {
	small := build("a", "b")
	large := build("a", "b", "c", "d", "e", "f", "g", "h", "i", "j")
	for _, m := range []*struct {
		Index map[string]int
		Ints  int
		Keys  []string
		Vals  []string
	}{small, large} {
		before := entries(m)
		replaced := Insert("a", "new", m)
		added := Insert("z", "z!", m)
		removed := Remove("b", m)
		if got := entries(m); !reflect.DeepEqual(got, before) {
			t.Fatalf("source changed to %v", got)
		}
		if got := entries(replaced)[0]; got != "a=new" {
			t.Fatalf("replaced first entry = %v", got)
		}
		if got := Keys(added); got[len(got)-1] != "z" || len(got) != len(before)+1 {
			t.Fatalf("added keys = %v", got)
		}
		if got := Keys(removed); len(got) != len(before)-1 || slices.Contains(got, "b") {
			t.Fatalf("removed keys = %v", got)
		}
		if Remove("missing", m) != m {
			t.Fatal("removing a missing key copied the object")
		}
	}
}

func TestOwnedWritesToALentObjectLeaveTheLenderUnchanged(t *testing.T) {
	lender := build("a", "b", "c", "d", "e", "f", "g", "h", "i", "j")
	mapped := FmapObject(lender, func(v string) string { return v + "?" })
	before := entries(lender)
	borrowed := Lend(mapped)
	for _, k := range []string{"a", "k", "l"} {
		borrowed = InsertOwned(k, k+"*", borrowed)
	}
	borrowed = RemoveOwned("b", borrowed)
	if got := entries(lender); !reflect.DeepEqual(got, before) {
		t.Fatalf("lender changed to %v", got)
	}
	if got := Keys(mapped); len(got) != 10 || got[1] != "b" {
		t.Fatalf("mapped keys changed to %v", got)
	}
	want := []string{"a=a*", "c=c!?", "d=d!?", "e=e!?", "f=f!?", "g=g!?", "h=h!?", "i=i!?", "j=j!?", "k=k*", "l=l*"}
	if got := entries(borrowed); !reflect.DeepEqual(got, want) {
		t.Fatalf("borrowed = %v", got)
	}
}

func TestOwnedInsertsGrowInLinearTime(t *testing.T) {
	m := Lend(Empty[int]())
	keys := make([]string, 20000)
	for i := range keys {
		keys[i] = string(rune('a'+i%26)) + string(rune(i))
	}
	allocations := testing.AllocsPerRun(1, func() {
		m = Lend(Empty[int]())
		for i, k := range keys {
			m = InsertOwned(k, i, m)
		}
	})
	if allocations > 200 {
		t.Fatalf("%v allocations for %d inserts", allocations, len(keys))
	}
	if int(Size(m)) != len(keys) {
		t.Fatalf("size = %d", Size(m))
	}
}

func BenchmarkInsertOwned(b *testing.B) {
	keys := make([]string, 10000)
	for i := range keys {
		keys[i] = string(rune('a'+i%26)) + string(rune(i))
	}
	for b.Loop() {
		m := Lend(Empty[int]())
		for i, k := range keys {
			m = InsertOwned(k, i, m)
		}
	}
}

func TestLendingTheEmptySlotOfAnotherConstructorKeepsItEmpty(t *testing.T) {
	var absent *struct {
		Index map[string]int
		Ints  int
		Keys  []string
		Vals  []int
	}
	if Lend(absent) != nil {
		t.Fatal("Lend(nil) is not nil")
	}
}
