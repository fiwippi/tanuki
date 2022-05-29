package human

import (
	"encoding/json"
	"errors"
)

type SeriesProgress struct {
	m map[string]EntryProgress
}

func NewSeriesProgress() *SeriesProgress {
	return &SeriesProgress{m: map[string]EntryProgress{}}
}

func (sp *SeriesProgress) Add(eid string, p EntryProgress) {
	sp.m[eid] = p
}

func (sp *SeriesProgress) Get(eid string) (EntryProgress, error) {
	ep, found := sp.m[eid]
	if !found {
		return EntryProgress{}, errors.New("entry progress does not exist")
	}
	return ep, nil
}

// TODO Tests for marshalling series progress

func (sp SeriesProgress) MarshalJSON() ([]byte, error) {
	return json.Marshal(sp.m)
}

func (sp *SeriesProgress) UnmarshalJSON(data []byte) error {
	var m map[string]EntryProgress
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	if m == nil {
		return errors.New("unmarshalled map is nil")
	}

	sp.m = m

	return nil
}
