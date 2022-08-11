package feed

type entry struct {
	Title   string   `xml:"title"`   // Title of the entry
	Updated opdsTime `xml:"updated"` // When the entry was last updated
	ID      string   `xml:"id"`      // Entry ID
	Content string   `xml:"content"` // Empty but tag should still exist
	Link    []link   `xml:"link"`    // Links to relevant items, e.g. cover/thumbnail/archive
}
