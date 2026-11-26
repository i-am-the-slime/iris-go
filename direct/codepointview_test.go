package direct

import "testing"

// A view answers every question exactly as the walking operation it stands
// for, bytes included, over ASCII, BMP and astral text, U+FFFD, invalid UTF-8
// and every index from before the start to past the end.

func TestCodePointsViewAgreesWithTheWalkingOperations(t *testing.T) {
	for _, s := range append(codePointSamples(), "plain ascii text") {
		view := CodePoints(s)
		length := CodePointLength(s)
		if got := view.Length(); got != length {
			t.Errorf("Length(%q) = %d, want %d", s, got, length)
		}
		for index := int(-2); index <= int(length)+2; index++ {
			if got, want := foundPoint(view.CodePointAt(index)), foundPoint(CodePointAt(nil, nil, index, s)); got != want {
				t.Errorf("CodePointAt(%d, %q) = %v, want %v", index, s, got, want)
			}
			if got, want := view.Take(index), CodePointTake(nil, index, s); got != want {
				t.Errorf("Take(%d, %q) = %q, want %q", index, s, got, want)
			}
		}
	}
}

func TestCodePointsViewReadsValidTextWithoutAllocating(t *testing.T) {
	view := CodePoints("naïve \U0001F600 café")
	allocations := testing.AllocsPerRun(100, func() {
		_, _ = view.CodePointAt(6)
		_ = view.Take(8)
	})
	if allocations != 0 {
		t.Errorf("indexing valid text allocated %v times", allocations)
	}
}
