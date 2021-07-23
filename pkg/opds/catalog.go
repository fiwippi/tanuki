package opds

import (
	"fmt"
)

type Catalog struct {
	*Feed
	Entries []*SeriesEntry `xml:"entry"`
}

func NewCatalog() *Catalog {
	c := &Catalog{
		Feed:    NewFeed(),
		Entries: make([]*SeriesEntry, 0),
	}
	c.Feed.ID = "root"
	c.Feed.Title = "GetCatalog"
	c.Feed.AddLink(&Link{Href: "/opds/v1.2/catalog", Rel: "self", Type: NavigationFeedType})
	return c
}

func (c *Catalog) AddEntry(e *SeriesEntry) {
	e.Link = Link{
		Href: fmt.Sprintf("/opds/v1.2/series/%s", e.ID),
		Rel:  "subsection",
		Type: AcquisitionFeedType,
	}
	c.Entries = append(c.Entries, e)
}
