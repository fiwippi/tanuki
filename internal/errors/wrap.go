package errors

import "fmt"

func Wrap(old, new error) error {
	if old == nil {
		return new
	}
	return fmt.Errorf("%w --> %s", old, new)
}
