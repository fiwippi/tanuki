package buckets

import (
	"errors"
	"github.com/fiwippi/tanuki/internal/json"
	"github.com/fiwippi/tanuki/pkg/store/entities/manga"
	bolt "go.etcd.io/bbolt"
)

var ErrPageNotExist = errors.New("page does not exist")

type PagesBucket struct {
	*bolt.Bucket
}

func (b *PagesBucket) Num() int {
	return b.Stats().KeyN
}

func (b *PagesBucket) SetPage(num int, p *manga.Page) error {
	return b.Put(json.Marshal(num), json.Marshal(p))
}

func (b *PagesBucket) GetPage(num int) (*manga.Page, error) {
	if num < 1 || num > b.Num() {
		return nil, ErrPageNotExist
	}
	return manga.UnmarshalPage(b.Get(json.Marshal(num))), nil
}
