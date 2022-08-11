package human

import "encoding/json"

type CatalogProgress struct {
	m map[string]SeriesProgress
}

func NewCatalogProgress() *CatalogProgress {
	return &CatalogProgress{m: map[string]SeriesProgress{}}
}

func (cp *CatalogProgress) Add(sid string, p SeriesProgress) {
	cp.m[sid] = p
}

func (cp CatalogProgress) MarshalJSON() ([]byte, error) {
	return json.Marshal(cp.m)
}
