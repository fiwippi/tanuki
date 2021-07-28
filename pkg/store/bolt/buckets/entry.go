package buckets

import (
	"github.com/fiwippi/tanuki/internal/json"
	"github.com/fiwippi/tanuki/pkg/store/bolt/keys"
	"github.com/fiwippi/tanuki/pkg/store/entities/manga"
	bolt "go.etcd.io/bbolt"
)

type EntryBucket struct {
	*bolt.Bucket
}

func (b *EntryBucket) PagesBucket() *PagesBucket {
	bucket := b.Bucket.Bucket(keys.Pages)
	if bucket == nil {
		return nil
	}
	return &PagesBucket{bucket}
}

func (b *EntryBucket) SetCover(c *manga.Cover) error {
	return b.Put(keys.Cover, json.Marshal(c))
}

func (b *EntryBucket) Cover() *manga.Cover {
	return manga.UnmarshalCover(b.Get(keys.Cover))
}

func (b *EntryBucket) SetArchive(a *manga.Archive) error {
	return b.Put(keys.Archive, json.Marshal(a))
}

func (b *EntryBucket) Archive() *manga.Archive {
	return manga.UnmarshalArchive(b.Get(keys.Archive))
}

func (b *EntryBucket) SetThumbnail(thumb []byte) error {
	return b.Put(keys.Thumbnail, thumb)
}

func (b *EntryBucket) HasThumbnail() bool {
	return len(b.Thumbnail()) > 0
}

func (b *EntryBucket) Thumbnail() []byte {
	return b.Get(keys.Thumbnail)
}

func (b *EntryBucket) SetOrder(o int) error {
	return b.Put(keys.Order, json.Marshal(o))
}

func (b *EntryBucket) Order() int {
	d := b.Get(keys.Order)
	if d == nil {
		return -1
	}
	return json.UnmarshalInt(d)
}

// Personal Metadata

func (b *EntryBucket) Metadata() *manga.EntryMetadata {
	m := b.Bucket.Get(keys.Metadata)
	if m == nil {
		return nil
	}
	return manga.UnmarshalEntryMetadata(m)
}

func (b *EntryBucket) SetMetadata(d *manga.EntryMetadata) error {
	return b.Bucket.Put(keys.Metadata, json.Marshal(d))
}
