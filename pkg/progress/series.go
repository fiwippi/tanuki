package progress

import (
	"encoding/json"
	"errors"
)

type Series struct {
	m map[string]Entry
}

func NewSeriesProgress() *Series {
	return &Series{m: map[string]Entry{}}
}

func (sp *Series) Add(eid string, p Entry) {
	sp.m[eid] = p
}

func (sp *Series) Get(eid string) (Entry, error) {
	ep, found := sp.m[eid]
	if !found {
		return Entry{}, errors.New("entry progress does not exist")
	}
	return ep, nil
}

// Marshaling

func (sp Series) MarshalJSON() ([]byte, error) {
	return json.Marshal(sp.m)
}

func (sp *Series) UnmarshalJSON(data []byte) error {
	var m map[string]Entry
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	if m == nil {
		return errors.New("unmarshalled map is nil")
	}

	sp.m = m

	return nil
}
