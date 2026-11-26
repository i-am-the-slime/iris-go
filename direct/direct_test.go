package direct

import (
	"math"
	"slices"
	"testing"

	"github.com/i-am-the-slime/iris-go/exception"
)

func TestSortByIsStable(t *testing.T) {
	type pair struct {
		key   int
		label string
	}
	byKey := func(left pair, right pair) int { return int(left.key - right.key) }
	sorted := SortBy([]pair{{2, "a"}, {1, "b"}, {2, "c"}, {1, "d"}}, byKey)
	labels := []string{}
	for _, element := range sorted {
		labels = append(labels, element.label)
	}
	if !slices.Equal(labels, []string{"b", "d", "a", "c"}) {
		t.Fatalf("sorted labels %v", labels)
	}
}

func TestCompareArrayInvertsLengthLikeJavaScript(t *testing.T) {
	equal := func(left int, right int) int { return 0 }
	if got := CompareArray(equal, []int{1, 2, 3}, []int{1, 2}); got != -1 {
		t.Fatalf("longer array compared as %d", got)
	}
	if got := CompareArray(equal, []int{1}, []int{1, 2}); got != 1 {
		t.Fatalf("shorter array compared as %d", got)
	}
}

func TestCodePointsTreatSurrogatePairsAsOne(t *testing.T) {
	text := "a😀b"
	if got := ToCodePointArray(nil, nil, text); !slices.Equal(got, []int{'a', 0x1F600, 'b'}) {
		t.Fatalf("code points %v", got)
	}
	if got := CodePointTake(nil, 2, text); got != "a😀" {
		t.Fatalf("took %q", got)
	}
	if got, _ := CodePointAt(nil, nil, 1, text); got != 0x1F600 {
		t.Fatalf("code point at 1 is %x", got)
	}
}

func TestCatchExceptionHandlesThrownErrors(t *testing.T) {
	handled := CatchException(
		func(failure exception.Error) string { return "caught " + failure.Message },
		func() string { panic(exception.NewError("boom")) },
	)
	if handled != "caught boom" {
		t.Fatalf("handled %q", handled)
	}
}

func TestPartitionReturnsFieldsByLabel(t *testing.T) {
	no, yes := Partition(func(x int) bool { return x > 1 }, []int{1, 2, 3})
	if !slices.Equal(no, []int{1}) || !slices.Equal(yes, []int{2, 3}) {
		t.Fatalf("no %v yes %v", no, yes)
	}
}

func TestAppenderNeverAliasesReusedAccumulators(t *testing.T) {
	snoc := Appender[int]()
	shared := snoc(snoc(nil)(1))(2)
	left := snoc(shared)(3)
	right := snoc(shared)(4)
	if !slices.Equal(left, []int{1, 2, 3}) || !slices.Equal(right, []int{1, 2, 4}) {
		t.Fatalf("appended %v and %v", left, right)
	}
}

func TestForeignRecordKeepsAssignmentOrder(t *testing.T) {
	record := ForeignRecord([]string{"tags", "ok"}, []any{[]any{"p"}, true})
	if got := record.(foreignObject).Keys; !slices.Equal(got, []string{"tags", "ok"}) {
		t.Fatalf("keys %v", got)
	}
	if got := ForeignField(record, "ok"); got != true {
		t.Fatalf("field ok is %v", got)
	}
}

// A comparison that is not a total order: Ord Number with NaN, which compares
// greater than everything and everything greater than it. The expected orders
// are what purescript-arrays' JS merge sort produces.
func TestSortByPlacesIncomparableElementsAsJavaScriptDoes(t *testing.T) {
	compareNumber := func(x float64, y float64) int {
		switch {
		case x < y:
			return -1
		case x == y:
			return 0
		}
		return 1
	}
	nan := math.NaN()
	cases := []struct{ input, want []float64 }{
		{[]float64{nan, 1, 0, 0}, []float64{0, 0, 1, nan}},
		{[]float64{nan, 3, 1, nan, 2, 0, nan, -1}, []float64{-1, nan, 0, 2, nan, 1, 3, nan}},
		{[]float64{5, nan, 4, 3, nan, 2, 1}, []float64{1, 2, nan, 3, 4, nan, 5}},
	}
	same := func(x float64, y float64) bool { return x == y || math.IsNaN(x) && math.IsNaN(y) }
	for _, c := range cases {
		if got := SortBy(c.input, compareNumber); !slices.EqualFunc(got, c.want, same) {
			t.Errorf("SortBy(%v) = %v, want %v", c.input, got, c.want)
		}
		mutable := slices.Clone(c.input)
		if got := *SortByMutable(&mutable, compareNumber); !slices.EqualFunc(got, c.want, same) {
			t.Errorf("SortByMutable(%v) = %v, want %v", c.input, got, c.want)
		}
	}
}
