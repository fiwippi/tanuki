package collections

type Set[T comparable] struct {
	m *Map[T, struct{}]
}

func NewSet[T comparable]() *Set[T] {
	s := &Set[T]{}
	s.m = NewMap[T, struct{}]()
	return s
}

func (s *Set[T]) Add(values ...T) {
	for _, v := range values {
		s.m.Set(v, struct{}{})
	}
}

func (s *Set[T]) Remove(values ...T) {
	for _, v := range values {
		s.m.Remove(v)
	}
}

func (s *Set[T]) Clear() {
	s.m = NewMap[T, struct{}]()
}

func (s *Set[T]) Has(v T) bool {
	return s.m.Has(v)
}

func (s *Set[T]) Empty() bool {
	return s.m.Empty()
}
