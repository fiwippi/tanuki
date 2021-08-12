package feed

import "encoding/xml"

const opensearchNs = "http://a9.com/-/spec/opensearch/1.1/"

type Search struct {
	XMLName        xml.Name   `xml:"OpenSearchDescription"`
	Xmlns          string     `xml:"xmlns,attr"`
	ShortName      string     `xml:"ShortName"`
	Description    string     `xml:"Description"`
	InputEncoding  string     `xml:"InputEncoding"`
	OutputEncoding string     `xml:"OutputEncoding"`
	URL            *SearchURL `xml:"Url"`
}

type SearchURL struct {
	Template string `xml:"template,attr"`
	Type     string `xml:"type,attr"`
}

func NewDefaultSearch() *Search {
	return &Search{
		Xmlns:          opensearchNs,
		ShortName:      "Search",
		Description:    "Search for Series",
		InputEncoding:  "UTF-8",
		OutputEncoding: "UTF-8",
		URL: &SearchURL{
			Template: "",
			Type:     AcquisitionFeedType,
		},
	}
}
