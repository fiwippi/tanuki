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

func (d *Date) String() string  {
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

	//parts := strings.Split(data, "-")
	//if len(parts) != 3 {
	//	return errors.New("date does not have 3 distinct parts, i.e. yyyy-mm-dd")
	//}
	//
	//year, err := strconv.Atoi(parts[0])
	//if err != nil {
	//	return err
	//}
	//month, err := strconv.Atoi(parts[1])
	//if err != nil {
	//	return err
	//}
	//dat, err := strconv.Atoi(parts[2])
	//if err != nil {
	//	return err
	//}

	return nil
}