package transfer

import (
	"sync"

	"github.com/fiwippi/tanuki/pkg/mangadex"
)

const poolStartCap = 20

// Pool implements a sync.Pool to improve performance when
// creating new []*Download for sending to each user
type Pool struct {
	pool *sync.Pool
}

func NewPool() *Pool {
	return &Pool{pool: &sync.Pool{}}
}

func (p Pool) Get() []*mangadex.Download {
	dp := p.pool.Get()
	if dp == nil {
		return make([]*mangadex.Download, 0, poolStartCap)
	}

	dl := dp.([]*mangadex.Download)
	for i := range dl {
		dl[i] = nil
	}

	return dl[0:0]
}

func (p Pool) Put(dl []*mangadex.Download) {
	p.pool.Put(dl)
}
