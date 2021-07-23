package core

import (
	"time"
)

const (
	ChapterZeroValue = -1
	VolumeZeroValue  = -1
	AuthorZeroValue  = ""
	TitleZeroValue   = ""
)

var (
	TimeZeroValue = time.Time{}
)

type ParsedEntryMetadata struct {
	Title        string `json:"title"`
	Chapter      int    `json:"chapter"`
	Volume       int    `json:"volume"`
	Author       string `json:"author"`
	DateReleased *Date  `json:"date_released"`
}

func NewEntryMetadata() *ParsedEntryMetadata {
	return &ParsedEntryMetadata{
		DateReleased: nil,
		Chapter:      ChapterZeroValue,
		Volume:       VolumeZeroValue,
		Author:       AuthorZeroValue,
		Title:        TitleZeroValue,
	}
}
