package array

import "testing"

func TestParallelMapKeepsOrder(t *testing.T) {
	xs := Range(1, 1000)
	got := ParallelMap(func(x int) int { return x * 2 }, xs)
	for i, x := range xs {
		if got[i] != x*2 {
			t.Fatalf("index %d = %d, want %d", i, got[i], x*2)
		}
	}
	if ArrayLength(got) != 1000 {
		t.Fatalf("length = %d", ArrayLength(got))
	}
}

func TestParallelMapReraisesPanics(t *testing.T) {
	defer func() {
		if recover() != "boom" {
			t.Fatalf("panic not re-raised")
		}
	}()
	ParallelMap(func(x int) int {
		if x == 500 {
			panic("boom")
		}
		return x
	}, Range(1, 1000))
}

func TestParallelMapReraisesTheFirstPanicInInputOrder(t *testing.T) {
	defer func() {
		if recovered := recover(); recovered != "first" {
			t.Fatalf("re-raised %v, want first", recovered)
		}
	}()
	ParallelMap(func(x int) int {
		switch x {
		case 10:
			panic("first")
		case 990:
			panic("last")
		}
		return x
	}, Range(1, 1000))
}
