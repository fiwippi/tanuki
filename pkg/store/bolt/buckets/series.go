package buckets

import (
	"errors"
	"github.com/fiwippi/tanuki/internal/hash"
	"github.com/fiwippi/tanuki/internal/json"
	"github.com/fiwippi/tanuki/pkg/store/bolt/keys"
	"github.com/fiwippi/tanuki/pkg/store/entities/api"
	"github.com/fiwippi/tanuki/pkg/store/entities/manga"
	bolt "go.etcd.io/bbolt"
	"time"

	"github.com/fiwippi/tanuki/internal/sets"
)

var (
	ErrEntryNotExist           = errors.New("entry does not exist")
	ErrEntryMetadataNotExist   = errors.New("entry metadata does not exist")
	ErrEntriesMetadataNotExist = errors.New("entries metadata does not exist")
)

type SeriesBucket struct {
	*bolt.Bucket
}

// Entry edit

func (b *SeriesBucket) getEntry(eid []byte) *EntryBucket {
	// The catalog bucket should be used to access entries,
	// this is a private helper function
	bucket := b.Bucket.Bucket(keys.EntriesData).Bucket(eid)
	if bucket == nil {
		return nil
	}
	return &EntryBucket{bucket}
}

func (b *SeriesBucket) AddEntry(e *manga.ParsedEntry, order int) error {
	entriesBucket := b.Bucket.Bucket(keys.EntriesData)

	// Creating bucket for the new manga entry
	eid := hash.SHA1(e.Archive.Title)
	tempBucket, err := entriesBucket.CreateBucketIfNotExists([]byte(eid))
	if err != nil {
		return err
	}
	entryBucket := &EntryBucket{tempBucket}

	// Set basic manga data
	if err := entryBucket.SetArchive(e.Archive); err != nil {
		return err
	}
	// If cover is nil then set it to a default
	cover := entryBucket.Cover()
	if cover == nil {
		if err := entryBucket.SetCover(&manga.Cover{}); err != nil {
			return err
		}
	}
	// Set the order of the entry
	if err := entryBucket.SetOrder(order); err != nil {
		return err
	}

	// Set the pages
	tempBucket, err = entryBucket.CreateBucketIfNotExists(keys.Pages)
	if err != nil {
		return err
	}
	pagesBucket := &PagesBucket{tempBucket}

	for i, p := range e.Pages {
		// Page indexing should start at 1
		err := pagesBucket.SetPage(i+1, p)
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *SeriesBucket) DeleteEntry(eid string) error {
	eb := b.getEntry([]byte(eid))
	if eb == nil {
		return ErrEntryNotExist
	}

	// Get index of the entry data in the metadata
	i := eb.Order() - 1

	// Delete it from the metadata if it still exists there
	m := b.EntriesMetadata()
	if i >= 0 && i < len(m) && m[i].Hash == eid {
		if m == nil {
			return ErrEntriesMetadataNotExist
		}
		m[i] = nil
		err := b.SetEntriesMetadata(m)
		if err != nil {
			return err
		}
	}

	// Delete the entry bucket
	err := b.Bucket.Bucket(keys.EntriesData).DeleteBucket([]byte(eid))
	if err != nil {
		return err
	}

	// Regenerate the metadata
	return b.RegenerateEntriesMetadata()
}

func (b *SeriesBucket) ForEachEntry(f func(hash string, b *EntryBucket) error) error {
	return b.Bucket.Bucket(keys.EntriesData).ForEach(func(k, v []byte) error {
		if v == nil {
			err := f(string(k), b.getEntry(k))
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (b *SeriesBucket) RegenerateEntriesMetadata() error {
	oldM := b.EntriesMetadata()
	newM := make(api.Entries, 0, len(oldM))

	o := 1
	for _, e := range oldM {
		if e != nil {
			eb := b.getEntry([]byte(e.Hash))
			if eb == nil {
				return ErrEntryNotExist
			}

			err := eb.SetOrder(o)
			if err != nil {
				return err
			}

			e.Order = o
			newM = append(newM, e)
			o++
		}
	}

	return b.SetEntriesMetadata(newM)
}

// Title

func (b *SeriesBucket) SetTitle(t string) error {
	return b.Bucket.Put(keys.Title, json.Marshal(t))
}

func (b *SeriesBucket) Title() string {
	return json.UnmarshalString(b.Bucket.Get(keys.Title))
}

// Tags

func (b *SeriesBucket) SetTags(t *sets.Set) error {
	return b.Bucket.Put(keys.Tags, json.Marshal(t))
}

func (b *SeriesBucket) Tags() *sets.Set {
	return sets.UnmarshalSet(b.Bucket.Get(keys.Tags))
}

// Cover

func (b *SeriesBucket) Cover() *manga.Cover {
	c := b.Get(keys.Cover)
	if c == nil {
		return nil
	}
	return manga.UnmarshalCover(c)
}

func (b *SeriesBucket) SetCover(c *manga.Cover) error {
	return b.Put(keys.Cover, json.Marshal(c))
}

// Thumbnail

func (b *SeriesBucket) Thumbnail() []byte {
	return b.Get(keys.Thumbnail)
}

func (b *SeriesBucket) HasThumbnail() bool {
	return len(b.Thumbnail()) > 0
}

func (b *SeriesBucket) SetThumbnail(thumb []byte) error {
	return b.Put(keys.Thumbnail, thumb)
}

// Personal Metadata

func (b *SeriesBucket) Metadata() *manga.SeriesMetadata {
	m := b.Bucket.Get(keys.Metadata)
	if m == nil {
		return nil
	}
	return manga.UnmarshalSeriesMetadata(m)
}

func (b *SeriesBucket) SetMetadata(d *manga.SeriesMetadata) error {
	return b.Bucket.Put(keys.Metadata, json.Marshal(d))
}

// Entries Metadata

func (b *SeriesBucket) EntryMetadata(eid string) (*api.Entry, error) {
	eb := b.getEntry([]byte(eid))
	if eb == nil {
		return nil, ErrEntryNotExist
	}
	i := eb.Order() - 1

	m := b.EntriesMetadata()
	if i >= len(m) {
		return nil, ErrEntryMetadataNotExist
	}

	return m[i], nil
}

func (b *SeriesBucket) SetEntryMetadata(eid string, m *api.Entry) error {
	eb := b.getEntry([]byte(eid))
	if eb == nil {
		return ErrEntryNotExist
	}
	i := eb.Order() - 1

	em := b.EntriesMetadata()
	if i >= len(em) {
		return ErrEntryMetadataNotExist
	}
	em[i] = m

	return b.SetEntriesMetadata(em)
}

func (b *SeriesBucket) EntriesMetadata() api.Entries {
	em := b.Bucket.Get(keys.EntriesMetadata)
	if em == nil {
		return nil
	}
	return api.UnmarshalEntries(em)
}

func (b *SeriesBucket) SetEntriesMetadata(d api.Entries) error {
	return b.Bucket.Put(keys.EntriesMetadata, json.Marshal(d))
}

// Order

func (b *SeriesBucket) SetOrder(o int) error {
	return b.Put(keys.Order, json.Marshal(o))
}

func (b *SeriesBucket) Order() int {
	d := b.Get(keys.Order)
	if d == nil {
		return -1
	}
	return json.UnmarshalInt(d)
}

// ModTime

func (b *SeriesBucket) SetModTime(t time.Time) error {
	return b.Put(keys.ModTime, json.Marshal(t))
}

func (b *SeriesBucket) ModTime() time.Time {
	return json.UnmarshalTime(b.Get(keys.ModTime))
}
