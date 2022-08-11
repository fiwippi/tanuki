package dbutil

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type Time time.Time

func (t Time) After(u Time) bool {
	return time.Time(t).After(time.Time(u))
}

func (t Time) Equal(u Time) bool {
	a := time.Time(t).Round(time.Second)
	b := time.Time(u).Round(time.Second)

	diff := a.Sub(b)
	if diff < 0 {
		diff *= -1
	}

	return a.Equal(b) || diff <= 1*time.Second
}

func (t Time) Time() time.Time {
	return time.Time(t)
}

func (t Time) String() string {
	return t.Time().String()
}

func (t Time) Value() (driver.Value, error) {
	return time.Time(t).Format(time.RFC3339), nil
}

func (t *Time) Scan(src interface{}) error {
	if src == nil {
		*(*time.Time)(t) = time.Time{}
		return nil
	}

	data, ok := src.(string)
	if !ok {
		return errors.New("bad string type assertion")
	}

	parsed, err := time.Parse(time.RFC3339, data)
	if err != nil {
		return err
	}
	*(*time.Time)(t) = parsed

	return nil
}

func (t Time) MarshalJSON() ([]byte, error) {
	v, _ := t.Value()
	return json.Marshal(v)
}

func (t *Time) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	parsed, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}
	*(*time.Time)(t) = parsed
	return nil
}
