// Package alloc hands out heap cells in chunks.
//
// Go's allocator costs one mallocgc call per object, which dominates code that
// builds a linked structure one cell at a time. A loop that allocates many
// cells of one type takes them from a Cells value instead: each chunk is one
// allocation that serves many cells.
//
// A Cells value belongs to one invocation of one generated function and is
// never shared, so goroutines that run PureScript at the same time (such as
// Data.Array.Parallel's workers) never contend for it.
//
// Every cell of a chunk lives as long as any of them does, so chunks start at
// a single cell and double up to chunkLimit: a loop that allocates few cells
// costs exactly what individual allocation costs, and the cells a chunk wastes
// never outnumber the cells handed out before it.
package alloc

// chunkLimit is the largest number of cells one chunk holds.
const chunkLimit = 128

// Cells hands out pointers to zeroed-then-filled cells of T. Its zero value
// is ready to use.
type Cells[T any] struct {
	free []T
	size int
}

// New returns a pointer to a fresh cell holding value. It grows the chunk
// inline rather than through a helper, since a call would put New over Go's
// inlining budget and cost every cell a call.
func (cells *Cells[T]) New(value T) *T {
	if len(cells.free) == 0 {
		cells.size = min(max(2*cells.size, 1), chunkLimit)
		cells.free = make([]T, cells.size)
	}
	cell := &cells.free[0]
	*cell = value
	cells.free = cells.free[1:]
	return cell
}
