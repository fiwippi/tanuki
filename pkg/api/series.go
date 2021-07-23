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

func UnmarshalEntries(data []byte) Entries {
	var s Entries
	err := json.Unmarshal(data, &s)
	if err != nil {
		panic(err)
	}
	return s
}

func UnmarshalCatalog(data []byte) Catalog {
	var s Catalog
	err := json.Unmarshal(data, &s)
	if err != nil {
		panic(err)
	}
	return s
}

// CatalogReply for /api/series
type CatalogReply struct {
	Success bool    `json:"success"`
	List    Catalog `json:"list"`
}

// SeriesReply for /api/series/:sid
type SeriesReply struct {
	Success bool   `json:"success"`
	Data    Series `json:"data"`
}

// SeriesEntriesReply for /api/series/:id/entries
type SeriesEntriesReply struct {
	Success bool    `json:"success"`
	List    Entries `json:"list"`
}

// SeriesEntryReply for /api/series/:id/entries/:eid
type SeriesEntryReply struct {
	Success bool  `json:"success"`
	Data    Entry `json:"data"`
}

// PatchCoverReply for /api/series/:sid/cover
type PatchCoverReply struct {
	Success bool `json:"success"`
}

// CatalogProgressReply for /api/catalog/progress
type CatalogProgressReply struct {
	Success  bool                            `json:"success"`
	Progress map[string]*core.SeriesProgress `json:"progress"`
}

// SeriesProgressRequest for /api/series/:sid/progress
type SeriesProgressRequest struct {
	Progress string `json:"progress"`
}

// SeriesProgressReply for /api/series/:sid/progress
type SeriesProgressReply struct {
	Success  bool                  `json:"success"`
	Progress []*core.EntryProgress `json:"progress"`
}

// EntryProgressRequest for /api/series/:sid/entries/:eid/progress
type EntryProgressRequest struct {
	Progress string `json:"progress"`
}

// EntriesProgressReply for /api/series/:sid/entries/:eid/progress
type EntriesProgressReply struct {
	Success  bool                `json:"success"`
	Progress *core.EntryProgress `json:"progress"`
}
