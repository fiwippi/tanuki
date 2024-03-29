package storage

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/fiwippi/tanuki/internal/archive"
	"github.com/fiwippi/tanuki/internal/fse"
	"github.com/fiwippi/tanuki/internal/hash"
	"github.com/fiwippi/tanuki/internal/image"
	"github.com/fiwippi/tanuki/internal/sqlutil"

	"github.com/fiwippi/tanuki/pkg/manga"
)

func testPopulateGetCatalog(t *testing.T) {
	t.Run("CatalogWithEntries", func(t *testing.T) {
		s := mustOpenStoreMem(t)
		defer mustCloseStore(t, s)

		require.Nil(t, s.PopulateCatalog())

		for _, d := range parsedData {
			dbSeries, err := s.GetSeries(d.s.SID)
			require.Nil(t, err)
			equalSeries(t, d.s, dbSeries)

			dbEntries, err := s.GetEntries(d.s.SID)
			require.Nil(t, err)
			equalEntries(t, d.e, dbEntries)
		}

		ctl, err := s.GetCatalog()
		require.Nil(t, err)
		require.Equal(t, 3, len(ctl))
		equalSeries(t, parsedData[0].s, ctl[0])
		equalSeries(t, parsedData[1].s, ctl[1])
		equalSeries(t, parsedData[2].s, ctl[2])
	})

	t.Run("CatalogWithoutEntries", func(t *testing.T) {
		s := mustOpenStoreMem(t)
		defer mustCloseStore(t, s)

		s.libraryPath = "."
		require.Nil(t, s.PopulateCatalog())
		ctl, err := s.GetCatalog()
		require.Nil(t, err)
		require.True(t, len(ctl) == 0)
	})
}

func TestStore_PopulateCatalog(t *testing.T) {
	testPopulateGetCatalog(t)
}

func TestStore_GetCatalog(t *testing.T) {
	testPopulateGetCatalog(t)
}

func TestStore_GenerateThumbnails(t *testing.T) {
	s := mustOpenStoreMem(t)

	for _, d := range parsedData {
		require.Nil(t, s.AddSeries(d.s, d.e))
	}

	// Thumbnail generation returns no errors
	require.Nil(t, s.GenerateThumbnails(true))

	// Thumbnails can be accessed for every series
	// and entry directly (without needing to regenerate
	//  the thumbnails)
	for _, d := range parsedData {
		var thumbSeries []byte
		require.Nil(t, s.pool.Get(&thumbSeries, `SELECT thumbnail FROM series WHERE sid = ?`, d.s.SID))
		require.True(t, len(thumbSeries) > 0)

		for _, e := range d.e {
			var thumbEntry []byte
			require.Nil(t, s.pool.Get(&thumbEntry, `SELECT thumbnail FROM entries WHERE sid = ? AND eid = ?`, e.SID, e.EID))
			require.True(t, len(thumbEntry) > 0)
		}
	}

	mustCloseStore(t, s)
}

func testGetDeleteMissingEntries(t *testing.T) {
	s := mustOpenStoreMem(t)

	series := manga.Series{
		SID:        hash.SHA1("a"),
		Title:      "a",
		NumEntries: 1,
		NumPages:   2,
		ModTime:    sqlutil.Time(time.Now()),
	}
	missingSeries := MissingItem{
		Type:  "Series",
		Title: "a",
		Path:  filepath.Join(s.libraryPath, "a"),
	}

	entries := []manga.Entry{
		{
			SID:     hash.SHA1("a"),
			EID:     hash.SHA1("b"),
			Title:   "b",
			Archive: manga.Archive{Type: archive.Zip, Path: "./b"},
			Pages: manga.Pages{
				{Path: "1.jpg", Type: image.JPEG},
				{Path: "2.jpg", Type: image.JPEG},
			},
			ModTime: sqlutil.Time(time.Now()),
		},
	}
	missingEntry := MissingItem{
		Type:  "Entry",
		Title: "b",
		Path:  "./b",
	}

	// Add the fake series and confirm they've been added
	require.Nil(t, s.AddSeries(series, entries))
	dbSeries, err := s.GetSeries(series.SID)
	require.Nil(t, err)
	require.NotNil(t, dbSeries)
	require.NotEqual(t, manga.Series{}, dbSeries)
	dbEntries, err := s.GetEntries(series.SID)
	require.Nil(t, err)
	require.True(t, len(dbEntries) == 1)

	// Check they exist as missing entries
	missingItems, err := s.GetMissingItems()
	require.Nil(t, err)
	require.Len(t, missingItems, 2)
	require.Equal(t, missingSeries, missingItems[0])
	require.Equal(t, missingEntry, missingItems[1])

	// Delete the missing items
	require.Nil(t, s.DeleteMissingItems())

	// Check they don't exist in DB
	dbSeries, err = s.GetSeries(series.SID)
	require.NotNil(t, err)
	require.Equal(t, manga.Series{}, dbSeries)
	dbEntries, err = s.GetEntries(series.SID)
	require.Nil(t, err)
	require.True(t, len(dbEntries) == 0)

	// Check they don't return as missing items
	missingItems, err = s.GetMissingItems()
	require.Nil(t, err)
	require.True(t, len(missingItems) == 0)

	mustCloseStore(t, s)
}

func TestStore_GetMissingEntries(t *testing.T) {
	testGetDeleteMissingEntries(t)
}

func TestStore_DeleteMissingEntries(t *testing.T) {
	testGetDeleteMissingEntries(t)
}

func TestStore_GenerateThumbnailsDoesNotBlockExcessively(t *testing.T) {
	dir := os.Getenv("TANUKI_CUSTOM_LIB_DIR")
	if dir == "" {
		dir = libPath
	}
	require.NotEqual(t, "", dir)
	require.True(t, fse.Exists(dir))

	s, tf := mustOpenStoreFile(t, nil, false)
	defer tf.Close()
	defer mustCloseStore(t, s)
	s.libraryPath = dir

	require.Nil(t, s.PopulateCatalog())

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		require.Nil(t, s.GenerateThumbnails(true))
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	doneChan := make(chan struct{})
	go func() {
		defer wg.Done()

		var err error

		select {
		case <-ctx.Done():
			err = ctx.Err()
		case <-doneChan:
			err = nil
		}

		require.Nil(t, err)
	}()

	go func() {
		t.Log("Populating catalog")
		require.Nil(t, s.PopulateCatalog())
		doneChan <- struct{}{}
		t.Log("Populated catalog")
	}()

	wg.Wait()
}
