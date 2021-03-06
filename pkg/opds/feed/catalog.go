package feed

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
	c.Feed.Title = "Catalog"
	c.Feed.AddLink(&Link{Href: "/opds/v1.2/catalog", Rel: "self", Type: NavigationFeedType})
	c.Feed.AddLink(&Link{Href: "/opds/v1.2/search", Rel: "search", Type: OpenSearchType})
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
