package feed

import (
	"encoding/xml"
	"time"
)

const ns = "http://www.w3.org/2005/Atom" // The main xmlns name space

type Feed struct {
	XMLName xml.Name `xml:"feed"` // sets the name of the xml object
	Xmlns   string   `xml:"xmlns,attr"`
	ID      string   `xml:"id"`
	Title   string   `xml:"title"`
	Updated *Time    `xml:"updated"`
	Author  *Author  `xml:"author"`
	Links   []*Link  `xml:"link"`
}

func (f *Feed) SetAuthor(a *Author) {
	f.Author = a
}

func (f *Feed) SetUpdated(t time.Time) {
	f.Updated = &Time{t}
}

func (f *Feed) AddLink(l *Link) {
	f.Links = append(f.Links, l)
}

func NewFeed() *Feed {
	return &Feed{
		Xmlns: ns,
		Links: []*Link{
			// All feeds should have a rel="start" linking to the catalog page
			{Href: "/opds/v1.2/catalog", Rel: "start", Type: NavigationFeedType},
		},
	}
}
