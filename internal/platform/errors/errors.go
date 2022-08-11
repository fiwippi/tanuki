package errors

import (
	"strings"
)

type Errors struct {
	errors []error
}

func (e *Errors) Add(err error) {
	if err != nil {
		e.errors = append(e.errors, err)
	}
}

func (e *Errors) Ret() error {
	if e == nil || e.IsEmpty() {
		return nil
	}
	return e
}

func (e *Errors) IsEmpty() bool {
	return e.Len() == 0
}

func (e *Errors) Len() int {
	return len(e.errors)
}

func (e *Errors) Error() string {
	asStr := make([]string, len(e.errors))
	for i, x := range e.errors {
		asStr[i] = x.Error()
	}
	return strings.Join(asStr, ". ")
}
