package db

import (
	"time"

	bolt "go.etcd.io/bbolt"

	"github.com/fiwippi/tanuki/internal/sets"
	"github.com/fiwippi/tanuki/pkg/api"
	"github.com/fiwippi/tanuki/pkg/auth"
	"github.com/fiwippi/tanuki/pkg/core"
)

var (
	keySeriesData      = []byte("data") // Holds state like the series order, is served to the user so they can request specific data
	keySeriesTags      = []byte("tags")
	keySeriesTitle     = []byte("title")
	keySeriesEntries   = []byte("entries")
	keySeriesMetadata  = []byte("metadata")
	keySeriesCover     = []byte("cover")
	keySeriesThumbnail = []byte("thumbnail")
)

type SeriesBucket struct {
	*bolt.Bucket
}

func (b *SeriesBucket) AddEntry(e *core.Manga) error {
	entriesBucket := b.Bucket.Bucket(keySeriesEntries)

	// Creating bucket for the new manga entry
	mangaHash := auth.HashSHA1(e.Title)
	emptyMangaBucket, err := entriesBucket.CreateBucketIfNotExists([]byte(mangaHash))
	if err != nil {
		return err
	}
	mangaBucket := &MangaBucket{emptyMangaBucket}

	// Set basic manga data
	if err := mangaBucket.SetArchive(e.Archive); err != nil {
		return err
	} else if err := mangaBucket.SetMetadata(e.Metadata); err != nil {
		return err
	} else if err := mangaBucket.SetTitle(e.Title); err != nil {
		return err
	}
	// If cover is nil then set it to a default
	cover := mangaBucket.Cover()
	if cover == nil {
		if err := mangaBucket.SetCover(&core.Cover{}); err != nil {
			return err
		}
	}

	// Set the pages
	emptyPagesBucket, err := mangaBucket.CreateBucketIfNotExists(keyMangaPages)
	if err != nil {
		return err
	}
	pagesBucket := &PagesBucket{emptyPagesBucket}

	for i, p := range e.Pages {
		// Page indexing should start at 1
		err := pagesBucket.SetPage(i+1, p)
		if err != nil {
			return err
		}
	}

	// Set the file info
	return nil
}

func (b *SeriesBucket) GetEntry(entryHashBytes []byte) *MangaBucket {
	bucket := b.Bucket.Bucket(keySeriesEntries).Bucket(entryHashBytes)
	if bucket == nil {
		return nil
	}
	return &MangaBucket{bucket}
}

func (b *SeriesBucket) DeleteEntry(entryHashBytes []byte) error {
	return b.Bucket.Bucket(keySeriesEntries).DeleteBucket(entryHashBytes)
}

func (b *SeriesBucket) Title() string {
	return core.UnmarshalString(b.Bucket.Get(keySeriesTitle))
}

func (b *SeriesBucket) Tags() *sets.Set {
	return core.UnmarshalSet(b.Bucket.Get(keySeriesTags))
}

func (b *SeriesBucket) Data() api.SeriesEntries {
	return api.UnmarshalSeriesEntries(b.Bucket.Get(keySeriesData))
}

func (b *SeriesBucket) Metadata() *core.SeriesMetadata {
	m := core.UnmarshalSeriesMetadata(b.Bucket.Get(keySeriesMetadata))
	if m != nil && m.DateReleased == nil {
		m.DateReleased = core.NewDate(time.Time{})
	}

	return m
}

func (b *SeriesBucket) Cover() *core.Cover {
	c := b.Get(keySeriesCover)
	if c == nil {
		return nil
	}
	return core.UnmarshalCover(c)
}

//func (b *SeriesBucket) ArchiveCoverBytes() ([]byte, error) {
//	cover := b.Cover()
//	if cover == nil {
//		return nil, nil
//	}
//	if cover.Fp == "" {
//		return nil, nil
//	}
//	data, err := cover.FromFS()
//	if err != nil {
//		return nil, err
//	}
//	return data, nil
//}

func (b *SeriesBucket) HasThumbnail() bool {
	return len(b.Thumbnail()) > 0
}

func (b *SeriesBucket) Thumbnail() []byte {
	return b.Get(keySeriesThumbnail)
}

func (b *SeriesBucket) ApiSeries() *api.Series {
	m := b.Metadata()

	return &api.Series{
		Hash:         auth.HashSHA1(b.Title()),
		Title:        m.Title,
		Entries:      len(b.Data()),
		Tags:         b.Tags().List(),
		Author:       m.Author,
		DateReleased: m.DateReleased,
	}
}

func (b *SeriesBucket) SetTitle(t string) error {
	return b.Bucket.Put(keySeriesTitle, core.MarshalJSON(t))
}

func (b *SeriesBucket) SetTags(t *sets.Set) error {
	return b.Bucket.Put(keySeriesTags, core.MarshalJSON(t))
}

func (b *SeriesBucket) SetData(d api.SeriesEntries) error {
	return b.Bucket.Put(keySeriesData, core.MarshalJSON(d))
}

func (b *SeriesBucket) SetMetadata(d *core.SeriesMetadata) error {
	// Ensure the date is not nil
	if d != nil && d.DateReleased == nil {
		d.DateReleased = core.NewDate(time.Time{})
	}

	return b.Bucket.Put(keySeriesMetadata, core.MarshalJSON(d))
}

func (b *SeriesBucket) SetCover(c *core.Cover) error {
	return b.Put(keySeriesCover, core.MarshalJSON(c))
}

func (b *SeriesBucket) SetThumbnail(thumb []byte) error {
	return b.Put(keySeriesThumbnail, thumb)
}

func (b *SeriesBucket) ForEachEntry(f func(hash string, b *MangaBucket) error) error {
	return b.Bucket.Bucket(keySeriesEntries).ForEach(func(k, v []byte) error {
		if v == nil {
			err := f(string(k), b.GetEntry(k))
			if err != nil {
				return err
			}
		}
		return nil
	})
}
