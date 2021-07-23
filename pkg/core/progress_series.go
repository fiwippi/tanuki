package core

import (
	"sync"
)

type SeriesProgress struct {
	Entries []*EntryProgress `json:"tracker"`
	M       sync.RWMutex     `json:"mutex"`
}

func NewSeriesProgress(entries int) *SeriesProgress {
	return &SeriesProgress{
		Entries: make([]*EntryProgress, entries),
		M:       sync.RWMutex{},
	}
}

func (p *SeriesProgress) GetEntryProgress(i int) *EntryProgress {
	p.M.RLock()
	defer p.M.RUnlock()

	if i >= 0 && i < len(p.Entries) {
		return p.Entries[i]
	}
	return nil
}

func (p *SeriesProgress) SetEntryProgress(i int, e *EntryProgress) error {
	p.M.RLock()
	defer p.M.RUnlock()

	if i >= 0 && i < len(p.Entries) {
		p.Entries[i] = e
		return nil
	}
	return ErrEntryNotExist
}

func (p *SeriesProgress) SetAllRead() {
	p.M.RLock()
	defer p.M.RUnlock()

	for _, e := range p.Entries {
		e.SetRead()
	}
}

func (p *SeriesProgress) SetAllUnread() {
	p.M.RLock()
	defer p.M.RUnlock()

	for _, e := range p.Entries {
		e.SetUnread()
	}
}

func (p *SeriesProgress) Count() int {
	p.M.RLock()
	defer p.M.RUnlock()

	return len(p.Entries)
}

func (p *SeriesProgress) TotalProgress() *EntryProgress {
	p.M.RLock()
	defer p.M.RUnlock()

	progress := &EntryProgress{}
	for _, e := range p.Entries {
		if e != nil {
			progress.Current += e.Current
			progress.Total += e.Total
		}
	}
	return progress
}

func (p *SeriesProgress) HasEntry(i int) bool {
	p.M.RLock()
	defer p.M.RUnlock()

	if i >= 0 && i < len(p.Entries) {
		return true
	}
	return false
}

func (p *SeriesProgress) DeleteEntry(i int) {
	p.M.RLock()
	defer p.M.RUnlock()

	if i >= 0 && i < len(p.Entries) {
		p.Entries[i] = nil
	}

	newP := make([]*EntryProgress, 0)
	for _, rp := range p.Entries {
		if rp != nil {
		}
		newP = append(newP, rp)
	}
	p.Entries = newP
}
