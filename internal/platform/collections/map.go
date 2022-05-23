package collections

import "errors"

var ErrDoesNotExist = errors.New("key does not exist in the map")

type Map[K comparable, V any] struct {
	m map[K]V
}

func NewMap[K comparable, V any]() *Map[K, V] {
	s := &Map[K, V]{}
	s.m = make(map[K]V)
	return s
}

func (s *Map[K, V]) Set(k K, v V) {
	s.m[k] = v
}

func (s *Map[K, V]) Get(k K) (V, error) {
	v, b := s.m[k]
	if !b {
		return *new(V), ErrDoesNotExist
	}
	return v, nil
}

func (s *Map[K, V]) Remove(k K) {
	delete(s.m, k)
}

func (s *Map[K, V]) Clear() {
	s.m = make(map[K]V)
}

func (s *Map[K, V]) Has(k K) bool {
	_, h := s.m[k]
	return h
}

func (s *Map[K, V]) ForEach(f func(value V) V) {
	for k, v := range s.m {
		s.m[k] = f(v)
	}
}

func (s *Map[K, V]) Empty() bool {
	return len(s.m) == 0
}
