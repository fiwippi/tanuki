package core

import (
	"encoding/json"
	"fmt"
	"time"
)

type Date struct {
	time.Time
}

func NewDate(t time.Time) *Date {
	return &Date{Time: t}
}

func (d *Date) String() string {
	year, month, day := d.Date()
	return fmt.Sprintf("%04d-%02d-%02d", year, month, day)
}

func (d *Date) MarshalJSON() ([]byte, error) {
	b, err := json.Marshal(d.String())
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (d *Date) UnmarshalJSON(b []byte) error {
	var data string
	err := json.Unmarshal(b, &data)
	if err != nil {
		return err
	}

	t, err := time.Parse("2006-01-02", data)
	if err != nil {
		return err
	}

	d.Time = t

	return nil
}
