package api

import "github.com/fiwippi/tanuki/pkg/core"

type Entry struct {
	Order        int        `json:"order"`
	Hash         string     `json:"hash"`
	Title        string     `json:"title"`
	Pages        int        `json:"pages"`
	Path         string     `json:"path"`
	Chapter      int        `json:"chapter"`
	Volume       int        `json:"volume"`
	Author       string     `json:"author"`
	DateReleased *core.Date `json:"date_released"`
}

type Entries []*Entry

type Series struct {
	Order        int        `json:"order"`
	Hash         string     `json:"hash"`
	Title        string     `json:"title"`
	Entries      int        `json:"entries"`
	Tags         []string   `json:"tags"`
	Author       string     `json:"author"`
	DateReleased *core.Date `json:"date_released"`
}

type Catalog []*Series

type MissingItem struct {
	Type  string `json:"type"`
	Title string `json:"title"`
	Path  string `json:"path"`
}

type MissingItems []*MissingItem

type EditableEntryMetadata struct {
	Title        string     `json:"title"`
	Chapter      int        `json:"chapter"`
	Volume       int        `json:"volume"`
	Author       string     `json:"author"`
	DateReleased *core.Date `json:"date_released"`
}

type EditableSeriesMetadata struct {
	Title        string     `json:"title"`
	Author       string     `json:"author"`
	DateReleased *core.Date `json:"date_released"`
}
