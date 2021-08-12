package manga

import (
	"encoding/json"
	"time"

	"github.com/fiwippi/tanuki/internal/date"
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

type EntryMetadata struct {
	Title        string     `json:"title"`
	Chapter      int        `json:"chapter"`
	Volume       int        `json:"volume"`
	Author       string     `json:"author"`
	DateReleased *date.Date `json:"date_released"`
}

func NewEntryMetadata() *EntryMetadata {
	return &EntryMetadata{
		DateReleased: nil,
		Chapter:      ChapterZeroValue,
		Volume:       VolumeZeroValue,
		Author:       AuthorZeroValue,
		Title:        TitleZeroValue,
	}
}

func UnmarshalEntryMetadata(data []byte) *EntryMetadata {
	var s EntryMetadata
	err := json.Unmarshal(data, &s)
	if err != nil {
		panic(err)
	}
	return &s
}

type SeriesMetadata struct {
	Title        string     `json:"title"`
	Author       string     `json:"author"`
	DateReleased *date.Date `json:"date_released"`
}

func UnmarshalSeriesMetadata(data []byte) *SeriesMetadata {
	var s SeriesMetadata
	err := json.Unmarshal(data, &s)
	if err != nil {
		panic(err)
	}
	return &s
}
