package db

import (
	bolt "go.etcd.io/bbolt"
	"time"

	"github.com/fiwippi/tanuki/internal/sets"
	"github.com/fiwippi/tanuki/pkg/api"
	"github.com/fiwippi/tanuki/pkg/auth"
	"github.com/fiwippi/tanuki/pkg/core"
)

type SeriesBucket struct {
	*bolt.Bucket
}

// Entry edit

func (b *SeriesBucket) getEntry(eid []byte) *EntryBucket {
	// The catalog bucket should be used to access entries,
	// this is a private helper function
	bucket := b.Bucket.Bucket(keyEntriesData).Bucket(eid)
	if bucket == nil {
		return nil
	}
	return &EntryBucket{bucket}
}

func (b *SeriesBucket) AddEntry(e *core.ParsedEntry, order int) error {
	entriesBucket := b.Bucket.Bucket(keyEntriesData)

	// Creating bucket for the new manga entry
	eid := auth.SHA1(e.Archive.Title)
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
		if err := entryBucket.SetCover(&core.Cover{}); err != nil {
			return err
		}
	}
	// Set the order of the entry
	if err := entryBucket.SetOrder(order); err != nil {
		return err
	}

	// Set the pages
	tempBucket, err = entryBucket.CreateBucketIfNotExists(keyPages)
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

	// Delete it from the metadata
	m := b.EntriesMetadata()
	if m == nil {
		return ErrEntriesMetadataNotExist
	}
	m[i] = nil
	err := b.SetEntriesMetadata(m)
	if err != nil {
		return err
	}

	// Delete the entry bucket
	err = b.Bucket.Bucket(keyEntriesData).DeleteBucket([]byte(eid))
	if err != nil {
		return err
	}

	// Regenerate the metadata
	return b.regenerateEntriesMetadata()
}

func (b *SeriesBucket) ForEachEntry(f func(hash string, b *EntryBucket) error) error {
	return b.Bucket.Bucket(keyEntriesData).ForEach(func(k, v []byte) error {
		if v == nil {
			err := f(string(k), b.getEntry(k))
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (b *SeriesBucket) regenerateEntriesMetadata() error {
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
	return b.Bucket.Put(keyTitle, core.MarshalJSON(t))
}

func (b *SeriesBucket) Title() string {
	return core.UnmarshalString(b.Bucket.Get(keyTitle))
}

// Tags

func (b *SeriesBucket) SetTags(t *sets.Set) error {
	return b.Bucket.Put(keyTags, core.MarshalJSON(t))
}

func (b *SeriesBucket) Tags() *sets.Set {
	return core.UnmarshalSet(b.Bucket.Get(keyTags))
}

// Cover

func (b *SeriesBucket) Cover() *core.Cover {
	c := b.Get(keyCover)
	if c == nil {
		return nil
	}
	return core.UnmarshalCover(c)
}

func (b *SeriesBucket) SetCover(c *core.Cover) error {
	return b.Put(keyCover, core.MarshalJSON(c))
}

// Thumbnail

func (b *SeriesBucket) Thumbnail() []byte {
	return b.Get(keyThumbnail)
}

func (b *SeriesBucket) HasThumbnail() bool {
	return len(b.Thumbnail()) > 0
}

func (b *SeriesBucket) SetThumbnail(thumb []byte) error {
	return b.Put(keyThumbnail, thumb)
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
	em := b.Bucket.Get(keyEntriesMetadata)
	if em == nil {
		return nil
	}
	return api.UnmarshalEntries(em)
}

func (b *SeriesBucket) SetEntriesMetadata(d api.Entries) error {
	return b.Bucket.Put(keyEntriesMetadata, core.MarshalJSON(d))
}

// Order

func (b *SeriesBucket) SetOrder(o int) error {
	return b.Put(keyOrder, core.MarshalJSON(o))
}

func (b *SeriesBucket) Order() int {
	return core.UnmarshalOrder(b.Get(keyOrder))
}

// ModTime

func (b *SeriesBucket) SetModTime(t time.Time) error {
	return b.Put(keyModTime, core.MarshalJSON(t))
}

func (b *SeriesBucket) ModTime() time.Time {
	return core.UnmarshalTime(b.Get(keyModTime))
}
