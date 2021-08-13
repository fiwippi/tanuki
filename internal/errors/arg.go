// Package errors provides utilities to format errors
package errors

import "fmt"

// ArgErr implements the default error with the ability
// to supply additional arguments when returning the error
type ArgErr struct {
	message string
	args    []interface{}
}

// New creates a new default error
func New(message string) *ArgErr {
	return &ArgErr{message: message}
}

// Fmt provides additional args to ArgErr which it will
// display in the error message
func (a *ArgErr) Fmt(args ...interface{}) *ArgErr {
	data := append(make([]interface{}, 0, len(args)), args...)

	return &ArgErr{
		message: a.message,
		args:    data,
	}
}

func (a *ArgErr) Error() string {
	if a.args != nil {
		return fmt.Sprintf("%s: '%v'", a.message, a.args)
	}
	return fmt.Sprintf("%s", a.message)
}
