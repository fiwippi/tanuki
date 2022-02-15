package users

import (
	"github.com/fiwippi/tanuki/internal/errors"
)

var ErrProgressEntryNotExist = errors.New("entry does not exist")

type SeriesProgress struct {
	Title   string
	Entries map[string]EntryProgress `json:"tracker"`
}

func NewSeriesProgress(entries int, title string) SeriesProgress {
	return SeriesProgress{
		Title:   title,
		Entries: make(map[string]EntryProgress, entries),
	}
}

func (p *SeriesProgress) HasEntry(eid string) bool {
	_, ok := p.Entries[eid]
	return ok
}

func (p *SeriesProgress) GetEntry(eid string) (EntryProgress, error) {
	e, ok := p.Entries[eid]
	if !ok {
		return EntryProgress{}, ErrProgressEntryNotExist.Fmt(eid)
	}
	return e, nil
}

func (p *SeriesProgress) SetEntry(eid string, e EntryProgress) {
	p.Entries[eid] = e
}

func (p *SeriesProgress) SetAllRead() {
	for eid, e := range p.Entries {
		e.SetRead()
		p.Entries[eid] = e
	}
}

func (p *SeriesProgress) SetAllUnread() {
	for eid, e := range p.Entries {
		e.SetUnread()
		p.Entries[eid] = e
	}
}

func (p *SeriesProgress) DeleteEntry(eid string) {
	delete(p.Entries, eid)
}

func (p *SeriesProgress) Empty() bool {
	return p.Entries == nil
}
