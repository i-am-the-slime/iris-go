package avar

// queue is AVar.js's MutableQueue, a doubly linked list whose cells can
// leave from anywhere: a canceler deletes its operation's cell.
type queue[T any] struct {
	head, last *cell[T]
	size       int
}

type cell[T any] struct {
	queue      *queue[T]
	value      T
	next, prev *cell[T]
}

func (q *queue[T]) putLast(value T) *cell[T] {
	added := &cell[T]{queue: q, value: value, prev: q.last}
	if q.last == nil {
		q.head = added
	} else {
		q.last.next = added
	}
	q.last = added
	q.size++
	return added
}

func (q *queue[T]) takeHead() (T, bool) {
	taken := q.head
	if taken == nil {
		var zero T
		return zero, false
	}
	value := taken.value
	taken.delete()
	return value, true
}

// delete removes the cell from its queue; a cell already taken or deleted
// belongs to none, so a canceler run after its operation completed does
// nothing.
func (c *cell[T]) delete() {
	q := c.queue
	if q == nil {
		return
	}
	if c.prev == nil {
		q.head = c.next
	} else {
		c.prev.next = c.next
	}
	if c.next == nil {
		q.last = c.prev
	} else {
		c.next.prev = c.prev
	}
	q.size--
	var zero T
	c.queue, c.next, c.prev, c.value = nil, nil, nil, zero
}
