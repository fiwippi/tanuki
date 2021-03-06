package feed

import (
	"github.com/fiwippi/tanuki/internal/image"
	"github.com/fiwippi/tanuki/pkg/store/entities/manga"
)

type SeriesEntry struct {
	Title   string `xml:"title"`   // Title of the series
	Updated Time   `xml:"updated"` // When the series was last updated
	ID      string `xml:"id"`      // Series hash
	Content string `xml:"content"` // Empty but tag should still exist
	Link    Link   `xml:"link"`    // NavigationFeed link to see all the series entries
}

type ArchiveEntry struct {
	Title   string        `xml:"title"`   // Title of the series
	Updated Time          `xml:"updated"` // When the series was last updated
	ID      string        `xml:"id"`      // EntryProgress hash
	Content string        `xml:"content"` // Empty but tag should still exist
	Links   []interface{} `xml:"link"`

	// Variables used to populate the entry
	CoverType, ThumbType, PageType image.Type     `xml:"-"`
	Archive                        *manga.Archive `xml:"-"`
	Pages                          int            `xml:"-"`
}

func (e *ArchiveEntry) AddLink(l *Link) {
	e.Links = append(e.Links, l)
}
