package feed

import "encoding/xml"

type Link struct {
	XMLName xml.Name `xml:"link"`
	Href    string   `xml:"href,attr"`
	Rel     string   `xml:"rel,attr,omitempty"`
	Type    string   `xml:"type,attr,omitempty"`
}

const (
	NavigationFeedType  = "application/atom+xml;profile=opds-catalog;kind=navigation"
	AcquisitionFeedType = "application/atom+xml;profile=opds-catalog;kind=acquisition"
	OpenSearchType      = "application/opensearchdescription+xml"
)
