package users

import (
	"sync"

	"github.com/fiwippi/tanuki/internal/errors"
)

var ErrProgressEntryNotExist = errors.New("entry does not exist")

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
	return ErrProgressEntryNotExist.Fmt(i)
}

func (p *SeriesProgress) SetAllRead() {
	p.M.RLock()
	defer p.M.RUnlock()

	for _, e := range p.Entries {
		if e != nil {
			e.SetRead()
		}

	}
}

func (p *SeriesProgress) SetAllUnread() {
	p.M.RLock()
	defer p.M.RUnlock()

	for _, e := range p.Entries {
		if e != nil {
			e.SetUnread()
		}
	}
}

func (p *SeriesProgress) DeleteEntry(i int) {
	p.M.RLock()
	defer p.M.RUnlock()

	if i >= 0 && i < len(p.Entries) {
		p.Entries[i] = nil
	}

	p.filter()
}

func (p *SeriesProgress) filter() {
	filtered := p.Entries[:0]
	for _, rp := range p.Entries {
		if rp != nil {
			filtered = append(filtered, rp)
		}
	}
	p.Entries = filtered
}
