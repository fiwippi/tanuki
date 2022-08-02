package feed

import (
	"time"
)

type CatalogFeed interface {
	SetAuthor(name, uri string)
	SetUpdated(t time.Time)
	AddSeries(id, title string, updated time.Time)
}

func NewCatalogFeed(root string) CatalogFeed {
	c := newFeed(root, "root", "Catalog")

	// All feeds should have a rel="start" linking to the catalog page
	c.addLink("/catalog", RelSelf, Navigation)
	c.addLink("/catalog", RelStart, Navigation)

	return c
}

func (f *feed) AddSeries(id, title string, updated time.Time) {
	f.Entries = append(f.Entries, entry{
		Title:   title,
		Updated: opdsTime{updated},
		ID:      id,
		Content: "",
		Link: []link{
			basicLink{Href: f.root + "/series/" + id, Rel: "subsection", Type: Acquisition},
		},
	})
}
