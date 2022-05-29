package storage

import (
	"io/ioutil"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	"github.com/fiwippi/tanuki/internal/platform/dbutil"
	"github.com/fiwippi/tanuki/pkg/manga"
)

// TODO use prepared statements to deduplicate code

// Core

func TestStore_HasEntry(t *testing.T) {
	s := mustOpenStoreMem(t)

	for i, d := range parsedData {
		require.Nil(t, s.AddSeries(parsedData[i].s, parsedData[i].e))
		has, err := s.HasSeries(d.s.SID)
		require.Nil(t, err)
		require.True(t, has)

		for _, e := range parsedData[i].e {
			has, err := s.HasEntry(d.s.SID, e.EID)
			require.Nil(t, err)
			require.True(t, has)
		}
	}

	mustCloseStore(t, s)
}

func TestStore_GetEntry(t *testing.T) {
	s := mustOpenStoreMem(t)

	require.Nil(t, s.AddSeries(parsedData[0].s, parsedData[0].e))
	for _, e := range parsedData[0].e {
		dbEntry, err := s.GetEntry(parsedData[0].s.SID, e.EID)
		require.Nil(t, err)
		require.Equal(t, e, dbEntry)
	}

	mustCloseStore(t, s)
}

func TestStore_getFirstEntry(t *testing.T) {
	s := mustOpenStoreMem(t)

	// Initial first entries are correct
	for _, d := range parsedData {
		require.Nil(t, s.AddSeries(d.s, d.e))
		firstEntry := d.e[0]

		var err error
		var dbFirstEntry *manga.Entry
		fn := func(tx *sqlx.Tx) error {
			dbFirstEntry, err = s.getFirstEntry(tx, d.s.SID)
			return err
		}
		require.Nil(t, s.tx(fn))
		require.Equal(t, firstEntry, dbFirstEntry)
	}

	// If the first entry is deleted and then
	// added again it should still stay correct
	for _, d := range parsedData {
		firstEntry := d.e[0]

		fn := func(tx *sqlx.Tx) error {
			return s.deleteEntry(tx, d.s.SID, firstEntry.EID)
		}
		require.Nil(t, s.tx(fn))

		fn = func(tx *sqlx.Tx) error {
			return s.addEntry(tx, firstEntry, 1)
		}
		require.Nil(t, s.tx(fn))

		var err error
		var dbFirstEntry *manga.Entry
		fn = func(tx *sqlx.Tx) error {
			dbFirstEntry, err = s.getFirstEntry(tx, d.s.SID)
			return err
		}
		require.Nil(t, s.tx(fn))
		require.Equal(t, firstEntry, dbFirstEntry)
	}

	// If there exists a "dead" entry (it doesn't exist on the filesystem)
	// then that means when adding entries to the store there may be cases
	// where a series has two or more entries with the same position, we
	// need to ensure that getFirstEntry() returns the most recently inserted
	// entry if there are identical positions
	for _, d := range parsedData {
		if len(d.e) < 2 {
			continue
		}

		firstEntry := d.e[0]
		secondEntry := d.e[1]

		fn := func(tx *sqlx.Tx) error {
			return s.addEntry(tx, firstEntry, 1)
		}
		require.Nil(t, s.tx(fn))
		fn = func(tx *sqlx.Tx) error {
			return s.addEntry(tx, secondEntry, 1)
		}
		require.Nil(t, s.tx(fn))

		// Second entry is most recent and both have position
		// zero so second entry should be returned
		var err error
		var returnedEntry *manga.Entry
		fn = func(tx *sqlx.Tx) error {
			returnedEntry, err = s.getFirstEntry(tx, d.s.SID)
			return err
		}
		require.Nil(t, s.tx(fn))
		require.Equal(t, secondEntry, returnedEntry)
	}

	mustCloseStore(t, s)
}

func TestStore_GetEntries(t *testing.T) {
	s := mustOpenStoreMem(t)

	series := parsedData[0].s
	entries := parsedData[0].e
	require.Nil(t, s.AddSeries(series, entries))
	dbEntries, err := s.GetEntries(series.SID)
	require.Nil(t, err)
	require.Equal(t, entries, dbEntries)

	mustCloseStore(t, s)
}

func TestStore_addEntry(t *testing.T) {
	s := mustOpenStoreMem(t)

	for _, d := range parsedData {
		require.Nil(t, s.AddSeries(d.s, nil))

		for i, e := range d.e {
			fn := func(tx *sqlx.Tx) error {
				return s.addEntry(tx, e, i+1)
			}
			require.Nil(t, s.tx(fn))

			has, err := s.HasEntry(d.s.SID, e.EID)
			require.Nil(t, err)
			require.True(t, has)
		}
	}

	mustCloseStore(t, s)
}

func TestStore_deleteEntry(t *testing.T) {
	s := mustOpenStoreMem(t)

	for _, d := range parsedData {
		series := parsedData[0].s
		entries := parsedData[0].e
		require.Nil(t, s.AddSeries(series, entries))

		for _, e := range entries {
			fn := func(tx *sqlx.Tx) error {
				return s.deleteEntry(tx, d.s.SID, e.EID)
			}
			require.Nil(t, s.tx(fn))

			has, err := s.HasEntry(d.s.SID, e.EID)
			require.Nil(t, err)
			require.False(t, has)
		}
	}

	mustCloseStore(t, s)
}

// Cover / Thumbnail / Page

func testGetDeleteEntryCover(t *testing.T) {
	s := mustOpenStoreMem(t)

	series := parsedData[0].s
	entries := parsedData[0].e
	require.Nil(t, s.AddSeries(series, entries))

	for _, e := range entries {
		// Get the normal cover
		coverA, err := s.GetEntryCover(series.SID, e.EID)
		require.Nil(t, err)
		require.True(t, len(coverA) > 0)

		// Get custom cover works if it exists
		require.Nil(t, s.SetEntryCover(series.SID, e.EID, customCover))
		coverB, err := s.GetEntryCover(series.SID, e.EID)
		require.Nil(t, err)
		require.True(t, len(coverB) > 0)
		require.NotEqual(t, coverA, coverB)

		// Delete the cover we should have the normal series cover
		require.Nil(t, s.DeleteEntryCustomCover(series.SID, e.EID))
		coverC, err := s.GetEntryCover(series.SID, e.EID)
		require.Nil(t, err)
		require.True(t, len(coverC) > 0)
		require.Equal(t, coverA, coverC)
	}

	mustCloseStore(t, s)
}

func testGetEntryCoverThumbnail(t *testing.T) {
	s := mustOpenStoreMem(t)

	series := parsedData[0].s
	require.Nil(t, s.AddSeries(series, parsedData[0].e))

	for _, e := range parsedData[0].e {
		// Keep track of the thumbnail of the original cover
		thumbA, err := s.GetEntryThumbnail(series.SID, e.EID)
		require.Nil(t, err)
		require.True(t, len(thumbA) > 0)

		// Cannot set a nil cover
		err = s.SetEntryCover(series.SID, e.EID, nil)
		require.NotNil(t, err)
		require.ErrorIs(t, ErrInvalidCover, err)

		// Can set a custom cover
		require.Nil(t, s.SetEntryCover(series.SID, e.EID, customCover))
		cover, err := s.GetEntryCover(series.SID, e.EID)
		require.Nil(t, err)
		require.True(t, len(cover) > 0)
		require.Equal(t, customCover, cover)

		// Thumbnail of the cover should not be of the normal cover
		thumbB, err := s.GetEntryThumbnail(series.SID, e.EID)
		require.Nil(t, err)
		require.True(t, len(thumbB) > 0)
		require.NotEqual(t, thumbA, thumbB)

		// If we access the thumbnail again it persists
		thumbC, err := s.GetEntryThumbnail(series.SID, e.EID)
		require.Nil(t, err)
		require.True(t, len(thumbC) > 0)
		require.Equal(t, thumbB, thumbC)

		// Once the custom cover gets deleted thumbnail goes back to normal
		require.Nil(t, s.DeleteEntryCustomCover(series.SID, e.EID))
		thumbD, err := s.GetEntryThumbnail(series.SID, e.EID)
		require.Nil(t, err)
		require.True(t, len(thumbD) > 0)
		require.Equal(t, thumbA, thumbD)
		require.NotEqual(t, thumbB, thumbD)
	}

	mustCloseStore(t, s)
}

func TestStore_GetEntryCover(t *testing.T) {
	testGetDeleteEntryCover(t)
}

func TestStore_SetEntryCover(t *testing.T) {
	testGetEntryCoverThumbnail(t)
}

func TestStore_DeleteEntryCustomCover(t *testing.T) {
	testGetDeleteEntryCover(t)
}

func TestStore_GetEntryThumbnail(t *testing.T) {
	testGetEntryCoverThumbnail(t)
}

func TestStore_GetPage(t *testing.T) {
	s := mustOpenStoreMem(t)

	for _, d := range parsedData {
		require.Nil(t, s.AddSeries(d.s, d.e))

		for _, e := range d.e {
			for i, p := range e.Pages {
				// Get the page from the archive file
				r, size, err := e.Archive.ReaderForFile(p)
				require.Nil(t, err)
				require.True(t, size > 0)
				originalData, err := ioutil.ReadAll(r)
				require.Nil(t, err)
				require.True(t, len(originalData) > 0)

				// Get the 0-indexed page from the store
				r, size, err = s.GetPage(d.s.SID, e.EID, i, true)
				require.Nil(t, err)
				require.True(t, size > 0)
				zeroData, err := ioutil.ReadAll(r)
				require.Nil(t, err)
				require.True(t, len(zeroData) > 0)
				require.Equal(t, originalData, zeroData)

				// Get the 1-indexed page from the store
				r, size, err = s.GetPage(d.s.SID, e.EID, i+1, false)
				require.Nil(t, err)
				require.True(t, size > 0)
				oneData, err := ioutil.ReadAll(r)
				require.Nil(t, err)
				require.True(t, len(oneData) > 0)
				require.Equal(t, originalData, oneData)
			}
		}
	}

	mustCloseStore(t, s)
}

// Metadata

func TestStore_SetEntryDisplayTitle(t *testing.T) {
	s := mustOpenStoreMem(t)

	series := parsedData[0].s
	require.Nil(t, s.AddSeries(series, parsedData[0].e))

	for _, e := range parsedData[0].e {
		require.Nil(t, s.SetEntryDisplayTitle(series.SID, e.EID, "HUH"))
		dbEntry, err := s.GetEntry(series.SID, e.EID)
		require.Nil(t, err)
		require.Equal(t, dbutil.NullString("HUH"), dbEntry.DisplayTile)

		require.Nil(t, s.SetEntryDisplayTitle(series.SID, e.EID, ""))
		dbEntry, err = s.GetEntry(series.SID, e.EID)
		require.Nil(t, err)
		require.Equal(t, dbutil.NullString(""), dbEntry.DisplayTile)
	}

	mustCloseStore(t, s)
}