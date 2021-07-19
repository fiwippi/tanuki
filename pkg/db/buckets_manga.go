package db

import (
	"time"

	bolt "go.etcd.io/bbolt"

	"github.com/fiwippi/tanuki/pkg/api"
	"github.com/fiwippi/tanuki/pkg/auth"
	"github.com/fiwippi/tanuki/pkg/core"
)

var (
	keyMangaTitle     = []byte("title")
	keyMangaArchive   = []byte("archive")
	keyMangaCover     = []byte("cover")
	keyMangaThumbnail = []byte("thumbnail")
	keyMangaPages     = []byte("pages")
	keyMangaMetadata  = []byte("metadata")
)

type MangaBucket struct {
	*bolt.Bucket
}

func (b *MangaBucket) ApiSeriesEntry() *api.SeriesEntry {
	meta := b.Metadata()

	return &api.SeriesEntry{
		Hash:  auth.HashSHA1(b.Title()),
		Title: meta.Title,
		Pages: b.Pages().Num(),
		Path:  b.Archive().Path,
		Chapter: meta.Chapter,
		Volume: meta.Volume,
		Author: meta.Author,
		DateReleased: meta.DateReleased,
	}
}

func (b *MangaBucket) Pages() *PagesBucket {
	bucket := b.Bucket.Bucket(keyMangaPages)
	if bucket == nil {
		return nil
	}
	return &PagesBucket{bucket}
}

func (b *MangaBucket) SetArchive(a *core.Archive) error {
	return b.Put(keyMangaArchive, core.MarshalJSON(a))
}

func (b *MangaBucket) SetCover(c *core.Cover) error {
	return b.Put(keyMangaCover, core.MarshalJSON(c))
}

func (b *MangaBucket) SetMetadata(m *core.EntryMetadata) error {
	if m != nil && m.DateReleased == nil {
		m.DateReleased = core.NewDate(time.Time{})
	}

	return b.Put(keyMangaMetadata, core.MarshalJSON(m))
}

func (b *MangaBucket) SetTitle(t string) error {
	return b.Put(keyMangaTitle, core.MarshalJSON(t))
}

func (b *MangaBucket) SetThumbnail(thumb []byte) error {
	return b.Put(keyMangaThumbnail, thumb)
}

func (b *MangaBucket) Cover() *core.Cover {
	return core.UnmarshalCover(b.Get(keyMangaCover))
}

func (b *MangaBucket) ArchiveCover() *core.Cover {
	return  b.Archive().Cover
}

func (b *MangaBucket) ArchiveCoverBytes() ([]byte, error) {
	data, err := b.Archive().CoverBytes()
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (b *MangaBucket) Archive() *core.Archive {
	return core.UnmarshalArchive(b.Get(keyMangaArchive))
}

func (b *MangaBucket) Metadata() *core.EntryMetadata {
	m := core.UnmarshalEntryMetadata(b.Bucket.Get(keySeriesMetadata))
	if m != nil && m.DateReleased == nil {
		m.DateReleased = core.NewDate(time.Time{})
	}

	return m
}

func (b *MangaBucket) Title() string {
	return core.UnmarshalString(b.Get(keyMangaTitle))
}

func (b *MangaBucket) HasThumbnail() bool {
	return len(b.Thumbnail()) > 0
}

func (b *MangaBucket) Thumbnail() []byte {
	return b.Get(keyMangaThumbnail)
}
