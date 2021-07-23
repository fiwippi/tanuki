package core

import (
	"fmt"
)

type EntryProgress struct {
	Current int `json:"current"`
	Total   int `json:"total"`
}

func NewEntryProgress(total int) *EntryProgress {
	return &EntryProgress{Current: 0, Total: total}
}

func (rp *EntryProgress) String() string {
	return fmt.Sprintf("%.2f%%", rp.Percent())
}

func (rp *EntryProgress) Set(n int) {
	if n > rp.Total {
		n = rp.Total
	}
	rp.Current = n
}

func (rp *EntryProgress) SetRead() {
	rp.Current = rp.Total
}

func (rp *EntryProgress) SetUnread() {
	rp.Current = 0
}

func (rp *EntryProgress) IsDone() bool {
	return rp.Current == rp.Total
}

func (rp *EntryProgress) Percent() float64 {
	if rp.Current == 0 && rp.Total == 0 {
		return 0
	}

	percent := (float64(rp.Current) / float64(rp.Total)) * 100
	if percent < 0 {
		percent = 0
	} else if percent > 100 {
		percent = 100
	}

	return percent
}
