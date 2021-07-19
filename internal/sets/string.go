package sets

import (
	"fmt"
	"strings"
)

// Structs take zero bytes unlike bools
var exists = struct{}{}

type Set struct {
	Data map[string]struct{} `json:"data"`
}

func NewSet() *Set {
	s := &Set{}
	s.Data = make(map[string]struct{})
	return s
}

func (s *Set) Add(values ...string) {
	for _, v := range values {
		s.Data[v] = exists
	}
}

func (s *Set) Clear() {
	s.Data = make(map[string]struct{})
}

func (s *Set) Copy() *Set {
	ns := NewSet()
	for x := range s.Data {
		ns.Add(x)
	}

	return ns
}

func (s *Set) Has(value string) bool {
	_, c := s.Data[value]
	return c
}

func (s *Set) IsEmpty() bool {
	return s.Size() == 0
}

func (s *Set) List() []string {
	list := make([]string, 0, len(s.Data))

	for item := range s.Data {
		list = append(list, item)
	}

	return list
}

func (s *Set) Pop() string {
	for item := range s.Data {
		delete(s.Data, item)
		return item
	}
	return ""
}

func (s *Set) Remove(value string) {
	delete(s.Data, value)
}

func (s *Set) Size() int {
	return len(s.Data)
}

func (s *Set) String() string {
	t := make([]string, 0, len(s.List()))
	for _, item := range s.List() {
		t = append(t, fmt.Sprintf("%v", item))
	}

	return fmt.Sprintf("[%s]", strings.Join(t, ", "))
}