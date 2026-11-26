package Data_HashMap_ffi

import (
	"math/rand/v2"
	"slices"
	"testing"
)

func pair(k int, v int) [2]int { return [2]int{k, v} }

func identityHash(k int) int { return k }

func intEqual(a int, b int) bool { return a == b }

func plus(a int, b int) int { return a + b }

func TestNativeVisitsEntriesInTheTriesOrder(t *testing.T) {
	random := rand.New(rand.NewPCG(1, 2))
	native := NativeEmpty[int, int]()
	generic := Empty[int, int]()
	for step := 0; step < 5000; step++ {
		key := random.IntN(300) - 150
		if random.IntN(1<<20) == 0 {
			key = int(random.Int32())
		}
		value := random.IntN(1000)
		switch random.IntN(4) {
		case 0:
			native = NativeInsert(nil, nil, key, value, native)
			generic = Insert(intEqual, identityHash, key, value, generic)
		case 1:
			native = NativeDelete(nil, key, nil, native)
			generic = Delete(intEqual, key, key, generic)
		case 2:
			native = NativeInsertWith(nil, nil, plus, key, value, native)
			generic = InsertWith(intEqual, identityHash, plus, key, value, generic)
		case 3:
			native = NativeFilterWithKey(func(k int, _ int) bool { return k != key }, native)
			generic = FilterWithKey(func(k int, _ int) bool { return k != key }, generic)
		}
		got := NativeToArrayBy(pair, native)
		want := ToArrayBy(pair, generic)
		if !slices.Equal(got, want) {
			t.Fatalf("step %d: native %v, trie %v", step, got, want)
		}
	}
}

func TestNativeOrderDoesNotDependOnHowTheMapWasBuilt(t *testing.T) {
	keys := []int{33, 1, 65, 2, -1, 1 << 30, 97, 0}
	built := NativeEmpty[int, int]()
	for _, key := range keys {
		built = NativeInsert(nil, nil, key, key, built)
	}
	fromArray := NativeFromArray(nil, nil, func(k int) int { return k }, func(k int) int { return k }, keys)
	if got, want := NativeToArrayBy(pair, fromArray), NativeToArrayBy(pair, built); !slices.Equal(got, want) {
		t.Fatalf("fromArray %v, inserts %v", got, want)
	}
}

func TestEveryVersionOfANativeMapKeepsItsEntries(t *testing.T) {
	first := NativeFromArray(nil, nil, func(k int) int { return k }, func(k int) int { return 1 }, []int{1, 2, 3})
	second := NativeInsert(nil, nil, 4, 1, first)
	lent := Lend(second)
	lent = NativeInsertOwned(nil, nil, 5, 1, lent)
	lent = NativeDeleteOwned(nil, 1, nil, lent)
	lent = NativeInsertWithOwned(nil, nil, plus, 2, 10, lent)
	sizes := []int{NativeSize(first), NativeSize(second), NativeSize(lent)}
	if !slices.Equal(sizes, []int{3, 4, 4}) {
		t.Fatalf("sizes %v", sizes)
	}
	nothing, just := int(-1), func(v int) int { return v }
	if got := NativeLookup(nothing, just, nil, 2, nil, lent); got != 11 {
		t.Fatalf("lent 2 is %d", got)
	}
	if got := NativeLookup(nothing, just, nil, 2, nil, second); got != 1 {
		t.Fatalf("second 2 is %d", got)
	}
	if got := NativeLookup(nothing, just, nil, 5, nil, second); got != nothing {
		t.Fatalf("second holds 5")
	}
}

func TestUnionIsLeftBiasedWhicheverWayItIsBuilt(t *testing.T) {
	small := NativeSingleton(int(1), nil, int(100))
	large := NativeFromArray(nil, nil, func(k int) int { return k }, func(k int) int { return k }, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
	first := func(a int, _ int) int { return a }
	nothing, just := int(-1), func(v int) int { return v }
	if got := NativeLookup(nothing, just, nil, 1, nil, NativeUnionWith(nil, nil, first, small, large)); got != 100 {
		t.Fatalf("small first: %d", got)
	}
	if got := NativeLookup(nothing, just, nil, 1, nil, NativeUnionWith(nil, nil, first, large, small)); got != 1 {
		t.Fatalf("large first: %d", got)
	}
}

func TestCollidingStringsShareATrieLeaf(t *testing.T) {
	if HashString("Aa") != HashString("BB") {
		t.Fatalf("Aa and BB should collide")
	}
	m := NativeEmpty[string, int]()
	m = NativeInsert(nil, nil, "BB", 1, m)
	m = NativeInsert(nil, nil, "Aa", 2, m)
	m = NativeDelete(nil, "BB", nil, m)
	keys := NativeToArrayBy(func(k string, _ int) string { return k }, m)
	if !slices.Equal(keys, []string{"Aa"}) {
		t.Fatalf("keys %v", keys)
	}
}

func TestHashStringReadsUTF16CodeUnits(t *testing.T) {
	cases := map[string]int{
		"":   0,
		"a":  97,
		"ab": 97*31 + 98,
		"😀":  0xD83D*31 + 0xDE00,
	}
	for text, want := range cases {
		if got := HashString(text); got != want {
			t.Errorf("HashString(%q) = %d, want %d", text, got, want)
		}
	}
}

func TestHashNumberOfWholeNumbersIsTheNumber(t *testing.T) {
	for _, n := range []float64{0, 1, -1, 2147483647, -2147483648} {
		if got := HashNumber(n); got != int(n) {
			t.Errorf("HashNumber(%v) = %d", n, got)
		}
	}
}
