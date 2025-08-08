package p

import "sync"

type Queue[T any] struct {
	elements []T
	mu       sync.Mutex
}

func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{elements: make([]T, 0)}
}

func (q *Queue[T]) Push(elem T) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.elements = append(q.elements, elem)
}

func (q *Queue[T]) Pop() (T, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	var elem T
	if len(q.elements) == 0 {
		return elem, false
	}
	elem = q.elements[0]
	q.elements = q.elements[1:]
	return elem, true
}

func (q *Queue[T]) Size() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.elements)
}

func (q *Queue[T]) IsEmpty() bool {
	return q.Size() == 0
}
