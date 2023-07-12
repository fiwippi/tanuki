package progress

import "encoding/json"

type Catalog struct {
	m map[string]Series
}

func NewCatalogProgress() *Catalog {
	return &Catalog{m: map[string]Series{}}
}

func (cp *Catalog) Add(sid string, p Series) {
	cp.m[sid] = p
}

func (cp Catalog) MarshalJSON() ([]byte, error) {
	return json.Marshal(cp.m)
}
