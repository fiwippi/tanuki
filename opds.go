package tanuki

import (
	"encoding/xml"
	"fmt"
	"time"
)

// Relations

type opdsRelation string

const (
	relSelf        opdsRelation = "self"
	relStart       opdsRelation = "start"
	relSearch      opdsRelation = "search"
	relSubsection  opdsRelation = "subsection"
	relCover       opdsRelation = "http://opds-spec.org/image"
	relThumbnail   opdsRelation = "http://opds-spec.org/image/thumbnail"
	relAcquisition opdsRelation = "http://opds-spec.org/acquisition"
	relPageStream  opdsRelation = "http://vaemendis.net/opds-pse/stream"
)

// Type

type opdsType string

const (
	typeNavigation  opdsType = "application/atom+xml;profile=opds-catalog;kind=navigation"
	typeAcquisition opdsType = "application/atom+xml;profile=opds-catalog;kind=acquisition"
	typeSearch      opdsType = "application/opensearchdescription+xml"
)

// Time

type opdsTime struct {
	time.Time
}

func (t *opdsTime) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(t.Format(time.RFC3339), start)
}

// Author

type opdsAuthor struct {
	XMLName xml.Name `xml:"author"`
	Name    string   `xml:"name"`
	URI     string   `xml:"uri"`
}

// Entry

type opdsEntry struct {
	Title       string     `xml:"title"`
	LastUpdated opdsTime   `xml:"updated"`
	ID          string     `xml:"id"`
	Content     string     `xml:"content"` // Empty but tag should still exist
	Link        []opdsLink `xml:"link"`
}

// Links

type opdsLink interface {
	isLink()
}

type simpleLink struct {
	XMLName xml.Name     `xml:"link"`
	Href    string       `xml:"href,attr"`
	Rel     opdsRelation `xml:"rel,attr,omitempty"`
	Type    opdsType     `xml:"type,attr,omitempty"`
}

func (simpleLink) isLink() {}

type streamingLink struct {
	simpleLink
	Namespace string `xml:"xmlns:pse,attr"`
	PageCount int    `xml:"pse:count,attr"`
}

func (streamingLink) isLink() {}

// Feed

const opdsRoot = "/opds/v1.2"

type opdsFeed struct {
	XMLName     xml.Name    `xml:"feed"`
	Namespace   string      `xml:"xmlns,attr"`
	ID          string      `xml:"id"`
	Links       []opdsLink  `xml:"link"`
	Title       string      `xml:"title"`
	LastUpdated opdsTime    `xml:"updated"`
	Author      opdsAuthor  `xml:"author"`
	Entries     []opdsEntry `xml:"entry"`
}

func newOpdsFeed(id, title string, lastUpdated time.Time, author opdsAuthor) *opdsFeed {
	f := &opdsFeed{
		Namespace:   "http://www.w3.org/2005/Atom",
		Links:       make([]opdsLink, 0),
		Entries:     make([]opdsEntry, 0),
		ID:          id,
		Title:       title,
		Author:      author,
		LastUpdated: opdsTime{lastUpdated},
	}

	// All feeds should have a rel="start" linking to the catalog page
	f.addLink("/catalog", relStart, typeNavigation)

	return f
}

func (f *opdsFeed) addLink(href string, r opdsRelation, t opdsType) {
	f.Links = append(f.Links, simpleLink{
		Href: opdsRoot + href,
		Rel:  r,
		Type: t,
	})
}

func (f *opdsFeed) addSeries(s *Series) {
	f.Entries = append(f.Entries, opdsEntry{
		Title:       s.Title,
		LastUpdated: opdsTime{s.ModTime},
		ID:          s.SID,
		Content:     "",
		Link: []opdsLink{
			simpleLink{Href: opdsRoot + "/series/" + s.SID, Rel: relSubsection, Type: typeAcquisition},
		},
	})
}

func (f *opdsFeed) addEntry(e *Entry) {
	content := fmt.Sprintf("zip - %.1f MiB", float64(e.Filesize)/1024/1024)
	if float64(e.Filesize)/1024 < 500 { // Under 500 KiB
		content = fmt.Sprintf("zip - %.1f KiB", float64(e.Filesize)/1024)
	}
	entryPath := fmt.Sprintf("%s/series/%s/entries/%s", opdsRoot, f.ID, e.EID)
	coverType := opdsType(e.Pages[0].Mime)

	f.Entries = append(f.Entries, opdsEntry{
		Title:       e.Title,
		LastUpdated: opdsTime{e.ModTime},
		ID:          e.EID,
		Content:     content,
		Link: []opdsLink{
			simpleLink{Href: entryPath + "/cover?thumbnail=true", Rel: relThumbnail, Type: "image/jpeg"},
			simpleLink{Href: entryPath + "/cover", Rel: relCover, Type: coverType},
			simpleLink{Href: entryPath + "/archive", Rel: relAcquisition, Type: "application/zip"},
			streamingLink{
				simpleLink: simpleLink{
					Href: entryPath + "/page/{pageNumber}",
					Rel:  relPageStream,
					Type: coverType,
				},
				Namespace: "http://vaemendis.net/opds-pse/ns",
				PageCount: len(e.Pages),
			},
		},
	})
}

// Search

const opensearchNs = "http://a9.com/-/spec/opensearch/1.1/"

type opdsSearch struct {
	XMLName        xml.Name       `xml:"OpenSearchDescription"`
	Xmlns          string         `xml:"xmlns,attr"`
	ShortName      string         `xml:"ShortName"`
	Description    string         `xml:"Description"`
	InputEncoding  string         `xml:"InputEncoding"`
	OutputEncoding string         `xml:"OutputEncoding"`
	URL            *opdsSearchURL `xml:"Url"`
}

type opdsSearchURL struct {
	Template string   `xml:"template,attr"`
	Type     opdsType `xml:"type,attr"`
}

func newOpdsSearch() *opdsSearch {
	return &opdsSearch{
		Xmlns:          opensearchNs,
		ShortName:      "Search",
		Description:    "Search for Series",
		InputEncoding:  "UTF-8",
		OutputEncoding: "UTF-8",
		URL: &opdsSearchURL{
			Template: opdsRoot + "/catalog?search={searchTerms}",
			Type:     typeAcquisition,
		},
	}
}
