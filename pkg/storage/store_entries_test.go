package storage

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	"github.com/fiwippi/tanuki/internal/platform/dbutil"
	"github.com/fiwippi/tanuki/internal/platform/image"
	"github.com/fiwippi/tanuki/pkg/manga"
)

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
	t.Run("GetSingleEntry", func(t *testing.T) {
		s := mustOpenStoreMem(t)
		defer mustCloseStore(t, s)

		series := parsedData[0].s
		entries := parsedData[0].e
		require.Nil(t, s.AddSeries(series, entries))
		dbEntries, err := s.GetEntries(series.SID)
		require.Nil(t, err)
		require.Equal(t, entries, dbEntries)
	})

	t.Run("PreserveEntriesOrderAfterDelete", func(t *testing.T) {
		s := mustOpenStoreMem(t)
		defer mustCloseStore(t, s)

		series := parsedData[0].s
		entries := parsedData[0].e
		require.Nil(t, s.AddSeries(series, entries))

		require.Nil(t, s.tx(func(tx *sqlx.Tx) error {
			return s.deleteEntry(tx, series.SID, entries[0].EID)
		}))
		require.Nil(t, s.tx(func(tx *sqlx.Tx) error {
			return s.addEntry(tx, entries[0], 1)
		}))

		dbEntry, err := s.GetEntry(parsedData[0].s.SID, entries[0].EID)
		require.Nil(t, err)
		require.Equal(t, entries[0], dbEntry)

		dbEntry, err = s.GetEntry(parsedData[0].s.SID, entries[1].EID)
		require.Nil(t, err)
		require.Equal(t, entries[1], dbEntry)
	})
}

func TestStore_addEntry(t *testing.T) {
	t.Run("AddingEntries", func(t *testing.T) {
		s := mustOpenStoreMem(t)
		defer mustCloseStore(t, s)

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
	})

	t.Run("EntryDeletedOnModtimeChange", func(t *testing.T) {
		s := mustOpenStoreMem(t)
		defer mustCloseStore(t, s)

		for _, d := range parsedData {
			require.Nil(t, s.AddSeries(d.s, nil))

			for i := range d.e {
				eOriginal := *d.e[i]
				eOriginal.Title = "Before"

				eChangedModTime := *d.e[i]
				eChangedModTime.Title = "After"
				eChangedModTime.ModTime = dbutil.Time(time.Now())

				require.NotEqual(t, eOriginal.ModTime, eChangedModTime.ModTime)

				// Add the original entry
				fnBef := func(tx *sqlx.Tx) error {
					return s.addEntry(tx, &eOriginal, i+1)
				}
				require.Nil(t, s.tx(fnBef))
				e, err := s.GetEntry(d.s.SID, eOriginal.EID)
				require.Nil(t, err)
				require.Equal(t, e.Title, "Before")

				// Add progress for the original entry
				require.Nil(t, s.SetEntryProgressRead(d.s.SID, eOriginal.EID, defaultUID))

				// Ensure the progress is set successfully
				p, err := s.GetEntryProgress(d.s.SID, eOriginal.EID, defaultUID)
				require.Nil(t, err)
				require.NotZero(t, p.Current)
				require.Equal(t, p.Current, p.Total)

				// Add the entry with the changed mod time
				fnAft := func(tx *sqlx.Tx) error {
					return s.addEntry(tx, &eChangedModTime, i+1)
				}
				require.Nil(t, s.tx(fnAft))
				e, err = s.GetEntry(d.s.SID, eOriginal.EID)
				require.Nil(t, err)
				require.Equal(t, e.Title, "After")

				// Check progress for said entry is now deleted
				// Ensure the progress is set successfully
				p, err = s.GetEntryProgress(d.s.SID, eOriginal.EID, defaultUID)
				require.Nil(t, err)
				require.Zero(t, p.Current)
				require.NotEqual(t, p.Current, p.Total)
			}
		}
	})
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

func testGetSetDeleteEntryCover(t *testing.T) {
	s := mustOpenStoreMem(t)

	series := parsedData[0].s
	entries := parsedData[0].e
	require.Nil(t, s.AddSeries(series, entries))

	for _, e := range entries {
		// Get the normal cover
		coverA, it, err := s.GetEntryCover(series.SID, e.EID)
		require.Nil(t, err)
		require.True(t, len(coverA) > 0)
		require.NotEqual(t, image.Invalid, it)

		// Cannot set invalid custom cover
		require.ErrorIs(t, s.SetEntryCover(series.SID, e.EID, "c.png", nil), ErrInvalidCover)
		require.ErrorIs(t, s.SetEntryCover(series.SID, e.EID, "c.aaa", customCover), ErrInvalidCover)

		// Get custom cover works if it exists
		require.Nil(t, s.SetEntryCover(series.SID, e.EID, "c.png", customCover))
		coverB, it, err := s.GetEntryCover(series.SID, e.EID)
		require.Nil(t, err)
		require.True(t, len(coverB) > 0)
		require.NotEqual(t, coverA, coverB)
		require.Equal(t, image.PNG, it)

		// Delete the cover we should have the normal series cover
		require.Nil(t, s.DeleteEntryCustomCover(series.SID, e.EID))
		coverC, it, err := s.GetEntryCover(series.SID, e.EID)
		require.Nil(t, err)
		require.True(t, len(coverC) > 0)
		require.Equal(t, coverA, coverC)
		require.NotEqual(t, image.Invalid, it)
	}

	mustCloseStore(t, s)
}

func testGetEntryCoverThumbnail(t *testing.T) {
	s := mustOpenStoreMem(t)

	series := parsedData[0].s
	require.Nil(t, s.AddSeries(series, parsedData[0].e))

	for _, e := range parsedData[0].e {
		// Keep track of the thumbnail of the original cover
		thumbA, it, err := s.GetEntryThumbnail(series.SID, e.EID)
		require.Nil(t, err)
		require.True(t, len(thumbA) > 0)
		require.Equal(t, image.JPEG, it)

		// Cannot set a nil cover
		err = s.SetEntryCover(series.SID, e.EID, "c.png", nil)
		require.NotNil(t, err)
		require.ErrorIs(t, ErrInvalidCover, err)

		// Can set a custom cover
		require.Nil(t, s.SetEntryCover(series.SID, e.EID, "c.png", customCover))
		cover, it, err := s.GetEntryCover(series.SID, e.EID)
		require.Nil(t, err)
		require.True(t, len(cover) > 0)
		require.Equal(t, customCover, cover)
		require.Equal(t, image.PNG, it)

		// Thumbnail of the cover should not be of the normal cover
		thumbB, it, err := s.GetEntryThumbnail(series.SID, e.EID)
		require.Nil(t, err)
		require.True(t, len(thumbB) > 0)
		require.NotEqual(t, thumbA, thumbB)
		require.Equal(t, image.JPEG, it)

		// If we access the thumbnail again it persists
		thumbC, it, err := s.GetEntryThumbnail(series.SID, e.EID)
		require.Nil(t, err)
		require.True(t, len(thumbC) > 0)
		require.Equal(t, thumbB, thumbC)
		require.Equal(t, image.JPEG, it)

		// Once the custom cover gets deleted thumbnail goes back to normal
		require.Nil(t, s.DeleteEntryCustomCover(series.SID, e.EID))
		thumbD, it, err := s.GetEntryThumbnail(series.SID, e.EID)
		require.Nil(t, err)
		require.True(t, len(thumbD) > 0)
		require.Equal(t, thumbA, thumbD)
		require.NotEqual(t, thumbB, thumbD)
		require.Equal(t, image.JPEG, it)
	}

	mustCloseStore(t, s)
}

func TestStore_GetEntryCover(t *testing.T) {
	testGetSetDeleteEntryCover(t)
}

func TestStore_SetEntryCover(t *testing.T) {
	testGetEntryCoverThumbnail(t)
}

func TestStore_DeleteEntryCustomCover(t *testing.T) {
	testGetSetDeleteEntryCover(t)
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
				r, size, err := e.Archive.ReaderForFile(p.Path)
				require.Nil(t, err)
				require.True(t, size > 0)
				originalData, err := ioutil.ReadAll(r)
				require.Nil(t, err)
				require.True(t, len(originalData) > 0)

				// Get the 0-indexed page from the store
				r, size, it, err := s.GetPage(d.s.SID, e.EID, i, true)
				require.Nil(t, err)
				require.True(t, size > 0)
				zeroData, err := ioutil.ReadAll(r)
				require.Nil(t, err)
				require.True(t, len(zeroData) > 0)
				require.Equal(t, originalData, zeroData)
				require.NotEqual(t, image.Invalid, it)

				// Get the 1-indexed page from the store
				r, size, it, err = s.GetPage(d.s.SID, e.EID, i+1, false)
				require.Nil(t, err)
				require.True(t, size > 0)
				oneData, err := ioutil.ReadAll(r)
				require.Nil(t, err)
				require.True(t, len(oneData) > 0)
				require.Equal(t, originalData, oneData)
				require.NotEqual(t, image.Invalid, it)
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
