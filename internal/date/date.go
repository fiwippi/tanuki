// Package date implements a custom date format with marshaling
package date

import (
	"encoding/json"
	"fmt"
	"time"
)

// Date acts like a normal time.Time which can marshal/unmarshal
// to JSON and is of the format "2006-01-02"
type Date struct {
	time.Time
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
