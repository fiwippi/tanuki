package users

import "encoding/json"

type CatalogProgress struct {
	Data map[string]SeriesProgress
}

func NewCatalogProgress() CatalogProgress {
	return CatalogProgress{
		Data: make(map[string]SeriesProgress),
	}
}

func (p *CatalogProgress) AddSeries(sid string, sp SeriesProgress) {
	p.Data[sid] = sp
}

func (p *CatalogProgress) DeleteSeries(sid string) {
	delete(p.Data, sid)
}

func (p *CatalogProgress) HasSeries(sid string) bool {
	_, ok := p.Data[sid]
	return ok
}

func (p *CatalogProgress) GetSeries(sid string) SeriesProgress {
	return p.Data[sid]
}

func (p *CatalogProgress) SetSeries(sid string, sp SeriesProgress) {
	p.Data[sid] = sp
}

func UnmarshalCatalogProgress(data []byte) CatalogProgress {
	var s CatalogProgress
	err := json.Unmarshal(data, &s)
	if err != nil {
		panic(err)
	}
	return s
}
