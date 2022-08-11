package human

import "fmt"

type EntryProgress struct {
	EID     string `json:"eid" db:"eid"`
	Current int    `json:"current" db:"current"`
	Total   int    `json:"total" db:"total"`
}

func (ep EntryProgress) String() string {
	return fmt.Sprintf("%.2f%%", ep.percent())
}

func (ep EntryProgress) percent() float64 {
	if ep.Total == 0 || ep.Current == 0 && ep.Total == 0 {
		return 0
	}

	percent := (float64(ep.Current) / float64(ep.Total)) * 100
	if percent < 0 {
		percent = 0
	} else if percent > 100 {
		percent = 100
	}

	return percent
}

func (ep *EntryProgress) set(n int) {
	if n > ep.Total {
		n = ep.Total
	}
	ep.Current = n
}

func (ep *EntryProgress) setRead() {
	ep.Current = ep.Total
}

func (ep *EntryProgress) setUnread() {
	ep.Current = 0
}
