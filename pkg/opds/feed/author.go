package feed

import "encoding/xml"

type author struct {
	XMLName xml.Name `xml:"author"`
	Name    string   `xml:"name"`
	URI     string   `xml:"uri"`
}
