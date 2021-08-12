package feed

import "encoding/xml"

const pseNs = "http://vaemendis.net/opds-pse/ns"

type pseLink struct {
	XMLName   xml.Name `xml:"link"`
	Ns        string   `xml:"xmlns:pse,attr"`
	PageCount int      `xml:"pse:count,attr"`
	Href      string   `xml:"href,attr"`
	Rel       string   `xml:"rel,attr,omitempty"`
	Type      string   `xml:"type,attr,omitempty"`
}
