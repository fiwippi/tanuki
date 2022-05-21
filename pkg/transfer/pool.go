package transfer

import "sync"

const poolStartCap = 20

// Pool implements a sync.Pool to improve performance when
// creating new []*Download for sending to each user
type Pool struct {
	pool *sync.Pool
}

func NewPool() *Pool {
	return &Pool{pool: &sync.Pool{}}
}

func (p Pool) Get() []*Download {
	dp := p.pool.Get()
	if dp == nil {
		return make([]*Download, 0, poolStartCap)
	}

	dl := dp.([]*Download)
	for i := range dl {
		dl[i] = nil
	}

	return dl[0:0]
}

func (p Pool) Put(dl []*Download) {
	p.pool.Put(dl)
}
