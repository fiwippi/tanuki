package feed

import (
	"encoding/xml"
	"strings"
	"time"
)

type feed struct {
	XMLName xml.Name `xml:"feed"`
	Xmlns   string   `xml:"xmlns,attr"` // Namespace of the feed
	ID      string   `xml:"id"`         // IRI of the feed
	Links   []link   `xml:"link"`       // Links the feed references
	Title   string   `xml:"title"`      // Title of the catalog
	Updated opdsTime `xml:"updated"`    // Last modified time of the feed
	Author  author   `xml:"author"`     // Owner of the OPDS feed
	Entries []entry  `xml:"entry"`

	root string // opds root url
}

func newFeed(root, id, title string) *feed {
	return &feed{
		Xmlns:   "http://www.w3.org/2005/Atom",
		Links:   []link{},
		Entries: []entry{},
		root:    strings.TrimRight(root, "/"),
		ID:      id,
		Title:   title,
	}
}

func (f *feed) SetAuthor(name, uri string) {
	f.Author = author{
		Name: name,
		URI:  uri,
	}
}

func (f *feed) SetUpdated(t time.Time) {
	f.Updated = opdsTime{t}
}

func (f *feed) addLink(href string, r Relation, t Type) {
	f.Links = append(f.Links, basicLink{
		Href: f.root + href,
		Rel:  r,
		Type: t,
	})
}
