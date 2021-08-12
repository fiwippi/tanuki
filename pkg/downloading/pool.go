package downloading

import (
	"github.com/fiwippi/tanuki/pkg/store/entities/api"
	"sync"
)

const poolStartCap = 20

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
