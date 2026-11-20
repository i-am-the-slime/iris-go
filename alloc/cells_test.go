package alloc

import (
	"slices"
	"testing"
)

type list struct {
	head int32
	tail *list
}

func TestCellsHoldTheirValues(t *testing.T) {
	var cells Cells[list]
	var built *list
	for value := int32(1); value <= 1000; value++ {
		built = cells.New(list{head: value, tail: built})
	}
	for expected := int32(1000); expected >= 1; expected-- {
		if built.head != expected {
			t.Fatalf("cell holds %d, want %d", built.head, expected)
		}
		built = built.tail
	}
	if built != nil {
		t.Fatalf("list continues past its first cell")
	}
}

func TestCellsAreDistinct(t *testing.T) {
	var cells Cells[list]
	seen := map[*list]bool{}
	for value := int32(0); value < 1000; value++ {
		cell := cells.New(list{head: value})
		if seen[cell] {
			t.Fatalf("cell %d handed out twice", value)
		}
		seen[cell] = true
	}
}

func TestWritingACellLeavesItsNeighboursAlone(t *testing.T) {
	var cells Cells[list]
	first := cells.New(list{head: 1})
	second := cells.New(list{head: 2})
	*first = list{head: 3, tail: second}
	if second.head != 2 || second.tail != nil {
		t.Fatalf("neighbour changed to %+v", *second)
	}
}

func TestChunksDoubleUpToTheLimit(t *testing.T) {
	var cells Cells[list]
	var sizes []int
	for range 1 + 2 + 4 + 8 + 16 + 32 + 64 + 128 + 128 {
		grows := len(cells.free) == 0
		cells.New(list{})
		if grows {
			sizes = append(sizes, cells.size)
		}
	}
	expected := []int{1, 2, 4, 8, 16, 32, 64, 128, 128}
	if !slices.Equal(sizes, expected) {
		t.Fatalf("chunk sizes %v, want %v", sizes, expected)
	}
}

func TestOneCellCostsOneAllocation(t *testing.T) {
	allocations := testing.AllocsPerRun(100, func() {
		var cells Cells[list]
		cells.New(list{head: 1})
	})
	if allocations > 1 {
		t.Fatalf("%v allocations for one cell, want at most 1", allocations)
	}
}

func TestChunksAmortiseAllocations(t *testing.T) {
	allocations := testing.AllocsPerRun(10, func() {
		var cells Cells[list]
		var built *list
		for value := int32(0); value < 900; value++ {
			built = cells.New(list{head: value, tail: built})
		}
	})
	if allocations > 16 {
		t.Fatalf("%v allocations for 900 cells, want at most 16", allocations)
	}
}
