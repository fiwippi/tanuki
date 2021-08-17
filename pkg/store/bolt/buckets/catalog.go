package buckets

import (
	"time"

	bolt "go.etcd.io/bbolt"

	"github.com/fiwippi/tanuki/internal/errors"
	"github.com/fiwippi/tanuki/internal/hash"
	"github.com/fiwippi/tanuki/internal/json"
	"github.com/fiwippi/tanuki/internal/sets"
	"github.com/fiwippi/tanuki/pkg/store/bolt/keys"
	"github.com/fiwippi/tanuki/pkg/store/entities/api"
	"github.com/fiwippi/tanuki/pkg/store/entities/manga"
)

var (
	ErrCatalogNotExist      = errors.New("catalog does not exist")
	ErrCatalogEntryNotExist = errors.New("catalog entry does not exist")
	ErrSeriesNotExist       = errors.New("series does not exist")
)

type CatalogBucket struct {
	*bolt.Bucket
}

func (b *CatalogBucket) Series(sid string) (*SeriesBucket, error) {
	bucket := b.Bucket.Bucket([]byte(sid))
	if bucket == nil {
		return nil, ErrSeriesNotExist.Fmt(sid)
	}

	return &SeriesBucket{Bucket: bucket}, nil
}

func (b *CatalogBucket) Entry(sid, eid string) (*EntryBucket, error) {
	seriesBucket, err := b.Series(sid)
	if err != nil {
		return nil, err
	}

	entryBucket := seriesBucket.getEntry([]byte(eid))
	if entryBucket == nil {
		return nil, ErrEntryNotExist.Fmt(eid)
	}

	return entryBucket, nil
}

func (b *CatalogBucket) FirstEntry(sid string) (*EntryBucket, error) {
	seriesBucket, err := b.Series(sid)
	if err != nil {
		return nil, err
	}

	eid := seriesBucket.EntriesMetadata()[0].Hash
	entryBucket, err := b.Entry(sid, eid)
	if err != nil {
		return nil, err
	}

	return entryBucket, nil
}

func (b *CatalogBucket) AddSeries(s *manga.ParsedSeries) error {
	if len(s.Entries) == 0 {
		return ErrEntriesNotExist.Fmt(s.Title)
	}

	// Create the bucket for the series
	sid := hash.SHA1(s.Title)
	tempBucket, err := b.Bucket.CreateBucketIfNotExists([]byte(sid))
	if err != nil {
		return err
	}
	seriesBucket := &SeriesBucket{tempBucket}

	// Set the series title
	if err := seriesBucket.SetTitle(s.Title); err != nil {
		return err
	}
	// Only set series tags if one doesn't already exist
	if t := seriesBucket.Tags(); t == nil {
		if err := seriesBucket.SetTags(sets.NewSet()); err != nil {
			return err
		}
	}
	// Create new cover entry if it doesn't exist
	cover := seriesBucket.Cover()
	if cover == nil {
		if err := seriesBucket.SetCover(&manga.Cover{}); err != nil {
			return err
		}
	}

	// Ensure the bucket for entries exists
	_, err = seriesBucket.CreateBucketIfNotExists(keys.EntriesData)
	if err != nil {
		return err
	}
	//
	seriesData := make(api.Entries, len(s.Entries))
	seriesModTime := s.Entries[0].Archive.ModTime
	for i, m := range s.Entries {
		eid := hash.SHA1(m.Archive.Title)

		// If the archive file has changed then delete the currently stored entry
		if e := seriesBucket.getEntry([]byte(eid)); e != nil {
			if !e.Archive().ModTime.Equal(m.Archive.ModTime) {
				err := seriesBucket.DeleteEntry(eid)
				if err != nil {
					return err
				}
			}
		}

		err := seriesBucket.AddEntry(m, m.Order)
		if err != nil {
			return err
		}

		e := &api.Entry{
			Order:        m.Order,
			Hash:         eid,
			Title:        m.Archive.Title,
			Pages:        len(m.Pages),
			Path:         m.Archive.Path,
			Chapter:      m.Metadata.Chapter,
			Volume:       m.Metadata.Volume,
			Author:       m.Metadata.Author,
			DateReleased: m.Metadata.DateReleased,
		}

		eb, err := b.Entry(sid, eid)
		if err != nil {
			return err
		}
		metadata := eb.Metadata()
		if metadata != nil {
			if metadata.Title != manga.TitleZeroValue {
				e.Title = metadata.Title
			}
			if metadata.Author != manga.AuthorZeroValue {
				e.Author = metadata.Author
			}
			if metadata.DateReleased != nil && metadata.DateReleased.Time != manga.TimeZeroValue {
				e.DateReleased = metadata.DateReleased
			}
			if metadata.Chapter != manga.ChapterZeroValue {
				e.Chapter = metadata.Chapter
			}
			if metadata.Volume != manga.VolumeZeroValue {
				e.Volume = metadata.Volume
			}
		}
		seriesData[i] = e

		err = eb.SetMetadata(&manga.EntryMetadata{
			Title:        e.Title,
			Author:       e.Author,
			DateReleased: e.DateReleased,
			Chapter:      e.Chapter,
			Volume:       e.Volume,
		})
		if err != nil {
			return err
		}

		if m.Archive.ModTime.Before(seriesModTime) {
			seriesModTime = m.Archive.ModTime
		}
	}

	if err := seriesBucket.SetModTime(seriesModTime); err != nil {
		return err
	}

	if err := seriesBucket.SetEntriesMetadata(seriesData); err != nil {
		return err
	}

	return nil
}

func (b *CatalogBucket) DeleteSeries(sid string) error {
	sb, err := b.Series(sid)
	if err != nil {
		return err
	}

	// Get index of the series data in the metadata
	i := sb.Order() - 1

	// Delete it from the metadata if it exists in the metadata
	c := b.Catalog()
	if c == nil {
		return ErrCatalogNotExist
	}
	if i >= 0 && i < len(c) && c[i].Hash == sid {
		c[i] = nil
		err = b.SetCatalog(c)
		if err != nil {
			return err
		}
	}

	// Delete the series bucket
	err = b.Bucket.DeleteBucket([]byte(sid))
	if err != nil {
		return err
	}

	// Regenerate the catalog
	return b.RegenerateCatalog()
}

func (b *CatalogBucket) ForEachSeries(f func(sid string, b *SeriesBucket) error) error {
	return b.Bucket.ForEach(func(k, v []byte) error {
		if v == nil {
			s, _ := b.Series(string(k))
			err := f(string(k), s)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// Metadata about the the catalog

func (b *CatalogBucket) SeriesMetadata(sid string) (*api.Series, error) {
	sb, err := b.Series(sid)
	if err != nil {
		return nil, ErrSeriesNotExist.Fmt(sid)
	}
	i := sb.Order() - 1

	c := b.Catalog()
	if i >= len(c) {
		return nil, ErrCatalogEntryNotExist.Fmt(sid)
	}

	return c[i], nil
}

func (b *CatalogBucket) SetSeriesMetadata(sid string, s *api.Series) error {
	sb, err := b.Series(sid)
	if err != nil {
		return ErrSeriesNotExist.Fmt(sid)
	}
	i := sb.Order() - 1

	c := b.Catalog()
	if i >= len(c) {
		return ErrCatalogEntryNotExist.Fmt(sid)
	}
	c[i] = s

	return b.SetCatalog(c)
}

func (b *CatalogBucket) RegenerateCatalog() error {
	oldC := b.Catalog()
	newC := make(api.Catalog, 0)

	o := 1
	for _, s := range oldC {
		if s != nil {
			sb, err := b.Series(s.Hash)
			if err != nil {
				return err
			}

			err = sb.SetOrder(o)
			if err != nil {
				return err
			}

			s.Order = o
			s.Entries = len(sb.EntriesMetadata())
			newC = append(newC, s)
			o++
		}
	}

	return b.SetCatalog(newC)
}

func (b *CatalogBucket) SetCatalog(c api.Catalog) error {
	return b.Put(keys.Catalog, json.Marshal(c))
}

func (b *CatalogBucket) Catalog() api.Catalog {
	c := b.Get(keys.Catalog)
	if c == nil {
		return nil
	}
	return api.UnmarshalCatalog(c)
}

// ModTime

func (b *CatalogBucket) SetModTime(t time.Time) error {
	return b.Put(keys.ModTime, json.Marshal(t))
}

func (b *CatalogBucket) ModTime() time.Time {
	return json.UnmarshalTime(b.Get(keys.ModTime))
}
