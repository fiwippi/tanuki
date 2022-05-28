package dbutil

import (
	"database/sql/driver"
	"errors"
)

type NullString string

func (s NullString) Value() (driver.Value, error) {
	if len(s) == 0 {
		return nil, nil
	}
	return string(s), nil
}

func (s *NullString) Scan(src interface{}) error {
	if src == nil {
		*s = ""
		return nil
	}

	v, ok := src.(string)
	if !ok {
		return errors.New("invalid string assertion")
	}

	*s = NullString(v)
	return nil
}
