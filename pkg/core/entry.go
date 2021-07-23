package core

import (
	"fmt"
)

// ParsedEntry represents an entry which you read, i.e. an archive file
type ParsedEntry struct {
	Order    int                  // 1-indexed order
	Archive  *Archive             // EntriesMetadata about the manga archive file
	Metadata *ParsedEntryMetadata // Metadata of the manga
	Pages    []*Page              // Pages of the manga
}

func newEntry() *ParsedEntry {
	return &ParsedEntry{
		Archive:  &Archive{Cover: &Cover{}},
		Metadata: NewEntryMetadata(),
		Pages:    make([]*Page, 0),
	}
}

func (e *ParsedEntry) String() string {
	return fmt.Sprintf("EntryProgress::Title=%s, Format=%s, Metadata={%s}", e.Archive.Title, e.Archive.Type, e.Metadata)
}
