package storage

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/fiwippi/tanuki/internal/sqlutil"
	"github.com/fiwippi/tanuki/pkg/mangadex"
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

	require.Nil(t, s.AddDownloads(dls...))
	require.Equal(t, 3, s.mustGetRows(t, "downloads"))
	require.Nil(t, s.AddDownloads(dls...))
	require.Equal(t, 6, s.mustGetRows(t, "downloads"))

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

	require.Nil(t, s.AddDownloads(dls...))
	require.Equal(t, 6, s.mustGetRows(t, "downloads"))

	dbDls, err := s.GetAllDownloads()
	require.Nil(t, err)
	for i := range dls {
		// Downloads returned in reverse order so
		// thats why messing around with indexes
		require.Equal(t, *dls[i], *dbDls[len(dbDls)-1-i])
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

	require.Nil(t, s.AddDownloads(dls...))
	require.Equal(t, 6, s.mustGetRows(t, "downloads"))

	dbDls, err := s.GetFailedDownloads()
	require.Nil(t, err)
	require.Equal(t, 1, len(dbDls))
	require.Equal(t, mangadex.DownloadFailed, dbDls[0].Status)

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

	require.Nil(t, s.AddDownloads(dls...))
	require.Nil(t, s.DeleteAllDownloads())
	require.Equal(t, 0, s.mustGetRows(t, "downloads"))

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

	require.Nil(t, s.AddDownloads(dls...))
	require.Nil(t, s.DeleteSuccessfulDownloads())
	require.Equal(t, 3, s.mustGetRows(t, "downloads"))

	dbDls, err := s.GetAllDownloads()
	require.Nil(t, err)
	for _, d := range dbDls {
		require.NotEqual(t, mangadex.DownloadCancelled, d.Status)
		require.NotEqual(t, mangadex.DownloadExists, d.Status)
		require.NotEqual(t, mangadex.DownloadFinished, d.Status)
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
			PublishedAt:     sqlutil.Time(time.Now().Round(time.Second)),
			Pages:           1,
			VolumeNo:        "e",
			ChapterNo:       "f",
		},
		Status:      mangadex.DownloadFinished,
		CurrentPage: 1,
		TotalPages:  1,
		TimeTaken:   -1,
	}

	require.Nil(t, s.AddDownloads(d))
	dls, err := s.GetAllDownloads()
	require.Nil(t, err)
	require.Equal(t, 1, len(dls))
	require.Equal(t, d, dls[0])

	mustCloseStore(t, s)
}
