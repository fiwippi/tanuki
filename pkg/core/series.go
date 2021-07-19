package core

// Series is a collection of Manga volumes/chapters
type Series struct {
	Title   string    // Title of the Series
}

func newSeries() *Series {
	return &Series{}
}