package progress

import "fmt"

type Entry struct {
	EID     string `json:"eid" db:"eid"`
	Current int    `json:"current" db:"current"`
	Total   int    `json:"total" db:"total"`
}

func (p Entry) String() string {
	return fmt.Sprintf("%.2f%%", p.percent())
}

func (p Entry) percent() float64 {
	if p.Total == 0 {
		return 0
	}

	percent := (float64(p.Current) / float64(p.Total)) * 100
	if percent < 0 {
		percent = 0
	} else if percent > 100 {
		percent = 100
	}

	return percent
}
