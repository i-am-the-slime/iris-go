package object

import (
	"reflect"
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
