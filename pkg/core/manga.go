package core

import (
	"fmt"
)

// Manga represents an entry which you read, i.e. an archive file
type Manga struct {
	Title    string         // Title of the archive file, the metadata title should be displayed if possible
	Archive  *Archive       // Data about the manga archive file
	Metadata *EntryMetadata // Metadata of the manga
	Pages    []*Page        // Pages of the manga
}

func newManga() *Manga {
	return &Manga{
		Archive:  &Archive{Cover: &Cover{}},
		Metadata: &EntryMetadata{Chapter: -1, Volume: -1},
		Pages:    make([]*Page, 0),
	}
}

func (m *Manga) String() string {
	return fmt.Sprintf("Manga::Title=%s, Format=%s, Metadata={%s}", m.Title, m.Archive.Type, m.Metadata)
}
