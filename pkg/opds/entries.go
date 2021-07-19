package opds

import "github.com/fiwippi/tanuki/pkg/core"

type SeriesEntry struct {
	Title   string `xml:"title"`   // Title of the series
	Updated Time   `xml:"updated"` // When the series was last updated
	ID      string `xml:"id"`      // Series hash
	Content string `xml:"content"` // Empty but tag should still exist
	Link    Link   `xml:"link"`    // NavigationFeed link to see all the series entries
}

type ArchiveEntry struct {
	Title   string  `xml:"title"`   // Title of the series
	Updated Time    `xml:"updated"` // When the series was last updated
	ID      string  `xml:"id"`      // Entry hash
	Content string  `xml:"content"` // Empty but tag should still exist
	Links   []*Link `xml:"link"`

	//
	CoverType, ThumbType core.ImageType `xml:"-"`
	Archive              *core.Archive  `xml:"-"`
}

func (e *ArchiveEntry) AddLink(l *Link) {
	e.Links = append(e.Links, l)
}