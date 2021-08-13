package downloading

import (
	"sync"

	"github.com/fiwippi/tanuki/pkg/store/entities/api"
)

const poolStartCap = 20

// Pool implements a sync.Pool to improve performance when
// creating new []*api.Download for sending to each user
type Pool struct {
	pool *sync.Pool
}

func NewPool() *Pool {
	return &Pool{pool: &sync.Pool{}}
}

func (p Pool) Get() []*api.Download {
	dp := p.pool.Get()
	if dp == nil {
		return make([]*api.Download, 0, poolStartCap)
	}

	dl := dp.([]*api.Download)
	for i := range dl {
		dl[i] = nil
	}

	return dl[0:0]
}

func (p Pool) Put(dl []*api.Download) {
	p.pool.Put(dl)
}
