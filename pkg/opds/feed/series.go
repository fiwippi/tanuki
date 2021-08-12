package feed

import (
	"fmt"
)

type Series struct {
	*Feed
	Entries []*ArchiveEntry `xml:"entry"`
}

func NewSeries(sid, title string) *Series {
	s := &Series{
		Feed:    NewFeed(),
		Entries: make([]*ArchiveEntry, 0),
	}
	s.Feed.ID = sid
	s.Feed.Title = title
	s.Feed.AddLink(&Link{Href: "/opds/v1.2/series/" + sid, Rel: "self", Type: AcquisitionFeedType})
	return s
}

func (s *Series) AddEntry(e *ArchiveEntry) {
	e.Content = fmt.Sprintf("%s - %.1f MiB", e.Archive.Type.String(), e.Archive.Filesize())

	entryPath := fmt.Sprintf("/opds/v1.2/series/%s/entries/%s", s.ID, e.ID)

	// Thumbnail and cover links
	e.AddLink(&Link{
		Href: entryPath + "/cover?thumbnail=true",
		Rel:  RelThumbnail,
		Type: e.ThumbType.MimeType(),
	})
	e.AddLink(&Link{
		Href: entryPath + "/cover",
		Rel:  RelCover,
		Type: e.CoverType.MimeType(),
	})

	// Archive link
	e.AddLink(&Link{
		Href: entryPath + "/archive",
		Rel:  RelAcquisition,
		Type: e.Archive.Type.MimeType(),
	})

	// Page streaming link
	pse := &pseLink{
		Ns:        pseNs,
		PageCount: e.Pages,
		Href:      entryPath + "/page/{pageNumber}?zero_based=true",
		Rel:       RelPageStream,
		Type:      e.PageType.MimeType(),
	}
	e.Links = append(e.Links, pse)

	s.Entries = append(s.Entries, e)
}
