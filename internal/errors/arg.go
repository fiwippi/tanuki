package errors

import "fmt"

type ArgErr struct {
	message string
	args    []interface{}
}

func New(message string) *ArgErr {
	return &ArgErr{message: message}
}

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
