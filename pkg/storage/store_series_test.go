package storage

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/fiwippi/tanuki/internal/platform/dbutil"
	"github.com/fiwippi/tanuki/internal/platform/image"
	"github.com/fiwippi/tanuki/pkg/manga"
)

func equalSeries(t *testing.T, s1, s2 *manga.Series) {
	require.NotNil(t, s1)
	require.NotNil(t, s2)
	require.Equal(t, s1.FolderTitle, s2.FolderTitle)
	require.Equal(t, s1.SID, s2.SID)
	require.Equal(t, s1.NumEntries, s2.NumEntries)
	require.Equal(t, s1.NumPages, s2.NumPages)
	require.Equal(t, s1.DisplayTitle, s2.DisplayTitle)
	require.Equal(t, s1.Tags, s2.Tags)
	require.True(t, s1.ModTime.Equal(s2.ModTime))
}

func equalEntries(t *testing.T, e1, e2 []*manga.Entry) {
	require.NotNil(t, e1)
	require.NotNil(t, e2)
	require.Equal(t, len(e1), len(e2))
	for i := range e1 {
		require.Equal(t, e1[i].SID, e2[i].SID)
		require.Equal(t, e1[i].EID, e2[i].EID)
		require.Equal(t, e1[i].Title, e2[i].Title)
		require.Equal(t, e1[i].Archive, e2[i].Archive)
		require.Equal(t, e1[i].Pages, e2[i].Pages)
		require.Equal(t, e1[i].DisplayTile, e2[i].DisplayTile)
		require.True(t, e1[i].ModTime.Equal(e2[i].ModTime))
	}
}

// Core

func TestStore_HasSeries(t *testing.T) {
	s := mustOpenStoreMem(t)

	for i, d := range parsedData {
		require.Nil(t, s.AddSeries(parsedData[i].s, parsedData[i].e))
		has, err := s.HasSeries(d.s.SID)
		require.Nil(t, err)
		require.True(t, has)
	}

	mustCloseStore(t, s)
}

func TestStore_AddSeries(t *testing.T) {
	s := mustOpenStoreMem(t)

	for i, d := range parsedData {
		require.Nil(t, s.AddSeries(parsedData[i].s, parsedData[i].e))

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

	mustCloseStore(t, s)
}

func TestStore_GetSeries(t *testing.T) {
	t.Run("GetSingleSeries", func(t *testing.T) {
		s := mustOpenStoreMem(t)
		defer mustCloseStore(t, s)

		require.Nil(t, s.AddSeries(parsedData[0].s, parsedData[0].e))
		dbSeries, err := s.GetSeries(parsedData[0].s.SID)
		require.Nil(t, err)
		equalSeries(t, parsedData[0].s, dbSeries)
	})

	t.Run("PreserveSeriesOrderAfterDelete", func(t *testing.T) {
		s := mustOpenStoreMem(t)
		defer mustCloseStore(t, s)

		for _, d := range parsedData {
			require.Nil(t, s.AddSeries(d.s, d.e))
		}

		require.Nil(t, s.DeleteSeries(parsedData[0].s.SID))
		require.Nil(t, s.DeleteSeries(parsedData[2].s.SID))
		require.Nil(t, s.AddSeries(parsedData[0].s, parsedData[0].e))
		require.Nil(t, s.DeleteSeries(parsedData[1].s.SID))
		require.Nil(t, s.AddSeries(parsedData[2].s, parsedData[2].e))
		require.Nil(t, s.AddSeries(parsedData[1].s, parsedData[1].e))

		ctl, err := s.GetCatalog()
		require.Nil(t, err)
		require.Equal(t, 3, len(ctl))

		equalSeries(t, parsedData[0].s, ctl[0])
		equalSeries(t, parsedData[1].s, ctl[1])
		equalSeries(t, parsedData[2].s, ctl[2])
	})
}

func TestStore_DeleteSeries(t *testing.T) {
	s := mustOpenStoreMem(t)

	for _, d := range parsedData {
		require.Nil(t, s.AddSeries(d.s, d.e))
		require.Nil(t, s.DeleteSeries(d.s.SID))
	}

	ctl, err := s.GetCatalog()
	require.Nil(t, err)
	require.Equal(t, 0, len(ctl))

	mustCloseStore(t, s)
}

// Cover / Thumbnail

func testGetSetDeleteSeriesCover(t *testing.T) {
	s := mustOpenStoreMem(t)

	series := parsedData[0].s
	entries := parsedData[0].e
	require.Nil(t, s.AddSeries(series, entries))

	// Get the normal cover
	coverA, aType, err := s.GetSeriesCover(series.SID)
	require.Nil(t, err)
	require.True(t, len(coverA) > 0)
	require.Equal(t, image.JPEG, aType)

	// Cannot set invalid custom cover
	require.ErrorIs(t, s.SetSeriesCover(series.SID, "c.png", nil), ErrInvalidCover)
	require.ErrorIs(t, s.SetSeriesCover(series.SID, "c.aaa", customCover), ErrInvalidCover)

	// Get custom cover works if it exists
	require.Nil(t, s.SetSeriesCover(series.SID, "c.png", customCover))
	coverB, bType, err := s.GetSeriesCover(series.SID)
	require.Nil(t, err)
	require.True(t, len(coverB) > 0)
	require.NotEqual(t, coverA, coverB)
	require.Equal(t, image.PNG, bType)

	// Delete the cover we should have the normal series cover
	require.Nil(t, s.DeleteSeriesCustomCover(series.SID))
	coverC, cType, err := s.GetSeriesCover(series.SID)
	require.Nil(t, err)
	require.True(t, len(coverC) > 0)
	require.Equal(t, coverA, coverC)
	require.Equal(t, aType, cType)

	mustCloseStore(t, s)
}

func testGetSeriesCoverThumbnail(t *testing.T) {
	s := mustOpenStoreMem(t)

	series := parsedData[0].s
	require.Nil(t, s.AddSeries(series, parsedData[0].e))

	// Keep track of the thumbnail of the original cover
	thumbA, it, err := s.GetSeriesThumbnail(series.SID)
	require.Nil(t, err)
	require.True(t, len(thumbA) > 0)
	require.Equal(t, image.JPEG, it)

	// Cannot set a nil cover
	err = s.SetSeriesCover(series.SID, "c.png", nil)
	require.NotNil(t, err)
	require.ErrorIs(t, ErrInvalidCover, err)

	// Can set a custom cover
	require.Nil(t, s.SetSeriesCover(series.SID, "c.png", customCover))
	cover, it, err := s.GetSeriesCover(series.SID)
	require.Nil(t, err)
	require.True(t, len(cover) > 0)
	require.Equal(t, customCover, cover)
	require.Equal(t, image.PNG, it)

	// Thumbnail of the cover should not be of the normal cover
	thumbB, it, err := s.GetSeriesThumbnail(series.SID)
	require.Nil(t, err)
	require.True(t, len(thumbB) > 0)
	require.NotEqual(t, thumbA, thumbB)
	require.Equal(t, image.JPEG, it)

	// If we access the thumbnail again it persists
	thumbC, it, err := s.GetSeriesThumbnail(series.SID)
	require.Nil(t, err)
	require.True(t, len(thumbC) > 0)
	require.Equal(t, thumbB, thumbC)
	require.Equal(t, image.JPEG, it)

	// Once the custom cover gets deleted thumbnail goes back to normal
	require.Nil(t, s.DeleteSeriesCustomCover(series.SID))
	thumbD, it, err := s.GetSeriesThumbnail(series.SID)
	require.Nil(t, err)
	require.True(t, len(thumbD) > 0)
	require.Equal(t, thumbA, thumbD)
	require.NotEqual(t, thumbB, thumbD)
	require.Equal(t, image.JPEG, it)

	mustCloseStore(t, s)
}

func TestStore_GetSeriesCover(t *testing.T) {
	testGetSetDeleteSeriesCover(t)
}

func TestStore_SetSeriesCover(t *testing.T) {
	testGetSeriesCoverThumbnail(t)
}

func TestStore_DeleteSeriesCustomCover(t *testing.T) {
	testGetSetDeleteSeriesCover(t)
}

func TestStore_GetSeriesThumbnail(t *testing.T) {
	testGetSeriesCoverThumbnail(t)
}

// Tags / Metadata

func TestStore_SetSeriesTags(t *testing.T) {
	s := mustOpenStoreMem(t)

	series := parsedData[0].s
	require.Nil(t, s.AddSeries(series, parsedData[0].e))

	tags := manga.NewTags()
	tags.Add("Red", "Green", "Blue")
	require.Nil(t, s.SetSeriesTags(series.SID, tags))

	dbSeries, err := s.GetSeries(series.SID)
	require.Nil(t, err)
	require.NotNil(t, dbSeries.Tags)
	require.Equal(t, tags, dbSeries.Tags)

	mustCloseStore(t, s)
}

func TestStore_GetSeriesWithTag(t *testing.T) {
	s := mustOpenStoreMem(t)

	require.Nil(t, s.AddSeries(parsedData[0].s, parsedData[0].e))
	tagsA := manga.NewTags()
	tagsA.Add("Red")
	require.Nil(t, s.SetSeriesTags(parsedData[0].s.SID, tagsA))

	require.Nil(t, s.AddSeries(parsedData[1].s, parsedData[1].e))
	tagsB := manga.NewTags()
	tagsB.Add("Green")
	require.Nil(t, s.SetSeriesTags(parsedData[1].s.SID, tagsB))

	ctl, err := s.GetSeriesWithTag("Red")
	require.Nil(t, err)
	require.NotNil(t, ctl)
	require.True(t, len(ctl) == 1)
	require.Equal(t, tagsA, ctl[0].Tags)

	mustCloseStore(t, s)
}

func TestStore_GetAllTags(t *testing.T) {
	s := mustOpenStoreMem(t)

	final := manga.NewTags()
	final.Add("Red", "Green", "Blue")

	require.Nil(t, s.AddSeries(parsedData[0].s, parsedData[0].e))
	tagsA := manga.NewTags()
	tagsA.Add("Red")
	require.Nil(t, s.SetSeriesTags(parsedData[0].s.SID, tagsA))

	require.Nil(t, s.AddSeries(parsedData[1].s, parsedData[1].e))
	tagsB := manga.NewTags()
	tagsB.Add("Green")
	require.Nil(t, s.SetSeriesTags(parsedData[1].s.SID, tagsB))

	require.Nil(t, s.AddSeries(parsedData[2].s, parsedData[2].e))
	tagsC := manga.NewTags()
	tagsC.Add("Blue")
	require.Nil(t, s.SetSeriesTags(parsedData[2].s.SID, tagsC))

	all, err := s.GetAllTags()
	require.Nil(t, err)
	require.NotNil(t, all)
	require.Equal(t, final, all)

	mustCloseStore(t, s)
}

func TestStore_SetSeriesDisplayTitle(t *testing.T) {
	s := mustOpenStoreMem(t)

	series := parsedData[0].s
	require.Nil(t, s.AddSeries(series, parsedData[0].e))

	require.Nil(t, s.SetSeriesDisplayTitle(series.SID, "HUH"))
	dbSeries, err := s.GetSeries(series.SID)
	require.Nil(t, err)
	require.Equal(t, dbutil.NullString("HUH"), dbSeries.DisplayTitle)

	require.Nil(t, s.SetSeriesDisplayTitle(series.SID, ""))
	dbSeries, err = s.GetSeries(series.SID)
	require.Nil(t, err)
	require.Equal(t, dbutil.NullString(""), dbSeries.DisplayTitle)

	mustCloseStore(t, s)
}
