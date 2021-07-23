package core

// ParsedSeries is a collection of ParsedEntry volumes/chapters
type ParsedSeries struct {
	Title   string         // Title of the ParsedSeries
	Entries []*ParsedEntry // Slice of all entries in the series
}
