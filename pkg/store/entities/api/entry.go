package api

import (
	"encoding/json"
	"github.com/fiwippi/tanuki/internal/date"
)

type Entry struct {
	Order        int        `json:"order"`
	Hash         string     `json:"hash"`
	Title        string     `json:"title"`
	Pages        int        `json:"pages"`
	Path         string     `json:"path"`
	Chapter      int        `json:"chapter"`
	Volume       int        `json:"volume"`
	Author       string     `json:"author"`
	DateReleased *date.Date `json:"date_released"`
}

type Entries []*Entry

func UnmarshalEntries(data []byte) Entries {
	var s Entries
	err := json.Unmarshal(data, &s)
	if err != nil {
		panic(err)
	}
	return s
}
