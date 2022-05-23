package storage

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fiwippi/tanuki/internal/mangadex"
)

func (s *Store) mustGetRows(t *testing.T, table string) int {
	var count int
	require.Nil(t, s.pool.Get(&count, fmt.Sprintf("SELECT COUNT(*) FROM %s", table)))
	return count
}

func TestStore_AddDownloads(t *testing.T) {
	s := mustOpenStoreMem(t)

	dls := []*mangadex.Download{
		{Status: mangadex.DownloadFailed},
		{Status: mangadex.DownloadFailed},
		{Status: mangadex.DownloadFailed},
	}

	assert.Nil(t, s.AddDownloads(dls...))
	assert.Equal(t, 3, s.mustGetRows(t, "downloads"))
	assert.Nil(t, s.AddDownloads(dls...))
	assert.Equal(t, 6, s.mustGetRows(t, "downloads"))

	mustCloseStore(t, s)
}

func TestStore_GetAllDownloads(t *testing.T) {
	s := mustOpenStoreMem(t)

	dls := []*mangadex.Download{
		{Status: mangadex.DownloadQueued},
		{Status: mangadex.DownloadStarted},
		{Status: mangadex.DownloadFinished},
		{Status: mangadex.DownloadCancelled},
		{Status: mangadex.DownloadExists},
		{Status: mangadex.DownloadFailed},
	}

	assert.Nil(t, s.AddDownloads(dls...))
	assert.Equal(t, 6, s.mustGetRows(t, "downloads"))

	dbDls, err := s.GetAllDownloads()
	assert.Nil(t, err)
	for i := range dls {
		// Downloads returned in reverse order so
		// thats why messing around with indexes
		assert.Equal(t, *dls[i], *dbDls[len(dbDls)-1-i])
	}

	mustCloseStore(t, s)
}

func TestStore_GetFailedDownloads(t *testing.T) {
	s := mustOpenStoreMem(t)

	dls := []*mangadex.Download{
		{Status: mangadex.DownloadQueued},
		{Status: mangadex.DownloadStarted},
		{Status: mangadex.DownloadFinished},
		{Status: mangadex.DownloadCancelled},
		{Status: mangadex.DownloadExists},
		{Status: mangadex.DownloadFailed},
	}

	assert.Nil(t, s.AddDownloads(dls...))
	assert.Equal(t, 6, s.mustGetRows(t, "downloads"))

	dbDls, err := s.GetFailedDownloads()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(dbDls))
	assert.Equal(t, mangadex.DownloadFailed, dbDls[0].Status)

	mustCloseStore(t, s)
}

func TestStore_DeleteAllDownloads(t *testing.T) {
	s := mustOpenStoreMem(t)

	dls := []*mangadex.Download{
		{Status: mangadex.DownloadQueued},
		{Status: mangadex.DownloadStarted},
		{Status: mangadex.DownloadFinished},
		{Status: mangadex.DownloadCancelled},
		{Status: mangadex.DownloadExists},
		{Status: mangadex.DownloadFailed},
	}

	assert.Nil(t, s.AddDownloads(dls...))
	assert.Nil(t, s.DeleteAllDownloads())
	assert.Equal(t, 0, s.mustGetRows(t, "downloads"))

	mustCloseStore(t, s)
}

func TestStore_DeleteSuccessfulDownloads(t *testing.T) {
	s := mustOpenStoreMem(t)

	dls := []*mangadex.Download{
		{Status: mangadex.DownloadQueued},
		{Status: mangadex.DownloadStarted},
		{Status: mangadex.DownloadFinished},
		{Status: mangadex.DownloadCancelled},
		{Status: mangadex.DownloadExists},
		{Status: mangadex.DownloadFailed},
	}

	assert.Nil(t, s.AddDownloads(dls...))
	assert.Nil(t, s.DeleteSuccessfulDownloads())
	assert.Equal(t, 3, s.mustGetRows(t, "downloads"))

	dbDls, err := s.GetAllDownloads()
	assert.Nil(t, err)
	for _, d := range dbDls {
		assert.NotEqual(t, mangadex.DownloadCancelled, d.Status)
		assert.NotEqual(t, mangadex.DownloadExists, d.Status)
		assert.NotEqual(t, mangadex.DownloadFinished, d.Status)
	}

	mustCloseStore(t, s)
}

func TestStore_DownloadMarshalling(t *testing.T) {
	s := mustOpenStoreMem(t)

	d := &mangadex.Download{
		MangaTitle: "a",
		Chapter: mangadex.Chapter{
			ID:              "b",
			Title:           "c",
			ScanlationGroup: "d",
			PublishedAt:     time.Unix(1, 0),
			Pages:           1,
			VolumeNo:        "e",
			ChapterNo:       "f",
		},
		Status:      mangadex.DownloadFinished,
		CurrentPage: 1,
		TotalPages:  1,
		TimeTaken:   "h",
	}

	assert.Nil(t, s.AddDownloads(d))
	dls, err := s.GetAllDownloads()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(dls))
	assert.Equal(t, d, dls[0])

	mustCloseStore(t, s)
}
