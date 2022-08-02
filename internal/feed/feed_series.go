package feed

import (
	"fmt"
	"time"

	"github.com/fiwippi/tanuki/pkg/manga"
)

type SeriesFeed interface {
	SetAuthor(name, uri string)
	SetUpdated(t time.Time)
	AddEntry(id, title, thumbType, coverType, pageType string, pageCount int, updated time.Time, a *manga.Archive)
}

func NewSeriesFeed(root, sid, title string) SeriesFeed {
	s := newFeed(root, sid, title)

	s.addLink("/catalog", RelStart, Navigation)
	s.addLink("/series/"+sid, RelSelf, Acquisition)

	return s
}

func (f *feed) AddEntry(id, title, thumbType, coverType, pageType string, pageCount int, updated time.Time, a *manga.Archive) {
	entryPath := fmt.Sprintf("/opds/v1.2/series/%s/entries/%s", f.ID, id)

	f.Entries = append(f.Entries, entry{
		Title:   title,
		Updated: opdsTime{updated},
		ID:      id,
		Content: fmt.Sprintf("%s - %.1f MiB", a.Type.String(), a.Filesize()),
		Link: []link{
			basicLink{Href: entryPath + "/cover?thumbnail=true", Rel: RelThumbnail, Type: Type(thumbType)},
			basicLink{Href: entryPath + "/cover", Rel: RelCover, Type: Type(coverType)},
			basicLink{Href: entryPath + "/archive", Rel: RelAcquisition, Type: Type(a.Type.MimeType())},
			streamingLink{
				basicLink: basicLink{
					Href: entryPath + "/page/{pageNumber}?zero_based=true",
					Rel:  RelPageStream,
					Type: Type(pageType),
				},
				Ns:        "http://vaemendis.net/opds-pse/ns",
				PageCount: pageCount,
			},
		},
	})
}
