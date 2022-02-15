package users

import "fmt"

type EntryProgress struct {
	Title   string
	Current int `json:"current"`
	Total   int `json:"total"`
}

func NewEntryProgress(total int, title string) EntryProgress {
	return EntryProgress{Current: 0, Total: total, Title: title}
}

func (ep EntryProgress) String() string {
	return fmt.Sprintf("%.2f%%", ep.Percent())
}

func (ep EntryProgress) Percent() float64 {
	if ep.Current == 0 && ep.Total == 0 {
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

func (ep *EntryProgress) Set(n int) {
	if n > ep.Total {
		n = ep.Total
	}
	ep.Current = n
}

func (ep *EntryProgress) SetRead() {
	ep.Current = ep.Total
}

func (ep *EntryProgress) SetUnread() {
	ep.Current = 0
}
