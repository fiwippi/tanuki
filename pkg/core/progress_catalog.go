package core

type CatalogProgress struct {
	Data map[string]*SeriesProgress
}

func NewCatalogProgress() *CatalogProgress {
	return &CatalogProgress{
		Data: make(map[string]*SeriesProgress),
	}
}

func (p *CatalogProgress) AddSeries(sid string, entries int) {
	p.Data[sid] = NewSeriesProgress(entries)
}

func (p *CatalogProgress) DeleteSeries(sid string) {
	delete(p.Data, sid)
}

func (p *CatalogProgress) GetSeries(sid string) *SeriesProgress {
	return p.Data[sid]
}
