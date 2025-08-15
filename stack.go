package p

import "sync"

type Stack[T any] struct {
	elements []T
	mu       sync.Mutex
}

func NewStack[T any]() *Stack[T] {
	return &Stack[T]{
		elements: make([]T, 0),
	}
}

func (s *Stack[T]) Push(elem T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.elements = append(s.elements, elem)
}

func (s *Stack[T]) Pop() (T, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var elem T
	if len(s.elements) == 0 {
		return elem, false
	}
	idx := len(s.elements) - 1
	elem = s.elements[idx]
	s.elements = s.elements[:idx]
	return elem, true
}
