package errors

import "fmt"

// Wrap wraps a given error with a new one enabling
// more thorough errors as they are returned up the
// function calls
func Wrap(old, new error) error {
	if old == nil {
		return new
	}
	return fmt.Errorf("%w, %s", old, new)
}
