package db

import (
	"github.com/fiwippi/tanuki/pkg/api"
	bolt "go.etcd.io/bbolt"

	"github.com/fiwippi/tanuki/pkg/core"
)

type EntryBucket struct {
	*bolt.Bucket
}

func (b *EntryBucket) PagesBucket() *PagesBucket {
	bucket := b.Bucket.Bucket(keyPages)
	if bucket == nil {
		return nil
	}
	return &PagesBucket{bucket}
}

func (b *EntryBucket) SetCover(c *core.Cover) error {
	return b.Put(keyCover, core.MarshalJSON(c))
}

func (b *EntryBucket) Cover() *core.Cover {
	return core.UnmarshalCover(b.Get(keyCover))
}

func (b *EntryBucket) SetArchive(a *core.Archive) error {
	return b.Put(keyArchive, core.MarshalJSON(a))
}

func (b *EntryBucket) Archive() *core.Archive {
	return core.UnmarshalArchive(b.Get(keyArchive))
}

func (b *EntryBucket) SetThumbnail(thumb []byte) error {
	return b.Put(keyThumbnail, thumb)
}

func (b *EntryBucket) HasThumbnail() bool {
	return len(b.Thumbnail()) > 0
}

func (b *EntryBucket) Thumbnail() []byte {
	return b.Get(keyThumbnail)
}

func (b *EntryBucket) SetOrder(o int) error {
	return b.Put(keyOrder, core.MarshalJSON(o))
}

func (b *EntryBucket) Order() int {
	return core.UnmarshalOrder(b.Get(keyOrder))
}

// Personal Metadata

func (b *EntryBucket) Metadata() *api.EditableEntryMetadata {
	m := b.Bucket.Get(keyMetadata)
	if m == nil {
		return nil
	}
	return api.UnmarshalEditableEntryMetadata(m)
}

func (b *EntryBucket) SetMetadata(d *api.EditableEntryMetadata) error {
	return b.Bucket.Put(keyMetadata, core.MarshalJSON(d))
}
