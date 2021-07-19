package db

import (
	bolt "go.etcd.io/bbolt"

	"github.com/fiwippi/tanuki/pkg/core"
)

type PagesBucket struct {
	*bolt.Bucket
}

func (b *PagesBucket) Num() int {
	return b.Stats().KeyN
}

func (b *PagesBucket) SetPage(num int, p *core.Page) error {
	return b.Put(core.MarshalJSON(num), core.MarshalJSON(p))
}

func (b *PagesBucket) GetPage(num int) *core.Page {
	d := b.Get(core.MarshalJSON(num))
	if d == nil {
		return nil
	}
	return core.UnmarshalPage(d)
}
