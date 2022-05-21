package human

import "github.com/fiwippi/tanuki/internal/collections"

type SeriesProgress struct {
	Raw *collections.Map[string, EntryProgress] `json:"entries"`
}

func NewSeriesProgress() *SeriesProgress {
	return &SeriesProgress{Raw: collections.NewMap[string, EntryProgress]()}
}

func (sp *SeriesProgress) Set(eid string, n int) error {
	ep, err := sp.Raw.Get(eid)
	if err != nil {
		return err
	}

	ep.set(n)
	sp.Raw.Set(eid, ep)
	return nil
}

func (sp *SeriesProgress) SetAllRead() {
	sp.Raw.ForEach(func(e EntryProgress) EntryProgress {
		e.setRead()
		return e
	})
}

func (sp *SeriesProgress) SetAllUnread() {
	sp.Raw.ForEach(func(e EntryProgress) EntryProgress {
		e.setUnread()
		return e
	})
}
