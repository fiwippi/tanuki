package feed

import "encoding/xml"

type link interface {
	isLink()
}

type basicLink struct {
	XMLName xml.Name `xml:"link"`
	Href    string   `xml:"href,attr"`
	Rel     Relation `xml:"rel,attr,omitempty"`
	Type    Type     `xml:"type,attr,omitempty"`
}

func (basicLink) isLink() {}

type streamingLink struct {
	basicLink
	Ns        string `xml:"xmlns:pse,attr"`
	PageCount int    `xml:"pse:count,attr"`
}

func (streamingLink) isLink() {}
