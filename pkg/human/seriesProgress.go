package human

import "errors"

var ErrEntryProgressNotFound = errors.New("entry progress not found")

type SeriesProgress struct {
	m map[string]*EntryProgress
}

func NewSeriesProgress() *SeriesProgress {
	return &SeriesProgress{m: map[string]*EntryProgress{}}
}

func (sp *SeriesProgress) AddEntry(eid string, total int) {
	sp.m[eid] = &EntryProgress{Total: total}
}

func (sp *SeriesProgress) Set(eid string, n int) error {
	ep, found := sp.m[eid]
	if !found {
		return ErrEntryProgressNotFound
	}

	ep.set(n)
	sp.m[eid] = ep
	return nil
}

func (sp *SeriesProgress) SetAllRead() {
	for k := range sp.m {
		sp.m[k].setRead()
	}
}

func (sp *SeriesProgress) SetAllUnread() {
	for k := range sp.m {
		sp.m[k].setUnread()
	}
}
