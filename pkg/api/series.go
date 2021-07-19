package api

import (
	"encoding/json"

	"github.com/fiwippi/tanuki/pkg/core"
)

// GET /api/series
// GET /api/series/:sid/
// GET, PATCH /api/series/:sid/tags
// GET /api/series/:sid/cover?thumbnail=true
// GET /api/series/:sid/entries
// GET /api/series/:sid/entries/eid
// GET /api/series/:sid/entries/:eid/cover?thumbnail=true
// GET /api/series/:sid/entries/:eid/page/:num
// GET /api/series/:sid/entries/:eid/archive

// SeriesEntries is the data sent to the user for a specific
// series so they can request data about its entries
type SeriesEntries []*SeriesEntry

type SeriesEntry struct {
	Hash         string     `json:"hash"`
	Title        string     `json:"title"`
	Pages        int        `json:"pages"`
	Path         string     `json:"path"`
	Chapter      int        `json:"chapter"`
	Volume       int        `json:"volume"`
	Author       string     `json:"author"`
	DateReleased *core.Date `json:"date_released"`
}

func UnmarshalSeriesEntries(data []byte) SeriesEntries {
	var s SeriesEntries
	err := json.Unmarshal(data, &s)
	if err != nil {
		panic(err)
	}
	return s
}

// SeriesList is the data sent to the user for about all series
// series so the user can request further data about them
type SeriesList []*Series

type Series struct {
	Hash         string     `json:"hash"`
	Title        string     `json:"title"`
	Entries      int        `json:"entries"`
	Tags         []string   `json:"tags"`
	Author       string     `json:"author"`
	DateReleased *core.Date `json:"date_released"`
}

func UnmarshalSeriesList(data []byte) SeriesList {
	var s SeriesList
	err := json.Unmarshal(data, &s)
	if err != nil {
		panic(err)
	}
	return s
}

// SeriesListReply for /api/series
type SeriesListReply struct {
	Success bool       `json:"success"`
	List    SeriesList `json:"list"`
}

// SeriesReply for /api/series/:sid
type SeriesReply struct {
	Success bool   `json:"success"`
	Data    Series `json:"data"`
}

// SeriesEntriesReply for /api/series/:id/entries
type SeriesEntriesReply struct {
	Success    bool          `json:"success"`
	List       SeriesEntries `json:"list"`
}

// SeriesEntryReply for /api/series/:id/entries/:eid
type SeriesEntryReply struct {
	Success bool        `json:"success"`
	Data    SeriesEntry `json:"data"`
}

type SeriesTagsRequest struct {
	Tags    []string `json:"tags"`
}

type SeriesTagsReply struct {
	Success bool     `json:"success"`
	Tags    []string `json:"tags"`
}

type PatchCoverReply struct {
	Success bool     `json:"success"`
}

type MissingEntry struct {
	Type  string `json:"type"`
	Title string `json:"title"`
	Path  string `json:"path"`
}

type MissingEntries []*MissingEntry