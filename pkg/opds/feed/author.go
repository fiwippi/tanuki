package feed

// Author is the owner of the OPDS feed
type Author struct {
	Name string `xml:"name"`
	URI  string `xml:"uri"`
}
